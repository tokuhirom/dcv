package views

import (
	"testing"

	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
)

func createTestDockerImages() []models.DockerImage {
	return []models.DockerImage{
		{
			ID:           "sha256:abc123def456",
			Repository:   "nginx",
			Tag:          "latest",
			CreatedSince: "2 weeks ago",
			Size:         "142MB",
			CreatedAt:    "2024-01-01 10:00:00 +0000 UTC",
		},
		{
			ID:           "sha256:789012345678",
			Repository:   "alpine",
			Tag:          "3.18",
			CreatedSince: "3 days ago",
			Size:         "7.34MB",
			CreatedAt:    "2024-01-15 14:30:00 +0000 UTC",
		},
		{
			ID:           "sha256:fedcba987654",
			Repository:   "<none>",
			Tag:          "<none>",
			CreatedSince: "1 month ago",
			Size:         "256MB",
			CreatedAt:    "2023-12-01 08:00:00 +0000 UTC",
		},
	}
}

func TestNewImageListView(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewImageListView(dockerClient)

	assert.NotNil(t, view)
	assert.NotNil(t, view.docker)
	assert.NotNil(t, view.table)
	assert.False(t, view.showAll)
	assert.Empty(t, view.dockerImages)
}

func TestImageListView_GetPrimitive(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewImageListView(dockerClient)

	primitive := view.GetPrimitive()

	assert.NotNil(t, primitive)
	_, ok := primitive.(*tview.Pages)
	assert.True(t, ok)
}

func TestImageListView_GetTitle(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewImageListView(dockerClient)

	// Test without showAll
	title := view.GetTitle()
	assert.Equal(t, "Docker Images", title)

	// Test with showAll
	view.showAll = true
	title = view.GetTitle()
	assert.Equal(t, "Docker Images (all)", title)
}

func TestImageListView_UpdateTable(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewImageListView(dockerClient)
	view.dockerImages = createTestDockerImages()

	// Create a test app
	app := tview.NewApplication()
	SetApp(app)

	// Update the table
	view.updateTable()

	// Check headers
	headerCell := view.table.GetCell(0, 0)
	assert.NotNil(t, headerCell)
	assert.Equal(t, "REPOSITORY", headerCell.Text)

	headerCell = view.table.GetCell(0, 1)
	assert.NotNil(t, headerCell)
	assert.Equal(t, "TAG", headerCell.Text)

	headerCell = view.table.GetCell(0, 2)
	assert.NotNil(t, headerCell)
	assert.Equal(t, "IMAGE ID", headerCell.Text)

	headerCell = view.table.GetCell(0, 3)
	assert.NotNil(t, headerCell)
	assert.Equal(t, "CREATED", headerCell.Text)

	headerCell = view.table.GetCell(0, 4)
	assert.NotNil(t, headerCell)
	assert.Equal(t, "SIZE", headerCell.Text)

	// Check image data
	repoCell := view.table.GetCell(1, 0)
	assert.NotNil(t, repoCell)
	assert.Equal(t, "nginx", repoCell.Text)

	tagCell := view.table.GetCell(1, 1)
	assert.NotNil(t, tagCell)
	assert.Equal(t, "latest", tagCell.Text)

	idCell := view.table.GetCell(1, 2)
	assert.NotNil(t, idCell)
	assert.Equal(t, "sha256:abc12", idCell.Text) // Truncated to 12 chars

	createdCell := view.table.GetCell(1, 3)
	assert.NotNil(t, createdCell)
	assert.Equal(t, "2 weeks ago", createdCell.Text)

	sizeCell := view.table.GetCell(1, 4)
	assert.NotNil(t, sizeCell)
	assert.Equal(t, "142MB", sizeCell.Text)

	// Check second image
	repoCell2 := view.table.GetCell(2, 0)
	assert.NotNil(t, repoCell2)
	assert.Equal(t, "alpine", repoCell2.Text)

	tagCell2 := view.table.GetCell(2, 1)
	assert.NotNil(t, tagCell2)
	assert.Equal(t, "3.18", tagCell2.Text)

	// Check <none> repository image
	repoCell3 := view.table.GetCell(3, 0)
	assert.NotNil(t, repoCell3)
	assert.Equal(t, "<none>", repoCell3.Text)

	tagCell3 := view.table.GetCell(3, 1)
	assert.NotNil(t, tagCell3)
	assert.Equal(t, "<none>", tagCell3.Text)

	// Verify table selection
	row, _ := view.table.GetSelection()
	assert.Equal(t, 1, row) // First data row should be selected
}

func TestImageListView_ShowAll(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewImageListView(dockerClient)

	// Initially showAll is false
	assert.False(t, view.showAll)

	// Toggle showAll
	view.showAll = true
	assert.True(t, view.showAll)

	// Toggle back
	view.showAll = false
	assert.False(t, view.showAll)
}

func TestImageListView_TableProperties(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewImageListView(dockerClient)

	// Check table properties
	selectable, selectableByRow := view.table.GetSelectable()
	assert.True(t, selectable)
	assert.False(t, selectableByRow)

	// Check that table is configured correctly
	assert.NotNil(t, view.table)
}

func TestImageListView_EmptyImageList(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewImageListView(dockerClient)
	view.dockerImages = []models.DockerImage{}

	// Create a test app
	app := tview.NewApplication()
	SetApp(app)

	// Update the table with empty image list
	view.updateTable()

	// Should still have headers
	headerCell := view.table.GetCell(0, 0)
	assert.NotNil(t, headerCell)
	assert.Equal(t, "REPOSITORY", headerCell.Text)

	// Check table row count - should only have header row
	rowCount := view.table.GetRowCount()
	assert.Equal(t, 1, rowCount) // Only header row
}

func TestImageListView_GetSelectedImage(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewImageListView(dockerClient)
	view.dockerImages = createTestDockerImages()

	// Create a test app
	app := tview.NewApplication()
	SetApp(app)

	// Update the table
	view.updateTable()

	// First row should be selected by default
	row, _ := view.table.GetSelection()
	assert.Equal(t, 1, row)

	// Get the selected image
	selectedImage := view.GetSelectedImage()
	assert.NotNil(t, selectedImage)
	assert.Equal(t, "nginx", selectedImage.Repository)
	assert.Equal(t, "latest", selectedImage.Tag)

	// Move selection down
	view.table.Select(2, 0)
	row, _ = view.table.GetSelection()
	assert.Equal(t, 2, row)

	// Get the new selected image
	selectedImage = view.GetSelectedImage()
	assert.NotNil(t, selectedImage)
	assert.Equal(t, "alpine", selectedImage.Repository)
	assert.Equal(t, "3.18", selectedImage.Tag)
}

func TestImageListView_SearchImages(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewImageListView(dockerClient)
	view.dockerImages = createTestDockerImages()

	// Create a test app
	app := tview.NewApplication()
	SetApp(app)

	// Search for "nginx"
	view.SearchImages("nginx")

	// Table should show filtered results
	// Note: The actual filtering happens in SearchImages but we can't easily test
	// the display without more complex mocking

	// Search for "alpine"
	view.SearchImages("alpine")

	// Test empty search (should reset)
	view.SearchImages("")

	// Verify original images are still intact
	assert.Len(t, view.dockerImages, 3)
}

func TestImageListView_ImageOperations(t *testing.T) {
	// Test that image operations don't crash
	// Note: These operations will fail in tests since we don't have actual images,
	// but we're checking that the methods exist and don't panic

	t.Run("DeleteImage", func(t *testing.T) {
		dockerClient := docker.NewClient()
		view := NewImageListView(dockerClient)
		view.dockerImages = createTestDockerImages()

		assert.NotPanics(t, func() {
			view.deleteImage(view.dockerImages[0])
		})
	})

	t.Run("ForceDeleteImage", func(t *testing.T) {
		dockerClient := docker.NewClient()
		view := NewImageListView(dockerClient)
		view.dockerImages = createTestDockerImages()

		assert.NotPanics(t, func() {
			view.forceDeleteImage(view.dockerImages[0])
		})
	})

	t.Run("PullImage", func(t *testing.T) {
		dockerClient := docker.NewClient()
		view := NewImageListView(dockerClient)
		view.dockerImages = createTestDockerImages()

		assert.NotPanics(t, func() {
			view.pullImage(view.dockerImages[0])
		})
	})

	t.Run("ShowHistory", func(t *testing.T) {
		dockerClient := docker.NewClient()
		view := NewImageListView(dockerClient)
		view.dockerImages = createTestDockerImages()

		assert.NotPanics(t, func() {
			view.showHistory(view.dockerImages[0])
		})
	})

	t.Run("PruneImages", func(t *testing.T) {
		dockerClient := docker.NewClient()
		view := NewImageListView(dockerClient)

		assert.NotPanics(t, func() {
			view.PruneImages()
		})
	})
}

func TestImageListView_ExportImage(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewImageListView(dockerClient)
	testImages := createTestDockerImages()

	// Test export - it may succeed or fail depending on whether the image exists
	// We're mainly testing that the function doesn't panic
	_ = view.ExportImage(testImages[0], "/tmp/test.tar")
	// No assertion on error since it depends on Docker state
}

func TestImageListView_KeyHandling(t *testing.T) {
	// This test verifies that key handlers are properly set up
	dockerClient := docker.NewClient()
	view := NewImageListView(dockerClient)

	// The table should have input capture function set
	assert.NotNil(t, view.table)

	// We can't easily test the actual key handling without running the app,
	// but we can verify the structure is in place
	primitive := view.GetPrimitive()
	pages, ok := primitive.(*tview.Pages)
	assert.True(t, ok)
	assert.NotNil(t, pages)
}

func TestImageListView_GetRepoTag(t *testing.T) {
	testCases := []struct {
		name     string
		image    models.DockerImage
		expected string
	}{
		{
			name: "Normal image",
			image: models.DockerImage{
				Repository: "nginx",
				Tag:        "latest",
				ID:         "abc123",
			},
			expected: "nginx:latest",
		},
		{
			name: "Image with <none> repository",
			image: models.DockerImage{
				Repository: "<none>",
				Tag:        "<none>",
				ID:         "def456",
			},
			expected: "def456",
		},
		{
			name: "Image with <none> tag",
			image: models.DockerImage{
				Repository: "myimage",
				Tag:        "<none>",
				ID:         "ghi789",
			},
			expected: "myimage",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.image.GetRepoTag()
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestImageListView_PullImageWithNoneRepository(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewImageListView(dockerClient)

	// Test pulling image with <none> repository
	noneImage := models.DockerImage{
		Repository: "<none>",
		Tag:        "<none>",
		ID:         "abc123",
	}

	// Should not panic and should log warning
	assert.NotPanics(t, func() {
		view.pullImage(noneImage)
	})
}

func TestImageListView_ImageIDTruncation(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewImageListView(dockerClient)

	// Create images with various ID lengths
	view.dockerImages = []models.DockerImage{
		{
			ID:           "sha256:abcdefghijklmnopqrstuvwxyz",
			Repository:   "test1",
			Tag:          "latest",
			CreatedSince: "1 day ago",
			Size:         "100MB",
		},
		{
			ID:           "short",
			Repository:   "test2",
			Tag:          "latest",
			CreatedSince: "2 days ago",
			Size:         "50MB",
		},
	}

	// Create a test app
	app := tview.NewApplication()
	SetApp(app)

	// Update the table
	view.updateTable()

	// Check that long ID is truncated
	idCell1 := view.table.GetCell(1, 2)
	assert.NotNil(t, idCell1)
	assert.Equal(t, "sha256:abcde", idCell1.Text) // Truncated to 12 chars

	// Check that short ID is not truncated
	idCell2 := view.table.GetCell(2, 2)
	assert.NotNil(t, idCell2)
	assert.Equal(t, "short", idCell2.Text) // Not truncated
}
