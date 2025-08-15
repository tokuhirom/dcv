package ui

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

// RenderUnfocusedTable renders a table without selection/focus capability
func RenderUnfocusedTable(columns []table.Column, rows []table.Row, availableHeight int) string {
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(false),
		table.WithHeight(availableHeight-3),
	)

	// Set styles
	tableStyle := table.DefaultStyles()
	tableStyle.Header = tableStyle.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	tableStyle.Cell = tableStyle.Cell.
		BorderForeground(lipgloss.Color("240"))

	t.SetStyles(tableStyle)

	return t.View()
}
