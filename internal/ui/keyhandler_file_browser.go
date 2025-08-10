package ui

import tea "github.com/charmbracelet/bubbletea"

// File browser specific commands

func (m *Model) CmdFileBrowse(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case ComposeProcessListView:
		return m, m.composeProcessListViewModel.HandleFileBrowse(m)
	case DockerContainerListView:
		return m, m.dockerContainerListViewModel.HandleFileBrowse(m)
	case DindProcessListView:
		return m, m.dindProcessListViewModel.HandleFileBrowse(m)
	default:
		return m, nil
	}
}

func (m *Model) CmdOpenFileOrDirectory(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.fileBrowserViewModel.HandleOpenFileOrDirectory(m)
}

func (m *Model) CmdGoToParentDirectory(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.fileBrowserViewModel.HandleGoToParentDirectory(m)
}
