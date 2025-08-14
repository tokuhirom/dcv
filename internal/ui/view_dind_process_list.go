package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
)

type dindContainersLoadedMsg struct {
	containers []models.DockerContainer
	err        error
}

var _ ContainerAware = (*DindProcessListViewModel)(nil)
var _ UpdateAware = (*DindProcessListViewModel)(nil)
var _ ContainerSearchAware = (*DindProcessListViewModel)(nil)

// DindProcessListViewModel manages the state and rendering of the Docker-in-Docker process list view
type DindProcessListViewModel struct {
	TableViewModel
	dindContainers []models.DockerContainer
	showAll        bool

	hostContainer *docker.Container
}

func (m *DindProcessListViewModel) Init(_ *Model) {
	m.InitTableViewModel(func() {
		m.performSearch()
	})
}

func (m *DindProcessListViewModel) Update(model *Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case dindContainersLoadedMsg:
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

func (m *DindProcessListViewModel) performSearch() {
	if m.searchText == "" {
		m.searchResults = nil
		m.currentSearchIdx = 0
		return
	}

	var results []int
	for i, container := range m.dindContainers {
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

// buildRows builds the table rows from dind containers
func (m *DindProcessListViewModel) buildRows() []table.Row {
	rows := make([]table.Row, 0, len(m.dindContainers))
	for i, container := range m.dindContainers {
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

		state := container.State

		// Status with color
		status := container.Status
		if strings.Contains(status, "Up") {
			status = statusUpStyle.Render(status)
		} else {
			status = statusDownStyle.Render(status)
		}

		// Truncate ports if too long
		ports := container.Ports
		if len(ports) > 25 {
			ports = ports[:22] + "..."
		}

		name := container.Names
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

// render renders the dind process list view
func (m *DindProcessListViewModel) render(model *Model, availableHeight int) string {
	if len(m.dindContainers) == 0 {
		var s strings.Builder
		s.WriteString("\nNo containers running inside this dind container.\n")
		s.WriteString("\nPress ESC to go back\n")
		return s.String()
	}

	// Create table
	columns := []table.Column{
		{Title: "CONTAINER ID", Width: 15},
		{Title: "IMAGE", Width: 30},
		{Title: "STATE", Width: 10},
		{Title: "STATUS", Width: 20},
		{Title: "PORTS", Width: 25},
		{Title: "NAMES", Width: 25},
	}

	// Reduce available height if search info will be displayed
	tableHeight := availableHeight
	if m.IsSearchActive() && m.GetSearchText() != "" && m.GetSearchInfo() != "" {
		tableHeight -= 2 // Reserve lines for search info
	}

	return m.RenderTable(model, columns, tableHeight, func(row, col int) lipgloss.Style {
		if row == m.Cursor {
			return tableSelectedCellStyle
		}
		return tableNormalCellStyle
	})
}

// Load switches to the dind process list view and loads containers
func (m *DindProcessListViewModel) Load(model *Model, hostContainer *docker.Container) tea.Cmd {
	m.hostContainer = hostContainer
	model.SwitchView(DindProcessListView)
	return m.DoLoad(model)
}

// DoLoad reloads the dind container list
func (m *DindProcessListViewModel) DoLoad(model *Model) tea.Cmd {
	model.loading = true
	return func() tea.Msg {
		containers, err := model.dockerClient.ListDindContainers(m.hostContainer.GetContainerID(), m.showAll)
		return dindContainersLoadedMsg{
			containers: containers,
			err:        err,
		}
	}
}

// HandleBack returns to the compose process list view
func (m *DindProcessListViewModel) HandleBack(model *Model) tea.Cmd {
	model.SwitchToPreviousView()
	return nil
}

// HandleToggleAll toggles showing all containers including stopped ones
func (m *DindProcessListViewModel) HandleToggleAll(model *Model) tea.Cmd {
	m.showAll = !m.showAll
	return m.DoLoad(model)
}

// Loaded updates the dind container list after loading
func (m *DindProcessListViewModel) Loaded(model *Model, containers []models.DockerContainer) {
	m.dindContainers = containers
	m.SetRows(m.buildRows(), model.ViewHeight())
}

func (m *DindProcessListViewModel) GetContainer(model *Model) *docker.Container {
	if m.Cursor < len(m.dindContainers) {
		container := m.dindContainers[m.Cursor]
		return docker.NewDindContainer(m.hostContainer.GetContainerID(), m.hostContainer.GetName(), container.ID, container.Names, container.State)
	}
	return nil
}

func (m *DindProcessListViewModel) Title() string {
	title := fmt.Sprintf("Docker in Docker: %s", m.hostContainer.GetName())
	if m.showAll {
		title = fmt.Sprintf("Docker in Docker: %s (all)", m.hostContainer.GetName())
	}
	// Add search info if searching
	if m.IsSearchActive() && m.GetSearchText() != "" {
		info := m.GetSearchInfo()
		if info != "" {
			title += " - " + info
		}
	}
	return title
}
