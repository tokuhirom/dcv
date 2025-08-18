package tui

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
	"github.com/tokuhirom/dcv/internal/tui/views"
	"github.com/tokuhirom/dcv/internal/ui"
)

// App represents the main tview application
type App struct {
	app          *tview.Application
	pages        *tview.Pages
	state        *State
	docker       *docker.Client
	views        map[ui.ViewType]views.View
	initialView  ui.ViewType
	navbar       *tview.TextView
	navbarHidden bool
	layout       *tview.Flex
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
	dockerView.SetSwitchToLogViewCallback(a.SwitchToLogView)
	dockerView.SetSwitchToFileBrowserCallback(a.SwitchToFileBrowser)
	dockerView.SetSwitchToInspectViewCallback(a.SwitchToInspectView)
	a.views[ui.DockerContainerListView] = dockerView
	a.pages.AddPage("docker", dockerView.GetPrimitive(), true, false)

	// Create Compose Process List view
	composeView := views.NewComposeProcessListView(a.docker)
	composeView.SetSwitchToLogViewCallback(a.SwitchToLogView)
	composeView.SetSwitchToFileBrowserCallback(a.SwitchToFileBrowser)
	composeView.SetSwitchToInspectViewCallback(a.SwitchToInspectView)
	a.views[ui.ComposeProcessListView] = composeView
	a.pages.AddPage("compose", composeView.GetPrimitive(), true, false)

	// Create Project List view
	projectView := views.NewProjectListView(a.docker)
	projectView.SetOnProjectSelected(a.SwitchToComposeProcessList)
	a.views[ui.ComposeProjectListView] = projectView
	a.pages.AddPage("projects", projectView.GetPrimitive(), true, false)

	// Create Image List view
	imageView := views.NewImageListView(a.docker)
	imageView.SetSwitchToInspectViewCallback(a.SwitchToInspectView)
	a.views[ui.ImageListView] = imageView
	a.pages.AddPage("images", imageView.GetPrimitive(), true, false)

	// Create Network List view
	networkView := views.NewNetworkListView(a.docker)
	networkView.SetSwitchToInspectViewCallback(a.SwitchToInspectView)
	a.views[ui.NetworkListView] = networkView
	a.pages.AddPage("networks", networkView.GetPrimitive(), true, false)

	// Create Volume List view
	volumeView := views.NewVolumeListView(a.docker)
	volumeView.SetSwitchToInspectViewCallback(a.SwitchToInspectView)
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

	// Create File Browser view
	fileBrowserView := views.NewFileBrowserView(a.docker)
	a.views[ui.FileBrowserView] = fileBrowserView
	a.pages.AddPage("file-browser", fileBrowserView.GetPrimitive(), true, false)

	// Create File Content view
	fileContentView := views.NewFileContentView(a.docker)
	a.views[ui.FileContentView] = fileContentView
	a.pages.AddPage("file-content", fileContentView.GetPrimitive(), true, false)

	// Create Inspect view
	inspectView := views.NewInspectView(a.docker)
	a.views[ui.InspectView] = inspectView
	a.pages.AddPage("inspect", inspectView.GetPrimitive(), true, false)

	// Set callbacks for file browser
	fileBrowserView.SetSwitchToContentViewCallback(a.SwitchToFileContent)
}

// setupLayout creates the main layout with navbar and status bar
func (a *App) setupLayout() {
	// Create navbar
	a.navbar = a.createNavbar()

	// Create status bar
	statusBar := a.createStatusBar()

	// Create main layout
	a.layout = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.navbar, 1, 0, false).
		AddItem(a.pages, 0, 1, true).
		AddItem(statusBar, 1, 0, false)

	a.app.SetRoot(a.layout, true)
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

	statusBar.SetText(" [yellow]?[white] Help | [yellow]x[white] Actions/Execute | [yellow]q[white] Quit | [yellow]1-9[white] Switch View")
	return statusBar
}

// updateNavbar updates the navbar based on current view
func (a *App) updateNavbar(navbar *tview.TextView) {
	// Define views with their numbers and names
	type navItem struct {
		viewType ui.ViewType
		number   string
		name     string
	}

	items := []navItem{
		{ui.DockerContainerListView, "1", "Containers"},
		{ui.ComposeProjectListView, "2", "Projects"},
		{ui.ImageListView, "3", "Images"},
		{ui.NetworkListView, "4", "Networks"},
		{ui.VolumeListView, "5", "Volumes"},
		{ui.StatsView, "6", "Stats"},
	}

	var parts []string
	for _, item := range items {
		// Format: [number] name
		itemText := fmt.Sprintf("[%s] %s", item.number, item.name)

		// Highlight current view
		if item.viewType == a.state.CurrentView {
			// White text on cyan background for current view
			parts = append(parts, fmt.Sprintf("[black:cyan] %s [-:-]", itemText))
		} else if item.viewType == ui.ComposeProjectListView && a.state.CurrentView == ui.ComposeProcessListView {
			// Also highlight Projects when in Compose Process view
			parts = append(parts, fmt.Sprintf("[black:cyan] %s [-:-]", itemText))
		} else {
			// Normal gray text for inactive views
			parts = append(parts, fmt.Sprintf("[gray]%s[-]", itemText))
		}
	}

	// Join with separator
	text := strings.Join(parts, " [darkgray]|[-] ")

	// Add help hint at the end
	if a.navbarHidden {
		text += "  [darkgray]|[-] [yellow][H][-] Show navbar"
	} else {
		text += "  [darkgray]|[-] [yellow][H][-] Hide navbar"
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
				ui.DockerContainerListView, // 1
				ui.ComposeProjectListView,  // 2
				ui.ImageListView,           // 3
				ui.NetworkListView,         // 4
				ui.VolumeListView,          // 5
				ui.StatsView,               // 6
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
		case 'h', 'H':
			// Toggle navbar visibility
			a.toggleNavbar()
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
	case ui.FileBrowserView:
		a.pages.SwitchToPage("file-browser")
		if view, ok := a.views[viewType]; ok {
			view.Refresh()
		}
	case ui.FileContentView:
		a.pages.SwitchToPage("file-content")
	case ui.InspectView:
		a.pages.SwitchToPage("inspect")
		if view, ok := a.views[viewType]; ok {
			view.Refresh()
		}
	}

	a.app.ForceDraw()
}

// SwitchToLogView switches to the log view with a specific container
func (a *App) SwitchToLogView(containerID string, container interface{}) {
	// Set the container in the log view
	if logView, ok := a.views[ui.LogView]; ok {
		if lv, ok := logView.(*views.LogView); ok {
			lv.SetContainer(containerID, container)
		}
	}

	// Switch to the log view
	a.SwitchView(ui.LogView)
}

// SwitchToComposeProcessList switches to the compose process list view with a specific project
func (a *App) SwitchToComposeProcessList(project models.ComposeProject) {
	// Set the project in the compose process list view
	if composeView, ok := a.views[ui.ComposeProcessListView]; ok {
		if cv, ok := composeView.(*views.ComposeProcessListView); ok {
			cv.SetProject(project.Name)
		}
	}

	// Switch to the compose process list view
	a.SwitchView(ui.ComposeProcessListView)
}

// SwitchToFileBrowser switches to the file browser view with a specific container
func (a *App) SwitchToFileBrowser(containerID string, container interface{}) {
	// Set the container in the file browser view
	if fileBrowserView, ok := a.views[ui.FileBrowserView]; ok {
		if fbv, ok := fileBrowserView.(*views.FileBrowserView); ok {
			containerName := ""
			// Extract container name based on type
			switch c := container.(type) {
			case models.DockerContainer:
				containerName = c.Names
			case models.ComposeContainer:
				containerName = c.Name
			}
			fbv.SetContainer(containerID, containerName)
		}
	}

	// Switch to the file browser view
	a.SwitchView(ui.FileBrowserView)
}

// SwitchToFileContent switches to the file content view with a specific file
func (a *App) SwitchToFileContent(containerID, path string, file models.ContainerFile) {
	// Set the file in the file content view
	if fileContentView, ok := a.views[ui.FileContentView]; ok {
		if fcv, ok := fileContentView.(*views.FileContentView); ok {
			// Get container name from file browser
			containerName := ""
			if fileBrowserView, ok := a.views[ui.FileBrowserView]; ok {
				if fbv, ok := fileBrowserView.(*views.FileBrowserView); ok {
					containerName = fbv.GetContainerName()
				}
			}
			fcv.SetFile(containerID, containerName, path, file)
		}
	}

	// Switch to the file content view
	a.SwitchView(ui.FileContentView)
}

// SwitchToInspectView switches to the inspect view with a specific target
func (a *App) SwitchToInspectView(targetID string, target interface{}) {
	// Set the target in the inspect view
	if inspectView, ok := a.views[ui.InspectView]; ok {
		if iv, ok := inspectView.(*views.InspectView); ok {
			targetName := ""
			targetType := ""
			// Extract target name and type based on the actual type
			switch t := target.(type) {
			case models.DockerContainer:
				targetName = t.Names
				targetType = "container"
			case models.ComposeContainer:
				targetName = t.Name
				targetType = "container"
			case models.DockerVolume:
				targetName = t.Name
				targetType = "volume"
			case models.DockerNetwork:
				targetName = t.Name
				targetType = "network"
			case models.DockerImage:
				targetName = t.GetRepoTag()
				targetType = "image"
			}
			iv.SetTarget(targetID, targetName, targetType)
		}
	}

	// Switch to the inspect view
	a.SwitchView(ui.InspectView)
}

// toggleNavbar toggles the navbar visibility
func (a *App) toggleNavbar() {
	a.navbarHidden = !a.navbarHidden

	if a.navbarHidden {
		// Remove navbar from layout
		a.layout.RemoveItem(a.navbar)
	} else {
		// Add navbar back to layout at the top
		a.layout.Clear()
		a.layout.AddItem(a.navbar, 1, 0, false)
		a.layout.AddItem(a.pages, 0, 1, true)

		// Re-add status bar if it exists
		statusBar := a.createStatusBar()
		a.layout.AddItem(statusBar, 1, 0, false)
	}

	// Update navbar text
	a.updateNavbar(a.navbar)
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
