package ui

import (
	"log/slog"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tokuhirom/dcv/internal/models"
)

// projectsLoadedMsg contains the loaded Compose projects
type projectsLoadedMsg struct {
	projects []models.ComposeProject
	err      error
}

type ComposeProjectListViewModel struct {
	TableViewModel
	// Compose list state
	projects []models.ComposeProject
}

// Update handles messages for the compose project list view
func (m *ComposeProjectListViewModel) Update(model *Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case projectsLoadedMsg:
		model.loading = false
		if msg.err != nil {
			model.err = msg.err
			return model, nil
		} else {
			model.err = nil
		}
		m.Loaded(model, msg.projects)
		return model, nil
	default:
		return model, nil
	}
}

func (m *ComposeProjectListViewModel) buildRows() []table.Row {
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
	return rows
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

	return m.RenderTable(model, columns, availableHeight, func(row, col int) lipgloss.Style {
		if row == m.Cursor {
			return tableSelectedCellStyle
		}
		return tableNormalCellStyle
	})
}

func (m *ComposeProjectListViewModel) HandleUp(model *Model) tea.Cmd {
	height := model.ViewHeight()
	if height <= 0 {
		height = 10 // fallback
	}
	if m.Cursor > 0 {
		m.Cursor--
		if m.Cursor < m.Start {
			m.Start = m.Cursor
		}
		m.End = clamp(m.Start+height, 0, len(m.Rows))
	}
	return nil
}

func (m *ComposeProjectListViewModel) HandleDown(model *Model) tea.Cmd {
	height := model.ViewHeight()
	if height <= 0 {
		height = 10 // fallback
	}
	if m.Cursor < len(m.Rows)-1 {
		m.Cursor++
		if m.Cursor >= m.End {
			m.Start = m.Cursor - height + 1
			if m.Start < 0 {
				m.Start = 0
			}
		}
		m.End = clamp(m.Start+height, 0, len(m.Rows))
	}
	return nil
}

func (m *ComposeProjectListViewModel) HandleSelectProject(model *Model) tea.Cmd {
	if m.Cursor < len(m.projects) {
		project := m.projects[m.Cursor]
		return model.composeProcessListViewModel.Load(model, project)
	}
	return nil
}

func (m *ComposeProjectListViewModel) Loaded(model *Model, projects []models.ComposeProject) {
	m.projects = projects
	m.SetRows(m.buildRows(), model.ViewHeight())
}

func (m *ComposeProjectListViewModel) DoLoad(model *Model) tea.Cmd {
	model.loading = true

	return func() tea.Msg {
		projects, err := model.dockerClient.ListComposeProjects()
		return projectsLoadedMsg{
			projects: projects,
			err:      err,
		}
	}
}

func (m *ComposeProjectListViewModel) Show(model *Model) tea.Cmd {
	model.SwitchView(ComposeProjectListView)
	return m.DoLoad(model)
}
