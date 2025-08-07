package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

	"github.com/tokuhirom/dcv/internal/models"
)

// NetworkListViewModel manages the state and rendering of the network list view
type NetworkListViewModel struct {
	dockerNetworks        []models.DockerNetwork
	selectedDockerNetwork int
}

// render renders the network list view
func (m *NetworkListViewModel) render(model *Model, availableHeight int) string {
	var content strings.Builder

	if len(m.dockerNetworks) == 0 {
		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		content.WriteString(dimStyle.Render("No networks found"))
		return content.String()
	}

	// Define dimStyle for ID column
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	// Create table with lipgloss
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			baseStyle := normalStyle
			if row == m.selectedDockerNetwork {
				baseStyle = selectedStyle
			}
			// Dim the ID column for non-selected rows
			if col == 0 && row != m.selectedDockerNetwork {
				return dimStyle
			}
			return baseStyle
		}).
		Headers("NETWORK ID", "NAME", "DRIVER", "SCOPE", "CONTAINERS").
		Height(availableHeight).
		Width(model.width)

	// Add rows
	for _, network := range m.dockerNetworks {
		// Format row data
		id := truncate(network.ID, 12)
		name := truncate(network.Name, 30)
		driver := truncate(network.Driver, 15)
		scope := truncate(network.Scope, 10)
		containers := fmt.Sprintf("%d", network.GetContainerCount())

		t.Row(id, name, driver, scope, containers)
	}

	// Set the scroll offset based on selection
	t.Offset(m.selectedDockerNetwork)

	return t.String()
}

// Show switches to the network list view
func (m *NetworkListViewModel) Show(model *Model) tea.Cmd {
	model.currentView = NetworkListView
	model.loading = true
	return loadDockerNetworks(model.dockerClient)
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
		model.loading = true
		return removeNetwork(model.dockerClient, network.ID)
	}
	return nil
}

// HandleInspect shows the inspect view for the selected network
func (m *NetworkListViewModel) HandleInspect(model *Model) tea.Cmd {
	if m.selectedDockerNetwork < len(m.dockerNetworks) {
		network := m.dockerNetworks[m.selectedDockerNetwork]
		return model.inspectViewModel.InspectNetwork(model, network)
	}
	return nil
}

// HandleBack returns to the compose process list view
func (m *NetworkListViewModel) HandleBack(model *Model) tea.Cmd {
	model.SwitchToPreviousView()
	return nil
}

// HandleRefresh reloads the network list
func (m *NetworkListViewModel) HandleRefresh(model *Model) tea.Cmd {
	model.loading = true
	return loadDockerNetworks(model.dockerClient)
}

// Loaded updates the networks list after loading
func (m *NetworkListViewModel) Loaded(networks []models.DockerNetwork) {
	m.dockerNetworks = networks
	if len(m.dockerNetworks) > 0 && m.selectedDockerNetwork >= len(m.dockerNetworks) {
		m.selectedDockerNetwork = 0
	}
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
