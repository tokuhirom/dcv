package ui

import (
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
)

// Navigation commands (up, down, go to start/end, back)

func (m *Model) CmdUp(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case NetworkListView:
		return m, m.networkListViewModel.HandleUp()
	case HelpView:
		return m, m.helpViewModel.HandleUp()
	case DindProcessListView:
		return m, m.dindProcessListViewModel.HandleUp()
	case VolumeListView:
		return m, m.volumeListViewModel.HandleUp()
	case ImageListView:
		return m, m.imageListViewModel.HandleUp()
	case FileContentView:
		return m, m.fileContentViewModel.HandleUp()
	case FileBrowserView:
		return m, m.fileBrowserViewModel.HandleUp()
	case LogView:
		return m, m.logViewModel.HandleUp()
	case InspectView:
		return m, m.inspectViewModel.HandleUp()
	case DockerContainerListView:
		return m, m.dockerContainerListViewModel.HandleUp(m)
	case ComposeProjectListView:
		return m, m.composeProjectListViewModel.HandleUp(m)
	case ComposeProcessListView:
		return m, m.composeProcessListViewModel.HandleUp()
	case CommandExecutionView:
		return m, m.commandExecutionViewModel.HandleUp()
	default:
		slog.Info("Unhandled key up in current view",
			slog.String("view", m.currentView.String()))
		return m, nil
	}
}

func (m *Model) CmdDown(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	slog.Info("CmdDown called",
		slog.String("view", m.currentView.String()),
		slog.Int("selectedContainer", m.composeProcessListViewModel.selectedContainer))

	switch m.currentView {
	case NetworkListView:
		return m, m.networkListViewModel.HandleDown()
	case HelpView:
		return m, m.helpViewModel.HandleDown(m)
	case DindProcessListView:
		return m, m.dindProcessListViewModel.HandleDown()
	case VolumeListView:
		return m, m.volumeListViewModel.HandleDown()
	case ImageListView:
		return m, m.imageListViewModel.HandleDown()
	case FileContentView:
		return m, m.fileContentViewModel.HandleDown(m.Height)
	case FileBrowserView:
		return m, m.fileBrowserViewModel.HandleDown()
	case LogView:
		return m, m.logViewModel.HandleDown(m)
	case InspectView:
		return m, m.inspectViewModel.HandleDown(m)
	case DockerContainerListView:
		return m, m.dockerContainerListViewModel.HandleDown(m)
	case ComposeProjectListView:
		return m, m.composeProjectListViewModel.HandleDown(m)
	case ComposeProcessListView:
		return m, m.composeProcessListViewModel.HandleDown()
	case CommandExecutionView:
		return m, m.commandExecutionViewModel.HandleDown(m)
	default:
		slog.Info("Unhandled key down in current view",
			slog.String("view", m.currentView.String()))
		return m, nil
	}
}

func (m *Model) CmdGoToEnd(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case LogView:
		return m, m.logViewModel.HandleGoToEnd(m)
	case FileContentView:
		return m, m.fileContentViewModel.HandleGoToEnd(m.Height)
	case InspectView:
		return m, m.inspectViewModel.HandleGoToEnd(m)
	case CommandExecutionView:
		return m, m.commandExecutionViewModel.HandleGoToEnd(m)
	default:
		slog.Info("GoToEnd not supported in current view",
			slog.String("view", m.currentView.String()))
		return m, nil
	}
}

func (m *Model) CmdGoToStart(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case LogView:
		return m, m.logViewModel.HandleGoToStart()
	case FileContentView:
		return m, m.fileContentViewModel.HandleGoToStart()
	case InspectView:
		return m, m.inspectViewModel.HandleGoToStart()
	case CommandExecutionView:
		return m, m.commandExecutionViewModel.HandleGoToStart()
	default:
		slog.Info("GoToStart not supported in current view",
			slog.String("view", m.currentView.String()))
		return m, nil
	}
}

func (m *Model) CmdPageUp(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case LogView:
		return m, m.logViewModel.HandlePageUp(m)
	case InspectView:
		return m, m.inspectViewModel.HandlePageUp(m)
	default:
		slog.Info("PageUp not supported in current view",
			slog.String("view", m.currentView.String()))
		return m, nil
	}
}

func (m *Model) CmdPageDown(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case LogView:
		return m, m.logViewModel.HandlePageDown(m)
	case InspectView:
		return m, m.inspectViewModel.HandlePageDown(m)
	default:
		slog.Info("PageDown not supported in current view",
			slog.String("view", m.currentView.String()))
		return m, nil
	}
}

func (m *Model) CmdBack(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case LogView:
		return m, m.logViewModel.HandleBack(m)
	case FileContentView:
		return m, m.fileContentViewModel.HandleBack(m)
	case FileBrowserView:
		return m, m.fileBrowserViewModel.HandleBack(m)
	case DindProcessListView:
		return m, m.dindProcessListViewModel.HandleBack(m)
	case TopView:
		return m, m.topViewModel.HandleBack(m)
	case StatsView:
		return m, m.statsViewModel.HandleBack(m)
	case DockerContainerListView:
		return m, m.dockerContainerListViewModel.HandleBack(m)
	case ImageListView:
		return m, m.imageListViewModel.HandleBack(m)
	case InspectView:
		return m, m.inspectViewModel.HandleBack(m)
	case HelpView:
		return m, m.helpViewModel.HandleBack(m)
	case NetworkListView:
		return m, m.networkListViewModel.HandleBack(m)
	case VolumeListView:
		return m, m.volumeListViewModel.HandleBack(m)
	case CommandExecutionView:
		return m, m.commandExecutionViewModel.HandleBack(m)
	case ComposeProcessListView:
		// Should not happen in ComposeProcessListView, but handle it gracefully
		// This is the main view, nowhere to go back to
		return m, nil
	default:
		slog.Info("Back not supported in current view",
			slog.String("view", m.currentView.String()))
		return m, nil
	}
}
