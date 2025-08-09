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
	m.currentPath = "/"
	m.pathHistory = []string{"/"}
	model.SwitchView(FileBrowserView)
	model.loading = true
	return loadContainerFiles(model.dockerClient, containerID, "/")
}

func (m *FileBrowserViewModel) HandleBack(model *Model) tea.Cmd {
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
		m.currentPath = filepath.Dir(m.currentPath)
		if len(m.pathHistory) > 1 {
			m.pathHistory = m.pathHistory[:len(m.pathHistory)-1]
		}
		model.loading = true
		m.selectedFile = 0
		return loadContainerFiles(model.dockerClient, m.browsingContainerID, m.currentPath)
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
				m.currentPath = filepath.Dir(m.currentPath)
				if len(m.pathHistory) > 1 {
					m.pathHistory = m.pathHistory[:len(m.pathHistory)-1]
				}
			}
			model.loading = true
			m.selectedFile = 0
			return loadContainerFiles(model.dockerClient, m.browsingContainerID, m.currentPath)
		}

		newPath := filepath.Join(m.currentPath, file.Name)

		if file.IsDir {
			// Navigate into directory
			m.currentPath = newPath
			m.pathHistory = append(m.pathHistory, newPath)
			model.loading = true
			m.selectedFile = 0
			return loadContainerFiles(model.dockerClient, m.browsingContainerID, newPath)
		} else {
			// View file content
			return model.fileContentViewModel.Load(model, m.browsingContainerID, m.browsingContainerName, newPath)
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
	return loadContainerFiles(model.dockerClient, m.browsingContainerID, m.currentPath)
}

func (m *FileBrowserViewModel) Title() string {
	return fmt.Sprintf("File Browser: %s [%s]", m.browsingContainerName, m.currentPath)
}
