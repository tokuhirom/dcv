package ui

import (
	"fmt"
	"log/slog"
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

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("226"))

	statusUpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42"))

	statusDownStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	searchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")).
			Bold(true)

	navActiveStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("86")).
			Foreground(lipgloss.Color("234")).
			Bold(true).
			Padding(0, 1)

	navInactiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("245")).
				Padding(0, 1)

	navSeparatorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("238"))
)

type ContainerAware interface {
	GetContainer(model *Model) *docker.Container
}

const ResetAll = "\x1b[0m"

// View returns the view for the current model
func (m *Model) View() string {
	if m.width == 0 || m.Height == 0 {
		return "Loading..."
	}

	// Get navigation header (only if not hidden)
	var navHeader string
	var navHeight int
	if !m.navbarHidden {
		navHeader = m.viewNavigationHeader()
		navHeight = lipgloss.Height(navHeader)
	}

	// Get title
	title := m.viewTitle()
	titleHeight := lipgloss.Height(titleStyle.Render(title))

	footer := m.viewFooter()
	footerHeight := lipgloss.Height(footer)

	// Available Height = total Height - nav Height - title Height - footer Height
	availableBodyHeight := m.Height - navHeight - titleHeight - footerHeight
	if availableBodyHeight < 1 {
		availableBodyHeight = 1
	}

	// Get body content with available Height
	body := m.viewBody(availableBodyHeight)

	// Note: if body is too long, you can truncate by following code.
	body = lipgloss.NewStyle().MaxHeight(availableBodyHeight).Height(availableBodyHeight).Render(body)

	// Join all components
	components := []string{}
	if !m.navbarHidden {
		components = append(components, navHeader)
	}
	components = append(components,
		titleStyle.Render(title),
		body,
		footer)

	retval := lipgloss.JoinVertical(
		lipgloss.Left,
		components...,
	)

	slog.Info("View rendered",
		slog.Int("expectedHeight", m.Height),
		slog.Int("realHeight", lipgloss.Height(retval)),
		slog.Int("navHeight", lipgloss.Height(navHeader)),
		slog.Int("titleHeight", lipgloss.Height(title)),
		slog.Int("bodyHeight", lipgloss.Height(body)),
		slog.Int("availableBodyHeight", availableBodyHeight),
		slog.Int("footerHeight", lipgloss.Height(footer)))
	return retval
}

func (m *Model) viewNavigationHeader() string {
	// Determine which nav item should be highlighted
	activeView := m.getActiveNavigationView()

	navItems := []string{}

	// Helper function to create nav item
	createNavItem := func(key, label string, viewType ViewType) string {
		item := fmt.Sprintf("[%s] %s", key, label)
		if activeView == viewType {
			return navActiveStyle.Render(item)
		}
		return navInactiveStyle.Render(item)
	}

	// Add navigation items
	navItems = append(navItems, createNavItem("1", "Containers", DockerContainerListView))
	navItems = append(navItems, createNavItem("2", "Projects", ComposeProjectListView))
	navItems = append(navItems, createNavItem("3", "Images", ImageListView))
	navItems = append(navItems, createNavItem("4", "Networks", NetworkListView))
	navItems = append(navItems, createNavItem("5", "Volumes", VolumeListView))
	navItems = append(navItems, createNavItem("6", "Stats", StatsView))

	// Add toggle hint
	toggleHint := helpStyle.Render("[H]ide navbar")
	navItems = append(navItems, toggleHint)

	// Join items with separator
	separator := navSeparatorStyle.Render(" | ")
	navLine := strings.Join(navItems, separator)

	marginBottom := 0
	if m.Height > 20 {
		marginBottom = 1 // Add margin only if there's enough space
	}

	// Add a bottom border
	navContainer := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(lipgloss.Color("238")).
		MarginBottom(marginBottom).
		Render(navLine)

	return navContainer
}

// getActiveNavigationView determines which navigation item should be highlighted
// based on the current view and view history
func (m *Model) getActiveNavigationView() ViewType {
	// Check if current view is one of the main navigation views
	navViews := []ViewType{
		DockerContainerListView,
		ComposeProjectListView,
		ImageListView,
		NetworkListView,
		VolumeListView,
		StatsView,
	}

	// If current view is a main nav view, return it
	for _, v := range navViews {
		if m.currentView == v {
			return m.currentView
		}
	}

	// Special case: ComposeProcessListView should highlight Projects
	if m.currentView == ComposeProcessListView {
		return ComposeProjectListView
	}

	// Otherwise, find the most recent main nav view in history
	for i := len(m.viewHistory) - 1; i >= 0; i-- {
		for _, navView := range navViews {
			if m.viewHistory[i] == navView {
				return m.viewHistory[i]
			}
		}
		// Also check for ComposeProcessListView in history
		if m.viewHistory[i] == ComposeProcessListView {
			return ComposeProjectListView
		}
	}

	// Default to DockerContainerListView if nothing found
	return DockerContainerListView
}

func (m *Model) viewTitle() string {
	switch m.currentView {
	case ComposeProcessListView:
		if m.composeProcessListViewModel.projectName != "" {
			if m.composeProcessListViewModel.showAll {
				return fmt.Sprintf("Docker Compose: %s (all)", m.composeProcessListViewModel.projectName)
			}
			return fmt.Sprintf("Docker Compose: %s", m.composeProcessListViewModel.projectName)
		}
		if m.composeProcessListViewModel.showAll {
			return "Docker Compose (all)"
		}
		return "Docker Compose"
	case LogView:
		return m.logViewModel.Title()
	case DindProcessListView:
		return m.dindProcessListViewModel.Title()
	case TopView:
		return m.topViewModel.Title()
	case StatsView:
		return "Stats"
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
	case CommandActionView:
		return "Select Action"
	case ComposeProjectActionView:
		return "Select Project Action"
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
		return m.dindProcessListViewModel.render(m, availableHeight)
	case TopView:
		return m.topViewModel.render(availableHeight)
	case StatsView:
		return m.statsViewModel.render(m, availableHeight)
	case ComposeProjectListView:
		return m.composeProjectListViewModel.render(m, availableHeight)
	case DockerContainerListView:
		return m.dockerContainerListViewModel.renderDockerList(m, availableHeight)
	case ImageListView:
		return m.imageListViewModel.render(m, availableHeight)
	case NetworkListView:
		return m.networkListViewModel.render(m, availableHeight)
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
	case CommandActionView:
		return m.commandActionViewModel.render(m)
	case ComposeProjectActionView:
		return m.composeProjectActionViewModel.render(m)
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
		// Check if current view supports container search and is in search mode
		vm := m.GetCurrentViewModel()
		if searchable, ok := vm.(ContainerSearchAware); ok && searchable.IsSearchActive() {
			return searchable.RenderSearchLine()
		}

		helpText := "Press ? for help"

		// Add action menu hint for process list views
		switch m.currentView {
		case ComposeProcessListView, DockerContainerListView, DindProcessListView:
			helpText += " | Press x for actions"
		}

		if m.navbarHidden {
			helpText += " | Press H to show navbar"
		}
		return helpStyle.Render(helpText)
	}
}
