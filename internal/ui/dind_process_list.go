package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/tokuhirom/dcv/internal/models"
)

type DindProcessListView struct {
	app           *App
	view          *tview.Table
	containerName string
	containers    []models.Container
}

func NewDindProcessListView(app *App) *DindProcessListView {
	v := &DindProcessListView{
		app:  app,
		view: tview.NewTable(),
	}

	v.setupTable()
	v.setupKeyBindings()

	return v
}

func (v *DindProcessListView) setupTable() {
	v.view.SetBorders(true).
		SetSelectable(true, false).
		SetSelectedStyle(tcell.StyleDefault.Background(tcell.ColorDarkCyan).Foreground(tcell.ColorWhite))

	// Set headers
	headers := []string{"CONTAINER ID", "IMAGE", "CREATED", "STATUS", "NAME"}
	for i, header := range headers {
		cell := tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignLeft).
			SetSelectable(false)
		v.view.SetCell(0, i, cell)
	}
}

func (v *DindProcessListView) setupKeyBindings() {
	v.view.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			v.app.ShowProcessList()
			return nil
		case tcell.KeyEnter:
			row, _ := v.view.GetSelection()
			if row > 0 && row <= len(v.containers) {
				container := v.containers[row-1]
				v.app.logView.SetContainer(container.Name, true)
				v.app.logView.view.SetTitle(fmt.Sprintf(" Logs: %s (in %s) ", container.Name, v.containerName))
				v.app.pages.SwitchToPage("logs")
			}
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q', 'Q':
				v.app.ShowProcessList()
				return nil
			case 'r', 'R':
				v.Refresh()
				return nil
			}
		}
		return event
	})
}

func (v *DindProcessListView) SetContainer(containerName string) {
	v.containerName = containerName
	v.view.SetTitle(fmt.Sprintf(" Docker in Docker: %s ", containerName))
	v.Refresh()
}

func (v *DindProcessListView) Refresh() error {
	containers, err := v.app.dockerClient.ListDindContainers(v.containerName)
	if err != nil {
		// Show error in the view
		v.view.Clear()
		v.view.SetCell(1, 0, tview.NewTableCell(fmt.Sprintf("Error: %v", err)).
			SetTextColor(tcell.ColorRed).
			SetExpansion(5))
		return err
	}

	v.containers = containers
	v.updateTable()
	return nil
}

func (v *DindProcessListView) updateTable() {
	// Clear existing rows (except header)
	for row := 1; row < v.view.GetRowCount(); row++ {
		v.view.RemoveRow(row)
	}

	// Add container rows
	for i, container := range v.containers {
		row := i + 1

		v.view.SetCell(row, 0, tview.NewTableCell(container.ID))
		v.view.SetCell(row, 1, tview.NewTableCell(container.Image))
		v.view.SetCell(row, 2, tview.NewTableCell(container.CreatedAt))
		
		statusCell := tview.NewTableCell(container.Status)
		if container.Status == "Up" || container.Status == "running" {
			statusCell.SetTextColor(tcell.ColorGreen)
		} else {
			statusCell.SetTextColor(tcell.ColorRed)
		}
		v.view.SetCell(row, 3, statusCell)
		
		v.view.SetCell(row, 4, tview.NewTableCell(container.Name))
	}

	// Add help text
	helpText := "Enter: View logs | Esc/q: Back | r: Refresh"
	v.view.SetTitle(fmt.Sprintf(" Docker in Docker: %s (%d containers) ", v.containerName, len(v.containers))).
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
	
	v.app.pages.RemovePage("dind")
	v.app.pages.AddPage("dind", flex, true, false)
}