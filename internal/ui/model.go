package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
)

// ViewType represents the current view
type ViewType int

const (
	ProcessListView ViewType = iota
	LogView
	DindProcessListView
)

// Model represents the application state
type Model struct {
	// Current view
	currentView ViewType

	// Docker client
	dockerClient *docker.ComposeClient

	// Process list state
	processes       []models.Process
	selectedProcess int

	// Dind state
	dindContainers       []models.Container
	selectedDindContainer int
	currentDindHost      string

	// Log view state
	logs           []string
	logScrollY     int
	containerName  string
	isDindLog      bool
	hostContainer  string

	// Search state
	searchMode bool
	searchText string

	// Error state
	err error

	// Window dimensions
	width  int
	height int

	// Loading state
	loading bool
}

// NewModel creates a new model with initial state
func NewModel() Model {
	return Model{
		currentView:  ProcessListView,
		dockerClient: docker.NewComposeClient(""),
		loading:      true,
	}
}

// Init returns an initial command for the application
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		loadProcesses(m.dockerClient),
		tea.WindowSize(),
	)
}

// Messages

type processesLoadedMsg struct {
	processes []models.Process
	err       error
}

type dindContainersLoadedMsg struct {
	containers []models.Container
	err        error
}

type logLineMsg struct {
	line string
}

type errorMsg struct {
	err error
}

// Commands

func loadProcesses(client *docker.ComposeClient) tea.Cmd {
	return func() tea.Msg {
		processes, err := client.ListContainers()
		return processesLoadedMsg{
			processes: processes,
			err:       err,
		}
	}
}

func loadDindContainers(client *docker.ComposeClient, containerName string) tea.Cmd {
	return func() tea.Msg {
		containers, err := client.ListDindContainers(containerName)
		return dindContainersLoadedMsg{
			containers: containers,
			err:        err,
		}
	}
}

func streamLogs(client *docker.ComposeClient, containerName string, isDind bool, hostContainer string) tea.Cmd {
	return streamLogsReal(client, containerName, isDind, hostContainer)
}