package ui

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
)

// SortField represents the field to sort processes by
type SortField int

const (
	SortByPID SortField = iota
	SortByCPU
	SortByMem
	SortByTime
	SortByCommand
)

func (s SortField) String() string {
	switch s {
	case SortByPID:
		return "PID"
	case SortByCPU:
		return "CPU%"
	case SortByMem:
		return "MEM%"
	case SortByTime:
		return "TIME"
	case SortByCommand:
		return "COMMAND"
	default:
		return "PID"
	}
}

// TopViewModel manages the state and rendering of the process info view
type TopViewModel struct {
	processes       []models.Process
	containerStats  *models.ContainerStats
	sortField       SortField
	sortReverse     bool
	scrollY         int
	autoRefresh     bool
	refreshInterval time.Duration

	container *docker.Container
}

// render renders the top view
func (m *TopViewModel) render(availableHeight int) string {
	var s strings.Builder

	if len(m.processes) == 0 {
		s.WriteString("No process information available.\n")
		return s.String()
	}

	// Display container stats header
	if m.containerStats != nil {
		s.WriteString(m.renderStatsHeader())
		s.WriteString("\n\n")
	}

	// Display sorting info and auto-refresh status
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

	s.WriteString(helpStyle.Render("[c]PU [m]EM [p]ID [t]IME [n]ame [r]everse [a]uto-refresh"))
	s.WriteString("\n\n")

	// Sort processes
	m.sortProcesses()

	// Display process header
	header := m.renderProcessHeader()
	s.WriteString(header)
	s.WriteString("\n")
	s.WriteString(strings.Repeat("─", 100))
	s.WriteString("\n")

	// Calculate visible area
	headerLines := 7 // Stats header + sort info + process header
	if m.containerStats == nil {
		headerLines = 5
	}
	visibleHeight := availableHeight - headerLines

	// Display processes
	for i := m.scrollY; i < len(m.processes) && i < m.scrollY+visibleHeight; i++ {
		s.WriteString(m.renderProcess(&m.processes[i]))
		s.WriteString("\n")
	}

	return s.String()
}

func (m *TopViewModel) renderStatsHeader() string {
	if m.containerStats == nil {
		return ""
	}

	// Create a styled header with container resource usage
	cpuStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
	memStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))

	return fmt.Sprintf("%s: %s  %s: %s (%s)  PIDs: %s",
		cpuStyle.Render("CPU"),
		m.containerStats.CPUPerc,
		memStyle.Render("Memory"),
		m.containerStats.MemUsage,
		m.containerStats.MemPerc,
		m.containerStats.PIDs,
	)
}

func (m *TopViewModel) renderProcessHeader() string {
	// Highlight the current sort field
	pidHeader := "PID"
	cpuHeader := "CPU%"
	memHeader := "MEM%"
	timeHeader := "TIME"
	cmdHeader := "COMMAND"

	highlightStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))

	switch m.sortField {
	case SortByPID:
		pidHeader = highlightStyle.Render(pidHeader)
	case SortByCPU:
		cpuHeader = highlightStyle.Render(cpuHeader)
	case SortByMem:
		memHeader = highlightStyle.Render(memHeader)
	case SortByTime:
		timeHeader = highlightStyle.Render(timeHeader)
	case SortByCommand:
		cmdHeader = highlightStyle.Render(cmdHeader)
	}

	return fmt.Sprintf("%-8s %-8s %-8s %-6s %-6s %-10s %-10s %s",
		"UID", pidHeader, "PPID", cpuHeader, memHeader, "STIME", timeHeader, cmdHeader)
}

func (m *TopViewModel) renderProcess(p *models.Process) string {
	// Color code based on CPU usage
	cpuStr := fmt.Sprintf("%.1f%%", p.CPUPerc)
	memStr := fmt.Sprintf("%.1f%%", p.MemPerc)

	if p.CPUPerc > 50 {
		cpuStr = errorStyle.Render(cpuStr)
	} else if p.CPUPerc > 20 {
		cpuStr = searchStyle.Render(cpuStr)
	}

	if p.MemPerc > 50 {
		memStr = errorStyle.Render(memStr)
	} else if p.MemPerc > 20 {
		memStr = searchStyle.Render(memStr)
	}

	// Truncate command if too long
	cmd := p.CMD
	if len(cmd) > 50 {
		cmd = cmd[:47] + "..."
	}

	return fmt.Sprintf("%-8s %-8s %-8s %-6s %-6s %-10s %-10s %s",
		p.UID, p.PID, p.PPID, cpuStr, memStr, p.STIME, p.TIME, cmd)
}

func (m *TopViewModel) sortProcesses() {
	sort.Slice(m.processes, func(i, j int) bool {
		var less bool
		switch m.sortField {
		case SortByPID:
			pid1, _ := strconv.Atoi(m.processes[i].PID)
			pid2, _ := strconv.Atoi(m.processes[j].PID)
			less = pid1 < pid2
		case SortByCPU:
			less = m.processes[i].CPUPerc < m.processes[j].CPUPerc
		case SortByMem:
			less = m.processes[i].MemPerc < m.processes[j].MemPerc
		case SortByTime:
			less = m.processes[i].TIME < m.processes[j].TIME
		case SortByCommand:
			less = m.processes[i].CMD < m.processes[j].CMD
		default:
			pid1, _ := strconv.Atoi(m.processes[i].PID)
			pid2, _ := strconv.Atoi(m.processes[j].PID)
			less = pid1 < pid2
		}

		if m.sortReverse {
			return !less
		}
		return less
	})
}

// Load switches to the top view and loads process info
func (m *TopViewModel) Load(model *Model, container *docker.Container) tea.Cmd {
	m.container = container
	m.autoRefresh = true                // Enable auto-refresh by default
	m.refreshInterval = 2 * time.Second // Default refresh interval
	model.SwitchView(TopView)
	return tea.Batch(
		m.DoLoad(model),
		m.startAutoRefresh(),
	)
}

// DoLoad reloads the process info
func (m *TopViewModel) DoLoad(model *Model) tea.Cmd {
	model.loading = true
	return m.doLoadInternal(model)
}

// DoLoadSilent reloads the process info without showing loading indicator
func (m *TopViewModel) DoLoadSilent(model *Model) tea.Cmd {
	// Don't set loading = true for silent refresh
	return m.doLoadInternal(model)
}

func (m *TopViewModel) doLoadInternal(model *Model) tea.Cmd {
	return func() tea.Msg {
		// Get process list
		args := m.container.OperationArgs("top")
		topOutput, err := model.dockerClient.ExecuteCaptured(args...)
		if err != nil {
			return topLoadedMsg{err: err}
		}

		// Get container stats
		statsArgs := []string{"stats", "--no-stream", "--format", "json", m.container.GetContainerID()}
		statsOutput, statsErr := model.dockerClient.ExecuteCaptured(statsArgs...)

		var stats *models.ContainerStats
		if statsErr == nil && len(statsOutput) > 0 {
			stats = &models.ContainerStats{}
			if err := json.Unmarshal(statsOutput, stats); err != nil {
				stats = nil
			}
		}

		return topLoadedMsg{
			processes: m.parseProcesses(string(topOutput)),
			stats:     stats,
			err:       err,
		}
	}
}

// parseProcesses parses the docker top output into Process structs
func (m *TopViewModel) parseProcesses(output string) []models.Process {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		return nil
	}

	var processes []models.Process

	// Skip the header line
	for i := 1; i < len(lines); i++ {
		fields := strings.Fields(lines[i])
		if len(fields) < 8 {
			continue
		}

		p := models.Process{
			UID:   fields[0],
			PID:   fields[1],
			PPID:  fields[2],
			C:     fields[3],
			STIME: fields[4],
			TTY:   fields[5],
			TIME:  fields[6],
			CMD:   strings.Join(fields[7:], " "),
		}

		// Parse CPU percentage from C field if available
		if cpu, err := strconv.ParseFloat(fields[3], 64); err == nil {
			p.CPUPerc = cpu
		}

		processes = append(processes, p)
	}

	// If we have container stats, distribute CPU/Memory proportionally
	// This is a simplified approach - in reality, we'd need per-process stats
	if m.containerStats != nil {
		totalCPU := models.ParsePercentage(m.containerStats.CPUPerc)
		totalMem := models.ParsePercentage(m.containerStats.MemPerc)

		if len(processes) > 0 {
			// Distribute evenly for now (in a real implementation, we'd use /proc/[pid]/stat)
			perProcessCPU := totalCPU / float64(len(processes))
			perProcessMem := totalMem / float64(len(processes))

			for i := range processes {
				processes[i].CPUPerc = perProcessCPU
				processes[i].MemPerc = perProcessMem
			}
		}
	}

	return processes
}

// HandleBack returns to the compose process list view
func (m *TopViewModel) HandleBack(model *Model) tea.Cmd {
	model.SwitchToPreviousView()
	return nil
}

// Loaded updates the top output after loading
func (m *TopViewModel) Loaded(processes []models.Process, stats *models.ContainerStats) {
	m.processes = processes
	m.containerStats = stats
	m.scrollY = 0
}

func (m *TopViewModel) Title() string {
	return fmt.Sprintf("Process Info: %s", m.container.Title())
}

// HandleUp scrolls up in the process list
func (m *TopViewModel) HandleUp() {
	if m.scrollY > 0 {
		m.scrollY--
	}
}

// HandleDown scrolls down in the process list
func (m *TopViewModel) HandleDown() {
	if m.scrollY < len(m.processes)-1 {
		m.scrollY++
	}
}

// HandleSortByCPU sorts processes by CPU usage
func (m *TopViewModel) HandleSortByCPU() {
	if m.sortField == SortByCPU {
		m.sortReverse = !m.sortReverse
	} else {
		m.sortField = SortByCPU
		m.sortReverse = true // Default to descending for CPU
	}
}

// HandleSortByMem sorts processes by memory usage
func (m *TopViewModel) HandleSortByMem() {
	if m.sortField == SortByMem {
		m.sortReverse = !m.sortReverse
	} else {
		m.sortField = SortByMem
		m.sortReverse = true // Default to descending for memory
	}
}

// HandleSortByPID sorts processes by PID
func (m *TopViewModel) HandleSortByPID() {
	if m.sortField == SortByPID {
		m.sortReverse = !m.sortReverse
	} else {
		m.sortField = SortByPID
		m.sortReverse = false // Default to ascending for PID
	}
}

// HandleSortByTime sorts processes by CPU time
func (m *TopViewModel) HandleSortByTime() {
	if m.sortField == SortByTime {
		m.sortReverse = !m.sortReverse
	} else {
		m.sortField = SortByTime
		m.sortReverse = true // Default to descending for time
	}
}

// HandleSortByCommand sorts processes by command name
func (m *TopViewModel) HandleSortByCommand() {
	if m.sortField == SortByCommand {
		m.sortReverse = !m.sortReverse
	} else {
		m.sortField = SortByCommand
		m.sortReverse = false // Default to ascending for command
	}
}

// HandleReverseSort reverses the current sort order
func (m *TopViewModel) HandleReverseSort() {
	m.sortReverse = !m.sortReverse
}

// HandleToggleAutoRefresh toggles the auto-refresh feature
func (m *TopViewModel) HandleToggleAutoRefresh() {
	m.autoRefresh = !m.autoRefresh
}

// startAutoRefresh returns a command to trigger periodic refresh
func (m *TopViewModel) startAutoRefresh() tea.Cmd {
	return tea.Tick(m.refreshInterval, func(time.Time) tea.Msg {
		return autoRefreshTickMsg{}
	})
}
