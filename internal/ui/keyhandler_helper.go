package ui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/tokuhirom/dcv/internal/docker"
)

// CmdInjectHelper is triggered when the user wants to inject a helper script into a container
func (m *Model) CmdInjectHelper(_ tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	return m, m.useContainerAware(func(container *docker.Container) tea.Cmd {
		return m.helperInjectorViewModel.HandleInjectHelper(m, container)
	})
}
