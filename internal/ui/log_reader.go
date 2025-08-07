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
	client              *docker.Client
	targetContainerID   string
	isDind              bool
	dindHostContainerID string
	cmd                 *exec.Cmd
	stdout              io.ReadCloser
	stderr              io.ReadCloser
	lines               []string
	mu                  sync.Mutex
	done                bool
}

// newLogReader creates a new log reader
func newLogReader(client *docker.Client, targetContainerID string, isDind bool, dindHostContainerID string) (*logReader, error) {
	lr := &logReader{
		client:              client,
		targetContainerID:   targetContainerID,
		isDind:              isDind,
		dindHostContainerID: dindHostContainerID,
		lines:               make([]string, 0),
	}

	if isDind && dindHostContainerID != "" {
		// For dind, targetContainerID is the container name inside dind
		lr.cmd = client.Dind(dindHostContainerID).Execute(
			"logs", targetContainerID, "--tail", "1000", "--timestamps", "--follow")
	} else {
		// For regular logs, use service name
		lr.cmd = client.Execute("logs", targetContainerID, "--tail", "1000", "--timestamps", "--follow")
	}

	var err error
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
		slog.String("targetContainerID", lr.targetContainerID),
		slog.Bool("isDind", lr.isDind),
		slog.String("dindHostContainerID", lr.dindHostContainerID))

	// HandleStart reading logs in background
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

type LogReaderManager struct {
	logReaderMu     sync.Mutex
	lastLogIndex    int
	activeLogReader *logReader
}

// streamLogsReal creates a command that starts log streaming
func (lrm *LogReaderManager) streamLogsReal(client *docker.Client, targetContainerID string, isDind bool, dindHostContainerID string) tea.Cmd {
	return func() tea.Msg {
		lrm.logReaderMu.Lock()
		defer lrm.logReaderMu.Unlock()

		// Stop any existing log reader
		if lrm.activeLogReader != nil && lrm.activeLogReader.cmd != nil {
			slog.Debug("Stopping existing log reader")
			if lrm.activeLogReader.cmd.Process != nil {
				if err := lrm.activeLogReader.cmd.Process.Kill(); err != nil {
					slog.Warn("Failed to kill log reader process", slog.Any("error", err))
				}
				if err := lrm.activeLogReader.cmd.Wait(); err != nil {
					slog.Debug("Log reader process wait error", slog.Any("error", err))
				}
			}
		}

		// Create new log reader
		slog.Debug("Creating new log reader",
			slog.String("targetContainerID", targetContainerID))
		lr, err := newLogReader(client, targetContainerID, isDind, dindHostContainerID)
		if err != nil {
			slog.Info("Failed to create log reader",
				slog.String("targetContainerID", targetContainerID),
				slog.Any("error", err))
			return errorMsg{err: err}
		}

		lrm.activeLogReader = lr
		lrm.lastLogIndex = 0
		slog.Info("Log reader created",
			slog.String("targetContainerID", targetContainerID),
			slog.Bool("isDind", isDind),
			slog.String("dindHostContainerID", dindHostContainerID))

		// Send command info message
		cmdStr := strings.Join(lr.cmd.Args, " ")
		return commandExecutedMsg{command: cmdStr}
	}
}

// stopLogReader stops the active log reader
func (lrm *LogReaderManager) stopLogReader() {
	lrm.logReaderMu.Lock()
	defer lrm.logReaderMu.Unlock()

	if lrm.activeLogReader != nil {
		if lrm.activeLogReader.cmd != nil && lrm.activeLogReader.cmd.Process != nil {
			if err := lrm.activeLogReader.cmd.Process.Kill(); err != nil {
				slog.Warn("Failed to kill log reader process in stopLogReader", slog.Any("error", err))
			}
			// Don't wait here as it might block
		}
		lrm.activeLogReader = nil
		lrm.lastLogIndex = 0 // Reset the index too
	}
}

// pollForLogs polls for new log lines
func (lrm *LogReaderManager) pollForLogs() tea.Cmd {
	return func() tea.Msg {
		lrm.logReaderMu.Lock()
		defer lrm.logReaderMu.Unlock()

		if lrm.activeLogReader == nil {
			// Don't log here as we don't have access to client
			return logLineMsg{line: "[Log reader stopped]"}
		}

		newLines, newIndex, done := lrm.activeLogReader.getNewLines(lrm.lastLogIndex)
		lrm.lastLogIndex = newIndex
		slog.Debug("Got new log lines",
			slog.Int("newLinesCount", len(newLines)))

		if len(newLines) > 0 {
			// Return all new lines at once
			return logLinesMsg{lines: newLines}
		}

		if done {
			// Log streaming finished
			if lrm.lastLogIndex == 0 {
				lrm.activeLogReader = nil
				return logLineMsg{line: "[No logs available for this container]"}
			}
			lrm.activeLogReader = nil
			return nil
		}

		// No new lines, continue polling
		return pollLogsContinueMsg{}
	}
}
