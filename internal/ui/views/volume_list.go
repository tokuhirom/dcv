package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
)

// VolumeListView represents the Docker volume list view
type VolumeListView struct {
	// View state
	width          int
	height         int
	selectedVolume int
	volumes        []models.DockerVolume

	// Loading/error state
	loading bool
	err     error

	// Dependencies
	dockerClient *docker.Client
	rootScreen   tea.Model
}

// NewVolumeListView creates a new volume list view
func NewVolumeListView(dockerClient *docker.Client) *VolumeListView {
	return &VolumeListView{
		dockerClient: dockerClient,
	}
}

// SetRootScreen sets the root screen reference
func (v *VolumeListView) SetRootScreen(root tea.Model) {
	v.rootScreen = root
}

// Init initializes the view
func (v *VolumeListView) Init() tea.Cmd {
	v.loading = true
	return loadVolumes(v.dockerClient)
}

// Update handles messages for this view
func (v *VolumeListView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		return v, nil

	case tea.KeyMsg:
		return v.handleKeyPress(msg)

	case volumesLoadedMsg:
		v.loading = false
		if msg.err != nil {
			v.err = msg.err
			return v, nil
		}
		v.volumes = msg.volumes
		v.err = nil
		if len(v.volumes) > 0 && v.selectedVolume >= len(v.volumes) {
			v.selectedVolume = 0
		}
		return v, nil

	case RefreshMsg:
		v.loading = true
		v.err = nil
		return v, loadVolumes(v.dockerClient)

	case serviceActionCompleteMsg:
		if msg.err != nil {
			v.err = msg.err
			return v, nil
		}
		// Refresh after deletion
		v.loading = true
		return v, loadVolumes(v.dockerClient)
	}

	return v, nil
}

// View renders the volume list
func (v *VolumeListView) View() string {
	if v.loading {
		return renderLoadingView(v.width, v.height, "Loading Docker volumes...")
	}

	if v.err != nil {
		return renderErrorView(v.width, v.height, v.err)
	}

	return v.renderVolumeList()
}

func (v *VolumeListView) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if v.selectedVolume > 0 {
			v.selectedVolume--
		}
		return v, nil

	case "down", "j":
		if v.selectedVolume < len(v.volumes)-1 {
			v.selectedVolume++
		}
		return v, nil

	case "enter", "I":
		// Inspect volume
		if v.selectedVolume < len(v.volumes) && v.rootScreen != nil {
			volume := v.volumes[v.selectedVolume]
			if switcher, ok := v.rootScreen.(interface {
				SwitchScreen(tea.Model) (tea.Model, tea.Cmd)
			}); ok {
				inspectView := NewInspectView(v.dockerClient, volume.Name, "volume", volume.Name)
				inspectView.SetRootScreen(v.rootScreen)
				return switcher.SwitchScreen(inspectView)
			}
		}
		return v, nil

	case "r":
		// Send refresh message
		return v, func() tea.Msg { return RefreshMsg{} }

	case "D":
		// Delete volume
		if v.selectedVolume < len(v.volumes) {
			volume := v.volumes[v.selectedVolume]
			return v, removeVolume(v.dockerClient, volume.Name)
		}
		return v, nil

	case "1":
		// Switch to Docker container list
		if v.rootScreen != nil {
			if switcher, ok := v.rootScreen.(interface {
				SwitchScreen(tea.Model) (tea.Model, tea.Cmd)
			}); ok {
				dockerView := NewDockerListView(v.dockerClient)
				dockerView.SetRootScreen(v.rootScreen)
				return switcher.SwitchScreen(dockerView)
			}
		}
		return v, nil

	case "esc", "q":
		// Go back to Docker container list
		if v.rootScreen != nil {
			if switcher, ok := v.rootScreen.(interface {
				SwitchScreen(tea.Model) (tea.Model, tea.Cmd)
			}); ok {
				dockerView := NewDockerListView(v.dockerClient)
				dockerView.SetRootScreen(v.rootScreen)
				return switcher.SwitchScreen(dockerView)
			}
		}
		return v, nil

	case "?":
		// Show help
		if v.rootScreen != nil {
			if switcher, ok := v.rootScreen.(interface {
				SwitchScreen(tea.Model) (tea.Model, tea.Cmd)
			}); ok {
				helpView := NewHelpView("Docker Volumes", v)
				helpView.SetRootScreen(v.rootScreen)
				return switcher.SwitchScreen(helpView)
			}
		}
		return v, nil
	}

	return v, nil
}

func (v *VolumeListView) renderVolumeList() string {
	var s strings.Builder

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("7")).
		Background(lipgloss.Color("4")).
		Width(v.width).
		Padding(0, 1).
		Render("Docker Volumes")
	s.WriteString(header + "\n")

	// Volume list
	if len(v.volumes) == 0 {
		s.WriteString("\nNo volumes found.\n")
	} else {
		// Column headers
		headers := fmt.Sprintf("%-30s %-15s %s",
			"VOLUME NAME", "DRIVER", "MOUNTPOINT")
		s.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Bold(true).
			Render(headers) + "\n")

		for i, volume := range v.volumes {
			selected := i == v.selectedVolume
			line := formatVolumeLine(volume, v.width, selected)
			s.WriteString(line + "\n")
		}
	}

	// Footer
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Width(v.width).
		Align(lipgloss.Center).
		Render("↑/↓: Navigate • Enter/I: Inspect • r: Refresh • D: Delete • 1: Docker • ESC: Back")

	// Pad to fill screen
	content := s.String()
	lines := strings.Split(content, "\n")
	for len(lines) < v.height-2 {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n") + "\n" + footer
}

func formatVolumeLine(volume models.DockerVolume, width int, selected bool) string {
	// Truncate long fields
	name := volume.Name
	if len(name) > 28 {
		name = name[:28]
	}

	mountpoint := volume.Mountpoint
	if len(mountpoint) > 40 {
		mountpoint = "..." + mountpoint[len(mountpoint)-37:]
	}

	line := fmt.Sprintf("%-30s %-15s %s",
		name, volume.Driver, mountpoint)

	if len(line) > width-3 {
		line = line[:width-3]
	}

	style := lipgloss.NewStyle()
	if selected {
		style = style.Background(lipgloss.Color("240"))
	}

	return style.Render(line)
}

// Messages
type volumesLoadedMsg struct {
	volumes []models.DockerVolume
	err     error
}

// Commands
func loadVolumes(client *docker.Client) tea.Cmd {
	return func() tea.Msg {
		volumes, err := client.ListVolumes()
		return volumesLoadedMsg{volumes: volumes, err: err}
	}
}

func removeVolume(client *docker.Client, volumeName string) tea.Cmd {
	return func() tea.Msg {
		err := client.RemoveVolume(volumeName, false)
		return serviceActionCompleteMsg{err: err}
	}
}
