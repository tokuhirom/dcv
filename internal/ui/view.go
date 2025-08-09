package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/tokuhirom/dcv/internal/docker"
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

type ContainerAware interface {
	GetContainer(model *Model) (docker.Container, error)
}

const ResetAll = "\x1b[0m"

// View returns the view for the current model
func (m *Model) View() string {
	if m.width == 0 || m.Height == 0 {
		return "Loading..."
	}

	// Get title
	title := m.viewTitle()
	titleHeight := lipgloss.Height(titleStyle.Render(title))

	footer := m.viewFooter()
	footerHeight := lipgloss.Height(footer)

	// Available Height = total Height - title Height - footer Height
	availableBodyHeight := m.Height - titleHeight - footerHeight
	if availableBodyHeight < 1 {
		availableBodyHeight = 1
	}

	// Get body content with available Height
	body := m.viewBody(availableBodyHeight)
	bodyHeight := lipgloss.Height(body)

	totalContentHeight := titleHeight + bodyHeight + footerHeight + 1 // +1 for the bottom padding

	// Add padding if needed to push footer to the bottom
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
		if m.composeProcessListViewModel.projectName != "" {
			return fmt.Sprintf("Docker Compose: %s", m.composeProcessListViewModel.projectName)
		}
		return "Docker Compose"
	case LogView:
		return m.logViewModel.Title()
	case DindProcessListView:
		return fmt.Sprintf("Docker in Docker: %s", m.dindProcessListViewModel.currentDindHostName)
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
		return m.imageListViewModel.Title()
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
		return fmt.Sprintf("Help for %s", headerStyle.Render(m.helpViewModel.parentView.String()))
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
		return m.dockerContainerListViewModel.renderDockerList(m, availableHeight)
	case ImageListView:
		return m.imageListViewModel.render(m, availableHeight)
	case NetworkListView:
		return m.networkListViewModel.render(availableHeight)
	case VolumeListView:
		return m.volumeListViewModel.render(m, availableHeight)
	case FileBrowserView:
		return m.fileBrowserViewModel.render(m, availableHeight)
	case FileContentView:
		return m.fileContentViewModel.render(m)
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

func (m *Model) viewFooter() string {
	if m.quitConfirmation {
		// Show quit confirmation dialog
		return errorStyle.Render("Really quit? (y/n)")
	}

	if m.currentView == LogView {
		if m.logViewModel.filterMode {
			return m.logViewModel.RenderFilterCmdLine()
		} else if m.logViewModel.searchMode {
			return m.logViewModel.RenderSearchCmdLine()
		}
	}

	if m.currentView == InspectView && m.inspectViewModel.searchMode {
		return m.inspectViewModel.RenderSearchCmdLine()
	} else if m.commandViewModel.commandMode {
		return m.commandViewModel.RenderCmdLine()
	} else if m.currentView == HelpView {
		return helpStyle.Render("Press ESC or q to go back")
	} else {
		return helpStyle.Render("Press ? for help")
	}
}
