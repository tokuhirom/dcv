package views

import (
	"github.com/rivo/tview"
)

// HelpView displays help information
type HelpView struct {
	textView *tview.TextView
}

// NewHelpView creates a new help view
func NewHelpView() *HelpView {
	v := &HelpView{
		textView: tview.NewTextView(),
	}
	
	v.setupHelp()
	return v
}

func (v *HelpView) setupHelp() {
	helpText := `[yellow]Docker Container Viewer - Help[-]

[cyan]Global Keys:[-]
  [white]?[-]      Show this help
  [white]q[-]      Quit application
  [white]r[-]      Refresh current view
  [white]ESC[-]    Go back to previous view
  [white]1-9[-]    Switch between views

[cyan]Navigation:[-]
  [white]j/↓[-]    Move down
  [white]k/↑[-]    Move up
  [white]g[-]      Go to top
  [white]G[-]      Go to bottom
  [white]PgUp[-]   Page up
  [white]PgDn[-]   Page down

[cyan]Container Operations:[-]
  [white]s[-]      Stop container
  [white]S[-]      Start container
  [white]k[-]      Kill container
  [white]d[-]      Delete container
  [white]r[-]      Restart container
  [white]l[-]      View logs
  [white]e[-]      Execute shell
  [white]i[-]      Inspect container
  [white]a[-]      Toggle show all containers

[cyan]Views:[-]
  [white]1[-]      Docker Containers
  [white]2[-]      Compose Processes
  [white]3[-]      Projects
  [white]4[-]      Images
  [white]5[-]      Networks
  [white]6[-]      Volumes

Press [yellow]ESC[-] to go back`

	v.textView.SetText(helpText).
		SetDynamicColors(true).
		SetScrollable(true)
}

func (v *HelpView) GetPrimitive() tview.Primitive {
	return v.textView
}

func (v *HelpView) Refresh() {
	// Help doesn't need refresh
}

func (v *HelpView) GetTitle() string {
	return "Help"
}