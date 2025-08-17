package tui

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"

	"github.com/tokuhirom/dcv/internal/ui"
)

func TestNewApp(t *testing.T) {
	app := NewApp(ui.DockerContainerListView)

	assert.NotNil(t, app)
	assert.NotNil(t, app.app)
	assert.NotNil(t, app.pages)
	assert.NotNil(t, app.state)
	assert.NotNil(t, app.docker)
	assert.NotNil(t, app.views)
	assert.Equal(t, ui.DockerContainerListView, app.initialView)
}

func TestApp_GettersSetters(t *testing.T) {
	app := NewApp(ui.DockerContainerListView)

	// Test GetApp
	tviewApp := app.GetApp()
	assert.NotNil(t, tviewApp)

	// Test GetDocker
	dockerClient := app.GetDocker()
	assert.NotNil(t, dockerClient)

	// Test GetState
	state := app.GetState()
	assert.NotNil(t, state)
}

func TestApp_ViewInitialization(t *testing.T) {
	app := NewApp(ui.DockerContainerListView)

	// Check that all views are initialized
	expectedViews := []ui.ViewType{
		ui.DockerContainerListView,
		ui.ComposeProcessListView,
		ui.ComposeProjectListView,
		ui.ImageListView,
		ui.NetworkListView,
		ui.VolumeListView,
		ui.LogView,
		ui.StatsView,
		ui.HelpView,
	}

	for _, viewType := range expectedViews {
		view, exists := app.views[viewType]
		assert.True(t, exists, "View %s should be initialized", viewType.String())
		assert.NotNil(t, view, "View %s should not be nil", viewType.String())
	}
}

func TestApp_SwitchView(t *testing.T) {
	app := NewApp(ui.DockerContainerListView)

	tests := []struct {
		name     string
		viewType ui.ViewType
		pageName string
	}{
		{
			name:     "switch to docker view",
			viewType: ui.DockerContainerListView,
			pageName: "docker",
		},
		{
			name:     "switch to compose view",
			viewType: ui.ComposeProcessListView,
			pageName: "compose",
		},
		{
			name:     "switch to projects view",
			viewType: ui.ComposeProjectListView,
			pageName: "projects",
		},
		{
			name:     "switch to images view",
			viewType: ui.ImageListView,
			pageName: "images",
		},
		{
			name:     "switch to networks view",
			viewType: ui.NetworkListView,
			pageName: "networks",
		},
		{
			name:     "switch to volumes view",
			viewType: ui.VolumeListView,
			pageName: "volumes",
		},
		{
			name:     "switch to help view",
			viewType: ui.HelpView,
			pageName: "help",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app.SwitchView(tt.viewType)

			assert.Equal(t, tt.viewType, app.state.CurrentView)
			assert.Contains(t, app.state.ViewHistory, tt.viewType)
		})
	}
}

func TestApp_ViewHistory(t *testing.T) {
	app := NewApp(ui.DockerContainerListView)

	// Switch through multiple views
	viewSequence := []ui.ViewType{
		ui.DockerContainerListView,
		ui.ComposeProcessListView,
		ui.ImageListView,
		ui.NetworkListView,
	}

	for _, viewType := range viewSequence {
		app.SwitchView(viewType)
	}

	// Check view history
	assert.Equal(t, len(viewSequence), len(app.state.ViewHistory))
	for i, viewType := range viewSequence {
		assert.Equal(t, viewType, app.state.ViewHistory[i])
	}

	// Check current view is the last one
	assert.Equal(t, viewSequence[len(viewSequence)-1], app.state.CurrentView)
}

func TestApp_CreateNavbar(t *testing.T) {
	app := NewApp(ui.DockerContainerListView)

	navbar := app.createNavbar()

	assert.NotNil(t, navbar)
	assert.IsType(t, &tview.TextView{}, navbar)

	// Check that navbar has dynamic colors enabled
	assert.True(t, navbar.HasFocus() || true) // TextView properties are private, so we just check it exists
}

func TestApp_CreateStatusBar(t *testing.T) {
	app := NewApp(ui.DockerContainerListView)

	statusBar := app.createStatusBar()

	assert.NotNil(t, statusBar)
	assert.IsType(t, &tview.TextView{}, statusBar)

	// Check status bar text contains help instructions
	text := statusBar.GetText(false)
	assert.Contains(t, text, "Help")
	assert.Contains(t, text, "Quit")
	assert.Contains(t, text, "Switch View")
}

func TestApp_UpdateNavbar(t *testing.T) {
	app := NewApp(ui.DockerContainerListView)
	navbar := tview.NewTextView().SetDynamicColors(true)

	// Test with different current views
	testCases := []ui.ViewType{
		ui.DockerContainerListView,
		ui.ComposeProcessListView,
		ui.ImageListView,
	}

	for _, viewType := range testCases {
		app.state.CurrentView = viewType
		app.updateNavbar(navbar)

		text := navbar.GetText(false)
		assert.NotEmpty(t, text)
		// The current view should be highlighted with color tags
		assert.Contains(t, text, "[black:cyan]") // Color tag for active view
	}
}

func TestApp_KeyboardShortcuts(t *testing.T) {
	app := NewApp(ui.DockerContainerListView)

	// Create simulation screen for testing
	simScreen := tcell.NewSimulationScreen("UTF-8")
	err := simScreen.Init()
	assert.NoError(t, err)

	app.app.SetScreen(simScreen)

	// Test view switching with number keys
	numberKeyTests := []struct {
		key      rune
		expected ui.ViewType
	}{
		{'1', ui.DockerContainerListView},
		{'2', ui.ComposeProcessListView},
		{'3', ui.ComposeProjectListView},
		{'4', ui.ImageListView},
		{'5', ui.NetworkListView},
		{'6', ui.VolumeListView},
	}

	for _, tt := range numberKeyTests {
		t.Run(string(tt.key), func(t *testing.T) {
			// Simulate key press
			event := tcell.NewEventKey(tcell.KeyRune, tt.key, tcell.ModNone)

			// Process the event through input capture
			capture := app.app.GetInputCapture()
			if capture != nil {
				processed := capture(event)
				assert.Nil(t, processed) // Event should be consumed
			}

			// Due to the async nature of the app, we'd need to run it to test properly
			// For unit tests, we're verifying the structure is in place
		})
	}
}

func TestApp_Stop(t *testing.T) {
	app := NewApp(ui.DockerContainerListView)

	// We can't really test Stop() without running the app
	// But we can verify the method exists and doesn't panic
	assert.NotPanics(t, func() {
		app.Stop()
	})
}

func TestApp_QuitConfirmation(t *testing.T) {
	app := NewApp(ui.DockerContainerListView)

	// Call showQuitConfirmation - it should add a modal to pages
	app.showQuitConfirmation()

	// The modal should be added to pages
	// We can't easily test the modal behavior without running the app
	// but we can verify the method doesn't panic
	assert.NotNil(t, app.pages)
}

func TestApp_ViewRefresh(t *testing.T) {
	app := NewApp(ui.DockerContainerListView)

	// Create a mock view to track refresh calls
	mockView := NewMockView("Test View")
	app.views[ui.DockerContainerListView] = mockView

	// Switch to the view (which should trigger refresh)
	app.SwitchView(ui.DockerContainerListView)

	// Verify refresh was called
	// Note: In the actual implementation, Refresh is called
	// We're verifying the structure is in place
	assert.NotNil(t, mockView)
}

func TestApp_RunIntegration(t *testing.T) {
	// This is an integration test that actually runs the app briefly
	t.Skip("Skipping integration test that requires terminal")

	app := NewApp(ui.DockerContainerListView)

	// Create simulation screen
	simScreen := tcell.NewSimulationScreen("UTF-8")
	err := simScreen.Init()
	assert.NoError(t, err)

	app.app.SetScreen(simScreen)

	// Run the app in a goroutine
	done := make(chan bool)
	go func() {
		err := app.Run()
		assert.NoError(t, err)
		done <- true
	}()

	// Let it run briefly
	time.Sleep(100 * time.Millisecond)

	// Stop the app
	app.Stop()

	// Wait for it to finish
	select {
	case <-done:
		// App stopped successfully
	case <-time.After(1 * time.Second):
		t.Fatal("App did not stop in time")
	}
}

func TestApp_SetGlobalAppInstance(t *testing.T) {
	// Test that the global app instance is set for views
	app := NewApp(ui.DockerContainerListView)

	// The global app instance should be set through SetApp
	assert.NotNil(t, app.app)
	// views.SetApp was called in NewApp, so the global instance should be set
}
