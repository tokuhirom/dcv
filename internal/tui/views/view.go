package views

import (
	"github.com/rivo/tview"
)

// View represents a view in the application
type View interface {
	// GetPrimitive returns the tview primitive for this view
	GetPrimitive() tview.Primitive
	
	// Refresh refreshes the view's data
	Refresh()
	
	// GetTitle returns the title of the view
	GetTitle() string
}