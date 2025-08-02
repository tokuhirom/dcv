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
	case HelpView:
		return m.renderHelpView()
	default:
		return "Unknown view"
	}
}

func (m *Model) renderComposeProcessList() string {
	var s strings.Builder

	slog.Info("Rendering container list",
		slog.Int("selectedContainer", m.selectedContainer),
		slog.Int("numContainers", len(m.composeContainers)))

	// Title with project name
	title := "Docker Compose"
	if m.projectName != "" {
		title = fmt.Sprintf("Docker Compose: %s", m.projectName)
	}
	s.WriteString(titleStyle.Render(title) + "\n")

	// Loading state
	if m.loading {
		s.WriteString("\nLoading...\n")
		return s.String()
	}

	// Error state
	if m.err != nil {
		s.WriteString("\n" + errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n")
		s.WriteString("\nPress q to quit\n")
		return s.String()
	}

	// Empty state
	if len(m.composeContainers) == 0 {
		s.WriteString("\nNo containers found.\n")
		s.WriteString("\nPress u to start services or p to switch to project list\n")
		return s.String()
	}

	// Container list
	s.WriteString("\n")

	// Create table with fixed widths
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return headerStyle
			}
			if row-1 == m.selectedContainer {
				return selectedStyle
			}
			return normalStyle
		}).
		Headers("SERVICE", "IMAGE", "STATUS", "PORTS")

	// Add rows with width control
	for _, container := range m.composeContainers {
		// Service name with dind indicator
		service := container.Service
		if container.IsDind() {
			service = dindStyle.Render("⬢ ") + service
		}

		// Truncate image name if too long
		image := container.Image
		if len(image) > 30 {
			image = image[:27] + "..."
		}

		// Status with color
		status := container.GetStatus()
		if strings.Contains(status, "Up") || strings.Contains(status, "running") {
			status = statusUpStyle.Render(status)
		} else {
			status = statusDownStyle.Render(status)
		}

		// Truncate ports if too long
		ports := container.GetPortsString()
		if len(ports) > 40 {
			ports = ports[:37] + "..."
		}

		t.Row(service, image, status, ports)
	}

	s.WriteString(t.Render() + "\n\n")

	// Show help hint
	s.WriteString(helpStyle.Render("Press ? for help"))

	return s.String()
}

func (m *Model) renderLogView() string {
	var s strings.Builder

	title := fmt.Sprintf("Logs: %s", m.containerName)
	if m.isDindLog {
		title = fmt.Sprintf("Logs: %s (in %s)", m.containerName, m.hostContainer)
	}
	s.WriteString(titleStyle.Render(title) + "\n\n")

	if m.loading && len(m.logs) == 0 {
		s.WriteString("Loading logs...\n")
		return s.String()
	}

	// Search mode indicator
	if m.searchMode {
		s.WriteString(searchStyle.Render(fmt.Sprintf("Search: %s", m.searchText)) + "\n\n")
	}

	// Calculate visible logs based on scroll position
	visibleHeight := m.height - 4 // Account for title and help
	if m.searchMode {
		visibleHeight -= 2
	}

	startIdx := m.logScrollY
	endIdx := startIdx + visibleHeight

	if endIdx > len(m.logs) {
		endIdx = len(m.logs)
	}

	// Display logs
	if len(m.logs) == 0 {
		s.WriteString("No logs available.\n")
	} else {
		for i := startIdx; i < endIdx; i++ {
			if i < len(m.logs) {
				s.WriteString(m.logs[i] + "\n")
			}
		}
	}

	// Scroll indicator
	if len(m.logs) > visibleHeight {
		scrollInfo := fmt.Sprintf(" [%d-%d/%d] ", startIdx+1, endIdx, len(m.logs))
		s.WriteString("\n" + helpStyle.Render(scrollInfo))
	}

	// Show help hint
	s.WriteString(helpStyle.Render("Press ? for help"))

	return s.String()
}

func (m *Model) renderDindList() string {
	var s strings.Builder

	title := titleStyle.Render(fmt.Sprintf("Docker in Docker: %s", m.currentDindHost))
	s.WriteString(title + "\n")

	if m.loading {
		s.WriteString("\nLoading...\n")
		return s.String()
	}

	if m.err != nil {
		s.WriteString("\n" + errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n")
		return s.String()
	}

	if len(m.dindContainers) == 0 {
		s.WriteString("\nNo containers running inside this dind container.\n")
		s.WriteString("\nPress ESC to go back\n")
		return s.String()
	}

	// Container list
	s.WriteString("\n")

	// Create table
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return headerStyle
			}
			if row-1 == m.selectedDindContainer {
				return selectedStyle
			}
			return normalStyle
		}).
		Headers("CONTAINER ID", "IMAGE", "STATUS", "NAMES")

	for _, container := range m.dindContainers {
		// Truncate container ID
		id := container.ID
		if len(id) > 12 {
			id = id[:12]
		}

		// Truncate image name
		image := container.Image
		if len(image) > 30 {
			image = image[:27] + "..."
		}

		// Status with color
		status := container.Status
		if strings.Contains(status, "Up") {
			status = statusUpStyle.Render(status)
		} else {
			status = statusDownStyle.Render(status)
		}

		t.Row(id, image, status, container.Names)
	}

	s.WriteString(t.Render() + "\n\n")

	// Show help hint
	s.WriteString(helpStyle.Render("Press ? for help"))

	return s.String()
}

func (m *Model) renderTopView() string {
	var s strings.Builder

	title := titleStyle.Render(fmt.Sprintf("Process Info: %s", m.topService))
	s.WriteString(title + "\n\n")

	if m.loading {
		s.WriteString("Loading...\n")
		return s.String()
	}

	if m.err != nil {
		s.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n")
		return s.String()
	}

	if m.topOutput == "" {
		s.WriteString("No process information available.\n")
	} else {
		// Display the raw top output
		lines := strings.Split(m.topOutput, "\n")
		visibleHeight := m.height - 5 // Account for title and help

		for i, line := range lines {
			if i >= visibleHeight {
				break
			}
			s.WriteString(line + "\n")
		}
	}

	s.WriteString("\n")

	// Show help hint
	s.WriteString(helpStyle.Render("Press ? for help"))

	return s.String()
}

func (m *Model) renderStatsView() string {
	var s strings.Builder

	title := titleStyle.Render("Container Resource Usage")
	s.WriteString(title + "\n")

	if m.loading {
		s.WriteString("\nLoading...\n")
		return s.String()
	}

	if m.err != nil {
		s.WriteString("\n" + errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n")
		return s.String()
	}

	if len(m.stats) == 0 {
		s.WriteString("\nNo stats available.\n")
		return s.String()
	}

	// Stats table
	s.WriteString("\n")

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return headerStyle
			}
			return normalStyle
		}).
		Headers("NAME", "CPU %", "MEM USAGE", "MEM %", "NET I/O", "BLOCK I/O")

	for _, stat := range m.stats {
		// Truncate name if too long
		name := stat.Name
		if len(name) > 20 {
			name = name[:17] + "..."
		}

		// Color CPU usage
		cpu := stat.CPUPerc
		if cpuVal := strings.TrimSuffix(cpu, "%"); cpuVal != "" {
			if val, err := fmt.Sscanf(cpuVal, "%f", new(float64)); err == nil && val > 0 {
				if *new(float64) > 80.0 {
					cpu = errorStyle.Render(cpu)
				} else if *new(float64) > 50.0 {
					cpu = searchStyle.Render(cpu)
				}
			}
		}

		t.Row(name, cpu, stat.MemUsage, stat.MemPerc, stat.NetIO, stat.BlockIO)
	}

	s.WriteString(t.Render() + "\n\n")

	// Show help hint
	s.WriteString(helpStyle.Render("Press ? for help"))

	return s.String()
}

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

func (m *Model) renderDockerList() string {
	var s strings.Builder

	title := "Docker Containers"
	if m.showAll {
		title += " (all)"
	}
	s.WriteString(titleStyle.Render(title) + "\n")

	if m.loading {
		s.WriteString("\nLoading...\n")
		return s.String()
	}

	if m.err != nil {
		s.WriteString("\n" + errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n")
		return s.String()
	}

	if len(m.dockerContainers) == 0 {
		s.WriteString("\nNo containers found.\n")
		return s.String()
	}

	// Container list
	s.WriteString("\n")

	// Define consistent styles for table cells
	idStyle := lipgloss.NewStyle().Width(12)
	imageStyle := lipgloss.NewStyle().Width(30)
	statusStyle := lipgloss.NewStyle().Width(20)
	portsStyle := lipgloss.NewStyle().Width(30)
	nameStyle := lipgloss.NewStyle().Width(20)

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			baseStyle := normalStyle
			if row == 0 {
				baseStyle = headerStyle
			} else if row-1 == m.selectedDockerContainer {
				baseStyle = selectedStyle
			}

			// Apply column-specific styling
			switch col {
			case 0:
				return baseStyle.Inherit(idStyle)
			case 1:
				return baseStyle.Inherit(imageStyle)
			case 2:
				return baseStyle.Inherit(statusStyle)
			case 3:
				return baseStyle.Inherit(portsStyle)
			case 4:
				return baseStyle.Inherit(nameStyle)
			default:
				return baseStyle
			}
		}).
		Headers("CONTAINER ID", "IMAGE", "STATUS", "PORTS", "NAMES")

	for _, container := range m.dockerContainers {
		// Truncate container ID
		id := idStyle.Render(container.ID[:12])

		// Truncate image name
		image := container.Image
		if len(image) > 30 {
			image = image[:27] + "..."
		}
		image = imageStyle.Render(image)

		// Status with color
		status := container.Status
		if strings.Contains(status, "Up") || strings.Contains(status, "running") {
			status = statusUpStyle.Render(status)
		} else {
			status = statusDownStyle.Render(status)
		}

		// Truncate ports
		ports := container.Ports
		if len(ports) > 30 {
			ports = ports[:27] + "..."
		}
		ports = portsStyle.Render(ports)

		name := nameStyle.Render(container.Names)

		t.Row(id, image, status, ports, name)
	}

	s.WriteString(t.Render() + "\n\n")

	// Show help hint
	s.WriteString(helpStyle.Render("Press ? for help"))

	return s.String()
}

func (m *Model) renderHelpView() string {
	var s strings.Builder

	// Title
	title := titleStyle.Render("Help")
	s.WriteString(title + "\n\n")

	// Get key configurations based on previous view
	var configs []KeyConfig
	viewName := ""

	switch m.previousView {
	case ComposeProcessListView:
		configs = m.processListViewHandlers
		viewName = "Compose Process List"
	case LogView:
		configs = m.logViewHandlers
		viewName = "Log View"
	case DindComposeProcessListView:
		configs = m.dindListViewHandlers
		viewName = "Docker in Docker"
	case TopView:
		configs = m.topViewHandlers
		viewName = "Process Info"
	case StatsView:
		configs = m.statsViewHandlers
		viewName = "Container Stats"
	case ProjectListView:
		configs = m.projectListViewHandlers
		viewName = "Project List"
	case DockerContainerListView:
		configs = m.dockerListViewHandlers
		viewName = "Docker Containers"
	case ImageListView:
		configs = m.imageListViewHandlers
		viewName = "Docker Images"
	case NetworkListView:
		configs = m.networkListViewHandlers
		viewName = "Docker Networks"
	case FileBrowserView:
		configs = m.fileBrowserHandlers
		viewName = "File Browser"
	case FileContentView:
		configs = m.fileContentHandlers
		viewName = "File Content"
	case InspectView:
		configs = m.inspectViewHandlers
		viewName = "Inspect"
	}

	// Show view name
	s.WriteString(headerStyle.Render(fmt.Sprintf("Keyboard shortcuts for: %s", viewName)) + "\n\n")

	// Calculate max scroll
	totalLines := len(configs) + 5 // title + view name + margins
	visibleLines := m.height - 4   // footer + margins
	maxScroll := totalLines - visibleLines
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.helpScrollY > maxScroll {
		m.helpScrollY = maxScroll
	}

	// Render key bindings
	visibleConfigs := configs
	if m.helpScrollY > 0 && m.helpScrollY < len(configs) {
		endIdx := m.helpScrollY + visibleLines - 5
		if endIdx > len(configs) {
			endIdx = len(configs)
		}
		visibleConfigs = configs[m.helpScrollY:endIdx]
	}

	for _, config := range visibleConfigs {
		if len(config.Keys) > 0 {
			key := config.Keys[0]
			if len(config.Keys) > 1 {
				key = strings.Join(config.Keys, "/")
			}
			keyStr := lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Bold(true).
				Render(fmt.Sprintf("%-15s", key))
			descStr := normalStyle.Render(config.Description)
			s.WriteString(fmt.Sprintf("%s %s\n", keyStr, descStr))
		}
	}

	// Footer
	footer := "\n" + helpStyle.Render("Press ESC or q to go back")
	footerHeight := 2
	contentHeight := m.height - footerHeight
	currentHeight := strings.Count(s.String(), "\n") + 1
	if currentHeight < contentHeight {
		s.WriteString(strings.Repeat("\n", contentHeight-currentHeight))
	}
	s.WriteString(footer)

	return s.String()
}