package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) renderHelpView() string {
	var s strings.Builder

	// Get key configurations based on previous view
	var configs []KeyConfig
	viewName := ""

	switch m.previousView {
	case ComposeProcessListView:
		configs = m.processListViewHandlers
		viewName = "Compose Process List"
	case LogView:
		configs = m.logViewHandlers
		viewName = "Log View"
	case DindComposeProcessListView:
		configs = m.dindListViewHandlers
		viewName = "Docker in Docker"
	case TopView:
		configs = m.topViewHandlers
		viewName = "Process Info"
	case StatsView:
		configs = m.statsViewHandlers
		viewName = "Container Stats"
	case ProjectListView:
		configs = m.projectListViewHandlers
		viewName = "Project List"
	case DockerContainerListView:
		configs = m.dockerListViewHandlers
		viewName = "Docker Containers"
	case ImageListView:
		configs = m.imageListViewHandlers
		viewName = "Docker Images"
	case NetworkListView:
		configs = m.networkListViewHandlers
		viewName = "Docker Networks"
	case FileBrowserView:
		configs = m.fileBrowserHandlers
		viewName = "File Browser"
	case FileContentView:
		configs = m.fileContentHandlers
		viewName = "File Content"
	case InspectView:
		configs = m.inspectViewHandlers
		viewName = "Inspect"
	}

	// Show view name
	s.WriteString(headerStyle.Render(fmt.Sprintf("Keyboard shortcuts for: %s", viewName)) + "\n\n")

	// Calculate max scroll
	totalLines := len(configs) + 5 // title + view name + margins
	visibleLines := m.height - 4   // footer + margins
	maxScroll := totalLines - visibleLines
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.helpScrollY > maxScroll {
		m.helpScrollY = maxScroll
	}

	// Render key bindings
	visibleConfigs := configs
	if m.helpScrollY > 0 && m.helpScrollY < len(configs) {
		endIdx := m.helpScrollY + visibleLines - 5
		if endIdx > len(configs) {
			endIdx = len(configs)
		}
		visibleConfigs = configs[m.helpScrollY:endIdx]
	}

	for _, config := range visibleConfigs {
		if len(config.Keys) > 0 {
			key := config.Keys[0]
			if len(config.Keys) > 1 {
				key = strings.Join(config.Keys, "/")
			}
			keyStr := lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Bold(true).
				Render(fmt.Sprintf("%-15s", key))
			descStr := normalStyle.Render(config.Description)
			s.WriteString(fmt.Sprintf("%s %s\n", keyStr, descStr))
		}
	}

	// Footer
	footer := "\n" + helpStyle.Render("Press ESC or q to go back")
	footerHeight := 2
	contentHeight := m.height - footerHeight
	currentHeight := strings.Count(s.String(), "\n") + 1
	if currentHeight < contentHeight {
		s.WriteString(strings.Repeat("\n", contentHeight-currentHeight))
	}
	s.WriteString(footer)

	return s.String()
}
