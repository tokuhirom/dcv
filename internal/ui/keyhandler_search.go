package ui

import (
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
)

// Search and filter commands

func (m *Model) CmdSearch(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case LogView:
		return m, m.logViewModel.HandleSearch()
	case InspectView:
		return m, m.inspectViewModel.HandleSearch()
	default:
		// Check if current view supports container search
		vm := m.GetCurrentViewModel()
		if searchable, ok := vm.(ContainerSearchAware); ok {
			searchable.StartSearch()
			return m, nil
		}

		slog.Info("Search not supported in current view",
			slog.String("view", m.currentView.String()))
		return m, nil
	}
}

func (m *Model) CmdNextSearchResult(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case LogView:
		return m, m.logViewModel.HandleNextSearchResult(m)
	case InspectView:
		return m, m.inspectViewModel.HandleNextSearchResult(m)
	default:
		// Check if current view supports container search
		vm := m.GetCurrentViewModel()
		if searchable, ok := vm.(ContainerSearchAware); ok {
			searchable.HandleNextSearchResult()
			return m, nil
		}

		slog.Info("Next search result not supported in current view",
			slog.String("view", m.currentView.String()))
		return m, nil
	}
}

func (m *Model) CmdPrevSearchResult(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case LogView:
		return m, m.logViewModel.HandlePrevSearchResult(m)
	case InspectView:
		return m, m.inspectViewModel.HandlePrevSearchResult(m)
	default:
		// Check if current view supports container search
		vm := m.GetCurrentViewModel()
		if searchable, ok := vm.(ContainerSearchAware); ok {
			searchable.HandlePrevSearchResult()
			return m, nil
		}

		slog.Info("Previous search result not supported in current view",
			slog.String("view", m.currentView.String()))
		return m, nil
	}
}

func (m *Model) CmdFilter(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case LogView:
		return m, m.logViewModel.HandleFilter()
	default:
		slog.Info("Filter not supported in current view",
			slog.String("view", m.currentView.String()))
		return m, nil
	}
}
