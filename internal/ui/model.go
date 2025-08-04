package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
)

// ContainerStats holds resource usage statistics for a container
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
}

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

	// File browser state
	containerFiles        []models.ContainerFile
	selectedFile          int
	currentPath           string
	browsingContainerID   string
	browsingContainerName string
	pathHistory           []string

	// Search state
	searchMode       bool
	searchText       string
	searchIgnoreCase bool
	searchRegex      bool
	searchResults    []int // Line indices of search results
	currentSearchIdx int   // Current position in searchResults
	searchCursorPos  int   // Cursor position in search text

	// Error state
	err error

	// Window dimensions
	width  int
	Height int

	// Loading state
	loading bool

	// Command line options
	projectName string // TODO: Make this a part of the model?

	// Key handler maps and configurations for all views
	processListViewKeymap           map[string]KeyHandler
	processListViewHandlers         []KeyConfig
	logViewKeymap                   map[string]KeyHandler
	logViewHandlers                 []KeyConfig
	dindListViewKeymap              map[string]KeyHandler
	dindListViewHandlers            []KeyConfig
	topViewKeymap                   map[string]KeyHandler
	topViewHandlers                 []KeyConfig
	statsViewKeymap                 map[string]KeyHandler
	statsViewHandlers               []KeyConfig
	projectListViewKeymap           map[string]KeyHandler
	projectListViewHandlers         []KeyConfig
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
	commandMode        bool
	commandBuffer      string
	commandHistory     []string
	commandHistoryIdx  int
	commandCursorPos   int
	quitConfirmation   bool
	quitConfirmMessage string
}

// NewModel creates a new model with initial state
func NewModel(initialView ViewType, projectName string) Model {
	client := docker.NewClient()

	slog.Info("Creating new model",
		slog.String("initial_view", initialView.String()))

	return Model{
		currentView:  initialView,
		dockerClient: client,
		loading:      true,
		projectName:  projectName,
	}
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
			loadDockerContainers(m.dockerClient, m.dockerContainerListViewModel.showAll),
			tea.WindowSize(),
		)
	default:
		// Otherwise, try to load composeContainers first - if it fails due to a missing compose file,
		// we'll switch to the project list view in the update
		return tea.Batch(
			loadProcesses(m.dockerClient, m.projectName, m.dockerContainerListViewModel.showAll),
			tea.WindowSize(),
		)
	}
}

func (m *Model) CmdCancel(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case CommandExecutionView:
		return m, m.commandExecutionViewModel.HandleCancel()
	default:
		slog.Info("Cancel command not implemented for current view",
			slog.String("current_view", m.currentView.String()))
		return m, nil
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

type logLineMsg struct {
	line string
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

type serviceActionCompleteMsg struct {
	service string
	err     error
}

type upActionCompleteMsg struct {
	err error
}

type statsLoadedMsg struct {
	stats []ContainerStats
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
	content string
	err     error
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

func loadProcesses(client *docker.Client, projectName string, showAll bool) tea.Cmd {
	return func() tea.Msg {
		slog.Info("Loading composeContainers",
			slog.Bool("showAll", showAll))
		processes, err := client.Compose(projectName).ListContainers(showAll)
		return processesLoadedMsg{
			processes: processes,
			err:       err,
		}
	}
}

func loadTop(client *docker.Client, projectName, serviceName string) tea.Cmd {
	return func() tea.Msg {
		output, err := client.Compose(projectName).GetContainerTop(serviceName)
		return topLoadedMsg{
			output: output,
			err:    err,
		}
	}
}

func removeService(client *docker.Client, containerID string) tea.Cmd {
	return func() tea.Msg {
		err := client.RemoveContainer(containerID)
		return serviceActionCompleteMsg{
			service: containerID,
			err:     err,
		}
	}
}

func pauseService(client *docker.Client, containerID string) tea.Cmd {
	return func() tea.Msg {
		err := client.PauseContainer(containerID)
		return serviceActionCompleteMsg{
			service: containerID,
			err:     err,
		}
	}
}

func unpauseService(client *docker.Client, containerID string) tea.Cmd {
	return func() tea.Msg {
		err := client.UnpauseContainer(containerID)
		return serviceActionCompleteMsg{
			service: containerID,
			err:     err,
		}
	}
}

func loadStats(client *docker.Client) tea.Cmd {
	return func() tea.Msg {
		// TODO: support periodic update
		output, err := client.GetStats()
		if err != nil {
			return statsLoadedMsg{
				stats: nil,
				err:   err,
			}
		}

		// Parse JSON lines format
		var stats []ContainerStats
		lines := strings.Split(strings.TrimSpace(output), "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}

			var stat ContainerStats
			if err := json.Unmarshal([]byte(line), &stat); err != nil {
				return statsLoadedMsg{
					stats: nil,
					err:   fmt.Errorf("failed to parse stats JSON: %w", err),
				}
			}
			stats = append(stats, stat)
		}

		return statsLoadedMsg{
			stats: stats,
			err:   nil,
		}
	}
}

func loadProjects(client *docker.Client) tea.Cmd {
	return func() tea.Msg {
		projects, err := client.ListComposeProjects()
		return projectsLoadedMsg{
			projects: projects,
			err:      err,
		}
	}
}

func loadDockerImages(client *docker.Client, showAll bool) tea.Cmd {
	return func() tea.Msg {
		images, err := client.ListImages(showAll)
		return dockerImagesLoadedMsg{
			images: images,
			err:    err,
		}
	}
}

func removeImage(client *docker.Client, imageID string, force bool) tea.Cmd {
	return func() tea.Msg {
		err := client.RemoveImage(imageID, force)
		return serviceActionCompleteMsg{
			service: imageID,
			err:     err,
		}
	}
}

func loadDockerNetworks(client *docker.Client) tea.Cmd {
	return func() tea.Msg {
		networks, err := client.ListNetworks()
		return dockerNetworksLoadedMsg{
			networks: networks,
			err:      err,
		}
	}
}

func removeNetwork(client *docker.Client, networkID string) tea.Cmd {
	return func() tea.Msg {
		err := client.RemoveNetwork(networkID)
		return serviceActionCompleteMsg{
			service: networkID,
			err:     err,
		}
	}
}

func loadContainerFiles(client *docker.Client, containerID, path string) tea.Cmd {
	return func() tea.Msg {
		files, err := client.ListContainerFiles(containerID, path)
		return containerFilesLoadedMsg{
			files: files,
			err:   err,
		}
	}
}

func loadFileContent(client *docker.Client, containerID, path string) tea.Cmd {
	return func() tea.Msg {
		content, err := client.ReadContainerFile(containerID, path)
		return fileContentLoadedMsg{
			content: content,
			path:    path,
			err:     err,
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
