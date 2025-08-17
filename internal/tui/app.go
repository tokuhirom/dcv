package tui

import (
	"fmt"
	"log/slog"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/tui/views"
	"github.com/tokuhirom/dcv/internal/ui"
)

// App represents the main tview application
type App struct {
	app         *tview.Application
	pages       *tview.Pages
	state       *State
	docker      *docker.Client
	views       map[ui.ViewType]views.View
	initialView ui.ViewType
}

// NewApp creates a new tview application
func NewApp(initialView ui.ViewType) *App {
	app := &App{
		app:         tview.NewApplication(),
		pages:       tview.NewPages(),
		views:       make(map[ui.ViewType]views.View),
		initialView: initialView,
	}

	// Set the global app instance for views
	views.SetApp(app.app)

	// Initialize Docker client
	app.docker = docker.NewClient()

	// Initialize state
	app.state = NewState()
	app.state.CurrentView = initialView

	// Initialize views
	app.initializeViews()

	// Setup the main layout
	app.setupLayout()

	// Setup global key handlers
	app.setupGlobalKeys()

	return app
}

// Run starts the tview application
func (a *App) Run() error {
	// Switch to initial view
	a.SwitchView(a.initialView)

	return a.app.Run()
}

// Stop stops the application
func (a *App) Stop() {
	a.app.Stop()
}

// initializeViews creates all the views
func (a *App) initializeViews() {
	// Create Docker Container List view
	dockerView := views.NewDockerContainerListView(a.docker)
	a.views[ui.DockerContainerListView] = dockerView
	a.pages.AddPage("docker", dockerView.GetPrimitive(), true, false)

	// Create Compose Process List view
	composeView := views.NewComposeProcessListView(a.docker)
	a.views[ui.ComposeProcessListView] = composeView
	a.pages.AddPage("compose", composeView.GetPrimitive(), true, false)

	// Create Project List view
	projectView := views.NewProjectListView(a.docker)
	a.views[ui.ComposeProjectListView] = projectView
	a.pages.AddPage("projects", projectView.GetPrimitive(), true, false)

	// Create Image List view
	imageView := views.NewImageListView(a.docker)
	a.views[ui.ImageListView] = imageView
	a.pages.AddPage("images", imageView.GetPrimitive(), true, false)

	// Create Network List view
	networkView := views.NewNetworkListView(a.docker)
	a.views[ui.NetworkListView] = networkView
	a.pages.AddPage("networks", networkView.GetPrimitive(), true, false)

	// Create Volume List view
	volumeView := views.NewVolumeListView(a.docker)
	a.views[ui.VolumeListView] = volumeView
	a.pages.AddPage("volumes", volumeView.GetPrimitive(), true, false)

	// Create Log view
	logView := views.NewLogView(a.docker)
	a.views[ui.LogView] = logView
	a.pages.AddPage("logs", logView.GetPrimitive(), true, false)

	// Create Stats view
	statsView := views.NewStatsView(a.docker)
	a.views[ui.StatsView] = statsView
	a.pages.AddPage("stats", statsView.GetPrimitive(), true, false)

	// Create Help view
	helpView := views.NewHelpView()
	a.views[ui.HelpView] = helpView
	a.pages.AddPage("help", helpView.GetPrimitive(), true, false)
}

// setupLayout creates the main layout with navbar and status bar
func (a *App) setupLayout() {
	// Create navbar
	navbar := a.createNavbar()

	// Create status bar
	statusBar := a.createStatusBar()

	// Create main layout
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(navbar, 1, 0, false).
		AddItem(a.pages, 0, 1, true).
		AddItem(statusBar, 1, 0, false)

	a.app.SetRoot(flex, true)
}

// createNavbar creates the navigation bar
func (a *App) createNavbar() *tview.TextView {
	navbar := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

	a.updateNavbar(navbar)
	return navbar
}

// createStatusBar creates the status bar
func (a *App) createStatusBar() *tview.TextView {
	statusBar := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)

	statusBar.SetText(" [yellow]?[white] Help | [yellow]q[white] Quit | [yellow]1-9[white] Switch View")
	return statusBar
}

// updateNavbar updates the navbar based on current view
func (a *App) updateNavbar(navbar *tview.TextView) {
	viewNames := map[ui.ViewType]string{
		ui.DockerContainerListView: "Docker",
		ui.ComposeProcessListView:  "Compose",
		ui.ComposeProjectListView:  "Projects",
		ui.ImageListView:           "Images",
		ui.NetworkListView:         "Networks",
		ui.VolumeListView:          "Volumes",
		ui.LogView:                 "Logs",
		ui.StatsView:               "Stats",
		ui.HelpView:                "Help",
	}

	var text string
	for viewType, name := range viewNames {
		if viewType == a.state.CurrentView {
			text += fmt.Sprintf("[black:cyan] %s [-:-] ", name)
		} else {
			text += fmt.Sprintf(" %s  ", name)
		}
	}

	navbar.SetText(text)
}

// setupGlobalKeys sets up global keyboard shortcuts
func (a *App) setupGlobalKeys() {
	a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Handle view switching (1-9 keys)
		if event.Rune() >= '1' && event.Rune() <= '9' {
			viewIndex := int(event.Rune() - '1')
			viewTypes := []ui.ViewType{
				ui.DockerContainerListView,
				ui.ComposeProcessListView,
				ui.ComposeProjectListView,
				ui.ImageListView,
				ui.NetworkListView,
				ui.VolumeListView,
			}
			if viewIndex < len(viewTypes) {
				a.SwitchView(viewTypes[viewIndex])
				return nil
			}
		}

		switch event.Key() {
		case tcell.KeyEscape:
			// Go back to previous view
			if len(a.state.ViewHistory) > 1 {
				a.state.ViewHistory = a.state.ViewHistory[:len(a.state.ViewHistory)-1]
				previousView := a.state.ViewHistory[len(a.state.ViewHistory)-1]
				a.SwitchView(previousView)
				return nil
			}
		case tcell.KeyCtrlC:
			a.Stop()
			return nil
		}

		switch event.Rune() {
		case 'q', 'Q':
			// Show quit confirmation
			a.showQuitConfirmation()
			return nil
		case '?':
			// Show help
			a.SwitchView(ui.HelpView)
			return nil
		case 'r', 'R':
			// Refresh current view
			if view, ok := a.views[a.state.CurrentView]; ok {
				view.Refresh()
			}
			return nil
		}

		return event
	})
}

// SwitchView switches to a different view
func (a *App) SwitchView(viewType ui.ViewType) {
	slog.Info("Switching view", slog.String("view", viewType.String()))

	a.state.CurrentView = viewType
	a.state.ViewHistory = append(a.state.ViewHistory, viewType)

	// Update navbar
	if navbar := a.app.GetFocus(); navbar != nil {
		if tv, ok := navbar.(*tview.TextView); ok {
			a.updateNavbar(tv)
		}
	}

	// Switch page based on view type
	switch viewType {
	case ui.DockerContainerListView:
		a.pages.SwitchToPage("docker")
		if view, ok := a.views[viewType]; ok {
			view.Refresh()
		}
	case ui.ComposeProcessListView:
		a.pages.SwitchToPage("compose")
		if view, ok := a.views[viewType]; ok {
			view.Refresh()
		}
	case ui.ComposeProjectListView:
		a.pages.SwitchToPage("projects")
		if view, ok := a.views[viewType]; ok {
			view.Refresh()
		}
	case ui.ImageListView:
		a.pages.SwitchToPage("images")
		if view, ok := a.views[viewType]; ok {
			view.Refresh()
		}
	case ui.NetworkListView:
		a.pages.SwitchToPage("networks")
		if view, ok := a.views[viewType]; ok {
			view.Refresh()
		}
	case ui.VolumeListView:
		a.pages.SwitchToPage("volumes")
		if view, ok := a.views[viewType]; ok {
			view.Refresh()
		}
	case ui.LogView:
		a.pages.SwitchToPage("logs")
	case ui.StatsView:
		a.pages.SwitchToPage("stats")
		if view, ok := a.views[viewType]; ok {
			view.Refresh()
		}
	case ui.HelpView:
		a.pages.SwitchToPage("help")
	}

	a.app.ForceDraw()
}

// showQuitConfirmation shows a confirmation dialog for quitting
func (a *App) showQuitConfirmation() {
	modal := tview.NewModal().
		SetText("Are you sure you want to quit?").
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Yes" {
				a.Stop()
			}
			a.pages.RemovePage("quit-confirm")
		})

	a.pages.AddPage("quit-confirm", modal, true, true)
}

// GetApp returns the tview application instance
func (a *App) GetApp() *tview.Application {
	return a.app
}

// GetDocker returns the Docker client
func (a *App) GetDocker() *docker.Client {
	return a.docker
}

// GetState returns the application state
func (a *App) GetState() *State {
	return a.state
}
