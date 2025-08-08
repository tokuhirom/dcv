package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
)

// VolumeListViewModel manages the state and rendering of the Docker volume list view
type VolumeListViewModel struct {
	dockerVolumes        []models.DockerVolume
	selectedDockerVolume int
}

// render renders the volume list view
func (m *VolumeListViewModel) render(model *Model, availableHeight int) string {
	if len(m.dockerVolumes) == 0 {
		s := strings.Builder{}
		s.WriteString("No volumes found.\n")
		s.WriteString(helpStyle.Render("\nPress 'q' to go back"))
		return s.String()
	}

	// Create table columns
	columns := []table.Column{
		{Title: "Name", Width: 40},
		{Title: "Driver", Width: 10},
		{Title: "Scope", Width: 10},
	}

	// Create table rows
	rows := make([]table.Row, len(m.dockerVolumes))
	for i, volume := range m.dockerVolumes {
		rows[i] = table.Row{
			volume.Name,
			volume.Driver,
			volume.Scope,
		}
	}

	// Create table
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(availableHeight-2),
	)

	tableStyle := table.DefaultStyles()
	tableStyle.Header = tableStyle.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	tableStyle.Selected = tableStyle.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)

	t.SetStyles(tableStyle)
	t.Focus()

	// Move to selected row
	t.MoveDown(m.selectedDockerVolume)

	return t.View()
}

// Show switches to the volume list view
func (m *VolumeListViewModel) Show(model *Model) tea.Cmd {
	model.currentView = VolumeListView
	model.loading = true
	m.selectedDockerVolume = 0
	m.dockerVolumes = []models.DockerVolume{}
	model.err = nil
	return loadDockerVolumes(model.dockerClient)
}

// HandleUp moves selection up in the volume list
func (m *VolumeListViewModel) HandleUp() tea.Cmd {
	if m.selectedDockerVolume > 0 {
		m.selectedDockerVolume--
	}
	return nil
}

// HandleDown moves selection down in the volume list
func (m *VolumeListViewModel) HandleDown() tea.Cmd {
	if m.selectedDockerVolume < len(m.dockerVolumes)-1 {
		m.selectedDockerVolume++
	}
	return nil
}

// HandleInspect shows the inspect view for the selected volume
func (m *VolumeListViewModel) HandleInspect(model *Model) tea.Cmd {
	if len(m.dockerVolumes) == 0 || m.selectedDockerVolume >= len(m.dockerVolumes) {
		return nil
	}

	volume := m.dockerVolumes[m.selectedDockerVolume]
	model.loading = true
	model.err = nil
	return model.inspectViewModel.InspectVolume(model, volume)
}

// HandleDelete removes the selected volume
func (m *VolumeListViewModel) HandleDelete(model *Model, force bool) tea.Cmd {
	if len(m.dockerVolumes) == 0 || m.selectedDockerVolume >= len(m.dockerVolumes) {
		return nil
	}

	volume := m.dockerVolumes[m.selectedDockerVolume]
	model.loading = true
	model.err = nil

	return removeVolume(model.dockerClient, volume.Name, force)
}

// HandleBack returns to the compose process list view
func (m *VolumeListViewModel) HandleBack(model *Model) tea.Cmd {
	model.SwitchToPreviousView()
	return nil
}

// HandleRefresh reloads the volume list
func (m *VolumeListViewModel) HandleRefresh(model *Model) tea.Cmd {
	model.loading = true
	return loadDockerVolumes(model.dockerClient)
}

// Loaded updates the volume list after loading
func (m *VolumeListViewModel) Loaded(volumes []models.DockerVolume) {
	m.dockerVolumes = volumes
	if len(m.dockerVolumes) > 0 && m.selectedDockerVolume >= len(m.dockerVolumes) {
		m.selectedDockerVolume = 0
	}
}

func loadDockerVolumes(dockerClient *docker.Client) tea.Cmd {
	return func() tea.Msg {
		volumes, err := dockerClient.ListVolumes()
		return dockerVolumesLoadedMsg{volumes: volumes, err: err}
	}
}

func removeVolume(dockerClient *docker.Client, volumeName string, force bool) tea.Cmd {
	return func() tea.Msg {
		err := dockerClient.RemoveVolume(volumeName, force)
		return serviceActionCompleteMsg{err: err}
	}
}
