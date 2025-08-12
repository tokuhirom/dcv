package ui

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

func RenderTable(columns []table.Column, rows []table.Row, availableHeight int, selectedRow int, width int) string {
	// Apply highlighting to the selected row
	highlightedRows := make([]table.Row, len(rows))
	selectedFg := lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57"))

	for i, row := range rows {
		if i == selectedRow {
			// Apply background color to all cells in the selected row
			highlightedRow := make(table.Row, len(row))
			for j, cell := range row {
				highlightedRow[j] = selectedFg.Render(cell)
			}
			highlightedRows[i] = highlightedRow
		} else {
			highlightedRows[i] = row
		}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(highlightedRows),
		table.WithFocused(false), // Disable focus to prevent cell highlighting
		table.WithHeight(availableHeight-3),
		table.WithWidth(width),
	)

	// Set styles
	tableStyle := table.DefaultStyles()
	tableStyle.Header = tableStyle.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	tableStyle.Cell = tableStyle.Cell.
		Foreground(lipgloss.Color("252"))

	t.SetStyles(tableStyle)

	return t.View()
}

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
