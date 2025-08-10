package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tokuhirom/dcv/internal/docker"
)

// CommandAction represents a container operation
type CommandAction struct {
	Key         string
	Name        string
	Description string
	Aggressive  bool
	Handler     func(m *Model, container *docker.Container) tea.Cmd
}

// CommandActionViewModel manages the command action selection view
type CommandActionViewModel struct {
	actions         []CommandAction
	selectedAction  int
	targetContainer *docker.Container
}

// Initialize sets up the action view with available commands for a container
func (m *CommandActionViewModel) Initialize(container *docker.Container) {
	m.targetContainer = container
	m.selectedAction = 0

	// Define available actions based on container state
	m.actions = []CommandAction{}

	// Always available actions
	m.actions = append(m.actions, CommandAction{
		Key:         "Enter",
		Name:        "View Logs",
		Description: "Show container logs",
		Aggressive:  false,
		Handler: func(model *Model, c *docker.Container) tea.Cmd {
			_, cmd := model.CmdLog(tea.KeyMsg{})
			return cmd
		},
	})

	m.actions = append(m.actions, CommandAction{
		Key:         "I",
		Name:        "Inspect",
		Description: "View container configuration",
		Aggressive:  false,
		Handler: func(model *Model, c *docker.Container) tea.Cmd {
			_, cmd := model.CmdInspect(tea.KeyMsg{})
			return cmd
		},
	})

	m.actions = append(m.actions, CommandAction{
		Key:         "f",
		Name:        "Browse Files",
		Description: "Browse container filesystem",
		Aggressive:  false,
		Handler: func(model *Model, c *docker.Container) tea.Cmd {
			_, cmd := model.CmdFileBrowse(tea.KeyMsg{})
			return cmd
		},
	})

	m.actions = append(m.actions, CommandAction{
		Key:         "!",
		Name:        "Execute Shell",
		Description: "Execute /bin/sh in container",
		Aggressive:  false,
		Handler: func(model *Model, c *docker.Container) tea.Cmd {
			_, cmd := model.CmdShell(tea.KeyMsg{})
			return cmd
		},
	})

	// State-dependent actions
	if container.GetState() == "running" {
		m.actions = append(m.actions, CommandAction{
			Key:         "S",
			Name:        "Stop",
			Description: "Stop the container",
			Aggressive:  true,
			Handler: func(model *Model, c *docker.Container) tea.Cmd {
				_, cmd := model.CmdStop(tea.KeyMsg{})
				return cmd
			},
		})

		m.actions = append(m.actions, CommandAction{
			Key:         "R",
			Name:        "Restart",
			Description: "Restart the container",
			Aggressive:  true,
			Handler: func(model *Model, c *docker.Container) tea.Cmd {
				_, cmd := model.CmdRestart(tea.KeyMsg{})
				return cmd
			},
		})

		m.actions = append(m.actions, CommandAction{
			Key:         "K",
			Name:        "Kill",
			Description: "Force kill the container",
			Aggressive:  true,
			Handler: func(model *Model, c *docker.Container) tea.Cmd {
				_, cmd := model.CmdKill(tea.KeyMsg{})
				return cmd
			},
		})

		m.actions = append(m.actions, CommandAction{
			Key:         "P",
			Name:        "Pause",
			Description: "Pause the container",
			Aggressive:  true,
			Handler: func(model *Model, c *docker.Container) tea.Cmd {
				_, cmd := model.CmdPause(tea.KeyMsg{})
				return cmd
			},
		})
	} else if container.GetState() == "paused" {
		m.actions = append(m.actions, CommandAction{
			Key:         "P",
			Name:        "Unpause",
			Description: "Resume the container",
			Aggressive:  true,
			Handler: func(model *Model, c *docker.Container) tea.Cmd {
				_, cmd := model.CmdPause(tea.KeyMsg{})
				return cmd
			},
		})
	} else if container.GetState() == "exited" || container.GetState() == "created" {
		m.actions = append(m.actions, CommandAction{
			Key:         "U",
			Name:        "Start",
			Description: "Start the container",
			Aggressive:  false,
			Handler: func(model *Model, c *docker.Container) tea.Cmd {
				_, cmd := model.CmdStart(tea.KeyMsg{})
				return cmd
			},
		})

		m.actions = append(m.actions, CommandAction{
			Key:         "D",
			Name:        "Delete",
			Description: "Remove the container",
			Aggressive:  true,
			Handler: func(model *Model, c *docker.Container) tea.Cmd {
				_, cmd := model.CmdDelete(tea.KeyMsg{})
				return cmd
			},
		})
	}
}

// render displays the action selection menu
func (m *CommandActionViewModel) render(model *Model) string {
	if m.targetContainer == nil {
		return "No container selected"
	}

	var s strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("7")).
		Background(lipgloss.Color("4")).
		Width(model.width).
		Padding(0, 1)

	header := fmt.Sprintf("Select Action for %s", m.targetContainer.GetName())
	s.WriteString(headerStyle.Render(header))
	s.WriteString("\n\n")

	// Container info
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	s.WriteString(infoStyle.Render(fmt.Sprintf("Container: %s\n", m.targetContainer.GetName())))
	s.WriteString(infoStyle.Render(fmt.Sprintf("State: %s\n", m.targetContainer.GetState())))
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
func (m *CommandActionViewModel) HandleUp() tea.Cmd {
	if m.selectedAction > 0 {
		m.selectedAction--
	}
	return nil
}

// HandleDown moves selection down
func (m *CommandActionViewModel) HandleDown() tea.Cmd {
	if m.selectedAction < len(m.actions)-1 {
		m.selectedAction++
	}
	return nil
}

// HandleSelect executes the selected action
func (m *CommandActionViewModel) HandleSelect(model *Model) tea.Cmd {
	if m.selectedAction >= 0 && m.selectedAction < len(m.actions) {
		action := m.actions[m.selectedAction]
		// Remove CommandActionView from history by going back
		model.SwitchToPreviousView()
		// Now when the command execution view shows and user presses ESC,
		// they'll go back to the container list, not the action view
		return action.Handler(model, m.targetContainer)
	}
	return nil
}

// HandleBack returns to the previous view
func (m *CommandActionViewModel) HandleBack(model *Model) tea.Cmd {
	model.SwitchToPreviousView()
	return nil
}
