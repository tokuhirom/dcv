package ui

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

func (m *Model) renderComposeProcessList() string {
	var s strings.Builder

	slog.Info("Rendering container list",
		slog.Int("selectedContainer", m.selectedContainer),
		slog.Int("numContainers", len(m.composeContainers)))

	// Title with project name
	title := "Docker Compose"
	if m.projectName != "" {
		title = fmt.Sprintf("Docker Compose: %s", m.projectName)
	}
	s.WriteString(titleStyle.Render(title) + "\n")

	// Loading state
	if m.loading {
		s.WriteString("\nLoading...\n")
		return s.String()
	}

	// Error state
	if m.err != nil {
		s.WriteString("\n" + errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n")
		s.WriteString("\nPress q to quit\n")
		return s.String()
	}

	// Empty state
	if len(m.composeContainers) == 0 {
		s.WriteString("\nNo containers found.\n")
		s.WriteString("\nPress u to start services or p to switch to project list\n")
		return s.String()
	}

	// Container list
	s.WriteString("\n")

	// Create table with fixed widths
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return headerStyle
			}
			if row-1 == m.selectedContainer {
				return selectedStyle
			}
			return normalStyle
		}).
		Headers("SERVICE", "IMAGE", "STATUS", "PORTS")

	// Add rows with width control
	for _, container := range m.composeContainers {
		// Service name with dind indicator
		service := container.Service
		if container.IsDind() {
			service = dindStyle.Render("⬢ ") + service
		}

		// Truncate image name if too long
		image := container.Image
		if len(image) > 30 {
			image = image[:27] + "..."
		}

		// Status with color
		status := container.GetStatus()
		if strings.Contains(status, "Up") || strings.Contains(status, "running") {
			status = statusUpStyle.Render(status)
		} else {
			status = statusDownStyle.Render(status)
		}

		// Truncate ports if too long
		ports := container.GetPortsString()
		if len(ports) > 40 {
			ports = ports[:37] + "..."
		}

		t.Row(service, image, status, ports)
	}

	s.WriteString(t.Render() + "\n\n")

	// Show help hint
	s.WriteString(helpStyle.Render("Press ? for help"))

	return s.String()
}
