package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

var (
	tableHeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("240")).
				BorderBottom(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("240"))

	tableSelectedCellStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("229")).
				Background(lipgloss.Color("57")).
				Bold(false)

	tableNormalCellStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Bold(false)

	tableFooterStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Align(lipgloss.Right)
)

type TableViewModel struct {
	Rows   []table.Row
	Start  int
	End    int
	Cursor int
}

func (t *TableViewModel) SetRows(rows []table.Row, height int) {
	t.Rows = rows
	if len(t.Rows) == 0 {
		t.Cursor = 0
		t.Start = 0
		t.End = 0
		return
	}
	if t.Cursor >= len(t.Rows) || t.Cursor < 0 {
		t.Cursor = 0
	}
	if height <= 0 {
		height = 10 // fallback
	}
	// Start/End調整: Cursorが画面内に来るように
	if t.Cursor < t.Start {
		t.Start = t.Cursor
	}
	if t.Cursor >= t.End || t.End-t.Start != height {
		t.Start = t.Cursor - height/2
		if t.Start < 0 {
			t.Start = 0
		}
	}
	t.End = clamp(t.Start+height, 0, len(t.Rows))
	if t.End > len(t.Rows) {
		t.End = len(t.Rows)
	}
}

func (t *TableViewModel) RenderTable(model *Model, columns []table.Column, _ int, styleCallback func(int, int) lipgloss.Style) string {
	// RenderTable the table header
	var s strings.Builder
	var headerLine strings.Builder
	for _, column := range columns {
		headerLine.WriteString(tableNormalCellStyle.Width(column.Width).
			Render(column.Title) + "  ")
	}
	s.WriteString(tableHeaderStyle.Render(headerLine.String()))
	s.WriteString("\n")

	// RenderTable the table rows
	for i, row := range t.Rows[t.Start:t.End] {
		i += t.Start
		for j, cell := range row {
			// Apply the selected style if this is the selected row
			base := styleCallback(i, j)
			s.WriteString(base.Width(columns[j].Width).Height(1).
				Render(runewidth.Truncate(cell, columns[j].Width, "…")))
			// Add space between cells
			if j < len(row)-1 {
				s.WriteString(base.Render("  "))
			}
		}

		s.WriteString("\n")
	}

	// If we exceed the available height, break to avoid overflow
	s.WriteString(tableFooterStyle.
		Width(model.width).
		Render(fmt.Sprintf("... [%d/%d-%d/%d]\n", t.Cursor, t.Start, t.End, len(t.Rows))))

	return s.String()
}

func clamp(v, low, high int) int {
	return min(max(v, low), high)
}

// HandleUp moves selection up in the table
func (t *TableViewModel) HandleUp(model *Model) tea.Cmd {
	height := model.ViewHeight()
	if height <= 0 {
		height = 10 // fallback
	}
	if t.Cursor > 0 {
		t.Cursor--
		if t.Cursor < t.Start {
			t.Start = t.Cursor
		}
		t.End = clamp(t.Start+height, 0, len(t.Rows))
	}
	return nil
}

// HandleDown moves selection down in the table
func (t *TableViewModel) HandleDown(model *Model) tea.Cmd {
	height := model.ViewHeight()
	if height <= 0 {
		height = 10 // fallback
	}
	if t.Cursor < len(t.Rows)-1 {
		t.Cursor++
		if t.Cursor >= t.End {
			t.Start = t.Cursor - height + 1
			if t.Start < 0 {
				t.Start = 0
			}
		}
		t.End = clamp(t.Start+height, 0, len(t.Rows))
	}
	return nil
}
