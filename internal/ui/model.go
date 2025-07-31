package ui

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
)

// ContainerStats holds resource usage statistics for a container
type ContainerStats struct {
	Container   string  `json:"Container"`
	Name        string  `json:"Name"`
	Service     string  `json:"Service"`
	CPUPerc     string  `json:"CPUPerc"`
	MemUsage    string  `json:"MemUsage"`
	MemPerc     string  `json:"MemPerc"`
	NetIO       string  `json:"NetIO"`
	BlockIO     string  `json:"BlockIO"`
	PIDs        string  `json:"PIDs"`
}

// CommandLog represents a command execution log entry
type CommandLog struct {
	Timestamp time.Time
	Command   string
	ExitCode  int
	Output    string
	Error     string
	Duration  time.Duration
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
	DebugLogView
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
	showAll         bool // Toggle to show all containers including stopped ones

	// Project list state
	projects         []models.ComposeProject
	selectedProject  int
	showProjectList  bool // Show project list when no compose file

	// Dind state
	dindContainers       []models.Container
	selectedDindContainer int
	currentDindHost      string  // Container name (for display)
	currentDindService   string  // Service name (for docker compose exec)

	// Log view state
	logs           []string
	logScrollY     int
	containerName  string
	isDindLog      bool
	hostContainer  string

	// Top view state
	topOutput    string
	topService   string

	// Stats view state
	stats []ContainerStats

	// Debug log view state
	commandLogs       []CommandLog
	sharedCommandLogs []docker.CommandLog // Shared across all docker clients
	debugLogScrollY   int

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

	// Last executed command
	lastCommand string

	// Command line options
	projectName string
	composeFile string
}

// NewModel creates a new model with initial state
func NewModel() Model {
	sharedLogs := make([]docker.CommandLog, 0)
	client := docker.NewComposeClient("")
	client.SetCommandLogs(&sharedLogs)
	
	return Model{
		currentView:       ProcessListView,
		dockerClient:      client,
		loading:           true,
		sharedCommandLogs: sharedLogs,
	}
}

// NewModelWithOptions creates a new model with command line options
func NewModelWithOptions(projectName, composeFile string, showProjects bool) Model {
	sharedLogs := make([]docker.CommandLog, 0)
	client := docker.NewComposeClientWithOptions("", projectName, composeFile)
	client.SetCommandLogs(&sharedLogs)
	
	// Determine initial view
	initialView := ProcessListView
	if showProjects {
		initialView = ProjectListView
	}
	
	return Model{
		currentView:       initialView,
		dockerClient:      client,
		loading:           true,
		projectName:       projectName,
		composeFile:       composeFile,
		showProjectList:   showProjects,
		sharedCommandLogs: sharedLogs,
	}
}

// Init returns an initial command for the application
func (m Model) Init() tea.Cmd {
	// If showProjectList is true, start with project list
	if m.showProjectList {
		return tea.Batch(
			loadProjects(m.dockerClient),
			tea.WindowSize(),
		)
	}
	
	// Otherwise, try to load processes first - if it fails due to missing compose file,
	// we'll switch to project list view in the update
	return tea.Batch(
		loadProcesses(m.dockerClient, m.showAll),
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

type commandExecutedMsg struct {
	command string
}

type topLoadedMsg struct {
	output string
	err    error
}

type serviceActionCompleteMsg struct {
	action string
	service string
	err    error
}

type statsLoadedMsg struct {
	stats []ContainerStats
	err   error
}

type projectsLoadedMsg struct {
	projects []models.ComposeProject
	err      error
}

type commandLogsMsg struct {
	logs []CommandLog
}

// Commands

func loadProcesses(client *docker.ComposeClient, showAll bool) tea.Cmd {
	return func() tea.Msg {
		processes, err := client.ListContainers(showAll)
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

func streamLogs(client *docker.ComposeClient, serviceName string, isDind bool, hostService string) tea.Cmd {
	return streamLogsReal(client, serviceName, isDind, hostService)
}

func loadTop(client *docker.ComposeClient, serviceName string) tea.Cmd {
	return func() tea.Msg {
		output, err := client.GetContainerTop(serviceName)
		return topLoadedMsg{
			output: output,
			err:    err,
		}
	}
}

func killService(client *docker.ComposeClient, serviceName string) tea.Cmd {
	return func() tea.Msg {
		err := client.KillService(serviceName)
		return serviceActionCompleteMsg{
			action: "kill",
			service: serviceName,
			err:    err,
		}
	}
}

func stopService(client *docker.ComposeClient, serviceName string) tea.Cmd {
	return func() tea.Msg {
		err := client.StopService(serviceName)
		return serviceActionCompleteMsg{
			action: "stop",
			service: serviceName,
			err:    err,
		}
	}
}

func startService(client *docker.ComposeClient, serviceName string) tea.Cmd {
	return func() tea.Msg {
		err := client.StartService(serviceName)
		return serviceActionCompleteMsg{
			action: "start",
			service: serviceName,
			err:    err,
		}
	}
}

func restartService(client *docker.ComposeClient, serviceName string) tea.Cmd {
	return func() tea.Msg {
		err := client.RestartService(serviceName)
		return serviceActionCompleteMsg{
			action: "restart",
			service: serviceName,
			err:    err,
		}
	}
}

func removeService(client *docker.ComposeClient, serviceName string) tea.Cmd {
	return func() tea.Msg {
		err := client.RemoveService(serviceName)
		return serviceActionCompleteMsg{
			action: "rm",
			service: serviceName,
			err:    err,
		}
	}
}

func upService(client *docker.ComposeClient, serviceName string) tea.Cmd {
	return func() tea.Msg {
		err := client.UpService(serviceName)
		return serviceActionCompleteMsg{
			action: "up -d",
			service: serviceName,
			err:    err,
		}
	}
}
func loadStats(client *docker.ComposeClient) tea.Cmd {
	return func() tea.Msg {
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

func loadProjects(client *docker.ComposeClient) tea.Cmd {
	return func() tea.Msg {
		projects, err := client.ListProjects()
		return projectsLoadedMsg{
			projects: projects,
			err:      err,
		}
	}
}

func loadCommandLogs(client *docker.ComposeClient) tea.Cmd {
	return func() tea.Msg {
		// Convert docker.CommandLog to ui.CommandLog
		dockerLogs := client.GetCommandLogs()
		uiLogs := make([]CommandLog, len(dockerLogs))
		for i, log := range dockerLogs {
			uiLogs[i] = CommandLog{
				Timestamp: log.Timestamp,
				Command:   log.Command,
				ExitCode:  log.ExitCode,
				Output:    log.Output,
				Error:     log.Error,
				Duration:  log.Duration,
			}
		}
		return commandLogsMsg{logs: uiLogs}
	}
}