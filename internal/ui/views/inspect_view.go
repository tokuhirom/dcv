package views

import (
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tokuhirom/dcv/internal/docker"
)

// InspectView represents the inspect view for containers, images, networks, volumes
type InspectView struct {
	// View state
	width        int
	height       int
	content      []string
	scrollY      int
	resourceID   string
	resourceType string
	resourceName string

	// Loading/error state
	loading bool
	err     error

	// Dependencies
	dockerClient *docker.Client
	rootScreen   tea.Model
}

// NewInspectView creates a new inspect view
func NewInspectView(dockerClient *docker.Client, resourceID, resourceType, resourceName string) *InspectView {
	return &InspectView{
		dockerClient: dockerClient,
		resourceID:   resourceID,
		resourceType: resourceType,
		resourceName: resourceName,
	}
}

// SetRootScreen sets the root screen reference
func (v *InspectView) SetRootScreen(root tea.Model) {
	v.rootScreen = root
}

// Init initializes the view
func (v *InspectView) Init() tea.Cmd {
	v.loading = true
	return loadInspectData(v.dockerClient, v.resourceID, v.resourceType)
}

// Update handles messages for this view
func (v *InspectView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		return v, nil

	case tea.KeyMsg:
		return v.handleKeyPress(msg)

	case inspectDataLoadedMsg:
		v.loading = false
		if msg.err != nil {
			v.err = msg.err
			return v, nil
		}
		// Format JSON data
		var data interface{}
		if err := json.Unmarshal([]byte(msg.data), &data); err != nil {
			v.content = []string{msg.data}
		} else {
			formatted, err := json.MarshalIndent(data, "", "  ")
			if err != nil {
				v.content = []string{msg.data}
			} else {
				v.content = strings.Split(string(formatted), "\n")
			}
		}
		v.err = nil
		v.scrollY = 0
		return v, nil
	}

	return v, nil
}

// View renders the inspect view
func (v *InspectView) View() string {
	if v.loading {
		return renderLoadingView(v.width, v.height, "Loading inspect data...")
	}

	if v.err != nil {
		return renderErrorView(v.width, v.height, v.err)
	}

	return v.renderInspect()
}

func (v *InspectView) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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

	case "G":
		// Jump to end
		maxScroll := len(v.content) - (v.height - 4)
		if maxScroll > 0 {
			v.scrollY = maxScroll
		}
		return v, nil

	case "g":
		// Jump to start
		v.scrollY = 0
		return v, nil

	case "?":
		// Show help
		if v.rootScreen != nil {
			if switcher, ok := v.rootScreen.(interface {
				SwitchScreen(tea.Model) (tea.Model, tea.Cmd)
			}); ok {
				helpView := NewHelpView("Inspect View", v)
				helpView.SetRootScreen(v.rootScreen)
				return switcher.SwitchScreen(helpView)
			}
		}
		return v, nil

	case "esc", "q":
		// Go back
		// TODO: Navigate back to the appropriate view based on resourceType
		return v, tea.Quit
	}

	return v, nil
}

func (v *InspectView) renderInspect() string {
	var s strings.Builder

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("7")).
		Background(lipgloss.Color("4")).
		Width(v.width).
		Padding(0, 1).
		Render("Inspect - " + v.resourceName)
	s.WriteString(header + "\n")

	// Calculate visible lines
	visibleHeight := v.height - 3
	start := v.scrollY
	end := start + visibleHeight
	if end > len(v.content) {
		end = len(v.content)
	}

	// Render content
	for i := start; i < end; i++ {
		if i < len(v.content) {
			s.WriteString(v.content[i] + "\n")
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
		Render("↑/↓: Scroll • G: End • g: Start • ESC: Back")

	return strings.Join(lines[:v.height-1], "\n") + "\n" + footer
}

// Messages
type inspectDataLoadedMsg struct {
	data string
	err  error
}

// Commands
func loadInspectData(client *docker.Client, resourceID, resourceType string) tea.Cmd {
	return func() tea.Msg {
		var data string
		var err error

		switch resourceType {
		case "container":
			data, err = client.InspectContainer(resourceID)
		case "image":
			data, err = client.InspectImage(resourceID)
		case "network":
			data, err = client.InspectNetwork(resourceID)
		case "volume":
			data, err = client.InspectVolume(resourceID)
		default:
			return inspectDataLoadedMsg{err: fmt.Errorf("unknown resource type: %s", resourceType)}
		}

		return inspectDataLoadedMsg{data: data, err: err}
	}
}
