package ui

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
