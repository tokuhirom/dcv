package ui

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tokuhirom/dcv/internal/docker"

	"github.com/tokuhirom/dcv/internal/models"
)

type composeProcessesLoadedMsg struct {
	processes []models.ComposeContainer
	err       error
}

var _ ContainerAware = (*ComposeProcessListViewModel)(nil)
var _ UpdateAware = (*ComposeProcessListViewModel)(nil)
var _ ContainerSearchAware = (*ComposeProcessListViewModel)(nil)

type ComposeProcessListViewModel struct {
	ContainerSearchViewModel
	TableViewModel
	// Process list state
	composeContainers []models.ComposeContainer
	showAll           bool   // Toggle to show all composeContainers including stopped ones
	projectName       string // Current Docker Compose project name
}

func (m *ComposeProcessListViewModel) Init(_ *Model) {
	m.InitContainerSearchViewModel(
		func(idx int) {
			m.Cursor = idx
		}, func() {
			m.performSearch()
		})
}

func (m *ComposeProcessListViewModel) Update(model *Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case composeProcessesLoadedMsg:
		model.loading = false
		if msg.err != nil {
			model.err = msg.err
		} else {
			model.err = nil
			m.Loaded(model, msg.processes)
		}
		return model, nil

	default:
		return model, nil
	}
}

func (m *ComposeProcessListViewModel) Load(model *Model, project models.ComposeProject) tea.Cmd {
	m.projectName = project.Name
	model.SwitchView(ComposeProcessListView)
	return m.DoLoad(model)
}

func (m *ComposeProcessListViewModel) DoLoad(model *Model) tea.Cmd {
	model.loading = true
	return func() tea.Msg {
		slog.Info("Loading composeContainers",
			slog.Bool("showAll", m.showAll))
		processes, err := model.dockerClient.ListComposeContainers(m.projectName, m.showAll)
		return composeProcessesLoadedMsg{
			processes: processes,
			err:       err,
		}
	}
}

func (m *ComposeProcessListViewModel) performSearch() {
	if m.searchText == "" {
		m.searchResults = nil
		m.currentSearchIdx = 0
		return
	}

	var results []int
	for i, container := range m.composeContainers {
		// Create searchable text from container fields
		searchableText := container.Service + " " + container.Image + " " +
			container.Name + " " + container.GetStatus() + " " + container.GetPortsString()

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

func (m *ComposeProcessListViewModel) buildRows() []table.Row {
	rows := make([]table.Row, 0, len(m.composeContainers))
	// Add rows with width control
	for i, container := range m.composeContainers {
		// Service name with dind indicator
		service := container.Service
		if container.IsDind() {
			service = "🔄 " + service
		}

		// Highlight if this container matches search
		if m.IsSearchActive() && m.GetSearchText() != "" {
			for _, idx := range m.searchResults {
				if idx == i {
					// Apply search highlight style to service name
					service = searchStyle.Render(service)
					break
				}
			}
		}

		// Truncate image name if too long
		image := container.Image
		if len(image) > 30 {
			image = image[:27] + "..."
		}

		state := container.State

		// Status with color
		status := container.GetStatus()
		if strings.Contains(status, "Up") || strings.Contains(status, "running") {
			status = statusUpStyle.Render(status)
		} else {
			status = statusDownStyle.Render(status)
		}

		// Truncate ports if too long
		ports := container.GetPortsString()
		if len(ports) > 40 {
			ports = ports[:37] + "..."
		}

		rows = append(rows, table.Row{service, image, state, status, ports})
	}
	return rows
}

func (m *ComposeProcessListViewModel) render(model *Model, availableHeight int) string {
	slog.Info("Rendering container list",
		slog.Int("cursor", m.Cursor),
		slog.Int("numContainers", len(m.composeContainers)))

	// Empty state
	if len(m.composeContainers) == 0 {
		var s strings.Builder
		s.WriteString("\nNo containers found.\n")
		s.WriteString("\nPress u to start services or p to switch to project list\n")
		return s.String()
	}

	// Create table with fixed widths
	columns := []table.Column{
		{Title: "SERVICE", Width: 20},
		{Title: "IMAGE", Width: 30},
		{Title: "STATE", Width: 10},
		{Title: "STATUS", Width: 20},
		{Title: "PORTS", Width: model.width - 75},
	}

	tableOutput := m.RenderTable(model, columns, availableHeight, func(row, col int) lipgloss.Style {
		if row == m.Cursor {
			return tableSelectedCellStyle
		}
		return tableNormalCellStyle
	})

	// Add search info if searching
	if m.IsSearchActive() && m.GetSearchText() != "" {
		searchInfo := m.GetSearchInfo()
		if searchInfo != "" {
			return tableOutput + "\n" + searchInfo
		}
	}

	return tableOutput
}

func (m *ComposeProcessListViewModel) HandleToggleAll(model *Model) tea.Cmd {
	m.showAll = !m.showAll
	return m.DoLoad(model)
}

func (m *ComposeProcessListViewModel) HandleDindProcessList(model *Model) tea.Cmd {
	container := m.GetContainer(model)
	if container == nil {
		slog.Error("Failed to get selected container for DIND process list",
			slog.Any("error", fmt.Errorf("no container selected")))
		return nil
	}

	return model.dindProcessListViewModel.Load(model, container)
}

func (m *ComposeProcessListViewModel) GetContainer(model *Model) *docker.Container {
	if m.Cursor < len(m.composeContainers) {
		container := m.composeContainers[m.Cursor]
		return docker.NewContainer(container.ID, container.Name, fmt.Sprintf("%s(project:%s)", container.Service, m.projectName), container.State)
	}
	return nil
}

func (m *ComposeProcessListViewModel) HandleBack(model *Model) tea.Cmd {
	model.SwitchToPreviousView()
	return nil
}

func (m *ComposeProcessListViewModel) Loaded(model *Model, processes []models.ComposeContainer) {
	m.composeContainers = processes
	m.SetRows(m.buildRows(), model.ViewHeight())
}
