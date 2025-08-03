package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderFileContent renders the file content view
func (m *Model) renderFileContent(availableHeight int) string {
	var content strings.Builder

	if m.err != nil {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		return content.String() + errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	// File content with line numbers
	lines := strings.Split(m.fileContent, "\n")
	viewHeight := availableHeight
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
