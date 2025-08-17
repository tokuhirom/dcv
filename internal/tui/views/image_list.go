package views

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
)

// ImageListView displays Docker images
type ImageListView struct {
	docker                *docker.Client
	table                 *tview.Table
	dockerImages          []models.DockerImage
	showAll               bool
	pages                 *tview.Pages
	switchToInspectViewFn func(targetID string, target interface{})
}

// NewImageListView creates a new image list view
func NewImageListView(dockerClient *docker.Client) *ImageListView {
	v := &ImageListView{
		docker:  dockerClient,
		table:   tview.NewTable(),
		showAll: false,
		pages:   tview.NewPages(),
	}

	v.setupTable()
	v.setupKeyHandlers()

	return v
}

// setupTable configures the table widget
func (v *ImageListView) setupTable() {
	v.table.SetBorders(false).
		SetSelectable(true, false).
		SetSeparator(' ').
		SetFixed(1, 0)

	// Set header style
	v.table.SetSelectedStyle(tcell.StyleDefault.
		Background(tcell.ColorDarkCyan).
		Foreground(tcell.ColorWhite))
}

// setupKeyHandlers sets up keyboard shortcuts for the view
func (v *ImageListView) setupKeyHandlers() {
	v.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := v.table.GetSelection()

		switch event.Rune() {
		case 'j':
			// Move down (vim style)
			if row < v.table.GetRowCount()-1 {
				v.table.Select(row+1, 0)
			}
			return nil

		case 'k':
			// Move up (vim style)
			if row > 1 { // Skip header row
				v.table.Select(row-1, 0)
			}
			return nil

		case 'g':
			// Go to top (vim style)
			v.table.Select(1, 0)
			return nil

		case 'G':
			// Go to bottom (vim style)
			rowCount := v.table.GetRowCount()
			if rowCount > 1 {
				v.table.Select(rowCount-1, 0)
			}
			return nil

		case 'a', 'A':
			// Toggle show all images
			v.showAll = !v.showAll
			v.Refresh()
			return nil

		case 'd':
			// Delete image
			if row > 0 && row <= len(v.dockerImages) {
				image := v.dockerImages[row-1]
				v.showConfirmation("rmi", image, false)
			}
			return nil

		case 'D':
			// Force delete image
			if row > 0 && row <= len(v.dockerImages) {
				image := v.dockerImages[row-1]
				v.showConfirmation("rmi -f", image, true)
			}
			return nil

		case 'p':
			// Pull image
			if row > 0 && row <= len(v.dockerImages) {
				image := v.dockerImages[row-1]
				go v.pullImage(image)
			}
			return nil

		case 't':
			// Tag image
			if row > 0 && row <= len(v.dockerImages) {
				image := v.dockerImages[row-1]
				// TODO: Show dialog to get new tag
				slog.Info("Tag image", slog.String("image", image.GetRepoTag()))
			}
			return nil

		case 'i':
			// Inspect image
			if row > 0 && row <= len(v.dockerImages) {
				image := v.dockerImages[row-1]
				if v.switchToInspectViewFn != nil {
					v.switchToInspectViewFn(image.ID, image)
				} else {
					slog.Info("Inspect image", slog.String("image", image.GetRepoTag()))
				}
			}
			return nil

		case 'h':
			// Show history
			if row > 0 && row <= len(v.dockerImages) {
				image := v.dockerImages[row-1]
				go v.showHistory(image)
			}
			return nil

		case '/':
			// Search images
			// TODO: Implement search functionality
			slog.Info("Search images")
			return nil

		case 'r':
			// Refresh list
			v.Refresh()
			return nil
		}

		switch event.Key() {
		case tcell.KeyEnter:
			// View image details
			if row > 0 && row <= len(v.dockerImages) {
				image := v.dockerImages[row-1]
				// TODO: Show image details
				slog.Info("View image details", slog.String("image", image.GetRepoTag()))
			}
			return nil

		case tcell.KeyCtrlR:
			// Force refresh
			v.Refresh()
			return nil
		}

		return event
	})
}

// GetPrimitive returns the tview primitive for this view
func (v *ImageListView) GetPrimitive() tview.Primitive {
	if v.pages.HasPage("main") {
		return v.pages
	}
	v.pages.AddPage("main", v.table, true, true)
	return v.pages
}

// Refresh refreshes the image list
func (v *ImageListView) Refresh() {
	go v.loadImages()
}

// GetTitle returns the title of the view
func (v *ImageListView) GetTitle() string {
	if v.showAll {
		return "Docker Images (all)"
	}
	return "Docker Images"
}

// SetSwitchToInspectViewCallback sets the callback for switching to inspect view
func (v *ImageListView) SetSwitchToInspectViewCallback(fn func(targetID string, target interface{})) {
	v.switchToInspectViewFn = fn
}

// loadImages loads the image list from Docker
func (v *ImageListView) loadImages() {
	slog.Info("Loading docker images", slog.Bool("showAll", v.showAll))

	images, err := v.docker.ListImages(v.showAll)
	if err != nil {
		slog.Error("Failed to load docker images", slog.Any("error", err))
		return
	}

	v.dockerImages = images

	// Update table in UI thread
	QueueUpdateDraw(func() {
		v.updateTable()
	})
}

// updateTable updates the table with image data
func (v *ImageListView) updateTable() {
	v.table.Clear()

	// Set headers
	headers := []string{"REPOSITORY", "TAG", "IMAGE ID", "CREATED", "SIZE"}
	for col, header := range headers {
		cell := tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetAttributes(tcell.AttrBold).
			SetSelectable(false)
		v.table.SetCell(0, col, cell)
	}

	// Add image rows
	for row, image := range v.dockerImages {
		// Repository
		repo := image.Repository
		if repo == "<none>" {
			repo = "<none>"
		}
		repoCell := tview.NewTableCell(repo).
			SetTextColor(tcell.ColorWhite)
		v.table.SetCell(row+1, 0, repoCell)

		// Tag
		tag := image.Tag
		if tag == "<none>" {
			tag = "<none>"
		}
		tagCell := tview.NewTableCell(tag).
			SetTextColor(tcell.ColorWhite)
		v.table.SetCell(row+1, 1, tagCell)

		// Image ID (truncated to 12 chars)
		imageID := image.ID
		if len(imageID) > 12 {
			imageID = imageID[:12]
		}
		idCell := tview.NewTableCell(imageID).
			SetTextColor(tcell.ColorBlue)
		v.table.SetCell(row+1, 2, idCell)

		// Created time
		createdCell := tview.NewTableCell(image.CreatedSince).
			SetTextColor(tcell.ColorWhite)
		v.table.SetCell(row+1, 3, createdCell)

		// Size
		sizeCell := tview.NewTableCell(image.Size).
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignRight)
		v.table.SetCell(row+1, 4, sizeCell)
	}

	// Select first row if available
	if len(v.dockerImages) > 0 {
		v.table.Select(1, 0)
	}
}

// Image operations
func (v *ImageListView) deleteImage(image models.DockerImage) {
	slog.Info("Deleting image", slog.String("image", image.GetRepoTag()))

	_, err := docker.ExecuteCaptured("rmi", image.GetRepoTag())
	if err != nil {
		slog.Error("Failed to delete image", slog.Any("error", err))
		// Try with image ID if repo:tag failed
		if strings.Contains(err.Error(), "No such image") {
			_, err = docker.ExecuteCaptured("rmi", image.ID)
			if err != nil {
				slog.Error("Failed to delete image by ID", slog.Any("error", err))
				return
			}
		} else {
			return
		}
	}
	time.Sleep(500 * time.Millisecond)
	v.Refresh()
}

func (v *ImageListView) forceDeleteImage(image models.DockerImage) {
	slog.Info("Force deleting image", slog.String("image", image.GetRepoTag()))

	_, err := docker.ExecuteCaptured("rmi", "-f", image.GetRepoTag())
	if err != nil {
		slog.Error("Failed to force delete image", slog.Any("error", err))
		// Try with image ID if repo:tag failed
		if strings.Contains(err.Error(), "No such image") {
			_, err = docker.ExecuteCaptured("rmi", "-f", image.ID)
			if err != nil {
				slog.Error("Failed to force delete image by ID", slog.Any("error", err))
				return
			}
		} else {
			return
		}
	}
	time.Sleep(500 * time.Millisecond)
	v.Refresh()
}

func (v *ImageListView) pullImage(image models.DockerImage) {
	repoTag := image.GetRepoTag()
	slog.Info("Pulling image", slog.String("image", repoTag))

	// For <none> images, we can't pull them
	if image.Repository == "<none>" {
		slog.Warn("Cannot pull image with <none> repository")
		return
	}

	_, err := docker.ExecuteCaptured("pull", repoTag)
	if err != nil {
		slog.Error("Failed to pull image", slog.Any("error", err))
		return
	}
	time.Sleep(500 * time.Millisecond)
	v.Refresh()
}

func (v *ImageListView) showHistory(image models.DockerImage) {
	slog.Info("Showing image history", slog.String("image", image.GetRepoTag()))

	output, err := docker.ExecuteCaptured("history", image.ID)
	if err != nil {
		slog.Error("Failed to get image history", slog.Any("error", err))
		return
	}

	// TODO: Display history in a dialog or switch to a history view
	slog.Info("Image history", slog.String("output", string(output)))
}

// showConfirmation shows a confirmation dialog for aggressive operations
func (v *ImageListView) showConfirmation(operation string, image models.DockerImage, force bool) {
	var commandText string
	if force {
		commandText = fmt.Sprintf("docker rmi -f %s", image.GetRepoTag())
	} else {
		commandText = fmt.Sprintf("docker rmi %s", image.GetRepoTag())
	}

	modal := tview.NewModal().
		SetText(fmt.Sprintf("Are you sure you want to execute:\n\n%s\n\nImage: %s", commandText, image.GetRepoTag())).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			v.pages.RemovePage("confirm")
			if buttonLabel == "Yes" {
				if force {
					go v.forceDeleteImage(image)
				} else {
					go v.deleteImage(image)
				}
			}
		})

	v.pages.AddPage("confirm", modal, true, true)
}

// GetSelectedImage returns the currently selected image
func (v *ImageListView) GetSelectedImage() *models.DockerImage {
	row, _ := v.table.GetSelection()
	if row > 0 && row <= len(v.dockerImages) {
		return &v.dockerImages[row-1]
	}
	return nil
}

// PruneImages removes unused images
func (v *ImageListView) PruneImages() {
	slog.Info("Pruning unused images")

	output, err := docker.ExecuteCaptured("image", "prune", "-f")
	if err != nil {
		slog.Error("Failed to prune images", slog.Any("error", err))
		return
	}

	slog.Info("Images pruned", slog.String("output", string(output)))
	time.Sleep(500 * time.Millisecond)
	v.Refresh()
}

// SearchImages searches for images in the list
func (v *ImageListView) SearchImages(query string) {
	if query == "" {
		// Reset to show all loaded images
		v.updateTable()
		return
	}

	// Filter images based on query
	var filteredImages []models.DockerImage
	lowerQuery := strings.ToLower(query)

	for _, image := range v.dockerImages {
		if strings.Contains(strings.ToLower(image.Repository), lowerQuery) ||
			strings.Contains(strings.ToLower(image.Tag), lowerQuery) ||
			strings.Contains(strings.ToLower(image.ID), lowerQuery) {
			filteredImages = append(filteredImages, image)
		}
	}

	// Temporarily replace dockerImages for display
	originalImages := v.dockerImages
	v.dockerImages = filteredImages
	v.updateTable()
	v.dockerImages = originalImages
}

// ExportImage exports the selected image to a tar file
func (v *ImageListView) ExportImage(image models.DockerImage, filename string) error {
	slog.Info("Exporting image",
		slog.String("image", image.GetRepoTag()),
		slog.String("filename", filename))

	_, err := docker.ExecuteCaptured("save", "-o", filename, image.GetRepoTag())
	if err != nil {
		return fmt.Errorf("failed to export image: %w", err)
	}

	slog.Info("Image exported successfully")
	return nil
}
