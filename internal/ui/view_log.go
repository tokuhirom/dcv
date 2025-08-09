package ui

import (
	"fmt"
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tokuhirom/dcv/internal/models"
)

type LogViewModel struct {
	SearchViewModel
	FilterViewModel

	logs              []string
	logScrollY        int
	containerName     string
	isDindLog         bool
	hostContainerName string

	LogReaderManager
}

func (m *LogViewModel) SwitchToLogView(model *Model, containerName string) {
	model.SwitchView(LogView)

	m.containerName = containerName
	m.isDindLog = false
	m.logs = []string{}
	m.logScrollY = 0
}

func (m *LogViewModel) StreamContainerLogs(model *Model, container models.DockerContainer) tea.Cmd {
	m.SwitchToLogView(model, container.Names)
	cmd := model.dockerClient.Execute("logs", container.ID, "--tail", "1000", "--timestamps", "--follow")
	return m.streamLogsReal(cmd)
}

func (m *LogViewModel) StreamComposeLogs(model *Model, composeContainer models.ComposeContainer) tea.Cmd {
	m.SwitchToLogView(model, composeContainer.Name)

	cmd := model.dockerClient.Execute("logs", composeContainer.ID, "--tail", "1000", "--timestamps", "--follow")
	return m.streamLogsReal(cmd)
}

func (m *LogViewModel) StreamLogsDind(model *Model, dindHostContainerID string, container models.DockerContainer) tea.Cmd {
	m.SwitchToLogView(model, container.Names)
	m.hostContainerName = model.dindProcessListViewModel.hostContainer.GetName()
	m.isDindLog = true

	cmd := model.dockerClient.Dind(dindHostContainerID).Execute(
		"logs", container.ID, "--tail", "1000", "--timestamps", "--follow")
	return m.streamLogsReal(cmd)
}

func (m *LogViewModel) HandleBack(model *Model) tea.Cmd {
	m.stopLogReader()
	model.SwitchToPreviousView()
	return nil
}

func (m *LogViewModel) render(model *Model, availableHeight int) string {
	var s strings.Builder

	if model.loading && len(m.logs) == 0 {
		s.WriteString("Loading logs...\n")
		return s.String()
	}

	// Determine which logs to display
	logsToDisplay := m.logs
	if m.filterMode && m.filterText != "" {
		logsToDisplay = m.filteredLogs
	}

	// Calculate visible logs based on scroll position
	visibleHeight := availableHeight - 2

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

func (m *LogViewModel) highlightLine(line string, style lipgloss.Style) string {
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

func (m *LogViewModel) highlightFilterMatch(line string, style lipgloss.Style) string {
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

func (m *LogViewModel) HandleUp() tea.Cmd {
	if m.logScrollY > 0 {
		m.logScrollY--
	}
	return nil
}

func (m *LogViewModel) HandleDown(model *Model) tea.Cmd {
	maxScroll := len(m.logs) - (model.Height - 4)
	if m.logScrollY < maxScroll && maxScroll > 0 {
		m.logScrollY++
	}
	return nil
}

func (m *LogViewModel) HandleGoToEnd(model *Model) tea.Cmd {
	maxScroll := len(m.logs) - (model.Height - 4)
	if maxScroll > 0 {
		m.logScrollY = maxScroll
	}
	return nil
}

func (m *LogViewModel) HandleGoToStart() tea.Cmd {
	m.logScrollY = 0
	return nil
}

func (m *LogViewModel) HandleSearch() tea.Cmd {
	m.searchMode = true
	m.searchText = ""
	m.searchCursorPos = 0
	m.searchResults = nil
	m.currentSearchIdx = 0
	return nil
}

func (m *LogViewModel) HandleFilter() tea.Cmd {
	m.filterMode = true
	m.filterText = ""
	m.filterCursorPos = 0
	m.filteredLogs = nil
	return nil
}

func (m *LogViewModel) HandleNextSearchResult(model *Model) tea.Cmd {
	if len(m.searchResults) > 0 {
		m.currentSearchIdx = (m.currentSearchIdx + 1) % len(m.searchResults)
		// Jump to the line
		if m.currentSearchIdx < len(m.searchResults) {
			targetLine := m.searchResults[m.currentSearchIdx]
			m.logScrollY = targetLine - model.Height/2 + 3 // Center the result
			if m.logScrollY < 0 {
				m.logScrollY = 0
			}
		}
	}
	return nil
}

func (m *LogViewModel) HandlePrevSearchResult(model *Model) tea.Cmd {
	if len(m.searchResults) > 0 {
		m.currentSearchIdx--
		if m.currentSearchIdx < 0 {
			m.currentSearchIdx = len(m.searchResults) - 1
		}
		// Jump to the line
		if m.currentSearchIdx < len(m.searchResults) {
			targetLine := m.searchResults[m.currentSearchIdx]
			m.logScrollY = targetLine - model.Height/2 + 3 // Center the result
			if m.logScrollY < 0 {
				m.logScrollY = 0
			}
		}
	}
	return nil
}

func (m *LogViewModel) performFilter() {
	m.filteredLogs = nil
	if m.filterText == "" {
		return
	}

	filterText := strings.ToLower(m.filterText)

	for _, line := range m.logs {
		lineToSearch := strings.ToLower(line)
		if strings.Contains(lineToSearch, filterText) {
			m.filteredLogs = append(m.filteredLogs, line)
		}
	}

	// Reset scroll position when filter changes
	m.logScrollY = 0
}

func (m *LogViewModel) LogLines(model *Model, lines []string) {
	m.logs = append(m.logs, lines...)
	// Keep only last 10000 lines to prevent unbounded memory growth
	if len(m.logs) > 10000 {
		m.logs = m.logs[len(m.logs)-10000:]
	}

	// If we're in filter mode, update filtered logs
	if m.filterMode && m.filterText != "" {
		m.performFilter()
	} else {
		// Auto-scroll to bottom only when not filtering
		maxScroll := len(m.logs) - (model.Height - 4)
		if maxScroll > 0 {
			m.logScrollY = maxScroll
		}
	}
}

func (m *LogViewModel) FilterDeleteLastChar() {
	updated := m.FilterViewModel.FilterDeleteLastChar()
	if updated {
		m.performFilter()
	}
}

func (m *LogViewModel) Title() string {
	title := ""
	if m.isDindLog {
		title = fmt.Sprintf("Logs: %s (in %s)", m.containerName, m.hostContainerName)
	} else {
		title = fmt.Sprintf("Logs: %s", m.containerName)
	}

	// Add search or filter status to title
	if m.filterMode && m.filterText != "" {
		filterCount := len(m.filteredLogs)
		title += fmt.Sprintf(" - Filter: '%s' (%d/%d lines)", m.filterText, filterCount, len(m.logs))
	} else if len(m.searchResults) > 0 {
		var statusParts []string
		if m.searchIgnoreCase {
			statusParts = append(statusParts, "i")
		}
		if m.searchRegex {
			statusParts = append(statusParts, "r")
		}

		statusStr := ""
		if len(statusParts) > 0 {
			statusStr = fmt.Sprintf(" [%s]", strings.Join(statusParts, ""))
		}

		title += fmt.Sprintf(" - Search: %d/%d%s", m.currentSearchIdx+1, len(m.searchResults), statusStr)
	} else if m.searchText != "" && !m.searchMode {
		title += " - No matches found"
	}

	return title
}

func (m *LogViewModel) HandleCancel() tea.Cmd {
	m.stopLogReader()
	return nil
}
