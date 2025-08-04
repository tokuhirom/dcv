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

// VolumeListViewModel manages the state and rendering of the Docker volume list view
type VolumeListViewModel struct {
	dockerVolumes        []models.DockerVolume
	selectedDockerVolume int
}

// render renders the volume list view
func (m *VolumeListViewModel) render(model *Model, availableHeight int) string {
	s := strings.Builder{}

	if len(m.dockerVolumes) == 0 {
		s.WriteString("No volumes found.\n")
		s.WriteString(helpStyle.Render("\nPress 'q' to go back"))
		return s.String()
	}

	// Create table columns
	columns := []table.Column{
		{Title: "Name", Width: 40},
		{Title: "Driver", Width: 10},
		{Title: "Scope", Width: 10},
		{Title: "Size", Width: 12},
		{Title: "Created", Width: 20},
		{Title: "Ref Count", Width: 10},
	}

	// Create table rows
	rows := make([]table.Row, len(m.dockerVolumes))
	for i, volume := range m.dockerVolumes {
		size := formatBytes(volume.Size)
		if volume.Size == 0 {
			size = "-"
		}

		refCount := fmt.Sprintf("%d", volume.RefCount)
		if volume.RefCount == 0 && volume.Size == 0 {
			refCount = "-"
		}

		created := volume.CreatedAt.Format("2006-01-02 15:04:05")
		if volume.CreatedAt.IsZero() {
			created = "-"
		}

		rows[i] = table.Row{
			volume.Name,
			volume.Driver,
			volume.Scope,
			size,
			created,
			refCount,
		}
	}

	// Create table
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(min(len(rows), model.Height-8)),
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
	for i := 0; i < m.selectedDockerVolume; i++ {
		t.MoveDown(1)
	}

	s.WriteString(t.View())
	s.WriteString("\n")

	return s.String()
}

// Show switches to the volume list view
func (m *VolumeListViewModel) Show(model *Model) tea.Cmd {
	model.currentView = VolumeListView
	model.loading = true
	m.selectedDockerVolume = 0
	model.err = nil
	return loadDockerVolumes(model.dockerClient)
}

// HandleSelectUp moves selection up in the volume list
func (m *VolumeListViewModel) HandleSelectUp() tea.Cmd {
	if m.selectedDockerVolume > 0 {
		m.selectedDockerVolume--
	}
	return nil
}

// HandleSelectDown moves selection down in the volume list
func (m *VolumeListViewModel) HandleSelectDown() tea.Cmd {
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
func (m *VolumeListViewModel) HandleDelete(model *Model) tea.Cmd {
	if len(m.dockerVolumes) == 0 || m.selectedDockerVolume >= len(m.dockerVolumes) {
		return nil
	}

	volume := m.dockerVolumes[m.selectedDockerVolume]
	model.loading = true
	model.err = nil

	return removeVolume(model.dockerClient, volume.Name, false)
}

// HandleForceDelete forcefully removes the selected volume
func (m *VolumeListViewModel) HandleForceDelete(model *Model) tea.Cmd {
	if len(m.dockerVolumes) == 0 || m.selectedDockerVolume >= len(m.dockerVolumes) {
		return nil
	}

	volume := m.dockerVolumes[m.selectedDockerVolume]
	model.loading = true
	model.err = nil

	return removeVolume(model.dockerClient, volume.Name, true)
}

// HandleBack returns to the compose process list view
func (m *VolumeListViewModel) HandleBack(model *Model) tea.Cmd {
	model.currentView = ComposeProcessListView
	model.err = nil
	return loadProcesses(model.dockerClient, model.projectName, model.composeProcessListViewModel.showAll)
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

// formatBytes formats bytes into human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
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
