package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
)

// ImageListView represents the Docker image list view
type ImageListView struct {
	// View state
	width         int
	height        int
	selectedImage int
	images        []models.DockerImage
	showAll       bool

	// Loading/error state
	loading bool
	err     error

	// Dependencies
	dockerClient *docker.Client
	rootScreen   tea.Model
}

// NewImageListView creates a new image list view
func NewImageListView(dockerClient *docker.Client) *ImageListView {
	return &ImageListView{
		dockerClient: dockerClient,
		showAll:      false,
	}
}

// SetRootScreen sets the root screen reference
func (v *ImageListView) SetRootScreen(root tea.Model) {
	v.rootScreen = root
}

// Init initializes the view
func (v *ImageListView) Init() tea.Cmd {
	v.loading = true
	return loadDockerImages(v.dockerClient, v.showAll)
}

// Update handles messages for this view
func (v *ImageListView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		return v, nil

	case tea.KeyMsg:
		return v.handleKeyPress(msg)

	case dockerImagesLoadedMsg:
		v.loading = false
		if msg.err != nil {
			v.err = msg.err
			return v, nil
		}
		v.images = msg.images
		v.err = nil
		if len(v.images) > 0 && v.selectedImage >= len(v.images) {
			v.selectedImage = 0
		}
		return v, nil

	case RefreshMsg:
		v.loading = true
		v.err = nil
		return v, loadDockerImages(v.dockerClient, v.showAll)

	case serviceActionCompleteMsg:
		if msg.err != nil {
			v.err = msg.err
			return v, nil
		}
		// Refresh after deletion
		v.loading = true
		return v, loadDockerImages(v.dockerClient, v.showAll)
	}

	return v, nil
}

// View renders the image list
func (v *ImageListView) View() string {
	if v.loading {
		return renderLoadingView(v.width, v.height, "Loading Docker images...")
	}

	if v.err != nil {
		return renderErrorView(v.width, v.height, v.err)
	}

	return v.renderImageList()
}

func (v *ImageListView) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if v.selectedImage > 0 {
			v.selectedImage--
		}
		return v, nil

	case "down", "j":
		if v.selectedImage < len(v.images)-1 {
			v.selectedImage++
		}
		return v, nil

	case "r":
		// Send refresh message
		return v, func() tea.Msg { return RefreshMsg{} }

	case "a":
		// Toggle show all
		v.showAll = !v.showAll
		v.loading = true
		return v, loadDockerImages(v.dockerClient, v.showAll)

	case "D":
		// Delete image
		if v.selectedImage < len(v.images) {
			image := v.images[v.selectedImage]
			return v, removeImage(v.dockerClient, image.ID, false)
		}
		return v, nil

	case "F":
		// Force delete image
		if v.selectedImage < len(v.images) {
			image := v.images[v.selectedImage]
			return v, removeImage(v.dockerClient, image.ID, true)
		}
		return v, nil

	case "1":
		// Switch to Docker container list
		if v.rootScreen != nil {
			if switcher, ok := v.rootScreen.(interface {
				SwitchScreen(tea.Model) (tea.Model, tea.Cmd)
			}); ok {
				dockerView := NewDockerListView(v.dockerClient)
				dockerView.SetRootScreen(v.rootScreen)
				return switcher.SwitchScreen(dockerView)
			}
		}
		return v, nil

	case "esc", "q":
		// Go back to Docker container list
		if v.rootScreen != nil {
			if switcher, ok := v.rootScreen.(interface {
				SwitchScreen(tea.Model) (tea.Model, tea.Cmd)
			}); ok {
				dockerView := NewDockerListView(v.dockerClient)
				dockerView.SetRootScreen(v.rootScreen)
				return switcher.SwitchScreen(dockerView)
			}
		}
		return v, nil
	}

	return v, nil
}

func (v *ImageListView) renderImageList() string {
	var s strings.Builder

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("7")).
		Background(lipgloss.Color("4")).
		Width(v.width).
		Padding(0, 1).
		Render("Docker Images")
	s.WriteString(header + "\n")

	// Image list
	if len(v.images) == 0 {
		s.WriteString("\nNo images found.\n")
	} else {
		// Column headers
		headers := fmt.Sprintf("%-30s %-20s %-15s %-20s %s",
			"REPOSITORY", "TAG", "IMAGE ID", "CREATED", "SIZE")
		s.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Bold(true).
			Render(headers) + "\n")

		for i, image := range v.images {
			selected := i == v.selectedImage
			line := formatImageLine(image, v.width, selected)
			s.WriteString(line + "\n")
		}
	}

	// Footer
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Width(v.width).
		Align(lipgloss.Center).
		Render("↑/↓: Navigate • r: Refresh • a: Toggle All • D: Delete • F: Force Delete • 1: Docker • ESC: Back")

	// Pad to fill screen
	content := s.String()
	lines := strings.Split(content, "\n")
	for len(lines) < v.height-2 {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n") + "\n" + footer
}

func formatImageLine(image models.DockerImage, width int, selected bool) string {
	// Truncate long fields
	repo := image.Repository
	if repo == "<none>" {
		repo = image.ID
		if len(repo) > 12 {
			repo = repo[:12]
		}
	}
	if len(repo) > 28 {
		repo = repo[:28]
	}

	tag := image.Tag
	if len(tag) > 18 {
		tag = tag[:18]
	}

	id := image.ID
	if len(id) > 12 {
		id = id[:12]
	}

	line := fmt.Sprintf("%-30s %-20s %-15s %-20s %s",
		repo, tag, id, image.CreatedSince, image.Size)

	if len(line) > width-3 {
		line = line[:width-3]
	}

	style := lipgloss.NewStyle()
	if selected {
		style = style.Background(lipgloss.Color("240"))
	}

	return style.Render(line)
}

// Messages
type dockerImagesLoadedMsg struct {
	images []models.DockerImage
	err    error
}

// Commands
func loadDockerImages(client *docker.Client, showAll bool) tea.Cmd {
	return func() tea.Msg {
		images, err := client.ListImages(showAll)
		return dockerImagesLoadedMsg{images: images, err: err}
	}
}

func removeImage(client *docker.Client, imageID string, force bool) tea.Cmd {
	return func() tea.Msg {
		err := client.RemoveImage(imageID, force)
		return serviceActionCompleteMsg{err: err}
	}
}
