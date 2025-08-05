package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FileContentViewModel struct {
	containerName string
	content       string
	contentPath   string
	scrollY       int
}

// render renders the file content view
func (m *FileContentViewModel) render(model *Model, availableHeight int) string {
	var content strings.Builder

	if model.err != nil {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		return content.String() + errorStyle.Render(fmt.Sprintf("Error: %v", model.err))
	}

	// File content with line numbers
	lines := strings.Split(m.content, "\n")
	viewHeight := availableHeight
	startIdx := m.scrollY
	endIdx := startIdx + viewHeight

	if endIdx > len(lines) {
		endIdx = len(lines)
	}

	lineNumStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	contentStyle := lipgloss.NewStyle()

	if len(lines) == 0 {
		content.WriteString("(empty file)\n")
	} else if startIdx < len(lines) {
		for i := startIdx; i < endIdx; i++ {
			lineNum := lineNumStyle.Render(fmt.Sprintf("%4d ", i+1))
			lineContent := contentStyle.Render(lines[i])
			content.WriteString(lineNum + lineContent + "\n")
		}
	}

	// Fill remaining space
	linesShown := endIdx - startIdx
	for i := linesShown; i < viewHeight; i++ {
		content.WriteString("\n")
	}

	// Show position indicator
	if len(lines) > viewHeight {
		posStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		position := fmt.Sprintf("Lines %d-%d of %d", startIdx+1, endIdx, len(lines))
		content.WriteString(posStyle.Render(position))
	}

	return content.String()
}

func (m *FileContentViewModel) Load(model *Model, containerID, path string) tea.Cmd {
	// no one calls this directly, it's used by the Model
	model.currentView = FileContentView
	model.loading = true
	m.scrollY = 0
	return loadFileContent(model.dockerClient, containerID, path)
}

func (m *FileContentViewModel) HandleScrollUp() tea.Cmd {
	if m.scrollY > 0 {
		m.scrollY--
	}
	return nil
}

func (m *FileContentViewModel) HandleScrollDown(height int) tea.Cmd {
	lines := strings.Split(m.content, "\n")
	maxScroll := len(lines) - (height - 5)
	if m.scrollY < maxScroll && maxScroll > 0 {
		m.scrollY++
	}
	return nil
}

func (m *FileContentViewModel) HandleGoToStart() tea.Cmd {
	m.scrollY = 0
	return nil
}

func (m *FileContentViewModel) HandleGoToEnd(height int) tea.Cmd {
	lines := strings.Split(m.content, "\n")
	maxScroll := len(lines) - (height - 5)
	if maxScroll > 0 {
		m.scrollY = maxScroll
	}
	return nil
}

func (m *FileContentViewModel) HandleBack(model *Model) tea.Cmd {
	model.currentView = FileBrowserView
	m.content = ""
	m.contentPath = ""
	m.scrollY = 0
	return nil
}

func (m *FileContentViewModel) Title() string {
	return fmt.Sprintf("File: %s [%s]", filepath.Base(m.contentPath), m.containerName)
}
