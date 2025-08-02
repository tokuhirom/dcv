package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

func (m *Model) renderProjectList() string {
	var s strings.Builder

	if len(m.projects) == 0 {
		s.WriteString("\nNo Docker Compose projects found.\n")
		s.WriteString("\nPress q to quit\n")
		return s.String()
	}

	// Project list

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == m.selectedProject {
				return selectedStyle
			}
			return normalStyle
		}).
		Headers("NAME", "STATUS", "CONFIG FILES").
		Height(m.height - 4). // Reserve space for footer
		Width(m.width).
		Offset(m.selectedDockerImage)

	for _, project := range m.projects {
		// Status with color
		status := project.Status
		if status == "running" {
			status = statusUpStyle.Render(status)
		} else {
			status = statusDownStyle.Render(status)
		}

		// Truncate config files if too long
		configFiles := project.ConfigFiles
		if len(configFiles) > 50 {
			configFiles = configFiles[:47] + "..."
		}

		t.Row(project.Name, status, configFiles)
	}

	s.WriteString(t.Render() + "\n")

	return s.String()
}
