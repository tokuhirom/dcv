package ui

import (
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// RefreshMsg signals that the current view should be refreshed
type RefreshMsg struct{}

// Update handles messages and updates the model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.Height = msg.Height
		return m, nil

	case processesLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		} else {
			m.err = nil
		}

		m.composeProcessListViewModel.Loaded(msg.processes)
		return m, nil

	case dindContainersLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		} else {
			m.err = nil
		}

		m.dindProcessListViewModel.Loaded(msg.containers)
		return m, nil

	// Following 2 cases seems very similar, so we can combine them?
	case logLinesMsg:
		m.logViewModel.LogLines(m, msg.lines)
		// Continue polling for more logs with a small delay
		return m, tea.Tick(time.Millisecond*50, func(time.Time) tea.Msg {
			return m.logViewModel.pollForLogs()()
		})

	case pollLogsContinueMsg:
		// Continue polling with a delay
		return m, tea.Tick(time.Millisecond*50, func(time.Time) tea.Msg {
			return m.logViewModel.pollForLogs()()
		})

	case commandExecutedMsg:
		// HandleStart polling for logs after command is set
		return m, m.logViewModel.pollForLogs()

	case errorMsg:
		m.err = msg.err
		m.loading = false
		return m, nil

	case topLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		} else {
			m.err = nil
		}

		m.topViewModel.Loaded(msg.output)
		return m, nil

	case statsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		} else {
			m.err = nil
		}

		m.statsViewModel.Loaded(msg.stats)
		return m, nil

	case projectsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		} else {
			m.err = nil
		}
		m.composeProjectListViewModel.Loaded(msg.projects)
		return m, nil

	case dockerContainersLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		} else {
			m.err = nil
		}

		m.dockerContainerListViewModel.dockerContainers = msg.containers
		if len(m.dockerContainerListViewModel.dockerContainers) > 0 && m.dockerContainerListViewModel.selectedDockerContainer >= len(m.dockerContainerListViewModel.dockerContainers) {
			m.dockerContainerListViewModel.selectedDockerContainer = 0
		}
		return m, nil

	case dockerImagesLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		} else {
			m.err = nil
		}

		m.imageListViewModel.Loaded(msg.images)
		return m, nil

	case dockerNetworksLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		} else {
			m.err = nil
		}

		m.networkListViewModel.Loaded(msg.networks)
		return m, nil

	case dockerVolumesLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		} else {
			m.err = nil
		}

		m.volumeListViewModel.Loaded(msg.volumes)
		return m, nil

	case containerFilesLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		} else {
			m.err = nil
		}

		m.fileBrowserViewModel.Loaded(msg.files)
		return m, nil

	case fileContentLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		} else {
			m.err = nil
		}

		m.fileContentViewModel.Loaded(msg.content, msg.path)
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
		} else {
			m.err = nil
		}

		m.inspectViewModel.Set(msg.content, msg.targetName)
		m.SwitchView(InspectView)
		return m, nil

	case commandExecStartedMsg:
		// HandleStart reading output
		return m, m.commandExecutionViewModel.ExecStarted(msg.cmd, msg.stdout, msg.stderr)

	case commandExecOutputMsg:
		return m, m.commandExecutionViewModel.ExecOutput(m, msg.line)

	case commandExecCompleteMsg:
		m.commandExecutionViewModel.Complete(msg.exitCode)
		return m, nil

	case RefreshMsg:
		// Handle refresh based on current view
		m.loading = true
		m.err = nil

		switch m.currentView {
		case ComposeProcessListView:
			return m, loadComposeProcesses(m.dockerClient, m.composeProcessListViewModel.projectName, m.composeProcessListViewModel.showAll)
		case DindProcessListView:
			return m, m.dindProcessListViewModel.DoLoad(m)
		case LogView:
			// Logs are continuously streamed, no need to refresh
			return m, nil
		case TopView:
			return m, m.topViewModel.HandleRefresh(m)
		case StatsView:
			return m, m.statsViewModel.HandleRefresh(m)
		case ComposeProjectListView:
			return m, loadProjects(m.dockerClient)
		case DockerContainerListView:
			return m, m.dockerContainerListViewModel.DoLoad(m)
		case ImageListView:
			return m, loadDockerImages(m.dockerClient, m.imageListViewModel.showAll)
		case NetworkListView:
			return m, m.networkListViewModel.Show(m, true)
		case VolumeListView:
			return m, m.volumeListViewModel.HandleRefresh(m)
		case FileBrowserView:
			return m, m.fileBrowserViewModel.HandleRefresh(m)
		case FileContentView:
			// File content doesn't need refresh, it's static
			return m, nil
		case InspectView:
			// Inspect view doesn't need refresh, it's static
			return m, nil
		case HelpView:
			// Help view doesn't need refresh
			return m, nil
		case CommandExecutionView:
			// Command execution is already running, no refresh needed
			return m, nil
		default:
			m.loading = false
			return m, nil
		}

	default:
		return m, nil
	}
}

func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle command mode first
	if m.commandViewModel.commandMode {
		return m.commandViewModel.HandleKeys(m, msg)
	}

	// Handle quit confirmation dialog
	if m.quitConfirmation {
		return m.handleQuitConfirmation(msg)
	}

	// Handle command execution confirmation dialog
	if m.currentView == CommandExecutionView && m.commandExecutionViewModel.pendingConfirmation {
		switch msg.String() {
		case "y", "Y":
			cmd := m.commandExecutionViewModel.HandleConfirmation(m, true)
			return m, cmd
		case "n", "N", "esc":
			cmd := m.commandExecutionViewModel.HandleConfirmation(m, false)
			return m, cmd
		default:
			// Ignore other keys during confirmation
			return m, nil
		}
	}

	// Handle search mode
	if m.currentView == LogView && m.logViewModel.searchMode {
		return m.handleSearchMode(msg, &m.logViewModel.SearchViewModel)
	} else if m.currentView == InspectView && m.inspectViewModel.searchMode {
		return m.handleSearchMode(msg, &m.inspectViewModel.SearchViewModel)
	}

	// Handle filter mode
	if m.currentView == LogView && m.logViewModel.filterMode {
		return m.handleFilterMode(msg)
	}

	handler, ok := m.globalKeymap[msg.String()]
	if ok {
		return handler(msg)
	}

	// Handle view-specific keys
	return m.handleViewKeys(msg)
}

// handleViewKeys handles key presses for the current view using the generic keymap
func (m *Model) handleViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Special case for ComposeProcessListView logging
	if m.currentView == ComposeProcessListView {
		slog.Info(fmt.Sprintf("Key: %s", msg.String()),
			slog.Bool("ok", m.composeProcessListViewKeymap != nil),
			slog.Any("handler", m.composeProcessListViewKeymap))
	}

	keymap := m.GetViewKeymap(m.currentView)
	if keymap != nil {
		handler, ok := keymap[msg.String()]
		if ok {
			return handler(msg)
		}
	}
	return m, nil
}

func (m *Model) handleSearchMode(msg tea.KeyMsg, searchViewModel *SearchViewModel) (tea.Model, tea.Cmd) {
	performSearch := func() {
		switch m.currentView {
		case LogView:
			m.logViewModel.PerformSearch(m, m.logViewModel.logs, func(scrollY int) { m.logViewModel.logScrollY = scrollY })
		case InspectView:
			m.inspectViewModel.PerformSearch(m, strings.Split(m.inspectViewModel.inspectContent, "\n"), func(scrollY int) { m.inspectViewModel.inspectScrollY = scrollY })
		default:
			panic("unhandled default case")
		}
	}

	// TODO: support CtrlD/Del
	switch msg.Type {
	case tea.KeyEsc:
		searchViewModel.InputEscape()
		return m, nil

	case tea.KeyEnter:
		searchViewModel.searchMode = false
		performSearch()
		return m, nil

	case tea.KeyBackspace, tea.KeyCtrlH:
		updated := searchViewModel.DeleteLastChar()
		if updated {
			performSearch()
		}
		return m, nil

	case tea.KeyLeft, tea.KeyCtrlB:
		searchViewModel.CursorLeft()
		return m, nil

	case tea.KeyRight, tea.KeyCtrlF:
		searchViewModel.CursorRight()
		return m, nil

	case tea.KeyCtrlI:
		searchViewModel.ToggleIgnoreCase()
		performSearch()
		return m, nil

	case tea.KeyCtrlR:
		searchViewModel.ToggleRegex()
		performSearch()
		return m, nil

	default:
		if msg.Type == tea.KeyRunes {
			searchViewModel.AppendString(msg.String())
			performSearch()
		}
		return m, nil
	}
}

func (m *Model) handleQuitConfirmation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		// Confirm quit
		m.quitConfirmation = false
		return m, tea.Quit
	case "n", "N", "esc":
		// Cancel quit
		m.quitConfirmation = false
		return m, nil
	}
	return m, nil
}

func (m *Model) handleFilterMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Check if ESC was pressed to clear filter
	if msg.Type == tea.KeyEsc {
		m.logViewModel.ClearFilter()
		m.logViewModel.logScrollY = 0 // Reset scroll position when clearing filter
		return m, nil
	}

	perform := m.logViewModel.HandleKey(msg)
	if perform {
		m.logViewModel.performFilter()
	}
	return m, nil
}
