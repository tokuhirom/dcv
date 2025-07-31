package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case processesLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.processes = msg.processes
		m.err = nil
		if len(m.processes) > 0 && m.selectedProcess >= len(m.processes) {
			m.selectedProcess = 0
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
		// Auto-scroll to bottom
		maxScroll := len(m.logs) - (m.height - 4)
		if maxScroll > 0 {
			m.logScrollY = maxScroll
		}
		// Continue polling for more logs
		return m, pollForLogs()

	case errorMsg:
		m.err = msg.err
		m.loading = false
		return m, nil

	default:
		return m, nil
	}
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle search mode first
	if m.searchMode {
		return m.handleSearchMode(msg)
	}

	// Handle quit globally
	if msg.String() == "q" || msg.String() == "ctrl+c" {
		if m.currentView != ProcessListView {
			// Go back to process list
			m.currentView = ProcessListView
			m.err = nil
			return m, loadProcesses(m.dockerClient)
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
	default:
		return m, nil
	}
}

func (m Model) handleProcessListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selectedProcess > 0 {
			m.selectedProcess--
		}
		return m, nil

	case "down", "j":
		if m.selectedProcess < len(m.processes)-1 {
			m.selectedProcess++
		}
		return m, nil

	case "enter":
		if m.selectedProcess < len(m.processes) {
			process := m.processes[m.selectedProcess]
			m.containerName = process.Name
			m.isDindLog = false
			m.currentView = LogView
			m.logs = []string{}
			m.logScrollY = 0
			// Use service name for docker compose logs
			return m, streamLogs(m.dockerClient, process.Service, false, "")
		}
		return m, nil

	case "d":
		if m.selectedProcess < len(m.processes) {
			process := m.processes[m.selectedProcess]
			if process.IsDind {
				m.currentDindHost = process.Name
				m.currentView = DindProcessListView
				m.loading = true
				return m, loadDindContainers(m.dockerClient, process.Name)
			}
		}
		return m, nil

	case "r":
		m.loading = true
		return m, loadProcesses(m.dockerClient)

	default:
		return m, nil
	}
}

func (m Model) handleLogViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		if m.isDindLog {
			m.currentView = DindProcessListView
			return m, loadDindContainers(m.dockerClient, m.currentDindHost)
		}
		m.currentView = ProcessListView
		return m, loadProcesses(m.dockerClient)

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

func (m Model) handleDindListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
			return m, streamLogs(m.dockerClient, container.Name, true, m.currentDindHost)
		}
		return m, nil

	case "esc":
		m.currentView = ProcessListView
		return m, loadProcesses(m.dockerClient)

	case "r":
		m.loading = true
		return m, loadDindContainers(m.dockerClient, m.currentDindHost)

	default:
		return m, nil
	}
}

func (m Model) handleSearchMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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

// Helper function to check if a process is dind
func isDind(imageName string) bool {
	lower := strings.ToLower(imageName)
	return strings.Contains(lower, "dind") || strings.Contains(lower, "docker:dind")
}