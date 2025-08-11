package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// NavbarItem represents a clickable item in the navbar
type NavbarItem struct {
	Label    string
	ViewType ViewType
	Key      string
	StartX   int
	EndX     int
}

// NavbarMouseZone tracks the clickable areas in the navbar
type NavbarMouseZone struct {
	Items      []NavbarItem
	HideButton NavbarItem
}

// calculateNavbarZones calculates the clickable zones for the navbar
func (m *Model) calculateNavbarZones() NavbarMouseZone {
	zones := NavbarMouseZone{
		Items: []NavbarItem{},
	}

	// Track current X position
	currentX := 0

	// Helper function to add nav item and track its position
	addNavItem := func(key, label string, viewType ViewType) {
		item := NavbarItem{
			Label:    label,
			ViewType: viewType,
			Key:      key,
			StartX:   currentX,
		}
		// Format: "[key] label"
		itemText := "[" + key + "] " + label
		itemWidth := lipgloss.Width(itemText)
		item.EndX = currentX + itemWidth
		zones.Items = append(zones.Items, item)

		// Update position for next item (add separator width)
		currentX = item.EndX + 3 // " | " separator
	}

	// Add navigation items in the same order as viewNavigationHeader
	addNavItem("1", "Containers", DockerContainerListView)
	addNavItem("2", "Projects", ComposeProjectListView)
	addNavItem("3", "Images", ImageListView)
	addNavItem("4", "Networks", NetworkListView)
	addNavItem("5", "Volumes", VolumeListView)
	addNavItem("6", "Stats", StatsView)

	// Add hide navbar button
	zones.HideButton = NavbarItem{
		Label:  "[H]ide navbar",
		StartX: currentX,
		EndX:   currentX + lipgloss.Width("[H]ide navbar"),
	}

	return zones
}

// handleMouseClick handles mouse click events on the navbar
func (m *Model) handleNavbarMouseClick(x, y int) (tea.Model, tea.Cmd) {
	// Only handle clicks on the first line (navbar)
	if y != 0 || m.navbarHidden {
		return m, nil
	}

	zones := m.calculateNavbarZones()

	// Check if click is on a navigation item
	for _, item := range zones.Items {
		if x >= item.StartX && x <= item.EndX {
			// Switch to the clicked view
			return m.handleNavigationKey(item.Key)
		}
	}

	// Check if click is on hide button
	if x >= zones.HideButton.StartX && x <= zones.HideButton.EndX {
		m.navbarHidden = true
		return m, nil
	}

	return m, nil
}

// handleNavigationKey handles keyboard navigation shortcuts
func (m *Model) handleNavigationKey(key string) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch key {
	case "1":
		cmd = m.dockerContainerListViewModel.Show(m)
	case "2":
		cmd = m.composeProjectListViewModel.Show(m)
	case "3":
		cmd = m.imageListViewModel.Show(m)
	case "4":
		cmd = m.networkListViewModel.Show(m)
	case "5":
		cmd = m.volumeListViewModel.Show(m)
	case "6":
		cmd = m.statsViewModel.Show(m)
	default:
		return m, nil
	}
	return m, cmd
}

// isNavbarClick checks if a click is in the navbar area
func isNavbarClick(msg tea.MouseMsg) bool {
	// Navbar is on the first line (y=0) and only when left button is pressed
	return msg.Y == 0 && msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft
}
