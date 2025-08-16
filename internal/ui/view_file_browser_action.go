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
			return m.handleCopyToLocal(model, f, c)
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

// handleCopyToLocal handles copying a file from container to local machine
func (m *FileBrowserActionViewModel) handleCopyToLocal(model *Model, file *models.ContainerFile, container *docker.Container) tea.Cmd {
	// Build the source path
	sourcePath := filepath.Join(m.containerPath, file.Name)

	// Get the home directory for destination
	homeDir, err := os.UserHomeDir()
	if err != nil {
		model.err = fmt.Errorf("failed to get home directory: %w", err)
		return nil
	}

	// Create a Downloads directory if it doesn't exist
	downloadDir := filepath.Join(homeDir, "Downloads", "dcv")
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		model.err = fmt.Errorf("failed to create download directory: %w", err)
		return nil
	}

	// Generate destination path
	destPath := filepath.Join(downloadDir, file.Name)

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
	if m.selectedAction > 0 {
		m.selectedAction--
	}
}

// HandleDown moves selection down
func (m *FileBrowserActionViewModel) HandleDown() {
	if m.selectedAction < len(m.actions)-1 {
		m.selectedAction++
	}
}

// HandleSelect executes the selected action
func (m *FileBrowserActionViewModel) HandleSelect(model *Model) tea.Cmd {
	if m.selectedAction < len(m.actions) {
		action := m.actions[m.selectedAction]
		model.SwitchToPreviousView() // Go back to file browser
		return action.Handler(model, m.targetFile, m.targetContainer)
	}
	return nil
}

// HandleBack returns to the file browser
func (m *FileBrowserActionViewModel) HandleBack(model *Model) tea.Cmd {
	model.SwitchToPreviousView()
	return nil
}
