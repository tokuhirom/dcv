package ui

import (
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
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	switch m.currentView {
	case ComposeProcessListView:
		return m.renderComposeProcessList()
	case LogView:
		return m.renderLogView()
	case DindComposeProcessListView:
		return m.renderDindList()
	case TopView:
		return m.renderTopView()
	case StatsView:
		return m.renderStatsView()
	case ProjectListView:
		return m.renderProjectList()
	case DockerContainerListView:
		return m.renderDockerList()
	case ImageListView:
		return m.renderImageList()
	case NetworkListView:
		return m.renderNetworkList()
	case FileBrowserView:
		return m.renderFileBrowser()
	case FileContentView:
		return m.renderFileContent()
	case InspectView:
		return m.renderInspectView()
	case HelpView:
		return m.renderHelpView()
	default:
		return "Unknown view"
	}
}
