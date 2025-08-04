package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// StatsViewModel manages the state and rendering of the stats view
type StatsViewModel struct {
	stats []ContainerStats
}

// render renders the stats view
func (m *StatsViewModel) render(model *Model, availableHeight int) string {
	var s strings.Builder

	if len(m.stats) == 0 {
		s.WriteString("\nNo stats available.\n")
		return s.String()
	}

	// Stats table

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return headerStyle
			}
			return normalStyle
		}).
		Headers("NAME", "CPU %", "MEM USAGE", "MEM %", "NET I/O", "BLOCK I/O")

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

		t.Row(name, cpu, stat.MemUsage, stat.MemPerc, stat.NetIO, stat.BlockIO)
	}

	s.WriteString(t.Render() + "\n")

	return s.String()
}

// Show switches to the stats view
func (m *StatsViewModel) Show(model *Model) tea.Cmd {
	model.currentView = StatsView
	model.loading = true
	return loadStats(model.dockerClient)
}

// HandleRefresh reloads the stats
func (m *StatsViewModel) HandleRefresh(model *Model) tea.Cmd {
	model.loading = true
	return loadStats(model.dockerClient)
}

// HandleBack returns to the compose process list view
func (m *StatsViewModel) HandleBack(model *Model) tea.Cmd {
	model.currentView = ComposeProcessListView
	return loadProcesses(model.dockerClient, model.projectName, model.composeProcessListViewModel.showAll)
}

// Loaded updates the stats list after loading
func (m *StatsViewModel) Loaded(stats []ContainerStats) {
	m.stats = stats
}
