package tui

import (
	"github.com/tokuhirom/dcv/internal/ui"
)

// State holds the application state
type State struct {
	CurrentView       ui.ViewType
	ViewHistory       []ui.ViewType
	SelectedProject   string
	SelectedContainer string
	ShowAll           bool
	SearchText        string
	CommandMode       bool
	CommandText       string
	Error             error
	Loading           bool
}

// NewState creates a new state instance
func NewState() *State {
	return &State{
		ViewHistory: make([]ui.ViewType, 0),
		ShowAll:     false,
	}
}

// PushView adds a view to the history
func (s *State) PushView(view ui.ViewType) {
	s.ViewHistory = append(s.ViewHistory, view)
	s.CurrentView = view
}

// PopView removes the current view and returns to the previous one
func (s *State) PopView() ui.ViewType {
	if len(s.ViewHistory) > 1 {
		s.ViewHistory = s.ViewHistory[:len(s.ViewHistory)-1]
		s.CurrentView = s.ViewHistory[len(s.ViewHistory)-1]
		return s.CurrentView
	}
	return s.CurrentView
}

// ClearError clears the error state
func (s *State) ClearError() {
	s.Error = nil
}

// SetError sets an error state
func (s *State) SetError(err error) {
	s.Error = err
	s.Loading = false
}

// SetLoading sets the loading state
func (s *State) SetLoading(loading bool) {
	s.Loading = loading
	if loading {
		s.Error = nil
	}
}
