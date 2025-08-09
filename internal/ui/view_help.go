package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

type HelpViewModel struct {
	scrollY    int
	parentView ViewType
}

func (m *HelpViewModel) render(model *Model, availableHeight int) string {
	var s strings.Builder

	// Get key configurations based on previous view
	viewConfigs := model.GetViewKeyHandlers(m.parentView)
	viewName := m.parentView.String()

	// Show view name
	s.WriteString(headerStyle.Render(fmt.Sprintf("Keyboard shortcuts for: %s", viewName)) + "\n\n")

	// Build table rows
	var allRows [][]string

	// Add view-specific configs
	if len(viewConfigs) > 0 {
		for _, config := range viewConfigs {
			if len(config.Keys) > 0 {
				key := config.Keys[0]
				if len(config.Keys) > 1 {
					key = strings.Join(config.Keys, "/")
				}

				// Get command name for this handler
				cmdName := model.getCommandForHandler(config.KeyHandler)
				if cmdName != "" {
					cmdName = ":" + cmdName
				}

				allRows = append(allRows, []string{key, cmdName, config.Description})
			}
		}
	}

	allRows = append(allRows, []string{"", "", ""})

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
				cmdName := model.getCommandForHandler(config.KeyHandler)
				if cmdName != "" {
					cmdName = ":" + cmdName
				}

				allRows = append(allRows, []string{key, cmdName, config.Description})
			}
		}
	}

	// Create columns
	columns := []table.Column{
		{Title: "Key", Width: 15},
		{Title: "Command", Width: 30},
		{Title: "Description", Width: 40},
	}

	// Convert visible table rows to table.Row format
	rows := make([]table.Row, len(allRows))
	for i, row := range allRows {
		if len(row) >= 3 {
			rows[i] = table.Row{row[0], row[1], row[2]}
		} else {
			// Handle empty separator rows
			rows[i] = table.Row{"", "", ""}
		}
	}

	// Create the table
	tableHeight := availableHeight - 2
	if tableHeight <= 0 {
		tableHeight = 10
	}

	tableString := RenderTable(columns, rows, tableHeight, m.scrollY)
	s.WriteString(tableString)

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
