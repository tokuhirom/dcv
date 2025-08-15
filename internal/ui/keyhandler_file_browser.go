package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/tokuhirom/dcv/internal/docker"
)

// File browser specific commands

// CmdFileBrowse is triggered when the user wants to browse files in a container
// It loads the file browser view model for the specified container.
func (m *Model) CmdFileBrowse(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.useContainerAware(func(container *docker.Container) tea.Cmd {
		return m.fileBrowserViewModel.LoadContainer(m, container)
	})
}

func (m *Model) CmdOpenFileOrDirectory(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.fileBrowserViewModel.HandleOpenFileOrDirectory(m)
}

func (m *Model) CmdGoToParentDirectory(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.fileBrowserViewModel.HandleGoToParentDirectory(m)
}

func (m *Model) CmdInjectHelper(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.fileBrowserViewModel.HandleInjectHelper(m)
}

func (m *Model) CmdInjectHelperDind(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	// For dind view, inject helper into the dind container itself
	return m, m.dindProcessListViewModel.HandleInjectHelper(m)
}
