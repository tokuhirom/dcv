package ui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) renderLogView(availableHeight int) string {
	var s strings.Builder

	if m.loading && len(m.logs) == 0 {
		s.WriteString("Loading logs...\n")
		return s.String()
	}

	// Calculate visible logs based on scroll position
	visibleHeight := availableHeight

	startIdx := m.logScrollY
	endIdx := startIdx + visibleHeight

	if endIdx > len(m.logs) {
		endIdx = len(m.logs)
	}

	// Define highlight style
	highlightStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("226")).
		Foreground(lipgloss.Color("235"))

	// Display logs
	if len(m.logs) == 0 {
		s.WriteString("No logs available.\n")
	} else {
		for i := startIdx; i < endIdx; i++ {
			if i < len(m.logs) {
				line := m.logs[i]

				// Highlight search matches if we have results and search text
				if m.searchText != "" && !m.searchMode {
					line = m.highlightLine(line, highlightStyle)
				}

				// Mark current search result line
				if len(m.searchResults) > 0 && m.currentSearchIdx < len(m.searchResults) &&
					i == m.searchResults[m.currentSearchIdx] {
					// Add a marker in the margin
					s.WriteString("> ")
				} else {
					s.WriteString("  ")
				}

				s.WriteString(line + "\n")
			}
		}
	}

	// Scroll indicator
	if len(m.logs) > visibleHeight {
		scrollInfo := fmt.Sprintf(" [%d-%d/%d] ", startIdx+1, endIdx, len(m.logs))
		s.WriteString("\n" + helpStyle.Render(scrollInfo))
	}

	return s.String()
}

func (m *Model) highlightLine(line string, style lipgloss.Style) string {
	if m.searchText == "" {
		return line
	}

	if m.searchRegex {
		pattern := m.searchText
		if m.searchIgnoreCase {
			pattern = "(?i)" + pattern
		}
		if re, err := regexp.Compile(pattern); err == nil {
			// Find all matches
			matches := re.FindAllStringIndex(line, -1)
			if len(matches) == 0 {
				return line
			}

			// Build the line with highlights
			var result strings.Builder
			lastEnd := 0
			for _, match := range matches {
				start, end := match[0], match[1]
				result.WriteString(line[lastEnd:start])
				result.WriteString(style.Render(line[start:end]))
				lastEnd = end
			}
			result.WriteString(line[lastEnd:])
			return result.String()
		}
	} else {
		// Simple string search
		searchStr := m.searchText
		lineToSearch := line

		if m.searchIgnoreCase {
			searchStr = strings.ToLower(searchStr)
			lineToSearch = strings.ToLower(line)
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
			result.WriteString(line[lastEnd:realIdx])
			result.WriteString(style.Render(line[realIdx : realIdx+len(m.searchText)]))
			lastEnd = realIdx + len(m.searchText)
		}
		result.WriteString(line[lastEnd:])
		return result.String()
	}

	return line
}
