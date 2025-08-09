package ui

import tea "github.com/charmbracelet/bubbletea"

// Docker Compose specific commands

func (m *Model) CmdComposeUp(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.composeProcessListViewModel.HandleCommandExecution(m, "up")
}

func (m *Model) CmdComposeDown(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.composeProcessListViewModel.HandleCommandExecution(m, "down")
}

func (m *Model) CmdComposeLS(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.SwitchView(ComposeProjectListView)
	m.loading = true
	return m, loadProjects(m.dockerClient)
}

func (m *Model) CmdSelectProject(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case ComposeProjectListView:
		return m, m.composeProjectListViewModel.HandleSelectProject(m)
	default:
		return m, nil
	}
}

func (m *Model) CmdTop(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case ComposeProcessListView:
		return m, m.composeProcessListViewModel.HandleTop(m)
	default:
		return m, nil
	}
}

func (m *Model) CmdStats(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.statsViewModel.Show(m)
}

func (m *Model) CmdDind(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case ComposeProcessListView:
		return m, m.composeProcessListViewModel.HandleDindProcessList(m)
	default:
		return m, nil
	}
}
