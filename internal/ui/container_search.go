package ui

import (
	"fmt"
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ContainerSearchViewModel handles search functionality for container views
type ContainerSearchViewModel struct {
	searchMode       bool
	searchText       string
	searchIgnoreCase bool
	searchRegex      bool
	searchResults    []int // Indices of matching containers
	currentSearchIdx int   // Current position in searchResults
	searchCursorPos  int   // Cursor position in search text
	onSelectIndex    func(idx int)
	onPerformSearch  func()
}

// RenderSearchLine renders the search input line
func (m *ContainerSearchViewModel) RenderSearchLine() string {
	searchText := m.searchText
	searchCursorPos := m.searchCursorPos

	cursor := " "
	if searchCursorPos < len(searchText) {
		cursor = string(searchText[searchCursorPos])
	}

	// Build search line with cursor
	before := searchText[:searchCursorPos]
	after := ""
	if searchCursorPos < len(searchText) {
		after = searchText[searchCursorPos+1:]
	}

	cursorStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("226")).
		Foreground(lipgloss.Color("235"))

	return "/" + before + cursorStyle.Render(cursor) + after
}

// StartSearch initializes search mode
func (m *ContainerSearchViewModel) StartSearch() {
	m.searchMode = true
	m.searchText = ""
	m.searchCursorPos = 0
	m.searchResults = nil
	m.currentSearchIdx = 0
}

// ExitSearch exits search mode
func (m *ContainerSearchViewModel) ExitSearch() {
	m.searchMode = false
}

// ClearSearch clears search state but keeps search mode active
func (m *ContainerSearchViewModel) ClearSearch() {
	m.searchText = ""
	m.searchCursorPos = 0
	m.searchResults = nil
	m.currentSearchIdx = 0
}

// IsSearchActive returns true if search mode is active
func (m *ContainerSearchViewModel) IsSearchActive() bool {
	return m.searchMode
}

// HasSearchResults returns true if there are search results
func (m *ContainerSearchViewModel) HasSearchResults() bool {
	return len(m.searchResults) > 0
}

// GetCurrentResult returns the current search result index (-1 if no results)
func (m *ContainerSearchViewModel) GetCurrentResult() int {
	if m.currentSearchIdx < len(m.searchResults) {
		return m.searchResults[m.currentSearchIdx]
	}
	return -1
}

// MatchContainer checks if a container matches the current search
func (m *ContainerSearchViewModel) MatchContainer(containerText string) bool {
	if m.searchText == "" {
		return false
	}

	searchText := m.searchText
	containerToSearch := containerText

	if m.searchIgnoreCase && !m.searchRegex {
		searchText = strings.ToLower(searchText)
		containerToSearch = strings.ToLower(containerText)
	}

	if m.searchRegex {
		pattern := searchText
		if m.searchIgnoreCase {
			pattern = "(?i)" + pattern
		}
		if re, err := regexp.Compile(pattern); err == nil {
			return re.MatchString(containerText)
		}
		return false
	}

	return strings.Contains(containerToSearch, searchText)
}

// AppendChar adds a character to the search text at cursor position
func (m *ContainerSearchViewModel) AppendChar(char string) {
	m.searchText = m.searchText[:m.searchCursorPos] + char + m.searchText[m.searchCursorPos:]
	m.searchCursorPos += len(char)
}

// DeleteChar removes the character before the cursor
func (m *ContainerSearchViewModel) DeleteChar() bool {
	if m.searchCursorPos > 0 && len(m.searchText) > 0 {
		m.searchText = m.searchText[:m.searchCursorPos-1] + m.searchText[m.searchCursorPos:]
		m.searchCursorPos--
		return true
	}
	return false
}

// MoveCursorLeft moves the cursor left
func (m *ContainerSearchViewModel) MoveCursorLeft() {
	if m.searchCursorPos > 0 {
		m.searchCursorPos--
	}
}

// MoveCursorRight moves the cursor right
func (m *ContainerSearchViewModel) MoveCursorRight() {
	if m.searchCursorPos < len(m.searchText) {
		m.searchCursorPos++
	}
}

// ToggleIgnoreCase toggles case-insensitive search
func (m *ContainerSearchViewModel) ToggleIgnoreCase() {
	m.searchIgnoreCase = !m.searchIgnoreCase
}

// ToggleRegex toggles regex search mode
func (m *ContainerSearchViewModel) ToggleRegex() {
	m.searchRegex = !m.searchRegex
}

// NextResult moves to the next search result
func (m *ContainerSearchViewModel) NextResult() int {
	if len(m.searchResults) > 0 {
		m.currentSearchIdx = (m.currentSearchIdx + 1) % len(m.searchResults)
		return m.searchResults[m.currentSearchIdx]
	}
	return -1
}

// PrevResult moves to the previous search result
func (m *ContainerSearchViewModel) PrevResult() int {
	if len(m.searchResults) > 0 {
		m.currentSearchIdx = (m.currentSearchIdx - 1 + len(m.searchResults)) % len(m.searchResults)
		return m.searchResults[m.currentSearchIdx]
	}
	return -1
}

// SetResults sets the search results
func (m *ContainerSearchViewModel) SetResults(results []int) {
	m.searchResults = results
	m.currentSearchIdx = 0
}

// GetSearchText returns the current search text
func (m *ContainerSearchViewModel) GetSearchText() string {
	return m.searchText
}

// GetSearchInfo returns information about the current search
func (m *ContainerSearchViewModel) GetSearchInfo() string {
	if !m.searchMode || m.searchText == "" {
		return ""
	}

	if len(m.searchResults) == 0 {
		return "No matches"
	}

	return strings.TrimSpace(searchStyle.Render(
		fmt.Sprintf("Match %d/%d", m.currentSearchIdx+1, len(m.searchResults))))
}

func (m *ContainerSearchViewModel) InitContainerSearchViewModel(onSelectIndex func(idx int), onPerformSearch func()) {
	m.onSelectIndex = onSelectIndex
	m.onPerformSearch = onPerformSearch
}

func (m *ContainerSearchViewModel) HandleSearchInput(_ *Model, msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		m.ExitSearch()
	case "enter":
		m.ExitSearch()
		m.onPerformSearch()
	case "backspace":
		if m.DeleteChar() {
			m.onPerformSearch()
		}
	case "left":
		m.MoveCursorLeft()
	case "right":
		m.MoveCursorRight()
	case "ctrl+i":
		m.ToggleIgnoreCase()
		m.onPerformSearch()
	case "ctrl+r":
		m.ToggleRegex()
		m.onPerformSearch()
	default:
		if len(msg.String()) == 1 {
			m.AppendChar(msg.String())
			m.onPerformSearch()
		}
	}
	return nil
}

func (m *ContainerSearchViewModel) HandleNextSearchResult() {
	if m.HasSearchResults() {
		idx := m.NextResult()
		if idx >= 0 {
			m.onSelectIndex(idx)
		}
	}
}

func (m *ContainerSearchViewModel) HandlePrevSearchResult() {
	if m.HasSearchResults() {
		idx := m.PrevResult()
		if idx >= 0 {
			m.onSelectIndex(idx)
		}
	}
}
