package ui

import (
	"bufio"
	"io"
	"os/exec"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tokuhirom/dcv/internal/docker"
)

// tickCmd sends periodic ticks for updates
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type tickMsg time.Time

// logReader manages log streaming from a container
type logReader struct {
	client        *docker.ComposeClient
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
func newLogReader(client *docker.ComposeClient, containerName string, isDind bool, hostContainer string) (*logReader, error) {
	lr := &logReader{
		client:        client,
		containerName: containerName,
		isDind:        isDind,
		hostContainer: hostContainer,
		lines:         make([]string, 0),
	}

	var err error
	if isDind && hostContainer != "" {
		lr.cmd, err = client.GetDindContainerLogs(hostContainer, containerName, true)
	} else {
		lr.cmd, err = client.GetContainerLogs(containerName, true)
	}

	if err != nil {
		return nil, err
	}

	lr.stdout, err = lr.cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	lr.stderr, err = lr.cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := lr.cmd.Start(); err != nil {
		return nil, err
	}

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
		scanner := bufio.NewScanner(lr.stdout)
		for scanner.Scan() {
			lr.mu.Lock()
			lr.lines = append(lr.lines, scanner.Text())
			lr.mu.Unlock()
		}
	}()

	// Read stderr
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(lr.stderr)
		for scanner.Scan() {
			lr.mu.Lock()
			lr.lines = append(lr.lines, scanner.Text())
			lr.mu.Unlock()
		}
	}()

	wg.Wait()
	lr.cmd.Wait()

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
func streamLogsReal(client *docker.ComposeClient, containerName string, isDind bool, hostContainer string) tea.Cmd {
	return func() tea.Msg {
		logReaderMu.Lock()
		defer logReaderMu.Unlock()

		// Stop any existing log reader
		if activeLogReader != nil && activeLogReader.cmd != nil {
			activeLogReader.cmd.Process.Kill()
		}

		// Create new log reader
		lr, err := newLogReader(client, containerName, isDind, hostContainer)
		if err != nil {
			return errorMsg{err: err}
		}

		activeLogReader = lr
		lastLogIndex = 0

		// Start polling for logs
		return pollForLogs()()
	}
}

// pollForLogs polls for new log lines
func pollForLogs() tea.Cmd {
	return func() tea.Msg {
		logReaderMu.Lock()
		defer logReaderMu.Unlock()

		if activeLogReader == nil {
			return nil
		}

		newLines, newIndex, done := activeLogReader.getNewLines(lastLogIndex)
		lastLogIndex = newIndex

		if len(newLines) > 0 {
			// Return first new line
			return logLineMsg{line: newLines[0]}
		}

		if done {
			// Log streaming finished
			activeLogReader = nil
			return nil
		}

		// No new lines, wait a bit
		time.Sleep(100 * time.Millisecond)
		return pollForLogs()()
	}
}