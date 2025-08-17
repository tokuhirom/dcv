package views

import "github.com/rivo/tview"

// appInstance holds a reference to the main tview application for QueueUpdateDraw
var appInstance *tview.Application

// SetApp sets the global application instance
func SetApp(app *tview.Application) {
	appInstance = app
}

// QueueUpdateDraw queues an update draw on the application
func QueueUpdateDraw(f func()) {
	if appInstance != nil {
		appInstance.QueueUpdateDraw(f)
	}
}
