package views

import (
	"github.com/rivo/tview"
	"github.com/tokuhirom/dcv/internal/docker"
)

// ComposeProcessListView displays Docker Compose processes
type ComposeProcessListView struct {
	docker *docker.Client
	table  *tview.Table
}

// NewComposeProcessListView creates a new compose process list view
func NewComposeProcessListView(dockerClient *docker.Client) *ComposeProcessListView {
	return &ComposeProcessListView{
		docker: dockerClient,
		table:  tview.NewTable(),
	}
}

func (v *ComposeProcessListView) GetPrimitive() tview.Primitive {
	return v.table
}

func (v *ComposeProcessListView) Refresh() {
	// TODO: Implement
}

func (v *ComposeProcessListView) GetTitle() string {
	return "Compose Processes"
}