package ui

import (
	"context"
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

type HelperInjectorViewModel struct {
	container       *docker.Container
	commands        []string
	currentStep     int
	totalSteps      int
	output          []string
	scrollY         int
	done            bool
	success         bool
	err             error
	currentCmd      *exec.Cmd
	currentCmdStr   string // Store the current command string for display
	pendingCommands []string
}

func (m *HelperInjectorViewModel) render(model *Model) string {
	if model.width == 0 || model.Height == 0 {
		return "Loading..."
	}

	// Create viewport
	vp := viewport.New(model.width, model.Height-4)

	// Build content
	var content strings.Builder

	// Show progress
	content.WriteString(lipgloss.NewStyle().Bold(true).Render("Helper Binary Injection"))
	content.WriteString("\n\n")

	// Show container info
	content.WriteString(fmt.Sprintf("Container: %s\n", m.container.Title()))
	if m.container.IsDind() {
		content.WriteString(fmt.Sprintf("Host Container: %s\n", m.container.HostContainerID()))
	}
	content.WriteString("\n")

	// Show current step
	if m.totalSteps > 0 {
		progress := fmt.Sprintf("Step %d/%d", m.currentStep, m.totalSteps)
		content.WriteString(lipgloss.NewStyle().Bold(true).Render(progress))
		content.WriteString("\n")
	}

	// Show current or last executed command
	if m.currentCmdStr != "" {
		if !m.done {
			content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render("Executing: "))
		} else if !m.success {
			content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render("Failed command: "))
		} else {
			content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("Last command: "))
		}
		content.WriteString(lipgloss.NewStyle().Bold(true).Render(m.currentCmdStr))
		content.WriteString("\n\n")
	}

	// Show output
	if len(m.output) > 0 {
		content.WriteString(lipgloss.NewStyle().Bold(true).Render("Output:"))
		content.WriteString("\n")
		for _, line := range m.output {
			content.WriteString(line)
			content.WriteString("\n")
		}
	}

	// Show all commands for reference
	if m.totalSteps > 0 {
		content.WriteString("\n")
		content.WriteString(lipgloss.NewStyle().Bold(true).Render("Commands to execute:"))
		content.WriteString("\n")
		for i, cmd := range m.commands {
			prefix := "  "
			style := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
			if i+1 < m.currentStep {
				prefix = "✓ "
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
			} else if i+1 == m.currentStep && !m.done {
				prefix = "▶ "
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true)
			} else if i+1 == m.currentStep && m.done && !m.success {
				prefix = "✗ "
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
			}
			content.WriteString(fmt.Sprintf("%s%d. %s\n", prefix, i+1, style.Render(cmd)))
		}
	}

	// Show status
	if m.done {
		content.WriteString("\n")
		if m.success {
			content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true).Render("✓ Helper binary injected successfully"))
			content.WriteString("\n")
			content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("The helper is now available at: /.dcv-helper"))
		} else {
			content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true).Render("✗ Helper injection failed"))
			if m.err != nil {
				content.WriteString("\n")
				content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render(fmt.Sprintf("Error: %v", m.err)))
			}
			content.WriteString("\n\n")
			content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render("💡 Tip: Check if the container has write permissions to /"))
		}
	} else if m.currentStep > 0 {
		content.WriteString("\n")
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render("⠋ Injecting helper binary..."))
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
		Render("Helper Injection")

	// Footer
	var footerText string
	if m.done {
		footerText = "Press ESC to go back"
	} else {
		footerText = "Press Ctrl+C to cancel, ESC to go back"
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

// HandleInjectHelper starts the helper injection process
func (m *HelperInjectorViewModel) HandleInjectHelper(model *Model, container *docker.Container) tea.Cmd {
	model.SwitchView(HelperInjectorView)

	m.container = container
	m.output = []string{}
	m.scrollY = 0
	m.done = false
	m.success = false
	m.err = nil
	m.currentStep = 0
	m.currentCmdStr = ""

	// Build commands using the helper injector
	if model.dockerSDKClient == nil {
		m.err = fmt.Errorf("docker SDK client not available")
		m.done = true
		return nil
	}
	injector := docker.NewHelperInjector(model.dockerSDKClient)

	// Get temporary file path for helper binary
	tempFile, err := m.prepareHelperBinary(injector, container)
	if err != nil {
		m.err = err
		m.done = true
		return nil
	}

	m.commands = injector.BuildCommands(container, tempFile)
	m.totalSteps = len(m.commands)
	m.pendingCommands = m.commands

	// Start executing the first command
	return m.executeNextCommand()
}

func (m *HelperInjectorViewModel) prepareHelperBinary(injector *docker.HelperInjector, container *docker.Container) (string, error) {
	// Get the temporary file with the helper binary
	ctx := context.Background()
	return injector.GetHelperTempFile(ctx, container)
}

func (m *HelperInjectorViewModel) executeNextCommand() tea.Cmd {
	if len(m.pendingCommands) == 0 {
		// All commands executed
		m.done = true
		m.success = true
		return nil
	}

	m.currentStep++
	cmdStr := m.pendingCommands[0]
	m.pendingCommands = m.pendingCommands[1:]
	m.currentCmdStr = cmdStr // Store for display

	return func() tea.Msg {
		// Parse the command string
		parts := strings.Fields(cmdStr)
		if len(parts) == 0 {
			return helperInjectorCompleteMsg{success: false, err: fmt.Errorf("empty command")}
		}

		cmd := exec.Command(parts[0], parts[1:]...)

		// Create pipes for stdout and stderr
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return helperInjectorCompleteMsg{success: false, err: fmt.Errorf("failed to create stdout pipe: %w", err)}
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return helperInjectorCompleteMsg{success: false, err: fmt.Errorf("failed to create stderr pipe: %w", err)}
		}

		// Start the command
		if err := cmd.Start(); err != nil {
			return helperInjectorCompleteMsg{success: false, err: fmt.Errorf("failed to start command: %w", err)}
		}

		return helperInjectorStartedMsg{
			cmd:    cmd,
			stdout: stdout,
			stderr: stderr,
		}
	}
}

// HandleUp handles up arrow key to scroll up
func (m *HelperInjectorViewModel) HandleUp() tea.Cmd {
	if m.scrollY > 0 {
		m.scrollY--
	}
	return nil
}

func (m *HelperInjectorViewModel) HandleDown(model *Model) tea.Cmd {
	maxScroll := len(m.output) - model.PageSize()
	if m.scrollY < maxScroll && maxScroll > 0 {
		m.scrollY++
	}
	return nil
}

func (m *HelperInjectorViewModel) HandleGoToEnd(model *Model) tea.Cmd {
	maxScroll := len(m.output) - model.PageSize()
	if maxScroll > 0 {
		m.scrollY = maxScroll
	}
	return nil
}

func (m *HelperInjectorViewModel) HandleGoToBeginning() tea.Cmd {
	m.scrollY = 0
	return nil
}

func (m *HelperInjectorViewModel) HandleCancel() tea.Cmd {
	if m.currentCmd != nil && !m.done {
		// Kill the process
		if err := m.currentCmd.Process.Kill(); err != nil {
			m.output = append(m.output, fmt.Sprintf("Error cancelling command: %v", err))
		} else {
			m.output = append(m.output, "Command cancelled by user")
		}
		m.done = true
		m.success = false
	}
	return nil
}

func (m *HelperInjectorViewModel) HandleBack(model *Model) tea.Cmd {
	// If the command is still running, cancel it first
	if m.currentCmd != nil && !m.done {
		if err := m.currentCmd.Process.Kill(); err != nil {
			// Log error but continue
			model.err = err
		}
	}

	// Go back to previous view
	model.SwitchToPreviousView()

	// Trigger a refresh to reload the list with updated state
	return func() tea.Msg {
		return RefreshMsg{}
	}
}

func (m *HelperInjectorViewModel) ExecStarted(cmd *exec.Cmd, stdout io.ReadCloser, stderr io.ReadCloser) tea.Cmd {
	m.currentCmd = cmd
	// Read all output at once for simplicity
	return m.readCommandOutput(stdout, stderr)
}

func (m *HelperInjectorViewModel) readCommandOutput(stdout, stderr io.ReadCloser) tea.Cmd {
	return func() tea.Msg {
		// Read stdout
		stdoutBytes, _ := io.ReadAll(stdout)
		stderrBytes, _ := io.ReadAll(stderr)

		// Wait for command to complete
		var exitCode int
		if err := m.currentCmd.Wait(); err != nil {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				exitCode = exitErr.ExitCode()
			} else {
				exitCode = 1
			}
		}

		output := string(stdoutBytes)
		if len(stderrBytes) > 0 {
			if output != "" {
				output += "\n"
			}
			output += string(stderrBytes)
		}

		return helperInjectorOutputMsg{
			output:   output,
			exitCode: exitCode,
		}
	}
}

func (m *HelperInjectorViewModel) ExecOutput(output string, exitCode int) tea.Cmd {
	// Add command execution info
	m.output = append(m.output, fmt.Sprintf("[Step %d/%d] %s", m.currentStep, m.totalSteps, m.currentCmdStr))

	// Add output
	if output != "" {
		lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
		m.output = append(m.output, lines...)
	}

	// Check if command succeeded
	if exitCode != 0 {
		m.done = true
		m.success = false
		m.err = fmt.Errorf("command failed with exit code %d: %s", exitCode, m.currentCmdStr)
		m.output = append(m.output, fmt.Sprintf("ERROR: Exit code %d", exitCode))
		return nil
	} else {
		m.output = append(m.output, "SUCCESS")
	}

	// Execute next command
	return m.executeNextCommand()
}

func (m *HelperInjectorViewModel) Complete(success bool, err error) {
	m.done = true
	m.success = success
	m.err = err
}
