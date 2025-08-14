package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type HelpViewModel struct {
	TableViewModel
	parentView ViewType
}

func (m *HelpViewModel) render(model *Model, availableHeight int) string {
	// Create columns
	columns := []table.Column{
		{Title: "Key", Width: 15}, // Fixed width for keys
		{Title: "Command", Width: -1},
		{Title: "Description", Width: -1},
	}

	return m.RenderTable(model, columns, availableHeight-2, func(row, col int) lipgloss.Style {
		if row == m.Cursor {
			return tableSelectedCellStyle
		}
		return tableNormalCellStyle
	})
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
	m.Cursor = 0
	m.SetRows(m.buildRows(model), model.ViewHeight())
	return nil
}

func (m *HelpViewModel) HandleUp(model *Model) tea.Cmd {
	return m.TableViewModel.HandleUp(model)
}

func (m *HelpViewModel) HandleDown(model *Model) tea.Cmd {
	return m.TableViewModel.HandleDown(model)
}

func (m *HelpViewModel) HandleBack(model *Model) tea.Cmd {
	model.SwitchToPreviousView()
	m.Cursor = 0
	return nil
}
