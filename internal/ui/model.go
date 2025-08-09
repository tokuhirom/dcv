package ui

import (
	"io"
	"log/slog"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
)

// ViewType represents the current view
type ViewType int

const (
	ComposeProcessListView ViewType = iota
	LogView
	DindProcessListView
	TopView
	StatsView
	ComposeProjectListView
	DockerContainerListView
	ImageListView
	NetworkListView
	VolumeListView
	FileBrowserView
	FileContentView
	InspectView
	HelpView
	CommandExecutionView
)

func (view ViewType) String() string {
	switch view {
	case ComposeProcessListView:
		return "Compose Process List"
	case LogView:
		return "Log View"
	case DindProcessListView:
		return "Docker in Docker"
	case TopView:
		return "Process Info"
	case StatsView:
		return "Container Stats"
	case ComposeProjectListView:
		return "Project List"
	case DockerContainerListView:
		return "Docker Containers"
	case ImageListView:
		return "Docker Images"
	case NetworkListView:
		return "Docker Networks"
	case VolumeListView:
		return "Docker Volumes"
	case FileBrowserView:
		return "File Browser"
	case FileContentView:
		return "File Content"
	case InspectView:
		return "Inspect"
	case HelpView:
		return "Help"
	case CommandExecutionView:
		return "Command Execution"
	default:
		return "Unknown View"
	}
}

// Model represents the application state
type Model struct {
	// Current view
	currentView ViewType
	viewHistory []ViewType

	// Docker client
	dockerClient *docker.Client

	dockerContainerListViewModel DockerContainerListViewModel
	logViewModel                 LogViewModel
	commandExecutionViewModel    CommandExecutionViewModel
	fileBrowserViewModel         FileBrowserViewModel
	inspectViewModel             InspectViewModel
	composeProjectListViewModel  ComposeProjectListViewModel
	composeProcessListViewModel  ComposeProcessListViewModel
	topViewModel                 TopViewModel
	dindProcessListViewModel     DindProcessListViewModel
	imageListViewModel           ImageListViewModel
	fileContentViewModel         FileContentViewModel
	helpViewModel                HelpViewModel
	networkListViewModel         NetworkListViewModel
	statsViewModel               StatsViewModel
	volumeListViewModel          VolumeListViewModel

	// Error state
	err error

	// Window dimensions
	width  int
	Height int

	// Loading state
	loading bool

	globalKeymap   map[string]KeyHandler
	globalHandlers []KeyConfig

	// Key handler maps and configurations for all views
	composeProcessListViewKeymap    map[string]KeyHandler
	composeProcessListViewHandlers  []KeyConfig
	logViewKeymap                   map[string]KeyHandler
	logViewHandlers                 []KeyConfig
	dindListViewKeymap              map[string]KeyHandler
	dindListViewHandlers            []KeyConfig
	topViewKeymap                   map[string]KeyHandler
	topViewHandlers                 []KeyConfig
	statsViewKeymap                 map[string]KeyHandler
	statsViewHandlers               []KeyConfig
	composeProjectListViewKeymap    map[string]KeyHandler
	composeProjectListViewHandlers  []KeyConfig
	dockerListViewKeymap            map[string]KeyHandler
	dockerContainerListViewHandlers []KeyConfig
	imageListViewKeymap             map[string]KeyHandler
	imageListViewHandlers           []KeyConfig
	networkListViewKeymap           map[string]KeyHandler
	networkListViewHandlers         []KeyConfig
	volumeListViewKeymap            map[string]KeyHandler
	volumeListViewHandlers          []KeyConfig
	fileBrowserKeymap               map[string]KeyHandler
	fileBrowserHandlers             []KeyConfig
	fileContentKeymap               map[string]KeyHandler
	fileContentHandlers             []KeyConfig
	inspectViewKeymap               map[string]KeyHandler
	inspectViewHandlers             []KeyConfig
	helpViewKeymap                  map[string]KeyHandler
	helpViewHandlers                []KeyConfig
	commandExecKeymap               map[string]KeyHandler
	commandExecHandlers             []KeyConfig

	// Command-line mode state
	commandViewModel CommandViewModel

	quitConfirmation bool

	// Command registry - stores all available commands
	viewCommandRegistry map[ViewType]map[string]CommandHandler
	allCommands         map[string]struct{}
}

// NewModel creates a new model with initial state
func NewModel(initialView ViewType) *Model {
	client := docker.NewClient()

	slog.Info("Creating new model",
		slog.String("initial_view", initialView.String()))

	m := &Model{
		currentView:  initialView,
		dockerClient: client,
		loading:      true,
	}

	return m
}

// Init returns an initial command for the application
func (m *Model) Init() tea.Cmd {
	m.initializeKeyHandlers()

	switch m.currentView {
	case ComposeProjectListView:
		return tea.Batch(
			loadProjects(m.dockerClient),
			tea.WindowSize(),
		)
	case DockerContainerListView:
		return tea.Batch(
			func() tea.Msg {
				return RefreshMsg{}
			},
			tea.WindowSize(),
		)
	default:
		// Otherwise, try to load composeContainers first - if it fails due to a missing compose file,
		// we'll switch to the project list view in the update
		return tea.Batch(
			func() tea.Msg {
				return RefreshMsg{}
			},
			tea.WindowSize(),
		)
	}
}

func (m *Model) SwitchView(view ViewType) {
	if view == m.currentView {
		slog.Info("SwitchView called with the same view, ignoring",
			slog.String("view", view.String()))
		return
	}

	m.viewHistory = append(m.viewHistory, m.currentView)
	m.currentView = view
}

func (m *Model) SwitchToPreviousView() {
	m.err = nil

	for len(m.viewHistory) > 0 && m.viewHistory[len(m.viewHistory)-1] == m.currentView {
		// Remove consecutive duplicates from the history
		m.viewHistory = m.viewHistory[:len(m.viewHistory)-1]
	}

	if len(m.viewHistory) == 0 {
		slog.Info("No previous view to switch to, staying in current view",
			slog.String("current_view", m.currentView.String()))
		return
	}

	previousView := m.viewHistory[len(m.viewHistory)-1]
	slog.Info("Switching to previous view",
		slog.String("previous_view", previousView.String()))
	m.viewHistory = m.viewHistory[:len(m.viewHistory)-1] // Remove the last
	m.currentView = previousView
}

// GetViewKeyHandlers returns the key handlers for the specified view
func (m *Model) GetViewKeyHandlers(view ViewType) []KeyConfig {
	switch view {
	case ComposeProcessListView:
		return m.composeProcessListViewHandlers
	case LogView:
		return m.logViewHandlers
	case DindProcessListView:
		return m.dindListViewHandlers
	case TopView:
		return m.topViewHandlers
	case StatsView:
		return m.statsViewHandlers
	case ComposeProjectListView:
		return m.composeProjectListViewHandlers
	case DockerContainerListView:
		return m.dockerContainerListViewHandlers
	case ImageListView:
		return m.imageListViewHandlers
	case NetworkListView:
		return m.networkListViewHandlers
	case VolumeListView:
		return m.volumeListViewHandlers
	case FileBrowserView:
		return m.fileBrowserHandlers
	case FileContentView:
		return m.fileContentHandlers
	case InspectView:
		return m.inspectViewHandlers
	case HelpView:
		return m.helpViewHandlers
	case CommandExecutionView:
		return m.commandExecHandlers
	default:
		return nil
	}
}

// GetViewKeymap returns the keymap for the specified view
func (m *Model) GetViewKeymap(view ViewType) map[string]KeyHandler {
	switch view {
	case ComposeProcessListView:
		return m.composeProcessListViewKeymap
	case LogView:
		return m.logViewKeymap
	case DindProcessListView:
		return m.dindListViewKeymap
	case TopView:
		return m.topViewKeymap
	case StatsView:
		return m.statsViewKeymap
	case ComposeProjectListView:
		return m.composeProjectListViewKeymap
	case DockerContainerListView:
		return m.dockerListViewKeymap
	case ImageListView:
		return m.imageListViewKeymap
	case NetworkListView:
		return m.networkListViewKeymap
	case VolumeListView:
		return m.volumeListViewKeymap
	case FileBrowserView:
		return m.fileBrowserKeymap
	case FileContentView:
		return m.fileContentKeymap
	case InspectView:
		return m.inspectViewKeymap
	case HelpView:
		return m.helpViewKeymap
	case CommandExecutionView:
		return m.commandExecKeymap
	default:
		return nil
	}
}

// Messages

type processesLoadedMsg struct {
	processes []models.ComposeContainer
	err       error
}

type dindContainersLoadedMsg struct {
	containers []models.DockerContainer
	err        error
}

type logLinesMsg struct {
	lines []string
}

type pollLogsContinueMsg struct{}

type errorMsg struct {
	err error
}

type commandExecutedMsg struct {
	command string
}

type topLoadedMsg struct {
	output string
	err    error
}

type statsLoadedMsg struct {
	stats []models.ContainerStats
	err   error
}

type projectsLoadedMsg struct {
	projects []models.ComposeProject
	err      error
}

type dockerContainersLoadedMsg struct {
	containers []models.DockerContainer
	err        error
}

type dockerImagesLoadedMsg struct {
	images []models.DockerImage
	err    error
}

type dockerNetworksLoadedMsg struct {
	networks []models.DockerNetwork
	err      error
}

type dockerVolumesLoadedMsg struct {
	volumes []models.DockerVolume
	err     error
}

type containerFilesLoadedMsg struct {
	files []models.ContainerFile
	err   error
}

type fileContentLoadedMsg struct {
	content string
	path    string
	err     error
}

type executeCommandMsg struct {
	containerID string
	command     []string
}

type inspectLoadedMsg struct {
	content    string
	err        error
	targetName string
}

type commandExecOutputMsg struct {
	line string
}

type commandExecCompleteMsg struct {
	exitCode int
}

type commandExecStartedMsg struct {
	cmd    *exec.Cmd
	stdout io.ReadCloser
	stderr io.ReadCloser
}

// Commands

func loadProjects(client *docker.Client) tea.Cmd {
	return func() tea.Msg {
		projects, err := client.ListComposeProjects()
		return projectsLoadedMsg{
			projects: projects,
			err:      err,
		}
	}
}

func executeInteractiveCommand(containerID string, command []string) tea.Cmd {
	return func() tea.Msg {
		return executeCommandMsg{
			containerID: containerID,
			command:     command,
		}
	}
}
