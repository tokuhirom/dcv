package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

	"github.com/tokuhirom/dcv/internal/docker"

	"github.com/tokuhirom/dcv/internal/models"
)

type DockerContainerListViewModel struct {
	dockerContainers        []models.DockerContainer
	selectedDockerContainer int
	showAll                 bool
}

func (m *DockerContainerListViewModel) renderDockerList(model *Model, availableHeight int) string {
	var s strings.Builder

	// Container list
	if len(m.dockerContainers) == 0 {
		s.WriteString("\nNo containers found.\n")
		return s.String()
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			baseStyle := normalStyle
			if row == m.selectedDockerContainer {
				baseStyle = selectedStyle
			}
			return baseStyle
		}).
		Headers("CONTAINER ID", "IMAGE", "STATUS", "PORTS", "NAMES").
		Height(availableHeight).
		Width(model.width)

	for _, container := range m.dockerContainers {
		// Truncate container ID
		id := container.ID
		if len(id) > 12 {
			id = id[:12]
		}

		// Truncate image name
		image := container.Image
		if len(image) > 30 {
			image = image[:27] + "..."
		}

		// Status with color
		status := container.Status
		if strings.Contains(status, "Up") || strings.Contains(status, "running") {
			status = statusUpStyle.Render(status)
		} else {
			status = statusDownStyle.Render(status)
		}

		// Truncate ports
		ports := container.Ports
		if len(ports) > 30 {
			ports = ports[:27] + "..."
		}

		name := container.Names
		if container.IsDind() {
			name = dindStyle.Render("⬢ ") + name
		}

		t.Row(id, image, status, ports, name)
	}

	t.Offset(m.selectedDockerContainer)

	s.WriteString(t.Render() + "\n")

	return s.String()
}

func (m *DockerContainerListViewModel) HandleUp(_ *Model) tea.Cmd {
	if m.selectedDockerContainer > 0 {
		m.selectedDockerContainer--
	}
	return nil
}

func (m *DockerContainerListViewModel) HandleDown(*Model) tea.Cmd {
	if m.selectedDockerContainer < len(m.dockerContainers)-1 {
		m.selectedDockerContainer++
	}
	return nil
}

func (m *DockerContainerListViewModel) HandleLog(model *Model) tea.Cmd {
	if m.selectedDockerContainer < len(m.dockerContainers) {
		container := m.dockerContainers[m.selectedDockerContainer]
		return model.logViewModel.StreamContainerLogs(model, container)
	}
	return nil
}

func (m *DockerContainerListViewModel) HandleKill(model *Model) tea.Cmd {
	if m.selectedDockerContainer < len(m.dockerContainers) {
		container := m.dockerContainers[m.selectedDockerContainer]
		return model.commandExecutionViewModel.ExecuteContainerCommand(model, DockerContainerListView, container.ID, "kill")
	}
	return nil
}

func (m *DockerContainerListViewModel) HandleStop(model *Model) tea.Cmd {
	if m.selectedDockerContainer < len(m.dockerContainers) {
		container := m.dockerContainers[m.selectedDockerContainer]
		return model.commandExecutionViewModel.ExecuteContainerCommand(model, DockerContainerListView, container.ID, "stop")
	}
	return nil
}

func (m *DockerContainerListViewModel) HandleStart(model *Model) tea.Cmd {
	if m.selectedDockerContainer < len(m.dockerContainers) {
		container := m.dockerContainers[m.selectedDockerContainer]
		return model.commandExecutionViewModel.ExecuteContainerCommand(model, DockerContainerListView, container.ID, "start")
	}
	return nil
}

func (m *DockerContainerListViewModel) HandleRestart(model *Model) tea.Cmd {
	if m.selectedDockerContainer < len(m.dockerContainers) {
		container := m.dockerContainers[m.selectedDockerContainer]
		return model.commandExecutionViewModel.ExecuteContainerCommand(model, DockerContainerListView, container.ID, "restart")
	}
	return nil
}

func (m *DockerContainerListViewModel) HandleRemove(model *Model) tea.Cmd {
	// Delete the selected Docker container
	if m.selectedDockerContainer < len(m.dockerContainers) {
		container := m.dockerContainers[m.selectedDockerContainer]
		return model.commandExecutionViewModel.ExecuteContainerCommand(model, DockerContainerListView, container.ID, "rm")
	}
	return nil
}

func (m *DockerContainerListViewModel) HandleFileBrowse(model *Model) tea.Cmd {
	if m.selectedDockerContainer < len(m.dockerContainers) {
		container := m.dockerContainers[m.selectedDockerContainer]
		return model.fileBrowserViewModel.Load(model, container.ID, container.Names)
	}
	return nil
}

func (m *DockerContainerListViewModel) HandleShell(model *Model) tea.Cmd {
	// Execute shell in the selected Docker container
	if m.selectedDockerContainer < len(m.dockerContainers) {
		container := m.dockerContainers[m.selectedDockerContainer]
		// Default to /bin/sh as it's most commonly available
		return executeInteractiveCommand(container.ID, []string{"/bin/sh"})
	}
	return nil
}

func (m *DockerContainerListViewModel) HandleInspect(model *Model) tea.Cmd {
	// Inspect the selected Docker container
	if m.selectedDockerContainer < len(m.dockerContainers) {
		container := m.dockerContainers[m.selectedDockerContainer]
		return model.inspectViewModel.InspectContainer(model, container.ID)
	}
	return nil
}

func (m *DockerContainerListViewModel) Show(model *Model) tea.Cmd {
	model.SwitchView(DockerContainerListView)
	model.loading = true
	return loadDockerContainers(model.dockerClient, m.showAll)
}

func (m *DockerContainerListViewModel) HandleBack(model *Model) tea.Cmd {
	model.SwitchToPreviousView()
	return nil
}

func (m *DockerContainerListViewModel) HandleToggleAll(model *Model) tea.Cmd {
	m.showAll = !m.showAll
	model.loading = true
	return loadDockerContainers(model.dockerClient, m.showAll)
}

func (m *DockerContainerListViewModel) HandleDindProcessList(model *Model) tea.Cmd {
	if m.selectedDockerContainer < len(m.dockerContainers) {
		container := m.dockerContainers[m.selectedDockerContainer]
		if container.IsDind() {
			return model.dindProcessListViewModel.Load(model, container)
		}
	}
	return nil
}

func loadDockerContainers(client *docker.Client, showAll bool) tea.Cmd {
	return func() tea.Msg {
		containers, err := client.ListContainers(showAll)
		return dockerContainersLoadedMsg{
			containers: containers,
			err:        err,
		}
	}
}
