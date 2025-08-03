package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
)

// NetworkListView represents the Docker network list view
type NetworkListView struct {
	// View state
	width           int
	height          int
	selectedNetwork int
	networks        []models.DockerNetwork
	
	// Loading/error state
	loading bool
	err     error
	
	// Dependencies
	dockerClient *docker.Client
	rootScreen   tea.Model
}

// NewNetworkListView creates a new network list view
func NewNetworkListView(dockerClient *docker.Client) *NetworkListView {
	return &NetworkListView{
		dockerClient: dockerClient,
	}
}

// SetRootScreen sets the root screen reference
func (v *NetworkListView) SetRootScreen(root tea.Model) {
	v.rootScreen = root
}

// Init initializes the view
func (v *NetworkListView) Init() tea.Cmd {
	v.loading = true
	return loadNetworks(v.dockerClient)
}

// Update handles messages for this view
func (v *NetworkListView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		return v, nil
		
	case tea.KeyMsg:
		return v.handleKeyPress(msg)
		
	case networksLoadedMsg:
		v.loading = false
		if msg.err != nil {
			v.err = msg.err
			return v, nil
		}
		v.networks = msg.networks
		v.err = nil
		if len(v.networks) > 0 && v.selectedNetwork >= len(v.networks) {
			v.selectedNetwork = 0
		}
		return v, nil
		
	case RefreshMsg:
		v.loading = true
		v.err = nil
		return v, loadNetworks(v.dockerClient)
		
	case serviceActionCompleteMsg:
		if msg.err != nil {
			v.err = msg.err
			return v, nil
		}
		// Refresh after deletion
		v.loading = true
		return v, loadNetworks(v.dockerClient)
	}
	
	return v, nil
}

// View renders the network list
func (v *NetworkListView) View() string {
	if v.loading {
		return renderLoadingView(v.width, v.height, "Loading Docker networks...")
	}
	
	if v.err != nil {
		return renderErrorView(v.width, v.height, v.err)
	}
	
	return v.renderNetworkList()
}

func (v *NetworkListView) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if v.selectedNetwork > 0 {
			v.selectedNetwork--
		}
		return v, nil
		
	case "down", "j":
		if v.selectedNetwork < len(v.networks)-1 {
			v.selectedNetwork++
		}
		return v, nil
		
	case "enter", "I":
		// Inspect network
		if v.selectedNetwork < len(v.networks) && v.rootScreen != nil {
			network := v.networks[v.selectedNetwork]
			if switcher, ok := v.rootScreen.(interface {
				SwitchScreen(tea.Model) (tea.Model, tea.Cmd)
			}); ok {
				inspectView := NewInspectView(v.dockerClient, network.ID, "network", network.Name)
				inspectView.SetRootScreen(v.rootScreen)
				return switcher.SwitchScreen(inspectView)
			}
		}
		return v, nil
		
	case "r":
		// Send refresh message
		return v, func() tea.Msg { return RefreshMsg{} }
		
	case "D":
		// Delete network
		if v.selectedNetwork < len(v.networks) {
			network := v.networks[v.selectedNetwork]
			// Don't allow deleting default networks
			if network.Name == "bridge" || network.Name == "host" || network.Name == "none" {
				v.err = fmt.Errorf("cannot delete default network: %s", network.Name)
				return v, nil
			}
			return v, removeNetwork(v.dockerClient, network.ID)
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
				helpView := NewHelpView("Docker Networks", v)
				helpView.SetRootScreen(v.rootScreen)
				return switcher.SwitchScreen(helpView)
			}
		}
		return v, nil
	}
	
	return v, nil
}

func (v *NetworkListView) renderNetworkList() string {
	var s strings.Builder
	
	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("7")).
		Background(lipgloss.Color("4")).
		Width(v.width).
		Padding(0, 1).
		Render("Docker Networks")
	s.WriteString(header + "\n")
	
	// Network list
	if len(v.networks) == 0 {
		s.WriteString("\nNo networks found.\n")
	} else {
		// Column headers
		headers := fmt.Sprintf("%-15s %-20s %-10s %-10s %s",
			"NETWORK ID", "NAME", "DRIVER", "SCOPE", "CONTAINERS")
		s.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Bold(true).
			Render(headers) + "\n")
		
		for i, network := range v.networks {
			selected := i == v.selectedNetwork
			line := formatNetworkLine(network, v.width, selected)
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

func formatNetworkLine(network models.DockerNetwork, width int, selected bool) string {
	// Truncate long fields
	id := network.ID
	if len(id) > 12 {
		id = id[:12]
	}
	
	name := network.Name
	if len(name) > 18 {
		name = name[:18]
	}
	
	containerCount := fmt.Sprintf("%d", network.ContainerCount)
	
	line := fmt.Sprintf("%-15s %-20s %-10s %-10s %s",
		id, name, network.Driver, network.Scope, containerCount)
	
	if len(line) > width-3 {
		line = line[:width-3]
	}
	
	style := lipgloss.NewStyle()
	if selected {
		style = style.Background(lipgloss.Color("240"))
	}
	
	// Color special networks
	if network.Name == "bridge" || network.Name == "host" || network.Name == "none" {
		style = style.Foreground(lipgloss.Color("241")) // Dim for default networks
	}
	
	return style.Render(line)
}

// Messages
type networksLoadedMsg struct {
	networks []models.DockerNetwork
	err      error
}

// Commands
func loadNetworks(client *docker.Client) tea.Cmd {
	return func() tea.Msg {
		networks, err := client.ListNetworks()
		return networksLoadedMsg{networks: networks, err: err}
	}
}

func removeNetwork(client *docker.Client, networkID string) tea.Cmd {
	return func() tea.Msg {
		err := client.RemoveNetwork(networkID)
		return serviceActionCompleteMsg{err: err}
	}
}