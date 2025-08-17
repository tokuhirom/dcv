package views

import (
	"github.com/rivo/tview"

	"github.com/tokuhirom/dcv/internal/docker"
)

// VolumeListView displays Docker volumes
type VolumeListView struct {
	docker *docker.Client
	table  *tview.Table
}

// NewVolumeListView creates a new volume list view
func NewVolumeListView(dockerClient *docker.Client) *VolumeListView {
	return &VolumeListView{
		docker: dockerClient,
		table:  tview.NewTable(),
	}
}

func (v *VolumeListView) GetPrimitive() tview.Primitive {
	return v.table
}

func (v *VolumeListView) Refresh() {
	// TODO: Implement
}

func (v *VolumeListView) GetTitle() string {
	return "Docker Volumes"
}
