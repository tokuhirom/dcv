package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderFileBrowser renders the file browser view
func (m *Model) renderFileBrowser() string {
	var content strings.Builder

	if len(m.containerFiles) == 0 {
		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		content.WriteString(dimStyle.Render("No files found or directory is empty"))
		return content.String()
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
