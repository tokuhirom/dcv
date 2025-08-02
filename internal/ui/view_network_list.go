package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderNetworkList renders the network list view
func (m *Model) renderNetworkList() string {
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
