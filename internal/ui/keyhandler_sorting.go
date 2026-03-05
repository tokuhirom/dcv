package ui

import (
	tea "charm.land/bubbletea/v2"
)

// Sorting commands for Top and Stats views

func (m *Model) CmdSortByCPU(_ tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case TopView:
		m.topViewModel.HandleSortByCPU()
		return m, nil
	case StatsView:
		m.statsViewModel.HandleSortByCPU(m)
		return m, nil
	default:
		return m, nil
	}
}

func (m *Model) CmdSortByMem(_ tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case TopView:
		m.topViewModel.HandleSortByMem()
		return m, nil
	case StatsView:
		m.statsViewModel.HandleSortByMem(m)
		return m, nil
	default:
		return m, nil
	}
}

func (m *Model) CmdSortByPID(_ tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case TopView:
		m.topViewModel.HandleSortByPID()
		return m, nil
	default:
		return m, nil
	}
}

func (m *Model) CmdSortByTime(_ tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case TopView:
		m.topViewModel.HandleSortByTime()
		return m, nil
	default:
		return m, nil
	}
}

func (m *Model) CmdSortByCommand(_ tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case TopView:
		m.topViewModel.HandleSortByCommand()
		return m, nil
	case StatsView:
		m.statsViewModel.HandleSortByName(m)
		return m, nil
	default:
		return m, nil
	}
}

func (m *Model) CmdReverseSort(_ tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case TopView:
		m.topViewModel.HandleReverseSort()
		return m, nil
	case StatsView:
		m.statsViewModel.HandleReverseSort(m)
		return m, nil
	default:
		return m, nil
	}
}

func (m *Model) CmdToggleAutoRefresh(_ tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case TopView:
		m.topViewModel.HandleToggleAutoRefresh()
		if m.topViewModel.autoRefresh {
			return m, m.topViewModel.startAutoRefresh()
		}
		return m, nil
	case StatsView:
		m.statsViewModel.HandleToggleAutoRefresh()
		if m.statsViewModel.autoRefresh {
			return m, m.statsViewModel.startAutoRefresh()
		}
		return m, nil
	default:
		return m, nil
	}
}
