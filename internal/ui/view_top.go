package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tokuhirom/dcv/internal/docker"
)

// TopViewModel manages the state and rendering of the process info view
type TopViewModel struct {
	content string

	container docker.Container
}

// render renders the top view
func (m *TopViewModel) render(availableHeight int) string {
	var s strings.Builder

	if m.content == "" {
		s.WriteString("No process information available.\n")
	} else {
		// Display the raw top output
		lines := strings.Split(m.content, "\n")
		visibleHeight := availableHeight - 2

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
func (m *TopViewModel) Load(model *Model, container docker.Container) tea.Cmd {
	m.container = container
	model.SwitchView(TopView)
	return m.DoLoad(model)
}

// DoLoad reloads the process info
func (m *TopViewModel) DoLoad(model *Model) tea.Cmd {
	model.loading = true

	return func() tea.Msg {
		output, err := m.container.Top()
		return topLoadedMsg{
			output: string(output),
			err:    err,
		}
	}
}

// HandleBack returns to the compose process list view
func (m *TopViewModel) HandleBack(model *Model) tea.Cmd {
	model.SwitchToPreviousView()
	return nil
}

// Loaded updates the top output after loading
func (m *TopViewModel) Loaded(output string) {
	m.content = output
}

func (m *TopViewModel) Title() string {
	return fmt.Sprintf("Process Info: %s", m.container.Title())
}
