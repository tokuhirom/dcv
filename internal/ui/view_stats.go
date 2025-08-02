package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

func (m *Model) renderStatsView() string {
	var s strings.Builder

	title := titleStyle.Render("Container Resource Usage")
	s.WriteString(title + "\n")

	if m.loading {
		s.WriteString("\nLoading...\n")
		return s.String()
	}

	if m.err != nil {
		s.WriteString("\n" + errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n")
		return s.String()
	}

	if len(m.stats) == 0 {
		s.WriteString("\nNo stats available.\n")
		return s.String()
	}

	// Stats table
	s.WriteString("\n")

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
			if val, err := fmt.Sscanf(cpuVal, "%f", new(float64)); err == nil && val > 0 {
				if *new(float64) > 80.0 {
					cpu = errorStyle.Render(cpu)
				} else if *new(float64) > 50.0 {
					cpu = searchStyle.Render(cpu)
				}
			}
		}

		t.Row(name, cpu, stat.MemUsage, stat.MemPerc, stat.NetIO, stat.BlockIO)
	}

	s.WriteString(t.Render() + "\n\n")

	// Show help hint
	s.WriteString(helpStyle.Render("Press ? for help"))

	return s.String()
}
