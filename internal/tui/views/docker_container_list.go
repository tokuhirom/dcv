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

// DockerContainerListView displays Docker containers
type DockerContainerListView struct {
	docker     *docker.Client
	table      *tview.Table
	containers []models.DockerContainer
	showAll    bool
}

// NewDockerContainerListView creates a new Docker container list view
func NewDockerContainerListView(dockerClient *docker.Client) *DockerContainerListView {
	v := &DockerContainerListView{
		docker:  dockerClient,
		table:   tview.NewTable(),
		showAll: false,
	}

	v.setupTable()
	v.setupKeyHandlers()
	
	return v
}

// setupTable configures the table widget
func (v *DockerContainerListView) setupTable() {
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
func (v *DockerContainerListView) setupKeyHandlers() {
	v.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := v.table.GetSelection()
		
		switch event.Rune() {
		case 'a', 'A':
			// Toggle show all containers
			v.showAll = !v.showAll
			v.Refresh()
			return nil
			
		case 's':
			// Stop container
			if row > 0 && row <= len(v.containers) {
				container := v.containers[row-1]
				go v.stopContainer(container.ID)
			}
			return nil
			
		case 'S':
			// Start container
			if row > 0 && row <= len(v.containers) {
				container := v.containers[row-1]
				go v.startContainer(container.ID)
			}
			return nil
			
		case 'k':
			// Kill container
			if row > 0 && row <= len(v.containers) {
				container := v.containers[row-1]
				go v.killContainer(container.ID)
			}
			return nil
			
		case 'd':
			// Delete container
			if row > 0 && row <= len(v.containers) {
				container := v.containers[row-1]
				go v.deleteContainer(container.ID)
			}
			return nil
			
		case 'l':
			// View logs
			if row > 0 && row <= len(v.containers) {
				container := v.containers[row-1]
				// TODO: Switch to log view
				slog.Info("View logs for container", slog.String("container", container.ID))
			}
			return nil
			
		case 'e':
			// Execute shell
			if row > 0 && row <= len(v.containers) {
				container := v.containers[row-1]
				go v.execShell(container.ID)
			}
			return nil
			
		case 'i':
			// Inspect container
			if row > 0 && row <= len(v.containers) {
				container := v.containers[row-1]
				// TODO: Switch to inspect view
				slog.Info("Inspect container", slog.String("container", container.ID))
			}
			return nil
		}

		switch event.Key() {
		case tcell.KeyEnter:
			// View container details
			if row > 0 && row <= len(v.containers) {
				container := v.containers[row-1]
				// TODO: Show container details
				slog.Info("View container details", slog.String("container", container.ID))
			}
			return nil
		}

		return event
	})
}

// GetPrimitive returns the tview primitive for this view
func (v *DockerContainerListView) GetPrimitive() tview.Primitive {
	return v.table
}

// Refresh refreshes the container list
func (v *DockerContainerListView) Refresh() {
	go v.loadContainers()
}

// GetTitle returns the title of the view
func (v *DockerContainerListView) GetTitle() string {
	return "Docker Containers"
}

// loadContainers loads the container list from Docker
func (v *DockerContainerListView) loadContainers() {
	slog.Info("Loading Docker containers", slog.Bool("showAll", v.showAll))
	
	containers, err := v.docker.ListContainers(v.showAll)
	if err != nil {
		slog.Error("Failed to load containers", slog.Any("error", err))
		return
	}

	v.containers = containers
	
	// Update table in UI thread
	QueueUpdateDraw(func() {
		v.updateTable()
	})
}

// updateTable updates the table with container data
func (v *DockerContainerListView) updateTable() {
	v.table.Clear()

	// Set headers
	headers := []string{"NAME", "IMAGE", "STATUS", "PORTS", "COMMAND"}
	for col, header := range headers {
		cell := tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetAttributes(tcell.AttrBold).
			SetSelectable(false)
		v.table.SetCell(0, col, cell)
	}

	// Add container rows
	for row, container := range v.containers {
		// Name
		nameCell := tview.NewTableCell(container.Names).
			SetTextColor(tcell.ColorWhite)
		v.table.SetCell(row+1, 0, nameCell)

		// Image
		imageCell := tview.NewTableCell(container.Image).
			SetTextColor(tcell.ColorWhite)
		v.table.SetCell(row+1, 1, imageCell)

		// Status
		statusColor := tcell.ColorRed
		if strings.HasPrefix(container.State, "running") {
			statusColor = tcell.ColorGreen
		}
		statusCell := tview.NewTableCell(container.Status).
			SetTextColor(statusColor)
		v.table.SetCell(row+1, 2, statusCell)

		// Ports
		portsCell := tview.NewTableCell(container.Ports).
			SetTextColor(tcell.ColorWhite)
		v.table.SetCell(row+1, 3, portsCell)

		// Command
		commandCell := tview.NewTableCell(container.Command).
			SetTextColor(tcell.ColorWhite).
			SetMaxWidth(30)
		v.table.SetCell(row+1, 4, commandCell)
	}

	// Select first row if available
	if len(v.containers) > 0 {
		v.table.Select(1, 0)
	}
}

// Container operations
func (v *DockerContainerListView) stopContainer(containerID string) {
	slog.Info("Stopping container", slog.String("container", containerID))
	_, err := docker.ExecuteCaptured("stop", containerID)
	if err != nil {
		slog.Error("Failed to stop container", slog.Any("error", err))
		return
	}
	time.Sleep(500 * time.Millisecond)
	v.Refresh()
}

func (v *DockerContainerListView) startContainer(containerID string) {
	slog.Info("Starting container", slog.String("container", containerID))
	_, err := docker.ExecuteCaptured("start", containerID)
	if err != nil {
		slog.Error("Failed to start container", slog.Any("error", err))
		return
	}
	time.Sleep(500 * time.Millisecond)
	v.Refresh()
}

func (v *DockerContainerListView) killContainer(containerID string) {
	slog.Info("Killing container", slog.String("container", containerID))
	_, err := docker.ExecuteCaptured("kill", containerID)
	if err != nil {
		slog.Error("Failed to kill container", slog.Any("error", err))
		return
	}
	time.Sleep(500 * time.Millisecond)
	v.Refresh()
}

func (v *DockerContainerListView) deleteContainer(containerID string) {
	slog.Info("Deleting container", slog.String("container", containerID))
	_, err := docker.ExecuteCaptured("rm", "-f", containerID)
	if err != nil {
		slog.Error("Failed to delete container", slog.Any("error", err))
		return
	}
	time.Sleep(500 * time.Millisecond)
	v.Refresh()
}

func (v *DockerContainerListView) execShell(containerID string) {
	slog.Info("Executing shell in container", slog.String("container", containerID))
	// TODO: Implement shell execution
	// This would typically launch an external terminal or switch to a shell view
}