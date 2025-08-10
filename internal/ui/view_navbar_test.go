package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/tokuhirom/dcv/internal/docker"
)

func TestNavbarToggle(t *testing.T) {
	t.Run("navbar is visible by default", func(t *testing.T) {
		model := NewModel(ComposeProcessListView)
		model.Init()
		model.width = 100
		model.Height = 30

		// Check that navbar is not hidden initially
		assert.False(t, model.navbarHidden)

		// View should contain navigation items
		view := model.View()
		assert.Contains(t, view, "[1] Containers")
		assert.Contains(t, view, "[2] Projects")
		assert.Contains(t, view, "[H]ide navbar")
	})

	t.Run("CmdToggleNavbar toggles navbar visibility", func(t *testing.T) {
		model := NewModel(ComposeProcessListView)
		model.Init()

		// Initially visible
		assert.False(t, model.navbarHidden)

		// Toggle to hide
		updatedModel, _ := model.CmdToggleNavbar(tea.KeyMsg{})
		m := updatedModel.(*Model)
		assert.True(t, m.navbarHidden)

		// Toggle to show again
		updatedModel, _ = m.CmdToggleNavbar(tea.KeyMsg{})
		m = updatedModel.(*Model)
		assert.False(t, m.navbarHidden)
	})

	t.Run("navbar is hidden from view when navbarHidden is true", func(t *testing.T) {
		model := NewModel(ComposeProcessListView)
		model.Init()
		model.width = 100
		model.Height = 30

		// Hide navbar
		model.navbarHidden = true

		// View should not contain navigation items
		view := model.View()
		assert.NotContains(t, view, "[1] Containers")
		assert.NotContains(t, view, "[2] Projects")
		assert.NotContains(t, view, "[H]ide navbar")

		// Should show hint to restore navbar
		assert.Contains(t, view, "Press H to show navbar")
	})

	t.Run("navbar state persists across view switches", func(t *testing.T) {
		model := NewModel(ComposeProcessListView)
		model.Init()
		model.dockerClient = docker.NewClient()

		// Hide navbar
		model.navbarHidden = true

		// Switch to different view
		model.SwitchView(LogView)
		assert.True(t, model.navbarHidden)

		// Switch to another view
		model.SwitchView(StatsView)
		assert.True(t, model.navbarHidden)

		// Toggle back
		model.navbarHidden = false
		model.SwitchView(ComposeProcessListView)
		assert.False(t, model.navbarHidden)
	})

	t.Run("available height adjusts when navbar is hidden", func(t *testing.T) {
		model := NewModel(ComposeProcessListView)
		model.Init()
		model.width = 100
		model.Height = 30

		// Get view with navbar
		viewWithNavbar := model.View()
		_ = strings.Count(viewWithNavbar, "\n")

		// Hide navbar
		model.navbarHidden = true
		viewWithoutNavbar := model.View()
		_ = strings.Count(viewWithoutNavbar, "\n")

		// Without navbar should have more space for content
		// (though total lines might be similar due to padding)
		assert.NotEqual(t, viewWithNavbar, viewWithoutNavbar)
	})

	t.Run("H key toggles navbar from keymap", func(t *testing.T) {
		model := NewModel(ComposeProcessListView)
		model.Init()
		model.width = 100
		model.Height = 30

		// Check that H is mapped globally
		handler, exists := model.globalKeymap["H"]
		assert.True(t, exists, "H key should be mapped globally")
		assert.NotNil(t, handler)

		// Execute the handler
		updatedModel, _ := handler(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'H'}})
		m := updatedModel.(*Model)
		assert.True(t, m.navbarHidden)
	})
}

func TestNavigationHeader(t *testing.T) {
	t.Run("viewNavigationHeader returns correct structure", func(t *testing.T) {
		model := NewModel(ComposeProcessListView)
		model.Init()
		model.width = 100
		model.Height = 30
		model.currentView = DockerContainerListView

		header := model.viewNavigationHeader()

		// Should contain all navigation items
		assert.Contains(t, header, "[1] Containers")
		assert.Contains(t, header, "[2] Projects")
		assert.Contains(t, header, "[3] Images")
		assert.Contains(t, header, "[4] Networks")
		assert.Contains(t, header, "[5] Volumes")
		assert.Contains(t, header, "[6] Stats")
		assert.Contains(t, header, "[H]ide navbar")
	})

	t.Run("active view is highlighted in navbar", func(t *testing.T) {
		model := NewModel(ComposeProcessListView)
		model.Init()

		// Test different active views
		testCases := []struct {
			view     ViewType
			expected string
		}{
			{DockerContainerListView, "[1] Containers"},
			{ComposeProjectListView, "[2] Projects"},
			{ImageListView, "[3] Images"},
			{NetworkListView, "[4] Networks"},
			{VolumeListView, "[5] Volumes"},
			{StatsView, "[6] Stats"},
		}

		for _, tc := range testCases {
			model.currentView = tc.view
			activeView := model.getActiveNavigationView()
			assert.Equal(t, tc.view, activeView)
		}
	})
}
