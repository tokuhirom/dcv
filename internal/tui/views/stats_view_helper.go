package views

import (
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/tokuhirom/dcv/internal/models"
)

// loadStats loads container statistics from Docker
func (v *StatsView) loadStats() {
	v.isRefreshing = true
	defer func() { v.isRefreshing = false }()

	slog.Info("Loading container stats", slog.Bool("showAll", v.showAll))

	stats, err := v.docker.GetStats(v.showAll)
	if err != nil {
		slog.Error("Failed to load container stats", slog.Any("error", err))
		return
	}

	v.stats = stats

	// Update table in UI thread
	QueueUpdateDraw(func() {
		v.updateTable()
	})
}

// updateTable updates the table with stats data
func (v *StatsView) updateTable() {
	v.table.Clear()

	// Sort stats before displaying
	v.sortStats()

	// Set headers
	headers := []string{"NAME", "CPU %", "MEM USAGE", "MEM %", "NET I/O", "BLOCK I/O", "PIDS"}
	for col, header := range headers {
		cell := tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetAttributes(tcell.AttrBold).
			SetSelectable(false)

		// Highlight the current sort field
		if (col == 0 && v.sortField == StatsSortByName) ||
			(col == 1 && v.sortField == StatsSortByCPU) ||
			(col == 3 && v.sortField == StatsSortByMem) ||
			(col == 4 && v.sortField == StatsSortByNetIO) ||
			(col == 5 && v.sortField == StatsSortByBlockIO) {
			cell.SetTextColor(tcell.ColorBlue)
			if v.sortReverse {
				cell.SetText(header + " ↓")
			} else {
				cell.SetText(header + " ↑")
			}
		}

		v.table.SetCell(0, col, cell)
	}

	// Add stats rows
	for row, stat := range v.stats {
		// Name
		name := stat.Name
		if len(name) > 20 {
			name = name[:17] + "..."
		}
		nameCell := tview.NewTableCell(name).
			SetTextColor(tcell.ColorWhite)
		v.table.SetCell(row+1, 0, nameCell)

		// CPU %
		cpuColor := tcell.ColorWhite
		cpuVal := parsePercentage(stat.CPUPerc)
		if cpuVal > 80.0 {
			cpuColor = tcell.ColorRed
		} else if cpuVal > 50.0 {
			cpuColor = tcell.ColorYellow
		} else if cpuVal > 20.0 {
			cpuColor = tcell.ColorGreen
		}
		cpuCell := tview.NewTableCell(stat.CPUPerc).
			SetTextColor(cpuColor)
		v.table.SetCell(row+1, 1, cpuCell)

		// Memory Usage
		memUsageCell := tview.NewTableCell(stat.MemUsage).
			SetTextColor(tcell.ColorWhite)
		v.table.SetCell(row+1, 2, memUsageCell)

		// Memory %
		memColor := tcell.ColorWhite
		memVal := parsePercentage(stat.MemPerc)
		if memVal > 80.0 {
			memColor = tcell.ColorRed
		} else if memVal > 50.0 {
			memColor = tcell.ColorYellow
		} else if memVal > 20.0 {
			memColor = tcell.ColorGreen
		}
		memPercCell := tview.NewTableCell(stat.MemPerc).
			SetTextColor(memColor)
		v.table.SetCell(row+1, 3, memPercCell)

		// Network I/O
		netIOCell := tview.NewTableCell(stat.NetIO).
			SetTextColor(tcell.ColorWhite)
		v.table.SetCell(row+1, 4, netIOCell)

		// Block I/O
		blockIOCell := tview.NewTableCell(stat.BlockIO).
			SetTextColor(tcell.ColorWhite)
		v.table.SetCell(row+1, 5, blockIOCell)

		// PIDs
		pidsCell := tview.NewTableCell(stat.PIDs).
			SetTextColor(tcell.ColorWhite)
		v.table.SetCell(row+1, 6, pidsCell)
	}

	// Select first row if available
	if len(v.stats) > 0 {
		v.table.Select(1, 0)
	}
}

// sortStats sorts the stats based on the current sort field and order
func (v *StatsView) sortStats() {
	sort.Slice(v.stats, func(i, j int) bool {
		var less bool
		switch v.sortField {
		case StatsSortByName:
			less = v.stats[i].Name < v.stats[j].Name
		case StatsSortByCPU:
			cpu1 := parsePercentage(v.stats[i].CPUPerc)
			cpu2 := parsePercentage(v.stats[j].CPUPerc)
			less = cpu1 < cpu2
		case StatsSortByMem:
			mem1 := parsePercentage(v.stats[i].MemPerc)
			mem2 := parsePercentage(v.stats[j].MemPerc)
			less = mem1 < mem2
		case StatsSortByNetIO:
			net1 := parseIOBytes(v.stats[i].NetIO)
			net2 := parseIOBytes(v.stats[j].NetIO)
			less = net1 < net2
		case StatsSortByBlockIO:
			block1 := parseIOBytes(v.stats[i].BlockIO)
			block2 := parseIOBytes(v.stats[j].BlockIO)
			less = block1 < block2
		default:
			less = v.stats[i].Name < v.stats[j].Name
		}

		if v.sortReverse {
			return !less
		}
		return less
	})
}

// setSortField sets the sort field and adjusts the sort order
func (v *StatsView) setSortField(field StatsSortField) {
	if v.sortField == field {
		// Toggle reverse if clicking same field
		v.sortReverse = !v.sortReverse
	} else {
		v.sortField = field
		// Default sort order based on field
		switch field {
		case StatsSortByName:
			v.sortReverse = false // Ascending for names
		default:
			v.sortReverse = true // Descending for numeric values
		}
	}
}

// toggleAutoRefresh toggles the auto-refresh feature
func (v *StatsView) toggleAutoRefresh() {
	v.autoRefresh = !v.autoRefresh
	if v.autoRefresh {
		v.startAutoRefresh()
	} else {
		v.stopAutoRefresh()
	}
}

// startAutoRefresh starts the auto-refresh goroutine
func (v *StatsView) startAutoRefresh() {
	if !v.autoRefresh {
		return
	}

	go func() {
		ticker := time.NewTicker(v.refreshInterval)
		defer ticker.Stop()

		// Initial load
		v.Refresh()

		for {
			select {
			case <-ticker.C:
				if v.autoRefresh {
					v.Refresh()
				}
			case <-v.stopRefresh:
				return
			}
		}
	}()
}

// stopAutoRefresh stops the auto-refresh goroutine
func (v *StatsView) stopAutoRefresh() {
	select {
	case v.stopRefresh <- true:
	default:
	}
}

// restartAutoRefresh restarts auto-refresh with new interval
func (v *StatsView) restartAutoRefresh() {
	v.stopAutoRefresh()
	time.Sleep(100 * time.Millisecond) // Brief pause to ensure stop
	v.startAutoRefresh()
}

// parsePercentage parses a percentage string and returns a float64
func parsePercentage(s string) float64 {
	s = strings.TrimSuffix(s, "%")
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return val
}

// parseIOBytes parses I/O byte strings like "1.2MB / 3.4GB" and returns total bytes
func parseIOBytes(s string) float64 {
	if s == "--" || s == "" {
		return 0
	}

	// Split by " / " to get input and output
	parts := strings.Split(s, " / ")
	if len(parts) == 0 {
		return 0
	}

	total := 0.0
	for _, part := range parts {
		part = strings.TrimSpace(part)
		total += parseSizeString(part)
	}
	return total
}

// parseSizeString parses size strings like "1.2MB", "3.4GB" to bytes
func parseSizeString(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "--" {
		return 0
	}

	var value float64
	var unit string

	// Try to parse the number and unit
	for i, r := range s {
		if (r < '0' || r > '9') && r != '.' {
			valStr := s[:i]
			unit = s[i:]
			val, err := strconv.ParseFloat(valStr, 64)
			if err != nil {
				return 0
			}
			value = val
			break
		}
	}

	// Convert to bytes based on unit
	unit = strings.ToUpper(strings.TrimSpace(unit))
	switch unit {
	case "B":
		return value
	case "KB", "KIB":
		return value * 1024
	case "MB", "MIB":
		return value * 1024 * 1024
	case "GB", "GIB":
		return value * 1024 * 1024 * 1024
	case "TB", "TIB":
		return value * 1024 * 1024 * 1024 * 1024
	default:
		return value
	}
}

// GetSelectedStat returns the currently selected container stat
func (v *StatsView) GetSelectedStat() *models.ContainerStats {
	row, _ := v.table.GetSelection()
	if row > 0 && row <= len(v.stats) {
		return &v.stats[row-1]
	}
	return nil
}

// Stop stops the auto-refresh when view is hidden
func (v *StatsView) Stop() {
	v.stopAutoRefresh()
}

