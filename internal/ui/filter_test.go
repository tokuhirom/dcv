package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestFilterMode(t *testing.T) {
	t.Run("start_filter_mode", func(t *testing.T) {
		m := &Model{
			currentView: LogView,
			logs:        []string{"line1", "error: something", "line3", "error: another"},
			filterMode:  false,
		}
		m.initializeKeyHandlers()

		// Press 'f' to start filter mode
		newModel, _ := m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")})
		m = newModel.(*Model)

		assert.True(t, m.filterMode)
		assert.Equal(t, "", m.filterText)
		assert.Equal(t, 0, m.filterCursorPos)
	})

	t.Run("filter_logs", func(t *testing.T) {
		m := &Model{
			currentView:     LogView,
			logs:            []string{"line1", "error: something", "line3", "error: another", "info: test"},
			filterMode:      true,
			filterText:      "",
			filterCursorPos: 0,
		}

		// Type "error"
		for _, ch := range "error" {
			newModel, _ := m.handleFilterMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
			m = newModel.(*Model)
		}

		assert.Equal(t, "error", m.filterText)
		assert.Equal(t, 2, len(m.filteredLogs))
		assert.Equal(t, "error: something", m.filteredLogs[0])
		assert.Equal(t, "error: another", m.filteredLogs[1])
	})

	t.Run("exit_filter_mode_enter", func(t *testing.T) {
		m := &Model{
			currentView:  LogView,
			logs:         []string{"line1", "error: something", "line3"},
			filterMode:   true,
			filterText:   "error",
			filteredLogs: []string{"error: something"},
		}

		// Press Enter to exit filter mode
		newModel, _ := m.handleFilterMode(tea.KeyMsg{Type: tea.KeyEnter})
		m = newModel.(*Model)

		assert.False(t, m.filterMode)
		assert.Equal(t, "error", m.filterText) // Filter text is preserved
		assert.Equal(t, 1, len(m.filteredLogs))
	})

	t.Run("exit_filter_mode_escape", func(t *testing.T) {
		m := &Model{
			currentView:  LogView,
			logs:         []string{"line1", "error: something", "line3"},
			filterMode:   true,
			filterText:   "error",
			filteredLogs: []string{"error: something"},
			logScrollY:   5,
		}

		// Press Escape to exit filter mode and clear filter
		newModel, _ := m.handleFilterMode(tea.KeyMsg{Type: tea.KeyEsc})
		m = newModel.(*Model)

		assert.False(t, m.filterMode)
		assert.Equal(t, "", m.filterText)      // Filter text is cleared
		assert.Nil(t, m.filteredLogs)          // Filtered logs are cleared
		assert.Equal(t, 0, m.logScrollY)       // Scroll is reset
	})

	t.Run("case_insensitive_filter", func(t *testing.T) {
		m := &Model{
			currentView:     LogView,
			logs:            []string{"ERROR: big", "error: small", "Error: mixed", "info: test"},
			filterMode:      true,
			filterText:      "",
			filterCursorPos: 0,
		}

		// Type "error" - should match all case variations
		for _, ch := range "error" {
			newModel, _ := m.handleFilterMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
			m = newModel.(*Model)
		}

		assert.Equal(t, 3, len(m.filteredLogs))
		assert.Contains(t, m.filteredLogs, "ERROR: big")
		assert.Contains(t, m.filteredLogs, "error: small")
		assert.Contains(t, m.filteredLogs, "Error: mixed")
	})
}

func TestFilterHighlight(t *testing.T) {
	// Test the highlight logic works correctly
	m := &Model{
		filterText: "error",
		filterMode: true,
	}

	// Test case-insensitive matching
	testCases := []struct {
		line     string
		expected bool // Should find a match
	}{
		{"This is an error message", true},
		{"This is an ERROR message", true},
		{"This is an Error message", true},
		{"This is a warning message", false},
		{"", false},
	}

	for _, tc := range testCases {
		t.Run(tc.line, func(t *testing.T) {
			searchStr := strings.ToLower(m.filterText)
			lineToSearch := strings.ToLower(tc.line)
			hasMatch := strings.Contains(lineToSearch, searchStr)
			assert.Equal(t, tc.expected, hasMatch, "Match result for: %s", tc.line)
		})
	}
}