package ui

import (
	"fmt"
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
)

type LogViewModel struct {
}

func (m LogViewModel) Clear(model *Model, containerName string) {
	model.containerName = containerName
	model.isDindLog = false
	model.currentView = LogView
	model.logs = []string{}
	model.logScrollY = 0
}

func (m LogViewModel) StreamLogs(model *Model, process models.ComposeContainer, isDind bool, hostService string) tea.Cmd {
	model.containerName = process.Name
	model.isDindLog = false
	model.currentView = LogView
	model.logs = []string{}
	model.logScrollY = 0
	return streamLogs(model.dockerClient, process.ID, isDind, hostService)
}

func (m LogViewModel) ShowDindLog(model *Model, dindContainerID string, container models.DockerContainer) tea.Cmd {
	model.containerName = container.Names
	model.hostContainer = model.dindProcessListViewModel.currentDindHost
	model.isDindLog = true
	model.currentView = LogView
	model.logs = []string{}
	model.logScrollY = 0
	return streamLogs(model.dockerClient, container.Names, true, dindContainerID)
}

func (m *Model) renderLogView(availableHeight int) string {
	var s strings.Builder

	if m.loading && len(m.logs) == 0 {
		s.WriteString("Loading logs...\n")
		return s.String()
	}

	// Determine which logs to display
	logsToDisplay := m.logs
	if m.filterMode && m.filterText != "" {
		logsToDisplay = m.filteredLogs
	}

	// Calculate visible logs based on scroll position
	visibleHeight := availableHeight

	startIdx := m.logScrollY
	endIdx := startIdx + visibleHeight

	if endIdx > len(logsToDisplay) {
		endIdx = len(logsToDisplay)
	}

	// Define highlight style
	highlightStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("226")).
		Foreground(lipgloss.Color("235"))

	// Display logs
	if len(logsToDisplay) == 0 {
		if m.filterMode && m.filterText != "" {
			s.WriteString("No logs match the filter.\n")
		} else {
			s.WriteString("No logs available.\n")
		}
	} else {
		for i := startIdx; i < endIdx; i++ {
			if i < len(logsToDisplay) {
				line := logsToDisplay[i]

				// Highlight search matches if we have results and search text
				if m.searchText != "" && !m.searchMode && !m.filterMode {
					line = m.highlightLine(line, highlightStyle)
				}

				// Highlight filter matches in filter mode
				if m.filterMode && m.filterText != "" {
					line = m.highlightFilterMatch(line, highlightStyle)
				}

				// Mark current search result line (only in search mode)
				if !m.filterMode && len(m.searchResults) > 0 && m.currentSearchIdx < len(m.searchResults) &&
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
	if len(logsToDisplay) > visibleHeight {
		scrollInfo := fmt.Sprintf(" [%d-%d/%d] ", startIdx+1, endIdx, len(logsToDisplay))
		if m.filterMode && m.filterText != "" {
			scrollInfo += fmt.Sprintf(" (filtered from %d)", len(m.logs))
		}
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

func (m *Model) highlightFilterMatch(line string, style lipgloss.Style) string {
	if m.filterText == "" {
		return line
	}

	// Simple case-insensitive string search for filter
	searchStr := strings.ToLower(m.filterText)
	lineToSearch := strings.ToLower(line)

	// Find all occurrences
	var result strings.Builder
	lastEnd := 0
	searchLen := len(searchStr)

	for lastEnd < len(line) {
		idx := strings.Index(lineToSearch[lastEnd:], searchStr)
		if idx == -1 {
			// No more matches, append the rest
			result.WriteString(line[lastEnd:])
			break
		}

		// Found a match
		matchStart := lastEnd + idx
		matchEnd := matchStart + searchLen

		// Append text before the match
		result.WriteString(line[lastEnd:matchStart])
		// Append highlighted match
		result.WriteString(style.Render(line[matchStart:matchEnd]))
		// Move past this match
		lastEnd = matchEnd
	}

	return result.String()
}

func streamLogs(client *docker.Client, serviceName string, isDind bool, hostService string) tea.Cmd {
	return streamLogsReal(client, serviceName, isDind, hostService)
}
