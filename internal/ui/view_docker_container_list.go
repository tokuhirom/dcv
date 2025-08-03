package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

func (m *Model) renderDockerList(availableHeight int) string {
	var s strings.Builder

	if len(m.dockerContainers) == 0 {
		s.WriteString("\nNo containers found.\n")
		return s.String()
	}

	// Container list

	// Define consistent styles for table cells
	idStyle := lipgloss.NewStyle().Width(12)
	imageStyle := lipgloss.NewStyle().Width(30)
	statusStyle := lipgloss.NewStyle().Width(20)
	portsStyle := lipgloss.NewStyle().Width(30)
	nameStyle := lipgloss.NewStyle().Width(20)

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			baseStyle := normalStyle
			if row == m.selectedDockerContainer {
				baseStyle = selectedStyle
			}

			// Apply column-specific styling
			switch col {
			case 0:
				return baseStyle.Inherit(idStyle)
			case 1:
				return baseStyle.Inherit(imageStyle)
			case 2:
				return baseStyle.Inherit(statusStyle)
			case 3:
				return baseStyle.Inherit(portsStyle)
			case 4:
				return baseStyle.Inherit(nameStyle)
			default:
				return baseStyle
			}
		}).
		Headers("CONTAINER ID", "IMAGE", "STATUS", "PORTS", "NAMES")

	for _, container := range m.dockerContainers {
		// Truncate container ID
		id := container.ID
		if len(id) > 12 {
			id = id[:12]
		}
		id = idStyle.Render(id)

		// Truncate image name
		image := container.Image
		if len(image) > 30 {
			image = image[:27] + "..."
		}
		image = imageStyle.Render(image)

		// Status with color
		status := container.Status
		if strings.Contains(status, "Up") || strings.Contains(status, "running") {
			status = statusUpStyle.Render(status)
		} else {
			status = statusDownStyle.Render(status)
		}

		// Truncate ports
		ports := container.Ports
		if len(ports) > 30 {
			ports = ports[:27] + "..."
		}
		ports = portsStyle.Render(ports)

		name := nameStyle.Render(container.Names)

		t.Row(id, image, status, ports, name)
	}

	s.WriteString(t.Render() + "\n")

	return s.String()
}
