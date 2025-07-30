package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/tokuhirom/dcv/internal/models"
)

type DindProcessListView struct {
	app           *App
	view          *tview.Flex
	table         *tview.Table
	containerName string
	containers    []models.Container
}

func NewDindProcessListView(app *App) *DindProcessListView {
	v := &DindProcessListView{
		app:   app,
		table: tview.NewTable(),
	}

	v.setupTable()
	v.setupKeyBindings()
	
	// Create help text
	helpText := "Enter: View logs | Esc/q: Back | r: Refresh"
	helpView := tview.NewTextView().SetText(helpText).SetTextAlign(tview.AlignCenter)
	
	// Create flex layout
	v.view = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(v.table, 0, 1, true).
		AddItem(helpView, 1, 0, false)

	return v
}

func (v *DindProcessListView) setupTable() {
	v.table.SetBorders(true).
		SetSelectable(true, false).
		SetSelectedStyle(tcell.StyleDefault.Background(tcell.ColorDarkCyan).Foreground(tcell.ColorWhite))

	// Set headers
	headers := []string{"CONTAINER ID", "IMAGE", "CREATED", "STATUS", "NAME"}
	for i, header := range headers {
		cell := tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignLeft).
			SetSelectable(false)
		v.table.SetCell(0, i, cell)
	}
}

func (v *DindProcessListView) setupKeyBindings() {
	v.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			v.app.ShowProcessList()
			return nil
		case tcell.KeyEnter:
			row, _ := v.table.GetSelection()
			if row > 0 && row <= len(v.containers) {
				container := v.containers[row-1]
				// For dind logs, we need to pass both the host container and target container
				v.app.ShowDindLogs(v.containerName, container.Name)
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
	v.table.SetTitle(fmt.Sprintf(" Docker in Docker: %s ", containerName))
	v.Refresh()
}

func (v *DindProcessListView) Refresh() error {
	containers, err := v.app.dockerClient.ListDindContainers(v.containerName)
	if err != nil {
		// Show error in the view
		v.table.Clear()
		v.table.SetCell(1, 0, tview.NewTableCell(fmt.Sprintf("Error: %v", err)).
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
	for row := 1; row < v.table.GetRowCount(); row++ {
		v.table.RemoveRow(row)
	}

	// Add container rows
	for i, container := range v.containers {
		row := i + 1

		v.table.SetCell(row, 0, tview.NewTableCell(container.ID))
		v.table.SetCell(row, 1, tview.NewTableCell(container.Image))
		v.table.SetCell(row, 2, tview.NewTableCell(container.CreatedAt))
		
		statusCell := tview.NewTableCell(container.Status)
		if container.Status == "Up" || container.Status == "running" {
			statusCell.SetTextColor(tcell.ColorGreen)
		} else {
			statusCell.SetTextColor(tcell.ColorRed)
		}
		v.table.SetCell(row, 3, statusCell)
		
		v.table.SetCell(row, 4, tview.NewTableCell(container.Name))
	}

	// Update title
	v.table.SetTitle(fmt.Sprintf(" Docker in Docker: %s (%d containers) ", v.containerName, len(v.containers))).
		SetTitleAlign(tview.AlignLeft).
		SetBorder(true).
		SetBorderPadding(0, 0, 1, 1)
	
	if v.table.GetRowCount() > 1 {
		v.table.Select(1, 0)
	}
}