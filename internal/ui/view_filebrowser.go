package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderFileBrowser renders the file browser view
func (m *Model) renderFileBrowser() string {
	var content strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	title := fmt.Sprintf("File Browser: %s [%s]", m.browsingContainerName, m.currentPath)
	content.WriteString(titleStyle.Render(title))
	content.WriteString("\n\n")

	if m.loading {
		return content.String() + "Loading files..."
	}

	if m.err != nil {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		return content.String() + errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	if len(m.containerFiles) == 0 {
		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		return content.String() + dimStyle.Render("No files found or directory is empty")
	}

	// Table headers
	headerStyle := lipgloss.NewStyle().Bold(true).Underline(true).Foreground(lipgloss.Color("86"))
	headers := []string{"PERMISSIONS", "NAME"}
	colWidths := []int{11, 60}

	// Render headers
	for i, header := range headers {
		content.WriteString(headerStyle.Render(padRight(header, colWidths[i])))
		if i < len(headers)-1 {
			content.WriteString(" ")
		}
	}
	content.WriteString("\n")

	// Render files
	normalStyle := lipgloss.NewStyle()
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220")).Background(lipgloss.Color("235"))
	dirStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("33"))
	linkStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("51"))

	for i, file := range m.containerFiles {
		style := normalStyle
		nameStyle := normalStyle

		if file.IsDir {
			nameStyle = dirStyle
		} else if file.LinkTarget != "" {
			nameStyle = linkStyle
		}

		if i == m.selectedFile {
			style = selectedStyle
			nameStyle = selectedStyle
		}

		// Format row data
		permissions := style.Render(padRight(file.Permissions, colWidths[0]))
		name := nameStyle.Render(padRight(file.GetDisplayName(), colWidths[1]))

		// Render row
		content.WriteString(permissions)
		content.WriteString(" ")
		content.WriteString(name)
		content.WriteString("\n")
	}

	// Show current path at bottom
	pathStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true)
	content.WriteString("\n")
	content.WriteString(pathStyle.Render(fmt.Sprintf("Path: %s", m.currentPath)))

	return content.String()
}

// renderFileContent renders the file content view
func (m *Model) renderFileContent() string {
	var content strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	title := fmt.Sprintf("File: %s [%s]", filepath.Base(m.fileContentPath), m.browsingContainerName)
	content.WriteString(titleStyle.Render(title))
	content.WriteString("\n\n")

	if m.loading {
		return content.String() + "Loading file content..."
	}

	if m.err != nil {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		return content.String() + errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	// File content with line numbers
	lines := strings.Split(m.fileContent, "\n")
	viewHeight := m.height - 5
	startIdx := m.fileScrollY
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
