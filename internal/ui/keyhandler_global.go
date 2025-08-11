package ui

import (
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
)

// Global commands that work across multiple views

func (m *Model) CmdQuit(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.quitConfirmation = true
	return m, nil
}

func (m *Model) CmdCommandMode(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Start command mode
	m.commandViewModel.Start()
	return m, nil
}

func (m *Model) CmdToggleNavbar(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Toggle navbar visibility
	m.navbarHidden = !m.navbarHidden
	return m, nil
}

// CmdRefresh sends a RefreshMsg to trigger a reload of the current view
func (m *Model) CmdRefresh(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, func() tea.Msg {
		return RefreshMsg{}
	}
}

// CmdCancel handles cancel/escape key for views that support it
func (m *Model) CmdCancel(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case CommandExecutionView:
		return m, m.commandExecutionViewModel.HandleCancel()
	case LogView:
		return m, m.logViewModel.HandleCancel()
	default:
		slog.Info("Cancel command not implemented for current view",
			slog.String("current_view", m.currentView.String()))
		return m, nil
	}
}

func (m *Model) CmdHelp(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.helpViewModel.Show(m, m.currentView)
}
