package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) renderCommandExecutionView() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	// Create viewport
	vp := viewport.New(m.width, m.height-4)

	// Build content
	var content strings.Builder

	// Show command
	content.WriteString(lipgloss.NewStyle().Bold(true).Render("Executing: "))
	content.WriteString(m.commandExecCmdString)
	content.WriteString("\n\n")

	// Show output
	for _, line := range m.commandExecOutput {
		content.WriteString(line)
		content.WriteString("\n")
	}

	// Show status
	if m.commandExecDone {
		content.WriteString("\n")
		if m.commandExecExitCode == 0 {
			content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("✓ Command completed successfully"))
		} else {
			content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render(fmt.Sprintf("✗ Command failed with exit code %d", m.commandExecExitCode)))
		}
	} else {
		content.WriteString("\n")
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render("⠋ Running... (Press Ctrl+C to cancel)"))
	}

	vp.SetContent(content.String())
	vp.YOffset = m.commandExecScrollY

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("7")).
		Background(lipgloss.Color("4")).
		Width(m.width).
		Padding(0, 1).
		Render("Command Execution")

	// Footer
	var footerText string
	if m.commandExecDone {
		footerText = "Press ESC or q to go back"
	} else {
		footerText = "Press Ctrl+C to cancel, ESC or q to go back"
	}
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("7")).
		Width(m.width).
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
