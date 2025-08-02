package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// KeyHandler represents a function that handles a key press
type KeyHandler func(msg tea.KeyMsg) (tea.Model, tea.Cmd)

// KeyConfig represents a key binding configuration
type KeyConfig struct {
	Keys        []string
	Description string
	KeyHandler  KeyHandler
}

// Common navigation handlers
func (m *Model) SelectUpContainer(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedContainer > 0 {
		m.selectedContainer--
	}
	return m, nil
}

func (m *Model) SelectDownContainer(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedContainer < len(m.containers)-1 {
		m.selectedContainer++
	}
	return m, nil
}

// Common selection handlers for different views
func (m *Model) SelectUpDindContainer(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedDindContainer > 0 {
		m.selectedDindContainer--
	}
	return m, nil
}

func (m *Model) SelectDownDindContainer(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedDindContainer < len(m.dindContainers)-1 {
		m.selectedDindContainer++
	}
	return m, nil
}

func (m *Model) SelectUpProject(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedProject > 0 {
		m.selectedProject--
	}
	return m, nil
}

func (m *Model) SelectDownProject(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedProject < len(m.projects)-1 {
		m.selectedProject++
	}
	return m, nil
}

func (m *Model) SelectUpDockerContainer(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedDockerContainer > 0 {
		m.selectedDockerContainer--
	}
	return m, nil
}

func (m *Model) SelectDownDockerContainer(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedDockerContainer < len(m.dockerContainers)-1 {
		m.selectedDockerContainer++
	}
	return m, nil
}

// Log view navigation handlers
func (m *Model) ScrollLogUp(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.logScrollY > 0 {
		m.logScrollY--
	}
	return m, nil
}

func (m *Model) ScrollLogDown(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	maxScroll := len(m.logs) - (m.height - 4)
	if m.logScrollY < maxScroll && maxScroll > 0 {
		m.logScrollY++
	}
	return m, nil
}

func (m *Model) GoToLogEnd(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	maxScroll := len(m.logs) - (m.height - 4)
	if maxScroll > 0 {
		m.logScrollY = maxScroll
	}
	return m, nil
}

func (m *Model) GoToLogStart(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.logScrollY = 0
	return m, nil
}

func (m *Model) StartSearch(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.searchMode = true
	m.searchText = ""
	return m, nil
}

// View-specific handlers
func (m *Model) ShowComposeLog(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedContainer < len(m.containers) {
		process := m.containers[m.selectedContainer]
		m.containerName = process.Name
		m.isDindLog = false
		m.currentView = LogView
		m.logs = []string{}
		m.logScrollY = 0
		return m, streamLogs(m.dockerClient, process.ID, false, "")
	}
	return m, nil
}

func (m *Model) ShowDindLog(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedDindContainer < len(m.dindContainers) {
		container := m.dindContainers[m.selectedDindContainer]
		m.containerName = container.Names
		m.hostContainer = m.currentDindHost
		m.isDindLog = true
		m.currentView = LogView
		m.logs = []string{}
		m.logScrollY = 0
		return m, streamLogs(m.dockerClient, container.Names, true, m.currentDindContainerID)
	}
	return m, nil
}

func (m *Model) ShowDindProcessList(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedContainer < len(m.containers) {
		container := m.containers[m.selectedContainer]
		if container.IsDind() {
			m.currentDindHost = container.Name
			m.currentDindContainerID = container.ID
			m.currentView = DindComposeProcessListView
			m.loading = true
			return m, loadDindContainers(m.dockerClient, container.ID)
		}
	}
	return m, nil
}

func (m *Model) ShowDockerContainerList(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.currentView = DockerContainerListView
	m.loading = true
	return m, loadDockerContainers(m.dockerClient, m.showAll)
}

func (m *Model) RefreshProcessList(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.loading = true
	return m, loadProcesses(m.dockerClient, m.projectName, m.showAll)
}

func (m *Model) RefreshDindList(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.loading = true
	return m, loadDindContainers(m.dockerClient, m.currentDindContainerID)
}

func (m *Model) RefreshProjects(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.loading = true
	return m, loadProjects(m.dockerClient)
}

func (m *Model) RefreshTop(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.loading = true
	return m, loadTop(m.dockerClient, m.projectName, m.topService)
}

func (m *Model) RefreshStats(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.loading = true
	return m, loadStats(m.dockerClient)
}

func (m *Model) ToggleAllContainers(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.showAll = !m.showAll
	m.loading = true
	return m, loadProcesses(m.dockerClient, m.projectName, m.showAll)
}

func (m *Model) ShowStatsView(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.currentView = StatsView
	m.loading = true
	return m, loadStats(m.dockerClient)
}

func (m *Model) ShowTopView(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedContainer < len(m.containers) {
		container := m.containers[m.selectedContainer]
		m.topService = container.Service
		m.currentView = TopView
		m.loading = true
		return m, loadTop(m.dockerClient, m.projectName, container.Service)
	}
	return m, nil
}

func (m *Model) KillContainer(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedContainer < len(m.containers) {
		container := m.containers[m.selectedContainer]
		m.loading = true
		return m, killService(m.dockerClient, container.ID)
	}
	return m, nil
}

func (m *Model) StopContainer(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedContainer < len(m.containers) {
		container := m.containers[m.selectedContainer]
		m.loading = true
		return m, stopService(m.dockerClient, container.ID)
	}
	return m, nil
}

func (m *Model) UpService(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedContainer < len(m.containers) {
		container := m.containers[m.selectedContainer]
		m.loading = true
		return m, startService(m.dockerClient, container.Service)
	}
	return m, nil
}

func (m *Model) RestartContainer(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedContainer < len(m.containers) {
		container := m.containers[m.selectedContainer]
		m.loading = true
		return m, restartService(m.dockerClient, container.ID)
	}
	return m, nil
}

func (m *Model) DeleteContainer(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedContainer < len(m.containers) {
		container := m.containers[m.selectedContainer]
		// Only allow removing stopped containers
		if !strings.Contains(container.GetStatus(), "Up") && !strings.Contains(container.State, "running") {
			m.loading = true
			return m, removeService(m.dockerClient, container.ID)
		}
	}
	return m, nil
}

func (m *Model) DeployProject(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedContainer < len(m.containers) {
		m.loading = true
		return m, up(m.dockerClient, m.projectName)
	}
	return m, nil
}

func (m *Model) ShowProjectList(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.currentView = ProjectListView
	m.loading = true
	return m, loadProjects(m.dockerClient)
}

func (m *Model) SelectProject(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedProject < len(m.projects) {
		project := m.projects[m.selectedProject]
		m.projectName = project.Name
		m.currentView = ComposeProcessListView
		m.loading = true
		return m, loadProcesses(m.dockerClient, m.projectName, m.showAll)
	}
	return m, nil
}

// Back navigation handlers
func (m *Model) BackToProcessList(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.currentView = ComposeProcessListView
	return m, loadProcesses(m.dockerClient, m.projectName, m.showAll)
}

func (m *Model) BackFromLogView(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	stopLogReader()
	if m.isDindLog {
		m.currentView = DindComposeProcessListView
		return m, loadDindContainers(m.dockerClient, m.currentDindContainerID)
	}
	m.currentView = ComposeProcessListView
	return m, loadProcesses(m.dockerClient, m.projectName, m.showAll)
}

func (m *Model) BackToDindList(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.currentView = ComposeProcessListView
	return m, loadProcesses(m.dockerClient, m.projectName, m.showAll)
}

// Docker container handlers
func (m *Model) ShowDockerLog(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedDockerContainer < len(m.dockerContainers) {
		container := m.dockerContainers[m.selectedDockerContainer]
		m.containerName = container.Names
		m.isDindLog = false
		m.currentView = LogView
		m.logs = []string{}
		m.logScrollY = 0
		return m, streamLogs(m.dockerClient, container.ID, false, "")
	}
	return m, nil
}

func (m *Model) RefreshDockerList(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.loading = true
	return m, loadDockerContainers(m.dockerClient, m.showAll)
}

func (m *Model) ToggleAllDockerContainers(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.showAll = !m.showAll
	m.loading = true
	return m, loadDockerContainers(m.dockerClient, m.showAll)
}

func (m *Model) KillDockerContainer(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedDockerContainer < len(m.dockerContainers) {
		container := m.dockerContainers[m.selectedDockerContainer]
		m.loading = true
		return m, killService(m.dockerClient, container.ID)
	}
	return m, nil
}

func (m *Model) StopDockerContainer(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedDockerContainer < len(m.dockerContainers) {
		container := m.dockerContainers[m.selectedDockerContainer]
		m.loading = true
		return m, stopService(m.dockerClient, container.ID)
	}
	return m, nil
}

func (m *Model) StartDockerContainer(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedDockerContainer < len(m.dockerContainers) {
		container := m.dockerContainers[m.selectedDockerContainer]
		m.loading = true
		return m, startService(m.dockerClient, container.ID)
	}
	return m, nil
}

func (m *Model) RestartDockerContainer(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedDockerContainer < len(m.dockerContainers) {
		container := m.dockerContainers[m.selectedDockerContainer]
		m.loading = true
		return m, restartService(m.dockerClient, container.ID)
	}
	return m, nil
}

func (m *Model) DeleteDockerContainer(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedDockerContainer < len(m.dockerContainers) {
		container := m.dockerContainers[m.selectedDockerContainer]
		// Only allow removing stopped containers
		if !strings.Contains(container.Status, "Up") && !strings.Contains(container.State, "running") {
			m.loading = true
			return m, removeService(m.dockerClient, container.ID)
		}
	}
	return m, nil
}

func (m *Model) BackFromDockerList(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.currentView = ComposeProcessListView
	return m, loadProcesses(m.dockerClient, m.projectName, m.showAll)
}

// renderHelpText generates help text based on key configurations and screen width
func renderHelpText(configs []KeyConfig, width int) string {
	if len(configs) == 0 {
		return ""
	}

	var helpItems []string
	for _, config := range configs {
		// Use the first key as the display key
		if len(config.Keys) > 0 {
			key := config.Keys[0]
			helpItems = append(helpItems, fmt.Sprintf("%s:%s", key, config.Description))
		}
	}

	// Calculate available width (leaving some margin)
	availableWidth := width - 4
	if availableWidth < 20 {
		// If screen is too narrow, show minimal help
		return "Press q to quit"
	}

	// Join items and wrap if necessary
	helpText := strings.Join(helpItems, " | ")

	if len(helpText) <= availableWidth {
		// All items fit on one line
		return helpText
	}

	// Need to wrap or truncate
	var lines []string
	var currentLine string

	for i, item := range helpItems {
		if i == 0 {
			currentLine = item
		} else {
			testLine := currentLine + " | " + item
			if len(testLine) <= availableWidth {
				currentLine = testLine
			} else {
				// Start new line
				lines = append(lines, currentLine)
				currentLine = item
			}
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	// Return first line (can be extended to show multiple lines)
	if len(lines) > 0 {
		return lines[0]
	}

	return helpText[:availableWidth-3] + "..."
}

// GetHelpText returns the help text for the current view
func (m *Model) GetHelpText() string {
	var configs []KeyConfig

	switch m.currentView {
	case ComposeProcessListView:
		configs = m.processListViewHandlers
	case LogView:
		configs = m.logViewHandlers
	case DindComposeProcessListView:
		configs = m.dindListViewHandlers
	case TopView:
		configs = m.topViewHandlers
	case StatsView:
		configs = m.statsViewHandlers
	case ProjectListView:
		configs = m.projectListViewHandlers
	case DockerContainerListView:
		configs = m.dockerListViewHandlers
	default:
		return ""
	}

	return renderHelpText(configs, m.width)
}

// GetStyledHelpText returns the help text with styling
func (m *Model) GetStyledHelpText() string {
	helpText := m.GetHelpText()
	if helpText == "" {
		return ""
	}

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true).
		Padding(0, 1)

	return style.Render(helpText)
}

// Image list handlers
func (m *Model) SelectUpImage(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedDockerImage > 0 {
		m.selectedDockerImage--
	}
	return m, nil
}

func (m *Model) SelectDownImage(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedDockerImage < len(m.dockerImages)-1 {
		m.selectedDockerImage++
	}
	return m, nil
}

func (m *Model) ShowImageList(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.currentView = ImageListView
	m.loading = true
	return m, loadDockerImages(m.dockerClient, m.showAll)
}

func (m *Model) RefreshImageList(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.loading = true
	return m, loadDockerImages(m.dockerClient, m.showAll)
}

func (m *Model) ToggleAllImages(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.showAll = !m.showAll
	m.loading = true
	return m, loadDockerImages(m.dockerClient, m.showAll)
}

func (m *Model) DeleteImage(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedDockerImage < len(m.dockerImages) {
		image := m.dockerImages[m.selectedDockerImage]
		m.loading = true
		return m, removeImage(m.dockerClient, image.GetRepoTag(), false)
	}
	return m, nil
}

func (m *Model) ForceDeleteImage(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedDockerImage < len(m.dockerImages) {
		image := m.dockerImages[m.selectedDockerImage]
		m.loading = true
		return m, removeImage(m.dockerClient, image.GetRepoTag(), true)
	}
	return m, nil
}

func (m *Model) BackFromImageList(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.currentView = ComposeProcessListView
	return m, loadProcesses(m.dockerClient, m.projectName, m.showAll)
}
