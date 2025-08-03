package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/ui/views"
)

// RootScreen is the main model that manages view switching
type RootScreen struct {
	// Current view model
	model tea.Model

	// Shared dependencies
	dockerClient *docker.Client

	// Global state
	width  int
	height int

	// Command mode state (shared across views)
	commandMode bool
	commandText string
}

// NewRootScreen creates a new root screen
func NewRootScreen(dockerClient *docker.Client, initialView ViewType, projectName string) *RootScreen {
	root := &RootScreen{
		dockerClient: dockerClient,
	}

	// Create initial view
	var initialModel tea.Model
	switch initialView {
	case ComposeProcessListView:
		composeView := views.NewComposeListView(dockerClient, projectName)
		composeView.SetRootScreen(root)
		initialModel = composeView
	case DockerContainerListView:
		dockerView := views.NewDockerListView(dockerClient)
		dockerView.SetRootScreen(root)
		initialModel = dockerView
	case ProjectListView:
		projectView := views.NewProjectListView(dockerClient)
		projectView.SetRootScreen(root)
		initialModel = projectView
	default:
		dockerView := views.NewDockerListView(dockerClient)
		dockerView.SetRootScreen(root)
		initialModel = dockerView
	}

	root.model = initialModel
	return root
}

// Init initializes the root screen
func (r *RootScreen) Init() tea.Cmd {
	return r.model.Init()
}

// Update handles messages for the root screen
func (r *RootScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		r.width = msg.Width
		r.height = msg.Height
		// Pass size to current model
		if sizable, ok := r.model.(interface{ SetSize(int, int) }); ok {
			sizable.SetSize(r.width, r.height)
		}

	case tea.KeyMsg:
		// Handle global keys
		switch msg.String() {
		case ":":
			if !r.commandMode {
				r.commandMode = true
				r.commandText = ""
				return r, nil
			}
		case "esc":
			if r.commandMode {
				r.commandMode = false
				r.commandText = ""
				return r, nil
			}
		}

		// If in command mode, handle command input
		if r.commandMode {
			return r.handleCommandMode(msg)
		}
	}

	// Delegate to current view
	updatedModel, cmd := r.model.Update(msg)
	r.model = updatedModel
	return r, cmd
}

// View renders the current view
func (r *RootScreen) View() string {
	view := r.model.View()

	// TODO: Add command line rendering if in command mode

	return view
}

// SwitchScreen switches to a new screen model
func (r *RootScreen) SwitchScreen(model tea.Model) (tea.Model, tea.Cmd) {
	r.model = model

	// Set size on new model if it supports it
	if sizable, ok := r.model.(interface{ SetSize(int, int) }); ok {
		sizable.SetSize(r.width, r.height)
	}

	// Initialize the new screen
	return r, r.model.Init()
}

func (r *RootScreen) handleCommandMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		// Execute command
		r.commandMode = false
		cmd := r.commandText
		r.commandText = ""
		return r.executeCommand(cmd)

	case tea.KeyBackspace:
		if len(r.commandText) > 0 {
			r.commandText = r.commandText[:len(r.commandText)-1]
		}

	case tea.KeyRunes:
		r.commandText += string(msg.Runes)
	}

	return r, nil
}

func (r *RootScreen) executeCommand(cmd string) (tea.Model, tea.Cmd) {
	// Handle commands
	switch cmd {
	case "q", "quit":
		return r, tea.Quit
	case "q!", "quit!":
		return r, tea.Quit
		// TODO: Add more commands
	}

	return r, nil
}
