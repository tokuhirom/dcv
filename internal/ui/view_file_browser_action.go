package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
)

// FileBrowserAction represents a file operation
type FileBrowserAction struct {
	Key         string
	Name        string
	Description string
	Handler     func(m *Model, file *models.ContainerFile, container *docker.Container) tea.Cmd
}

// FileBrowserActionViewModel manages the file browser action selection view
type FileBrowserActionViewModel struct {
	actions         []FileBrowserAction
	selectedAction  int
	targetFile      *models.ContainerFile
	targetContainer *docker.Container
	containerPath   string

	// Input mode for destination path
	inputMode      bool
	inputBuffer    string
	inputCursorPos int
	inputPrompt    string
}

// Initialize sets up the action view with available commands for a file
func (m *FileBrowserActionViewModel) Initialize(file *models.ContainerFile, container *docker.Container, containerPath string) {
	m.targetFile = file
	m.targetContainer = container
	m.containerPath = containerPath
	m.selectedAction = 0

	// Define available actions
	m.actions = []FileBrowserAction{}

	// Copy to local machine
	m.actions = append(m.actions, FileBrowserAction{
		Key:         "C",
		Name:        "Copy to Local",
		Description: "Copy file/directory to local machine",
		Handler: func(model *Model, f *models.ContainerFile, c *docker.Container) tea.Cmd {
			// Start input mode to get destination path
			m.startInputMode(f)
			return nil
		},
	})

	// View file (if it's a file)
	if !file.IsDir {
		m.actions = append(m.actions, FileBrowserAction{
			Key:         "V",
			Name:        "View File",
			Description: "View file contents",
			Handler: func(model *Model, f *models.ContainerFile, c *docker.Container) tea.Cmd {
				// Use the existing file content viewer
				fullPath := filepath.Join(containerPath, f.Name)
				return model.fileContentViewModel.LoadContainer(model, c, fullPath)
			},
		})
	}

	// Execute command in directory (if it's a directory)
	if file.IsDir {
		m.actions = append(m.actions, FileBrowserAction{
			Key:         "E",
			Name:        "Execute Command",
			Description: "Execute command in this directory",
			Handler: func(model *Model, f *models.ContainerFile, c *docker.Container) tea.Cmd {
				fullPath := filepath.Join(containerPath, f.Name)
				return m.handleExecuteInDirectory(model, fullPath, c)
			},
		})
	}

	// Delete file/directory
	m.actions = append(m.actions, FileBrowserAction{
		Key:         "D",
		Name:        "Delete",
		Description: "Delete file or directory",
		Handler: func(model *Model, f *models.ContainerFile, c *docker.Container) tea.Cmd {
			return m.handleDelete(model, f, c)
		},
	})
}

// startInputMode starts the input mode for destination path
func (m *FileBrowserActionViewModel) startInputMode(file *models.ContainerFile) {
	m.inputMode = true
	m.inputPrompt = fmt.Sprintf("Enter destination path for '%s': ", file.Name)

	// Set default path
	homeDir, err := os.UserHomeDir()
	if err == nil {
		m.inputBuffer = filepath.Join(homeDir, "Downloads", file.Name)
	} else {
		m.inputBuffer = filepath.Join("/tmp", file.Name)
	}
	m.inputCursorPos = len(m.inputBuffer)
}

// handleCopyToLocal handles copying a file from container to local machine
func (m *FileBrowserActionViewModel) handleCopyToLocal(model *Model, destPath string) tea.Cmd {
	file := m.targetFile
	container := m.targetContainer

	// Build the source path
	sourcePath := filepath.Join(m.containerPath, file.Name)

	// Expand tilde if present
	if strings.HasPrefix(destPath, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			model.err = fmt.Errorf("failed to get home directory: %w", err)
			return nil
		}
		destPath = filepath.Join(homeDir, destPath[2:])
	}

	// Create parent directory if it doesn't exist
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		model.err = fmt.Errorf("failed to create destination directory: %w", err)
		return nil
	}

	// Build the docker cp command
	var args []string
	if container.IsDind() {
		// For dind containers: docker exec <container> docker cp <inner-container>:path -
		// Then we need to extract it locally
		args = append(container.OperationArgs("cp"),
			fmt.Sprintf("%s:%s", container.ContainerID(), sourcePath), destPath)
	} else {
		// Regular container: docker cp <container>:path dest
		args = []string{"cp", fmt.Sprintf("%s:%s", container.ContainerID(), sourcePath), destPath}
	}

	// Execute the command
	return model.commandExecutionViewModel.ExecuteCommand(model, false, args...)
}

// handleDelete handles deleting a file or directory
func (m *FileBrowserActionViewModel) handleDelete(model *Model, file *models.ContainerFile, container *docker.Container) tea.Cmd {
	// Build the full path
	fullPath := filepath.Join(m.containerPath, file.Name)

	// Build the rm command
	var rmCmd string
	if file.IsDir {
		rmCmd = "rm -rf"
	} else {
		rmCmd = "rm -f"
	}

	// Execute the command
	args := append(container.OperationArgs("exec"),
		container.ContainerID(), "sh", "-c", fmt.Sprintf("%s %q", rmCmd, fullPath))

	// Show confirmation dialog for delete (true = aggressive operation)
	return model.commandExecutionViewModel.ExecuteCommand(model, true, args...)
}

// handleExecuteInDirectory handles executing a command in a specific directory
func (m *FileBrowserActionViewModel) handleExecuteInDirectory(model *Model, dirPath string, container *docker.Container) tea.Cmd {
	// Build args for shell execution in the specific directory
	// We need to add the -w flag before the shell command
	var args []string
	if container.IsDind() {
		// For DinD: docker exec -it <host> docker exec -it -w <dir> <container> /bin/sh
		args = []string{"exec", "-it", container.HostContainerID(), "docker", "exec", "-it", "-w", dirPath, container.ContainerID(), "/bin/sh"}
	} else {
		// For regular: docker exec -it -w <dir> <container> /bin/sh
		args = []string{"exec", "-it", "-w", dirPath, container.ContainerID(), "/bin/sh"}
	}

	return func() tea.Msg {
		return launchShellMsg{
			container: container,
			args:      args,
			shell:     "/bin/sh",
		}
	}
}

// render renders the file browser action menu
func (m *FileBrowserActionViewModel) render(model *Model) string {
	var s strings.Builder

	// If in input mode, render input prompt
	if m.inputMode {
		return m.renderInputMode()
	}

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86"))

	s.WriteString(titleStyle.Render("File Actions"))
	s.WriteString("\n\n")

	// File info
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	s.WriteString(infoStyle.Render(fmt.Sprintf("File: %s", m.targetFile.Name)))
	s.WriteString("\n")
	s.WriteString(infoStyle.Render(fmt.Sprintf("Path: %s", m.containerPath)))
	s.WriteString("\n")
	s.WriteString(infoStyle.Render(fmt.Sprintf("Type: %s", m.getFileType())))
	s.WriteString("\n\n")

	// Actions list
	for i, action := range m.actions {
		prefix := "  "
		if i == m.selectedAction {
			prefix = "> "
		}

		keyStyle := lipgloss.NewStyle().
			Bold(i == m.selectedAction).
			Foreground(lipgloss.Color("220"))
		nameStyle := lipgloss.NewStyle().
			Bold(i == m.selectedAction)
		descStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

		s.WriteString(prefix)
		s.WriteString(keyStyle.Render(fmt.Sprintf("[%s]", action.Key)))
		s.WriteString(" ")
		s.WriteString(nameStyle.Render(action.Name))
		s.WriteString(" - ")
		s.WriteString(descStyle.Render(action.Description))
		s.WriteString("\n")
	}

	s.WriteString("\n")
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	s.WriteString(helpStyle.Render("Use ↑/↓ to navigate, Enter to select, Esc to cancel"))

	return s.String()
}

// renderInputMode renders the input prompt for destination path
func (m *FileBrowserActionViewModel) renderInputMode() string {
	var s strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86"))

	s.WriteString(titleStyle.Render("Copy File to Local"))
	s.WriteString("\n\n")

	// File info
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	s.WriteString(infoStyle.Render(fmt.Sprintf("Source: %s/%s", m.containerPath, m.targetFile.Name)))
	s.WriteString("\n\n")

	// Input prompt
	promptStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	s.WriteString(promptStyle.Render(m.inputPrompt))
	s.WriteString("\n")

	// Input field with cursor
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("86")).
		Padding(0, 1)

	// Add cursor to the input
	inputWithCursor := m.inputBuffer
	if m.inputCursorPos >= 0 && m.inputCursorPos <= len(m.inputBuffer) {
		inputWithCursor = m.inputBuffer[:m.inputCursorPos] + "█" + m.inputBuffer[m.inputCursorPos:]
	}

	s.WriteString(inputStyle.Render(inputWithCursor))
	s.WriteString("\n\n")

	// Help text
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	s.WriteString(helpStyle.Render("Press Enter to confirm, Esc to cancel"))

	return s.String()
}

func (m *FileBrowserActionViewModel) getFileType() string {
	if m.targetFile.IsDir {
		return "Directory"
	} else if m.targetFile.LinkTarget != "" {
		return fmt.Sprintf("Symlink -> %s", m.targetFile.LinkTarget)
	} else {
		return "File"
	}
}

// HandleUp moves selection up
func (m *FileBrowserActionViewModel) HandleUp() {
	if !m.inputMode && m.selectedAction > 0 {
		m.selectedAction--
	}
}

// HandleDown moves selection down
func (m *FileBrowserActionViewModel) HandleDown() {
	if !m.inputMode && m.selectedAction < len(m.actions)-1 {
		m.selectedAction++
	}
}

// HandleSelect executes the selected action
func (m *FileBrowserActionViewModel) HandleSelect(model *Model) tea.Cmd {
	// If in input mode, confirm the input
	if m.inputMode {
		destPath := strings.TrimSpace(m.inputBuffer)
		if destPath != "" {
			m.inputMode = false
			model.SwitchToPreviousView() // Go back to file browser
			return m.handleCopyToLocal(model, destPath)
		}
		return nil
	}

	// Otherwise, execute the selected action
	if m.selectedAction < len(m.actions) {
		action := m.actions[m.selectedAction]
		// Don't switch view yet if the action starts input mode
		return action.Handler(model, m.targetFile, m.targetContainer)
	}
	return nil
}

// HandleBack returns to the file browser or cancels input
func (m *FileBrowserActionViewModel) HandleBack(model *Model) tea.Cmd {
	if m.inputMode {
		// Cancel input mode and return to action menu
		m.inputMode = false
		m.inputBuffer = ""
		m.inputCursorPos = 0
		return nil
	}
	// Return to file browser
	model.SwitchToPreviousView()
	return nil
}

// HandleInput handles keyboard input in input mode
func (m *FileBrowserActionViewModel) HandleInput(model *Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if !m.inputMode {
		return model, nil
	}

	switch msg.Type {
	case tea.KeyEnter:
		// Confirm input
		return model, m.HandleSelect(model)

	case tea.KeyEsc:
		// Cancel input
		return model, m.HandleBack(model)

	case tea.KeyBackspace, tea.KeyCtrlH:
		if len(m.inputBuffer) > 0 && m.inputCursorPos > 0 {
			m.inputBuffer = m.inputBuffer[:m.inputCursorPos-1] + m.inputBuffer[m.inputCursorPos:]
			m.inputCursorPos--
		}

	case tea.KeyLeft, tea.KeyCtrlB:
		if m.inputCursorPos > 0 {
			m.inputCursorPos--
		}

	case tea.KeyRight, tea.KeyCtrlF:
		if m.inputCursorPos < len(m.inputBuffer) {
			m.inputCursorPos++
		}

	case tea.KeyHome, tea.KeyCtrlA:
		m.inputCursorPos = 0

	case tea.KeyEnd, tea.KeyCtrlE:
		m.inputCursorPos = len(m.inputBuffer)

	case tea.KeyDelete:
		if m.inputCursorPos < len(m.inputBuffer) {
			m.inputBuffer = m.inputBuffer[:m.inputCursorPos] + m.inputBuffer[m.inputCursorPos+1:]
		}

	default:
		switch msg.Type {
		case tea.KeyRunes:
			m.inputBuffer = m.inputBuffer[:m.inputCursorPos] + msg.String() + m.inputBuffer[m.inputCursorPos:]
			m.inputCursorPos += len(msg.String())
		case tea.KeySpace:
			m.inputBuffer = m.inputBuffer[:m.inputCursorPos] + " " + m.inputBuffer[m.inputCursorPos:]
			m.inputCursorPos++
		}
	}

	return model, nil
}
