package ui

import (
	"fmt"
	"log/slog"
	"path/filepath"
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

// Refresh sends a RefreshMsg to trigger a reload of the current view
func (m *Model) Refresh(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, func() tea.Msg {
		return RefreshMsg{}
	}
}

// Common selection handlers for different views
func (m *Model) SelectUpDindContainer(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.dindProcessListViewModel.HandleSelectUp()
}

func (m *Model) SelectDownDindContainer(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.dindProcessListViewModel.HandleSelectDown()
}

func (m *Model) CmdUp(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case DockerContainerListView:
		return m, m.dockerContainerListViewModel.HandleUp(m)
	case ComposeProjectListView:
		return m, m.composeProjectListViewModel.HandleUp(m)
	case ComposeProcessListView:
		return m, m.composeProcessListViewModel.HandleUp()
	case CommandExecutionView:
		return m, m.commandExecutionViewModel.HandleUp()
	default:
		slog.Info("Unhandled key up in current view",
			slog.String("view", m.currentView.String()))
		return m, nil
	}
}

func (m *Model) CmdDown(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	slog.Info("CmdDown called",
		slog.String("view", m.currentView.String()),
		slog.Int("selectedContainer", m.composeProcessListViewModel.selectedContainer))

	switch m.currentView {
	case DockerContainerListView:
		return m, m.dockerContainerListViewModel.HandleDown(m)
	case ComposeProjectListView:
		return m, m.composeProjectListViewModel.HandleDown(m)
	case ComposeProcessListView:
		return m, m.composeProcessListViewModel.HandleDown()
	case CommandExecutionView:
		return m, m.commandExecutionViewModel.HandleDown(m)
	default:
		slog.Info("Unhandled key down in current view",
			slog.String("view", m.currentView.String()))
		return m, nil
	}
}

func (m *Model) CmdGoToEnd(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case CommandExecutionView:
		return m, m.commandExecutionViewModel.HandleGoToEnd(m)
	default:
		slog.Info("Unhandled go to end in current view",
			slog.String("view", m.currentView.String()))
		return m, nil
	}
}

func (m *Model) CmdGoToStart(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case CommandExecutionView:
		return m, m.commandExecutionViewModel.HandleGoToStart()
	default:
		slog.Info("Unhandled go to start in current view",
			slog.String("view", m.currentView.String()))
		return m, nil
	}
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
	m.searchCursorPos = 0
	m.searchResults = nil
	m.currentSearchIdx = 0
	return m, nil
}

func (m *Model) StartFilter(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.filterMode = true
	m.filterText = ""
	m.filterCursorPos = 0
	m.filteredLogs = nil
	return m, nil
}

func (m *Model) NextSearchResult(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if len(m.searchResults) > 0 {
		m.currentSearchIdx = (m.currentSearchIdx + 1) % len(m.searchResults)
		// Jump to the line
		if m.currentSearchIdx < len(m.searchResults) {
			targetLine := m.searchResults[m.currentSearchIdx]
			m.logScrollY = targetLine - m.height/2 + 3 // Center the result
			if m.logScrollY < 0 {
				m.logScrollY = 0
			}
		}
	}
	return m, nil
}

func (m *Model) PrevSearchResult(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if len(m.searchResults) > 0 {
		m.currentSearchIdx--
		if m.currentSearchIdx < 0 {
			m.currentSearchIdx = len(m.searchResults) - 1
		}
		// Jump to the line
		if m.currentSearchIdx < len(m.searchResults) {
			targetLine := m.searchResults[m.currentSearchIdx]
			m.logScrollY = targetLine - m.height/2 + 3 // Center the result
			if m.logScrollY < 0 {
				m.logScrollY = 0
			}
		}
	}
	return m, nil
}

func (m *Model) ShowDindLog(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.dindProcessListViewModel.HandleShowLog(m)
}

func (m *Model) ShowDindProcessList(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case ComposeProcessListView:
		return m, m.composeProcessListViewModel.HandleDindProcessList(m)
	default:
		slog.Info("Unhandled Dind process list command in current view",
			slog.String("view", m.currentView.String()))
	}
	return m, nil
}

func (m *Model) ShowDockerContainerList(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.dockerContainerListViewModel.Show(m)
}

func (m *Model) CmdToggleAll(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case ComposeProcessListView:
		return m, m.composeProcessListViewModel.HandleToggleAll(m)
	default:
		slog.Info("Unhandled toggle all command in current view",
			slog.String("view", m.currentView.String()))
	}
	return m, nil
}

func (m *Model) ShowStatsView(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.statsViewModel.Show(m)
}

func (m *Model) CmdTop(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case ComposeProcessListView:
		return m, m.composeProcessListViewModel.HandleTop(m)
	default:
		slog.Info("Unhandled top command in current view",
			slog.String("view", m.currentView.String()))
	}
	return m, nil
}

func (m *Model) DeleteContainer(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m *Model) DeployProject(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.commandExecutionViewModel.ExecuteComposeCommand(m, "up")
}

func (m *Model) DownProject(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.commandExecutionViewModel.ExecuteComposeCommand(m, "down")
}

func (m *Model) ShowProjectList(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.currentView = ComposeProjectListView
	m.loading = true
	return m, loadProjects(m.dockerClient)
}

func (m *Model) SelectProject(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case ComposeProjectListView:
		return m, m.composeProjectListViewModel.HandleSelectProject(m)
	default:
		slog.Info("Unhandled project selection in current view",
			slog.String("view", m.currentView.String()))
		return m, nil
	}
}

func (m *Model) BackFromLogView(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	stopLogReader()
	if m.isDindLog {
		m.currentView = DindProcessListView
		return m, loadDindContainers(m.dockerClient, m.dindProcessListViewModel.currentDindContainerID)
	}
	m.currentView = ComposeProcessListView
	return m, loadProcesses(m.dockerClient, m.projectName, m.composeProcessListViewModel.showAll)
}

// Docker container handlers
func (m *Model) CmdLog(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case DockerContainerListView:
		return m, m.dockerContainerListViewModel.HandleLog(m)
	case ComposeProcessListView:
		return m, m.composeProcessListViewModel.HandleLog(m)
	default:
		slog.Info("Unhandled :log command in current view",
			slog.String("view", m.currentView.String()))
	}
	return m, nil
}

func (m *Model) ToggleAllDockerContainers(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.dockerContainerListViewModel.showAll = !m.dockerContainerListViewModel.showAll
	m.loading = true
	return m, loadDockerContainers(m.dockerClient, m.dockerContainerListViewModel.showAll)
}

func (m *Model) CmdKill(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case DockerContainerListView:
		return m, m.dockerContainerListViewModel.HandleKill(m)
	case ComposeProcessListView:
		return m, m.composeProcessListViewModel.HandleKill(m)
	default:
		slog.Info("Unhandled :kill command in current view",
			slog.String("view", m.currentView.String()))
	}
	return m, nil
}

func (m *Model) CmdStop(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case DockerContainerListView:
		return m, m.dockerContainerListViewModel.HandleStop(m)
	case ComposeProcessListView:
		return m, m.composeProcessListViewModel.HandleStop(m)
	default:
		slog.Info("Unhandled :stop command in current view",
			slog.String("view", m.currentView.String()))
	}
	return m, nil
}

func (m *Model) CmdStart(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case DockerContainerListView:
		return m, m.dockerContainerListViewModel.HandleStart(m)
	case ComposeProcessListView:
		return m, m.composeProcessListViewModel.HandleStart(m)
	default:
		slog.Info("Unhandled :start command in current view",
			slog.String("view", m.currentView.String()))
	}
	return m, nil
}

func (m *Model) CmdRestart(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case DockerContainerListView:
		return m, m.dockerContainerListViewModel.HandleRestart(m)
	case ComposeProcessListView:
		return m, m.composeProcessListViewModel.HandleRestart(m)
	default:
		slog.Info("Unhandled :restart command in current view",
			slog.String("view", m.currentView.String()))
	}
	return m, nil
}

func (m *Model) CmdRemove(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case DockerContainerListView:
		return m, m.dockerContainerListViewModel.HandleRemove(m)
	case ComposeProcessListView:
		return m, m.composeProcessListViewModel.HandleRemove(m)
	default:
		slog.Info("Unhandled :delete command in current view",
			slog.String("view", m.currentView.String()))
	}
	return m, nil
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
				// HandleStart new line
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
	case DindProcessListView:
		configs = m.dindListViewHandlers
	case TopView:
		configs = m.topViewHandlers
	case StatsView:
		configs = m.statsViewHandlers
	case ComposeProjectListView:
		configs = m.projectListViewHandlers
	case DockerContainerListView:
		configs = m.dockerContainerListViewHandlers
	case ImageListView:
		configs = m.imageListViewHandlers
	case NetworkListView:
		configs = m.networkListViewHandlers
	case FileBrowserView:
		configs = m.fileBrowserHandlers
	case FileContentView:
		configs = m.fileContentHandlers
	case InspectView:
		configs = m.inspectViewHandlers
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
	return m, m.imageListViewModel.HandleSelectUp()
}

func (m *Model) SelectDownImage(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.imageListViewModel.HandleSelectDown()
}

func (m *Model) ShowImageList(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.imageListViewModel.Show(m)
}

func (m *Model) ToggleAllImages(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.imageListViewModel.HandleToggleAll(m)
}

func (m *Model) DeleteImage(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.imageListViewModel.HandleDelete(m)
}

func (m *Model) ForceDeleteImage(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.imageListViewModel.HandleForceDelete(m)
}

func (m *Model) ShowImageInspect(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.imageListViewModel.HandleInspect(m)
}

// Network list handlers
func (m *Model) SelectUpNetwork(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.networkListViewModel.HandleSelectUp()
}

func (m *Model) SelectDownNetwork(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.networkListViewModel.HandleSelectDown()
}

func (m *Model) ShowNetworkList(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.networkListViewModel.Show(m)
}

func (m *Model) DeleteNetwork(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.networkListViewModel.HandleDelete(m)
}

func (m *Model) ShowNetworkInspect(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.networkListViewModel.HandleInspect(m)
}

func (m *Model) CmdFileBrowse(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case DockerContainerListView:
		return m, m.dockerContainerListViewModel.HandleFileBrowse(m)
	case ComposeProcessListView:
		return m, m.composeProcessListViewModel.HandleFileBrowse(m)
	default:
		slog.Info("Unhandled :filebrowser command in current view",
			slog.String("view", m.currentView.String()))
	}
	return m, nil
}

func (m *Model) SelectUpFile(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedFile > 0 {
		m.selectedFile--
	}
	return m, nil
}

func (m *Model) SelectDownFile(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedFile < len(m.containerFiles)-1 {
		m.selectedFile++
	}
	return m, nil
}

func (m *Model) OpenFileOrDirectory(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedFile < len(m.containerFiles) {
		file := m.containerFiles[m.selectedFile]

		if file.Name == "." {
			return m, nil
		}

		if file.Name == ".." {
			// Go up one directory
			if m.currentPath != "/" {
				m.currentPath = filepath.Dir(m.currentPath)
				if len(m.pathHistory) > 1 {
					m.pathHistory = m.pathHistory[:len(m.pathHistory)-1]
				}
			}
			m.loading = true
			m.selectedFile = 0
			return m, loadContainerFiles(m.dockerClient, m.browsingContainerID, m.currentPath)
		}

		newPath := filepath.Join(m.currentPath, file.Name)

		if file.IsDir {
			// Navigate into directory
			m.currentPath = newPath
			m.pathHistory = append(m.pathHistory, newPath)
			m.loading = true
			m.selectedFile = 0
			return m, loadContainerFiles(m.dockerClient, m.browsingContainerID, newPath)
		} else {
			// View file content
			return m, m.fileContentViewModel.Load(m, m.browsingContainerID, newPath)
		}
	}
	return m, nil
}

func (m *Model) GoToParentDirectory(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Go up one directory
	if m.currentPath != "/" {
		m.currentPath = filepath.Dir(m.currentPath)
		if len(m.pathHistory) > 1 {
			m.pathHistory = m.pathHistory[:len(m.pathHistory)-1]
		}
		m.loading = true
		m.selectedFile = 0
		return m, loadContainerFiles(m.dockerClient, m.browsingContainerID, m.currentPath)
	}
	return m, nil
}

func (m *Model) CmdBack(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case FileContentView:
		return m, m.fileContentViewModel.HandleBack(m)
	case HelpView:
		return m, m.helpViewModel.HandleBack(m)
	case InspectView:
		return m, m.inspectViewModel.HandleBack(m)
	case VolumeListView:
		return m, m.volumeListViewModel.HandleBack(m)
	case NetworkListView:
		return m, m.networkListViewModel.HandleBack(m)
	case ImageListView:
		return m, m.imageListViewModel.HandleBack(m)
	case DockerContainerListView:
		return m, m.dockerContainerListViewModel.HandleBack(m)
	case FileBrowserView:
		return m, m.fileBrowserViewModel.HandleBack(m)
	case CommandExecutionView:
		return m, m.commandExecutionViewModel.HandleBack(m)
	case ComposeProcessListView:
		m.currentView = ComposeProjectListView
		return m, nil
	case DindProcessListView:
		return m, m.dindProcessListViewModel.HandleBack(m)
	case TopView:
		return m, m.topViewModel.HandleBack(m)
	case StatsView:
		return m, m.statsViewModel.HandleBack(m)
	default:
		slog.Info("Unhandled :back command in current view",
			slog.String("view", m.currentView.String()))
	}
	return m, nil
}

// File content handlers
func (m *Model) ScrollFileUp(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.fileContentViewModel.HandleScrollUp()
}

func (m *Model) ScrollFileDown(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.fileContentViewModel.HandleScrollDown(m.height)
}

func (m *Model) GoToFileEnd(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.fileContentViewModel.HandleGoToEnd(m.height)
}

func (m *Model) GoToFileStart(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.fileContentViewModel.HandleGoToStart()
}

func (m *Model) CmdShell(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case DockerContainerListView:
		return m, m.dockerContainerListViewModel.HandleShell(m)
	case ComposeProcessListView:
		return m, m.composeProcessListViewModel.HandleShell(m)
	default:
		slog.Info("Unhandled :shell command in current view",
			slog.String("view", m.currentView.String()))
	}
	return m, nil
}

// Inspect handlers
func (m *Model) CmdInspect(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case ComposeProcessListView:
		return m, m.composeProcessListViewModel.HandleInspect(m)
	default:
		slog.Info("Unhandled :inspect command in current view",
			slog.String("view", m.currentView.String()))
		return m, nil
	}
}

func (m *Model) ShowDockerInspect(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case DockerContainerListView:
		return m, m.dockerContainerListViewModel.HandleInspect(m)
	default:
		slog.Info("Unhandled :inspect command in current view",
			slog.String("view", m.currentView.String()))
	}
	return m, nil
}

func (m *Model) ScrollInspectUp(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.inspectScrollY > 0 {
		m.inspectScrollY--
	}
	return m, nil
}

func (m *Model) ScrollInspectDown(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	lines := strings.Split(m.inspectContent, "\n")
	maxScroll := len(lines) - (m.height - 5)
	if m.inspectScrollY < maxScroll && maxScroll > 0 {
		m.inspectScrollY++
	}
	return m, nil
}

func (m *Model) GoToInspectEnd(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	lines := strings.Split(m.inspectContent, "\n")
	maxScroll := len(lines) - (m.height - 5)
	if maxScroll > 0 {
		m.inspectScrollY = maxScroll
	}
	return m, nil
}

func (m *Model) GoToInspectStart(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.inspectScrollY = 0
	return m, nil
}

func (m *Model) StartInspectSearch(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.searchMode = true
	m.searchText = ""
	m.searchCursorPos = 0
	m.searchResults = nil
	m.currentSearchIdx = 0
	return m, nil
}

func (m *Model) NextInspectSearchResult(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if len(m.searchResults) > 0 {
		m.currentSearchIdx = (m.currentSearchIdx + 1) % len(m.searchResults)
		// Jump to the line
		if m.currentSearchIdx < len(m.searchResults) {
			targetLine := m.searchResults[m.currentSearchIdx]
			m.inspectScrollY = targetLine - m.height/2 + 3 // Center the result
			if m.inspectScrollY < 0 {
				m.inspectScrollY = 0
			}
		}
	}
	return m, nil
}

func (m *Model) PrevInspectSearchResult(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	if len(m.searchResults) > 0 {
		m.currentSearchIdx--
		if m.currentSearchIdx < 0 {
			m.currentSearchIdx = len(m.searchResults) - 1
		}
		// Jump to the line
		if m.currentSearchIdx < len(m.searchResults) {
			targetLine := m.searchResults[m.currentSearchIdx]
			m.inspectScrollY = targetLine - m.height/2 + 3 // Center the result
			if m.inspectScrollY < 0 {
				m.inspectScrollY = 0
			}
		}
	}
	return m, nil
}

// Help view handlers
func (m *Model) ShowHelp(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.helpViewModel.Show(m, m.currentView)
}

func (m *Model) ScrollHelpUp(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.helpViewModel.HandleScrollUp()
}

func (m *Model) ScrollHelpDown(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.helpViewModel.HandleScrollDown()
}

func (m *Model) CmdPause(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case DockerContainerListView:
		// Pause/Unpause the selected Docker container
		if m.dockerContainerListViewModel.selectedDockerContainer < len(m.dockerContainerListViewModel.dockerContainers) {
			container := m.dockerContainerListViewModel.dockerContainers[m.dockerContainerListViewModel.selectedDockerContainer]
			// Check if container is already paused
			if strings.Contains(container.State, "paused") || strings.Contains(container.Status, "(Paused)") {
				// Container is paused, so unpause it
				m.loading = true
				return m, unpauseService(m.dockerClient, container.ID)
			} else {
				// Container is running, so pause it
				m.loading = true
				return m, pauseService(m.dockerClient, container.ID)
			}
		}
	case ComposeProcessListView:
		if m.composeProcessListViewModel.selectedContainer < len(m.composeProcessListViewModel.composeContainers) {
			container := m.composeProcessListViewModel.composeContainers[m.composeProcessListViewModel.selectedContainer]
			// Check if container is already paused
			if strings.Contains(container.State, "paused") {
				// Container is paused, so unpause it
				m.loading = true
				return m, unpauseService(m.dockerClient, container.ID)
			} else {
				// Container is running, so pause it
				m.loading = true
				return m, pauseService(m.dockerClient, container.ID)
			}
		}
		return m, nil
	default:
		slog.Info("Unhandled :pause command in current view",
			slog.String("view", m.currentView.String()))
	}
	return m, nil
}
