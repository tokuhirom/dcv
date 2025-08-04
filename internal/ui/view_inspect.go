package ui

import (
	"fmt"
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tokuhirom/dcv/internal/docker"
)

type InspectViewModel struct {
}

// renderInspectView renders the container inspect view
func (m *Model) renderInspectView(availableHeight int) string {
	var content strings.Builder

	// Split content into lines for scrolling
	lines := strings.Split(m.inspectContent, "\n")
	viewHeight := availableHeight
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

	// Define highlight style for search matches
	highlightStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("226")).
		Foreground(lipgloss.Color("235"))

	if len(lines) == 0 {
		content.WriteString("No inspect data available\n")
	} else if startIdx < len(lines) {
		for i := startIdx; i < endIdx; i++ {
			line := lines[i]
			lineNum := lineNumStyle.Render(fmt.Sprintf("%4d ", i+1))

			// Mark current search result line
			if len(m.searchResults) > 0 && m.currentSearchIdx < len(m.searchResults) &&
				i == m.searchResults[m.currentSearchIdx] {
				// Add a marker in the margin
				lineNum = lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Render("▶") + lineNum[1:]
			}

			// Apply JSON syntax highlighting first
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

			// Then apply search highlighting if in search mode and have search text
			if m.searchText != "" && !m.searchMode {
				highlightedLine = m.highlightInspectLine(line, highlightedLine, highlightStyle)
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
		if len(m.searchResults) > 0 {
			position += fmt.Sprintf(" | Match %d/%d", m.currentSearchIdx+1, len(m.searchResults))
		}
		content.WriteString(posStyle.Render(position))
	}

	return content.String()
}

// highlightInspectLine highlights search matches in a line that may already have JSON syntax highlighting
func (m *Model) highlightInspectLine(originalLine, styledLine string, highlightStyle lipgloss.Style) string {
	if m.searchText == "" {
		return styledLine
	}

	// For inspect view, we need to be careful not to break the existing JSON syntax highlighting
	// We'll use a simpler approach - just highlight in the original line for display
	if m.searchRegex {
		pattern := m.searchText
		if m.searchIgnoreCase {
			pattern = "(?i)" + pattern
		}
		if re, err := regexp.Compile(pattern); err == nil {
			return re.ReplaceAllStringFunc(originalLine, func(match string) string {
				return highlightStyle.Render(match)
			})
		}
	} else {
		// Simple string search
		searchStr := m.searchText
		lineToSearch := originalLine

		if m.searchIgnoreCase {
			searchStr = strings.ToLower(searchStr)
			lineToSearch = strings.ToLower(originalLine)
		}

		// Find all occurrences
		var result strings.Builder
		lastEnd := 0
		for {
			idx := strings.Index(lineToSearch[lastEnd:], searchStr)
			if idx == -1 {
				break
			}

			realIdx := lastEnd + idx
			result.WriteString(originalLine[lastEnd:realIdx])
			result.WriteString(highlightStyle.Render(originalLine[realIdx : realIdx+len(m.searchText)]))
			lastEnd = realIdx + len(m.searchText)
		}
		result.WriteString(originalLine[lastEnd:])
		return result.String()
	}

	return styledLine
}

func loadInspect(client *docker.Client, containerID string) tea.Cmd {
	return func() tea.Msg {
		content, err := client.InspectContainer(containerID)
		return inspectLoadedMsg{
			content: content,
			err:     err,
		}
	}
}

func (m InspectViewModel) InspectContainer(model *Model, containerID string) tea.Cmd {
	model.inspectContainerID = containerID
	model.loading = true
	return loadInspect(model.dockerClient, containerID)
}
