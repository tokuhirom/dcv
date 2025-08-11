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

	case launchShellMsg:
		// Execute the interactive command in a subprocess
		c := exec.Command("docker", msg.args...)
		return m, tea.ExecProcess(c, func(err error) tea.Msg {
			// After the command exits, we'll get this message
			m.err = fmt.Errorf("command execution failed: %w, container %s may not contain the %s",
				err,
				msg.container.Title(),
				msg.shell)
			return nil
		})

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
			return m, m.composeProcessListViewModel.DoLoad(m)
		case DindProcessListView:
			return m, m.dindProcessListViewModel.DoLoad(m)
		case LogView:
			// Logs are continuously streamed, no need to refresh
			return m, nil
		case TopView:
			return m, m.topViewModel.DoLoad(m)
		case StatsView:
			return m, m.statsViewModel.DoLoad(m)
		case ComposeProjectListView:
			return m, m.composeProjectListViewModel.DoLoad(m)
		case DockerContainerListView:
			return m, m.dockerContainerListViewModel.DoLoad(m)
		case ImageListView:
			return m, m.imageListViewModel.DoLoad(m)
		case NetworkListView:
			return m, m.networkListViewModel.DoLoad(m)
		case VolumeListView:
			return m, m.volumeListViewModel.DoLoad(m)
		case FileBrowserView:
			return m, m.fileBrowserViewModel.DoLoad(m)
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
		case CommandActionView:
			// Action view doesn't need refresh
			return m, nil
		default:
			m.loading = false
			return m, nil
		}

	case autoRefreshTickMsg:
		// Handle auto-refresh for views that support it (without loading indicator)
		switch m.currentView {
		case TopView:
			if m.topViewModel.autoRefresh {
				// Don't show loading indicator for auto-refresh
				return m, tea.Batch(
					m.topViewModel.DoLoadSilent(m),
					m.topViewModel.startAutoRefresh(),
				)
			}
		case StatsView:
			if m.statsViewModel.autoRefresh {
				// Don't show loading indicator for auto-refresh
				return m, tea.Batch(
					m.statsViewModel.DoLoadSilent(m),
					m.statsViewModel.startAutoRefresh(),
				)
			}
		}
		return m, nil

	default:
		// if UpdateAware
		vm := m.GetCurrentViewModel()
		slog.Debug("Update method needs to call UpdateAware",
			slog.String("ViewType", m.currentView.String()))
		if aware, ok := vm.(UpdateAware); ok {
			// Call the Update method on the current view model
			return aware.Update(m, msg)
		}
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

	// Handle container search mode
	vm := m.GetCurrentViewModel()
	if searchable, ok := vm.(ContainerSearchAware); ok {
		if searchable.IsSearchActive() {
			return m, searchable.HandleSearchInput(m, msg)
		}
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
