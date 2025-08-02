package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderInspectView renders the container inspect view
func (m *Model) renderInspectView() string {
	var content strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	title := fmt.Sprintf("Container Inspect: %s", m.inspectContainerID)
	content.WriteString(titleStyle.Render(title))
	content.WriteString("\n\n")

	if m.loading {
		return content.String() + "Loading inspect data..."
	}

	if m.err != nil {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		return content.String() + errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	// Split content into lines for scrolling
	lines := strings.Split(m.inspectContent, "\n")
	viewHeight := m.height - 5
	startIdx := m.inspectScrollY
	endIdx := startIdx + viewHeight

	if endIdx > len(lines) {
		endIdx = len(lines)
	}

	// Render the JSON content with syntax highlighting
	lineNumStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	jsonKeyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("33"))
	jsonValueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("76"))
	jsonBraceStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	if len(lines) == 0 {
		content.WriteString("No inspect data available\n")
	} else if startIdx < len(lines) {
		for i := startIdx; i < endIdx; i++ {
			line := lines[i]
			lineNum := lineNumStyle.Render(fmt.Sprintf("%4d ", i+1))
			
			// Simple JSON syntax highlighting
			highlightedLine := line
			if strings.Contains(line, "\":") {
				// This line likely contains a key-value pair
				parts := strings.SplitN(line, "\":", 2)
				if len(parts) == 2 {
					// Extract key part
					keyStart := strings.LastIndex(parts[0], "\"")
					if keyStart >= 0 {
						indent := parts[0][:keyStart]
						key := parts[0][keyStart:]
						value := parts[1]
						
						// Apply styles
						highlightedLine = indent + jsonKeyStyle.Render(key+"\":") + jsonValueStyle.Render(value)
					}
				}
			} else if strings.TrimSpace(line) == "{" || strings.TrimSpace(line) == "}" || 
			          strings.TrimSpace(line) == "[" || strings.TrimSpace(line) == "]" ||
			          strings.TrimSpace(line) == "}," || strings.TrimSpace(line) == "]," {
				// Highlight braces and brackets
				highlightedLine = jsonBraceStyle.Render(line)
			}
			
			content.WriteString(lineNum + highlightedLine + "\n")
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