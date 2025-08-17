package views

import (
	"github.com/rivo/tview"

	"github.com/tokuhirom/dcv/internal/docker"
)

// NetworkListView displays Docker networks
type NetworkListView struct {
	docker *docker.Client
	table  *tview.Table
}

// NewNetworkListView creates a new network list view
func NewNetworkListView(dockerClient *docker.Client) *NetworkListView {
	return &NetworkListView{
		docker: dockerClient,
		table:  tview.NewTable(),
	}
}

func (v *NetworkListView) GetPrimitive() tview.Primitive {
	return v.table
}

func (v *NetworkListView) Refresh() {
	// TODO: Implement
}

func (v *NetworkListView) GetTitle() string {
	return "Docker Networks"
}
