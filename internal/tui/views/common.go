package views

import (
	"sync"

	"github.com/rivo/tview"
)

var (
	// appInstance holds a reference to the main tview application for QueueUpdateDraw
	appInstance *tview.Application
	// mu protects access to appInstance
	mu sync.RWMutex
)

// SetApp sets the global application instance
func SetApp(app *tview.Application) {
	mu.Lock()
	defer mu.Unlock()
	appInstance = app
}

// QueueUpdateDraw queues an update draw on the application
func QueueUpdateDraw(f func()) {
	mu.RLock()
	app := appInstance
	mu.RUnlock()

	if app != nil {
		app.QueueUpdateDraw(f)
	}
}
