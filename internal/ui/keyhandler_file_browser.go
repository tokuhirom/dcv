package ui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/tokuhirom/dcv/internal/docker"
)

// File browser specific commands

// CmdFileBrowse is triggered when the user wants to browse files in a container
// It loads the file browser view model for the specified container.
func (m *Model) CmdFileBrowse(_ tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	return m, m.useContainerAware(func(container *docker.Container) tea.Cmd {
		return m.fileBrowserViewModel.LoadContainer(m, container)
	})
}

func (m *Model) CmdOpenFileOrDirectory(_ tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	return m, m.fileBrowserViewModel.HandleOpenFileOrDirectory(m)
}

func (m *Model) CmdGoToParentDirectory(_ tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	return m, m.fileBrowserViewModel.HandleGoToParentDirectory(m)
}
