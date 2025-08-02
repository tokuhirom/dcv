package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

func (m *Model) renderImageList() string {
	var s strings.Builder

	// Title
	title := "Docker Images"
	if m.showAll {
		title += " (All)"
	}
	s.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		Render(title) + "\n\n")

	// Loading or error state
	if m.loading {
		s.WriteString("Loading images...\n")
		return s.String()
	}

	if m.err != nil {
		s.WriteString(fmt.Sprintf("Error: %v\n", m.err))
		return s.String()
	}

	// No images
	if len(m.dockerImages) == 0 {
		s.WriteString("No images found.\n")
		s.WriteString("Press 'r' to refresh.\n")
		return s.String()
	}

	// Create table
	t := table.New()
	t.Headers("REPOSITORY", "TAG", "IMAGE ID", "CREATED", "SIZE")

	// Configure column widths based on terminal width
	// Approximate widths: REPOSITORY(30), TAG(15), IMAGE ID(12), CREATED(15), SIZE(10)
	availableWidth := m.width - 10 // margin
	repoWidth := 30
	if availableWidth < 100 {
		repoWidth = 20
	}
	t.Width(availableWidth).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("99"))
			}
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

	s.WriteString(t.Render() + "\n\n")

	// Help text
	helpText := m.GetStyledHelpText()
	if helpText != "" {
		s.WriteString(helpText)
	}

	return s.String()
}
