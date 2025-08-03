package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
)

// ProjectListView represents the Docker Compose project list view
type ProjectListView struct {
	// View state
	width           int
	height          int
	selectedProject int
	projects        []models.ComposeProject

	// Loading/error state
	loading bool
	err     error

	// Dependencies
	dockerClient *docker.Client
	rootScreen   tea.Model
}

// NewProjectListView creates a new project list view
func NewProjectListView(dockerClient *docker.Client) *ProjectListView {
	return &ProjectListView{
		dockerClient: dockerClient,
	}
}

// SetRootScreen sets the root screen reference
func (v *ProjectListView) SetRootScreen(root tea.Model) {
	v.rootScreen = root
}

// Init initializes the view
func (v *ProjectListView) Init() tea.Cmd {
	v.loading = true
	return loadProjects(v.dockerClient)
}

// Update handles messages for this view
func (v *ProjectListView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		return v, nil

	case tea.KeyMsg:
		return v.handleKeyPress(msg)

	case projectsLoadedMsg:
		v.loading = false
		if msg.err != nil {
			v.err = msg.err
			return v, nil
		}
		v.projects = msg.projects
		v.err = nil
		if len(v.projects) > 0 && v.selectedProject >= len(v.projects) {
			v.selectedProject = 0
		}
		return v, nil

	case RefreshMsg:
		v.loading = true
		v.err = nil
		return v, loadProjects(v.dockerClient)
	}

	return v, nil
}

// View renders the project list
func (v *ProjectListView) View() string {
	if v.loading {
		return renderLoadingView(v.width, v.height, "Loading Docker Compose projects...")
	}

	if v.err != nil {
		return renderErrorView(v.width, v.height, v.err)
	}

	return v.renderProjectList()
}

func (v *ProjectListView) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if v.selectedProject > 0 {
			v.selectedProject--
		}
		return v, nil

	case "down", "j":
		if v.selectedProject < len(v.projects)-1 {
			v.selectedProject++
		}
		return v, nil

	case "enter":
		// Switch to compose list view for selected project
		if v.selectedProject < len(v.projects) && v.rootScreen != nil {
			project := v.projects[v.selectedProject]
			if switcher, ok := v.rootScreen.(interface {
				SwitchScreen(tea.Model) (tea.Model, tea.Cmd)
			}); ok {
				composeView := NewComposeListView(v.dockerClient, project.Name)
				composeView.SetRootScreen(v.rootScreen)
				return switcher.SwitchScreen(composeView)
			}
		}
		return v, nil

	case "r":
		// Send refresh message
		return v, func() tea.Msg { return RefreshMsg{} }

	case "1":
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

	case "q":
		return v, tea.Quit
	}

	return v, nil
}

func (v *ProjectListView) renderProjectList() string {
	var s strings.Builder

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("7")).
		Background(lipgloss.Color("4")).
		Width(v.width).
		Padding(0, 1).
		Render("Docker Compose Projects")
	s.WriteString(header + "\n")

	// Project list
	if len(v.projects) == 0 {
		s.WriteString("\nNo Docker Compose projects found.\n")
	} else {
		// Column headers
		headers := fmt.Sprintf("%-30s %-15s %s",
			"NAME", "STATUS", "CONFIG FILES")
		s.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Bold(true).
			Render(headers) + "\n")

		for i, project := range v.projects {
			selected := i == v.selectedProject
			line := formatProjectLine(project, v.width, selected)
			s.WriteString(line + "\n")
		}
	}

	// Footer
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Width(v.width).
		Align(lipgloss.Center).
		Render("↑/↓: Navigate • Enter: View Project • r: Refresh • 1: Docker • q: Quit")

	// Pad to fill screen
	content := s.String()
	lines := strings.Split(content, "\n")
	for len(lines) < v.height-2 {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n") + "\n" + footer
}

func formatProjectLine(project models.ComposeProject, width int, selected bool) string {
	// Truncate long fields
	name := project.Name
	if len(name) > 28 {
		name = name[:28]
	}

	configFiles := project.ConfigFiles
	if len(configFiles) > 18 {
		configFiles = configFiles[:18]
	}

	line := fmt.Sprintf("%-30s %-15s %s",
		name, project.Status, configFiles)

	if len(line) > width-3 {
		line = line[:width-3]
	}

	style := lipgloss.NewStyle()
	if selected {
		style = style.Background(lipgloss.Color("240"))
	}

	// Color based on status
	if strings.Contains(strings.ToLower(project.Status), "running") {
		style = style.Foreground(lipgloss.Color("2")) // Green
	} else {
		style = style.Foreground(lipgloss.Color("240")) // Gray
	}

	return style.Render(line)
}

// Messages
type projectsLoadedMsg struct {
	projects []models.ComposeProject
	err      error
}

// Commands
func loadProjects(client *docker.Client) tea.Cmd {
	return func() tea.Msg {
		projects, err := client.ListComposeProjects()
		return projectsLoadedMsg{projects: projects, err: err}
	}
}
