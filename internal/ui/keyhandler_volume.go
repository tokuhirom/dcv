package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/tokuhirom/dcv/internal/docker"
)

// Volume navigation handlers
func (m *Model) SelectUpVolume(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedDockerVolume > 0 {
		m.selectedDockerVolume--
	}
	return m, nil
}

func (m *Model) SelectDownVolume(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedDockerVolume < len(m.dockerVolumes)-1 {
		m.selectedDockerVolume++
	}
	return m, nil
}

// View change handlers
func (m *Model) ShowVolumeList(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.currentView = VolumeListView
	m.loading = true
	m.selectedDockerVolume = 0
	m.err = nil
	return m, loadDockerVolumes(m.dockerClient)
}

func (m *Model) BackFromVolumeList(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.currentView = ComposeProcessListView
	m.err = nil
	return m, loadProcesses(m.dockerClient, m.projectName, m.showAll)
}

// Action handlers
func (m *Model) ShowVolumeInspect(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if len(m.dockerVolumes) == 0 || m.selectedDockerVolume >= len(m.dockerVolumes) {
		return m, nil
	}

	volume := m.dockerVolumes[m.selectedDockerVolume]
	m.loading = true
	m.err = nil
	m.inspectVolumeID = volume.Name
	m.currentView = InspectView

	return m, inspectVolume(m.dockerClient, volume.Name)
}

// Deprecated: Use Refresh instead
func (m *Model) RefreshVolumeList(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m.Refresh(tea.KeyMsg{})
}

func (m *Model) DeleteVolume(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if len(m.dockerVolumes) == 0 || m.selectedDockerVolume >= len(m.dockerVolumes) {
		return m, nil
	}

	volume := m.dockerVolumes[m.selectedDockerVolume]
	m.loading = true
	m.err = nil

	return m, removeVolume(m.dockerClient, volume.Name, false)
}

func (m *Model) ForceDeleteVolume(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if len(m.dockerVolumes) == 0 || m.selectedDockerVolume >= len(m.dockerVolumes) {
		return m, nil
	}

	volume := m.dockerVolumes[m.selectedDockerVolume]
	m.loading = true
	m.err = nil

	return m, removeVolume(m.dockerClient, volume.Name, true)
}

// Command functions
func inspectVolume(dockerClient *docker.Client, volumeID string) tea.Cmd {
	return func() tea.Msg {
		output, err := dockerClient.InspectVolume(volumeID)
		if err != nil {
			return errorMsg{err: err}
		}
		return inspectLoadedMsg{content: output}
	}
}

func removeVolume(dockerClient *docker.Client, volumeName string, force bool) tea.Cmd {
	return func() tea.Msg {
		err := dockerClient.RemoveVolume(volumeName, force)
		return serviceActionCompleteMsg{err: err}
	}
}
