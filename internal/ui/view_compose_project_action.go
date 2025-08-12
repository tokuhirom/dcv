package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tokuhirom/dcv/internal/models"
)

// ComposeProjectAction represents a compose project operation
type ComposeProjectAction struct {
	Key         string
	Name        string
	Description string
	Aggressive  bool
	Handler     func(m *Model, project models.ComposeProject) tea.Cmd
}

// ComposeProjectActionViewModel manages the compose project action selection view
type ComposeProjectActionViewModel struct {
	actions        []ComposeProjectAction
	selectedAction int
	targetProject  *models.ComposeProject
}

// Initialize sets up the action view with available commands for a compose project
func (m *ComposeProjectActionViewModel) Initialize(project *models.ComposeProject) {
	m.targetProject = project
	m.selectedAction = 0

	// Define available actions based on project state
	m.actions = []ComposeProjectAction{}

	// Always available actions
	m.actions = append(m.actions, ComposeProjectAction{
		Key:         "Enter",
		Name:        "View Containers",
		Description: "Show project containers",
		Aggressive:  false,
		Handler: func(model *Model, p models.ComposeProject) tea.Cmd {
			return model.composeProcessListViewModel.Load(model, p)
		},
	})

	// Project status dependent actions
	if strings.Contains(project.Status, "running") {
		m.actions = append(m.actions, ComposeProjectAction{
			Key:         "U",
			Name:        "Deploy",
			Description: "Update containers (up -d)",
			Aggressive:  false,
			Handler: func(model *Model, p models.ComposeProject) tea.Cmd {
				args := []string{"compose", "-p", p.Name, "up", "-d"}
				return model.commandExecutionViewModel.ExecuteCommand(model, false, args...)
			},
		})

		m.actions = append(m.actions, ComposeProjectAction{
			Key:         "D",
			Name:        "Down",
			Description: "Stop and remove containers, networks",
			Aggressive:  true,
			Handler: func(model *Model, p models.ComposeProject) tea.Cmd {
				args := []string{"compose", "-p", p.Name, "down"}
				return model.commandExecutionViewModel.ExecuteCommand(model, true, args...) // down is aggressive
			},
		})

		m.actions = append(m.actions, ComposeProjectAction{
			Key:         "S",
			Name:        "Stop",
			Description: "Stop all services",
			Aggressive:  true,
			Handler: func(model *Model, p models.ComposeProject) tea.Cmd {
				args := []string{"compose", "-p", p.Name, "stop"}
				return model.commandExecutionViewModel.ExecuteCommand(model, true, args...)
			},
		})

		m.actions = append(m.actions, ComposeProjectAction{
			Key:         "R",
			Name:        "Restart",
			Description: "Restart all services",
			Aggressive:  true,
			Handler: func(model *Model, p models.ComposeProject) tea.Cmd {
				args := []string{"compose", "-p", p.Name, "restart"}
				return model.commandExecutionViewModel.ExecuteCommand(model, true, args...)
			},
		})
	} else {
		m.actions = append(m.actions, ComposeProjectAction{
			Key:         "U",
			Name:        "Up",
			Description: "Create and start containers",
			Aggressive:  false,
			Handler: func(model *Model, p models.ComposeProject) tea.Cmd {
				args := []string{"compose", "-p", p.Name, "up", "-d"}
				return model.commandExecutionViewModel.ExecuteCommand(model, false, args...) // up is not aggressive
			},
		})

		m.actions = append(m.actions, ComposeProjectAction{
			Key:         "S",
			Name:        "Start",
			Description: "Start existing containers",
			Aggressive:  false,
			Handler: func(model *Model, p models.ComposeProject) tea.Cmd {
				args := []string{"compose", "-p", p.Name, "start"}
				return model.commandExecutionViewModel.ExecuteCommand(model, false, args...)
			},
		})
	}

	// Always available build command
	m.actions = append(m.actions, ComposeProjectAction{
		Key:         "B",
		Name:        "Build",
		Description: "Build or rebuild services",
		Aggressive:  false,
		Handler: func(model *Model, p models.ComposeProject) tea.Cmd {
			args := []string{"compose", "-f", p.ConfigFiles, "build"}
			return model.commandExecutionViewModel.ExecuteCommand(model, false, args...)
		},
	})

	// Pull images
	m.actions = append(m.actions, ComposeProjectAction{
		Key:         "P",
		Name:        "Pull",
		Description: "Pull service images",
		Aggressive:  false,
		Handler: func(model *Model, p models.ComposeProject) tea.Cmd {
			args := []string{"compose", "-f", p.ConfigFiles, "pull"}
			return model.commandExecutionViewModel.ExecuteCommand(model, false, args...)
		},
	})

	// View logs
	m.actions = append(m.actions, ComposeProjectAction{
		Key:         "L",
		Name:        "Logs",
		Description: "View all service logs",
		Aggressive:  false,
		Handler: func(model *Model, p models.ComposeProject) tea.Cmd {
			args := []string{"compose", "-p", p.Name, "logs", "--follow", "--tail", "10"}
			return model.commandExecutionViewModel.ExecuteCommand(model, false, args...)
		},
	})
}

// render displays the action selection menu
func (m *ComposeProjectActionViewModel) render(model *Model) string {
	if m.targetProject == nil {
		return "No project selected"
	}

	var s strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("7")).
		Background(lipgloss.Color("4")).
		Width(model.width).
		Padding(0, 1)

	header := fmt.Sprintf("Select Action for %s", m.targetProject.Name)
	s.WriteString(headerStyle.Render(header))
	s.WriteString("\n\n")

	// Project info
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	s.WriteString(infoStyle.Render(fmt.Sprintf("Project: %s\n", m.targetProject.Name)))
	s.WriteString(infoStyle.Render(fmt.Sprintf("Status: %s\n", m.targetProject.Status)))
	s.WriteString(infoStyle.Render(fmt.Sprintf("Config Files: %s\n", m.targetProject.ConfigFiles)))
	s.WriteString("\n")

	// Actions list
	s.WriteString("Available Actions:\n\n")

	for i, action := range m.actions {
		prefix := "  "
		if i == m.selectedAction {
			prefix = "> "
		}

		// Color based on aggressive flag
		var actionStyle lipgloss.Style
		if action.Aggressive {
			actionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1")) // Red for aggressive
		} else {
			actionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // Green for safe
		}

		if i == m.selectedAction {
			actionStyle = actionStyle.Bold(true).Background(lipgloss.Color("237"))
		}

		line := fmt.Sprintf("%s[%s] %s - %s", prefix, action.Key, action.Name, action.Description)
		s.WriteString(actionStyle.Render(line))
		s.WriteString("\n")
	}

	// Footer
	s.WriteString("\n")
	footerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	s.WriteString(footerStyle.Render("Use ↑/↓ to select, Enter to execute, Esc to cancel"))

	return s.String()
}

// HandleUp moves selection up
func (m *ComposeProjectActionViewModel) HandleUp() tea.Cmd {
	if m.selectedAction > 0 {
		m.selectedAction--
	}
	return nil
}

// HandleDown moves selection down
func (m *ComposeProjectActionViewModel) HandleDown() tea.Cmd {
	if m.selectedAction < len(m.actions)-1 {
		m.selectedAction++
	}
	return nil
}

// HandleSelect executes the selected action
func (m *ComposeProjectActionViewModel) HandleSelect(model *Model) tea.Cmd {
	if m.selectedAction >= 0 && m.selectedAction < len(m.actions) {
		action := m.actions[m.selectedAction]
		// Remove ComposeProjectActionView from history by going back
		model.SwitchToPreviousView()
		// Now when the command execution view shows and user presses ESC,
		// they'll go back to the project list, not the action view
		return action.Handler(model, *m.targetProject)
	}
	return nil
}

// HandleBack returns to the previous view
func (m *ComposeProjectActionViewModel) HandleBack(model *Model) tea.Cmd {
	model.SwitchToPreviousView()
	return nil
}
