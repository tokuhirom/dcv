package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type HelpViewModel struct {
	scrollY      int
	previousView ViewType
}

func (m *HelpViewModel) render(model *Model, availableHeight int) string {
	var s strings.Builder

	// Get key configurations based on previous view
	var configs []KeyConfig
	viewName := ""

	switch m.previousView {
	case ComposeProcessListView:
		configs = model.processListViewHandlers
		viewName = "Compose Process List"
	case LogView:
		configs = model.logViewHandlers
		viewName = "Log View"
	case DindProcessListView:
		configs = model.dindListViewHandlers
		viewName = "Docker in Docker"
	case TopView:
		configs = model.topViewHandlers
		viewName = "Process Info"
	case StatsView:
		configs = model.statsViewHandlers
		viewName = "Container Stats"
	case ComposeProjectListView:
		configs = model.projectListViewHandlers
		viewName = "Project List"
	case DockerContainerListView:
		configs = model.dockerContainerListViewHandlers
		viewName = "Docker Containers"
	case ImageListView:
		configs = model.imageListViewHandlers
		viewName = "Docker Images"
	case NetworkListView:
		configs = model.networkListViewHandlers
		viewName = "Docker Networks"
	case VolumeListView:
		configs = model.volumeListViewHandlers
		viewName = "Docker Volumes"
	case FileBrowserView:
		configs = model.fileBrowserHandlers
		viewName = "File Browser"
	case FileContentView:
		configs = model.fileContentHandlers
		viewName = "File Content"
	case InspectView:
		configs = model.inspectViewHandlers
		viewName = "Inspect"
	}

	// Show view name
	s.WriteString(headerStyle.Render(fmt.Sprintf("Keyboard shortcuts for: %s", viewName)) + "\n\n")

	// Column styles
	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Bold(true)
	cmdStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214"))

	// Column headers
	headerKeyStr := keyStyle.Render(fmt.Sprintf("%-12s", "Key"))
	headerCmdStr := cmdStyle.Render(fmt.Sprintf("%-25s", "Command"))
	headerDescStr := normalStyle.Render("Description")
	s.WriteString(fmt.Sprintf("%s %s %s\n", headerKeyStr, headerCmdStr, headerDescStr))
	s.WriteString(strings.Repeat("─", 65) + "\n")

	// Calculate max scroll
	totalLines := len(configs) + 7      // title + view name + header + separator + margins
	visibleLines := availableHeight - 6 // footer + header + margins
	maxScroll := totalLines - visibleLines
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.scrollY > maxScroll {
		m.scrollY = maxScroll
	}

	// Render key bindings
	visibleConfigs := configs
	if m.scrollY > 0 && m.scrollY < len(configs) {
		endIdx := m.scrollY + visibleLines - 5
		if endIdx > len(configs) {
			endIdx = len(configs)
		}
		visibleConfigs = configs[m.scrollY:endIdx]
	}

	for _, config := range visibleConfigs {
		if len(config.Keys) > 0 {
			key := config.Keys[0]
			if len(config.Keys) > 1 {
				key = strings.Join(config.Keys, "/")
			}

			// Get command name for this handler
			cmdName := getCommandForHandler(config.KeyHandler)

			// Format columns
			keyStr := keyStyle.Render(fmt.Sprintf("%-12s", key))
			cmdStr := ""
			if cmdName != "" {
				cmdStr = cmdStyle.Render(fmt.Sprintf("%-25s", ":"+cmdName))
			} else {
				cmdStr = fmt.Sprintf("%-25s", "")
			}
			descStr := normalStyle.Render(config.Description)

			s.WriteString(fmt.Sprintf("%s %s %s\n", keyStr, cmdStr, descStr))
		}
	}

	// Footer
	footer := "\n" + helpStyle.Render("Press ESC or q to go back")
	footerHeight := 2
	contentHeight := availableHeight - footerHeight
	currentHeight := strings.Count(s.String(), "\n") + 1
	if currentHeight < contentHeight {
		s.WriteString(strings.Repeat("\n", contentHeight-currentHeight))
	}
	s.WriteString(footer)

	return s.String()
}

func (m *HelpViewModel) Show(model *Model, previousView ViewType) tea.Cmd {
	m.previousView = previousView
	model.currentView = HelpView
	m.scrollY = 0
	return nil
}

func (m *HelpViewModel) HandleScrollUp() tea.Cmd {
	if m.scrollY > 0 {
		m.scrollY--
	}
	return nil
}

func (m *HelpViewModel) HandleScrollDown() tea.Cmd {
	m.scrollY++
	return nil
}

func (m *HelpViewModel) HandleBack(model *Model) tea.Cmd {
	model.currentView = m.previousView
	m.scrollY = 0
	return nil
}
