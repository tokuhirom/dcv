package ui

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Update handles messages and updates the model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case processesLoadedMsg:
		m.loading = false
		// Example debug logging
		slog.Debug("Loaded containers", slog.Int("count", len(msg.processes)))
		if msg.err != nil {
			// Check if error is due to missing compose file
			if containsAny(msg.err.Error(), []string{"no configuration file provided", "not found", "no such file"}) {
				// Switch to project list view
				m.currentView = ProjectListView
				m.loading = true
				return m, loadProjects(m.dockerClient)
			}
			m.err = msg.err
			return m, nil
		}
		m.containers = msg.processes
		m.err = nil
		if len(m.containers) > 0 && m.selectedContainer >= len(m.containers) {
			m.selectedContainer = 0
		}
		return m, nil

	case dindContainersLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.dindContainers = msg.containers
		m.err = nil
		if len(m.dindContainers) > 0 && m.selectedDindContainer >= len(m.dindContainers) {
			m.selectedDindContainer = 0
		}
		return m, nil

	case logLineMsg:
		m.logs = append(m.logs, msg.line)
		// Keep only last 10000 lines to prevent unbounded memory growth
		if len(m.logs) > 10000 {
			m.logs = m.logs[len(m.logs)-10000:]
		}
		// Auto-scroll to bottom
		maxScroll := len(m.logs) - (m.height - 4)
		if maxScroll > 0 {
			m.logScrollY = maxScroll
		}
		// Don't continue polling for single lines (e.g., "[Log reader stopped]")
		return m, nil

	case logLinesMsg:
		m.logs = append(m.logs, msg.lines...)
		// Keep only last 10000 lines to prevent unbounded memory growth
		if len(m.logs) > 10000 {
			m.logs = m.logs[len(m.logs)-10000:]
		}
		// Auto-scroll to bottom
		maxScroll := len(m.logs) - (m.height - 4)
		if maxScroll > 0 {
			m.logScrollY = maxScroll
		}
		// Continue polling for more logs with a small delay
		return m, tea.Tick(time.Millisecond*50, func(time.Time) tea.Msg {
			return pollForLogs()()
		})

	case pollLogsContinueMsg:
		// Continue polling with a delay
		return m, tea.Tick(time.Millisecond*50, func(time.Time) tea.Msg {
			return pollForLogs()()
		})

	case errorMsg:
		m.err = msg.err
		m.loading = false
		return m, nil

	case commandExecutedMsg:
		// Start polling for logs after command is set
		return m, pollForLogs()

	case topLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.topOutput = msg.output
		m.err = nil
		return m, nil

	case serviceActionCompleteMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		// Reload the process list after action completes
		return m, loadProcesses(m.dockerClient, m.projectName, m.showAll)

	case upActionCompleteMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		return m, nil

	case statsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.stats = msg.stats
		m.err = nil
		return m, nil

	case projectsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.projects = msg.projects
		m.err = nil
		if len(m.projects) > 0 && m.selectedProject >= len(m.projects) {
			m.selectedProject = 0
		}
		return m, nil

	default:
		return m, nil
	}
}

func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle search mode first
	if m.searchMode {
		return m.handleSearchMode(msg)
	}

	// Handle quit globally
	if msg.String() == "q" || msg.String() == "ctrl+c" {
		// Stop log reader if in log view
		if m.currentView == LogView {
			stopLogReader()
		}
		if m.currentView == ProjectListView {
			return m, tea.Quit
		}
		if m.currentView != ProcessListView {
			// Go back to process list
			m.currentView = ProcessListView
			m.err = nil
			return m, loadProcesses(m.dockerClient, m.projectName, m.showAll)
		}
		return m, tea.Quit
	}

	// Handle view-specific keys
	switch m.currentView {
	case ProcessListView:
		return m.handleProcessListKeys(msg)
	case LogView:
		return m.handleLogViewKeys(msg)
	case DindProcessListView:
		return m.handleDindListKeys(msg)
	case TopView:
		return m.handleTopViewKeys(msg)
	case StatsView:
		return m.handleStatsViewKeys(msg)
	case ProjectListView:
		return m.handleProjectListKeys(msg)
	default:
		return m, nil
	}
}

func (m *Model) handleProcessListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	handler, ok := m.processListViewKeymap[msg.String()]
	slog.Info(fmt.Sprintf("Key: %s", msg.String()),
		slog.Bool("ok", ok),
		slog.Any("handler", m.processListViewKeymap))
	if ok {
		return handler(msg)
	}
	return m, nil
}

func (m *Model) handleLogViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Stop the log reader before switching views
		stopLogReader()
		if m.isDindLog {
			m.currentView = DindProcessListView
			return m, loadDindContainers(m.dockerClient, m.currentDindContainerID)
		}
		m.currentView = ProcessListView
		return m, loadProcesses(m.dockerClient, m.projectName, m.showAll)

	case "up", "k":
		if m.logScrollY > 0 {
			m.logScrollY--
		}
		return m, nil

	case "down", "j":
		maxScroll := len(m.logs) - (m.height - 4)
		if m.logScrollY < maxScroll && maxScroll > 0 {
			m.logScrollY++
		}
		return m, nil

	case "G":
		maxScroll := len(m.logs) - (m.height - 4)
		if maxScroll > 0 {
			m.logScrollY = maxScroll
		}
		return m, nil

	case "g":
		m.logScrollY = 0
		return m, nil

	case "/":
		m.searchMode = true
		m.searchText = ""
		return m, nil

	default:
		return m, nil
	}
}

func (m *Model) handleDindListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selectedDindContainer > 0 {
			m.selectedDindContainer--
		}
		return m, nil

	case "down", "j":
		if m.selectedDindContainer < len(m.dindContainers)-1 {
			m.selectedDindContainer++
		}
		return m, nil

	case "enter":
		if m.selectedDindContainer < len(m.dindContainers) {
			container := m.dindContainers[m.selectedDindContainer]
			m.containerName = container.Name
			m.hostContainer = m.currentDindHost
			m.isDindLog = true
			m.currentView = LogView
			m.logs = []string{}
			m.logScrollY = 0
			return m, streamLogs(m.dockerClient, container.Name, true, m.currentDindContainerID)
		}
		return m, nil

	case "esc":
		m.currentView = ProcessListView
		return m, loadProcesses(m.dockerClient, m.projectName, m.showAll)

	case "r":
		m.loading = true
		return m, loadDindContainers(m.dockerClient, m.currentDindContainerID)

	default:
		return m, nil
	}
}

func (m *Model) handleSearchMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.searchMode = false
		m.searchText = ""
		return m, nil

	case tea.KeyEnter:
		m.searchMode = false
		// Implement search functionality
		return m, nil

	case tea.KeyBackspace:
		if len(m.searchText) > 0 {
			m.searchText = m.searchText[:len(m.searchText)-1]
		}
		return m, nil

	default:
		if msg.Type == tea.KeyRunes {
			m.searchText += msg.String()
		}
		return m, nil
	}
}

func (m *Model) handleTopViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		// Go back to process list
		m.currentView = ProcessListView
		return m, loadProcesses(m.dockerClient, m.projectName, m.showAll)

	case "r":
		// Manual refresh
		m.loading = true
		return m, loadTop(m.dockerClient, m.projectName, m.topService)

	default:
		return m, nil
	}
}

func (m *Model) handleStatsViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		// Go back to process list
		m.currentView = ProcessListView
		return m, loadProcesses(m.dockerClient, m.projectName, m.showAll)

	case "r":
		// Refresh stats
		m.loading = true
		return m, loadStats(m.dockerClient)

	default:
		return m, nil
	}
}

func (m *Model) handleProjectListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selectedProject > 0 {
			m.selectedProject--
		}
		return m, nil

	case "down", "j":
		if m.selectedProject < len(m.projects)-1 {
			m.selectedProject++
		}
		return m, nil

	case "enter":
		if m.selectedProject < len(m.projects) {
			project := m.projects[m.selectedProject]
			// Create a new compose client with the selected project
			m.projectName = project.Name
			m.currentView = ProcessListView
			m.loading = true
			return m, loadProcesses(m.dockerClient, m.projectName, m.showAll)
		}
		return m, nil

	case "r":
		m.loading = true
		return m, loadProjects(m.dockerClient)

	default:
		return m, nil
	}
}

// containsAny checks if the string contains any of the substrings
func containsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}
