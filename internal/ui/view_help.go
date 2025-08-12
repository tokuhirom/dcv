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
	rows := m.buildRows(model)

	// Create columns
	columns := []table.Column{
		{Title: "Key", Width: 15},
		{Title: "Command", Width: 30},
		{Title: "Description", Width: 40},
	}

	return RenderTable(columns, rows, availableHeight-2, m.scrollY, model.width)
}

func (m *HelpViewModel) buildRows(model *Model) []table.Row {
	// Build table rows
	var allRows []table.Row

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

func keyHandlersToTableRows(model *Model, keyHandlers []KeyConfig) []table.Row {
	allRows := make([]table.Row, 0, len(keyHandlers))

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

			allRows = append(allRows, table.Row{key, cmdName, config.Description})
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
