package ui

import (
	"fmt"
	"log/slog"
	"os/exec"
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
		slog.Debug("Loaded composeContainers", slog.Int("count", len(msg.processes)))
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
		m.composeContainers = msg.processes
		m.err = nil
		if len(m.composeContainers) > 0 && m.selectedContainer >= len(m.composeContainers) {
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
		// Reload process list after up/down action
		return m, loadProcesses(m.dockerClient, m.projectName, m.showAll)

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

	case containerFilesLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.containerFiles = msg.files
		m.err = nil
		if len(m.containerFiles) > 0 && m.selectedFile >= len(m.containerFiles) {
			m.selectedFile = 0
		}
		return m, nil

	case fileContentLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.fileContent = msg.content
		m.fileContentPath = msg.path
		m.fileScrollY = 0
		m.err = nil
		m.currentView = FileContentView
		return m, nil

	case executeCommandMsg:
		// Execute the interactive command in a subprocess
		c := exec.Command("docker", append([]string{"exec", "-it", msg.containerID}, msg.command...)...)
		return m, tea.ExecProcess(c, func(err error) tea.Msg {
			// After the command exits, we'll get this message
			return nil
		})

	case inspectLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.inspectContent = msg.content
		m.inspectScrollY = 0
		m.err = nil
		m.currentView = InspectView
		return m, nil

	default:
		return m, nil
	}
}

func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle command mode first
	if m.commandMode {
		return m.handleCommandMode(msg)
	}

	// Handle quit confirmation dialog
	if m.quitConfirmation {
		return m.handleQuitConfirmation(msg)
	}

	// Handle search mode
	if m.searchMode {
		return m.handleSearchMode(msg)
	}

	// Handle ':' to enter command mode
	if msg.String() == ":" {
		m.commandMode = true
		m.commandBuffer = ":"
		m.commandCursorPos = 1
		return m, nil
	}

	// Handle quit globally
	if msg.String() == "q" {
		// For 'q' key, show confirmation dialog
		if m.currentView == ProjectListView || m.currentView == ComposeProcessListView {
			m.quitConfirmation = true
			m.quitConfirmMessage = "Really quit? (y/n)"
			return m, nil
		}
		// For other views, go back
		if m.currentView == LogView {
			stopLogReader()
		}
		m.currentView = ComposeProcessListView
		m.err = nil
		return m, loadProcesses(m.dockerClient, m.projectName, m.showAll)
	}
	
	// Handle ctrl+c for immediate quit
	if msg.String() == "ctrl+c" {
		if m.currentView == LogView {
			stopLogReader()
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
	case FileBrowserView:
		return m.handleFileBrowserKeys(msg)
	case FileContentView:
		return m.handleFileContentKeys(msg)
	case InspectView:
		return m.handleInspectKeys(msg)
	case HelpView:
		return m.handleHelpKeys(msg)
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
		m.searchResults = nil
		m.currentSearchIdx = 0
		m.searchCursorPos = 0
		return m, nil

	case tea.KeyEnter:
		m.searchMode = false
		m.performSearch()
		return m, nil

	case tea.KeyBackspace:
		if m.searchCursorPos > 0 && len(m.searchText) > 0 {
			m.searchText = m.searchText[:m.searchCursorPos-1] + m.searchText[m.searchCursorPos:]
			m.searchCursorPos--
			m.performSearch()
		}
		return m, nil

	case tea.KeyLeft:
		if m.searchCursorPos > 0 {
			m.searchCursorPos--
		}
		return m, nil

	case tea.KeyRight:
		if m.searchCursorPos < len(m.searchText) {
			m.searchCursorPos++
		}
		return m, nil

	case tea.KeyCtrlI:
		m.searchIgnoreCase = !m.searchIgnoreCase
		m.performSearch()
		return m, nil

	case tea.KeyCtrlR:
		m.searchRegex = !m.searchRegex
		m.performSearch()
		return m, nil

	default:
		if msg.Type == tea.KeyRunes {
			// Insert at cursor position
			m.searchText = m.searchText[:m.searchCursorPos] + msg.String() + m.searchText[m.searchCursorPos:]
			m.searchCursorPos += len(msg.String())
			m.performSearch()
		}
		return m, nil
	}
}

func (m *Model) handleCommandMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		// Exit command mode
		m.commandMode = false
		m.commandBuffer = ""
		m.commandCursorPos = 0
		return m, nil

	case tea.KeyEnter:
		// Execute command
		return m.executeCommand()

	case tea.KeyBackspace:
		if len(m.commandBuffer) > 1 && m.commandCursorPos > 1 {
			m.commandBuffer = m.commandBuffer[:m.commandCursorPos-1] + m.commandBuffer[m.commandCursorPos:]
			m.commandCursorPos--
		}
		return m, nil

	case tea.KeyLeft:
		if m.commandCursorPos > 1 {
			m.commandCursorPos--
		}
		return m, nil

	case tea.KeyRight:
		if m.commandCursorPos < len(m.commandBuffer) {
			m.commandCursorPos++
		}
		return m, nil

	case tea.KeyUp:
		// Navigate command history
		if m.commandHistoryIdx > 0 {
			m.commandHistoryIdx--
			if m.commandHistoryIdx < len(m.commandHistory) {
				m.commandBuffer = ":" + m.commandHistory[m.commandHistoryIdx]
				m.commandCursorPos = len(m.commandBuffer)
			}
		}
		return m, nil

	case tea.KeyDown:
		// Navigate command history
		if m.commandHistoryIdx < len(m.commandHistory)-1 {
			m.commandHistoryIdx++
			m.commandBuffer = ":" + m.commandHistory[m.commandHistoryIdx]
			m.commandCursorPos = len(m.commandBuffer)
		} else if m.commandHistoryIdx == len(m.commandHistory)-1 {
			m.commandHistoryIdx++
			m.commandBuffer = ":"
			m.commandCursorPos = 1
		}
		return m, nil

	default:
		if msg.Type == tea.KeyRunes {
			// Insert character at cursor position
			m.commandBuffer = m.commandBuffer[:m.commandCursorPos] + msg.String() + m.commandBuffer[m.commandCursorPos:]
			m.commandCursorPos += len(msg.String())
		}
		return m, nil
	}
}

func (m *Model) handleQuitConfirmation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		// Confirm quit
		m.quitConfirmation = false
		if m.currentView == LogView {
			stopLogReader()
		}
		return m, tea.Quit
	case "n", "N", "esc":
		// Cancel quit
		m.quitConfirmation = false
		m.quitConfirmMessage = ""
		return m, nil
	}
	return m, nil
}

func (m *Model) executeCommand() (tea.Model, tea.Cmd) {
	command := strings.TrimSpace(m.commandBuffer[1:]) // Remove leading ':'
	
	// Add to command history
	if command != "" && (len(m.commandHistory) == 0 || m.commandHistory[len(m.commandHistory)-1] != command) {
		m.commandHistory = append(m.commandHistory, command)
	}
	m.commandHistoryIdx = len(m.commandHistory)
	
	// Exit command mode
	m.commandMode = false
	m.commandBuffer = ""
	m.commandCursorPos = 0
	
	// Parse and execute command
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return m, nil
	}
	
	switch parts[0] {
	case "q", "quit":
		// Show quit confirmation
		m.quitConfirmation = true
		m.quitConfirmMessage = "Really quit? (y/n)"
		return m, nil
		
	case "q!", "quit!":
		// Force quit without confirmation
		if m.currentView == LogView {
			stopLogReader()
		}
		return m, tea.Quit
		
	case "h", "help":
		// Show help
		m.previousView = m.currentView
		m.currentView = HelpView
		m.helpScrollY = 0
		return m, nil
		
	case "set":
		// Handle set commands (e.g., :set showAll)
		if len(parts) > 1 {
			switch parts[1] {
			case "all", "showAll":
				m.showAll = true
				return m, m.refreshCurrentView()
			case "noall", "noshowAll":
				m.showAll = false
				return m, m.refreshCurrentView()
			}
		}
		return m, nil
		
	default:
		m.err = fmt.Errorf("unknown command: %s", command)
		return m, nil
	}
}


func (m *Model) refreshCurrentView() tea.Cmd {
	switch m.currentView {
	case ComposeProcessListView:
		return loadProcesses(m.dockerClient, m.projectName, m.showAll)
	case DockerContainerListView:
		return loadDockerContainers(m.dockerClient, m.showAll)
	case ImageListView:
		return loadDockerImages(m.dockerClient, m.showAll)
	case NetworkListView:
		return loadDockerNetworks(m.dockerClient)
	case ProjectListView:
		return loadProjects(m.dockerClient)
	case DindComposeProcessListView:
		if m.currentDindContainerID != "" {
			return loadDindContainers(m.dockerClient, m.currentDindContainerID)
		}
	case FileBrowserView:
		if m.browsingContainerID != "" {
			return loadContainerFiles(m.dockerClient, m.browsingContainerID, m.currentPath)
		}
	}
	return nil
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

func (m *Model) handleFileBrowserKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	handler, ok := m.fileBrowserKeymap[msg.String()]
	if ok {
		return handler(msg)
	}
	return m, nil
}

func (m *Model) handleFileContentKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	handler, ok := m.fileContentKeymap[msg.String()]
	if ok {
		return handler(msg)
	}
	return m, nil
}

func (m *Model) handleInspectKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle search mode
	if m.searchMode {
		return m.handleInspectSearchMode(msg)
	}
	
	handler, ok := m.inspectViewKeymap[msg.String()]
	if ok {
		return handler(msg)
	}
	return m, nil
}

func (m *Model) handleInspectSearchMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC, tea.KeyEsc:
		m.searchMode = false
		m.searchText = ""
		m.searchResults = nil
		m.currentSearchIdx = 0
		return m, nil
		
	case tea.KeyEnter:
		m.searchMode = false
		m.performInspectSearch()
		return m, nil
		
	case tea.KeyBackspace:
		if m.searchCursorPos > 0 {
			m.searchText = m.searchText[:m.searchCursorPos-1] + m.searchText[m.searchCursorPos:]
			m.searchCursorPos--
		}
		return m, nil
		
	case tea.KeyDelete:
		if m.searchCursorPos < len(m.searchText) {
			m.searchText = m.searchText[:m.searchCursorPos] + m.searchText[m.searchCursorPos+1:]
		}
		return m, nil
		
	case tea.KeyLeft:
		if m.searchCursorPos > 0 {
			m.searchCursorPos--
		}
		return m, nil
		
	case tea.KeyRight:
		if m.searchCursorPos < len(m.searchText) {
			m.searchCursorPos++
		}
		return m, nil
		
	case tea.KeyHome:
		m.searchCursorPos = 0
		return m, nil
		
	case tea.KeyEnd:
		m.searchCursorPos = len(m.searchText)
		return m, nil
		
	case tea.KeyCtrlI: // Tab - toggle case insensitive
		m.searchIgnoreCase = !m.searchIgnoreCase
		return m, nil
		
	case tea.KeyCtrlR: // Toggle regex
		m.searchRegex = !m.searchRegex
		return m, nil
		
	case tea.KeyRunes:
		// Insert characters at cursor position
		runes := string(msg.Runes)
		m.searchText = m.searchText[:m.searchCursorPos] + runes + m.searchText[m.searchCursorPos:]
		m.searchCursorPos += len(runes)
		return m, nil
	}
	
	return m, nil
}

func (m *Model) handleHelpKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	handler, ok := m.helpViewKeymap[msg.String()]
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
