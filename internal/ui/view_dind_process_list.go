package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
)

// DindProcessListViewModel manages the state and rendering of the Docker-in-Docker process list view
type DindProcessListViewModel struct {
	dindContainers         []models.DockerContainer
	selectedDindContainer  int
	currentDindHost        string // Container name (for display)
	currentDindContainerID string // Service name (for docker compose exec)
}

// render renders the dind process list view
func (m *DindProcessListViewModel) render(availableHeight int) string {
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
		{Title: "STATUS", Width: 20},
		{Title: "NAMES", Width: 30},
	}

	rows := make([]table.Row, 0, len(m.dindContainers))
	for _, container := range m.dindContainers {
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
		if strings.Contains(status, "Up") {
			status = statusUpStyle.Render(status)
		} else {
			status = statusDownStyle.Render(status)
		}

		rows = append(rows, table.Row{id, image, status, container.Names})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithHeight(availableHeight-2),
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
	if m.selectedDindContainer < len(rows) {
		t.MoveDown(m.selectedDindContainer)
	}

	return t.View()
}

// Load switches to the dind process list view and loads containers
func (m *DindProcessListViewModel) Load(model *Model, container models.GenericContainer) tea.Cmd {
	m.currentDindHost = container.GetName()
	m.currentDindContainerID = container.GetID()
	model.SwitchView(DindProcessListView)
	model.loading = true
	return loadDindContainers(model.dockerClient, container.GetID())
}

// HandleUp moves selection up in the dind container list
func (m *DindProcessListViewModel) HandleUp() tea.Cmd {
	if m.selectedDindContainer > 0 {
		m.selectedDindContainer--
	}
	return nil
}

// HandleDown moves selection down in the dind container list
func (m *DindProcessListViewModel) HandleDown() tea.Cmd {
	if m.selectedDindContainer < len(m.dindContainers)-1 {
		m.selectedDindContainer++
	}
	return nil
}

// HandleLog shows logs for the selected dind container
func (m *DindProcessListViewModel) HandleLog(model *Model) tea.Cmd {
	if len(m.dindContainers) == 0 || m.selectedDindContainer >= len(m.dindContainers) {
		return nil
	}

	container := m.dindContainers[m.selectedDindContainer]
	return model.logViewModel.StreamLogsDind(model, m.currentDindContainerID, container)
}

// HandleBack returns to the compose process list view
func (m *DindProcessListViewModel) HandleBack(model *Model) tea.Cmd {
	model.SwitchToPreviousView()
	return nil
}

// HandleRefresh reloads the dind container list
func (m *DindProcessListViewModel) HandleRefresh(model *Model) tea.Cmd {
	model.loading = true
	return loadDindContainers(model.dockerClient, m.currentDindContainerID)
}

// Loaded updates the dind container list after loading
func (m *DindProcessListViewModel) Loaded(containers []models.DockerContainer) {
	m.dindContainers = containers
	if len(m.dindContainers) > 0 && m.selectedDindContainer >= len(m.dindContainers) {
		m.selectedDindContainer = 0
	}
}

func loadDindContainers(client *docker.Client, containerID string) tea.Cmd {
	return func() tea.Msg {
		containers, err := client.Dind(containerID).ListContainers()
		return dindContainersLoadedMsg{
			containers: containers,
			err:        err,
		}
	}
}
