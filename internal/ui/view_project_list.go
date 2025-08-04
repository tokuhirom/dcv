package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

	"github.com/tokuhirom/dcv/internal/models"
)

type ComposeProjectListViewModel struct {
	// Compose list state
	projects        []models.ComposeProject
	selectedProject int
}

func (m *ComposeProjectListViewModel) render(model *Model, availableHeight int) string {
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
		Height(availableHeight).
		Width(model.width).
		Offset(m.selectedProject)

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

func (m *ComposeProjectListViewModel) HandleUp(_ *Model) tea.Cmd {
	if m.selectedProject > 0 {
		m.selectedProject--
	}
	return nil
}

func (m *ComposeProjectListViewModel) HandleDown(_ *Model) tea.Cmd {
	if m.selectedProject < len(m.projects)-1 {
		m.selectedProject++
	}
	return nil
}

func (m *ComposeProjectListViewModel) HandleSelectProject(model *Model) tea.Cmd {
	if m.selectedProject < len(m.projects) {
		project := m.projects[m.selectedProject]
		return model.composeProcessListViewModel.Load(model, project)
	}
	return nil
}

func (m *ComposeProjectListViewModel) Loaded(projects []models.ComposeProject) {
	m.projects = projects
	if len(m.projects) > 0 && m.selectedProject >= len(m.projects) {
		m.selectedProject = 0
	}
}
