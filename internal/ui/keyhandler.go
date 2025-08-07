package ui

import (
	"fmt"
	"log/slog"
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

func (m *Model) CmdQuit(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.quitConfirmation = true
	return m, nil
}

func (m *Model) CmdCommandMode(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Start command mode
	m.commandViewModel.Start()
	return m, nil
}

// CmdRefresh sends a RefreshMsg to trigger a reload of the current view
func (m *Model) CmdRefresh(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
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
	case ImageListView:
		return m, m.imageListViewModel.HandleUp()
	case FileContentView:
		return m, m.fileContentViewModel.HandleUp()
	case FileBrowserView:
		return m, m.fileBrowserViewModel.HandleUp()
	case LogView:
		return m, m.logViewModel.HandleUp()
	case InspectView:
		return m, m.inspectViewModel.HandleUp()
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
	case ImageListView:
		return m, m.imageListViewModel.HandleDown()
	case FileContentView:
		return m, m.fileContentViewModel.HandleDown(m.Height)
	case FileBrowserView:
		return m, m.fileBrowserViewModel.HandleDown()
	case LogView:
		return m, m.logViewModel.HandleDown(m)
	case InspectView:
		return m, m.inspectViewModel.HandleDown(m)
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
	case FileContentView:
		return m, m.fileContentViewModel.HandleGoToEnd(m.Height)
	case LogView:
		return m, m.logViewModel.HandleGoToEnd(m)
	case InspectView:
		return m, m.inspectViewModel.HandleGoToEnd(m)
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
	case FileContentView:
		return m, m.fileContentViewModel.HandleGoToStart()
	case LogView:
		return m, m.logViewModel.HandleGoToStart()
	case CommandExecutionView:
		return m, m.commandExecutionViewModel.HandleGoToStart()
	case InspectView:
		return m, m.inspectViewModel.HandleGoToStart()
	default:
		slog.Info("Unhandled go to start in current view",
			slog.String("view", m.currentView.String()))
		return m, nil
	}
}

func (m *Model) CmdFilter(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case LogView:
		return m, m.logViewModel.HandleFilter()
	default:
		slog.Info("Unhandled filter command in current view",
			slog.String("view", m.currentView.String()))
		return m, nil
	}
}

func (m *Model) ShowDindLog(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	// TODO: merge into CmdLog
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
		configs = m.composeProjectListViewHandlers
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

func (m *Model) OpenFileOrDirectory(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.fileBrowserViewModel.HandleOpenFileOrDirectory(m)
}

func (m *Model) GoToParentDirectory(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, m.fileBrowserViewModel.HandleGoToParentDirectory(m)
}

func (m *Model) CmdBack(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case LogView:
		return m, m.logViewModel.HandleBack(m)
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
		return m, m.composeProcessListViewModel.HandleBack(m)
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
	case VolumeListView:
		return m, m.volumeListViewModel.HandleInspect(m)
	case NetworkListView:
		return m, m.networkListViewModel.HandleInspect(m)
	case ImageListView:
		return m, m.imageListViewModel.HandleInspect(m)
	case DockerContainerListView:
		return m, m.dockerContainerListViewModel.HandleInspect(m)
	case ComposeProcessListView:
		return m, m.composeProcessListViewModel.HandleInspect(m)
	default:
		slog.Info("Unhandled :inspect command in current view",
			slog.String("view", m.currentView.String()))
		return m, nil
	}
}

func (m *Model) CmdSearch(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case LogView:
		return m, m.logViewModel.HandleSearch()
	case InspectView:
		return m, m.inspectViewModel.HandleSearch()
	default:
		slog.Info("Unhandled :search command in current view",
			slog.String("view", m.currentView.String()))
		return m, nil
	}
}

func (m *Model) CmdNextSearchResult(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case LogView:
		return m, m.logViewModel.HandleNextSearchResult(m)
	case InspectView:
		return m, m.inspectViewModel.HandleNextSearchResult(m)
	default:
		slog.Info("Unhandled next search result command in current view",
			slog.String("view", m.currentView.String()))
		return m, nil
	}
}

func (m *Model) CmdPrevSearchResult(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case LogView:
		return m, m.logViewModel.HandlePrevSearchResult(m)
	case InspectView:
		return m, m.inspectViewModel.HandlePrevSearchResult(m)
	default:
		slog.Info("Unhandled previous search result command in current view",
			slog.String("view", m.currentView.String()))
		return m, nil
	}
}

func (m *Model) CmdHelp(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
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
