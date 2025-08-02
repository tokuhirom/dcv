package ui

import (
	"fmt"
	"strings"
)

func (m *Model) renderLogView() string {
	var s strings.Builder

	title := fmt.Sprintf("Logs: %s", m.containerName)
	if m.isDindLog {
		title = fmt.Sprintf("Logs: %s (in %s)", m.containerName, m.hostContainer)
	}
	s.WriteString(titleStyle.Render(title) + "\n\n")

	if m.loading && len(m.logs) == 0 {
		s.WriteString("Loading logs...\n")
		return s.String()
	}

	// Search mode indicator
	if m.searchMode {
		s.WriteString(searchStyle.Render(fmt.Sprintf("Search: %s", m.searchText)) + "\n\n")
	}

	// Calculate visible logs based on scroll position
	visibleHeight := m.height - 4 // Account for title and help
	if m.searchMode {
		visibleHeight -= 2
	}

	startIdx := m.logScrollY
	endIdx := startIdx + visibleHeight

	if endIdx > len(m.logs) {
		endIdx = len(m.logs)
	}

	// Display logs
	if len(m.logs) == 0 {
		s.WriteString("No logs available.\n")
	} else {
		for i := startIdx; i < endIdx; i++ {
			if i < len(m.logs) {
				s.WriteString(m.logs[i] + "\n")
			}
		}
	}

	// Scroll indicator
	if len(m.logs) > visibleHeight {
		scrollInfo := fmt.Sprintf(" [%d-%d/%d] ", startIdx+1, endIdx, len(m.logs))
		s.WriteString("\n" + helpStyle.Render(scrollInfo))
	}

	// Show help hint
	s.WriteString(helpStyle.Render("Press ? for help"))

	return s.String()
}
