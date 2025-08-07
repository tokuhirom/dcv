package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Volume navigation handlers
func (m *Model) SelectUpVolume(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.volumeListViewModel.HandleSelectUp()
}

func (m *Model) SelectDownVolume(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.volumeListViewModel.HandleSelectDown()
}

// View change handlers
func (m *Model) ShowVolumeList(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.volumeListViewModel.Show(m)
}

func (m *Model) DeleteVolume(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.volumeListViewModel.HandleDelete(m)
}

func (m *Model) ForceDeleteVolume(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.volumeListViewModel.HandleForceDelete(m)
}
