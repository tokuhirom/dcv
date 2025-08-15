package ui

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
)

// containerFilesLoadedMsg contains the loaded container files
type containerFilesLoadedMsg struct {
	files []models.ContainerFile
	err   error
}

type FileBrowserViewModel struct {
	TableViewModel
	containerFiles    []models.ContainerFile
	currentPath       string
	browsingContainer *docker.Container // The container we're browsing
	pathHistory       []string
}

// Update handles messages for the file browser view
func (m *FileBrowserViewModel) Update(model *Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case containerFilesLoadedMsg:
		model.loading = false
		if msg.err != nil {
			model.err = msg.err
			return model, nil
		} else {
			model.err = nil
		}

		m.Loaded(model, msg.files)
		return model, nil
	default:
		return model, nil
	}
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
		{Title: "PERMISSIONS", Width: 15}, // Fixed width for permissions display
		{Title: "SIZE", Width: 10},        // Fixed width for size display
		{Title: "NAME", Width: -1},
	}

	return m.RenderTable(model, columns, availableHeight, func(row, col int) lipgloss.Style {
		if row == m.Cursor {
			return tableSelectedCellStyle
		}
		return tableNormalCellStyle
	})
}

// buildRows builds the table rows from container files
func (m *FileBrowserViewModel) buildRows() []table.Row {
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

		rows = append(rows, table.Row{file.Permissions, file.GetSizeString(), name})
	}
	return rows
}

func (m *FileBrowserViewModel) LoadContainer(model *Model, container *docker.Container) tea.Cmd {
	m.browsingContainer = container
	m.pathHistory = []string{}
	m.pushHistory("/")
	model.SwitchView(FileBrowserView)
	return m.DoLoad(model)
}

func (m *FileBrowserViewModel) HandleBack(model *Model) tea.Cmd {
	// Try to go back in path history
	if m.popHistory() {
		m.Cursor = 0
		return m.DoLoad(model)
	}
	// If no more history, go back to the previous view
	model.SwitchToPreviousView()
	return nil
}

func (m *FileBrowserViewModel) HandleUp(model *Model) tea.Cmd {
	return m.TableViewModel.HandleUp(model)
}

func (m *FileBrowserViewModel) HandleDown(model *Model) tea.Cmd {
	return m.TableViewModel.HandleDown(model)
}

func (m *FileBrowserViewModel) HandleGoToParentDirectory(model *Model) tea.Cmd {
	// Go up one directory
	if m.currentPath != "/" {
		parentPath := filepath.Dir(m.currentPath)
		m.pushHistory(parentPath)
		m.Cursor = 0
		return m.DoLoad(model)
	}
	return nil
}

func (m *FileBrowserViewModel) HandleOpenFileOrDirectory(model *Model) tea.Cmd {
	if m.Cursor < len(m.containerFiles) {
		file := m.containerFiles[m.Cursor]

		if file.Name == "." {
			return nil
		}

		if file.Name == ".." {
			// Go up one directory
			if m.currentPath != "/" {
				parentPath := filepath.Dir(m.currentPath)
				m.pushHistory(parentPath)
				m.Cursor = 0
				return m.DoLoad(model)
			}
			return nil
		}

		newPath := filepath.Join(m.currentPath, file.Name)

		if file.IsDir {
			// Navigate into directory
			m.pushHistory(newPath)
			m.Cursor = 0
			return m.DoLoad(model)
		} else {
			// View file content
			return model.fileContentViewModel.LoadContainer(model, m.browsingContainer, newPath)
		}
	}
	return nil
}

func (m *FileBrowserViewModel) HandleInjectHelper(model *Model) tea.Cmd {
	if m.browsingContainer == nil {
		return nil
	}

	model.loading = true
	return func() tea.Msg {
		// Manually inject the helper binary
		ctx := context.Background()
		_, err := model.fileOperations.InjectHelper(ctx, m.browsingContainer.ContainerID())

		if err != nil {
			return containerFilesLoadedMsg{
				err: fmt.Errorf("failed to inject helper: %w", err),
			}
		}

		// After successful injection, reload the current directory
		files, err := model.fileOperations.ListFiles(ctx, m.browsingContainer, m.currentPath)

		return containerFilesLoadedMsg{
			files: files,
			err:   err,
		}
	}
}

func (m *FileBrowserViewModel) Loaded(model *Model, files []models.ContainerFile) {
	m.containerFiles = files
	m.SetRows(m.buildRows(), model.ViewHeight())
}

func (m *FileBrowserViewModel) DoLoad(model *Model) tea.Cmd {
	model.loading = true
	return func() tea.Msg {
		// Use FileOperations for multi-strategy file listing
		ctx := context.Background()

		files, err := model.fileOperations.ListFiles(ctx, m.browsingContainer, m.currentPath)

		return containerFilesLoadedMsg{
			files: files,
			err:   err,
		}
	}
}

func (m *FileBrowserViewModel) Title() string {
	if m.browsingContainer != nil {
		return fmt.Sprintf("File Browser: %s [%s]", m.browsingContainer.Title(), m.currentPath)
	}
	return "File Browser"
}
