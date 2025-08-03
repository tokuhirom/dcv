package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
)

// DindView represents the Docker-in-Docker container list view
type DindView struct {
	// View state
	width             int
	height            int
	selectedContainer int
	containers        []models.DockerContainer
	hostContainerID   string
	hostContainerName string

	// Loading/error state
	loading bool
	err     error

	// Dependencies
	dockerClient *docker.Client
	rootScreen   tea.Model
}

// NewDindView creates a new dind view
func NewDindView(dockerClient *docker.Client, hostContainerID, hostContainerName string) *DindView {
	return &DindView{
		dockerClient:      dockerClient,
		hostContainerID:   hostContainerID,
		hostContainerName: hostContainerName,
	}
}

// SetRootScreen sets the root screen reference
func (v *DindView) SetRootScreen(root tea.Model) {
	v.rootScreen = root
}

// Init initializes the view
func (v *DindView) Init() tea.Cmd {
	v.loading = true
	return loadDindContainers(v.dockerClient, v.hostContainerID)
}

// Update handles messages for this view
func (v *DindView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		return v, nil

	case tea.KeyMsg:
		return v.handleKeyPress(msg)

	case dindContainersLoadedMsg:
		v.loading = false
		if msg.err != nil {
			v.err = msg.err
			return v, nil
		}
		v.containers = msg.containers
		v.err = nil
		if len(v.containers) > 0 && v.selectedContainer >= len(v.containers) {
			v.selectedContainer = 0
		}
		return v, nil

	case RefreshMsg:
		v.loading = true
		v.err = nil
		return v, loadDindContainers(v.dockerClient, v.hostContainerID)
	}

	return v, nil
}

// View renders the dind container list
func (v *DindView) View() string {
	if v.loading {
		return renderLoadingView(v.width, v.height, "Loading dind containers...")
	}

	if v.err != nil {
		return renderErrorView(v.width, v.height, v.err)
	}

	return v.renderDindList()
}

func (v *DindView) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if v.selectedContainer > 0 {
			v.selectedContainer--
		}
		return v, nil

	case "down", "j":
		if v.selectedContainer < len(v.containers)-1 {
			v.selectedContainer++
		}
		return v, nil

	case "enter":
		// View logs of dind container
		if v.selectedContainer < len(v.containers) && v.rootScreen != nil {
			container := v.containers[v.selectedContainer]
			if switcher, ok := v.rootScreen.(interface {
				SwitchScreen(tea.Model) (tea.Model, tea.Cmd)
			}); ok {
				// Extract container name
				name := strings.TrimPrefix(container.Names, "/")
				logView := NewLogView(v.dockerClient, container.ID, name, true)
				logView.SetRootScreen(v.rootScreen)
				return switcher.SwitchScreen(logView)
			}
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
				helpView := NewHelpView("Docker-in-Docker", v)
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
				composeView := NewComposeListView(v.dockerClient, "")
				composeView.SetRootScreen(v.rootScreen)
				return switcher.SwitchScreen(composeView)
			}
		}
		return v, nil
	}

	return v, nil
}

func (v *DindView) renderDindList() string {
	var s strings.Builder

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("7")).
		Background(lipgloss.Color("4")).
		Width(v.width).
		Padding(0, 1).
		Render(fmt.Sprintf("Docker-in-Docker - %s", v.hostContainerName))
	s.WriteString(header + "\n")

	// Container list
	if len(v.containers) == 0 {
		s.WriteString("\nNo containers found in dind.\n")
	} else {
		// Column headers
		headers := fmt.Sprintf("%-15s %-30s %-20s %-15s %s",
			"CONTAINER ID", "IMAGE", "COMMAND", "STATUS", "NAMES")
		s.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Bold(true).
			Render(headers) + "\n")

		for i, container := range v.containers {
			selected := i == v.selectedContainer
			line := formatDindContainerLine(container, v.width, selected)
			s.WriteString(line + "\n")
		}
	}

	// Footer
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Width(v.width).
		Align(lipgloss.Center).
		Render("↑/↓: Navigate • Enter: View Logs • r: Refresh • ESC: Back")

	// Pad to fill screen
	content := s.String()
	lines := strings.Split(content, "\n")
	for len(lines) < v.height-2 {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n") + "\n" + footer
}

func formatDindContainerLine(container models.DockerContainer, width int, selected bool) string {
	// Truncate long fields
	id := container.ID
	if len(id) > 12 {
		id = id[:12]
	}

	image := container.Image
	if len(image) > 28 {
		image = image[:28]
	}

	command := container.Command
	if len(command) > 18 {
		command = command[:18]
	}

	status := container.Status
	if len(status) > 13 {
		status = status[:13]
	}

	names := container.Names
	if len(names) > 20 {
		names = names[:20]
	}

	line := fmt.Sprintf("%-15s %-30s %-20s %-15s %s",
		id, image, command, status, names)

	if len(line) > width-3 {
		line = line[:width-3]
	}

	style := lipgloss.NewStyle()
	if selected {
		style = style.Background(lipgloss.Color("240"))
	}

	// Color based on state
	if strings.Contains(container.State, "running") {
		style = style.Foreground(lipgloss.Color("2")) // Green
	} else if strings.Contains(container.State, "exited") {
		style = style.Foreground(lipgloss.Color("240")) // Gray
	}

	return style.Render(line)
}

// Messages
type dindContainersLoadedMsg struct {
	containers []models.DockerContainer
	err        error
}

// Commands
func loadDindContainers(client *docker.Client, hostContainerID string) tea.Cmd {
	return func() tea.Msg {
		containers, err := client.ListDindContainers(hostContainerID)
		return dindContainersLoadedMsg{containers: containers, err: err}
	}
}
