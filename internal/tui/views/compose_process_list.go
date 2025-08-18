package views

import (
	"fmt"
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
	docker                *docker.Client
	table                 *tview.Table
	composeContainers     []models.ComposeContainer
	projectName           string
	showAll               bool
	pages                 *tview.Pages
	switchToLogViewFn     func(containerID string, container interface{})
	switchToFileBrowserFn func(containerID string, container interface{})
	switchToInspectViewFn func(containerID string, container interface{})
}

// NewComposeProcessListView creates a new compose process list view
func NewComposeProcessListView(dockerClient *docker.Client) *ComposeProcessListView {
	v := &ComposeProcessListView{
		docker:  dockerClient,
		table:   tview.NewTable(),
		showAll: false,
		pages:   tview.NewPages(),
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
				v.showConfirmation("compose stop", container)
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
				v.showConfirmation("compose restart", container)
			}
			return nil

		case 'd':
			// Delete container
			if row > 0 && row <= len(v.composeContainers) {
				container := v.composeContainers[row-1]
				v.showConfirmation("compose rm -f", container)
			}
			return nil

		case 'l':
			// View logs
			if row > 0 && row <= len(v.composeContainers) {
				container := v.composeContainers[row-1]
				if v.switchToLogViewFn != nil {
					v.switchToLogViewFn(container.ID, container)
				} else {
					slog.Info("View logs for compose container", slog.String("container", container.Name))
				}
			}
			return nil

		case 'f':
			// Browse files
			if row > 0 && row <= len(v.composeContainers) {
				container := v.composeContainers[row-1]
				if v.switchToFileBrowserFn != nil {
					v.switchToFileBrowserFn(container.ID, container)
				} else {
					slog.Info("Browse files for compose container", slog.String("container", container.Name))
				}
			}
			return nil

		case '!':
			// Execute shell (/bin/sh)
			if row > 0 && row <= len(v.composeContainers) {
				container := v.composeContainers[row-1]
				go v.execShell(container)
			}
			return nil

		case 'x':
			// Show actions menu
			if row > 0 && row <= len(v.composeContainers) {
				container := v.composeContainers[row-1]
				v.showActionsMenu(container)
			}
			return nil

		case 'i':
			// Inspect container
			if row > 0 && row <= len(v.composeContainers) {
				container := v.composeContainers[row-1]
				if v.switchToInspectViewFn != nil {
					v.switchToInspectViewFn(container.ID, container)
				} else {
					slog.Info("Inspect compose container", slog.String("container", container.Name))
				}
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
	if v.pages.HasPage("main") {
		return v.pages
	}
	v.pages.AddPage("main", v.table, true, true)
	return v.pages
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

// SetSwitchToLogViewCallback sets the callback for switching to log view
func (v *ComposeProcessListView) SetSwitchToLogViewCallback(fn func(containerID string, container interface{})) {
	v.switchToLogViewFn = fn
}

// SetSwitchToFileBrowserCallback sets the callback for switching to file browser view
func (v *ComposeProcessListView) SetSwitchToFileBrowserCallback(fn func(containerID string, container interface{})) {
	v.switchToFileBrowserFn = fn
}

// SetSwitchToInspectViewCallback sets the callback for switching to inspect view
func (v *ComposeProcessListView) SetSwitchToInspectViewCallback(fn func(containerID string, container interface{})) {
	v.switchToInspectViewFn = fn
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

// showActionsMenu shows a menu of available actions for the container
func (v *ComposeProcessListView) showActionsMenu(container models.ComposeContainer) {
	// Create a list of actions
	actions := []string{
		"View Logs (l)",
		"Browse Files (f)",
		"Inspect (i)",
		"Execute Shell (!)",
	}

	// Add state-dependent actions
	status := container.GetStatus()
	if strings.Contains(status, "Up") || strings.Contains(status, "running") {
		actions = append(actions, "Stop (s)")
	} else {
		actions = append(actions, "Start (S)")
	}
	actions = append(actions, "Restart (r)")
	actions = append(actions, "Delete (d)")
	actions = append(actions, "Cancel")

	// Create modal with action buttons
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Select action for service: %s", container.Name)).
		AddButtons(actions).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			v.pages.RemovePage("actions")
			// Handle the selected action
			switch buttonLabel {
			case "View Logs (l)":
				if v.switchToLogViewFn != nil {
					v.switchToLogViewFn(container.ID, container)
				}
			case "Browse Files (f)":
				if v.switchToFileBrowserFn != nil {
					v.switchToFileBrowserFn(container.ID, container)
				}
			case "Inspect (i)":
				if v.switchToInspectViewFn != nil {
					v.switchToInspectViewFn(container.ID, container)
				}
			case "Execute Shell (!)":
				go v.execShell(container)
			case "Stop (s)":
				v.showConfirmation("compose stop", container)
			case "Start (S)":
				go v.startContainer(container)
			case "Restart (r)":
				v.showConfirmation("compose restart", container)
			case "Delete (d)":
				v.showConfirmation("compose rm -f", container)
			case "Cancel":
				// Do nothing
			}
		})

	v.pages.AddPage("actions", modal, true, true)
}

// showConfirmation shows a confirmation dialog for aggressive operations
func (v *ComposeProcessListView) showConfirmation(operation string, container models.ComposeContainer) {
	var commandText string
	switch operation {
	case "compose stop":
		commandText = fmt.Sprintf("docker compose -p %s stop %s", v.projectName, container.Service)
	case "compose restart":
		commandText = fmt.Sprintf("docker compose -p %s restart %s", v.projectName, container.Service)
	case "compose rm -f":
		commandText = fmt.Sprintf("docker compose -p %s rm -f %s", v.projectName, container.Service)
	}

	modal := tview.NewModal().
		SetText(fmt.Sprintf("Are you sure you want to execute:\n\n%s\n\nService: %s", commandText, container.Name)).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			v.pages.RemovePage("confirm")
			if buttonLabel == "Yes" {
				switch operation {
				case "compose stop":
					go v.stopContainer(container)
				case "compose restart":
					go v.restartContainer(container)
				case "compose rm -f":
					go v.deleteContainer(container)
				}
			}
		})

	v.pages.AddPage("confirm", modal, true, true)
}
