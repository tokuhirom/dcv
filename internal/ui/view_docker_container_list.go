package ui

import (
	"log/slog"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tokuhirom/dcv/internal/docker"

	"github.com/tokuhirom/dcv/internal/models"
)

type DockerContainerListViewModel struct {
	dockerContainers        []models.DockerContainer
	selectedDockerContainer int
	showAll                 bool
}

type ColumnMap struct {
	containerID table.Column
	image       table.Column
	status      table.Column
	ports       table.Column
	names       table.Column
}

func NewColumnMap(model *Model) ColumnMap {
	sideMargin := 2 * 2    // 2 for left and right padding
	cellMargin := 2        // 2 for cell margin
	containerIDWidth := 12 // Fixed width for container ID
	widthPerColumn := (model.width - containerIDWidth - cellMargin*4 - sideMargin) / 4

	return ColumnMap{
		containerID: table.Column{Title: "CONTAINER ID", Width: containerIDWidth},
		image:       table.Column{Title: "IMAGE", Width: widthPerColumn},
		status:      table.Column{Title: "STATUS", Width: widthPerColumn},
		ports:       table.Column{Title: "PORTS", Width: widthPerColumn},
		names:       table.Column{Title: "NAMES", Width: widthPerColumn},
	}
}

func (c *ColumnMap) ToArray() []table.Column {
	return []table.Column{
		c.containerID,
		c.image,
		c.status,
		c.ports,
		c.names,
	}
}

func (m *DockerContainerListViewModel) renderDockerList(model *Model, availableHeight int) string {
	// Container list
	if len(m.dockerContainers) == 0 {
		var s strings.Builder
		s.WriteString("\nNo containers found.\n")
		return s.String()
	}

	columns := NewColumnMap(model)

	rows := make([]table.Row, 0, len(m.dockerContainers))
	for _, container := range m.dockerContainers {
		// Truncate container ID
		id := container.ID
		if len(id) > 12 { // shorten ID to 12 characters
			id = id[:12]
		}

		// Truncate image name
		image := container.Image

		// Status with color
		status := container.Status

		// Truncate ports
		ports := lipgloss.NewStyle().MaxWidth(columns.ports.Width).Render(container.Ports)

		name := container.Names
		if container.IsDind() {
			name = dindStyle.Render("⬢ ") + name
		}

		rows = append(rows, table.Row{id, image, status, ports, name})
	}

	return RenderTable(columns.ToArray(), rows, availableHeight, m.selectedDockerContainer)
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
		return model.commandExecutionViewModel.ExecuteCommand(model, true, "kill", container.ID) // kill is aggressive
	}
	return nil
}

func (m *DockerContainerListViewModel) HandleStop(model *Model) tea.Cmd {
	if m.selectedDockerContainer < len(m.dockerContainers) {
		container := m.dockerContainers[m.selectedDockerContainer]
		return model.commandExecutionViewModel.ExecuteCommand(model, true, "stop", container.ID) // stop is aggressive
	}
	return nil
}

func (m *DockerContainerListViewModel) HandleStart(model *Model) tea.Cmd {
	if m.selectedDockerContainer < len(m.dockerContainers) {
		container := m.dockerContainers[m.selectedDockerContainer]
		return model.commandExecutionViewModel.ExecuteCommand(model, true, "start", container.ID) // start is aggressive
	}
	return nil
}

func (m *DockerContainerListViewModel) HandleRestart(model *Model) tea.Cmd {
	if m.selectedDockerContainer < len(m.dockerContainers) {
		container := m.dockerContainers[m.selectedDockerContainer]
		return model.commandExecutionViewModel.ExecuteCommand(model, true, "restart", container.ID) // restart is aggressive
	}
	return nil
}

func (m *DockerContainerListViewModel) HandleRemove(model *Model) tea.Cmd {
	// Delete the selected Docker container
	if m.selectedDockerContainer < len(m.dockerContainers) {
		container := m.dockerContainers[m.selectedDockerContainer]
		return model.commandExecutionViewModel.ExecuteCommand(model, true, "rm", container.ID) // rm is aggressive
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
		return func() tea.Msg {
			return executeCommandMsg{
				containerID: container.ID,
				command:     []string{"/bin/sh"},
			}
		}
	}
	return nil
}

func (m *DockerContainerListViewModel) GetContainer(model *Model) docker.Container {
	// Get the selected Docker container
	if m.selectedDockerContainer < len(m.dockerContainers) {
		container := m.dockerContainers[m.selectedDockerContainer]
		return docker.NewContainer(model.dockerClient, container.ID, container.Names, container.Names)
	}
	return nil
}

func (m *DockerContainerListViewModel) HandleInspect(model *Model) tea.Cmd {
	container := m.GetContainer(model)
	if container == nil {
		slog.Error("Failed to get selected container for inspection")
		return nil
	}

	return model.inspectViewModel.Inspect(model,
		"container "+container.GetName(),
		func() ([]byte, error) {
			return container.Inspect()
		})
}

func (m *DockerContainerListViewModel) Show(model *Model) tea.Cmd {
	model.SwitchView(DockerContainerListView)
	return m.DoLoad(model)
}

func (m *DockerContainerListViewModel) HandleBack(model *Model) tea.Cmd {
	model.SwitchToPreviousView()
	return nil
}

func (m *DockerContainerListViewModel) HandleToggleAll(model *Model) tea.Cmd {
	m.showAll = !m.showAll
	return m.DoLoad(model)
}

func (m *DockerContainerListViewModel) DoLoad(model *Model) tea.Cmd {
	model.loading = true
	return func() tea.Msg {
		containers, err := model.dockerClient.ListContainers(m.showAll)
		return dockerContainersLoadedMsg{
			containers: containers,
			err:        err,
		}
	}
}

func (m *DockerContainerListViewModel) HandleDindProcessList(model *Model) tea.Cmd {
	container := m.GetContainer(model)
	if container == nil {
		slog.Error("Failed to get selected container for DinD process list")
		return nil
	}

	return model.dindProcessListViewModel.Load(model, container)
}

func (m *DockerContainerListViewModel) HandleTop(model *Model) tea.Cmd {
	container := m.GetContainer(model)
	if container == nil {
		slog.Error("Failed to get selected container for DinD process list")
		return nil
	}

	return model.topViewModel.Load(model, container)
}
