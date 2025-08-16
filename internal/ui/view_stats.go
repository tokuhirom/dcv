package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tokuhirom/dcv/internal/models"
)

// statsLoadedMsg contains the loaded container stats
type statsLoadedMsg struct {
	stats []models.ContainerStats
	err   error
}

// TODO: support compose-stats
// TODO: stream support

// StatsSortField represents the field to sort container stats by
type StatsSortField int

const (
	StatsSortByName StatsSortField = iota
	StatsSortByCPU
	StatsSortByMem
)

func (s StatsSortField) String() string {
	switch s {
	case StatsSortByName:
		return "NAME"
	case StatsSortByCPU:
		return "CPU%"
	case StatsSortByMem:
		return "MEM%"
	default:
		return "NAME"
	}
}

// StatsViewModel manages the state and rendering of the stats view
type StatsViewModel struct {
	TableViewModel
	stats           []models.ContainerStats
	sortField       StatsSortField
	sortReverse     bool
	autoRefresh     bool
	refreshInterval time.Duration
}

// Update handles messages for the stats view
func (m *StatsViewModel) Update(model *Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case statsLoadedMsg:
		model.loading = false
		if msg.err != nil {
			model.err = msg.err
			return model, nil
		} else {
			model.err = nil
		}

		m.Loaded(model, msg.stats)
		return model, nil
	default:
		return model, nil
	}
}

// render renders the stats view
func (m *StatsViewModel) render(model *Model, availableHeight int) string {
	if len(m.stats) == 0 {
		return "\nNo stats available.\n"
	}

	// Display sorting info and auto-refresh status
	var s strings.Builder
	sortInfo := fmt.Sprintf("Sort: %s", m.sortField.String())
	if m.sortReverse {
		sortInfo += " (desc)"
	} else {
		sortInfo += " (asc)"
	}
	s.WriteString(helpStyle.Render(sortInfo))
	s.WriteString("  ")

	// Show auto-refresh status
	if m.autoRefresh {
		refreshStatus := fmt.Sprintf("Auto-refresh: ON (%ds)", int(m.refreshInterval.Seconds()))
		s.WriteString(searchStyle.Render(refreshStatus))
	} else {
		s.WriteString(helpStyle.Render("Auto-refresh: OFF"))
	}
	s.WriteString("  ")

	s.WriteString(helpStyle.Render("[c]PU [m]EM [n]ame [R]everse [a]uto-refresh"))
	s.WriteString("\n\n")

	// Sort stats
	m.sortStats()

	// Stats table
	columns := []table.Column{
		{Title: "NAME", Width: 20},
		{Title: "CPU %", Width: 10},
		{Title: "MEM USAGE", Width: 15},
		{Title: "MEM %", Width: 10},
		{Title: "NET I/O", Width: 15},
		{Title: "BLOCK I/O", Width: 15},
	}

	// Highlight the current sort field header
	highlightStyle := searchStyle
	switch m.sortField {
	case StatsSortByName:
		columns[0].Title = highlightStyle.Render("NAME")
	case StatsSortByCPU:
		columns[1].Title = highlightStyle.Render("CPU %")
	case StatsSortByMem:
		columns[3].Title = highlightStyle.Render("MEM %")
	}

	// Build rows for the TableViewModel
	m.buildRows()

	// Use RenderTable for consistent table rendering
	return s.String() + m.RenderTable(model, columns, availableHeight-3, func(row, col int) lipgloss.Style {
		if row == m.Cursor {
			return tableSelectedCellStyle
		}
		return tableNormalCellStyle
	})
}

// Show switches to the stats view
func (m *StatsViewModel) Show(model *Model) tea.Cmd {
	m.autoRefresh = true                // Enable auto-refresh by default
	m.refreshInterval = 2 * time.Second // Default refresh interval
	model.SwitchView(StatsView)
	return tea.Batch(
		m.DoLoad(model),
		m.startAutoRefresh(),
	)
}

// DoLoad reloads the stats
func (m *StatsViewModel) DoLoad(model *Model) tea.Cmd {
	model.loading = true
	return m.doLoadInternal(model)
}

// DoLoadSilent reloads the stats without showing loading indicator
func (m *StatsViewModel) DoLoadSilent(model *Model) tea.Cmd {
	// Don't set loading = true for silent refresh
	return m.doLoadInternal(model)
}

func (m *StatsViewModel) doLoadInternal(model *Model) tea.Cmd {
	return func() tea.Msg {
		// TODO: suppport toggle-all stats
		stats, err := model.dockerClient.GetStats(false)
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
func (m *StatsViewModel) Loaded(model *Model, stats []models.ContainerStats) {
	m.stats = stats
	m.buildRows()
	m.SetRows(m.Rows, model.ViewHeight())
}

func (m *StatsViewModel) sortStats() {
	sort.Slice(m.stats, func(i, j int) bool {
		var less bool
		switch m.sortField {
		case StatsSortByName:
			less = m.stats[i].Name < m.stats[j].Name
		case StatsSortByCPU:
			cpu1 := models.ParsePercentage(m.stats[i].CPUPerc)
			cpu2 := models.ParsePercentage(m.stats[j].CPUPerc)
			less = cpu1 < cpu2
		case StatsSortByMem:
			mem1 := models.ParsePercentage(m.stats[i].MemPerc)
			mem2 := models.ParsePercentage(m.stats[j].MemPerc)
			less = mem1 < mem2
		default:
			less = m.stats[i].Name < m.stats[j].Name
		}

		if m.sortReverse {
			return !less
		}
		return less
	})
}

// buildRows builds the table rows for the stats
func (m *StatsViewModel) buildRows() {
	// Sort stats first
	m.sortStats()

	rows := make([]table.Row, 0, len(m.stats))
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

	m.Rows = rows
}

// HandleUp scrolls up in the stats list
func (m *StatsViewModel) HandleUp(model *Model) tea.Cmd {
	return m.TableViewModel.HandleUp(model)
}

// HandleDown scrolls down in the stats list
func (m *StatsViewModel) HandleDown(model *Model) tea.Cmd {
	return m.TableViewModel.HandleDown(model)
}

// HandleSortByCPU sorts containers by CPU usage
func (m *StatsViewModel) HandleSortByCPU(model *Model) {
	if m.sortField == StatsSortByCPU {
		m.sortReverse = !m.sortReverse
	} else {
		m.sortField = StatsSortByCPU
		m.sortReverse = true // Default to descending for CPU
	}
	m.buildRows()
	m.SetRows(m.Rows, model.ViewHeight())
}

// HandleSortByMem sorts containers by memory usage
func (m *StatsViewModel) HandleSortByMem(model *Model) {
	if m.sortField == StatsSortByMem {
		m.sortReverse = !m.sortReverse
	} else {
		m.sortField = StatsSortByMem
		m.sortReverse = true // Default to descending for memory
	}
	m.buildRows()
	m.SetRows(m.Rows, model.ViewHeight())
}

// HandleSortByName sorts containers by name
func (m *StatsViewModel) HandleSortByName(model *Model) {
	if m.sortField == StatsSortByName {
		m.sortReverse = !m.sortReverse
	} else {
		m.sortField = StatsSortByName
		m.sortReverse = false // Default to ascending for name
	}
	m.buildRows()
	m.SetRows(m.Rows, model.ViewHeight())
}

// HandleReverseSort reverses the current sort order
func (m *StatsViewModel) HandleReverseSort(model *Model) {
	m.sortReverse = !m.sortReverse
	m.buildRows()
	m.SetRows(m.Rows, model.ViewHeight())
}

// HandleToggleAutoRefresh toggles the auto-refresh feature
func (m *StatsViewModel) HandleToggleAutoRefresh() {
	m.autoRefresh = !m.autoRefresh
}

// startAutoRefresh returns a command to trigger periodic refresh
func (m *StatsViewModel) startAutoRefresh() tea.Cmd {
	return tea.Tick(m.refreshInterval, func(time.Time) tea.Msg {
		return autoRefreshTickMsg{}
	})
}
