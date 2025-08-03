package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tokuhirom/dcv/internal/docker"
)

// FileContentView represents the file content viewer
type FileContentView struct {
	// View state
	width       int
	height      int
	content     []string
	scrollY     int
	containerID string
	filePath    string
	fileName    string

	// Loading/error state
	loading bool
	err     error

	// Dependencies
	dockerClient *docker.Client
	rootScreen   tea.Model
}

// NewFileContentView creates a new file content view
func NewFileContentView(dockerClient *docker.Client, containerID, filePath, fileName string) *FileContentView {
	return &FileContentView{
		dockerClient: dockerClient,
		containerID:  containerID,
		filePath:     filePath,
		fileName:     fileName,
	}
}

// SetRootScreen sets the root screen reference
func (v *FileContentView) SetRootScreen(root tea.Model) {
	v.rootScreen = root
}

// Init initializes the view
func (v *FileContentView) Init() tea.Cmd {
	v.loading = true
	return loadFileContent(v.dockerClient, v.containerID, v.filePath)
}

// Update handles messages for this view
func (v *FileContentView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		return v, nil

	case tea.KeyMsg:
		return v.handleKeyPress(msg)

	case fileContentLoadedMsg:
		v.loading = false
		if msg.err != nil {
			v.err = msg.err
			return v, nil
		}
		v.content = strings.Split(msg.content, "\n")
		v.err = nil
		v.scrollY = 0
		return v, nil
	}

	return v, nil
}

// View renders the file content
func (v *FileContentView) View() string {
	if v.loading {
		return renderLoadingView(v.width, v.height, "Loading file content...")
	}

	if v.err != nil {
		return renderErrorView(v.width, v.height, v.err)
	}

	return v.renderContent()
}

func (v *FileContentView) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if v.scrollY > 0 {
			v.scrollY--
		}
		return v, nil

	case "down", "j":
		maxScroll := len(v.content) - (v.height - 4)
		if v.scrollY < maxScroll && maxScroll > 0 {
			v.scrollY++
		}
		return v, nil

	case "G":
		// Jump to end
		maxScroll := len(v.content) - (v.height - 4)
		if maxScroll > 0 {
			v.scrollY = maxScroll
		}
		return v, nil

	case "g":
		// Jump to start
		v.scrollY = 0
		return v, nil

	case "?":
		// Show help
		if v.rootScreen != nil {
			if switcher, ok := v.rootScreen.(interface {
				SwitchScreen(tea.Model) (tea.Model, tea.Cmd)
			}); ok {
				helpView := NewHelpView("File Viewer", v)
				helpView.SetRootScreen(v.rootScreen)
				return switcher.SwitchScreen(helpView)
			}
		}
		return v, nil

	case "esc", "q":
		// Go back to file browser
		if v.rootScreen != nil {
			if switcher, ok := v.rootScreen.(interface {
				SwitchScreen(tea.Model) (tea.Model, tea.Cmd)
			}); ok {
				// Extract directory from file path
				lastSlash := strings.LastIndex(v.filePath, "/")
				dir := "/"
				if lastSlash > 0 {
					dir = v.filePath[:lastSlash]
				}

				browserView := NewFileBrowserView(v.dockerClient, v.containerID, "", dir)
				browserView.SetRootScreen(v.rootScreen)
				return switcher.SwitchScreen(browserView)
			}
		}
		return v, nil
	}

	return v, nil
}

func (v *FileContentView) renderContent() string {
	var s strings.Builder

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("7")).
		Background(lipgloss.Color("4")).
		Width(v.width).
		Padding(0, 1).
		Render("File: " + v.fileName)
	s.WriteString(header + "\n")

	// Calculate visible lines
	visibleHeight := v.height - 3
	start := v.scrollY
	end := start + visibleHeight
	if end > len(v.content) {
		end = len(v.content)
	}

	// Render content with line numbers
	for i := start; i < end; i++ {
		if i < len(v.content) {
			lineNum := lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Render(fmt.Sprintf("%4d ", i+1))
			s.WriteString(lineNum + v.content[i] + "\n")
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
		Render("↑/↓: Scroll • G: End • g: Start • ESC: Back")

	return strings.Join(lines[:v.height-1], "\n") + "\n" + footer
}

// Messages
type fileContentLoadedMsg struct {
	content string
	err     error
}

// Commands
func loadFileContent(client *docker.Client, containerID, filePath string) tea.Cmd {
	return func() tea.Msg {
		content, err := client.ReadContainerFile(containerID, filePath)
		return fileContentLoadedMsg{content: content, err: err}
	}
}
