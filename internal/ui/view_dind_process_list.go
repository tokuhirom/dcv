package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
)

type dindContainersLoadedMsg struct {
	containers []models.DockerContainer
	err        error
}

var _ ContainerAware = (*DindProcessListViewModel)(nil)
var _ UpdateAware = (*DindProcessListViewModel)(nil)

// DindProcessListViewModel manages the state and rendering of the Docker-in-Docker process list view
type DindProcessListViewModel struct {
	dindContainers        []models.DockerContainer
	selectedDindContainer int
	showAll               bool

	hostContainer *docker.Container
}

func (m *DindProcessListViewModel) Update(model *Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case dindContainersLoadedMsg:
		model.loading = false
		if msg.err != nil {
			model.err = msg.err
		} else {
			model.err = nil
			m.Loaded(msg.containers)
		}
		return model, nil

	default:
		return model, nil
	}
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
		{Title: "STATE", Width: 10},
		{Title: "STATUS", Width: 25},
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

		state := container.State

		// Status with color
		status := container.Status
		if strings.Contains(status, "Up") {
			status = statusUpStyle.Render(status)
		} else {
			status = statusDownStyle.Render(status)
		}

		rows = append(rows, table.Row{id, image, state, status, container.Names})
	}

	return RenderTable(columns, rows, availableHeight, m.selectedDindContainer)
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
func (m *DindProcessListViewModel) Loaded(containers []models.DockerContainer) {
	m.dindContainers = containers
	if len(m.dindContainers) > 0 && m.selectedDindContainer >= len(m.dindContainers) {
		m.selectedDindContainer = 0
	}
}

func (m *DindProcessListViewModel) GetContainer(model *Model) *docker.Container {
	if m.selectedDindContainer < len(m.dindContainers) {
		container := m.dindContainers[m.selectedDindContainer]
		return docker.NewDindContainer(m.hostContainer.GetContainerID(), m.hostContainer.GetName(), container.ID, container.Names, container.State)
	}
	return nil
}

func (m *DindProcessListViewModel) Title() string {
	if m.showAll {
		return fmt.Sprintf("Docker in Docker: %s (all)", m.hostContainer.GetName())
	}
	return fmt.Sprintf("Docker in Docker: %s", m.hostContainer.GetName())
}
