package ui

import (
	"log/slog"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/tokuhirom/dcv/internal/models"
)

type ComposeProcessListViewModel struct {
}

func (m *ComposeProcessListViewModel) Load(model *Model, project models.ComposeProject) tea.Cmd {
	model.projectName = project.Name
	model.currentView = ComposeProcessListView
	model.loading = true
	return loadProcesses(model.dockerClient, model.projectName, model.showAll)
}

func (m *Model) renderComposeProcessList(availableHeight int) string {
	var s strings.Builder

	slog.Info("Rendering container list",
		slog.Int("selectedContainer", m.selectedContainer),
		slog.Int("numContainers", len(m.composeContainers)))

	// Empty state
	if len(m.composeContainers) == 0 {
		s.WriteString("\nNo containers found.\n")
		s.WriteString("\nPress u to start services or p to switch to project list\n")
		return s.String()
	}

	// Create table with fixed widths
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == m.selectedContainer {
				return selectedStyle
			}
			return normalStyle
		}).
		Headers("SERVICE", "IMAGE", "STATUS", "PORTS").
		Height(availableHeight).
		Width(m.width).
		Offset(m.selectedContainer)

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

	return s.String()
}
