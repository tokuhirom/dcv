package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

type HelpViewModel struct {
	scrollY    int
	parentView ViewType
}

func (m *HelpViewModel) render(model *Model, availableHeight int) string {
	var s strings.Builder

	// Get key configurations based on previous view
	var viewConfigs []KeyConfig
	viewName := ""

	switch m.parentView {
	case ComposeProcessListView:
		viewConfigs = model.processListViewHandlers
		viewName = "Compose Process List"
	case LogView:
		viewConfigs = model.logViewHandlers
		viewName = "Log View"
	case DindProcessListView:
		viewConfigs = model.dindListViewHandlers
		viewName = "Docker in Docker"
	case TopView:
		viewConfigs = model.topViewHandlers
		viewName = "Process Info"
	case StatsView:
		viewConfigs = model.statsViewHandlers
		viewName = "Container Stats"
	case ComposeProjectListView:
		viewConfigs = model.composeProjectListViewHandlers
		viewName = "Project List"
	case DockerContainerListView:
		viewConfigs = model.dockerContainerListViewHandlers
		viewName = "Docker Containers"
	case ImageListView:
		viewConfigs = model.imageListViewHandlers
		viewName = "Docker Images"
	case NetworkListView:
		viewConfigs = model.networkListViewHandlers
		viewName = "Docker Networks"
	case VolumeListView:
		viewConfigs = model.volumeListViewHandlers
		viewName = "Docker Volumes"
	case FileBrowserView:
		viewConfigs = model.fileBrowserHandlers
		viewName = "File Browser"
	case FileContentView:
		viewConfigs = model.fileContentHandlers
		viewName = "File Content"
	case InspectView:
		viewConfigs = model.inspectViewHandlers
		viewName = "Inspect"
	case HelpView:
		viewConfigs = model.helpViewHandlers
		viewName = "Help"
	case CommandExecutionView:
		viewConfigs = model.commandExecHandlers
		viewName = "Command Execution"
	}

	// Show view name
	s.WriteString(headerStyle.Render(fmt.Sprintf("Keyboard shortcuts for: %s", viewName)) + "\n\n")

	// Build table rows
	var allRows [][]string

	// Add global configs
	if len(model.globalHandlers) > 0 {
		// Add section header as a special row
		for _, config := range model.globalHandlers {
			if len(config.Keys) > 0 {
				key := config.Keys[0]
				if len(config.Keys) > 1 {
					key = strings.Join(config.Keys, "/")
				}

				// Get command name for this handler
				cmdName := getCommandForHandler(config.KeyHandler)
				if cmdName != "" {
					cmdName = ":" + cmdName
				}

				allRows = append(allRows, []string{key, cmdName, config.Description})
			}
		}
	}

	// Add view-specific configs
	if len(viewConfigs) > 0 {
		// Add separator row if we have both
		if len(model.globalHandlers) > 0 {
			allRows = append(allRows, []string{"", "", ""})
		}

		// Add section header
		allRows = append(allRows, []string{"", "", ""})

		for _, config := range viewConfigs {
			if len(config.Keys) > 0 {
				key := config.Keys[0]
				if len(config.Keys) > 1 {
					key = strings.Join(config.Keys, "/")
				}

				// Get command name for this handler
				cmdName := getCommandForHandler(config.KeyHandler)
				if cmdName != "" {
					cmdName = ":" + cmdName
				}

				allRows = append(allRows, []string{key, cmdName, config.Description})
			}
		}
	}

	// Calculate scrolling
	visibleRows := availableHeight - 8 // Account for title, headers, footer
	if visibleRows < 5 {
		visibleRows = 5
	}

	// Adjust scroll position
	maxScroll := len(allRows) - visibleRows
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.scrollY > maxScroll {
		m.scrollY = maxScroll
	}
	if m.scrollY < 0 {
		m.scrollY = 0
	}

	// Apply scrolling to rows
	startIdx := m.scrollY
	endIdx := startIdx + visibleRows
	if endIdx > len(allRows) {
		endIdx = len(allRows)
	}

	// Get visible rows
	var visibleTableRows [][]string
	if len(allRows) > 0 && startIdx < len(allRows) {
		visibleTableRows = allRows[startIdx:endIdx]
	}

	// Create the table
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			// Header row
			if row == 0 {
				return headerStyle
			}

			// Regular rows
			switch col {
			case 0: // Key column
				return lipgloss.NewStyle().
					Foreground(lipgloss.Color("86")).
					Bold(true).
					Width(15)
			case 1: // Command column
				return lipgloss.NewStyle().
					Foreground(lipgloss.Color("214")).
					Width(30)
			case 2: // Description column
				newStyle := normalStyle
				return newStyle.Width(40)
			default:
				return normalStyle
			}
		}).
		Headers("Key", "Command", "Description").
		Rows(visibleTableRows...)

	s.WriteString(t.String())

	// Scroll indicator
	if len(allRows) > visibleRows {
		scrollInfo := fmt.Sprintf(" [Showing %d-%d of %d] ", startIdx+1, endIdx, len(allRows))
		s.WriteString("\n" + helpStyle.Render(scrollInfo))
	}

	// Footer
	footer := "\n" + helpStyle.Render("Press ESC or q to go back")
	s.WriteString(footer)

	return s.String()
}

func (m *HelpViewModel) Show(model *Model, parentView ViewType) tea.Cmd {
	m.parentView = parentView
	model.SwitchView(HelpView)
	m.scrollY = 0
	return nil
}

func (m *HelpViewModel) HandleUp() tea.Cmd {
	if m.scrollY > 0 {
		m.scrollY--
	}
	return nil
}

func (m *HelpViewModel) HandleDown() tea.Cmd {
	m.scrollY++
	return nil
}

func (m *HelpViewModel) HandleBack(model *Model) tea.Cmd {
	model.SwitchToPreviousView()
	m.scrollY = 0
	return nil
}
