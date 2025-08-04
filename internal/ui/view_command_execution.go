package ui

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tokuhirom/dcv/internal/docker"
)

type CommandExecutionViewModel struct {
}

func (m *CommandExecutionViewModel) render(model *Model) string {
	if model.width == 0 || model.height == 0 {
		return "Loading..."
	}

	// Create viewport
	vp := viewport.New(model.width, model.height-4)

	// Build content
	var content strings.Builder

	// Show command
	content.WriteString(lipgloss.NewStyle().Bold(true).Render("Executing: "))
	content.WriteString(model.commandExecCmdString)
	content.WriteString("\n\n")

	// Show output
	for _, line := range model.commandExecOutput {
		content.WriteString(line)
		content.WriteString("\n")
	}

	// Show status
	if model.commandExecDone {
		content.WriteString("\n")
		if model.commandExecExitCode == 0 {
			content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("✓ Command completed successfully"))
		} else {
			content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render(fmt.Sprintf("✗ Command failed with exit code %d", model.commandExecExitCode)))
		}
	} else {
		content.WriteString("\n")
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render("⠋ Running... (Press Ctrl+C to cancel)"))
	}

	vp.SetContent(content.String())
	vp.YOffset = model.commandExecScrollY

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
	if model.commandExecDone {
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
func (m *Model) ScrollCommandExecUp(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.commandExecScrollY > 0 {
		m.commandExecScrollY--
	}
	return m, nil
}

func (m *Model) ScrollCommandExecDown(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	maxScroll := len(m.commandExecOutput) - (m.height - 6)
	if m.commandExecScrollY < maxScroll && maxScroll > 0 {
		m.commandExecScrollY++
	}
	return m, nil
}

func (m *Model) GoToCommandExecEnd(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	maxScroll := len(m.commandExecOutput) - (m.height - 6)
	if maxScroll > 0 {
		m.commandExecScrollY = maxScroll
	}
	return m, nil
}

func (m *Model) GoToCommandExecStart(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.commandExecScrollY = 0
	return m, nil
}

func (m *Model) CancelCommandExec(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.commandExecCmd != nil && !m.commandExecDone {
		// Kill the process
		if err := m.commandExecCmd.Process.Kill(); err != nil {
			m.commandExecOutput = append(m.commandExecOutput, fmt.Sprintf("Error cancelling command: %v", err))
		} else {
			m.commandExecOutput = append(m.commandExecOutput, "Command cancelled by user")
		}
		m.commandExecDone = true
		m.commandExecExitCode = -1
	}
	return m, nil
}

func (m *Model) BackFromCommandExec(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If command is still running, cancel it first
	if m.commandExecCmd != nil && !m.commandExecDone {
		if err := m.commandExecCmd.Process.Kill(); err != nil {
			// Log error but continue
			m.err = err
		}
	}

	// Go back to previous view
	m.currentView = m.previousView
	m.loading = true

	// Reload data for the previous view
	switch m.previousView {
	case ComposeProcessListView:
		return m, loadProcesses(m.dockerClient, m.projectName, m.showAll)
	case DockerContainerListView:
		return m, loadDockerContainers(m.dockerClient, m.showAll)
	default:
		m.loading = false
		return m, nil
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
	model.previousView = previousView
	model.currentView = CommandExecutionView
	model.commandExecOutput = []string{}
	model.commandExecScrollY = 0
	model.commandExecDone = false
	model.commandExecCmdString = fmt.Sprintf("docker %s %s", operation, containerID)

	return executeContainerCommand(model.dockerClient, containerID, operation)
}
