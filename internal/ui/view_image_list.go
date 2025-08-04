package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

type ImageListViewModel struct {
	showAll bool
}

func (m *Model) renderImageList(availableHeight int) string {
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
		Height(availableHeight).
		Width(m.width).
		Offset(m.selectedDockerImage)

	// Configure column widths based on terminal width
	// Approximate widths: REPOSITORY(30), TAG(15), IMAGE ID(12), CREATED(15), SIZE(10)
	availableWidth := m.width - 10 // margin
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
