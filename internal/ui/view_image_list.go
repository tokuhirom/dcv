package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tokuhirom/dcv/internal/models"
)

type dockerImagesLoadedMsg struct {
	images []models.DockerImage
	err    error
}

var _ HandleInspectAware = (*ImageListViewModel)(nil)
var _ UpdateAware = (*ImageListViewModel)(nil)

// ImageListViewModel manages the state and rendering of the Docker image list view
type ImageListViewModel struct {
	TableViewModel
	dockerImages []models.DockerImage
	showAll      bool
}

func (m *ImageListViewModel) Update(model *Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case dockerImagesLoadedMsg:
		model.loading = false
		if msg.err != nil {
			model.err = msg.err
		} else {
			model.err = nil
			m.Loaded(model, msg.images)
		}
		return model, nil

	default:
		return model, nil
	}
}

func (m *ImageListViewModel) Loaded(model *Model, images []models.DockerImage) {
	m.dockerImages = images
	m.SetRows(m.buildRows(), model.ViewHeight())
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
	cellWidth := (availableWidth - repoWidth) / 5

	// Create columns
	columns := []table.Column{
		{Title: "REPOSITORY", Width: repoWidth},
		{Title: "TAG", Width: cellWidth},
		{Title: "IMAGE ID", Width: cellWidth},
		{Title: "CREATED", Width: cellWidth},
		{Title: "SIZE", Width: cellWidth},
	}

	return m.RenderTable(model, columns, availableHeight, func(row, col int) lipgloss.Style {
		var base lipgloss.Style
		if row == m.Cursor {
			base = tableSelectedCellStyle
		} else {
			base = tableNormalCellStyle
		}
		if col == 4 {
			base = base.Align(lipgloss.Right)
		}
		return base
	})
}

func (m *ImageListViewModel) buildRows() []table.Row {
	// Create rows
	rows := make([]table.Row, len(m.dockerImages))
	for i, image := range m.dockerImages {
		// Handle <none> repository
		repo := image.Repository

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
	return rows
}

// Show switches to the image list view
func (m *ImageListViewModel) Show(model *Model) tea.Cmd {
	model.SwitchView(ImageListView)
	return m.DoLoad(model)
}

func (m *ImageListViewModel) DoLoad(model *Model) tea.Cmd {
	model.loading = true
	return func() tea.Msg {
		images, err := model.dockerClient.ListImages(m.showAll)
		return dockerImagesLoadedMsg{
			images: images,
			err:    err,
		}
	}
}

// HandleToggleAll toggles showing all images including intermediate layers
func (m *ImageListViewModel) HandleToggleAll(model *Model) tea.Cmd {
	m.showAll = !m.showAll
	return m.DoLoad(model)
}

// HandleDelete removes the selected image
func (m *ImageListViewModel) HandleDelete(model *Model) tea.Cmd {
	if len(m.dockerImages) == 0 || m.Cursor >= len(m.dockerImages) {
		return nil
	}
	image := m.dockerImages[m.Cursor]
	// Use CommandExecutionView to show real-time output
	args := []string{"rmi", image.GetRepoTag()}
	return model.commandExecutionViewModel.ExecuteCommand(model, true, args...) // rmi is aggressive
}

// HandleInspect shows the inspect view for the selected image
func (m *ImageListViewModel) HandleInspect(model *Model) tea.Cmd {
	if len(m.dockerImages) == 0 || m.Cursor >= len(m.dockerImages) {
		return nil
	}

	image := m.dockerImages[m.Cursor]
	return model.inspectViewModel.Inspect(model, fmt.Sprintf("Image: %s:%s %s", image.Repository, image.Tag, image.ID), func() ([]byte, error) {
		return model.dockerClient.ExecuteCaptured("image", "inspect", image.ID)
	})
}

// HandleBack returns to the compose process list view
func (m *ImageListViewModel) HandleBack(model *Model) tea.Cmd {
	model.SwitchToPreviousView()
	return nil
}
