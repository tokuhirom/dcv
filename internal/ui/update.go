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
		// Reload the appropriate view after action completes
		switch m.currentView {
		case ImageListView:
			return m, loadDockerImages(m.dockerClient, m.showAll)
		case DockerContainerListView:
			return m, loadDockerContainers(m.dockerClient, m.showAll)
		case NetworkListView:
			return m, loadDockerNetworks(m.dockerClient)
		default:
			return m, loadProcesses(m.dockerClient, m.projectName, m.showAll)
		}

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

	case dockerContainersLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.dockerContainers = msg.containers
		m.err = nil
		if len(m.dockerContainers) > 0 && m.selectedDockerContainer >= len(m.dockerContainers) {
			m.selectedDockerContainer = 0
		}
		return m, nil

	case dockerImagesLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.dockerImages = msg.images
		m.err = nil
		if len(m.dockerImages) > 0 && m.selectedDockerImage >= len(m.dockerImages) {
			m.selectedDockerImage = 0
		}
		return m, nil

	case dockerNetworksLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.dockerNetworks = msg.networks
		m.err = nil
		if len(m.dockerNetworks) > 0 && m.selectedDockerNetwork >= len(m.dockerNetworks) {
			m.selectedDockerNetwork = 0
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
		if m.currentView != ComposeProcessListView {
			// Go back to process list
			m.currentView = ComposeProcessListView
			m.err = nil
			return m, loadProcesses(m.dockerClient, m.projectName, m.showAll)
		}
		return m, tea.Quit
	}

	// Handle view-specific keys
	switch m.currentView {
	case ComposeProcessListView:
		return m.handleProcessListKeys(msg)
	case LogView:
		return m.handleLogViewKeys(msg)
	case DindComposeProcessListView:
		return m.handleDindListKeys(msg)
	case TopView:
		return m.handleTopViewKeys(msg)
	case StatsView:
		return m.handleStatsViewKeys(msg)
	case ProjectListView:
		return m.handleProjectListKeys(msg)
	case DockerContainerListView:
		return m.handleDockerListKeys(msg)
	case ImageListView:
		return m.handleImageListKeys(msg)
	case NetworkListView:
		return m.handleNetworkListKeys(msg)
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
	handler, ok := m.logViewKeymap[msg.String()]
	if ok {
		return handler(msg)
	}
	return m, nil
}

func (m *Model) handleDindListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	handler, ok := m.dindListViewKeymap[msg.String()]
	if ok {
		return handler(msg)
	}
	return m, nil
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
	handler, ok := m.topViewKeymap[msg.String()]
	if ok {
		return handler(msg)
	}
	return m, nil
}

func (m *Model) handleStatsViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	handler, ok := m.statsViewKeymap[msg.String()]
	if ok {
		return handler(msg)
	}
	return m, nil
}

func (m *Model) handleProjectListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	handler, ok := m.projectListViewKeymap[msg.String()]
	if ok {
		return handler(msg)
	}
	return m, nil
}

func (m *Model) handleDockerListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	handler, ok := m.dockerListViewKeymap[msg.String()]
	if ok {
		return handler(msg)
	}
	return m, nil
}

func (m *Model) handleImageListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	handler, ok := m.imageListViewKeymap[msg.String()]
	if ok {
		return handler(msg)
	}
	return m, nil
}

func (m *Model) handleNetworkListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	handler, ok := m.networkListViewKeymap[msg.String()]
	if ok {
		return handler(msg)
	}
	return m, nil
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
