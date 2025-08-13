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

type dockerContainersLoadedMsg struct {
	containers []models.DockerContainer
	err        error
}

var _ ContainerAware = (*DockerContainerListViewModel)(nil)
var _ UpdateAware = (*DockerContainerListViewModel)(nil)
var _ ContainerSearchAware = (*DockerContainerListViewModel)(nil)

type DockerContainerListViewModel struct {
	ContainerSearchViewModel
	TableViewModel
	dockerContainers []models.DockerContainer
	showAll          bool
}

func (m *DockerContainerListViewModel) Init(_ *Model) {
	m.InitContainerSearchViewModel(
		func(idx int) {
			m.Cursor = idx
		}, func() {
			m.performSearch()
		})
}

func (m *DockerContainerListViewModel) Update(model *Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case dockerContainersLoadedMsg:
		model.loading = false
		if msg.err != nil {
			model.err = msg.err
		} else {
			model.err = nil
			m.Loaded(model, msg.containers)
		}
		return model, nil

	default:
		return model, nil
	}
}

func (m *DockerContainerListViewModel) Loaded(model *Model, containers []models.DockerContainer) {
	m.dockerContainers = containers
	m.SetRows(m.buildRows(), model.ViewHeight())
}

type ColumnMap struct {
	containerID table.Column
	image       table.Column
	state       table.Column
	status      table.Column
	ports       table.Column
	names       table.Column
}

func NewColumnMap(model *Model) ColumnMap {
	sideMargin := 2 * 2    // 2 for left and right padding
	cellMargin := 2        // 2 for cell margin
	containerIDWidth := 12 // Fixed width for container ID
	stateWidth := 10
	widthPerColumn := (model.width - containerIDWidth - stateWidth - cellMargin*4 - sideMargin) / 4

	return ColumnMap{
		containerID: table.Column{Title: "CONTAINER ID", Width: containerIDWidth},
		image:       table.Column{Title: "IMAGE", Width: widthPerColumn},
		state:       table.Column{Title: "STATE", Width: stateWidth}, // Fixed width for state
		status:      table.Column{Title: "STATUS", Width: widthPerColumn},
		ports:       table.Column{Title: "PORTS", Width: widthPerColumn},
		names:       table.Column{Title: "NAMES", Width: widthPerColumn},
	}
}

func (c *ColumnMap) ToArray() []table.Column {
	return []table.Column{
		c.containerID,
		c.image,
		c.state,
		c.status,
		c.ports,
		c.names,
	}
}

func (m *DockerContainerListViewModel) performSearch() {
	if m.searchText == "" {
		m.searchResults = nil
		m.currentSearchIdx = 0
		return
	}

	var results []int
	for i, container := range m.dockerContainers {
		// Create searchable text from container fields
		searchableText := container.ID + " " + container.Image + " " +
			container.Names + " " + container.Status + " " + container.Ports

		if m.MatchContainer(searchableText) {
			results = append(results, i)
		}
	}

	m.SetResults(results)

	// Jump to first result if found
	if len(results) > 0 {
		m.Cursor = results[0]
	}
}

func (m *DockerContainerListViewModel) buildRows() []table.Row {
	rows := make([]table.Row, 0, len(m.dockerContainers))
	for i, container := range m.dockerContainers {
		// Truncate container ID
		id := container.ID
		if len(id) > 12 { // shorten ID to 12 characters
			id = id[:12]
		}

		// Truncate image name
		image := container.Image

		state := container.State

		// Status with color
		status := container.Status

		// Truncate ports
		ports := container.Ports

		name := container.Names
		if container.IsDind() {
			name = dindStyle.Render("⬢ ") + name
		}

		// Highlight if this container matches search
		if m.IsSearchActive() && m.GetSearchText() != "" {
			for _, idx := range m.searchResults {
				if idx == i {
					// Apply search highlight style to name
					name = searchStyle.Render(name)
					break
				}
			}
		}

		rows = append(rows, table.Row{id, image, state, status, ports, name})
	}
	return rows
}

func (m *DockerContainerListViewModel) renderDockerList(model *Model, availableHeight int) string {
	// Container list
	if len(m.dockerContainers) == 0 {
		var s strings.Builder
		s.WriteString("\nNo containers found.\n")
		return s.String()
	}

	columns := NewColumnMap(model)

	return m.RenderTable(model, columns.ToArray(), availableHeight, func(row, col int) lipgloss.Style {
		if row == m.Cursor {
			return tableSelectedCellStyle
		}
		return tableNormalCellStyle
	})
}

func (m *DockerContainerListViewModel) HandleUp(model *Model) tea.Cmd {
	height := model.ViewHeight()
	if height <= 0 {
		height = 10 // fallback
	}
	if m.Cursor > 0 {
		m.Cursor--
		if m.Cursor < m.Start {
			m.Start = m.Cursor
		}
		m.End = clamp(m.Start+height, 0, len(m.Rows))
	}
	return nil
}

func (m *DockerContainerListViewModel) HandleDown(model *Model) tea.Cmd {
	height := model.ViewHeight()
	if height <= 0 {
		height = 10 // fallback
	}
	if m.Cursor < len(m.Rows)-1 {
		m.Cursor++
		if m.Cursor >= m.End {
			m.Start = m.Cursor - height + 1
			if m.Start < 0 {
				m.Start = 0
			}
		}
		m.End = clamp(m.Start+height, 0, len(m.Rows))
	}
	return nil
}

func (m *DockerContainerListViewModel) GetContainer(model *Model) *docker.Container {
	// Get the selected Docker container
	if m.Cursor < len(m.dockerContainers) {
		container := m.dockerContainers[m.Cursor]
		return docker.NewContainer(container.ID, container.Names, container.Names, container.State)
	}
	return nil
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
