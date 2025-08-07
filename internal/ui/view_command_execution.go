package ui

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tokuhirom/dcv/internal/docker"
)

type CommandExecutionViewModel struct {
	cmd          *exec.Cmd
	output       []string
	scrollY      int
	done         bool
	exitCode     int
	cmdString    string
	reader       *bufio.Reader
	previousView ViewType
}

func (m *CommandExecutionViewModel) render(model *Model) string {
	if model.width == 0 || model.Height == 0 {
		return "Loading..."
	}

	// Create viewport
	vp := viewport.New(model.width, model.Height-4)

	// Build content
	var content strings.Builder

	// Show command
	content.WriteString(lipgloss.NewStyle().Bold(true).Render("Executing: "))
	content.WriteString(m.cmdString)
	content.WriteString("\n\n")

	// Show output
	for _, line := range m.output {
		content.WriteString(line)
		content.WriteString("\n")
	}

	// Show status
	if m.done {
		content.WriteString("\n")
		if m.exitCode == 0 {
			content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("✓ Command completed successfully"))
		} else {
			content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render(fmt.Sprintf("✗ Command failed with exit code %d", m.exitCode)))
		}
	} else {
		content.WriteString("\n")
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render("⠋ Running... (Press Ctrl+C to cancel)"))
	}

	vp.SetContent(content.String())
	vp.YOffset = m.scrollY

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("7")).
		Background(lipgloss.Color("4")).
		Width(model.width).
		Padding(0, 1).
		Render("Command Execution")

	// Footer
	var footerText string
	if m.done {
		footerText = "Press ESC or q to go back"
	} else {
		footerText = "Press Ctrl+C to cancel, ESC or q to go back"
	}
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("7")).
		Width(model.width).
		Padding(0, 1).
		Align(lipgloss.Center).
		Render(footerText)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		vp.View(),
		footer,
	)
}

// Command execution view key handlers
func (m *CommandExecutionViewModel) HandleUp() tea.Cmd {
	if m.scrollY > 0 {
		m.scrollY--
	}
	return nil
}

func (m *CommandExecutionViewModel) HandleDown(model *Model) tea.Cmd {
	maxScroll := len(m.output) - (model.Height - 6)
	if m.scrollY < maxScroll && maxScroll > 0 {
		m.scrollY++
	}
	return nil
}

func (m *CommandExecutionViewModel) HandleGoToEnd(model *Model) tea.Cmd {
	maxScroll := len(m.output) - (model.Height - 6)
	if maxScroll > 0 {
		m.scrollY = maxScroll
	}
	return nil
}

func (m *CommandExecutionViewModel) HandleGoToStart() tea.Cmd {
	m.scrollY = 0
	return nil
}

func (m *CommandExecutionViewModel) HandleCancel() tea.Cmd {
	if m.cmd != nil && !m.done {
		// Kill the process
		if err := m.cmd.Process.Kill(); err != nil {
			m.output = append(m.output, fmt.Sprintf("Error cancelling command: %v", err))
		} else {
			m.output = append(m.output, "Command cancelled by user")
		}
		m.done = true
		m.exitCode = -1
	}
	return nil
}

func (m *CommandExecutionViewModel) HandleBack(model *Model) tea.Cmd {
	// If the command is still running, cancel it first
	if m.cmd != nil && !m.done {
		if err := m.cmd.Process.Kill(); err != nil {
			// Log error but continue
			model.err = err
		}
	}

	// Go back to previous view
	model.currentView = m.previousView
	model.loading = true

	// TODO: CommandExecutionModelView でここを管理するのは微妙｡refresh してもらう旨だけ､メッセージ送ればよろしいかもしれず｡

	// Reload data for the previous view
	switch m.previousView {
	case ComposeProcessListView:
		return loadProcesses(model.dockerClient, model.projectName, model.composeProcessListViewModel.showAll)
	case DockerContainerListView:
		return loadDockerContainers(model.dockerClient, model.dockerContainerListViewModel.showAll)
	default:
		model.loading = false
		return nil
	}
}

func executeComposeCommandWithStreaming(client *docker.Client, projectName string, operation string) tea.Cmd {
	return func() tea.Msg {
		// Create the command based on operation
		var cmd *exec.Cmd
		switch operation {
		case "up":
			cmd = exec.Command("docker", "compose", "-p", projectName, "up", "-d")
		case "down":
			cmd = exec.Command("docker", "compose", "-p", projectName, "down")
		default:
			return errorMsg{err: fmt.Errorf("unknown operation: %s", operation)}
		}

		// Create pipes for stdout and stderr
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return errorMsg{err: fmt.Errorf("failed to create stdout pipe: %w", err)}
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return errorMsg{err: fmt.Errorf("failed to create stderr pipe: %w", err)}
		}

		// HandleStart the command
		if err := cmd.Start(); err != nil {
			return errorMsg{err: fmt.Errorf("failed to start command: %w", err)}
		}

		// Store the command reference and output channel
		return commandExecStartedMsg{cmd: cmd, stdout: stdout, stderr: stderr}
	}
}

func executeContainerCommand(client *docker.Client, containerID string, operation string) tea.Cmd {
	return func() tea.Msg {
		// Create the command based on operation
		var cmd *exec.Cmd
		switch operation {
		case "start":
			cmd = exec.Command("docker", "start", containerID)
		case "stop":
			cmd = exec.Command("docker", "stop", containerID)
		case "restart":
			cmd = exec.Command("docker", "restart", containerID)
		case "kill":
			cmd = exec.Command("docker", "kill", containerID)
		case "rm":
			cmd = exec.Command("docker", "rm", containerID)
		default:
			return errorMsg{err: fmt.Errorf("unknown operation: %s", operation)}
		}

		// Create pipes for stdout and stderr
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return errorMsg{err: fmt.Errorf("failed to create stdout pipe: %w", err)}
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return errorMsg{err: fmt.Errorf("failed to create stderr pipe: %w", err)}
		}

		// HandleStart the command
		if err := cmd.Start(); err != nil {
			return errorMsg{err: fmt.Errorf("failed to start command: %w", err)}
		}

		// Store the command reference and output channel
		return commandExecStartedMsg{cmd: cmd, stdout: stdout, stderr: stderr}
	}
}

func (m *CommandExecutionViewModel) ExecuteContainerCommand(model *Model, previousView ViewType, containerID string, operation string) tea.Cmd {
	m.previousView = previousView
	model.SwitchView(CommandExecutionView)
	m.output = []string{}
	m.scrollY = 0
	m.done = false
	m.cmdString = fmt.Sprintf("docker %s %s", operation, containerID)

	return executeContainerCommand(model.dockerClient, containerID, operation)
}

func (m *CommandExecutionViewModel) ExecuteComposeCommand(model *Model, operation string) tea.Cmd {
	m.previousView = model.currentView
	model.currentView = CommandExecutionView
	m.output = []string{}
	m.scrollY = 0
	m.done = false
	m.cmdString = fmt.Sprintf("docker compose -p %s up -d", model.projectName)
	return executeComposeCommandWithStreaming(model.dockerClient, model.projectName, operation)
}

func (m *CommandExecutionViewModel) ExecStarted(cmd *exec.Cmd, stdout io.ReadCloser, stderr io.ReadCloser) tea.Cmd {
	m.cmd = cmd
	m.reader = bufio.NewReader(io.MultiReader(stdout, stderr))
	return streamCommandFromReader(m)
}

func (m *CommandExecutionViewModel) ExecOutput(model *Model, line string) tea.Cmd {
	m.output = append(m.output, line)
	// Auto-scroll to bottom
	maxScroll := len(m.output) - (model.Height - 6)
	if maxScroll > 0 && m.scrollY == maxScroll-1 {
		m.scrollY = maxScroll
	}
	// Continue reading output
	return streamCommandFromReader(m)
}

func (m *CommandExecutionViewModel) Complete(code int) {
	m.done = true
	m.exitCode = code
}

func streamCommandFromReader(m *CommandExecutionViewModel) tea.Cmd {
	return func() tea.Msg {
		if m.reader == nil || m.cmd == nil {
			return commandExecCompleteMsg{exitCode: 1}
		}

		// Read one line
		line, err := m.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// Command finished, wait for exit code
				exitCode := 0
				if waitErr := m.cmd.Wait(); waitErr != nil {
					var exitErr *exec.ExitError
					if errors.As(waitErr, &exitErr) {
						exitCode = exitErr.ExitCode()
					}
				}
				return commandExecCompleteMsg{exitCode: exitCode}
			}
			// Other error, treat as completion
			return commandExecCompleteMsg{exitCode: 1}
		}

		// Remove trailing newline
		line = strings.TrimRight(line, "\n\r")
		return commandExecOutputMsg{line: line}
	}
}
