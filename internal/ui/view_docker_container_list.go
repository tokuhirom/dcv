package ui

import (
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
		if strings.Contains(status, "Up") || strings.Contains(status, "running") {
			status = statusUpStyle.MaxWidth(columns.status.Width).Inline(true).Render(status)
		} else {
			status = statusDownStyle.MaxWidth(columns.status.Width).Inline(true).Render(status)
		}

		// Truncate ports
		ports := lipgloss.NewStyle().MaxWidth(columns.ports.Width).Render(container.Ports)

		name := container.Names
		if container.IsDind() {
			name = dindStyle.Render("⬢ ") + name
		}

		rows = append(rows, table.Row{id, image, status, ports, name})
	}

	t := table.New(
		table.WithColumns(columns.ToArray()),
		table.WithRows(rows),
		table.WithHeight(availableHeight-1),
		table.WithFocused(true),
	)

	// Apply styles
	styles := table.DefaultStyles()
	styles.Header = styles.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	styles.Selected = selectedStyle
	styles.Cell = styles.Cell.
		BorderForeground(lipgloss.Color("240"))
	t.SetStyles(styles)

	// Set cursor position
	if m.selectedDockerContainer < len(rows) {
		t.MoveDown(m.selectedDockerContainer)
	}

	return t.View()
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
		return model.commandExecutionViewModel.ExecuteCommand(model, "kill", container.ID)
	}
	return nil
}

func (m *DockerContainerListViewModel) HandleStop(model *Model) tea.Cmd {
	if m.selectedDockerContainer < len(m.dockerContainers) {
		container := m.dockerContainers[m.selectedDockerContainer]
		return model.commandExecutionViewModel.ExecuteCommand(model, "stop", container.ID)
	}
	return nil
}

func (m *DockerContainerListViewModel) HandleStart(model *Model) tea.Cmd {
	if m.selectedDockerContainer < len(m.dockerContainers) {
		container := m.dockerContainers[m.selectedDockerContainer]
		return model.commandExecutionViewModel.ExecuteCommand(model, "start", container.ID)
	}
	return nil
}

func (m *DockerContainerListViewModel) HandleRestart(model *Model) tea.Cmd {
	if m.selectedDockerContainer < len(m.dockerContainers) {
		container := m.dockerContainers[m.selectedDockerContainer]
		return model.commandExecutionViewModel.ExecuteCommand(model, "restart", container.ID)
	}
	return nil
}

func (m *DockerContainerListViewModel) HandleRemove(model *Model) tea.Cmd {
	// Delete the selected Docker container
	if m.selectedDockerContainer < len(m.dockerContainers) {
		container := m.dockerContainers[m.selectedDockerContainer]
		return model.commandExecutionViewModel.ExecuteCommand(model, "rm", container.ID)
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
