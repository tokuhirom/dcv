package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/tokuhirom/dcv/internal/docker"
)

// View switching commands

func (m *Model) CmdPS(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.dockerContainerListViewModel.Show(m)
}

func (m *Model) CmdImages(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.imageListViewModel.Show(m)
}

func (m *Model) CmdNetworkLs(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.networkListViewModel.Show(m)
}

func (m *Model) CmdVolumeLs(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.volumeListViewModel.Show(m)
}

func (m *Model) CmdLog(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.useContainerAware(func(container *docker.Container) tea.Cmd {
		return m.logViewModel.StreamContainerLogs(m, container)
	})
}

// CmdShell executes a shell in the selected container
// It defaults to /bin/sh, which is commonly available in containers.
// If the container does not have /bin/sh, it will fail gracefully.
func (m *Model) CmdShell(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.useContainerAware(func(container *docker.Container) tea.Cmd {
		// Default to /bin/sh as it's most commonly available
		return func() tea.Msg {
			shell := "/bin/sh"
			args := container.InteractiveExecArgs(shell)
			return launchShellMsg{
				container: container,
				args:      args,
				shell:     shell,
			}
		}
	})
}

func (m *Model) CmdShowActions(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.useContainerAware(func(container *docker.Container) tea.Cmd {
		// Initialize the action view with the selected container
		m.commandActionViewModel.Initialize(container)
		m.SwitchView(CommandActionView)
		return nil
	})
}

func (m *Model) CmdSelectAction(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case CommandActionView:
		return m, m.commandActionViewModel.HandleSelect(m)
	default:
		return m, nil
	}
}

func (m *Model) CmdShowComposeProjectActions(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case ComposeProjectListView:
		if m.composeProjectListViewModel.Cursor < len(m.composeProjectListViewModel.projects) {
			project := m.composeProjectListViewModel.projects[m.composeProjectListViewModel.Cursor]
			m.composeProjectActionViewModel.Initialize(&project)
			m.SwitchView(ComposeProjectActionView)
			return m, nil
		}
	}
	return m, nil
}

func (m *Model) CmdSelectComposeProjectAction(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case ComposeProjectActionView:
		return m, m.composeProjectActionViewModel.HandleSelect(m)
	default:
		return m, nil
	}
}

func (m *Model) CmdShowFileActions(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case FileBrowserView:
		if m.fileBrowserViewModel.Cursor < len(m.fileBrowserViewModel.containerFiles) {
			file := m.fileBrowserViewModel.containerFiles[m.fileBrowserViewModel.Cursor]
			container := m.fileBrowserViewModel.browsingContainer
			path := m.fileBrowserViewModel.currentPath

			// Initialize the file browser action view
			m.fileBrowserActionViewModel.Initialize(&file, container, path)
			m.SwitchView(FileBrowserActionView)
			return m, nil
		}
	}
	return m, nil
}

func (m *Model) CmdSelectFileAction(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case FileBrowserActionView:
		return m, m.fileBrowserActionViewModel.HandleSelect(m)
	default:
		return m, nil
	}
}
