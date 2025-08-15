package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tokuhirom/dcv/internal/models"
)

// dockerVolumesLoadedMsg contains the loaded Docker volumes
type dockerVolumesLoadedMsg struct {
	volumes []models.DockerVolume
	err     error
}

// VolumeListViewModel manages the state and rendering of the Docker volume list view
type VolumeListViewModel struct {
	TableViewModel
	dockerVolumes []models.DockerVolume
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

		m.Loaded(model, msg.volumes)
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
		{Title: "Name", Width: -1},
		{Title: "Driver", Width: -1},
		{Title: "Scope", Width: -1},
	}

	return m.RenderTable(model, columns, availableHeight, func(row, col int) lipgloss.Style {
		if row == m.Cursor {
			return tableSelectedCellStyle
		}
		return tableNormalCellStyle
	})
}

// buildRows builds the table rows from docker volumes
func (m *VolumeListViewModel) buildRows() []table.Row {
	rows := make([]table.Row, 0, len(m.dockerVolumes))
	for _, volume := range m.dockerVolumes {
		rows = append(rows, table.Row{
			volume.Name,
			volume.Driver,
			volume.Scope,
		})
	}
	return rows
}

// Show switches to the volume list view
func (m *VolumeListViewModel) Show(model *Model) tea.Cmd {
	model.SwitchView(VolumeListView)
	m.Cursor = 0
	m.dockerVolumes = []models.DockerVolume{}
	model.err = nil
	return m.DoLoad(model)
}

// HandleUp moves selection up in the volume list
func (m *VolumeListViewModel) HandleUp(model *Model) tea.Cmd {
	return m.TableViewModel.HandleUp(model)
}

// HandleDown moves selection down in the volume list
func (m *VolumeListViewModel) HandleDown(model *Model) tea.Cmd {
	return m.TableViewModel.HandleDown(model)
}

// HandleInspect shows the inspect view for the selected volume
func (m *VolumeListViewModel) HandleInspect(model *Model) tea.Cmd {
	if len(m.dockerVolumes) == 0 || m.Cursor >= len(m.dockerVolumes) {
		return nil
	}

	volume := m.dockerVolumes[m.Cursor]
	model.loading = true
	model.err = nil
	return model.inspectViewModel.Inspect(model, "volume "+volume.Name, func() ([]byte, error) {
		return model.dockerClient.ExecuteCaptured("volume", "inspect", volume.Name)
	})
}

// HandleDelete removes the selected volume
func (m *VolumeListViewModel) HandleDelete(model *Model, force bool) tea.Cmd {
	if len(m.dockerVolumes) == 0 || m.Cursor >= len(m.dockerVolumes) {
		return nil
	}

	volume := m.dockerVolumes[m.Cursor]
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
func (m *VolumeListViewModel) Loaded(model *Model, volumes []models.DockerVolume) {
	m.dockerVolumes = volumes
	m.SetRows(m.buildRows(), model.ViewHeight())
}
