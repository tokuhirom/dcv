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

// NetworkListView displays Docker networks
type NetworkListView struct {
	docker                *docker.Client
	table                 *tview.Table
	dockerNetworks        []models.DockerNetwork
	pages                 *tview.Pages
	switchToInspectViewFn func(targetID string, target interface{})
}

// NewNetworkListView creates a new network list view
func NewNetworkListView(dockerClient *docker.Client) *NetworkListView {
	v := &NetworkListView{
		docker: dockerClient,
		table:  tview.NewTable(),
		pages:  tview.NewPages(),
	}

	v.setupTable()
	v.setupKeyHandlers()

	return v
}

// setupTable configures the table widget
func (v *NetworkListView) setupTable() {
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
func (v *NetworkListView) setupKeyHandlers() {
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
			// Delete network
			if row > 0 && row <= len(v.dockerNetworks) {
				network := v.dockerNetworks[row-1]
				v.showConfirmation("network rm", network)
			}
			return nil

		case 'c':
			// Create network
			// TODO: Show dialog to get network configuration
			slog.Info("Create network")
			return nil

		case 'i':
			// Inspect network
			if row > 0 && row <= len(v.dockerNetworks) {
				network := v.dockerNetworks[row-1]
				if v.switchToInspectViewFn != nil {
					v.switchToInspectViewFn(network.Name, network)
				} else {
					slog.Info("Inspect network", slog.String("network", network.Name))
				}
			}
			return nil

		case 'p':
			// Prune unused networks
			v.showPruneConfirmation()
			return nil

		case 'r':
			// Refresh list
			v.Refresh()
			return nil

		case '/':
			// Search networks
			// TODO: Implement search functionality
			slog.Info("Search networks")
			return nil
		}

		switch event.Key() {
		case tcell.KeyEnter:
			// View network details
			if row > 0 && row <= len(v.dockerNetworks) {
				network := v.dockerNetworks[row-1]
				// TODO: Show network details or connected containers
				slog.Info("View network details", slog.String("network", network.Name))
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
func (v *NetworkListView) GetPrimitive() tview.Primitive {
	if v.pages.HasPage("main") {
		return v.pages
	}
	v.pages.AddPage("main", v.table, true, true)
	return v.pages
}

// Refresh refreshes the network list
func (v *NetworkListView) Refresh() {
	go v.loadNetworks()
}

// GetTitle returns the title of the view
func (v *NetworkListView) GetTitle() string {
	return "Docker Networks"
}

// SetSwitchToInspectViewCallback sets the callback for switching to inspect view
func (v *NetworkListView) SetSwitchToInspectViewCallback(fn func(targetID string, target interface{})) {
	v.switchToInspectViewFn = fn
}

// loadNetworks loads the network list from Docker
func (v *NetworkListView) loadNetworks() {
	slog.Info("Loading docker networks")

	networks, err := v.docker.ListNetworks()
	if err != nil {
		slog.Error("Failed to load docker networks", slog.Any("error", err))
		return
	}

	v.dockerNetworks = networks

	// Update table in UI thread
	QueueUpdateDraw(func() {
		v.updateTable()
	})
}

// updateTable updates the table with network data
func (v *NetworkListView) updateTable() {
	v.table.Clear()

	// Set headers
	headers := []string{"NETWORK ID", "NAME", "DRIVER", "SCOPE", "CONTAINERS", "SUBNET"}
	for col, header := range headers {
		cell := tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetAttributes(tcell.AttrBold).
			SetSelectable(false)
		v.table.SetCell(0, col, cell)
	}

	// Add network rows
	for row, network := range v.dockerNetworks {
		// Network ID (truncated to 12 chars)
		networkID := network.ID
		if len(networkID) > 12 {
			networkID = networkID[:12]
		}
		idCell := tview.NewTableCell(networkID).
			SetTextColor(tcell.ColorBlue)
		v.table.SetCell(row+1, 0, idCell)

		// Name
		nameColor := tcell.ColorWhite
		// Highlight default networks
		if network.Name == "bridge" || network.Name == "host" || network.Name == "none" {
			nameColor = tcell.ColorGreen
		}
		nameCell := tview.NewTableCell(network.Name).
			SetTextColor(nameColor)
		v.table.SetCell(row+1, 1, nameCell)

		// Driver
		driverCell := tview.NewTableCell(network.Driver).
			SetTextColor(tcell.ColorWhite)
		v.table.SetCell(row+1, 2, driverCell)

		// Scope
		scopeCell := tview.NewTableCell(network.Scope).
			SetTextColor(tcell.ColorWhite)
		v.table.SetCell(row+1, 3, scopeCell)

		// Container count
		containerCount := fmt.Sprintf("%d", network.GetContainerCount())
		containerColor := tcell.ColorWhite
		if network.GetContainerCount() > 0 {
			containerColor = tcell.ColorYellow
		}
		containerCell := tview.NewTableCell(containerCount).
			SetTextColor(containerColor).
			SetAlign(tview.AlignRight)
		v.table.SetCell(row+1, 4, containerCell)

		// Subnet
		subnet := network.GetSubnet()
		if subnet == "" {
			subnet = "-"
		}
		subnetCell := tview.NewTableCell(subnet).
			SetTextColor(tcell.ColorWhite)
		v.table.SetCell(row+1, 5, subnetCell)
	}

	// Select first row if available
	if len(v.dockerNetworks) > 0 {
		v.table.Select(1, 0)
	}
}

// Network operations
func (v *NetworkListView) deleteNetwork(network models.DockerNetwork) {
	// Don't allow removing default networks
	if network.Name == "bridge" || network.Name == "host" || network.Name == "none" {
		slog.Warn("Cannot remove default network", slog.String("network", network.Name))
		return
	}

	slog.Info("Deleting network", slog.String("network", network.Name))

	_, err := docker.ExecuteCaptured("network", "rm", network.ID)
	if err != nil {
		slog.Error("Failed to delete network", slog.Any("error", err))
		return
	}
	time.Sleep(500 * time.Millisecond)
	v.Refresh()
}

func (v *NetworkListView) pruneNetworks() {
	slog.Info("Pruning unused networks")

	output, err := docker.ExecuteCaptured("network", "prune", "-f")
	if err != nil {
		slog.Error("Failed to prune networks", slog.Any("error", err))
		return
	}

	slog.Info("Networks pruned", slog.String("output", string(output)))
	time.Sleep(500 * time.Millisecond)
	v.Refresh()
}

// showConfirmation shows a confirmation dialog for aggressive operations
func (v *NetworkListView) showConfirmation(operation string, network models.DockerNetwork) {
	commandText := fmt.Sprintf("docker %s %s", operation, network.ID)
	text := fmt.Sprintf("Are you sure you want to execute:\n\n%s\n\nNetwork: %s", commandText, network.Name)

	onYes := func() {
		v.pages.RemovePage("confirm")
		go v.deleteNetwork(network)
	}

	onNo := func() {
		v.pages.RemovePage("confirm")
	}

	modal := CreateConfirmationModal(text, onYes, onNo)
	v.pages.AddPage("confirm", modal, true, true)
}

// showPruneConfirmation shows a confirmation dialog for prune operations
func (v *NetworkListView) showPruneConfirmation() {
	commandText := "docker network prune -f"
	message := "This will remove all unused networks."

	modal := tview.NewModal().
		SetText(fmt.Sprintf("Are you sure you want to execute:\n\n%s\n\n%s", commandText, message)).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			v.pages.RemovePage("confirm")
			if buttonLabel == "Yes" {
				go v.pruneNetworks()
			}
		})

	v.pages.AddPage("confirm", modal, true, true)
}

// GetSelectedNetwork returns the currently selected network
func (v *NetworkListView) GetSelectedNetwork() *models.DockerNetwork {
	row, _ := v.table.GetSelection()
	if row > 0 && row <= len(v.dockerNetworks) {
		return &v.dockerNetworks[row-1]
	}
	return nil
}

// CreateNetwork creates a new Docker network
func (v *NetworkListView) CreateNetwork(name, driver string, internal bool) error {
	slog.Info("Creating network",
		slog.String("name", name),
		slog.String("driver", driver),
		slog.Bool("internal", internal))

	args := []string{"network", "create"}
	if driver != "" {
		args = append(args, "--driver", driver)
	}
	if internal {
		args = append(args, "--internal")
	}
	args = append(args, name)

	_, err := docker.ExecuteCaptured(args...)
	if err != nil {
		return fmt.Errorf("failed to create network: %w", err)
	}

	time.Sleep(500 * time.Millisecond)
	v.Refresh()
	return nil
}

// ConnectContainer connects a container to the selected network
func (v *NetworkListView) ConnectContainer(network models.DockerNetwork, containerID string) error {
	slog.Info("Connecting container to network",
		slog.String("network", network.Name),
		slog.String("container", containerID))

	_, err := docker.ExecuteCaptured("network", "connect", network.ID, containerID)
	if err != nil {
		return fmt.Errorf("failed to connect container to network: %w", err)
	}

	time.Sleep(500 * time.Millisecond)
	v.Refresh()
	return nil
}

// DisconnectContainer disconnects a container from the selected network
func (v *NetworkListView) DisconnectContainer(network models.DockerNetwork, containerID string) error {
	slog.Info("Disconnecting container from network",
		slog.String("network", network.Name),
		slog.String("container", containerID))

	_, err := docker.ExecuteCaptured("network", "disconnect", network.ID, containerID)
	if err != nil {
		return fmt.Errorf("failed to disconnect container from network: %w", err)
	}

	time.Sleep(500 * time.Millisecond)
	v.Refresh()
	return nil
}

// SearchNetworks searches for networks in the list
func (v *NetworkListView) SearchNetworks(query string) {
	if query == "" {
		// Reset to show all loaded networks
		v.updateTable()
		return
	}

	// Filter networks based on query
	var filteredNetworks []models.DockerNetwork
	lowerQuery := strings.ToLower(query)

	for _, network := range v.dockerNetworks {
		if strings.Contains(strings.ToLower(network.Name), lowerQuery) ||
			strings.Contains(strings.ToLower(network.ID), lowerQuery) ||
			strings.Contains(strings.ToLower(network.Driver), lowerQuery) ||
			strings.Contains(strings.ToLower(network.Scope), lowerQuery) {
			filteredNetworks = append(filteredNetworks, network)
		}
	}

	// Temporarily replace dockerNetworks for display
	originalNetworks := v.dockerNetworks
	v.dockerNetworks = filteredNetworks
	v.updateTable()
	v.dockerNetworks = originalNetworks
}

// InspectNetwork returns detailed information about the selected network
func (v *NetworkListView) InspectNetwork(network models.DockerNetwork) (string, error) {
	output, err := docker.ExecuteCaptured("network", "inspect", network.ID)
	if err != nil {
		return "", fmt.Errorf("failed to inspect network: %w", err)
	}
	return string(output), nil
}
