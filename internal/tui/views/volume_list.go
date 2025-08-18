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

// VolumeListView displays Docker volumes
type VolumeListView struct {
	docker                *docker.Client
	table                 *tview.Table
	dockerVolumes         []models.DockerVolume
	pages                 *tview.Pages
	switchToInspectViewFn func(targetID string, target interface{})
}

// NewVolumeListView creates a new volume list view
func NewVolumeListView(dockerClient *docker.Client) *VolumeListView {
	v := &VolumeListView{
		docker: dockerClient,
		table:  tview.NewTable(),
		pages:  tview.NewPages(),
	}

	v.setupTable()
	v.setupKeyHandlers()

	return v
}

// setupTable configures the table widget
func (v *VolumeListView) setupTable() {
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
func (v *VolumeListView) setupKeyHandlers() {
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

		case 'd':
			// Delete volume
			if row > 0 && row <= len(v.dockerVolumes) {
				volume := v.dockerVolumes[row-1]
				v.showConfirmation("volume rm", volume.Name, false)
			}
			return nil

		case 'D':
			// Force delete volume
			if row > 0 && row <= len(v.dockerVolumes) {
				volume := v.dockerVolumes[row-1]
				v.showConfirmation("volume rm -f", volume.Name, true)
			}
			return nil

		case 'p':
			// Prune unused volumes
			v.showPruneConfirmation(false)
			return nil

		case 'P':
			// Prune all volumes
			v.showPruneConfirmation(true)
			return nil

		case 'c':
			// Create volume
			// TODO: Show dialog to get volume configuration
			slog.Info("Create volume")
			return nil

		case 'i':
			// Inspect volume
			if row > 0 && row <= len(v.dockerVolumes) {
				volume := v.dockerVolumes[row-1]
				if v.switchToInspectViewFn != nil {
					v.switchToInspectViewFn(volume.Name, volume)
				} else {
					slog.Info("Inspect volume", slog.String("volume", volume.Name))
				}
			}
			return nil

		case 'r':
			// Refresh list
			v.Refresh()
			return nil

		case '/':
			// Search volumes
			// TODO: Implement search functionality
			slog.Info("Search volumes")
			return nil
		}

		switch event.Key() {
		case tcell.KeyEnter:
			// View volume details
			if row > 0 && row <= len(v.dockerVolumes) {
				volume := v.dockerVolumes[row-1]
				// TODO: Show volume details
				slog.Info("View volume details", slog.String("volume", volume.Name))
			}
			return nil

		case tcell.KeyCtrlR:
			// Force refresh
			v.Refresh()
			return nil
		}

		return event
	})
}

// GetPrimitive returns the tview primitive for this view
func (v *VolumeListView) GetPrimitive() tview.Primitive {
	if v.pages.HasPage("main") {
		return v.pages
	}
	v.pages.AddPage("main", v.table, true, true)
	return v.pages
}

// Refresh refreshes the volume list
func (v *VolumeListView) Refresh() {
	go v.loadVolumes()
}

// GetTitle returns the title of the view
func (v *VolumeListView) GetTitle() string {
	return "Docker Volumes"
}

// SetSwitchToInspectViewCallback sets the callback for switching to inspect view
func (v *VolumeListView) SetSwitchToInspectViewCallback(fn func(targetID string, target interface{})) {
	v.switchToInspectViewFn = fn
}

// loadVolumes loads the volume list from Docker
func (v *VolumeListView) loadVolumes() {
	slog.Info("Loading docker volumes")

	volumes, err := v.docker.ListVolumes()
	if err != nil {
		slog.Error("Failed to load docker volumes", slog.Any("error", err))
		return
	}

	v.dockerVolumes = volumes

	// Update table in UI thread
	QueueUpdateDraw(func() {
		v.updateTable()
	})
}

// updateTable updates the table with volume data
func (v *VolumeListView) updateTable() {
	v.table.Clear()

	// Set headers
	headers := []string{"NAME", "DRIVER", "SCOPE", "MOUNTPOINT", "LABELS"}
	for col, header := range headers {
		cell := tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetAttributes(tcell.AttrBold).
			SetSelectable(false)
		v.table.SetCell(0, col, cell)
	}

	// Add volume rows
	for row, volume := range v.dockerVolumes {
		// Name
		nameColor := tcell.ColorWhite
		// Highlight volumes created by compose (usually have project label)
		if strings.Contains(volume.Labels, "com.docker.compose.project") {
			nameColor = tcell.ColorGreen
		}
		nameCell := tview.NewTableCell(volume.Name).
			SetTextColor(nameColor)
		v.table.SetCell(row+1, 0, nameCell)

		// Driver
		driverColor := tcell.ColorWhite
		if !volume.IsLocal() {
			driverColor = tcell.ColorYellow // Non-local drivers are special
		}
		driverCell := tview.NewTableCell(volume.Driver).
			SetTextColor(driverColor)
		v.table.SetCell(row+1, 1, driverCell)

		// Scope
		scopeCell := tview.NewTableCell(volume.Scope).
			SetTextColor(tcell.ColorWhite)
		v.table.SetCell(row+1, 2, scopeCell)

		// Mountpoint (truncated for display)
		mountpoint := volume.Mountpoint
		if len(mountpoint) > 50 {
			mountpoint = "..." + mountpoint[len(mountpoint)-47:]
		}
		mountpointCell := tview.NewTableCell(mountpoint).
			SetTextColor(tcell.ColorWhite)
		v.table.SetCell(row+1, 3, mountpointCell)

		// Labels (show project if exists)
		labelText := ""
		if projectName := volume.GetLabel("com.docker.compose.project"); projectName != "" {
			labelText = fmt.Sprintf("project=%s", projectName)
			if volumeName := volume.GetLabel("com.docker.compose.volume"); volumeName != "" {
				labelText += fmt.Sprintf(", volume=%s", volumeName)
			}
		} else if volume.Labels != "" {
			// Show first label if no compose labels
			if len(volume.Labels) > 30 {
				labelText = volume.Labels[:27] + "..."
			} else {
				labelText = volume.Labels
			}
		}
		labelsCell := tview.NewTableCell(labelText).
			SetTextColor(tcell.ColorWhite)
		v.table.SetCell(row+1, 4, labelsCell)
	}

	// Select first row if available
	if len(v.dockerVolumes) > 0 {
		v.table.Select(1, 0)
	}
}

// Volume operations
func (v *VolumeListView) deleteVolume(volume models.DockerVolume, force bool) {
	slog.Info("Deleting volume",
		slog.String("volume", volume.Name),
		slog.Bool("force", force))

	args := []string{"volume", "rm"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, volume.Name)

	_, err := docker.ExecuteCaptured(args...)
	if err != nil {
		slog.Error("Failed to delete volume", slog.Any("error", err))
		return
	}
	time.Sleep(500 * time.Millisecond)
	v.Refresh()
}

func (v *VolumeListView) pruneVolumes() {
	slog.Info("Pruning unused volumes")

	output, err := docker.ExecuteCaptured("volume", "prune", "-f")
	if err != nil {
		slog.Error("Failed to prune volumes", slog.Any("error", err))
		return
	}

	slog.Info("Volumes pruned", slog.String("output", string(output)))
	time.Sleep(500 * time.Millisecond)
	v.Refresh()
}

func (v *VolumeListView) pruneAllVolumes() {
	slog.Info("Pruning all unused volumes (including named volumes)")

	// --all flag removes all unused volumes, not just anonymous ones
	output, err := docker.ExecuteCaptured("volume", "prune", "-f", "--all")
	if err != nil {
		slog.Error("Failed to prune all volumes", slog.Any("error", err))
		return
	}

	slog.Info("All volumes pruned", slog.String("output", string(output)))
	time.Sleep(500 * time.Millisecond)
	v.Refresh()
}

// showConfirmation shows a confirmation dialog for aggressive operations
func (v *VolumeListView) showConfirmation(operation string, volumeName string, force bool) {
	var commandText string
	if force {
		commandText = fmt.Sprintf("docker %s -f %s", operation, volumeName)
	} else {
		commandText = fmt.Sprintf("docker %s %s", operation, volumeName)
	}

	text := fmt.Sprintf("Are you sure you want to execute:\n\n%s\n\nVolume: %s", commandText, volumeName)

	onYes := func() {
		v.pages.RemovePage("confirm")
		if row, _ := v.table.GetSelection(); row > 0 && row <= len(v.dockerVolumes) {
			volume := v.dockerVolumes[row-1]
			go v.deleteVolume(volume, force)
		}
	}

	onNo := func() {
		v.pages.RemovePage("confirm")
	}

	modal := CreateConfirmationModal(text, onYes, onNo)
	v.pages.AddPage("confirm", modal, true, true)
}

// showPruneConfirmation shows a confirmation dialog for prune operations
func (v *VolumeListView) showPruneConfirmation(all bool) {
	var commandText, message string
	if all {
		commandText = "docker volume prune -f --all"
		message = "This will remove ALL unused volumes, including named volumes!"
	} else {
		commandText = "docker volume prune -f"
		message = "This will remove all unused anonymous volumes."
	}

	modal := tview.NewModal().
		SetText(fmt.Sprintf("Are you sure you want to execute:\n\n%s\n\n%s", commandText, message)).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			v.pages.RemovePage("confirm")
			if buttonLabel == "Yes" {
				if all {
					go v.pruneAllVolumes()
				} else {
					go v.pruneVolumes()
				}
			}
		})

	v.pages.AddPage("confirm", modal, true, true)
}

// GetSelectedVolume returns the currently selected volume
func (v *VolumeListView) GetSelectedVolume() *models.DockerVolume {
	row, _ := v.table.GetSelection()
	if row > 0 && row <= len(v.dockerVolumes) {
		return &v.dockerVolumes[row-1]
	}
	return nil
}

// CreateVolume creates a new Docker volume
func (v *VolumeListView) CreateVolume(name, driver string, labels map[string]string) error {
	slog.Info("Creating volume",
		slog.String("name", name),
		slog.String("driver", driver))

	args := []string{"volume", "create"}
	if driver != "" && driver != "local" {
		args = append(args, "--driver", driver)
	}
	for key, value := range labels {
		args = append(args, "--label", fmt.Sprintf("%s=%s", key, value))
	}
	if name != "" {
		args = append(args, name)
	}

	_, err := docker.ExecuteCaptured(args...)
	if err != nil {
		return fmt.Errorf("failed to create volume: %w", err)
	}

	time.Sleep(500 * time.Millisecond)
	v.Refresh()
	return nil
}

// SearchVolumes searches for volumes in the list
func (v *VolumeListView) SearchVolumes(query string) {
	if query == "" {
		// Reset to show all loaded volumes
		v.updateTable()
		return
	}

	// Filter volumes based on query
	var filteredVolumes []models.DockerVolume
	lowerQuery := strings.ToLower(query)

	for _, volume := range v.dockerVolumes {
		if strings.Contains(strings.ToLower(volume.Name), lowerQuery) ||
			strings.Contains(strings.ToLower(volume.Driver), lowerQuery) ||
			strings.Contains(strings.ToLower(volume.Labels), lowerQuery) ||
			strings.Contains(strings.ToLower(volume.Scope), lowerQuery) {
			filteredVolumes = append(filteredVolumes, volume)
		}
	}

	// Temporarily replace dockerVolumes for display
	originalVolumes := v.dockerVolumes
	v.dockerVolumes = filteredVolumes
	v.updateTable()
	v.dockerVolumes = originalVolumes
}

// InspectVolume returns detailed information about the selected volume
func (v *VolumeListView) InspectVolume(volume models.DockerVolume) (string, error) {
	output, err := docker.ExecuteCaptured("volume", "inspect", volume.Name)
	if err != nil {
		return "", fmt.Errorf("failed to inspect volume: %w", err)
	}
	return string(output), nil
}

// GetVolumeSize attempts to get the actual size of a volume
// Note: This is not directly supported by Docker API, so it's an approximation
func (v *VolumeListView) GetVolumeSize(volume models.DockerVolume) (string, error) {
	if !volume.IsLocal() {
		return "N/A", nil
	}

	// Use du command to get size (requires appropriate permissions)
	output, err := docker.ExecuteCaptured("run", "--rm", "-v",
		fmt.Sprintf("%s:/volume", volume.Name),
		"alpine", "du", "-sh", "/volume")
	if err != nil {
		return "", fmt.Errorf("failed to get volume size: %w", err)
	}

	// Parse output to get size
	parts := strings.Fields(string(output))
	if len(parts) > 0 {
		return parts[0], nil
	}
	return "0", nil
}
