package views

import (
	"github.com/rivo/tview"

	"github.com/tokuhirom/dcv/internal/docker"
)

// StatsView displays container statistics
type StatsView struct {
	docker *docker.Client
	table  *tview.Table
}

// NewStatsView creates a new stats view
func NewStatsView(dockerClient *docker.Client) *StatsView {
	return &StatsView{
		docker: dockerClient,
		table:  tview.NewTable(),
	}
}

func (v *StatsView) GetPrimitive() tview.Primitive {
	return v.table
}

func (v *StatsView) Refresh() {
	// TODO: Implement
}

func (v *StatsView) GetTitle() string {
	return "Container Statistics"
}
