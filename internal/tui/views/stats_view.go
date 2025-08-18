package views

import (
	"fmt"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
)

// StatsSortField represents the field to sort container stats by
type StatsSortField int

const (
	StatsSortByName StatsSortField = iota
	StatsSortByCPU
	StatsSortByMem
	StatsSortByNetIO
	StatsSortByBlockIO
)

func (s StatsSortField) String() string {
	switch s {
	case StatsSortByName:
		return "NAME"
	case StatsSortByCPU:
		return "CPU%"
	case StatsSortByMem:
		return "MEM%"
	case StatsSortByNetIO:
		return "NET I/O"
	case StatsSortByBlockIO:
		return "BLOCK I/O"
	default:
		return "NAME"
	}
}

// StatsView displays container statistics
type StatsView struct {
	docker            *docker.Client
	table             *tview.Table
	stats             []models.ContainerStats
	sortField         StatsSortField
	sortReverse       bool
	autoRefresh       bool
	refreshInterval   time.Duration
	stopRefresh       chan bool
	isRefreshing      bool
	showAll           bool         // Show all containers or just running ones
	mu                sync.RWMutex // Protects showAll, autoRefresh, isRefreshing
	switchToLogViewFn func(containerID string, container interface{})
	switchToTopViewFn func(containerID string, container interface{})
}

// NewStatsView creates a new stats view
func NewStatsView(dockerClient *docker.Client) *StatsView {
	v := &StatsView{
		docker:          dockerClient,
		table:           tview.NewTable(),
		sortField:       StatsSortByCPU,
		sortReverse:     true, // Default to descending for CPU
		autoRefresh:     true,
		refreshInterval: 2 * time.Second,
		stopRefresh:     make(chan bool, 1),
		showAll:         false,
	}

	v.setupTable()
	v.setupKeyHandlers()
	v.startAutoRefresh()

	return v
}

// setupTable configures the table widget
func (v *StatsView) setupTable() {
	v.table.SetBorders(false).
		SetSelectable(true, false).
		SetSeparator(' ').
		SetFixed(1, 0)

	// Set header style
	v.table.SetSelectedStyle(tcell.StyleDefault.
		Background(tcell.ColorDarkCyan).
		Foreground(tcell.ColorWhite))
}

// setupKeyHandlers sets up keyboard shortcuts for the view
func (v *StatsView) setupKeyHandlers() {
	v.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := v.table.GetSelection()

		switch event.Rune() {
		case 'j':
			// Move down (vim style)
			if row < v.table.GetRowCount()-1 {
				v.table.Select(row+1, 0)
			}
			return nil

		case 'k':
			// Move up (vim style)
			if row > 1 { // Skip header row
				v.table.Select(row-1, 0)
			}
			return nil

		case 'g':
			// Go to top (vim style)
			v.table.Select(1, 0)
			return nil

		case 'G':
			// Go to bottom (vim style)
			rowCount := v.table.GetRowCount()
			if rowCount > 1 {
				v.table.Select(rowCount-1, 0)
			}
			return nil

		case 'c':
			// Sort by CPU
			v.setSortField(StatsSortByCPU)
			v.updateTable()
			return nil

		case 'm':
			// Sort by memory
			v.setSortField(StatsSortByMem)
			v.updateTable()
			return nil

		case 'n':
			// Sort by name
			v.setSortField(StatsSortByName)
			v.updateTable()
			return nil

		case 'N':
			// Sort by network I/O
			v.setSortField(StatsSortByNetIO)
			v.updateTable()
			return nil

		case 'B':
			// Sort by block I/O
			v.setSortField(StatsSortByBlockIO)
			v.updateTable()
			return nil

		case 'R':
			// Reverse sort order
			v.mu.Lock()
			v.sortReverse = !v.sortReverse
			v.mu.Unlock()
			v.updateTable()
			return nil

		case 'a':
			// Toggle auto-refresh
			v.toggleAutoRefresh()
			return nil

		case 'A':
			// Toggle show all containers
			v.mu.Lock()
			v.showAll = !v.showAll
			v.mu.Unlock()
			v.Refresh()
			return nil

		case 'r':
			// Manual refresh
			v.Refresh()
			return nil

		case 'l':
			// View logs
			if row > 0 && row <= len(v.stats) {
				stat := v.stats[row-1]
				if v.switchToLogViewFn != nil {
					// Use Container ID from stats
					v.switchToLogViewFn(stat.Container, stat)
				}
			}
			return nil

		case 't':
			// View top (processes)
			if row > 0 && row <= len(v.stats) {
				stat := v.stats[row-1]
				if v.switchToTopViewFn != nil {
					// Use Container ID from stats
					v.switchToTopViewFn(stat.Container, stat)
				}
			}
			return nil

		case '+':
			// Increase refresh interval
			v.mu.Lock()
			if v.refreshInterval < 10*time.Second {
				v.refreshInterval += time.Second
			}
			autoRefresh := v.autoRefresh
			v.mu.Unlock()
			if autoRefresh {
				v.restartAutoRefresh()
			}
			return nil

		case '-':
			// Decrease refresh interval
			v.mu.Lock()
			if v.refreshInterval > time.Second {
				v.refreshInterval -= time.Second
			}
			autoRefresh := v.autoRefresh
			v.mu.Unlock()
			if autoRefresh {
				v.restartAutoRefresh()
			}
			return nil
		}

		switch event.Key() {
		case tcell.KeyCtrlR:
			// Force refresh
			v.Refresh()
			return nil
		}

		return event
	})
}

func (v *StatsView) GetPrimitive() tview.Primitive {
	return v.table
}

func (v *StatsView) Refresh() {
	if !v.isRefreshing {
		go v.loadStats()
	}
}

func (v *StatsView) GetTitle() string {
	title := "Container Statistics"
	v.mu.RLock()
	autoRefresh := v.autoRefresh
	showAll := v.showAll
	refreshInterval := v.refreshInterval
	v.mu.RUnlock()
	if autoRefresh {
		title += fmt.Sprintf(" [Auto-refresh: %ds]", int(refreshInterval.Seconds()))
	} else {
		title += " [Auto-refresh: OFF]"
	}
	if showAll {
		title += " [All]"
	} else {
		title += " [Running]"
	}
	return title
}

// SetSwitchToLogViewCallback sets the callback for switching to log view
func (v *StatsView) SetSwitchToLogViewCallback(fn func(containerID string, container interface{})) {
	v.switchToLogViewFn = fn
}

// SetSwitchToTopViewCallback sets the callback for switching to top view
func (v *StatsView) SetSwitchToTopViewCallback(fn func(containerID string, container interface{})) {
	v.switchToTopViewFn = fn
}
