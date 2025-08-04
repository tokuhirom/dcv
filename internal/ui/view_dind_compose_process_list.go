package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

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
func (m *DindProcessListViewModel) render(model *Model, availableHeight int) string {
	var s strings.Builder

	if len(m.dindContainers) == 0 {
		s.WriteString("\nNo containers running inside this dind container.\n")
		s.WriteString("\nPress ESC to go back\n")
		return s.String()
	}

	// Container list
	s.WriteString("\n")

	// Create table
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return headerStyle
			}
			if row-1 == m.selectedDindContainer {
				return selectedStyle
			}
			return normalStyle
		}).
		Headers("CONTAINER ID", "IMAGE", "STATUS", "NAMES").
		Height(availableHeight)

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

		t.Row(id, image, status, container.Names)
	}

	s.WriteString(t.Render() + "\n\n")

	return s.String()
}

// Load switches to the dind process list view and loads containers
func (m *DindProcessListViewModel) Load(model *Model, container models.ComposeContainer) tea.Cmd {
	m.currentDindHost = container.Name
	m.currentDindContainerID = container.ID
	model.currentView = DindProcessListView
	model.loading = true
	return loadDindContainers(model.dockerClient, container.ID)
}

// HandleSelectUp moves selection up in the dind container list
func (m *DindProcessListViewModel) HandleSelectUp() tea.Cmd {
	if m.selectedDindContainer > 0 {
		m.selectedDindContainer--
	}
	return nil
}

// HandleSelectDown moves selection down in the dind container list
func (m *DindProcessListViewModel) HandleSelectDown() tea.Cmd {
	if m.selectedDindContainer < len(m.dindContainers)-1 {
		m.selectedDindContainer++
	}
	return nil
}

// HandleShowLog shows logs for the selected dind container
func (m *DindProcessListViewModel) HandleShowLog(model *Model) tea.Cmd {
	if len(m.dindContainers) == 0 || m.selectedDindContainer >= len(m.dindContainers) {
		return nil
	}

	container := m.dindContainers[m.selectedDindContainer]
	return model.logViewModel.ShowDindLog(model, m.currentDindContainerID, container)
}

// HandleBack returns to the compose process list view
func (m *DindProcessListViewModel) HandleBack(model *Model) tea.Cmd {
	model.currentView = ComposeProcessListView
	return loadProcesses(model.dockerClient, model.projectName, model.composeProcessListViewModel.showAll)
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
		containers, err := client.ListDindContainers(containerID)
		return dindContainersLoadedMsg{
			containers: containers,
			err:        err,
		}
	}
}
