package ui

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
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
	DindComposeProcessListView
	TopView
	StatsView
	ProjectListView
	DockerContainerListView
	ImageListView
	NetworkListView
	FileBrowserView
	FileContentView
	InspectView
	HelpView
)

// Model represents the application state
type Model struct {
	// Current view
	currentView ViewType

	// Docker client
	dockerClient *docker.Client

	// Process list state
	composeContainers []models.ComposeContainer
	selectedContainer int
	showAll           bool // Toggle to show all composeContainers including stopped ones

	// Compose list state
	projects        []models.ComposeProject
	selectedProject int

	// Dind state
	dindContainers         []models.DockerContainer
	selectedDindContainer  int
	currentDindHost        string // Container name (for display)
	currentDindContainerID string // Service name (for docker compose exec)

	// Docker composeContainers state (plain docker ps)
	dockerContainers        []models.DockerContainer
	selectedDockerContainer int

	// Docker images state
	dockerImages        []models.DockerImage
	selectedDockerImage int

	// Docker networks state
	dockerNetworks        []models.DockerNetwork
	selectedDockerNetwork int

	// File browser state
	containerFiles        []models.ContainerFile
	selectedFile          int
	currentPath           string
	browsingContainerID   string
	browsingContainerName string
	pathHistory           []string

	// File content view state
	fileContent     string
	fileContentPath string
	fileScrollY     int

	// Inspect view state
	inspectContent     string
	inspectScrollY     int
	inspectContainerID string
	inspectImageID     string
	inspectNetworkID   string

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
	searchMode        bool
	searchText        string
	searchIgnoreCase  bool
	searchRegex       bool
	searchResults     []int // Line indices of search results
	currentSearchIdx  int   // Current position in searchResults
	searchCursorPos   int   // Cursor position in search text

	// Error state
	err error

	// Window dimensions
	width  int
	height int

	// Loading state
	loading bool

	// Command line options
	projectName string

	// Key handler maps and configurations for all views
	processListViewKeymap   map[string]KeyHandler
	processListViewHandlers []KeyConfig
	logViewKeymap           map[string]KeyHandler
	logViewHandlers         []KeyConfig
	dindListViewKeymap      map[string]KeyHandler
	dindListViewHandlers    []KeyConfig
	topViewKeymap           map[string]KeyHandler
	topViewHandlers         []KeyConfig
	statsViewKeymap         map[string]KeyHandler
	statsViewHandlers       []KeyConfig
	projectListViewKeymap   map[string]KeyHandler
	projectListViewHandlers []KeyConfig
	dockerListViewKeymap    map[string]KeyHandler
	dockerListViewHandlers  []KeyConfig
	imageListViewKeymap     map[string]KeyHandler
	imageListViewHandlers   []KeyConfig
	networkListViewKeymap   map[string]KeyHandler
	networkListViewHandlers []KeyConfig
	fileBrowserKeymap       map[string]KeyHandler
	fileBrowserHandlers     []KeyConfig
	fileContentKeymap       map[string]KeyHandler
	fileContentHandlers     []KeyConfig
	inspectViewKeymap       map[string]KeyHandler
	inspectViewHandlers     []KeyConfig
	helpViewKeymap          map[string]KeyHandler
	helpViewHandlers        []KeyConfig

	// Help view state
	previousView ViewType
	helpScrollY  int

	// Command-line mode state
	commandMode        bool
	commandBuffer      string
	commandHistory     []string
	commandHistoryIdx  int
	commandCursorPos   int
	quitConfirmation   bool
	quitConfirmMessage string
}

func (m *Model) performSearch() {
	m.searchResults = nil
	if m.searchText == "" {
		return
	}

	searchText := m.searchText
	if m.searchIgnoreCase && !m.searchRegex {
		searchText = strings.ToLower(searchText)
	}

	for i, line := range m.logs {
		lineToSearch := line
		if m.searchIgnoreCase && !m.searchRegex {
			lineToSearch = strings.ToLower(line)
		}

		match := false
		if m.searchRegex {
			pattern := searchText
			if m.searchIgnoreCase {
				pattern = "(?i)" + pattern
			}
			if re, err := regexp.Compile(pattern); err == nil {
				match = re.MatchString(line)
			}
		} else {
			match = strings.Contains(lineToSearch, searchText)
		}

		if match {
			m.searchResults = append(m.searchResults, i)
		}
	}

	// If we have results, jump to the first one
	if len(m.searchResults) > 0 && m.currentSearchIdx < len(m.searchResults) {
		targetLine := m.searchResults[m.currentSearchIdx]
		m.logScrollY = targetLine - m.height/2 + 3
		if m.logScrollY < 0 {
			m.logScrollY = 0
		}
	}
}

func (m *Model) performInspectSearch() {
	m.searchResults = nil
	if m.searchText == "" {
		return
	}

	// Split inspect content into lines for searching
	lines := strings.Split(m.inspectContent, "\n")
	
	searchText := m.searchText
	if m.searchIgnoreCase && !m.searchRegex {
		searchText = strings.ToLower(searchText)
	}

	for i, line := range lines {
		lineToSearch := line
		if m.searchIgnoreCase && !m.searchRegex {
			lineToSearch = strings.ToLower(line)
		}

		match := false
		if m.searchRegex {
			pattern := searchText
			if m.searchIgnoreCase {
				pattern = "(?i)" + pattern
			}
			if re, err := regexp.Compile(pattern); err == nil {
				match = re.MatchString(line)
			}
		} else {
			match = strings.Contains(lineToSearch, searchText)
		}

		if match {
			m.searchResults = append(m.searchResults, i)
		}
	}

	// If we have results, jump to the first one
	if len(m.searchResults) > 0 && m.currentSearchIdx < len(m.searchResults) {
		targetLine := m.searchResults[m.currentSearchIdx]
		m.inspectScrollY = targetLine - m.height/2 + 3
		if m.inspectScrollY < 0 {
			m.inspectScrollY = 0
		}
	}
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
func (m *Model) Init() tea.Cmd {
	m.initializeKeyHandlers()

	if m.currentView == ProjectListView {
		return tea.Batch(
			loadProjects(m.dockerClient),
			tea.WindowSize(),
		)
	}

	// Otherwise, try to load composeContainers first - if it fails due to a missing compose file,
	// we'll switch to the project list view in the update
	return tea.Batch(
		loadProcesses(m.dockerClient, m.projectName, m.showAll),
		tea.WindowSize(),
	)
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
	action  string
	service string
	err     error
}

type upActionCompleteMsg struct {
	action string
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

func pauseService(client *docker.Client, containerID string) tea.Cmd {
	return func() tea.Msg {
		err := client.PauseContainer(containerID)
		return serviceActionCompleteMsg{
			action:  "pause",
			service: containerID,
			err:     err,
		}
	}
}

func unpauseService(client *docker.Client, containerID string) tea.Cmd {
	return func() tea.Msg {
		err := client.UnpauseContainer(containerID)
		return serviceActionCompleteMsg{
			action:  "unpause",
			service: containerID,
			err:     err,
		}
	}
}

func up(client *docker.Client, projectName string) tea.Cmd {
	return func() tea.Msg {
		err := client.Compose(projectName).Up()
		return upActionCompleteMsg{
			action: "up -d",
			err:    err,
		}
	}
}

func down(client *docker.Client, projectName string) tea.Cmd {
	return func() tea.Msg {
		err := client.Compose(projectName).Down()
		return upActionCompleteMsg{
			action: "down",
			err:    err,
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

func loadDockerContainers(client *docker.Client, showAll bool) tea.Cmd {
	return func() tea.Msg {
		containers, err := client.ListAllContainers(showAll)
		return dockerContainersLoadedMsg{
			containers: containers,
			err:        err,
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
			action:  "remove image",
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
			action:  "remove network",
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

func loadInspect(client *docker.Client, containerID string) tea.Cmd {
	return func() tea.Msg {
		content, err := client.InspectContainer(containerID)
		return inspectLoadedMsg{
			content: content,
			err:     err,
		}
	}
}

func loadImageInspect(client *docker.Client, imageID string) tea.Cmd {
	return func() tea.Msg {
		content, err := client.InspectImage(imageID)
		return inspectLoadedMsg{
			content: content,
			err:     err,
		}
	}
}

func loadNetworkInspect(client *docker.Client, networkID string) tea.Cmd {
	return func() tea.Msg {
		content, err := client.InspectNetwork(networkID)
		return inspectLoadedMsg{
			content: content,
			err:     err,
		}
	}
}
