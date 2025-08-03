package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tokuhirom/dcv/internal/docker"
)

// TopView represents the container process list view (docker compose top)
type TopView struct {
	// View state
	width       int
	height      int
	content     []string
	scrollY     int
	projectName string

	// Loading/error state
	loading bool
	err     error

	// Dependencies
	dockerClient *docker.Client
	rootScreen   tea.Model
}

// NewTopView creates a new top view
func NewTopView(dockerClient *docker.Client, projectName string) *TopView {
	return &TopView{
		dockerClient: dockerClient,
		projectName:  projectName,
	}
}

// SetRootScreen sets the root screen reference
func (v *TopView) SetRootScreen(root tea.Model) {
	v.rootScreen = root
}

// Init initializes the view
func (v *TopView) Init() tea.Cmd {
	v.loading = true
	return loadTopData(v.dockerClient, v.projectName)
}

// Update handles messages for this view
func (v *TopView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		return v, nil

	case tea.KeyMsg:
		return v.handleKeyPress(msg)

	case topDataLoadedMsg:
		v.loading = false
		if msg.err != nil {
			v.err = msg.err
			return v, nil
		}
		v.content = msg.lines
		v.err = nil
		v.scrollY = 0
		return v, nil

	case RefreshMsg:
		v.loading = true
		v.err = nil
		return v, loadTopData(v.dockerClient, v.projectName)
	}

	return v, nil
}

// View renders the top view
func (v *TopView) View() string {
	if v.loading {
		return renderLoadingView(v.width, v.height, "Loading process information...")
	}

	if v.err != nil {
		return renderErrorView(v.width, v.height, v.err)
	}

	return v.renderTop()
}

func (v *TopView) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if v.scrollY > 0 {
			v.scrollY--
		}
		return v, nil

	case "down", "j":
		maxScroll := len(v.content) - (v.height - 4)
		if v.scrollY < maxScroll && maxScroll > 0 {
			v.scrollY++
		}
		return v, nil

	case "r":
		// Send refresh message
		return v, func() tea.Msg { return RefreshMsg{} }

	case "?":
		// Show help
		if v.rootScreen != nil {
			if switcher, ok := v.rootScreen.(interface {
				SwitchScreen(tea.Model) (tea.Model, tea.Cmd)
			}); ok {
				helpView := NewHelpView("Process Information", v)
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

func (v *TopView) renderTop() string {
	var s strings.Builder

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("7")).
		Background(lipgloss.Color("4")).
		Width(v.width).
		Padding(0, 1).
		Render("Process Information - " + v.projectName)
	s.WriteString(header + "\n")

	// Calculate visible lines
	visibleHeight := v.height - 3
	start := v.scrollY
	end := start + visibleHeight
	if end > len(v.content) {
		end = len(v.content)
	}

	// Render content
	if len(v.content) == 0 {
		s.WriteString("\nNo process information available.\n")
	} else {
		for i := start; i < end; i++ {
			if i < len(v.content) {
				s.WriteString(v.content[i] + "\n")
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
		Render("↑/↓: Scroll • r: Refresh • ESC/q: Back")

	return strings.Join(lines[:v.height-1], "\n") + "\n" + footer
}

// Messages
type topDataLoadedMsg struct {
	lines []string
	err   error
}

// Commands
func loadTopData(client *docker.Client, projectName string) tea.Cmd {
	return func() tea.Msg {
		// Get all containers for the project first
		containers, err := client.Compose(projectName).ListContainers(false)
		if err != nil {
			return topDataLoadedMsg{err: err}
		}

		// Get top output for all containers
		var allLines []string
		for _, container := range containers {
			topOutput, err := client.Compose(projectName).GetContainerTop(container.Service)
			if err == nil && topOutput != "" {
				allLines = append(allLines, fmt.Sprintf("\n=== %s ===", container.Service))
				allLines = append(allLines, strings.Split(strings.TrimSpace(topOutput), "\n")...)
			}
		}

		if len(allLines) == 0 {
			allLines = []string{"No process information available"}
		}
		if err != nil {
			return topDataLoadedMsg{err: err}
		}

		return topDataLoadedMsg{lines: allLines, err: nil}
	}
}
