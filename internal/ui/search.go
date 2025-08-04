package ui

import (
	"regexp"
	"strings"
)

type SearchViewModel struct {
	searchMode       bool
	searchText       string
	searchIgnoreCase bool
	searchRegex      bool
	searchResults    []int // Line indices of search results
	currentSearchIdx int   // Current position in searchResults
	searchCursorPos  int   // Cursor position in search text
}

func (m *SearchViewModel) ClearSearch() {
	m.searchMode = true
	m.searchText = ""
	m.searchCursorPos = 0
	m.searchResults = nil
	m.currentSearchIdx = 0
}

func (m *SearchViewModel) PerformSearch(model *Model, logs []string, updateScrollY func(int)) {
	m.searchResults = nil
	if m.searchText == "" {
		return
	}

	searchText := m.searchText
	if m.searchIgnoreCase && !m.searchRegex {
		searchText = strings.ToLower(searchText)
	}

	for i, line := range logs {
		lineToSearch := line
		if m.searchIgnoreCase && !m.searchRegex {
			lineToSearch = strings.ToLower(line)
		}

		match := false
		if m.searchRegex {
			pattern := searchText
			if m.searchIgnoreCase {
				pattern = "(?i)" + pattern
			}
			if re, err := regexp.Compile(pattern); err == nil {
				match = re.MatchString(line)
			}
		} else {
			match = strings.Contains(lineToSearch, searchText)
		}

		if match {
			m.searchResults = append(m.searchResults, i)
		}
	}

	// If we have results, jump to the first one
	if len(m.searchResults) > 0 && m.currentSearchIdx < len(m.searchResults) {
		targetLine := m.searchResults[m.currentSearchIdx]
		logScrollY := targetLine - model.Height/2 + 3
		if logScrollY < 0 {
			logScrollY = 0
		}
		updateScrollY(logScrollY)
	}
}

func (m *SearchViewModel) InputEscape() {
	m.searchMode = false
	m.searchText = ""
	m.searchResults = nil
	m.currentSearchIdx = 0
	m.searchCursorPos = 0
}

func (m *SearchViewModel) DeleteLastChar() bool {
	if m.searchCursorPos > 0 && len(m.searchText) > 0 {
		m.searchText = m.searchText[:m.searchCursorPos-1] + m.searchText[m.searchCursorPos:]
		m.searchCursorPos--
		return true
	}
	return false
}

func (m *SearchViewModel) CursorLeft() {
	if m.searchCursorPos > 0 {
		m.searchCursorPos--
	}
}

func (m *SearchViewModel) CursorRight() {
	if m.searchCursorPos < len(m.searchText) {
		m.searchCursorPos++
	}
}

func (m *SearchViewModel) ToggleIgnoreCase() {
	m.searchIgnoreCase = !m.searchIgnoreCase
}

func (m *SearchViewModel) ToggleRegex() {
	m.searchRegex = !m.searchRegex
}

func (m *SearchViewModel) AppendString(str string) {
	// Insert at cursor position
	m.searchText = m.searchText[:m.searchCursorPos] + str + m.searchText[m.searchCursorPos:]
	m.searchCursorPos += len(str)
}
