package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/tokuhirom/dcv/internal/models"
)

// dockerVolumesLoadedMsg contains the loaded Docker volumes
type dockerVolumesLoadedMsg struct {
	volumes []models.DockerVolume
	err     error
}

// VolumeListViewModel manages the state and rendering of the Docker volume list view
type VolumeListViewModel struct {
	dockerVolumes        []models.DockerVolume
	selectedDockerVolume int
}

// Update handles messages for the volume list view
func (m *VolumeListViewModel) Update(model *Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case dockerVolumesLoadedMsg:
		model.loading = false
		if msg.err != nil {
			model.err = msg.err
			return model, nil
		} else {
			model.err = nil
		}

		m.Loaded(msg.volumes)
		return model, nil
	default:
		return model, nil
	}
}

// render renders the volume list view
func (m *VolumeListViewModel) render(model *Model, availableHeight int) string {
	if len(m.dockerVolumes) == 0 {
		s := strings.Builder{}
		s.WriteString("No volumes found.\n")
		s.WriteString(helpStyle.Render("\nPress 'Esc' to go back"))
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

	return RenderTable(columns, rows, availableHeight, m.selectedDockerVolume)
}

// Show switches to the volume list view
func (m *VolumeListViewModel) Show(model *Model) tea.Cmd {
	model.SwitchView(VolumeListView)
	m.selectedDockerVolume = 0
	m.dockerVolumes = []models.DockerVolume{}
	model.err = nil
	return m.DoLoad(model)
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
	return model.inspectViewModel.Inspect(model, "volume "+volume.Name, func() ([]byte, error) {
		return model.dockerClient.ExecuteCaptured("volume", "inspect", volume.Name)
	})
}

// HandleDelete removes the selected volume
func (m *VolumeListViewModel) HandleDelete(model *Model, force bool) tea.Cmd {
	if len(m.dockerVolumes) == 0 || m.selectedDockerVolume >= len(m.dockerVolumes) {
		return nil
	}

	volume := m.dockerVolumes[m.selectedDockerVolume]
	// Use CommandExecutionView to show real-time output
	args := []string{"volume", "rm"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, volume.Name)
	return model.commandExecutionViewModel.ExecuteCommand(model, true, args...) // volume rm is aggressive
}

// HandleBack returns to the compose process list view
func (m *VolumeListViewModel) HandleBack(model *Model) tea.Cmd {
	model.SwitchToPreviousView()
	return nil
}

// DoLoad reloads the volume list
func (m *VolumeListViewModel) DoLoad(model *Model) tea.Cmd {
	model.loading = true
	return func() tea.Msg {
		volumes, err := model.dockerClient.ListVolumes()
		return dockerVolumesLoadedMsg{volumes: volumes, err: err}
	}
}

// Loaded updates the volume list after loading
func (m *VolumeListViewModel) Loaded(volumes []models.DockerVolume) {
	m.dockerVolumes = volumes
	if len(m.dockerVolumes) > 0 && m.selectedDockerVolume >= len(m.dockerVolumes) {
		m.selectedDockerVolume = 0
	}
}
