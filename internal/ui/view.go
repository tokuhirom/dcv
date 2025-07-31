package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")).
			MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Background(lipgloss.Color("235"))

	normalStyle = lipgloss.NewStyle()

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("226"))

	dindStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42"))

	statusUpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42"))

	statusDownStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	searchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")).
			Bold(true)
)

// View returns the view for the current model
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	switch m.currentView {
	case ProcessListView:
		return m.renderProcessList()
	case LogView:
		return m.renderLogView()
	case DindProcessListView:
		return m.renderDindList()
	default:
		return "Unknown view"
	}
}

func (m Model) renderProcessList() string {
	var s strings.Builder

	title := titleStyle.Render("Docker Compose Processes")
	s.WriteString(title + "\n\n")

	if m.err != nil {
		if strings.Contains(m.err.Error(), "no configuration file provided") {
			s.WriteString(errorStyle.Render("No docker-compose.yml found in current directory") + "\n")
			s.WriteString("Please run from a directory with docker-compose.yml\n")
		} else {
			s.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n")
		}
		s.WriteString("\n" + helpStyle.Render("Press 'q' to quit"))
		return s.String()
	}

	if m.loading {
		s.WriteString("Loading containers...")
		return s.String()
	}

	if len(m.processes) == 0 {
		s.WriteString("No containers found\n")
		s.WriteString("\n" + helpStyle.Render("Press 'r' to refresh, 'q' to quit"))
		return s.String()
	}

	// Create table
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		Headers("NAME", "IMAGE", "SERVICE", "STATUS")

	// Style headers
	t.StyleFunc(func(row, col int) lipgloss.Style {
		if row == 0 {
			return headerStyle
		}
		return normalStyle
	})

	// Add rows
	for i, process := range m.processes {
		nameStyle := normalStyle
		if process.IsDind {
			nameStyle = dindStyle
		}
		if i == m.selectedProcess {
			nameStyle = selectedStyle
		}

		statusStyle := normalStyle
		if strings.Contains(process.Status, "Up") {
			statusStyle = statusUpStyle
		} else {
			statusStyle = statusDownStyle
		}
		if i == m.selectedProcess {
			statusStyle = selectedStyle
		}

		name := nameStyle.Render(process.Name)
		image := normalStyle.Render(process.Image)
		service := normalStyle.Render(process.Service)
		status := statusStyle.Render(process.Status)

		if i == m.selectedProcess {
			image = selectedStyle.Render(process.Image)
			service = selectedStyle.Render(process.Service)
		}

		t.Row(name, image, service, status)
	}

	s.WriteString(t.Render() + "\n\n")

	// Help text
	help := helpStyle.Render("↑/k: up • ↓/j: down • Enter: logs • d: dind • r: refresh • q: quit")
	s.WriteString(help)

	// Show last command if available
	if m.lastCommand != "" {
		s.WriteString("\n" + helpStyle.Render(fmt.Sprintf("Last command: %s", m.lastCommand)))
	}

	return s.String()
}

func (m Model) renderLogView() string {
	var s strings.Builder

	title := fmt.Sprintf("Logs: %s", m.containerName)
	if m.isDindLog {
		title = fmt.Sprintf("Logs: %s (in %s)", m.containerName, m.hostContainer)
	}
	if m.searchMode {
		title = fmt.Sprintf("Search: %s", m.searchText)
		title = searchStyle.Render(title)
	} else {
		title = titleStyle.Render(title)
	}
	s.WriteString(title + "\n\n")

	// Log content
	viewHeight := m.height - 4
	if m.lastCommand != "" {
		viewHeight-- // Account for command line
	}
	startIdx := m.logScrollY
	endIdx := startIdx + viewHeight

	if endIdx > len(m.logs) {
		endIdx = len(m.logs)
	}

	if len(m.logs) == 0 {
		s.WriteString("Loading logs... (fetching last 1000 lines)\n")
	} else if startIdx < len(m.logs) {
		for i := startIdx; i < endIdx; i++ {
			s.WriteString(m.logs[i] + "\n")
		}
	}

	// Fill remaining space
	linesShown := endIdx - startIdx
	for i := linesShown; i < viewHeight; i++ {
		s.WriteString("\n")
	}

	// Help text
	help := helpStyle.Render("↑/k: up • ↓/j: down • G: end • g: start • /: search • Esc/q: back")
	s.WriteString(help)

	// Show last command if available
	if m.lastCommand != "" {
		s.WriteString("\n" + helpStyle.Render(fmt.Sprintf("Command: %s", m.lastCommand)))
	}

	return s.String()
}

func (m Model) renderDindList() string {
	var s strings.Builder

	title := titleStyle.Render(fmt.Sprintf("Docker in Docker: %s", m.currentDindHost))
	s.WriteString(title + "\n\n")

	if m.err != nil {
		s.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n")
		s.WriteString("\n" + helpStyle.Render("Press 'Esc' to go back, 'q' to quit"))
		return s.String()
	}

	if m.loading {
		s.WriteString("Loading containers...")
		return s.String()
	}

	if len(m.dindContainers) == 0 {
		s.WriteString("No containers found in dind\n")
		s.WriteString("\n" + helpStyle.Render("Press 'r' to refresh, 'Esc' to go back"))
		return s.String()
	}

	// Create table
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		Headers("CONTAINER ID", "IMAGE", "STATUS", "NAME")

	// Style headers
	t.StyleFunc(func(row, col int) lipgloss.Style {
		if row == 0 {
			return headerStyle
		}
		return normalStyle
	})

	// Add rows
	for i, container := range m.dindContainers {
		style := normalStyle
		if i == m.selectedDindContainer {
			style = selectedStyle
		}

		id := style.Render(container.ID[:12])
		image := style.Render(container.Image)
		status := style.Render(container.Status)
		name := style.Render(container.Name)

		t.Row(id, image, status, name)
	}

	s.WriteString(t.Render() + "\n\n")

	// Help text
	help := helpStyle.Render("↑/k: up • ↓/j: down • Enter: logs • r: refresh • Esc: back • q: quit")
	s.WriteString(help)

	// Show last command if available
	if m.lastCommand != "" {
		s.WriteString("\n" + helpStyle.Render(fmt.Sprintf("Last command: %s", m.lastCommand)))
	}

	return s.String()
}