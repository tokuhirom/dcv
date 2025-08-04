package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

type FileBrowserViewModel struct {
}

// renderFileBrowser renders the file browser view
func (m *Model) renderFileBrowser(availableHeight int) string {
	var content strings.Builder

	if len(m.containerFiles) == 0 {
		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		content.WriteString("\n")
		content.WriteString(dimStyle.Render("No files found or directory is empty"))
		content.WriteString("\n")
		return content.String()
	}

	// Add spacing
	content.WriteString("\n")

	// Create table
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == m.selectedFile {
				return selectedStyle
			}
			return normalStyle
		}).
		Headers("PERMISSIONS", "NAME").
		Height(availableHeight - 3). // Reserve 3 lines for path display
		Width(m.width).
		Offset(m.selectedFile)

	// Define styles for different file types
	dirStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("33"))
	linkStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("51"))

	// Add rows
	for _, file := range m.containerFiles {
		// Style the name based on file type
		name := file.GetDisplayName()
		if file.IsDir {
			name = dirStyle.Render(name)
		} else if file.LinkTarget != "" {
			name = linkStyle.Render(name)
		}

		t.Row(file.Permissions, name)
	}

	content.WriteString(t.Render())
	content.WriteString("\n\n")

	// Show current path at bottom
	pathStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true)
	content.WriteString(pathStyle.Render(fmt.Sprintf("Path: %s", m.currentPath)))
	content.WriteString("\n")

	return content.String()
}

func (m *FileBrowserViewModel) Load(model *Model, containerID, containerName string) tea.Cmd {
	model.browsingContainerID = containerID
	model.browsingContainerName = containerName
	model.currentPath = "/"
	model.pathHistory = []string{"/"}
	model.currentView = FileBrowserView
	model.loading = true
	return loadContainerFiles(model.dockerClient, containerID, "/")
}

func (m *FileBrowserViewModel) HandleBack(model *Model) tea.Cmd {
	// TODO: care about the previous view...

	// Check where we came from based on the container name prefix
	for _, container := range model.dockerContainerListViewModel.dockerContainers {
		if container.ID == model.browsingContainerID {
			model.currentView = DockerContainerListView
			return loadDockerContainers(model.dockerClient, model.dockerContainerListViewModel.showAll)
		}
	}
	// Default to compose process list
	model.currentView = ComposeProcessListView
	return loadProcesses(model.dockerClient, model.projectName, model.composeProcessListViewModel.showAll)
}
