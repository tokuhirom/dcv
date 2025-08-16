package ui

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/docker/docker/client"

	"github.com/tokuhirom/dcv/internal/docker"
)

type HelperInjectorViewModel struct {
	container       *docker.Container
	commands        [][]string
	currentStep     int
	totalSteps      int
	output          []string
	scrollY         int
	done            bool
	success         bool
	err             error
	currentCmd      *exec.Cmd
	currentCmdStr   string // Store the current command string for display
	pendingCommands [][]string
	tempDir         string // Store temp directory to clean up later
	helperPath      string // Path where helper will be injected
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
			cmdStr := strings.Join(cmd, " ")
			content.WriteString(fmt.Sprintf("%s%d. %s\n", prefix, i+1, style.Render(cmdStr)))
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

// buildCommands returns the list of commands needed to inject the helper
func (m *HelperInjectorViewModel) buildCommands(container *docker.Container, tempFile string) [][]string {
	if container.IsDind() {
		return [][]string{
			{"docker", "cp", tempFile, fmt.Sprintf("%s:%s", container.HostContainerID(), m.helperPath)},
			{"docker", "exec", container.HostContainerID(), "docker", "cp", m.helperPath, fmt.Sprintf("%s:%s", container.ContainerID(), m.helperPath)},
		}
	} else {
		return [][]string{
			{"docker", "cp", tempFile, fmt.Sprintf("%s:%s", container.ContainerID(), m.helperPath)},
		}
	}
}

// detectArch tries to detect the container's architecture
func (m *HelperInjectorViewModel) detectArch(ctx context.Context, dockerClient *client.Client, container *docker.Container) string {
	// Inspect container to get architecture
	inspect, err := dockerClient.ContainerInspect(ctx, container.ContainerID())
	if err != nil {
		slog.Debug("Failed to inspect container for architecture", "error", err)
		return ""
	}

	// Architecture is in format like "amd64", "arm64", etc.
	if inspect.Platform != "" {
		// Platform might be like "linux/amd64"
		parts := strings.Split(inspect.Platform, "/")
		if len(parts) > 1 {
			return parts[1]
		}
	}

	// Try to get from image config
	if inspect.Config.Labels != nil {
		if arch, ok := inspect.Config.Labels["architecture"]; ok {
			return arch
		}
	}

	// Default detection failed
	return ""
}

// getHelperTempFile creates a temporary file with the helper binary and returns the temp directory and file path
func (m *HelperInjectorViewModel) getHelperTempFile(ctx context.Context, dockerClient *client.Client, container *docker.Container) (string, string, error) {
	// Detect container architecture (default to runtime arch)
	arch := m.detectArch(ctx, dockerClient, container)
	if arch == "" {
		arch = runtime.GOARCH
		slog.Info("Using runtime architecture",
			slog.String("arch", arch))
	} else {
		slog.Info("Detected container architecture",
			slog.String("arch", arch))
	}

	// Get the embedded binary
	binaryData, err := docker.GetHelperBinary(arch)
	if err != nil {
		return "", "", fmt.Errorf("failed to get helper binary: %w", err)
	}

	// Create a temporary file to store the binary
	tempDir, err := os.MkdirTemp("", "dcv-helper-*")
	if err != nil {
		return "", "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	tempFile := filepath.Join(tempDir, "dcv-helper")
	if err := os.WriteFile(tempFile, binaryData, 0755); err != nil {
		_ = os.RemoveAll(tempDir)
		return "", "", fmt.Errorf("failed to write helper binary to temp file: %w", err)
	}

	return tempDir, tempFile, nil
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
	m.helperPath = "/.dcv-helper"

	// Build commands using the moved logic
	if model.dockerSDKClient == nil {
		m.err = fmt.Errorf("docker SDK client not available")
		m.done = true
		return nil
	}

	// Get temporary file path for helper binary
	ctx := context.Background()
	tempDir, tempFile, err := m.getHelperTempFile(ctx, model.dockerSDKClient, container)
	if err != nil {
		m.err = err
		m.done = true
		return nil
	}
	m.tempDir = tempDir // Store for cleanup later
	slog.Info("Wrote temporary file",
		slog.String("tempDir", tempDir),
		slog.String("tempFile", tempFile))

	m.commands = m.buildCommands(container, tempFile)
	m.totalSteps = len(m.commands)
	m.pendingCommands = m.commands

	// Start executing the first command
	return m.executeNextCommand()
}

func (m *HelperInjectorViewModel) executeNextCommand() tea.Cmd {
	if len(m.pendingCommands) == 0 {
		// All commands executed
		m.done = true
		m.success = true
		return nil
	}

	m.currentStep++
	cmdArgs := m.pendingCommands[0]
	m.pendingCommands = m.pendingCommands[1:]
	m.currentCmdStr = strings.Join(cmdArgs, " ") // Store for display

	return func() tea.Msg {
		// Use the command arguments directly
		if len(cmdArgs) == 0 {
			return helperInjectorCompleteMsg{success: false, err: fmt.Errorf("empty command")}
		}

		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)

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

	// Clean up temp directory if it exists
	if m.tempDir != "" {
		if err := os.RemoveAll(m.tempDir); err != nil {
			slog.Debug("Failed to clean up temp directory", "error", err, "path", m.tempDir)
		}
		m.tempDir = ""
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

	// Clean up temp directory on completion
	if m.tempDir != "" {
		if cleanupErr := os.RemoveAll(m.tempDir); cleanupErr != nil {
			slog.Debug("Failed to clean up temp directory", "error", cleanupErr, "path", m.tempDir)
		}
		m.tempDir = ""
	}
}
