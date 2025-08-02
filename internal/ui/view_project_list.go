package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

func (m *Model) renderProjectList() string {
	var s strings.Builder

	title := titleStyle.Render("Docker Compose Projects")
	s.WriteString(title + "\n")

	if m.loading {
		s.WriteString("\nLoading...\n")
		return s.String()
	}

	if m.err != nil {
		s.WriteString("\n" + errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n")
		return s.String()
	}

	if len(m.projects) == 0 {
		s.WriteString("\nNo Docker Compose projects found.\n")
		s.WriteString("\nPress q to quit\n")
		return s.String()
	}

	// Project list
	s.WriteString("\n")

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return headerStyle
			}
			if row-1 == m.selectedProject {
				return selectedStyle
			}
			return normalStyle
		}).
		Headers("NAME", "STATUS", "CONFIG FILES")

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

	s.WriteString(t.Render() + "\n\n")

	// Show help hint
	s.WriteString(helpStyle.Render("Press ? for help"))

	return s.String()
}
