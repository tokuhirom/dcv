package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/tokuhirom/dcv/internal/docker"
)

// CmdInjectHelper is triggered when the user wants to inject a helper script into a container
func (m *Model) CmdInjectHelper(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.useContainerAware(func(container *docker.Container) tea.Cmd {
		return m.helperInjectorViewModel.HandleInjectHelper(m, container)
	})
}
