package ui

import (
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/tokuhirom/dcv/internal/docker"
)

// Docker container management commands

func (m *Model) CmdKill(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.useContainerAware(func(container *docker.Container) tea.Cmd {
		args := container.OperationArgs("kill")
		return m.commandExecutionViewModel.ExecuteCommand(m, true, args...) // kill is aggressive
	})
}

func (m *Model) CmdStop(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.useContainerAware(func(container *docker.Container) tea.Cmd {
		args := container.OperationArgs("stop")
		return m.commandExecutionViewModel.ExecuteCommand(m, true, args...) // stop is aggressive
	})
}

func (m *Model) CmdStart(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.useContainerAware(func(container *docker.Container) tea.Cmd {
		args := container.OperationArgs("start")
		return m.commandExecutionViewModel.ExecuteCommand(m, true, args...) // start is aggressive
	})
}

func (m *Model) CmdRestart(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.useContainerAware(func(container *docker.Container) tea.Cmd {
		args := container.OperationArgs("restart")
		return m.commandExecutionViewModel.ExecuteCommand(m, true, args...) // start is aggressive
	})
}

func (m *Model) CmdPause(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.useContainerAware(func(container *docker.Container) tea.Cmd {
		cmd := func() string {
			if container.GetState() == "paused" {
				return "unpause"
			} else {
				return "pause"
			}
		}()
		args := container.OperationArgs(cmd)
		return m.commandExecutionViewModel.ExecuteCommand(m, true, args...) // pause/unpause is aggressive
	})
}

func (m *Model) CmdDelete(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.isContainerAware() {
		return m, m.useContainerAware(func(container *docker.Container) tea.Cmd {
			args := container.OperationArgs("rm")
			return m.commandExecutionViewModel.ExecuteCommand(m, true, args...)
		})
	}

	switch m.currentView {
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
	case DindProcessListView:
		return m, m.dindProcessListViewModel.HandleToggleAll(m)
	case ImageListView:
		return m, m.imageListViewModel.HandleToggleAll(m)
	default:
		return m, nil
	}
}

func (m *Model) CmdInspect(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.isContainerAware() {
		return m, m.useContainerAware(func(container *docker.Container) tea.Cmd {
			return m.inspectViewModel.Inspect(m,
				container.Title(),
				func() ([]byte, error) {
					args := container.OperationArgs("inspect")
					return m.dockerClient.ExecuteCaptured(args...)
				})
		})
	}

	vm := m.GetCurrentViewModel()
	if vm == nil {
		slog.Info("Cannot get current view model for inspect command")
		return m, nil
	}

	if inspectAware, ok := vm.(HandleInspectAware); ok {
		// If the current view model implements HandleInspectAware, use its method
		return m, inspectAware.HandleInspect(m)
	}

	slog.Info("Current view model does not implement HandleInspectAware",
		slog.String("view", m.currentView.String()))
	return m, nil
}

func (m *Model) CmdTop(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.useContainerAware(func(container *docker.Container) tea.Cmd {
		return m.topViewModel.Load(m, container)
	})
}

func (m *Model) useContainerAware(cb func(container *docker.Container) tea.Cmd) tea.Cmd {
	container := m.getContainerByContainerAware()
	if container == nil {
		slog.Error("Failed to get selected container for top command")
		return nil
	}
	return cb(container)
}

func (m *Model) isContainerAware() bool {
	// if GetContainerAware, we can show top for containers
	// GetContainerAware is the interface that provides container-aware functionality
	vm := m.GetCurrentViewModel()
	if vm == nil {
		return false
	}

	if _, ok := vm.(GetContainerAware); ok {
		return true
	}
	return false
}

func (m *Model) getContainerByContainerAware() *docker.Container {
	// if GetContainerAware, we can show top for containers
	// GetContainerAware is the interface that provides container-aware functionality
	vm := m.GetCurrentViewModel()
	if vm == nil {
		return nil
	}

	if containerAware, ok := vm.(GetContainerAware); ok {
		container := containerAware.GetContainer(m)
		if container == nil {
			slog.Error("Failed to get selected container")
			return nil
		}
		return container
	}

	// this view model does not support container-aware functionality.
	return nil
}
