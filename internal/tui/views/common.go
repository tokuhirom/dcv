package views

import (
	"sync"

	"github.com/gdamore/tcell/v2"
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
	} else {
		// In test mode or when app is not running, execute directly
		f()
	}
}

// CreateConfirmationModal creates a confirmation modal with y/n keyboard shortcuts
func CreateConfirmationModal(text string, onYes, onNo func()) *tview.Modal {
	modal := tview.NewModal().
		SetText(text).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Yes" && onYes != nil {
				onYes()
			} else if onNo != nil {
				onNo()
			}
		})

	// Add keyboard shortcuts for y/n
	modal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'y', 'Y':
			// Yes - execute the onYes callback
			if onYes != nil {
				onYes()
			}
			return nil
		case 'n', 'N':
			// No - execute the onNo callback
			if onNo != nil {
				onNo()
			}
			return nil
		}
		return event
	})

	return modal
}
