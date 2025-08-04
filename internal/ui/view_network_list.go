package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

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

	// Table headers
	headerStyle := lipgloss.NewStyle().Bold(true).Underline(true).Foreground(lipgloss.Color("86"))
	headers := []string{"NETWORK ID", "NAME", "DRIVER", "SCOPE", "CONTAINERS"}
	colWidths := []int{12, 30, 15, 10, 10}

	// Render headers
	for i, header := range headers {
		content.WriteString(headerStyle.Render(padRight(header, colWidths[i])))
		if i < len(headers)-1 {
			content.WriteString(" ")
		}
	}
	content.WriteString("\n")

	// Render networks
	normalStyle := lipgloss.NewStyle()
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220")).Background(lipgloss.Color("235"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	for i, network := range m.dockerNetworks {
		style := normalStyle
		if i == m.selectedDockerNetwork {
			style = selectedStyle
		}

		// Format row data
		row := []string{
			truncate(network.ID, 12),
			truncate(network.Name, 30),
			truncate(network.Driver, 15),
			truncate(network.Scope, 10),
			fmt.Sprintf("%d", network.GetContainerCount()),
		}

		// Render row
		for j, col := range row {
			if j == 0 && i != m.selectedDockerNetwork {
				// Dim the ID for non-selected rows
				content.WriteString(dimStyle.Render(padRight(col, colWidths[j])))
			} else {
				content.WriteString(style.Render(padRight(col, colWidths[j])))
			}
			if j < len(row)-1 {
				content.WriteString(" ")
			}
		}
		content.WriteString("\n")
	}

	return content.String()
}

// Show switches to the network list view
func (m *NetworkListViewModel) Show(model *Model) tea.Cmd {
	model.currentView = NetworkListView
	model.loading = true
	return loadDockerNetworks(model.dockerClient)
}

// HandleSelectUp moves the selection up
func (m *NetworkListViewModel) HandleSelectUp() tea.Cmd {
	if m.selectedDockerNetwork > 0 {
		m.selectedDockerNetwork--
	}
	return nil
}

// HandleSelectDown moves the selection down
func (m *NetworkListViewModel) HandleSelectDown() tea.Cmd {
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
		model.inspectNetworkID = network.ID
		model.inspectContainerID = "" // Clear container ID
		model.inspectImageID = ""     // Clear image ID
		model.loading = true
		return loadNetworkInspect(model.dockerClient, network.ID)
	}
	return nil
}

// HandleBack returns to the compose process list view
func (m *NetworkListViewModel) HandleBack(model *Model) tea.Cmd {
	model.currentView = ComposeProcessListView
	return loadProcesses(model.dockerClient, model.projectName, model.composeProcessListViewModel.showAll)
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

func padRight(s string, length int) string {
	if len(s) >= length {
		return s[:length]
	}
	return s + strings.Repeat(" ", length-len(s))
}
