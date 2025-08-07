package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

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
	var s strings.Builder

	// No images
	if len(m.dockerImages) == 0 {
		s.WriteString("No images found.\n")
		s.WriteString("Press 'r' to refresh.\n")
		return s.String()
	}

	// Create table
	t := table.New().
		Headers("REPOSITORY", "TAG", "IMAGE ID", "CREATED", "SIZE").
		Height(availableHeight - 2).
		Width(model.width).
		Offset(m.selectedDockerImage)

	// Configure column widths based on terminal width
	// Approximate widths: REPOSITORY(30), TAG(15), IMAGE ID(12), CREATED(15), SIZE(10)
	availableWidth := model.width - 10 // margin
	repoWidth := 30
	if availableWidth < 100 {
		repoWidth = 20
	}
	t.Width(availableWidth).
		StyleFunc(func(row, col int) lipgloss.Style {
			return lipgloss.NewStyle()
		})

	// Styles
	selectedStyle := lipgloss.NewStyle().Background(lipgloss.Color("238"))
	repoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	tagStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("229"))
	idStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	createdStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	sizeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("213"))

	// Add rows
	for i, image := range m.dockerImages {
		var rowRepoStyle, rowTagStyle, rowIdStyle, rowCreatedStyle, rowSizeStyle lipgloss.Style

		if i == m.selectedDockerImage {
			rowRepoStyle = selectedStyle.Inherit(repoStyle)
			rowTagStyle = selectedStyle.Inherit(tagStyle)
			rowIdStyle = selectedStyle.Inherit(idStyle)
			rowCreatedStyle = selectedStyle.Inherit(createdStyle)
			rowSizeStyle = selectedStyle.Inherit(sizeStyle)
		} else {
			rowRepoStyle = repoStyle
			rowTagStyle = tagStyle
			rowIdStyle = idStyle
			rowCreatedStyle = createdStyle
			rowSizeStyle = sizeStyle
		}

		// Handle <none> repository
		repo := image.Repository
		if repo == "<none>" {
			repo = "<none>"
		}
		if len(repo) > repoWidth {
			repo = repo[:repoWidth-3] + "..."
		}
		repo = rowRepoStyle.Render(repo)

		tag := rowTagStyle.Render(image.Tag)

		// Show first 12 chars of ID
		id := image.ID
		if len(id) > 12 {
			id = id[:12]
		}
		id = rowIdStyle.Render(id)

		created := rowCreatedStyle.Render(image.CreatedSince)
		size := rowSizeStyle.Render(image.Size)

		t.Row(repo, tag, id, created, size)
	}

	s.WriteString(t.Render() + "\n")

	return s.String()
}

// Show switches to the image list view
func (m *ImageListViewModel) Show(model *Model) tea.Cmd {
	model.currentView = ImageListView
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
	model.loading = true
	return removeImage(model.dockerClient, image.GetRepoTag(), false)
}

// HandleForceDelete forcefully removes the selected image
func (m *ImageListViewModel) HandleForceDelete(model *Model) tea.Cmd {
	if len(m.dockerImages) == 0 || m.selectedDockerImage >= len(m.dockerImages) {
		return nil
	}
	image := m.dockerImages[m.selectedDockerImage]
	model.loading = true
	return removeImage(model.dockerClient, image.GetRepoTag(), true)
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
