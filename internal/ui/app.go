package ui

import (
	"github.com/rivo/tview"
	"github.com/tokuhirom/dcv/internal/docker"
)

type App struct {
	app           *tview.Application
	pages         *tview.Pages
	dockerClient  *docker.ComposeClient
	processView   *ProcessListView
	logView       *LogView
	dindView      *DindProcessListView
}

func NewApp() (*App, error) {
	app := &App{
		app:          tview.NewApplication(),
		pages:        tview.NewPages(),
		dockerClient: docker.NewComposeClient(""),
	}

	// Initialize views
	app.processView = NewProcessListView(app)
	app.logView = NewLogView(app)
	app.dindView = NewDindProcessListView(app)

	// Add pages
	app.pages.AddPage("processes", app.processView.view, true, true)
	app.pages.AddPage("logs", app.logView.view, true, false)
	app.pages.AddPage("dind", app.dindView.view, true, false)

	app.app.SetRoot(app.pages, true)

	return app, nil
}

func (a *App) Run() error {
	// Load initial data
	if err := a.processView.Refresh(); err != nil {
		return err
	}

	return a.app.Run()
}

func (a *App) ShowProcessList() {
	a.pages.SwitchToPage("processes")
}

func (a *App) ShowLogs(containerName string, isDind bool) {
	a.logView.SetContainer(containerName, isDind)
	a.pages.SwitchToPage("logs")
}

func (a *App) ShowDindProcessList(containerName string) {
	a.dindView.SetContainer(containerName)
	a.pages.SwitchToPage("dind")
}