package views

import (
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
)

// ProcessSortField represents the field to sort processes by
type ProcessSortField int

const (
	ProcessSortByPID ProcessSortField = iota
	ProcessSortByCPU
	ProcessSortByMem
	ProcessSortByTime
	ProcessSortByCommand
)

// TopView displays process information for a container
type TopView struct {
	docker          *docker.Client
	table           *tview.Table
	containerID     string
	containerName   string
	processes       []models.Process
	sortField       ProcessSortField
	sortReverse     bool
	autoRefresh     bool
	refreshInterval time.Duration
	stopRefresh     chan bool
	isRefreshing    bool
	mu              sync.RWMutex
}

// NewTopView creates a new top view
func NewTopView(dockerClient *docker.Client) *TopView {
	v := &TopView{
		docker:          dockerClient,
		table:           tview.NewTable(),
		sortField:       ProcessSortByCPU,
		sortReverse:     true, // Default to descending for CPU
		autoRefresh:     true,
		refreshInterval: 2 * time.Second,
		stopRefresh:     make(chan bool, 1),
	}

	v.setupTable()
	v.setupKeyHandlers()

	return v
}

// setupTable configures the table widget
func (v *TopView) setupTable() {
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
func (v *TopView) setupKeyHandlers() {
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
			v.setSortField(ProcessSortByCPU)
			v.updateTable()
			return nil

		case 'm':
			// Sort by memory
			v.setSortField(ProcessSortByMem)
			v.updateTable()
			return nil

		case 'p':
			// Sort by PID
			v.setSortField(ProcessSortByPID)
			v.updateTable()
			return nil

		case 't':
			// Sort by time
			v.setSortField(ProcessSortByTime)
			v.updateTable()
			return nil

		case 'C':
			// Sort by command
			v.setSortField(ProcessSortByCommand)
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

		case 'r':
			// Manual refresh
			v.Refresh()
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

// GetPrimitive returns the tview primitive for this view
func (v *TopView) GetPrimitive() tview.Primitive {
	return v.table
}

// Refresh refreshes the process list
func (v *TopView) Refresh() {
	if !v.isRefreshing {
		go v.loadProcesses()
	}
}

// GetTitle returns the title of the view
func (v *TopView) GetTitle() string {
	title := fmt.Sprintf("Process Info: %s", v.containerName)
	v.mu.RLock()
	autoRefresh := v.autoRefresh
	refreshInterval := v.refreshInterval
	v.mu.RUnlock()
	if autoRefresh {
		title += fmt.Sprintf(" [Auto-refresh: %ds]", int(refreshInterval.Seconds()))
	} else {
		title += " [Auto-refresh: OFF]"
	}
	return title
}

// SetContainer sets the container for this view
func (v *TopView) SetContainer(containerID string, container interface{}) {
	v.containerID = containerID
	v.containerName = ""

	// Extract container name based on type
	switch c := container.(type) {
	case models.DockerContainer:
		v.containerName = c.Names
	case models.ComposeContainer:
		v.containerName = c.Name
	case models.ContainerStats:
		v.containerName = c.Name
		// For stats, we need to use the container ID from the stats
		v.containerID = c.Container
	}

	// Start auto-refresh if enabled
	if v.autoRefresh {
		v.startAutoRefresh()
	}

	// Load processes immediately
	v.Refresh()
}

// loadProcesses loads the process list from Docker
func (v *TopView) loadProcesses() {
	v.mu.Lock()
	v.isRefreshing = true
	v.mu.Unlock()
	defer func() {
		v.mu.Lock()
		v.isRefreshing = false
		v.mu.Unlock()
	}()

	slog.Info("Loading processes for container",
		slog.String("container", v.containerID))

	// Execute docker top command
	output, err := docker.ExecuteCaptured("top", v.containerID)
	if err != nil {
		slog.Error("Failed to load processes", slog.Any("error", err))
		return
	}

	// Parse the output
	processes := v.parseProcesses(string(output))
	v.mu.Lock()
	v.processes = processes
	v.mu.Unlock()

	// Update table in UI thread
	QueueUpdateDraw(func() {
		v.updateTable()
	})
}

// parseProcesses parses the docker top output
func (v *TopView) parseProcesses(output string) []models.Process {
	var processes []models.Process
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) <= 1 {
		return processes
	}

	// Skip header line
	for i := 1; i < len(lines); i++ {
		fields := strings.Fields(lines[i])
		if len(fields) < 8 {
			continue
		}

		// Parse fields: UID PID PPID C STIME TTY TIME CMD
		process := models.Process{
			UID:   fields[0],
			PID:   fields[1],
			PPID:  fields[2],
			C:     fields[3],
			STIME: fields[4],
			TTY:   fields[5],
			TIME:  fields[6],
			CMD:   strings.Join(fields[7:], " "),
		}

		processes = append(processes, process)
	}

	return processes
}

// updateTable updates the table with process data
func (v *TopView) updateTable() {
	v.table.Clear()

	// Sort processes
	v.sortProcesses()

	// Set headers
	headers := []string{"PID", "USER", "CPU%", "TIME", "COMMAND"}
	for col, header := range headers {
		// Highlight current sort field
		color := tcell.ColorYellow
		attrs := tcell.AttrBold

		if (col == 0 && v.sortField == ProcessSortByPID) ||
			(col == 2 && v.sortField == ProcessSortByCPU) ||
			(col == 3 && v.sortField == ProcessSortByTime) ||
			(col == 4 && v.sortField == ProcessSortByCommand) {
			color = tcell.ColorGreen
		}

		cell := tview.NewTableCell(header).
			SetTextColor(color).
			SetAttributes(attrs).
			SetSelectable(false)
		v.table.SetCell(0, col, cell)
	}

	// Add process rows
	v.mu.RLock()
	processes := v.processes
	v.mu.RUnlock()

	for row, proc := range processes {
		// PID
		pidCell := tview.NewTableCell(proc.PID).
			SetTextColor(tcell.ColorWhite)
		v.table.SetCell(row+1, 0, pidCell)

		// User
		userCell := tview.NewTableCell(proc.UID).
			SetTextColor(tcell.ColorWhite)
		v.table.SetCell(row+1, 1, userCell)

		// CPU%
		cpuColor := tcell.ColorWhite
		cpuVal, _ := strconv.ParseFloat(proc.C, 64)
		if cpuVal > 50 {
			cpuColor = tcell.ColorRed
		} else if cpuVal > 20 {
			cpuColor = tcell.ColorYellow
		}
		cpuCell := tview.NewTableCell(fmt.Sprintf("%s%%", proc.C)).
			SetTextColor(cpuColor).
			SetAlign(tview.AlignRight)
		v.table.SetCell(row+1, 2, cpuCell)

		// Time
		timeCell := tview.NewTableCell(proc.TIME).
			SetTextColor(tcell.ColorWhite)
		v.table.SetCell(row+1, 3, timeCell)

		// Command (truncate if too long)
		cmd := proc.CMD
		if len(cmd) > 50 {
			cmd = cmd[:47] + "..."
		}
		cmdCell := tview.NewTableCell(cmd).
			SetTextColor(tcell.ColorWhite)
		v.table.SetCell(row+1, 4, cmdCell)
	}

	// Select first row if available
	if len(processes) > 0 {
		v.table.Select(1, 0)
	}
}

// sortProcesses sorts the process list based on current sort field
func (v *TopView) sortProcesses() {
	v.mu.Lock()
	defer v.mu.Unlock()

	sort.Slice(v.processes, func(i, j int) bool {
		var less bool
		switch v.sortField {
		case ProcessSortByPID:
			pid1, _ := strconv.Atoi(v.processes[i].PID)
			pid2, _ := strconv.Atoi(v.processes[j].PID)
			less = pid1 < pid2
		case ProcessSortByCPU:
			cpu1, _ := strconv.ParseFloat(v.processes[i].C, 64)
			cpu2, _ := strconv.ParseFloat(v.processes[j].C, 64)
			less = cpu1 < cpu2
		case ProcessSortByMem:
			// For now, sort by CPU since we don't have memory info
			cpu1, _ := strconv.ParseFloat(v.processes[i].C, 64)
			cpu2, _ := strconv.ParseFloat(v.processes[j].C, 64)
			less = cpu1 < cpu2
		case ProcessSortByTime:
			less = v.processes[i].TIME < v.processes[j].TIME
		case ProcessSortByCommand:
			less = v.processes[i].CMD < v.processes[j].CMD
		default:
			pid1, _ := strconv.Atoi(v.processes[i].PID)
			pid2, _ := strconv.Atoi(v.processes[j].PID)
			less = pid1 < pid2
		}

		if v.sortReverse {
			return !less
		}
		return less
	})
}

// setSortField sets the sort field and toggles reverse if same field
func (v *TopView) setSortField(field ProcessSortField) {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.sortField == field {
		v.sortReverse = !v.sortReverse
	} else {
		v.sortField = field
		// Default to descending for CPU and memory
		v.sortReverse = (field == ProcessSortByCPU || field == ProcessSortByMem)
	}
}

// toggleAutoRefresh toggles the auto-refresh feature
func (v *TopView) toggleAutoRefresh() {
	v.mu.Lock()
	v.autoRefresh = !v.autoRefresh
	autoRefresh := v.autoRefresh
	v.mu.Unlock()

	if autoRefresh {
		v.startAutoRefresh()
	} else {
		v.stopAutoRefresh()
	}
}

// startAutoRefresh starts the auto-refresh timer
func (v *TopView) startAutoRefresh() {
	v.stopAutoRefresh() // Stop any existing timer

	go func() {
		ticker := time.NewTicker(v.refreshInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				v.mu.RLock()
				autoRefresh := v.autoRefresh
				v.mu.RUnlock()
				if autoRefresh {
					v.Refresh()
				}
			case <-v.stopRefresh:
				return
			}
		}
	}()
}

// stopAutoRefresh stops the auto-refresh timer
func (v *TopView) stopAutoRefresh() {
	select {
	case v.stopRefresh <- true:
	default:
	}
}

// Stop stops the view and cleans up resources
func (v *TopView) Stop() {
	v.stopAutoRefresh()
}
