package views

import (
	"github.com/rivo/tview"
	"github.com/tokuhirom/dcv/internal/docker"
)

// ProjectListView displays Docker Compose projects
type ProjectListView struct {
	docker *docker.Client
	table  *tview.Table
}

// NewProjectListView creates a new project list view
func NewProjectListView(dockerClient *docker.Client) *ProjectListView {
	return &ProjectListView{
		docker: dockerClient,
		table:  tview.NewTable(),
	}
}

func (v *ProjectListView) GetPrimitive() tview.Primitive {
	return v.table
}

func (v *ProjectListView) Refresh() {
	// TODO: Implement
}

func (v *ProjectListView) GetTitle() string {
	return "Docker Compose Projects"
}