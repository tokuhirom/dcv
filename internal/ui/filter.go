package ui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

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

func (m *FilterViewModel) FilterCursorRight() {
	if m.filterCursorPos < len(m.filterText) {
		m.filterCursorPos++
	}
}

func (m *FilterViewModel) AppendString(str string) {
	m.filterText = m.filterText[:m.filterCursorPos] + str + m.filterText[m.filterCursorPos:]
	m.filterCursorPos += len(str)
}

func (m *FilterViewModel) RenderFilterCmdLine() string {
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

func (m *FilterViewModel) HandleKey(msg tea.KeyPressMsg) bool {
	switch msg.Code {
	case tea.KeyEsc:
		m.ClearFilter()
	case tea.KeyEnter:
		m.filterMode = false
		return true
	case tea.KeyBackspace:
		m.FilterDeleteLastChar()
	case tea.KeyLeft:
		m.FilterCursorLeft()
	case tea.KeyRight:
		m.FilterCursorRight()
	default:
		switch {
		case isCtrlKey(msg, 'h'):
			m.FilterDeleteLastChar()
		case isCtrlKey(msg, 'f'):
			m.FilterCursorLeft()
		case isCtrlKey(msg, 'b'):
			m.FilterCursorRight()
		case len(msg.Text) > 0:
			// Insert at the cursor position
			m.AppendString(msg.Text)
			return true
		}
	}
	return false
}

func (m *FilterViewModel) ClearFilter() {
	m.filterMode = false
	m.filterText = ""
	m.filteredLogs = nil
	m.filterCursorPos = 0
}
