package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type TopViewModel struct {
}

func (m TopViewModel) Load(model *Model, projectName string, service string) tea.Cmd {
	model.topService = service
	model.currentView = TopView
	model.loading = true
	return loadTop(model.dockerClient, projectName, service)
}

func (m *Model) renderTopView(availableHeight int) string {
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
