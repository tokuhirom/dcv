package ui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/tokuhirom/dcv/internal/models"
)

type ProcessListView struct {
	app       *App
	view      *tview.Flex
	table     *tview.Table
	processes []models.Process
}

func NewProcessListView(app *App) *ProcessListView {
	v := &ProcessListView{
		app:   app,
		table: tview.NewTable(),
	}

	v.setupTable()
	v.setupKeyBindings()
	
	// Create help text
	helpText := "Enter: View logs | d: View dind containers | r: Refresh | q: Quit"
	helpView := tview.NewTextView().SetText(helpText).SetTextAlign(tview.AlignCenter)
	
	// Create flex layout
	v.view = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(v.table, 0, 1, true).
		AddItem(helpView, 1, 0, false)

	return v
}

func (v *ProcessListView) setupTable() {
	v.table.SetBorders(true).
		SetSelectable(true, false).
		SetSelectedStyle(tcell.StyleDefault.Background(tcell.ColorDarkCyan).Foreground(tcell.ColorWhite))

	// Set headers
	headers := []string{"NAME", "IMAGE", "STATUS"}
	for i, header := range headers {
		cell := tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignLeft).
			SetSelectable(false)
		v.table.SetCell(0, i, cell)
	}
}

func (v *ProcessListView) setupKeyBindings() {
	v.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			row, _ := v.table.GetSelection()
			if row > 0 && row <= len(v.processes) {
				process := v.processes[row-1]
				v.app.ShowLogs(process.Name, false)
			}
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'd', 'D':
				row, _ := v.table.GetSelection()
				if row > 0 && row <= len(v.processes) {
					process := v.processes[row-1]
					if process.IsDind {
						v.app.ShowDindProcessList(process.Name)
					}
				}
				return nil
			case 'r', 'R':
				v.Refresh()
				return nil
			case 'q', 'Q':
				v.app.app.Stop()
				return nil
			}
		}
		return event
	})
}

func (v *ProcessListView) Refresh() error {
	processes, err := v.app.dockerClient.ListContainers()
	if err != nil {
		v.showError(err)
		return err
	}

	v.processes = processes
	v.updateTable()
	return nil
}

func (v *ProcessListView) showError(err error) {
	v.app.app.QueueUpdateDraw(func() {
		// Clear table and show error
		v.table.Clear()
		v.table.SetCell(0, 0, tview.NewTableCell("Error").
			SetTextColor(tcell.ColorRed).
			SetAlign(tview.AlignCenter).
			SetExpansion(3).
			SetSelectable(false))
		
		errorMsg := err.Error()
		if strings.Contains(errorMsg, "no configuration file provided") {
			errorMsg = "No docker-compose.yml found in current directory. Please run from a directory with docker-compose.yml"
		}
		
		v.table.SetCell(1, 0, tview.NewTableCell(errorMsg).
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignLeft).
			SetExpansion(3).
			SetSelectable(false))
		
		// Add help text
		v.table.SetCell(3, 0, tview.NewTableCell("Press 'q' to quit or navigate to a directory with docker-compose.yml").
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignCenter).
			SetExpansion(3).
			SetSelectable(false))
	})
}

func (v *ProcessListView) updateTable() {
	// Clear existing rows (except header)
	for row := 1; row < v.table.GetRowCount(); row++ {
		v.table.RemoveRow(row)
	}

	// Add process rows
	for i, process := range v.processes {
		row := i + 1

		nameCell := tview.NewTableCell(process.Name)
		if process.IsDind {
			nameCell.SetTextColor(tcell.ColorGreen)
		}
		v.table.SetCell(row, 0, nameCell)

		v.table.SetCell(row, 1, tview.NewTableCell(process.Image))
		
		statusCell := tview.NewTableCell(process.Status)
		if process.Status == "running" || process.Status == "Up" {
			statusCell.SetTextColor(tcell.ColorGreen)
		} else {
			statusCell.SetTextColor(tcell.ColorRed)
		}
		v.table.SetCell(row, 2, statusCell)
	}

	// Update title
	v.table.SetTitle(fmt.Sprintf(" Docker Compose Processes (%d) ", len(v.processes))).
		SetTitleAlign(tview.AlignLeft).
		SetBorder(true).
		SetBorderPadding(0, 0, 1, 1)
	
	if v.table.GetRowCount() > 1 {
		v.table.Select(1, 0)
	}
}