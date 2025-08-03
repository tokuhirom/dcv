package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
)

// ComposeListView represents the Docker Compose process list view
type ComposeListView struct {
	// View state
	width             int
	height            int
	selectedContainer int
	containers        []models.ComposeContainer
	projectName       string
	showAll           bool

	// Loading/error state
	loading bool
	err     error

	// Dependencies
	dockerClient *docker.Client
	rootScreen   tea.Model // Reference to root for switching views
}

// NewComposeListView creates a new compose list view
func NewComposeListView(dockerClient *docker.Client, projectName string) *ComposeListView {
	return &ComposeListView{
		dockerClient: dockerClient,
		projectName:  projectName,
		showAll:      false,
	}
}

// SetRootScreen sets the root screen reference
func (v *ComposeListView) SetRootScreen(root tea.Model) {
	v.rootScreen = root
}

// SetSize updates the view dimensions
func (v *ComposeListView) SetSize(width, height int) {
	v.width = width
	v.height = height
}

// Init initializes the view
func (v *ComposeListView) Init() tea.Cmd {
	v.loading = true
	return loadProcesses(v.dockerClient, v.projectName, v.showAll)
}

// Update handles messages for this view
func (v *ComposeListView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		return v, nil

	case tea.KeyMsg:
		return v.handleKeyPress(msg)

	case processesLoadedMsg:
		v.loading = false
		if msg.err != nil {
			v.err = msg.err
			return v, nil
		}
		v.containers = msg.processes
		v.err = nil
		if len(v.containers) > 0 && v.selectedContainer >= len(v.containers) {
			v.selectedContainer = 0
		}
		return v, nil

	case RefreshMsg:
		v.loading = true
		v.err = nil
		return v, loadProcesses(v.dockerClient, v.projectName, v.showAll)
	}

	return v, nil
}

// View renders the compose list view
func (v *ComposeListView) View() string {
	if v.loading {
		return renderLoadingView(v.width, v.height, "Loading compose containers...")
	}

	if v.err != nil {
		return renderErrorView(v.width, v.height, v.err)
	}

	return v.renderComposeList()
}

func (v *ComposeListView) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		// Switch to log view
		if v.selectedContainer < len(v.containers) && v.rootScreen != nil {
			container := v.containers[v.selectedContainer]
			// Use the root screen's SwitchScreen method
			if switcher, ok := v.rootScreen.(interface {
				SwitchScreen(tea.Model) (tea.Model, tea.Cmd)
			}); ok {
				logView := NewLogView(v.dockerClient, container.ID, container.Name, false)
				logView.SetRootScreen(v.rootScreen)
				return switcher.SwitchScreen(logView)
			}
		}
		return v, nil

	case "r":
		// Send refresh message
		return v, func() tea.Msg { return RefreshMsg{} }

	case "a":
		// Toggle show all
		v.showAll = !v.showAll
		v.loading = true
		return v, loadProcesses(v.dockerClient, v.projectName, v.showAll)

	case "d":
		// View dind containers
		if v.selectedContainer < len(v.containers) && v.rootScreen != nil {
			container := v.containers[v.selectedContainer]
			// Check if container name contains "dind"
			if strings.Contains(strings.ToLower(container.Name), "dind") ||
				strings.Contains(strings.ToLower(container.Service), "dind") {
				if switcher, ok := v.rootScreen.(interface {
					SwitchScreen(tea.Model) (tea.Model, tea.Cmd)
				}); ok {
					dindView := NewDindView(v.dockerClient, container.ID, container.Name)
					dindView.SetRootScreen(v.rootScreen)
					return switcher.SwitchScreen(dindView)
				}
			}
		}
		return v, nil

	case "f":
		// Browse container files
		if v.selectedContainer < len(v.containers) && v.rootScreen != nil {
			container := v.containers[v.selectedContainer]
			if switcher, ok := v.rootScreen.(interface {
				SwitchScreen(tea.Model) (tea.Model, tea.Cmd)
			}); ok {
				fileBrowser := NewFileBrowserView(v.dockerClient, container.ID, container.Name, "/")
				fileBrowser.SetRootScreen(v.rootScreen)
				return switcher.SwitchScreen(fileBrowser)
			}
		}
		return v, nil

	case "I":
		// Inspect container
		if v.selectedContainer < len(v.containers) && v.rootScreen != nil {
			container := v.containers[v.selectedContainer]
			if switcher, ok := v.rootScreen.(interface {
				SwitchScreen(tea.Model) (tea.Model, tea.Cmd)
			}); ok {
				inspectView := NewInspectView(v.dockerClient, container.ID, "container", container.Name)
				inspectView.SetRootScreen(v.rootScreen)
				return switcher.SwitchScreen(inspectView)
			}
		}
		return v, nil

	case "s":
		// Show stats
		if v.rootScreen != nil {
			if switcher, ok := v.rootScreen.(interface {
				SwitchScreen(tea.Model) (tea.Model, tea.Cmd)
			}); ok {
				statsView := NewStatsView(v.dockerClient, v.projectName)
				statsView.SetRootScreen(v.rootScreen)
				return switcher.SwitchScreen(statsView)
			}
		}
		return v, nil

	case "t":
		// Show top (process info)
		if v.rootScreen != nil {
			if switcher, ok := v.rootScreen.(interface {
				SwitchScreen(tea.Model) (tea.Model, tea.Cmd)
			}); ok {
				topView := NewTopView(v.dockerClient, v.projectName)
				topView.SetRootScreen(v.rootScreen)
				return switcher.SwitchScreen(topView)
			}
		}
		return v, nil

	case "p":
		// Switch to Docker container list
		if v.rootScreen != nil {
			if switcher, ok := v.rootScreen.(interface {
				SwitchScreen(tea.Model) (tea.Model, tea.Cmd)
			}); ok {
				dockerView := NewDockerListView(v.dockerClient)
				dockerView.SetRootScreen(v.rootScreen)
				return switcher.SwitchScreen(dockerView)
			}
		}
		return v, nil

	case "i":
		// Switch to image list
		if v.rootScreen != nil {
			if switcher, ok := v.rootScreen.(interface {
				SwitchScreen(tea.Model) (tea.Model, tea.Cmd)
			}); ok {
				imageView := NewImageListView(v.dockerClient)
				imageView.SetRootScreen(v.rootScreen)
				return switcher.SwitchScreen(imageView)
			}
		}
		return v, nil

	case "n":
		// Switch to network list
		if v.rootScreen != nil {
			if switcher, ok := v.rootScreen.(interface {
				SwitchScreen(tea.Model) (tea.Model, tea.Cmd)
			}); ok {
				networkView := NewNetworkListView(v.dockerClient)
				networkView.SetRootScreen(v.rootScreen)
				return switcher.SwitchScreen(networkView)
			}
		}
		return v, nil

	case "P":
		// Switch to project list
		if v.rootScreen != nil {
			if switcher, ok := v.rootScreen.(interface {
				SwitchScreen(tea.Model) (tea.Model, tea.Cmd)
			}); ok {
				projectView := NewProjectListView(v.dockerClient)
				projectView.SetRootScreen(v.rootScreen)
				return switcher.SwitchScreen(projectView)
			}
		}
		return v, nil

	case "?":
		// Show help
		if v.rootScreen != nil {
			if switcher, ok := v.rootScreen.(interface {
				SwitchScreen(tea.Model) (tea.Model, tea.Cmd)
			}); ok {
				helpView := NewHelpView("Docker Compose", v)
				helpView.SetRootScreen(v.rootScreen)
				return switcher.SwitchScreen(helpView)
			}
		}
		return v, nil

	case "q":
		return v, tea.Quit
	}

	return v, nil
}

func (v *ComposeListView) renderComposeList() string {
	var s strings.Builder

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("7")).
		Background(lipgloss.Color("4")).
		Width(v.width).
		Padding(0, 1).
		Render(fmt.Sprintf("Docker Compose - %s", v.projectName))
	s.WriteString(header + "\n")

	// Container list
	if len(v.containers) == 0 {
		s.WriteString("\nNo containers found.\n")
	} else {
		for i, container := range v.containers {
			selected := i == v.selectedContainer
			line := formatContainerLine(container, v.width, selected)
			s.WriteString(line + "\n")
		}
	}

	// Footer
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Width(v.width).
		Align(lipgloss.Center).
		Render("↑/↓: Navigate • Enter: Logs • d: Dind • f: Files • s: Stats • t: Top • p: Docker • P: Projects • q: Quit")

	// Pad to fill screen
	content := s.String()
	lines := strings.Split(content, "\n")
	for len(lines) < v.height-2 {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n") + "\n" + footer
}

// Helper functions
func formatContainerLine(container models.ComposeContainer, width int, selected bool) string {
	status := container.State
	if container.Health != "" {
		status = fmt.Sprintf("%s (%s)", status, container.Health)
	}

	line := fmt.Sprintf("%-30s %-15s %s", container.Name, container.Service, status)
	if len(line) > width-3 {
		line = line[:width-3]
	}

	style := lipgloss.NewStyle()
	if selected {
		style = style.Background(lipgloss.Color("240"))
	}

	return style.Render(line)
}

// Messages
type processesLoadedMsg struct {
	processes []models.ComposeContainer
	err       error
}

// Commands
func loadProcesses(client *docker.Client, projectName string, showAll bool) tea.Cmd {
	return func() tea.Msg {
		processes, err := client.Compose(projectName).ListContainers(showAll)
		return processesLoadedMsg{processes: processes, err: err}
	}
}

// Common view helpers
func renderLoadingView(width, height int, message string) string {
	return lipgloss.Place(width, height,
		lipgloss.Center, lipgloss.Center,
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Render(message),
	)
}

func renderErrorView(width, height int, err error) string {
	return lipgloss.Place(width, height,
		lipgloss.Center, lipgloss.Center,
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")).
			Render(fmt.Sprintf("Error: %v", err)),
	)
}
