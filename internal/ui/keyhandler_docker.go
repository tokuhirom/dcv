package ui

import (
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/tokuhirom/dcv/internal/docker"
)

// Docker container management commands

func (m *Model) CmdKill(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case ComposeProcessListView:
		return m, m.composeProcessListViewModel.HandleCommandExecution(m, "kill", true) // kill is aggressive
	case DockerContainerListView:
		return m, m.dockerContainerListViewModel.HandleKill(m)
	default:
		return m, nil
	}
}

func (m *Model) CmdStop(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case ComposeProcessListView:
		return m, m.composeProcessListViewModel.HandleCommandExecution(m, "stop", true) // stop is aggressive
	case DockerContainerListView:
		return m, m.dockerContainerListViewModel.HandleStop(m)
	default:
		return m, nil
	}
}

func (m *Model) CmdStart(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case ComposeProcessListView:
		return m, m.composeProcessListViewModel.HandleCommandExecution(m, "start", true) // start is aggressive
	case DockerContainerListView:
		return m, m.dockerContainerListViewModel.HandleStart(m)
	default:
		return m, nil
	}
}

func (m *Model) CmdRestart(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case ComposeProcessListView:
		return m, m.composeProcessListViewModel.HandleCommandExecution(m, "restart", true) // restart is aggressive
	case DockerContainerListView:
		return m, m.dockerContainerListViewModel.HandleRestart(m)
	default:
		return m, nil
	}
}

func (m *Model) CmdPause(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.useContainerAware(func(container docker.Container) tea.Cmd {
		args := container.PauseArgs()
		return m.commandExecutionViewModel.ExecuteCommand(m, true, args...) // pause/unpause is aggressive
	})
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
	case DindProcessListView:
		return m, m.dindProcessListViewModel.HandleInspect(m)
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

func (m *Model) CmdTop(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.useContainerAware(func(container docker.Container) tea.Cmd {
		return m.topViewModel.Load(m, container)
	})
}

func (m *Model) useContainerAware(cb func(container docker.Container) tea.Cmd) tea.Cmd {
	// if GetContainerAware, we can show top for containers
	// GetContainerAware is the interface that provides container-aware functionality
	vm := m.GetCurrentViewModel()
	if vm == nil {
		return nil
	}

	if containerAware, ok := vm.(GetContainerAware); ok {
		container := containerAware.GetContainer(m)
		if container == nil {
			slog.Error("Failed to get selected container for top command")
			return nil
		}
		return cb(container)
	}

	// this view model does not support container-aware functionality.
	return nil
}
