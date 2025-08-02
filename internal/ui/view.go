package ui

import (
	"fmt"
	"log/slog"
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
func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	switch m.currentView {
	case ComposeProcessListView:
		return m.renderComposeProcessList()
	case LogView:
		return m.renderLogView()
	case DindComposeProcessListView:
		return m.renderDindList()
	case TopView:
		return m.renderTopView()
	case StatsView:
		return m.renderStatsView()
	case ProjectListView:
		return m.renderProjectList()
	case DockerContainerListView:
		return m.renderDockerList()
	case ImageListView:
		return m.renderImageList()
	case NetworkListView:
		return m.renderNetworkList()
	case FileBrowserView:
		return m.renderFileBrowser()
	case FileContentView:
		return m.renderFileContent()
	case InspectView:
		return m.renderInspectView()
	default:
		return "Unknown view"
	}
}

func (m *Model) renderComposeProcessList() string {
	var s strings.Builder

	slog.Info("Rendering container list",
		slog.String("projectName", m.projectName))

	title := "Docker Compose Processes"
	if m.showAll {
		title += " (All)"
	}
	if m.projectName != "" {
		title += fmt.Sprintf(" [Compose: %s]", m.projectName)
	}
	s.WriteString(titleStyle.Render(title) + "\n\n")

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
		s.WriteString("Loading composeContainers...")
		return s.String()
	}

	if len(m.composeContainers) == 0 {
		s.WriteString("No composeContainers found\n")
		s.WriteString("\n" + helpStyle.Render("Press 'r' to refresh, 'q' to quit"))
		return s.String()
	}

	// Create table
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		Headers("NAME", "IMAGE", "SERVICE", "STATUS", "STATE")

	// Style headers
	t.StyleFunc(func(row, col int) lipgloss.Style {
		if row == 0 {
			return headerStyle
		}
		return normalStyle
	})

	// Add rows
	for i, container := range m.composeContainers {
		nameStyle := normalStyle
		if container.IsDind() {
			nameStyle = dindStyle
		}
		if i == m.selectedContainer {
			nameStyle = selectedStyle
		}

		var statusStyle lipgloss.Style
		if i == m.selectedContainer {
			statusStyle = selectedStyle
		} else if strings.Contains(container.GetStatus(), "Up") || strings.Contains(container.State, "running") {
			statusStyle = statusUpStyle
		} else {
			statusStyle = statusDownStyle
		}

		name := nameStyle.Render(container.Name)
		var image, service string
		if i == m.selectedContainer {
			image = selectedStyle.Render(container.Image)
			service = selectedStyle.Render(container.Service)
		} else {
			image = normalStyle.Render(container.Image)
			service = normalStyle.Render(container.Service)
		}
		status := statusStyle.Render(container.GetStatus())
		state := normalStyle.Render(container.State)

		t.Row(name, image, service, status, state)
	}

	s.WriteString(t.Render() + "\n\n")

	// Help text
	helpText := m.GetStyledHelpText()
	if helpText != "" {
		s.WriteString(helpText)
	}

	return s.String()
}

func (m *Model) renderLogView() string {
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
	helpText := m.GetStyledHelpText()
	if helpText != "" {
		s.WriteString(helpText)
	}

	return s.String()
}

func (m *Model) renderDindList() string {
	var s strings.Builder

	title := titleStyle.Render(fmt.Sprintf("Docker in Docker: %s", m.currentDindHost))
	s.WriteString(title + "\n\n")

	if m.err != nil {
		s.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n")
		s.WriteString("\n" + helpStyle.Render("Press 'Esc' to go back, 'q' to quit"))
		return s.String()
	}

	if m.loading {
		slog.Info("Loading dind composeContainers...")
		return s.String()
	}

	if len(m.dindContainers) == 0 {
		s.WriteString("No composeContainers found in dind\n")
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
		name := style.Render(container.Names)

		t.Row(id, image, status, name)
	}

	s.WriteString(t.Render() + "\n\n")

	// Help text
	helpText := m.GetStyledHelpText()
	if helpText != "" {
		s.WriteString(helpText)
	}

	return s.String()
}

func (m *Model) renderTopView() string {
	var s strings.Builder

	title := titleStyle.Render(fmt.Sprintf("Process Info: %s", m.topService))
	s.WriteString(title + "\n\n")

	if m.err != nil {
		s.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n")
		s.WriteString("\n" + helpStyle.Render("Press 'Esc' to go back, 'q' to quit"))
		return s.String()
	}

	if m.loading {
		s.WriteString("Loading process information...")
		return s.String()
	}

	// Display the top output
	if m.topOutput != "" {
		s.WriteString(m.topOutput)
	} else {
		s.WriteString("No process information available\n")
	}

	// Fill remaining space
	outputLines := strings.Count(m.topOutput, "\n")
	viewHeight := m.height - 5
	for i := outputLines; i < viewHeight; i++ {
		s.WriteString("\n")
	}

	// Help text
	helpText := m.GetStyledHelpText()
	if helpText != "" {
		s.WriteString(helpText)
	}

	return s.String()
}

func (m *Model) renderStatsView() string {
	var s strings.Builder

	title := titleStyle.Render("Container Resource Usage")
	s.WriteString(title + "\n\n")

	if m.err != nil {
		s.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n")
		s.WriteString("\n" + helpStyle.Render("Press 'Esc' to go back, 'q' to quit"))
		return s.String()
	}

	if m.loading {
		s.WriteString("Loading stats...")
		return s.String()
	}

	if len(m.stats) == 0 {
		s.WriteString("No stats available\n")
		s.WriteString("\n" + helpStyle.Render("Press 'r' to refresh, 'Esc' to go back"))
		return s.String()
	}

	// Create table
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		Headers("SERVICE", "CPU %", "MEM USAGE", "MEM %", "NET I/O", "BLOCK I/O", "PIDS")

	// Style headers
	t.StyleFunc(func(row, col int) lipgloss.Style {
		if row == 0 {
			return headerStyle
		}
		return normalStyle
	})

	// Add rows
	for _, stat := range m.stats {
		t.Row(
			stat.Service,
			stat.CPUPerc,
			stat.MemUsage,
			stat.MemPerc,
			stat.NetIO,
			stat.BlockIO,
			stat.PIDs,
		)
	}

	s.WriteString(t.Render() + "\n\n")

	// Help text
	helpText := m.GetStyledHelpText()
	if helpText != "" {
		s.WriteString(helpText)
	}

	return s.String()
}

func (m *Model) renderProjectList() string {
	var s strings.Builder

	title := titleStyle.Render("Docker Compose Projects")
	s.WriteString(title + "\n\n")

	if m.err != nil {
		s.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n")
		s.WriteString("\n" + helpStyle.Render("Press 'q' to quit"))
		return s.String()
	}

	if m.loading {
		s.WriteString("Loading projects...")
		return s.String()
	}

	if len(m.projects) == 0 {
		s.WriteString("No Docker Compose projects found\n")
		s.WriteString("\nStart a compose project or specify a compose file with -f flag\n")
		s.WriteString("\n" + helpStyle.Render("Press 'r' to refresh, 'q' to quit"))
		return s.String()
	}

	// Create table
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		Headers("NAME", "STATUS", "CONFIG FILES")

	// Style headers
	t.StyleFunc(func(row, col int) lipgloss.Style {
		if row == 0 {
			return headerStyle
		}
		return normalStyle
	})

	// Add rows
	for i, project := range m.projects {
		style := normalStyle
		if i == m.selectedProject {
			style = selectedStyle
		}

		name := style.Render(project.Name)
		status := style.Render(project.Status)
		configFiles := style.Render(project.ConfigFiles)

		t.Row(name, status, configFiles)
	}

	s.WriteString(t.Render() + "\n\n")

	// Help text
	helpText := m.GetStyledHelpText()
	if helpText != "" {
		s.WriteString(helpText)
	}

	return s.String()
}

func (m *Model) renderDockerList() string {
	var s strings.Builder

	title := "Docker Containers"
	if m.showAll {
		title += " (All)"
	}
	s.WriteString(titleStyle.Render(title) + "\n\n")

	if m.err != nil {
		s.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n")
		s.WriteString("\n" + helpStyle.Render("Press 'esc' or 'q' to go back"))
		return s.String()
	}

	if m.loading {
		s.WriteString("Loading composeContainers...")
		return s.String()
	}

	if len(m.dockerContainers) == 0 {
		s.WriteString("No composeContainers found\n")
		s.WriteString("\n" + helpStyle.Render("Press 'r' to refresh, 'a' to toggle all, 'esc' to go back"))
		return s.String()
	}

	// Create table
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		Headers("CONTAINER ID", "IMAGE", "STATUS", "PORTS", "NAMES")

	// Style headers
	t.StyleFunc(func(row, col int) lipgloss.Style {
		if row == 0 {
			return headerStyle
		}
		return normalStyle
	})

	// Add rows
	for i, container := range m.dockerContainers {
		var idStyle, imageStyle, statusStyle, portsStyle, nameStyle lipgloss.Style

		if i == m.selectedDockerContainer {
			idStyle = selectedStyle
			imageStyle = selectedStyle
			statusStyle = selectedStyle
			portsStyle = selectedStyle
			nameStyle = selectedStyle
		} else {
			idStyle = normalStyle
			imageStyle = normalStyle
			portsStyle = normalStyle
			nameStyle = normalStyle

			if strings.Contains(container.Status, "Up") || strings.Contains(container.State, "running") {
				statusStyle = statusUpStyle
			} else {
				statusStyle = statusDownStyle
			}
		}

		id := idStyle.Render(container.ID[:12])
		image := imageStyle.Render(container.Image)
		status := statusStyle.Render(container.Status)

		// Truncate ports if too long
		ports := container.Ports
		if len(ports) > 30 {
			ports = ports[:27] + "..."
		}
		ports = portsStyle.Render(ports)

		name := nameStyle.Render(container.Names)

		t.Row(id, image, status, ports, name)
	}

	s.WriteString(t.Render() + "\n\n")

	// Help text
	helpText := m.GetStyledHelpText()
	if helpText != "" {
		s.WriteString(helpText)
	}

	return s.String()
}
