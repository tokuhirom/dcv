package ui

import (
	"encoding/json"
	"fmt"
	"strings"

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

// ViewType represents the current view
type ViewType int

const (
	ProcessListView ViewType = iota
	LogView
	DindProcessListView
	TopView
	StatsView
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

	// Top view state
	topOutput    string
	topService   string

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

	// Last executed command
	lastCommand string
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