package ui

import tea "github.com/charmbracelet/bubbletea"

// ContainerSearchAware is an interface for views that support container search
type ContainerSearchAware interface {
	// StartSearch initializes search mode
	StartSearch()
	// HandleSearchInput handles search input in search mode
	HandleSearchInput(model *Model, msg tea.KeyMsg) tea.Cmd
	// HandleNextSearchResult moves to the next search result
	HandleNextSearchResult()
	// HandlePrevSearchResult moves to the previous search result
	HandlePrevSearchResult()
	// IsSearchActive returns true if the search mode is active
	IsSearchActive() bool
	// ExitSearch exits search mode
	ExitSearch()
	// RenderSearchLine renders the search input line
	RenderSearchLine() string
}
