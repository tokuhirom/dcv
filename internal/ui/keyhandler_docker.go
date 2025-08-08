package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Docker container management commands

func (m *Model) CmdKill(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case ComposeProcessListView:
		return m, m.composeProcessListViewModel.HandleCommandExecution(m, "kill")
	case DockerContainerListView:
		return m, m.dockerContainerListViewModel.HandleKill(m)
	default:
		return m, nil
	}
}

func (m *Model) CmdStop(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case ComposeProcessListView:
		return m, m.composeProcessListViewModel.HandleCommandExecution(m, "stop")
	case DockerContainerListView:
		return m, m.dockerContainerListViewModel.HandleStop(m)
	default:
		return m, nil
	}
}

func (m *Model) CmdStart(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case ComposeProcessListView:
		return m, m.composeProcessListViewModel.HandleCommandExecution(m, "start")
	case DockerContainerListView:
		return m, m.dockerContainerListViewModel.HandleStart(m)
	default:
		return m, nil
	}
}

func (m *Model) CmdRestart(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case ComposeProcessListView:
		return m, m.composeProcessListViewModel.HandleCommandExecution(m, "restart")
	case DockerContainerListView:
		return m, m.dockerContainerListViewModel.HandleRestart(m)
	default:
		return m, nil
	}
}

func (m *Model) CmdPause(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case ComposeProcessListView:
		// Check if selected container is paused
		if m.composeProcessListViewModel.selectedContainer < len(m.composeProcessListViewModel.composeContainers) {
			selected := m.composeProcessListViewModel.composeContainers[m.composeProcessListViewModel.selectedContainer]
			var args []string
			if selected.State == "paused" {
				args = []string{"unpause", selected.ID}
			} else {
				args = []string{"pause", selected.ID}
			}
			return m, m.commandExecutionViewModel.ExecuteCommand(m, args...)
		}
		return m, nil
	case DockerContainerListView:
		// Docker container list pause/unpause support
		if m.dockerContainerListViewModel.selectedDockerContainer < len(m.dockerContainerListViewModel.dockerContainers) {
			selected := m.dockerContainerListViewModel.dockerContainers[m.dockerContainerListViewModel.selectedDockerContainer]
			var args []string
			if strings.Contains(selected.Status, "Paused") {
				args = []string{"unpause", selected.ID}
			} else {
				args = []string{"pause", selected.ID}
			}
			return m, m.commandExecutionViewModel.ExecuteCommand(m, args...)
		}
		return m, nil
	default:
		return m, nil
	}
}

func (m *Model) CmdDelete(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case ComposeProcessListView:
		return m, m.composeProcessListViewModel.HandleRemove(m)
	case DockerContainerListView:
		return m, m.dockerContainerListViewModel.HandleRemove(m)
	case ImageListView:
		return m, m.imageListViewModel.HandleDelete(m)
	case NetworkListView:
		return m, m.networkListViewModel.HandleDelete(m)
	case VolumeListView:
		return m, m.volumeListViewModel.HandleDelete(m, false)
	default:
		return m, nil
	}
}

func (m *Model) CmdToggleAll(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case ComposeProcessListView:
		return m, m.composeProcessListViewModel.HandleToggleAll(m)
	case DockerContainerListView:
		return m, m.dockerContainerListViewModel.HandleToggleAll(m)
	case ImageListView:
		return m, m.imageListViewModel.HandleToggleAll(m)
	default:
		return m, nil
	}
}

func (m *Model) CmdInspect(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case ComposeProcessListView:
		return m, m.composeProcessListViewModel.HandleInspect(m)
	case DockerContainerListView:
		return m, m.dockerContainerListViewModel.HandleInspect(m)
	case ImageListView:
		return m, m.imageListViewModel.HandleInspect(m)
	case NetworkListView:
		return m, m.networkListViewModel.HandleInspect(m)
	case VolumeListView:
		return m, m.volumeListViewModel.HandleInspect(m)
	default:
		return m, nil
	}
}
