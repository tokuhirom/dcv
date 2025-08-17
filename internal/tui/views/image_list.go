package views

import (
	"github.com/rivo/tview"
	"github.com/tokuhirom/dcv/internal/docker"
)

// ImageListView displays Docker images
type ImageListView struct {
	docker *docker.Client
	table  *tview.Table
}

// NewImageListView creates a new image list view
func NewImageListView(dockerClient *docker.Client) *ImageListView {
	return &ImageListView{
		docker: dockerClient,
		table:  tview.NewTable(),
	}
}

func (v *ImageListView) GetPrimitive() tview.Primitive {
	return v.table
}

func (v *ImageListView) Refresh() {
	// TODO: Implement
}

func (v *ImageListView) GetTitle() string {
	return "Docker Images"
}