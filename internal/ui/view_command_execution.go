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
)

type CommandExecutionViewModel struct {
	cmd                 *exec.Cmd
	output              []string
	scrollY             int
	done                bool
	exitCode            int
	cmdString           string
	reader              *bufio.Reader
	pendingConfirmation bool
	pendingArgs         []string
	confirmationTarget  string // e.g., "container nginx" or "service web"
}

func (m *CommandExecutionViewModel) render(model *Model) string {
	if model.width == 0 || model.Height == 0 {
		return "Loading..."
	}

	// Show confirmation dialog if pending
	if m.pendingConfirmation {
		return m.renderConfirmationDialog(model)
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
	model.SwitchToPreviousView()
	return nil
}

func (m *CommandExecutionViewModel) ExecuteCommand(model *Model, args ...string) tea.Cmd {
	model.SwitchView(CommandExecutionView)

	m.output = []string{}
	m.scrollY = 0
	m.done = false

	// Check if this is an aggressive command that needs confirmation
	if m.needsConfirmation(args) {
		m.pendingConfirmation = true
		m.pendingArgs = args
		m.confirmationTarget = m.getConfirmationTarget(args)
		return nil
	}

	return func() tea.Msg {
		m.cmdString = fmt.Sprintf("docker %s", strings.Join(args, " "))

		// Create the command based on operation
		cmd := exec.Command("docker", args...)

		// Create pipes for stdout and stderr
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return errorMsg{err: fmt.Errorf("failed to create stdout pipe: %w", err)}
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return errorMsg{err: fmt.Errorf("failed to create stderr pipe: %w", err)}
		}

		// Start the command
		if err := cmd.Start(); err != nil {
			return errorMsg{err: fmt.Errorf("failed to start command: %w", err)}
		}

		// Store the command reference and output channel
		return commandExecStartedMsg{cmd: cmd, stdout: stdout, stderr: stderr}
	}
}

func (m *CommandExecutionViewModel) ExecuteComposeCommand(model *Model, projectName string, operation string) tea.Cmd {
	switch operation {
	case "up":
		args := []string{"compose", "-p", projectName, "up", "-d"}
		return m.ExecuteCommand(model, args...)
	case "down":
		args := []string{"compose", "-p", projectName, "down"}
		return m.ExecuteCommand(model, args...)
	default:
		return nil
	}
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

// needsConfirmation checks if a command is aggressive and needs user confirmation
func (m *CommandExecutionViewModel) needsConfirmation(args []string) bool {
	if len(args) == 0 {
		return false
	}

	// List of aggressive commands that need confirmation
	aggressiveCommands := map[string]bool{
		"stop":    true,
		"start":   true,
		"restart": true,
		"kill":    true,
		"pause":   true,
		"unpause": true,
		"rm":      true,
		"rmi":     true,
	}

	// Check the first argument (the docker command)
	if len(args) > 0 {
		if aggressive, exists := aggressiveCommands[args[0]]; exists && aggressive {
			return true
		}
	}

	// Check for compose commands
	if len(args) >= 2 && args[0] == "compose" {
		for i, arg := range args {
			if arg == "down" || arg == "stop" || arg == "restart" || arg == "kill" || arg == "rm" {
				return true
			}
			// Check for service-specific commands like "compose restart service_name"
			if i > 0 && (args[i-1] == "restart" || args[i-1] == "stop" || args[i-1] == "kill" || args[i-1] == "rm") {
				return true
			}
		}
	}

	// Check for network/volume removal
	if len(args) >= 2 && (args[0] == "network" || args[0] == "volume") && args[1] == "rm" {
		return true
	}

	return false
}

// getConfirmationTarget extracts a human-readable target from the command args
func (m *CommandExecutionViewModel) getConfirmationTarget(args []string) string {
	if len(args) == 0 {
		return "unknown"
	}

	// For compose commands
	if len(args) >= 2 && args[0] == "compose" {
		// Find the project name
		projectName := ""
		for i, arg := range args {
			if arg == "-p" && i+1 < len(args) {
				projectName = args[i+1]
				break
			}
		}
		if projectName != "" {
			return fmt.Sprintf("project '%s'", projectName)
		}
		return "compose services"
	}

	// For network/volume commands
	if len(args) >= 3 && (args[0] == "network" || args[0] == "volume") {
		return fmt.Sprintf("%s '%s'", args[0], args[2])
	}

	// For image commands
	if len(args) >= 2 && args[0] == "rmi" {
		return fmt.Sprintf("image '%s'", args[len(args)-1])
	}

	// For direct container commands (e.g., "stop container_id")
	if len(args) >= 2 {
		// Try to extract container name or ID
		containerID := args[len(args)-1]
		if len(containerID) > 12 {
			// Likely a container ID, truncate for display
			containerID = containerID[:12]
		}
		return fmt.Sprintf("container %s", containerID)
	}

	return strings.Join(args, " ")
}

// renderConfirmationDialog renders the confirmation prompt
func (m *CommandExecutionViewModel) renderConfirmationDialog(model *Model) string {
	var content strings.Builder

	// Center the dialog
	padding := (model.Height - 10) / 2
	for i := 0; i < padding; i++ {
		content.WriteString("\n")
	}

	// Dialog box
	warningStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("11")).
		Bold(true)

	questionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15"))

	targetStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("14")).
		Bold(true)

	instructionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8"))

	// Build the confirmation message
	content.WriteString(lipgloss.NewStyle().Width(model.width).Align(lipgloss.Center).Render(
		warningStyle.Render("⚠ WARNING"),
	))
	content.WriteString("\n\n")

	// Parse the command for display
	commandDisplay := m.getCommandDisplay(m.pendingArgs)
	content.WriteString(lipgloss.NewStyle().Width(model.width).Align(lipgloss.Center).Render(
		questionStyle.Render(fmt.Sprintf("Are you sure you want to %s", commandDisplay)),
	))
	content.WriteString("\n")
	content.WriteString(lipgloss.NewStyle().Width(model.width).Align(lipgloss.Center).Render(
		targetStyle.Render(m.confirmationTarget),
	))
	content.WriteString("\n\n")
	content.WriteString(lipgloss.NewStyle().Width(model.width).Align(lipgloss.Center).Render(
		instructionStyle.Render("Press 'y' to confirm, 'n' to cancel"),
	))

	return content.String()
}

// getCommandDisplay returns a human-readable command description
func (m *CommandExecutionViewModel) getCommandDisplay(args []string) string {
	if len(args) == 0 {
		return "execute command"
	}

	// Map Docker commands to human-readable descriptions
	commandDescriptions := map[string]string{
		"stop":    "stop",
		"start":   "start",
		"restart": "restart",
		"kill":    "forcefully stop",
		"pause":   "pause",
		"unpause": "unpause",
		"rm":      "remove",
		"rmi":     "remove image",
	}

	// For direct Docker commands
	if desc, exists := commandDescriptions[args[0]]; exists {
		return desc
	}

	// For compose commands
	if len(args) >= 2 && args[0] == "compose" {
		for _, arg := range args {
			if desc, exists := commandDescriptions[arg]; exists {
				return desc
			}
			if arg == "down" {
				return "stop and remove all services in"
			}
			if arg == "up" {
				return "start all services in"
			}
		}
	}

	// For network/volume commands
	if len(args) >= 2 {
		if args[0] == "network" && args[1] == "rm" {
			return "remove network"
		}
		if args[0] == "volume" && args[1] == "rm" {
			return "remove volume"
		}
	}

	return args[0]
}

// HandleConfirmation processes user's confirmation response
func (m *CommandExecutionViewModel) HandleConfirmation(model *Model, confirm bool) tea.Cmd {
	if !m.pendingConfirmation {
		return nil
	}

	if confirm {
		// User confirmed, execute the command
		m.pendingConfirmation = false
		args := m.pendingArgs
		m.pendingArgs = nil
		m.confirmationTarget = ""

		return func() tea.Msg {
			m.cmdString = fmt.Sprintf("docker %s", strings.Join(args, " "))

			// Create the command based on operation
			cmd := exec.Command("docker", args...)

			// Create pipes for stdout and stderr
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				return errorMsg{err: fmt.Errorf("failed to create stdout pipe: %w", err)}
			}
			stderr, err := cmd.StderrPipe()
			if err != nil {
				return errorMsg{err: fmt.Errorf("failed to create stderr pipe: %w", err)}
			}

			// Start the command
			if err := cmd.Start(); err != nil {
				return errorMsg{err: fmt.Errorf("failed to start command: %w", err)}
			}

			// Store the command reference and output channel
			return commandExecStartedMsg{cmd: cmd, stdout: stdout, stderr: stderr}
		}
	}

	// User cancelled, go back to previous view
	m.pendingConfirmation = false
	m.pendingArgs = nil
	m.confirmationTarget = ""
	model.SwitchToPreviousView()
	return nil
}
