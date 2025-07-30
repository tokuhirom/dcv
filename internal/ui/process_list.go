package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/tokuhirom/dcv/internal/models"
)

type ProcessListView struct {
	app       *App
	view      *tview.Table
	processes []models.Process
}

func NewProcessListView(app *App) *ProcessListView {
	v := &ProcessListView{
		app:  app,
		view: tview.NewTable(),
	}

	v.setupTable()
	v.setupKeyBindings()

	return v
}

func (v *ProcessListView) setupTable() {
	v.view.SetBorders(true).
		SetSelectable(true, false).
		SetSelectedStyle(tcell.StyleDefault.Background(tcell.ColorDarkCyan).Foreground(tcell.ColorWhite))

	// Set headers
	headers := []string{"NAME", "IMAGE", "STATUS"}
	for i, header := range headers {
		cell := tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignLeft).
			SetSelectable(false)
		v.view.SetCell(0, i, cell)
	}
}

func (v *ProcessListView) setupKeyBindings() {
	v.view.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			row, _ := v.view.GetSelection()
			if row > 0 && row <= len(v.processes) {
				process := v.processes[row-1]
				v.app.ShowLogs(process.Name, false)
			}
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'd', 'D':
				row, _ := v.view.GetSelection()
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
		return err
	}

	v.processes = processes
	v.updateTable()
	return nil
}

func (v *ProcessListView) updateTable() {
	// Clear existing rows (except header)
	for row := 1; row < v.view.GetRowCount(); row++ {
		v.view.RemoveRow(row)
	}

	// Add process rows
	for i, process := range v.processes {
		row := i + 1

		nameCell := tview.NewTableCell(process.Name)
		if process.IsDind {
			nameCell.SetTextColor(tcell.ColorGreen)
		}
		v.view.SetCell(row, 0, nameCell)

		v.view.SetCell(row, 1, tview.NewTableCell(process.Image))
		
		statusCell := tview.NewTableCell(process.Status)
		if process.Status == "running" || process.Status == "Up" {
			statusCell.SetTextColor(tcell.ColorGreen)
		} else {
			statusCell.SetTextColor(tcell.ColorRed)
		}
		v.view.SetCell(row, 2, statusCell)
	}

	// Add help text
	helpText := "Enter: View logs | d: View dind containers | r: Refresh | q: Quit"
	v.view.SetTitle(fmt.Sprintf(" Docker Compose Processes (%d) ", len(v.processes))).
		SetTitleAlign(tview.AlignLeft).
		SetBorder(true).
		SetBorderPadding(0, 0, 1, 1)
	
	if v.view.GetRowCount() > 1 {
		v.view.Select(1, 0)
	}
	
	// Create a flex layout to add help text at the bottom
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(v.view, 0, 1, true).
		AddItem(tview.NewTextView().SetText(helpText).SetTextAlign(tview.AlignCenter), 1, 0, false)
	
	v.app.pages.RemovePage("processes")
	v.app.pages.AddPage("processes", flex, true, true)
}