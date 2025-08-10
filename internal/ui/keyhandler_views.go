package ui

import tea "github.com/charmbracelet/bubbletea"

// View switching commands

func (m *Model) CmdPS(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.dockerContainerListViewModel.Show(m)
}

func (m *Model) CmdImages(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.imageListViewModel.Show(m)
}

func (m *Model) CmdNetworkLs(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.networkListViewModel.Show(m)
}

func (m *Model) CmdVolumeLs(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.volumeListViewModel.Show(m)
}

func (m *Model) CmdLog(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case ComposeProcessListView:
		return m, m.composeProcessListViewModel.HandleLog(m)
	case DockerContainerListView:
		return m, m.dockerContainerListViewModel.HandleLog(m)
	case DindProcessListView:
		return m, m.dindProcessListViewModel.HandleLog(m)
	default:
		return m, nil
	}
}

func (m *Model) CmdShell(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case ComposeProcessListView:
		return m, m.composeProcessListViewModel.HandleShell()
	case DockerContainerListView:
		return m, m.dockerContainerListViewModel.HandleShell(m)
	default:
		return m, nil
	}
}

func (m *Model) CmdShellNewWindow(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case ComposeProcessListView:
		return m, m.composeProcessListViewModel.HandleShellNewWindow()
	case DockerContainerListView:
		return m, m.dockerContainerListViewModel.HandleShellNewWindow(m)
	default:
		return m, nil
	}
}
