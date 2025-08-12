package ui

import (
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/tokuhirom/dcv/internal/models"
)

// Helper function to get the selected compose project
func (m *Model) getSelectedComposeProject() *models.ComposeProject {
	if m.currentView == ComposeProjectListView {
		if m.composeProjectListViewModel.selectedProject < len(m.composeProjectListViewModel.projects) {
			return &m.composeProjectListViewModel.projects[m.composeProjectListViewModel.selectedProject]
		}
	}
	return nil
}

// Helper function to execute commands on the selected compose project
func (m *Model) useComposeProjectAware(cb func(project *models.ComposeProject) tea.Cmd) tea.Cmd {
	project := m.getSelectedComposeProject()
	if project == nil {
		slog.Error("No compose project selected")
		return nil
	}
	return cb(project)
}

// CmdComposeUp runs docker compose up -d for the selected project
func (m *Model) CmdComposeUp(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.useComposeProjectAware(func(project *models.ComposeProject) tea.Cmd {
		args := []string{"compose", "-p", project.Name, "up", "-d"}
		return m.commandExecutionViewModel.ExecuteCommand(m, false, args...) // up is not aggressive
	})
}

// CmdComposeDown runs docker compose down for the selected project
func (m *Model) CmdComposeDown(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.useComposeProjectAware(func(project *models.ComposeProject) tea.Cmd {
		args := []string{"compose", "-p", project.Name, "down"}
		return m.commandExecutionViewModel.ExecuteCommand(m, true, args...) // down is aggressive
	})
}

// CmdComposeStop runs docker compose stop for the selected project
func (m *Model) CmdComposeStop(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.useComposeProjectAware(func(project *models.ComposeProject) tea.Cmd {
		args := []string{"compose", "-p", project.Name, "stop"}
		return m.commandExecutionViewModel.ExecuteCommand(m, true, args...) // stop is aggressive
	})
}

// CmdComposeStart runs docker compose start for the selected project
func (m *Model) CmdComposeStart(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.useComposeProjectAware(func(project *models.ComposeProject) tea.Cmd {
		args := []string{"compose", "-p", project.Name, "start"}
		return m.commandExecutionViewModel.ExecuteCommand(m, false, args...) // start is not aggressive
	})
}

// CmdComposeRestart runs docker compose restart for the selected project
func (m *Model) CmdComposeRestart(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.useComposeProjectAware(func(project *models.ComposeProject) tea.Cmd {
		args := []string{"compose", "-p", project.Name, "restart"}
		return m.commandExecutionViewModel.ExecuteCommand(m, true, args...) // restart is aggressive
	})
}

// CmdComposeBuild runs docker compose build for the selected project
func (m *Model) CmdComposeBuild(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.useComposeProjectAware(func(project *models.ComposeProject) tea.Cmd {
		args := []string{"compose", "-f", project.ConfigFiles, "build"}
		return m.commandExecutionViewModel.ExecuteCommand(m, false, args...) // build is not aggressive
	})
}

// CmdComposePull runs docker compose pull for the selected project
func (m *Model) CmdComposePull(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.useComposeProjectAware(func(project *models.ComposeProject) tea.Cmd {
		args := []string{"compose", "-f", project.ConfigFiles, "pull"}
		return m.commandExecutionViewModel.ExecuteCommand(m, false, args...) // pull is not aggressive
	})
}

// CmdComposeLogs runs docker compose logs for the selected project
func (m *Model) CmdComposeLogs(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.useComposeProjectAware(func(project *models.ComposeProject) tea.Cmd {
		args := []string{"compose", "-p", project.Name, "logs", "--follow", "--tail", "10"}
		return m.commandExecutionViewModel.ExecuteCommand(m, false, args...) // logs is not aggressive
	})
}
