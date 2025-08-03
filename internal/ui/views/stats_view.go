package views

import (
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tokuhirom/dcv/internal/docker"
)

// StatsView represents the container statistics view
type StatsView struct {
	// View state
	width       int
	height      int
	stats       []ContainerStats
	projectName string

	// Loading/error state
	loading bool
	err     error

	// Dependencies
	dockerClient *docker.Client
	rootScreen   tea.Model
}

// NewStatsView creates a new stats view
func NewStatsView(dockerClient *docker.Client, projectName string) *StatsView {
	return &StatsView{
		dockerClient: dockerClient,
		projectName:  projectName,
	}
}

// SetRootScreen sets the root screen reference
func (v *StatsView) SetRootScreen(root tea.Model) {
	v.rootScreen = root
}

// Init initializes the view
func (v *StatsView) Init() tea.Cmd {
	v.loading = true
	return loadStats(v.dockerClient, v.projectName)
}

// Update handles messages for this view
func (v *StatsView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		return v, nil

	case tea.KeyMsg:
		return v.handleKeyPress(msg)

	case statsLoadedMsg:
		v.loading = false
		if msg.err != nil {
			v.err = msg.err
			return v, nil
		}
		v.stats = msg.stats
		v.err = nil
		return v, nil

	case RefreshMsg:
		v.loading = true
		v.err = nil
		return v, loadStats(v.dockerClient, v.projectName)
	}

	return v, nil
}

// View renders the stats view
func (v *StatsView) View() string {
	if v.loading {
		return renderLoadingView(v.width, v.height, "Loading container statistics...")
	}

	if v.err != nil {
		return renderErrorView(v.width, v.height, v.err)
	}

	return v.renderStats()
}

func (v *StatsView) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "r":
		// Send refresh message
		return v, func() tea.Msg { return RefreshMsg{} }

	case "?":
		// Show help
		if v.rootScreen != nil {
			if switcher, ok := v.rootScreen.(interface {
				SwitchScreen(tea.Model) (tea.Model, tea.Cmd)
			}); ok {
				helpView := NewHelpView("Container Stats", v)
				helpView.SetRootScreen(v.rootScreen)
				return switcher.SwitchScreen(helpView)
			}
		}
		return v, nil

	case "esc", "q":
		// Go back to compose list
		if v.rootScreen != nil {
			if switcher, ok := v.rootScreen.(interface {
				SwitchScreen(tea.Model) (tea.Model, tea.Cmd)
			}); ok {
				composeView := NewComposeListView(v.dockerClient, v.projectName)
				composeView.SetRootScreen(v.rootScreen)
				return switcher.SwitchScreen(composeView)
			}
		}
		return v, nil
	}

	return v, nil
}

func (v *StatsView) renderStats() string {
	var s strings.Builder

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("7")).
		Background(lipgloss.Color("4")).
		Width(v.width).
		Padding(0, 1).
		Render("Container Statistics - " + v.projectName)
	s.WriteString(header + "\n")

	// Stats table
	if len(v.stats) == 0 {
		s.WriteString("\nNo statistics available.\n")
	} else {
		// Column headers
		headers := fmt.Sprintf("%-30s %-12s %-12s %-15s %-15s %-12s %-12s %s",
			"CONTAINER", "CPU %", "MEM USAGE", "MEM %", "NET I/O", "BLOCK I/O", "PIDS", "STATUS")
		s.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Bold(true).
			Render(headers) + "\n")

		for _, stat := range v.stats {
			line := formatStatsLine(stat, v.width)
			s.WriteString(line + "\n")
		}
	}

	// Footer
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Width(v.width).
		Align(lipgloss.Center).
		Render("r: Refresh • ESC/q: Back")

	// Pad to fill screen
	content := s.String()
	lines := strings.Split(content, "\n")
	for len(lines) < v.height-2 {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n") + "\n" + footer
}

func formatStatsLine(stat ContainerStats, width int) string {
	// Truncate container name if too long
	name := stat.Name
	if len(name) > 28 {
		name = name[:28]
	}

	line := fmt.Sprintf("%-30s %-12s %-12s %-15s %-15s %-12s %-12s %s",
		name,
		stat.CPUPerc,
		stat.MemUsage,
		stat.MemPerc,
		stat.NetIO,
		stat.BlockIO,
		stat.PIDs,
		stat.Status)

	if len(line) > width-3 {
		line = line[:width-3]
	}

	style := lipgloss.NewStyle()

	// Color based on status
	if strings.Contains(stat.Status, "running") {
		style = style.Foreground(lipgloss.Color("2")) // Green
	} else {
		style = style.Foreground(lipgloss.Color("240")) // Gray
	}

	return style.Render(line)
}

// Messages
// ContainerStats represents container resource usage statistics
type ContainerStats struct {
	Container string `json:"Container"`
	Name      string `json:"Name"`
	Service   string `json:"Service"`
	CPUPerc   string `json:"CPUPerc"`
	MemUsage  string `json:"MemUsage"`
	MemPerc   string `json:"MemPerc"`
	NetIO     string `json:"NetIO"`
	BlockIO   string `json:"BlockIO"`
	PIDs      string `json:"PIDs"`
	Status    string `json:"Status"`
}

type statsLoadedMsg struct {
	stats []ContainerStats
	err   error
}

// Commands
func loadStats(client *docker.Client, projectName string) tea.Cmd {
	return func() tea.Msg {
		// Get raw stats output
		output, err := client.GetStats()
		if err != nil {
			return statsLoadedMsg{err: err}
		}

		// Parse stats
		var stats []ContainerStats
		if output != "" {
			lines := strings.Split(strings.TrimSpace(output), "\n")
			for _, line := range lines {
				var stat ContainerStats
				if err := json.Unmarshal([]byte(line), &stat); err == nil {
					// Filter by project if specified
					if projectName == "" || stat.Service != "" {
						stats = append(stats, stat)
					}
				}
			}
		}
		return statsLoadedMsg{stats: stats, err: err}
	}
}
