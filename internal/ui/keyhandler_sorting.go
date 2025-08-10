package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Sorting commands for Top and Stats views

func (m *Model) CmdSortByCPU(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case TopView:
		m.topViewModel.HandleSortByCPU()
		return m, nil
	case StatsView:
		m.statsViewModel.HandleSortByCPU()
		return m, nil
	default:
		return m, nil
	}
}

func (m *Model) CmdSortByMem(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case TopView:
		m.topViewModel.HandleSortByMem()
		return m, nil
	case StatsView:
		m.statsViewModel.HandleSortByMem()
		return m, nil
	default:
		return m, nil
	}
}

func (m *Model) CmdSortByPID(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case TopView:
		m.topViewModel.HandleSortByPID()
		return m, nil
	default:
		return m, nil
	}
}

func (m *Model) CmdSortByTime(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case TopView:
		m.topViewModel.HandleSortByTime()
		return m, nil
	default:
		return m, nil
	}
}

func (m *Model) CmdSortByCommand(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case TopView:
		m.topViewModel.HandleSortByCommand()
		return m, nil
	case StatsView:
		m.statsViewModel.HandleSortByName()
		return m, nil
	default:
		return m, nil
	}
}

func (m *Model) CmdReverseSort(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case TopView:
		m.topViewModel.HandleReverseSort()
		return m, nil
	case StatsView:
		m.statsViewModel.HandleReverseSort()
		return m, nil
	default:
		return m, nil
	}
}
