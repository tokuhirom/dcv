package ui

import (
	"log/slog"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/tokuhirom/dcv/internal/models"
)

type ComposeProjectListViewModel struct {
	// Compose list state
	projects        []models.ComposeProject
	selectedProject int
}

func (m *ComposeProjectListViewModel) render(model *Model, availableHeight int) string {
	if len(m.projects) == 0 {
		var s strings.Builder
		s.WriteString("\nNo Docker Compose projects found.\n")
		s.WriteString("\nPress q to quit\n")
		return s.String()
	}

	// Project list

	columns := []table.Column{
		{Title: "NAME", Width: 20},
		{Title: "STATUS", Width: 15},
		{Title: "CONFIG FILES", Width: model.width - 40},
	}

	rows := make([]table.Row, 0, len(m.projects))
	for _, project := range m.projects {
		// Status with color
		status := project.Status
		slog.Info("Project status",
			slog.String("project", project.Name),
			slog.String("status", status))
		if strings.Contains(status, "running") {
			status = statusUpStyle.Render(status)
		} else {
			status = statusDownStyle.Render(status)
		}
		status += ResetAll

		// Truncate config files if too long
		configFiles := project.ConfigFiles
		if len(configFiles) > 50 {
			configFiles = configFiles[:47] + "..."
		}

		rows = append(rows, table.Row{project.Name, status, configFiles})
	}

	return RenderTable(columns, rows, availableHeight, m.selectedProject)
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
