package ui

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tokuhirom/dcv/internal/docker"
)

// logReader manages log streaming from a container
type logReader struct {
	client        *docker.Client
	containerName string
	isDind        bool
	hostContainer string
	cmd           *exec.Cmd
	stdout        io.ReadCloser
	stderr        io.ReadCloser
	lines         []string
	mu            sync.Mutex
	done          bool
}

// newLogReader creates a new log reader
func newLogReader(client *docker.Client, serviceName string, isDind bool, hostService string) (*logReader, error) {
	lr := &logReader{
		client:        client,
		containerName: serviceName, // Keep field name for compatibility but it's actually service name
		isDind:        isDind,
		hostContainer: hostService, // Keep field name for compatibility but it's actually service name
		lines:         make([]string, 0),
	}

	var err error
	if isDind && hostService != "" {
		// For dind, serviceName is the container name inside dind
		lr.cmd, err = client.GetDindContainerLogs(hostService, serviceName, true)
	} else {
		// For regular logs, use service name
		lr.cmd, err = client.GetContainerLogs(serviceName, true)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create log command: %w", err)
	}

	lr.stdout, err = lr.cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	lr.stderr, err = lr.cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Log the command execution
	startTime := time.Now()
	if err := lr.cmd.Start(); err != nil {
		slog.Error("Failed to start log command",
			slog.String("command", strings.Join(lr.cmd.Args, " ")),
			slog.Duration("startTime", time.Since(startTime)),
			slog.Any("error", err))
		return nil, fmt.Errorf("failed to start log command '%s': %w", strings.Join(lr.cmd.Args, " "), err)
	}

	// Log successful start
	slog.Info("Log command started",
		slog.String("command", strings.Join(lr.cmd.Args, " ")),
		slog.Duration("startTime", time.Since(startTime)),
		slog.String("containerName", lr.containerName),
		slog.Bool("isDind", lr.isDind),
		slog.String("hostContainer", lr.hostContainer))

	// Start reading logs in background
	go lr.readLogs()

	return lr, nil
}

// readLogs reads logs from stdout and stderr
func (lr *logReader) readLogs() {
	var wg sync.WaitGroup
	wg.Add(2)

	// Read stdout
	go func() {
		defer wg.Done()
		slog.Debug("Log reader started for stdout.")
		scanner := bufio.NewScanner(lr.stdout)
		for scanner.Scan() {
			got := scanner.Text()
			slog.Debug("Got stdout line",
				slog.String("line", got))
			lr.mu.Lock()
			lr.lines = append(lr.lines, got)
			lr.mu.Unlock()
		}
		slog.Debug("Log reader finished(stdout).")
		if err := scanner.Err(); err != nil {
			lr.mu.Lock()
			lr.lines = append(lr.lines, fmt.Sprintf("[ERROR reading stdout: %v]", err))
			lr.mu.Unlock()
		}
	}()

	// Read stderr
	go func() {
		defer wg.Done()
		slog.Debug("Log reader started for stderr.")
		scanner := bufio.NewScanner(lr.stderr)
		for scanner.Scan() {
			slog.Debug("Got stderr line")
			lr.mu.Lock()
			lr.lines = append(lr.lines, fmt.Sprintf("[STDERR] %s", scanner.Text()))
			lr.mu.Unlock()
		}
		slog.Debug("Log reader finished(stderr).")
		if err := scanner.Err(); err != nil {
			lr.mu.Lock()
			lr.lines = append(lr.lines, fmt.Sprintf("[ERROR reading stderr: %v]", err))
			lr.mu.Unlock()
		}
	}()

	wg.Wait()
	if err := lr.cmd.Wait(); err != nil {
		lr.mu.Lock()
		lr.lines = append(lr.lines, fmt.Sprintf("[ERROR: Command failed: %v]", err))
		lr.mu.Unlock()
	}

	slog.Debug("Log reader finished(wait).")
	lr.mu.Lock()
	lr.done = true
	lr.mu.Unlock()
}

// getNewLines returns any new log lines
func (lr *logReader) getNewLines(lastIndex int) ([]string, int, bool) {
	lr.mu.Lock()
	defer lr.mu.Unlock()

	if lastIndex >= len(lr.lines) {
		return nil, lastIndex, lr.done
	}

	newLines := lr.lines[lastIndex:]
	return newLines, len(lr.lines), lr.done
}

// Global log reader storage
var activeLogReader *logReader
var logReaderMu sync.Mutex
var lastLogIndex int

// streamLogsReal creates a command that starts log streaming
func streamLogsReal(client *docker.Client, serviceName string, isDind bool, hostService string) tea.Cmd {
	return func() tea.Msg {
		logReaderMu.Lock()
		defer logReaderMu.Unlock()

		// Stop any existing log reader
		if activeLogReader != nil && activeLogReader.cmd != nil {
			slog.Debug("Stopping existing log reader")
			if activeLogReader.cmd.Process != nil {
				activeLogReader.cmd.Process.Kill()
				activeLogReader.cmd.Wait() // Wait for process to terminate
			}
		}

		// Create new log reader
		slog.Debug("Creating new log reader",
			slog.String("serviceName", serviceName))
		lr, err := newLogReader(client, serviceName, isDind, hostService)
		if err != nil {
			slog.Info("Failed to create log reader",
				slog.String("serviceName", serviceName),
				slog.Any("error", err))
			return errorMsg{err: err}
		}

		activeLogReader = lr
		lastLogIndex = 0
		slog.Info("Log reader created",
			slog.String("serviceName", serviceName),
			slog.Bool("isDind", isDind),
			slog.String("hostService", hostService))

		// Send command info message
		cmdStr := strings.Join(lr.cmd.Args, " ")
		return commandExecutedMsg{command: cmdStr}
	}
}

// stopLogReader stops the active log reader
func stopLogReader() {
	logReaderMu.Lock()
	defer logReaderMu.Unlock()

	if activeLogReader != nil {
		if activeLogReader.cmd != nil && activeLogReader.cmd.Process != nil {
			activeLogReader.cmd.Process.Kill()
			// Don't wait here as it might block
		}
		activeLogReader = nil
		lastLogIndex = 0 // Reset the index too
	}
}

// pollForLogs polls for new log lines
func pollForLogs() tea.Cmd {
	return func() tea.Msg {
		logReaderMu.Lock()
		defer logReaderMu.Unlock()

		if activeLogReader == nil {
			// Don't log here as we don't have access to client
			return logLineMsg{line: "[Log reader stopped]"}
		}

		newLines, newIndex, done := activeLogReader.getNewLines(lastLogIndex)
		lastLogIndex = newIndex
		slog.Debug("Got new log lines",
			slog.Int("newLinesCount", len(newLines)))

		if len(newLines) > 0 {
			// Return all new lines at once
			return logLinesMsg{lines: newLines}
		}

		if done {
			// Log streaming finished
			if lastLogIndex == 0 {
				activeLogReader = nil
				return logLineMsg{line: "[No logs available for this container]"}
			}
			activeLogReader = nil
			return nil
		}

		// No new lines, continue polling
		return pollLogsContinueMsg{}
	}
}
