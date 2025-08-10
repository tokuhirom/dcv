package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tokuhirom/dcv/internal/models"
)

type FileBrowserViewModel struct {
	containerFiles        []models.ContainerFile
	selectedFile          int
	currentPath           string
	browsingContainerID   string
	browsingContainerName string
	pathHistory           []string
	isDind                bool   // Whether we're browsing a DinD container
	hostContainerID       string // Host container ID for DinD operations
}

// pushHistory adds a new path to the history
func (m *FileBrowserViewModel) pushHistory(path string) {
	m.pathHistory = append(m.pathHistory, path)
	m.currentPath = path
}

// popHistory removes the last path from history and returns to the previous one
// Returns false if there's no more history to pop
func (m *FileBrowserViewModel) popHistory() bool {
	if len(m.pathHistory) <= 1 {
		return false
	}
	// Remove current path from history
	m.pathHistory = m.pathHistory[:len(m.pathHistory)-1]
	// Set current path to the new last item
	m.currentPath = m.pathHistory[len(m.pathHistory)-1]
	return true
}

// render renders the file browser view
func (m *FileBrowserViewModel) render(model *Model, availableHeight int) string {
	if len(m.containerFiles) == 0 {
		var content strings.Builder
		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		content.WriteString("\n")
		content.WriteString(dimStyle.Render("No files found or directory is empty"))
		content.WriteString("\n")
		return content.String()
	}

	// Create table
	columns := []table.Column{
		{Title: "PERMISSIONS", Width: 15},
		{Title: "NAME", Width: model.width - 20},
	}

	// Define styles for different file types
	dirStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("33"))
	linkStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("51"))

	rows := make([]table.Row, 0, len(m.containerFiles))
	// Add rows
	for _, file := range m.containerFiles {
		// Style the name based on file type
		name := file.GetDisplayName()
		if file.IsDir {
			name = dirStyle.Render(name)
		} else if file.LinkTarget != "" {
			name = linkStyle.Render(name)
		}

		rows = append(rows, table.Row{file.Permissions, name})
	}

	return RenderTable(columns, rows, availableHeight-3, m.selectedFile)
}

func (m *FileBrowserViewModel) Load(model *Model, containerID, containerName string) tea.Cmd {
	m.browsingContainerID = containerID
	m.browsingContainerName = containerName
	m.pathHistory = []string{}
	m.pushHistory("/")
	m.isDind = false
	m.hostContainerID = ""
	model.SwitchView(FileBrowserView)
	return m.DoLoad(model)
}

func (m *FileBrowserViewModel) LoadDind(model *Model, hostContainerID, containerID, containerName string) tea.Cmd {
	m.browsingContainerID = containerID
	m.browsingContainerName = containerName
	m.pathHistory = []string{}
	m.pushHistory("/")
	m.isDind = true
	m.hostContainerID = hostContainerID
	model.SwitchView(FileBrowserView)
	return m.DoLoad(model)
}

func (m *FileBrowserViewModel) HandleBack(model *Model) tea.Cmd {
	// Try to go back in path history
	if m.popHistory() {
		m.selectedFile = 0
		return m.DoLoad(model)
	}
	// If no more history, go back to the previous view
	model.SwitchToPreviousView()
	return nil
}

func (m *FileBrowserViewModel) HandleUp() tea.Cmd {
	if m.selectedFile > 0 {
		m.selectedFile--
	}
	return nil
}

func (m *FileBrowserViewModel) HandleDown() tea.Cmd {
	if m.selectedFile < len(m.containerFiles)-1 {
		m.selectedFile++
	}
	return nil
}

func (m *FileBrowserViewModel) HandleGoToParentDirectory(model *Model) tea.Cmd {
	// Go up one directory
	if m.currentPath != "/" {
		parentPath := filepath.Dir(m.currentPath)
		m.pushHistory(parentPath)
		m.selectedFile = 0
		return m.DoLoad(model)
	}
	return nil
}

func (m *FileBrowserViewModel) HandleOpenFileOrDirectory(model *Model) tea.Cmd {
	if m.selectedFile < len(m.containerFiles) {
		file := m.containerFiles[m.selectedFile]

		if file.Name == "." {
			return nil
		}

		if file.Name == ".." {
			// Go up one directory
			if m.currentPath != "/" {
				parentPath := filepath.Dir(m.currentPath)
				m.pushHistory(parentPath)
				m.selectedFile = 0
				return m.DoLoad(model)
			}
			return nil
		}

		newPath := filepath.Join(m.currentPath, file.Name)

		if file.IsDir {
			// Navigate into directory
			m.pushHistory(newPath)
			m.selectedFile = 0
			return m.DoLoad(model)
		} else {
			// View file content
			if m.isDind {
				return model.fileContentViewModel.LoadDind(model, m.hostContainerID, m.browsingContainerID, m.browsingContainerName, newPath)
			} else {
				return model.fileContentViewModel.Load(model, m.browsingContainerID, m.browsingContainerName, newPath)
			}
		}
	}
	return nil
}

func (m *FileBrowserViewModel) Loaded(files []models.ContainerFile) {
	m.containerFiles = files
	if len(m.containerFiles) > 0 && m.selectedFile >= len(m.containerFiles) {
		m.selectedFile = 0
	}
}

func (m *FileBrowserViewModel) DoLoad(model *Model) tea.Cmd {
	model.loading = true
	return func() tea.Msg {
		var files []models.ContainerFile
		var err error

		if m.isDind {
			// Use DindClient for nested container file listing
			dindClient := model.dockerClient.Dind(m.hostContainerID)
			files, err = dindClient.ListContainerFiles(m.browsingContainerID, m.currentPath)
		} else {
			// Use regular client for normal containers
			files, err = model.dockerClient.ListContainerFiles(m.browsingContainerID, m.currentPath)
		}

		return containerFilesLoadedMsg{
			files: files,
			err:   err,
		}
	}
}

func (m *FileBrowserViewModel) Title() string {
	return fmt.Sprintf("File Browser: %s [%s]", m.browsingContainerName, m.currentPath)
}
