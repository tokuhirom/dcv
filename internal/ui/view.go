package ui

import (
	"fmt"
	"path/filepath"
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
		footer = errorStyle.Render(m.quitConfirmMessage)
	} else if m.filterMode && m.currentView == LogView {
		// Show filter prompt
		cursor := " "
		if m.filterCursorPos < len(m.filterText) {
			cursor = string(m.filterText[m.filterCursorPos])
		}

		// Build filter line with cursor
		before := m.filterText[:m.filterCursorPos]
		after := ""
		if m.filterCursorPos < len(m.filterText) {
			after = m.filterText[m.filterCursorPos+1:]
		}

		cursorStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("226")).
			Foreground(lipgloss.Color("235"))

		footer = "Filter: " + before + cursorStyle.Render(cursor) + after + " (ESC to clear)"
	} else if m.searchMode && (m.currentView == LogView || m.currentView == InspectView) {
		// Show search prompt
		cursor := " "
		if m.searchCursorPos < len(m.searchText) {
			cursor = string(m.searchText[m.searchCursorPos])
		}

		// Build search line with cursor
		before := m.searchText[:m.searchCursorPos]
		after := ""
		if m.searchCursorPos < len(m.searchText) {
			after = m.searchText[m.searchCursorPos+1:]
		}

		cursorStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("226")).
			Foreground(lipgloss.Color("235"))

		footer = "/" + before + cursorStyle.Render(cursor) + after
	} else if m.commandMode {
		// Show command line
		cursor := " "
		if m.commandCursorPos < len(m.commandBuffer) {
			cursor = string(m.commandBuffer[m.commandCursorPos])
		}

		// Build command line with cursor
		before := m.commandBuffer[:m.commandCursorPos]
		after := ""
		if m.commandCursorPos < len(m.commandBuffer) {
			after = m.commandBuffer[m.commandCursorPos+1:]
		}

		cursorStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("226")).
			Foreground(lipgloss.Color("235"))

		footer = before + cursorStyle.Render(cursor) + after
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
		title := ""
		if m.isDindLog {
			title = fmt.Sprintf("Logs: %s (in %s)", m.containerName, m.hostContainer)
		} else {
			title = fmt.Sprintf("Logs: %s", m.containerName)
		}

		// Add search or filter status to title
		if m.filterMode && m.filterText != "" {
			filterCount := len(m.filteredLogs)
			title += fmt.Sprintf(" - Filter: '%s' (%d/%d lines)", m.filterText, filterCount, len(m.logs))
		} else if len(m.searchResults) > 0 {
			statusParts := []string{}
			if m.searchIgnoreCase {
				statusParts = append(statusParts, "i")
			}
			if m.searchRegex {
				statusParts = append(statusParts, "r")
			}

			statusStr := ""
			if len(statusParts) > 0 {
				statusStr = fmt.Sprintf(" [%s]", strings.Join(statusParts, ""))
			}

			title += fmt.Sprintf(" - Search: %d/%d%s", m.currentSearchIdx+1, len(m.searchResults), statusStr)
		} else if m.searchText != "" && !m.searchMode {
			title += " - No matches found"
		}

		return title
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
		return fmt.Sprintf("File Browser: %s [%s]", m.browsingContainerName, m.currentPath)
	case FileContentView:
		return fmt.Sprintf("File: %s [%s]", filepath.Base(m.fileContentViewModel.contentPath), m.browsingContainerName)
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
		return m.renderLogView(availableHeight)
	case DindProcessListView:
		return m.dindProcessListViewModel.render(m, availableHeight)
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
		return m.renderFileBrowser(availableHeight)
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
