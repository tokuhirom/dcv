package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/tokuhirom/dcv/internal/models"
)

// StatsViewModel manages the state and rendering of the stats view
type StatsViewModel struct {
	stats []models.ContainerStats
}

// render renders the stats view
func (m *StatsViewModel) render(model *Model, availableHeight int) string {
	if len(m.stats) == 0 {
		var s strings.Builder
		s.WriteString("\nNo stats available.\n")
		return s.String()
	}

	// Stats table
	columns := []table.Column{
		{Title: "NAME", Width: 20},
		{Title: "CPU %", Width: 10},
		{Title: "MEM USAGE", Width: 15},
		{Title: "MEM %", Width: 10},
		{Title: "NET I/O", Width: 15},
		{Title: "BLOCK I/O", Width: 15},
	}

	rows := []table.Row{}
	for _, stat := range m.stats {
		// Truncate name if too long
		name := stat.Name
		if len(name) > 20 {
			name = name[:17] + "..."
		}

		// Color CPU usage
		cpu := stat.CPUPerc
		if cpuVal := strings.TrimSuffix(cpu, "%"); cpuVal != "" {
			var cpuPercent float64
			if _, err := fmt.Sscanf(cpuVal, "%f", &cpuPercent); err == nil {
				if cpuPercent > 80.0 {
					cpu = errorStyle.Render(cpu)
				} else if cpuPercent > 50.0 {
					cpu = searchStyle.Render(cpu)
				}
			}
		}

		rows = append(rows, table.Row{name, cpu, stat.MemUsage, stat.MemPerc, stat.NetIO, stat.BlockIO})
	}

	return RenderUnfocusedTable(columns, rows, availableHeight)
}

// Show switches to the stats view
func (m *StatsViewModel) Show(model *Model) tea.Cmd {
	model.SwitchView(StatsView)
	return m.DoLoad(model)
}

// DoLoad reloads the stats
func (m *StatsViewModel) DoLoad(model *Model) tea.Cmd {
	model.loading = true
	return func() tea.Msg {
		stats, err := model.dockerClient.GetStats()
		return statsLoadedMsg{
			stats: stats,
			err:   err,
		}
	}
}

// HandleBack returns to the compose process list view
func (m *StatsViewModel) HandleBack(model *Model) tea.Cmd {
	model.SwitchToPreviousView()
	return nil
}

// Loaded updates the stats list after loading
func (m *StatsViewModel) Loaded(stats []models.ContainerStats) {
	m.stats = stats
}
