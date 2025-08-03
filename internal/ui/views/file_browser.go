package views

import (
	"fmt"
	"path"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
)

// FileBrowserView represents the container file browser view
type FileBrowserView struct {
	// View state
	width         int
	height        int
	selectedFile  int
	files         []models.FileInfo
	currentPath   string
	containerID   string
	containerName string

	// Loading/error state
	loading bool
	err     error

	// Dependencies
	dockerClient *docker.Client
	rootScreen   tea.Model
}

// NewFileBrowserView creates a new file browser view
func NewFileBrowserView(dockerClient *docker.Client, containerID, containerName, initialPath string) *FileBrowserView {
	if initialPath == "" {
		initialPath = "/"
	}
	return &FileBrowserView{
		dockerClient:  dockerClient,
		containerID:   containerID,
		containerName: containerName,
		currentPath:   initialPath,
	}
}

// SetRootScreen sets the root screen reference
func (v *FileBrowserView) SetRootScreen(root tea.Model) {
	v.rootScreen = root
}

// Init initializes the view
func (v *FileBrowserView) Init() tea.Cmd {
	v.loading = true
	return loadFiles(v.dockerClient, v.containerID, v.currentPath)
}

// Update handles messages for this view
func (v *FileBrowserView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		return v, nil

	case tea.KeyMsg:
		return v.handleKeyPress(msg)

	case filesLoadedMsg:
		v.loading = false
		if msg.err != nil {
			v.err = msg.err
			return v, nil
		}
		v.files = msg.files
		v.err = nil
		v.selectedFile = 0
		return v, nil

	case RefreshMsg:
		v.loading = true
		v.err = nil
		return v, loadFiles(v.dockerClient, v.containerID, v.currentPath)
	}

	return v, nil
}

// View renders the file browser
func (v *FileBrowserView) View() string {
	if v.loading {
		return renderLoadingView(v.width, v.height, "Loading files...")
	}

	if v.err != nil {
		return renderErrorView(v.width, v.height, v.err)
	}

	return v.renderFileBrowser()
}

func (v *FileBrowserView) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if v.selectedFile > 0 {
			v.selectedFile--
		}
		return v, nil

	case "down", "j":
		if v.selectedFile < len(v.files)-1 {
			v.selectedFile++
		}
		return v, nil

	case "enter":
		// Enter directory or view file
		if v.selectedFile < len(v.files) {
			file := v.files[v.selectedFile]
			if file.IsDir {
				// Navigate into directory
				newPath := path.Join(v.currentPath, file.Name)
				v.currentPath = newPath
				v.loading = true
				return v, loadFiles(v.dockerClient, v.containerID, newPath)
			} else if v.rootScreen != nil {
				// View file content
				if switcher, ok := v.rootScreen.(interface {
					SwitchScreen(tea.Model) (tea.Model, tea.Cmd)
				}); ok {
					filePath := path.Join(v.currentPath, file.Name)
					fileView := NewFileContentView(v.dockerClient, v.containerID, filePath, file.Name)
					fileView.SetRootScreen(v.rootScreen)
					return switcher.SwitchScreen(fileView)
				}
			}
		}
		return v, nil

	case "u":
		// Go up one directory
		if v.currentPath != "/" {
			v.currentPath = path.Dir(v.currentPath)
			v.loading = true
			return v, loadFiles(v.dockerClient, v.containerID, v.currentPath)
		}
		return v, nil

	case "r":
		// Send refresh message
		return v, func() tea.Msg { return RefreshMsg{} }

	case "?":
		// Show help
		if v.rootScreen != nil {
			if switcher, ok := v.rootScreen.(interface {
				SwitchScreen(tea.Model) (tea.Model, tea.Cmd)
			}); ok {
				helpView := NewHelpView("File Browser", v)
				helpView.SetRootScreen(v.rootScreen)
				return switcher.SwitchScreen(helpView)
			}
		}
		return v, nil

	case "esc", "q":
		// Go back
		// TODO: Navigate back to the appropriate view
		return v, tea.Quit
	}

	return v, nil
}

func (v *FileBrowserView) renderFileBrowser() string {
	var s strings.Builder

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("7")).
		Background(lipgloss.Color("4")).
		Width(v.width).
		Padding(0, 1).
		Render(fmt.Sprintf("Files - %s:%s", v.containerName, v.currentPath))
	s.WriteString(header + "\n")

	// File list
	if len(v.files) == 0 {
		s.WriteString("\nNo files found.\n")
	} else {
		// Column headers
		headers := fmt.Sprintf("%-40s %-10s %-20s %s",
			"NAME", "SIZE", "MODIFIED", "PERMISSIONS")
		s.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Bold(true).
			Render(headers) + "\n")

		// Calculate visible range
		visibleHeight := v.height - 5
		start := 0
		if v.selectedFile >= visibleHeight {
			start = v.selectedFile - visibleHeight + 1
		}
		end := start + visibleHeight
		if end > len(v.files) {
			end = len(v.files)
		}

		for i := start; i < end; i++ {
			selected := i == v.selectedFile
			line := formatFileLine(v.files[i], v.width, selected)
			s.WriteString(line + "\n")
		}
	}

	// Pad to fill screen
	lines := strings.Split(s.String(), "\n")
	for len(lines) < v.height-1 {
		lines = append(lines, "")
	}

	// Footer
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Width(v.width).
		Align(lipgloss.Center).
		Render("↑/↓: Navigate • Enter: Open • u: Parent Dir • r: Refresh • ESC: Back")

	return strings.Join(lines[:v.height-1], "\n") + "\n" + footer
}

func formatFileLine(file models.FileInfo, width int, selected bool) string {
	// Format name with directory indicator
	name := file.Name
	if file.IsDir {
		name = name + "/"
	}
	if len(name) > 38 {
		name = name[:38]
	}

	// Format size
	size := file.Size
	if file.IsDir {
		size = "<DIR>"
	} else if len(size) > 8 {
		size = size[:8]
	}

	// Format modified time
	modified := file.ModTime
	if len(modified) > 18 {
		modified = modified[:18]
	}

	line := fmt.Sprintf("%-40s %-10s %-20s %s",
		name, size, modified, file.Mode)

	if len(line) > width-3 {
		line = line[:width-3]
	}

	style := lipgloss.NewStyle()
	if selected {
		style = style.Background(lipgloss.Color("240"))
	}

	// Color directories differently
	if file.IsDir {
		style = style.Foreground(lipgloss.Color("4")) // Blue for directories
	}

	return style.Render(line)
}

// Messages
type filesLoadedMsg struct {
	files []models.FileInfo
	err   error
}

// Commands
func loadFiles(client *docker.Client, containerID, path string) tea.Cmd {
	return func() tea.Msg {
		files, err := client.ListFiles(containerID, path)
		return filesLoadedMsg{files: files, err: err}
	}
}