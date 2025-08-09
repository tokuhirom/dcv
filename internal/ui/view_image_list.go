package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/tokuhirom/dcv/internal/models"
)

// ImageListViewModel manages the state and rendering of the Docker image list view
type ImageListViewModel struct {
	dockerImages        []models.DockerImage
	selectedDockerImage int
	showAll             bool
}

func (m *ImageListViewModel) Title() string {
	if m.showAll {
		return "Docker Images (all)"
	}
	return "Docker Images"
}

// render renders the image list view
func (m *ImageListViewModel) render(model *Model, availableHeight int) string {
	// No images
	if len(m.dockerImages) == 0 {
		var s strings.Builder
		s.WriteString("No images found.\n")
		s.WriteString("Press 'r' to refresh.\n")
		return s.String()
	}

	availableWidth := model.width - 10 // margin
	repoWidth := 30
	if availableWidth < 100 {
		repoWidth = 20
	}

	// Create columns
	columns := []table.Column{
		{Title: "REPOSITORY", Width: repoWidth},
		{Title: "TAG", Width: 15},
		{Title: "IMAGE ID", Width: 12},
		{Title: "CREATED", Width: 20},
		{Title: "SIZE", Width: 10},
	}

	// Create rows
	rows := make([]table.Row, len(m.dockerImages))
	for i, image := range m.dockerImages {
		// Handle <none> repository
		repo := image.Repository
		if len(repo) > repoWidth {
			repo = repo[:repoWidth-3] + "..."
		}

		// Show first 12 chars of ID
		id := image.ID
		if len(id) > 12 {
			id = id[:12]
		}

		rows[i] = table.Row{
			repo,
			image.Tag,
			id,
			image.CreatedSince,
			image.Size,
		}
	}

	return RenderTable(columns, rows, availableHeight, m.selectedDockerImage)
}

// Show switches to the image list view
func (m *ImageListViewModel) Show(model *Model) tea.Cmd {
	model.SwitchView(ImageListView)
	model.loading = true
	return loadDockerImages(model.dockerClient, m.showAll)
}

// HandleUp moves selection up in the image list
func (m *ImageListViewModel) HandleUp() tea.Cmd {
	if m.selectedDockerImage > 0 {
		m.selectedDockerImage--
	}
	return nil
}

// HandleDown moves selection down in the image list
func (m *ImageListViewModel) HandleDown() tea.Cmd {
	if m.selectedDockerImage < len(m.dockerImages)-1 {
		m.selectedDockerImage++
	}
	return nil
}

// HandleToggleAll toggles showing all images including intermediate layers
func (m *ImageListViewModel) HandleToggleAll(model *Model) tea.Cmd {
	m.showAll = !m.showAll
	model.loading = true
	return loadDockerImages(model.dockerClient, m.showAll)
}

// HandleDelete removes the selected image
func (m *ImageListViewModel) HandleDelete(model *Model) tea.Cmd {
	if len(m.dockerImages) == 0 || m.selectedDockerImage >= len(m.dockerImages) {
		return nil
	}
	image := m.dockerImages[m.selectedDockerImage]
	// Use CommandExecutionView to show real-time output
	args := []string{"rmi", image.GetRepoTag()}
	return model.commandExecutionViewModel.ExecuteCommand(model, true, args...) // rmi is aggressive
}

// HandleInspect shows the inspect view for the selected image
func (m *ImageListViewModel) HandleInspect(model *Model) tea.Cmd {
	if len(m.dockerImages) == 0 || m.selectedDockerImage >= len(m.dockerImages) {
		return nil
	}

	image := m.dockerImages[m.selectedDockerImage]
	return model.inspectViewModel.InspectImage(model, image)
}

// HandleBack returns to the compose process list view
func (m *ImageListViewModel) HandleBack(model *Model) tea.Cmd {
	model.SwitchToPreviousView()
	return nil
}

// HandleRefresh reloads the image list
func (m *ImageListViewModel) HandleRefresh(model *Model) tea.Cmd {
	model.loading = true
	return loadDockerImages(model.dockerClient, m.showAll)
}

// Loaded updates the image list after loading
func (m *ImageListViewModel) Loaded(images []models.DockerImage) {
	m.dockerImages = images
	if len(m.dockerImages) > 0 && m.selectedDockerImage >= len(m.dockerImages) {
		m.selectedDockerImage = 0
	}
}
