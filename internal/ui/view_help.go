package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

type HelpViewModel struct {
	scrollY    int
	parentView ViewType
}

func (m *HelpViewModel) render(model *Model, availableHeight int) string {
	allRows := m.buildRows(model)

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

	return RenderTable(columns, rows, availableHeight-2, m.scrollY)
}

func (m *HelpViewModel) buildRows(model *Model) [][]string {
	// Build table rows
	var allRows [][]string

	// Add view-specific configs
	viewKeyHandlers := model.GetViewKeyHandlers(m.parentView)
	if len(viewKeyHandlers) > 0 {
		allRows = append(allRows, keyHandlersToTableRows(model, viewKeyHandlers)...)
	}

	// Add global configs
	if len(model.globalHandlers) > 0 {
		allRows = append(allRows, []string{"", "", ""})

		allRows = append(allRows, keyHandlersToTableRows(model, model.globalHandlers)...)
	}

	return allRows
}

func keyHandlersToTableRows(model *Model, keyHandlers []KeyConfig) [][]string {
	allRows := make([][]string, 0, len(keyHandlers))

	// Add section header as a special row
	for _, config := range keyHandlers {
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
	return allRows
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

func (m *HelpViewModel) HandleDown(model *Model) tea.Cmd {
	if m.scrollY < len(m.buildRows(model))-1 { // Assuming buildRows returns all rows
		m.scrollY++
	}
	return nil
}

func (m *HelpViewModel) HandleBack(model *Model) tea.Cmd {
	model.SwitchToPreviousView()
	m.scrollY = 0
	return nil
}
