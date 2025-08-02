package ui

import (
	"encoding/json"
	"fmt"
	"log/slog"
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
	ProcessListView ViewType = iota
	LogView
	DindProcessListView
	TopView
	StatsView
	ProjectListView
)

// Model represents the application state
type Model struct {
	// Current view
	currentView ViewType

	// Docker client
	dockerClient *docker.Client

	// Process list state
	containers        []models.Container
	selectedContainer int
	showAll           bool // Toggle to show all containers including stopped ones

	// Compose list state
	projects        []models.ComposeProject
	selectedProject int

	// Dind state
	dindContainers         []models.Container
	selectedDindContainer  int
	currentDindHost        string // Container name (for display)
	currentDindContainerID string // Service name (for docker compose exec)

	// Log view state
	logs          []string
	logScrollY    int
	containerName string
	isDindLog     bool
	hostContainer string

	// Top view state
	topOutput  string
	topService string

	// Stats view state
	stats []ContainerStats

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

	// Command line options
	projectName string
}

// NewModel creates a new model with initial state
func NewModel(initialView ViewType, projectName string) Model {
	client := docker.NewClient()

	return Model{
		currentView:  initialView,
		dockerClient: client,
		loading:      true,
		projectName:  projectName,
	}
}

// Init returns an initial command for the application
func (m Model) Init() tea.Cmd {
	if m.currentView == ProjectListView {
		return tea.Batch(
			loadProjects(m.dockerClient),
			tea.WindowSize(),
		)
	}

	// Otherwise, try to load containers first - if it fails due to a missing compose file,
	// we'll switch to the project list view in the update
	return tea.Batch(
		loadProcesses(m.dockerClient, m.projectName, m.showAll),
		tea.WindowSize(),
	)
}

// Messages

type processesLoadedMsg struct {
	processes []models.Container
	err       error
}

type dindContainersLoadedMsg struct {
	containers []models.Container
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
	action  string
	service string
	err     error
}

type statsLoadedMsg struct {
	stats []ContainerStats
	err   error
}

type projectsLoadedMsg struct {
	projects []models.ComposeProject
	err      error
}

// Commands

func loadProcesses(client *docker.Client, projectName string, showAll bool) tea.Cmd {
	return func() tea.Msg {
		slog.Info("Loading containers",
			slog.Bool("showAll", showAll))
		processes, err := client.Compose(projectName).ListContainers(showAll)
		return processesLoadedMsg{
			processes: processes,
			err:       err,
		}
	}
}

func loadDindContainers(client *docker.Client, containerID string) tea.Cmd {
	return func() tea.Msg {
		containers, err := client.ListDindContainers(containerID)
		return dindContainersLoadedMsg{
			containers: containers,
			err:        err,
		}
	}
}

func streamLogs(client *docker.Client, serviceName string, isDind bool, hostService string) tea.Cmd {
	return streamLogsReal(client, serviceName, isDind, hostService)
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

func killService(client *docker.Client, containerID string) tea.Cmd {
	return func() tea.Msg {
		err := client.KillContainer(containerID)
		return serviceActionCompleteMsg{
			action:  "kill",
			service: containerID,
			err:     err,
		}
	}
}

func stopService(client *docker.Client, containerID string) tea.Cmd {
	return func() tea.Msg {
		err := client.StopContainer(containerID)
		return serviceActionCompleteMsg{
			action:  "stop",
			service: containerID,
			err:     err,
		}
	}
}

func startService(client *docker.Client, serviceName string) tea.Cmd {
	return func() tea.Msg {
		err := client.StartContainer(serviceName)
		return serviceActionCompleteMsg{
			action:  "start",
			service: serviceName,
			err:     err,
		}
	}
}

func restartService(client *docker.Client, containerID string) tea.Cmd {
	return func() tea.Msg {
		err := client.RestartContainer(containerID)
		return serviceActionCompleteMsg{
			action:  "restart",
			service: containerID,
			err:     err,
		}
	}
}

func removeService(client *docker.Client, containerID string) tea.Cmd {
	return func() tea.Msg {
		err := client.RemoveContainer(containerID)
		return serviceActionCompleteMsg{
			action:  "rm",
			service: containerID,
			err:     err,
		}
	}
}

func upService(client *docker.Client, projectName, serviceName string) tea.Cmd {
	return func() tea.Msg {
		err := client.Compose(projectName).UpService(serviceName)
		return serviceActionCompleteMsg{
			action:  "up -d",
			service: serviceName,
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
