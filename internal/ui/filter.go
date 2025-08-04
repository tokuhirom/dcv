package ui

import "github.com/charmbracelet/lipgloss"

type FilterViewModel struct {
	// Filter state
	filterMode      bool
	filterText      string
	filterCursorPos int
	filteredLogs    []string // Logs that match the filter
}

func (m *FilterViewModel) FilterDeleteLastChar() bool {
	if m.filterCursorPos > 0 && len(m.filterText) > 0 {
		m.filterText = m.filterText[:m.filterCursorPos-1] + m.filterText[m.filterCursorPos:]
		m.filterCursorPos--
		return true
	}
	return false
}

func (m *FilterViewModel) FilterCursorLeft() {
	if m.filterCursorPos > 0 {
		m.filterCursorPos--
	}
}

func (m *FilterViewModel) AppendString(str string) {
	m.filterText = m.filterText[:m.filterCursorPos] + str + m.filterText[m.filterCursorPos:]
	m.filterCursorPos += len(str)
}

func (m *FilterViewModel) RenderCmdLine() string {
	// Show filter prompt
	cursor := " "
	if m.filterCursorPos < len(m.filterText) {
		cursor = string(m.filterText[m.filterCursorPos])
	}

	// Build filter line with cursor
	before := m.filterText[:m.filterCursorPos]
	after := ""
	if m.filterCursorPos < len(m.filterText) {
		after = m.filterText[m.filterCursorPos+1:]
	}

	cursorStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("226")).
		Foreground(lipgloss.Color("235"))

	return "Filter: " + before + cursorStyle.Render(cursor) + after + " (ESC to clear)"
}
