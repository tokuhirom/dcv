package views

import (
	"log/slog"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
)

// ComposeProcessListView displays Docker Compose processes
type ComposeProcessListView struct {
	docker            *docker.Client
	table             *tview.Table
	composeContainers []models.ComposeContainer
	projectName       string
	showAll           bool
}

// NewComposeProcessListView creates a new compose process list view
func NewComposeProcessListView(dockerClient *docker.Client) *ComposeProcessListView {
	v := &ComposeProcessListView{
		docker:  dockerClient,
		table:   tview.NewTable(),
		showAll: false,
	}

	v.setupTable()
	v.setupKeyHandlers()

	return v
}

// setupTable configures the table widget
func (v *ComposeProcessListView) setupTable() {
	v.table.SetBorders(false).
		SetSelectable(true, false).
		SetSeparator(' ').
		SetFixed(1, 0)

	// Set header style
	v.table.SetSelectedStyle(tcell.StyleDefault.
		Background(tcell.ColorDarkCyan).
		Foreground(tcell.ColorWhite))
}

// setupKeyHandlers sets up keyboard shortcuts for the view
func (v *ComposeProcessListView) setupKeyHandlers() {
	v.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := v.table.GetSelection()

		switch event.Rune() {
		case 'j':
			// Move down (vim style)
			if row < v.table.GetRowCount()-1 {
				v.table.Select(row+1, 0)
			}
			return nil

		case 'k':
			// Move up (vim style)
			if row > 1 { // Skip header row
				v.table.Select(row-1, 0)
			}
			return nil

		case 'g':
			// Go to top (vim style)
			v.table.Select(1, 0)
			return nil

		case 'G':
			// Go to bottom (vim style)
			rowCount := v.table.GetRowCount()
			if rowCount > 1 {
				v.table.Select(rowCount-1, 0)
			}
			return nil

		case 'a', 'A':
			// Toggle show all containers
			v.showAll = !v.showAll
			v.Refresh()
			return nil

		case 's':
			// Stop container
			if row > 0 && row <= len(v.composeContainers) {
				container := v.composeContainers[row-1]
				go v.stopContainer(container)
			}
			return nil

		case 'S':
			// Start container
			if row > 0 && row <= len(v.composeContainers) {
				container := v.composeContainers[row-1]
				go v.startContainer(container)
			}
			return nil

		case 'r':
			// Restart container
			if row > 0 && row <= len(v.composeContainers) {
				container := v.composeContainers[row-1]
				go v.restartContainer(container)
			}
			return nil

		case 'd':
			// Delete container
			if row > 0 && row <= len(v.composeContainers) {
				container := v.composeContainers[row-1]
				go v.deleteContainer(container)
			}
			return nil

		case 'l':
			// View logs
			if row > 0 && row <= len(v.composeContainers) {
				container := v.composeContainers[row-1]
				// TODO: Switch to log view
				slog.Info("View logs for compose container", slog.String("container", container.Name))
			}
			return nil

		case 'e':
			// Execute shell
			if row > 0 && row <= len(v.composeContainers) {
				container := v.composeContainers[row-1]
				go v.execShell(container)
			}
			return nil

		case 'i':
			// Inspect container
			if row > 0 && row <= len(v.composeContainers) {
				container := v.composeContainers[row-1]
				// TODO: Switch to inspect view
				slog.Info("Inspect compose container", slog.String("container", container.Name))
			}
			return nil
		}

		switch event.Key() {
		case tcell.KeyEnter:
			// View container details
			if row > 0 && row <= len(v.composeContainers) {
				container := v.composeContainers[row-1]
				// TODO: Show container details
				slog.Info("View compose container details", slog.String("container", container.Name))
			}
			return nil
		}

		return event
	})
}

// SetProject sets the current Docker Compose project
func (v *ComposeProcessListView) SetProject(projectName string) {
	v.projectName = projectName
	v.Refresh()
}

// GetPrimitive returns the tview primitive for this view
func (v *ComposeProcessListView) GetPrimitive() tview.Primitive {
	return v.table
}

// Refresh refreshes the container list
func (v *ComposeProcessListView) Refresh() {
	if v.projectName == "" {
		// No project selected yet
		return
	}
	go v.loadContainers()
}

// GetTitle returns the title of the view
func (v *ComposeProcessListView) GetTitle() string {
	if v.projectName != "" {
		return "Compose: " + v.projectName
	}
	return "Compose Processes"
}

// loadContainers loads the container list from Docker Compose
func (v *ComposeProcessListView) loadContainers() {
	slog.Info("Loading compose containers",
		slog.String("project", v.projectName),
		slog.Bool("showAll", v.showAll))

	if v.projectName == "" {
		return
	}

	containers, err := v.docker.ListComposeContainers(v.projectName, v.showAll)
	if err != nil {
		slog.Error("Failed to load compose containers", slog.Any("error", err))
		return
	}

	v.composeContainers = containers

	// Update table in UI thread
	QueueUpdateDraw(func() {
		v.updateTable()
	})
}

// updateTable updates the table with container data
func (v *ComposeProcessListView) updateTable() {
	v.table.Clear()

	// Set headers
	headers := []string{"SERVICE", "IMAGE", "STATE", "STATUS", "PORTS"}
	for col, header := range headers {
		cell := tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetAttributes(tcell.AttrBold).
			SetSelectable(false)
		v.table.SetCell(0, col, cell)
	}

	// Add container rows
	for row, container := range v.composeContainers {
		// Service name with dind indicator
		service := container.Service
		if container.IsDind() {
			service = "🔄 " + service
		}
		serviceCell := tview.NewTableCell(service).
			SetTextColor(tcell.ColorWhite)
		v.table.SetCell(row+1, 0, serviceCell)

		// Image
		imageCell := tview.NewTableCell(container.Image).
			SetTextColor(tcell.ColorWhite)
		v.table.SetCell(row+1, 1, imageCell)

		// State
		stateCell := tview.NewTableCell(container.State).
			SetTextColor(tcell.ColorWhite)
		v.table.SetCell(row+1, 2, stateCell)

		// Status with color
		status := container.GetStatus()
		statusColor := tcell.ColorRed
		if strings.Contains(status, "Up") || strings.Contains(status, "running") {
			statusColor = tcell.ColorGreen
		}
		statusCell := tview.NewTableCell(status).
			SetTextColor(statusColor)
		v.table.SetCell(row+1, 3, statusCell)

		// Ports
		portsCell := tview.NewTableCell(container.GetPortsString()).
			SetTextColor(tcell.ColorWhite)
		v.table.SetCell(row+1, 4, portsCell)
	}

	// Select first row if available
	if len(v.composeContainers) > 0 {
		v.table.Select(1, 0)
	}
}

// Container operations for Docker Compose
func (v *ComposeProcessListView) stopContainer(container models.ComposeContainer) {
	slog.Info("Stopping compose container",
		slog.String("project", v.projectName),
		slog.String("service", container.Service))

	_, err := docker.ExecuteCaptured("compose", "-p", v.projectName, "stop", container.Service)
	if err != nil {
		slog.Error("Failed to stop compose container", slog.Any("error", err))
		return
	}
	time.Sleep(500 * time.Millisecond)
	v.Refresh()
}

func (v *ComposeProcessListView) startContainer(container models.ComposeContainer) {
	slog.Info("Starting compose container",
		slog.String("project", v.projectName),
		slog.String("service", container.Service))

	_, err := docker.ExecuteCaptured("compose", "-p", v.projectName, "start", container.Service)
	if err != nil {
		slog.Error("Failed to start compose container", slog.Any("error", err))
		return
	}
	time.Sleep(500 * time.Millisecond)
	v.Refresh()
}

func (v *ComposeProcessListView) restartContainer(container models.ComposeContainer) {
	slog.Info("Restarting compose container",
		slog.String("project", v.projectName),
		slog.String("service", container.Service))

	_, err := docker.ExecuteCaptured("compose", "-p", v.projectName, "restart", container.Service)
	if err != nil {
		slog.Error("Failed to restart compose container", slog.Any("error", err))
		return
	}
	time.Sleep(500 * time.Millisecond)
	v.Refresh()
}

func (v *ComposeProcessListView) deleteContainer(container models.ComposeContainer) {
	slog.Info("Deleting compose container",
		slog.String("project", v.projectName),
		slog.String("service", container.Service))

	_, err := docker.ExecuteCaptured("compose", "-p", v.projectName, "rm", "-f", container.Service)
	if err != nil {
		slog.Error("Failed to delete compose container", slog.Any("error", err))
		return
	}
	time.Sleep(500 * time.Millisecond)
	v.Refresh()
}

func (v *ComposeProcessListView) execShell(container models.ComposeContainer) {
	slog.Info("Executing shell in compose container",
		slog.String("project", v.projectName),
		slog.String("service", container.Service))
	// TODO: Implement shell execution
	// This would typically launch an external terminal or switch to a shell view
}
