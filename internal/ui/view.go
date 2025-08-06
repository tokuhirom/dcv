package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
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
	if m.width == 0 || m.Height == 0 {
		return "Loading..."
	}

	// Get title
	title := m.viewTitle()
	titleHeight := lipgloss.Height(titleStyle.Render(title))

	// Calculate available Height for body content
	// Layout: title (with margin) + body + footer
	titleRendered := titleStyle.Render(title)
	actualTitleHeight := lipgloss.Height(titleRendered)
	footerHeight := 1

	// Available Height = total Height - title Height - footer Height
	availableBodyHeight := m.Height - actualTitleHeight - footerHeight
	if availableBodyHeight < 1 {
		availableBodyHeight = 1
	}

	// Get body content with available Height
	body := m.viewBody(availableBodyHeight)
	bodyHeight := strings.Count(body, "\n") + 1

	// Special handling for HelpView (it has its own footer)
	if m.currentView == HelpView {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			titleStyle.Render(title),
			body,
		)
	}

	// Build footer content (command line or quit confirmation or help hint)
	var footer string

	if m.quitConfirmation {
		// Show quit confirmation dialog
		footer = errorStyle.Render("Really quit? (y/n)")
	} else if m.currentView == LogView && m.logViewModel.filterMode {
		footer = m.logViewModel.RenderCmdLine()
	} else if (m.currentView == LogView && m.logViewModel.searchMode) || (m.currentView == InspectView && m.inspectViewModel.searchMode) {
		// Show search prompt
		var searchText string
		var searchCursorPos int

		if m.currentView == LogView {
			searchText = m.logViewModel.searchText
			searchCursorPos = m.logViewModel.searchCursorPos
		} else {
			searchText = m.inspectViewModel.searchText
			searchCursorPos = m.inspectViewModel.searchCursorPos
		}

		cursor := " "
		if searchCursorPos < len(searchText) {
			cursor = string(searchText[searchCursorPos])
		}

		// Build search line with cursor
		before := searchText[:searchCursorPos]
		after := ""
		if searchCursorPos < len(searchText) {
			after = searchText[searchCursorPos+1:]
		}

		cursorStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("226")).
			Foreground(lipgloss.Color("235"))

		footer = "/" + before + cursorStyle.Render(cursor) + after
	} else if m.commandViewModel.commandMode {
		footer = m.commandViewModel.RenderCmdLine()
	} else {
		// Show help hint
		footer = helpStyle.Render("Press ? for help")
	}

	totalContentHeight := titleHeight + bodyHeight + footerHeight + 1 // +1 for spacing

	// Add padding if needed to push footer to bottom
	if totalContentHeight < m.Height {
		padding := m.Height - totalContentHeight
		body = body + strings.Repeat("\n", padding)
	}

	// Join all components
	return lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render(title),
		body,
		footer,
	)
}

func (m *Model) viewTitle() string {
	switch m.currentView {
	case ComposeProcessListView:
		if m.projectName != "" {
			return fmt.Sprintf("Docker Compose: %s", m.projectName)
		}
		return "Docker Compose"
	case LogView:
		return m.logViewModel.Title()
	case DindProcessListView:
		return fmt.Sprintf("Docker in Docker: %s", m.dindProcessListViewModel.currentDindHost)
	case TopView:
		return fmt.Sprintf("Process Info: %s", m.topViewModel.topService)
	case StatsView:
		return "Container Resource Usage"
	case ComposeProjectListView:
		return "Docker Compose Projects"
	case DockerContainerListView:
		if m.dockerContainerListViewModel.showAll {
			return "Docker Containers (all)"
		}
		return "Docker Containers"
	case ImageListView:
		if m.imageListViewModel.showAll {
			return "Docker Images (all)"
		}
		return "Docker Images"
	case NetworkListView:
		return "Docker Networks"
	case VolumeListView:
		return "Docker Volumes"
	case FileBrowserView:
		return m.fileBrowserViewModel.Title()
	case FileContentView:
		return m.fileContentViewModel.Title()
	case InspectView:
		return m.inspectViewModel.Title()
	case HelpView:
		return "Help"
	case CommandExecutionView:
		return "Command Execution"
	default:
		return "Unknown View"
	}
}

func (m *Model) viewBody(availableHeight int) string {
	// Handle loading state
	if m.loading && m.currentView != LogView {
		return "\nLoading...\n"
	}

	// Handle error state
	if m.err != nil && m.currentView != LogView && m.currentView != FileContentView {
		return "\n" + errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	switch m.currentView {
	case ComposeProcessListView:
		return m.composeProcessListViewModel.render(m, availableHeight)
	case LogView:
		return m.logViewModel.render(m, availableHeight)
	case DindProcessListView:
		return m.dindProcessListViewModel.render(availableHeight)
	case TopView:
		return m.topViewModel.render(m, availableHeight)
	case StatsView:
		return m.statsViewModel.render(m, availableHeight)
	case ComposeProjectListView:
		return m.composeProjectListViewModel.render(m, availableHeight)
	case DockerContainerListView:
		return m.dockerContainerListViewModel.renderDockerList(availableHeight)
	case ImageListView:
		return m.imageListViewModel.render(m, availableHeight)
	case NetworkListView:
		return m.networkListViewModel.render(m, availableHeight)
	case VolumeListView:
		return m.volumeListViewModel.render(m, availableHeight)
	case FileBrowserView:
		return m.fileBrowserViewModel.render(m, availableHeight)
	case FileContentView:
		return m.fileContentViewModel.render(m, availableHeight)
	case InspectView:
		return m.inspectViewModel.render(availableHeight)
	case HelpView:
		return m.helpViewModel.render(m, availableHeight)
	case CommandExecutionView:
		return m.commandExecutionViewModel.render(m)
	default:
		return "Unknown view"
	}
}
