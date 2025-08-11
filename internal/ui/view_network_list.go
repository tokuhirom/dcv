package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/tokuhirom/dcv/internal/models"
)

type dockerNetworksLoadedMsg struct {
	networks []models.DockerNetwork
	err      error
}

var _ HandleInspectAware = (*NetworkListViewModel)(nil)
var _ UpdateAware = (*NetworkListViewModel)(nil)

// NetworkListViewModel manages the state and rendering of the network list view
type NetworkListViewModel struct {
	dockerNetworks        []models.DockerNetwork
	selectedDockerNetwork int
}

func (m *NetworkListViewModel) Update(model *Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case dockerNetworksLoadedMsg:
		model.loading = false
		if msg.err != nil {
			model.err = msg.err
		} else {
			model.err = nil
			m.Loaded(msg.networks)
		}
		return model, nil

	default:
		return model, nil
	}
}

func (m *NetworkListViewModel) Loaded(networks []models.DockerNetwork) {
	m.dockerNetworks = networks
	if len(m.dockerNetworks) > 0 && m.selectedDockerNetwork >= len(m.dockerNetworks) {
		m.selectedDockerNetwork = 0
	}
}

// render renders the network list view
func (m *NetworkListViewModel) render(availableHeight int) string {
	if len(m.dockerNetworks) == 0 {
		s := strings.Builder{}
		s.WriteString("No networks found.\n")
		s.WriteString(helpStyle.Render("\nPress 'q' to go back"))
		return s.String()
	}

	// Create table columns
	columns := []table.Column{
		{Title: "NETWORK ID", Width: 12},
		{Title: "NAME", Width: 30},
		{Title: "DRIVER", Width: 15},
		{Title: "SCOPE", Width: 10},
		{Title: "CONTAINERS", Width: 10},
	}

	// Create table rows
	rows := make([]table.Row, len(m.dockerNetworks))
	for i, network := range m.dockerNetworks {
		rows[i] = table.Row{
			truncate(network.ID, 12),
			truncate(network.Name, 30),
			truncate(network.Driver, 15),
			truncate(network.Scope, 10),
			fmt.Sprintf("%d", network.GetContainerCount()),
		}
	}

	return RenderTable(columns, rows, availableHeight, m.selectedDockerNetwork)
}

// Show switches to the network list view
func (m *NetworkListViewModel) Show(model *Model) tea.Cmd {
	model.SwitchView(NetworkListView)
	return m.DoLoad(model)
}

func (m *NetworkListViewModel) DoLoad(model *Model) tea.Cmd {
	model.loading = true
	return func() tea.Msg {
		networks, err := model.dockerClient.ListNetworks()
		return dockerNetworksLoadedMsg{
			networks: networks,
			err:      err,
		}
	}
}

// HandleUp moves the selection up
func (m *NetworkListViewModel) HandleUp() tea.Cmd {
	if m.selectedDockerNetwork > 0 {
		m.selectedDockerNetwork--
	}
	return nil
}

// HandleDown moves the selection down
func (m *NetworkListViewModel) HandleDown() tea.Cmd {
	if m.selectedDockerNetwork < len(m.dockerNetworks)-1 {
		m.selectedDockerNetwork++
	}
	return nil
}

// HandleDelete removes the selected network
func (m *NetworkListViewModel) HandleDelete(model *Model) tea.Cmd {
	if m.selectedDockerNetwork < len(m.dockerNetworks) {
		network := m.dockerNetworks[m.selectedDockerNetwork]
		// Don't allow removing default networks
		if network.Name == "bridge" || network.Name == "host" || network.Name == "none" {
			model.err = fmt.Errorf("cannot remove default network: %s", network.Name)
			return nil
		}
		// Use CommandExecutionView to show real-time output
		args := []string{"network", "rm", network.ID}
		return model.commandExecutionViewModel.ExecuteCommand(model, true, args...) // network rm is aggressive
	}
	return nil
}

// HandleInspect shows the inspect view for the selected network
func (m *NetworkListViewModel) HandleInspect(model *Model) tea.Cmd {
	if m.selectedDockerNetwork < len(m.dockerNetworks) {
		network := m.dockerNetworks[m.selectedDockerNetwork]
		return model.inspectViewModel.Inspect(model, "network "+network.Name, func() ([]byte, error) {
			return model.dockerClient.ExecuteCaptured("network", "inspect", network.ID)
		})
	}
	return nil
}

// HandleBack returns to the compose process list view
func (m *NetworkListViewModel) HandleBack(model *Model) tea.Cmd {
	model.SwitchToPreviousView()
	return nil
}

// Helper functions
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
