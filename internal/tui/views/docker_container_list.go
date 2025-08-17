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

// DockerContainerListView displays Docker containers
type DockerContainerListView struct {
	docker                *docker.Client
	table                 *tview.Table
	containers            []models.DockerContainer
	showAll               bool
	pages                 *tview.Pages
	switchToLogViewFn     func(containerID string, container interface{})
	switchToFileBrowserFn func(containerID string, container interface{})
}

// NewDockerContainerListView creates a new Docker container list view
func NewDockerContainerListView(dockerClient *docker.Client) *DockerContainerListView {
	v := &DockerContainerListView{
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
			// Go to top (vim style - gg)
			// This is simplified - in real vim you'd need to track double 'g'
			v.table.Select(1, 0) // Select first data row
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
			if row > 0 && row <= len(v.containers) {
				container := v.containers[row-1]
				v.showConfirmation("stop", container.ID, container.Names)
			}
			return nil

		case 'S':
			// Start container
			if row > 0 && row <= len(v.containers) {
				container := v.containers[row-1]
				go v.startContainer(container.ID)
			}
			return nil

		case 'K':
			// Kill container (uppercase K to avoid conflict with vim navigation)
			if row > 0 && row <= len(v.containers) {
				container := v.containers[row-1]
				v.showConfirmation("kill", container.ID, container.Names)
			}
			return nil

		case 'd':
			// Delete container
			if row > 0 && row <= len(v.containers) {
				container := v.containers[row-1]
				v.showConfirmation("rm -f", container.ID, container.Names)
			}
			return nil

		case 'l':
			// View logs
			if row > 0 && row <= len(v.containers) {
				container := v.containers[row-1]
				if v.switchToLogViewFn != nil {
					v.switchToLogViewFn(container.ID, container)
				} else {
					slog.Info("View logs for container", slog.String("container", container.ID))
				}
			}
			return nil

		case 'f':
			// Browse files
			if row > 0 && row <= len(v.containers) {
				container := v.containers[row-1]
				if v.switchToFileBrowserFn != nil {
					v.switchToFileBrowserFn(container.ID, container)
				} else {
					slog.Info("Browse files for container", slog.String("container", container.ID))
				}
			}
			return nil

		case 'e':
			// Execute shell
			if row > 0 && row <= len(v.containers) {
				container := v.containers[row-1]
				go v.execShell(container.ID)
			}
			return nil

		case 'x':
			// Execute command in container
			if row > 0 && row <= len(v.containers) {
				container := v.containers[row-1]
				v.showExecDialog(container.ID, container.Names)
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
	if v.pages.HasPage("main") {
		return v.pages
	}
	v.pages.AddPage("main", v.table, true, true)
	return v.pages
}

// Refresh refreshes the container list
func (v *DockerContainerListView) Refresh() {
	go v.loadContainers()
}

// GetTitle returns the title of the view
func (v *DockerContainerListView) GetTitle() string {
	return "Docker Containers"
}

// SetSwitchToLogViewCallback sets the callback for switching to log view
func (v *DockerContainerListView) SetSwitchToLogViewCallback(fn func(containerID string, container interface{})) {
	v.switchToLogViewFn = fn
}

// SetSwitchToFileBrowserCallback sets the callback for switching to file browser view
func (v *DockerContainerListView) SetSwitchToFileBrowserCallback(fn func(containerID string, container interface{})) {
	v.switchToFileBrowserFn = fn
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

// showExecDialog shows a dialog for executing a command in the container
func (v *DockerContainerListView) showExecDialog(containerID string, containerName string) {
	// Create input field for command
	inputField := tview.NewInputField().
		SetLabel("Command: ").
		SetText("/bin/sh").
		SetFieldWidth(50)

	// Create a form for command input
	form := tview.NewForm().
		AddFormItem(inputField).
		AddButton("Execute", func() {
			command := inputField.GetText()
			if command != "" {
				v.executeCommand(containerID, command)
			}
			v.pages.RemovePage("exec")
		}).
		AddButton("Cancel", func() {
			v.pages.RemovePage("exec")
		})

	// Set form attributes
	form.SetBorder(true).
		SetTitle(fmt.Sprintf(" Execute Command in %s ", containerName)).
		SetTitleAlign(tview.AlignCenter)

	// Calculate position to center the form
	modal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(form, 7, 1, true).
			AddItem(nil, 0, 1, false), 60, 1, true).
		AddItem(nil, 0, 1, false)

	v.pages.AddPage("exec", modal, true, true)
}

// executeCommand executes a command in the container
func (v *DockerContainerListView) executeCommand(containerID string, command string) {
	slog.Info("Executing command in container",
		slog.String("container", containerID),
		slog.String("command", command))

	// Parse the command into parts
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return
	}

	// Execute the command using docker exec
	args := append([]string{"exec", "-it", containerID}, parts...)
	output, err := docker.ExecuteCaptured(args...)
	if err != nil {
		slog.Error("Failed to execute command", slog.Any("error", err))
		// Show error in a modal
		v.showErrorDialog(fmt.Sprintf("Failed to execute command: %v", err))
		return
	}

	// Show output in a modal
	v.showOutputDialog("Command Output", string(output))
}

// showErrorDialog shows an error message in a modal
func (v *DockerContainerListView) showErrorDialog(message string) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			v.pages.RemovePage("error")
		})

	v.pages.AddPage("error", modal, true, true)
}

// showOutputDialog shows command output in a modal
func (v *DockerContainerListView) showOutputDialog(title string, output string) {
	textView := tview.NewTextView().
		SetText(output).
		SetScrollable(true).
		SetWrap(true)

	textView.SetBorder(true).
		SetTitle(fmt.Sprintf(" %s ", title)).
		SetTitleAlign(tview.AlignCenter)

	// Add key handler to close the dialog
	textView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape || event.Rune() == 'q' {
			v.pages.RemovePage("output")
			return nil
		}
		return event
	})

	// Create a flex to center the output
	modal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(textView, 0, 8, true).
			AddItem(nil, 0, 1, false), 0, 8, true).
		AddItem(nil, 0, 1, false)

	v.pages.AddPage("output", modal, true, true)
}

// showConfirmation shows a confirmation dialog for aggressive operations
func (v *DockerContainerListView) showConfirmation(operation string, containerID string, containerName string) {
	var commandText string
	if operation == "rm -f" {
		commandText = fmt.Sprintf("docker rm -f %s", containerID)
	} else {
		commandText = fmt.Sprintf("docker %s %s", operation, containerID)
	}

	modal := tview.NewModal().
		SetText(fmt.Sprintf("Are you sure you want to execute:\n\n%s\n\nContainer: %s", commandText, containerName)).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			v.pages.RemovePage("confirm")
			if buttonLabel == "Yes" {
				switch operation {
				case "stop":
					go v.stopContainer(containerID)
				case "kill":
					go v.killContainer(containerID)
				case "rm -f":
					go v.deleteContainer(containerID)
				}
			}
		})

	v.pages.AddPage("confirm", modal, true, true)
}
