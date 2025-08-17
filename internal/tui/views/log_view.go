package views

import (
	"github.com/rivo/tview"
	"github.com/tokuhirom/dcv/internal/docker"
)

// LogView displays container logs
type LogView struct {
	docker   *docker.Client
	textView *tview.TextView
}

// NewLogView creates a new log view
func NewLogView(dockerClient *docker.Client) *LogView {
	return &LogView{
		docker:   dockerClient,
		textView: tview.NewTextView(),
	}
}

func (v *LogView) GetPrimitive() tview.Primitive {
	return v.textView
}

func (v *LogView) Refresh() {
	// TODO: Implement
}

func (v *LogView) GetTitle() string {
	return "Container Logs"
}