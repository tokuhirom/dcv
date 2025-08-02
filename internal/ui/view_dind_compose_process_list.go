package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

func (m *Model) renderDindList() string {
	var s strings.Builder

	if len(m.dindContainers) == 0 {
		s.WriteString("\nNo containers running inside this dind container.\n")
		s.WriteString("\nPress ESC to go back\n")
		return s.String()
	}

	// Container list
	s.WriteString("\n")

	// Create table
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return headerStyle
			}
			if row-1 == m.selectedDindContainer {
				return selectedStyle
			}
			return normalStyle
		}).
		Headers("CONTAINER ID", "IMAGE", "STATUS", "NAMES")

	for _, container := range m.dindContainers {
		// Truncate container ID
		id := container.ID
		if len(id) > 12 {
			id = id[:12]
		}

		// Truncate image name
		image := container.Image
		if len(image) > 30 {
			image = image[:27] + "..."
		}

		// Status with color
		status := container.Status
		if strings.Contains(status, "Up") {
			status = statusUpStyle.Render(status)
		} else {
			status = statusDownStyle.Render(status)
		}

		t.Row(id, image, status, container.Names)
	}

	s.WriteString(t.Render() + "\n\n")

	return s.String()
}
