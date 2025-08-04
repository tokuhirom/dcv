package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// TopViewModel manages the state and rendering of the process info view
type TopViewModel struct {
	topOutput  string
	topService string
}

// render renders the top view
func (m *TopViewModel) render(model *Model, availableHeight int) string {
	var s strings.Builder

	if m.topOutput == "" {
		s.WriteString("No process information available.\n")
	} else {
		// Display the raw top output
		lines := strings.Split(m.topOutput, "\n")
		visibleHeight := availableHeight

		for i, line := range lines {
			if i >= visibleHeight {
				break
			}
			s.WriteString(line + "\n")
		}
	}

	return s.String()
}

// Load switches to the top view and loads process info
func (m *TopViewModel) Load(model *Model, projectName string, service string) tea.Cmd {
	m.topService = service
	model.currentView = TopView
	model.loading = true
	return loadTop(model.dockerClient, projectName, service)
}

// HandleRefresh reloads the process info
func (m *TopViewModel) HandleRefresh(model *Model) tea.Cmd {
	model.loading = true
	return loadTop(model.dockerClient, model.projectName, m.topService)
}

// HandleBack returns to the compose process list view
func (m *TopViewModel) HandleBack(model *Model) tea.Cmd {
	model.currentView = ComposeProcessListView
	return loadProcesses(model.dockerClient, model.projectName, model.composeProcessListViewModel.showAll)
}

// Loaded updates the top output after loading
func (m *TopViewModel) Loaded(output string) {
	m.topOutput = output
}
