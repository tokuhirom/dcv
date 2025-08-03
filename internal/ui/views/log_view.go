package views

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tokuhirom/dcv/internal/docker"
)

// LogView represents the container log view
type LogView struct {
	// View state
	width         int
	height        int
	logs          []string
	scrollY       int
	containerID   string
	containerName string
	isDind        bool

	// Search state
	searchMode bool
	searchText string

	// Dependencies
	dockerClient *docker.Client
	rootScreen   tea.Model
}

// NewLogView creates a new log view
func NewLogView(dockerClient *docker.Client, containerID, containerName string, isDind bool) *LogView {
	return &LogView{
		dockerClient:  dockerClient,
		containerID:   containerID,
		containerName: containerName,
		isDind:        isDind,
		logs:          []string{},
	}
}

// SetRootScreen sets the root screen reference
func (v *LogView) SetRootScreen(root tea.Model) {
	v.rootScreen = root
}

// Init initializes the log view
func (v *LogView) Init() tea.Cmd {
	// Start streaming logs
	return streamLogs(v.dockerClient, v.containerID, v.isDind, "")
}

// Update handles messages for the log view
func (v *LogView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		return v, nil

	case tea.KeyMsg:
		if v.searchMode {
			return v.handleSearchMode(msg)
		}
		return v.handleKeyPress(msg)

	case logLineMsg:
		v.logs = append(v.logs, msg.line)
		// Keep only last 10000 lines
		if len(v.logs) > 10000 {
			v.logs = v.logs[len(v.logs)-10000:]
		}
		// Auto-scroll to bottom
		maxScroll := len(v.logs) - (v.height - 4)
		if maxScroll > 0 {
			v.scrollY = maxScroll
		}
		return v, nil

	case logLinesMsg:
		v.logs = append(v.logs, msg.lines...)
		if len(v.logs) > 10000 {
			v.logs = v.logs[len(v.logs)-10000:]
		}
		maxScroll := len(v.logs) - (v.height - 4)
		if maxScroll > 0 {
			v.scrollY = maxScroll
		}
		// Continue polling
		return v, tea.Tick(time.Millisecond*50, func(time.Time) tea.Msg {
			return pollForLogs()()
		})
	}

	return v, nil
}

// View renders the log view
func (v *LogView) View() string {
	var s strings.Builder

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("7")).
		Background(lipgloss.Color("4")).
		Width(v.width).
		Padding(0, 1).
		Render("Logs - " + v.containerName)
	s.WriteString(header + "\n")

	// Calculate visible logs
	visibleHeight := v.height - 3 // Header + footer
	start := v.scrollY
	end := start + visibleHeight
	if end > len(v.logs) {
		end = len(v.logs)
	}

	// Render logs
	if len(v.logs) == 0 {
		s.WriteString("\nWaiting for logs...\n")
	} else {
		for i := start; i < end; i++ {
			if i < len(v.logs) {
				s.WriteString(v.logs[i] + "\n")
			}
		}
	}

	// Pad to fill screen
	lines := strings.Split(s.String(), "\n")
	for len(lines) < v.height-1 {
		lines = append(lines, "")
	}

	// Footer
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Width(v.width).
		Align(lipgloss.Center).
		Render("↑/↓: Scroll • /: Search • ESC: Back • q: Back")

	return strings.Join(lines[:v.height-1], "\n") + "\n" + footer
}

func (v *LogView) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if v.scrollY > 0 {
			v.scrollY--
		}
		return v, nil

	case "down", "j":
		maxScroll := len(v.logs) - (v.height - 4)
		if v.scrollY < maxScroll && maxScroll > 0 {
			v.scrollY++
		}
		return v, nil

	case "G":
		// Go to end
		maxScroll := len(v.logs) - (v.height - 4)
		if maxScroll > 0 {
			v.scrollY = maxScroll
		}
		return v, nil

	case "g":
		// Go to start
		v.scrollY = 0
		return v, nil

	case "/":
		// Enter search mode
		v.searchMode = true
		v.searchText = ""
		return v, nil

	case "esc", "q":
		// Go back to compose list
		if v.rootScreen != nil {
			if switcher, ok := v.rootScreen.(interface {
				SwitchScreen(tea.Model) (tea.Model, tea.Cmd)
			}); ok {
				// Stop log reader
				stopLogReader()

				// Switch back to compose list
				composeView := NewComposeListView(v.dockerClient, "")
				composeView.SetRootScreen(v.rootScreen)
				return switcher.SwitchScreen(composeView)
			}
		}
		return v, nil
	}

	return v, nil
}

func (v *LogView) handleSearchMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		v.searchMode = false
		v.searchText = ""
		return v, nil

	case tea.KeyEnter:
		v.searchMode = false
		// TODO: Implement search
		return v, nil

	case tea.KeyBackspace:
		if len(v.searchText) > 0 {
			v.searchText = v.searchText[:len(v.searchText)-1]
		}
		return v, nil

	case tea.KeyRunes:
		v.searchText += string(msg.Runes)
		return v, nil
	}

	return v, nil
}

// Messages
type logLineMsg struct {
	line string
}

type logLinesMsg struct {
	lines []string
}

type pollLogsContinueMsg struct{}

// Commands
func streamLogs(client *docker.Client, containerID string, isDind bool, hostContainer string) tea.Cmd {
	// TODO: Implement actual log streaming
	return func() tea.Msg {
		return logLineMsg{line: "Starting container logs..."}
	}
}

func pollForLogs() tea.Cmd {
	// TODO: Implement actual log polling
	return func() tea.Msg {
		return pollLogsContinueMsg{}
	}
}

func stopLogReader() {
	// TODO: Implement stopping log reader
}
