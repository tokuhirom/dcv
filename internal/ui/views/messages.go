package views

// RefreshMsg signals that the current view should be refreshed
type RefreshMsg struct{}

// serviceActionCompleteMsg indicates a service action (like delete) completed
type serviceActionCompleteMsg struct {
	err error
}
