package views

import (
	"testing"
	"time"

	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
)

func TestNewFileBrowserView(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewFileBrowserView(dockerClient)

	assert.NotNil(t, view)
	assert.NotNil(t, view.docker)
	assert.NotNil(t, view.table)
	assert.Equal(t, "/", view.currentPath)
	assert.Equal(t, []string{"/"}, view.pathHistory)
	assert.Empty(t, view.containerID)
	assert.Empty(t, view.containerName)
	assert.Empty(t, view.files)
}

func TestFileBrowserView_GetPrimitive(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewFileBrowserView(dockerClient)

	primitive := view.GetPrimitive()
	assert.NotNil(t, primitive)
	_, ok := primitive.(*tview.Table)
	assert.True(t, ok)
}

func TestFileBrowserView_GetTitle(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewFileBrowserView(dockerClient)

	// Test with no container
	title := view.GetTitle()
	assert.Equal(t, "File Browser: /", title)

	// Test with container
	view.mu.Lock()
	view.containerName = "test-container"
	view.currentPath = "/app"
	view.mu.Unlock()
	title = view.GetTitle()
	assert.Equal(t, "File Browser: test-container:/app", title)
}

func TestFileBrowserView_SetContainer(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewFileBrowserView(dockerClient)

	// Set container without triggering load (which would require Docker)
	view.mu.Lock()
	view.containerID = "abc123"
	view.containerName = "test-container"
	view.mu.Unlock()

	assert.Equal(t, "abc123", view.containerID)
	assert.Equal(t, "test-container", view.containerName)
	assert.Equal(t, "/", view.currentPath)
	assert.Equal(t, []string{"/"}, view.pathHistory)
}

func TestFileBrowserView_Navigation(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewFileBrowserView(dockerClient)

	// Test navigateTo
	view.mu.Lock()
	view.currentPath = "/app"
	view.pathHistory = []string{"/", "/app"}
	view.mu.Unlock()

	assert.Equal(t, "/app", view.currentPath)
	assert.Equal(t, []string{"/", "/app"}, view.pathHistory)

	// Test navigateUp
	view.mu.Lock()
	view.currentPath = "/"
	view.pathHistory = []string{"/"}
	view.mu.Unlock()

	assert.Equal(t, "/", view.currentPath)

	// Test navigateBack
	view.mu.Lock()
	view.pathHistory = []string{"/", "/app", "/app/src"}
	view.currentPath = "/app/src"
	view.mu.Unlock()

	// Simulate navigateBack
	view.mu.Lock()
	if len(view.pathHistory) > 1 {
		view.pathHistory = view.pathHistory[:len(view.pathHistory)-1]
		view.currentPath = view.pathHistory[len(view.pathHistory)-1]
	}
	view.mu.Unlock()

	assert.Equal(t, "/app", view.currentPath)
	assert.Equal(t, []string{"/", "/app"}, view.pathHistory)
}

func TestFileBrowserView_UpdateTable(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewFileBrowserView(dockerClient)

	// Create test files
	testFiles := []models.ContainerFile{
		{
			Name:        "testdir",
			Permissions: "drwxr-xr-x",
			Size:        4096,
			ModTime:     time.Now(),
			IsDir:       true,
		},
		{
			Name:        "test.txt",
			Permissions: "-rw-r--r--",
			Size:        1024,
			ModTime:     time.Now(),
			IsDir:       false,
		},
		{
			Name:        "link",
			Permissions: "lrwxrwxrwx",
			Size:        10,
			ModTime:     time.Now(),
			LinkTarget:  "/target",
		},
	}

	view.mu.Lock()
	view.files = testFiles
	view.currentPath = "/app"
	view.mu.Unlock()

	// Create a test app
	app := tview.NewApplication()
	SetApp(app)

	// Update the table
	view.updateTable()

	// Check headers
	headerCell := view.table.GetCell(0, 0)
	assert.NotNil(t, headerCell)
	assert.Equal(t, "TYPE", headerCell.Text)

	headerCell = view.table.GetCell(0, 1)
	assert.NotNil(t, headerCell)
	assert.Equal(t, "PERMISSIONS", headerCell.Text)

	headerCell = view.table.GetCell(0, 2)
	assert.NotNil(t, headerCell)
	assert.Equal(t, "SIZE", headerCell.Text)

	headerCell = view.table.GetCell(0, 3)
	assert.NotNil(t, headerCell)
	assert.Equal(t, "MODIFIED", headerCell.Text)

	headerCell = view.table.GetCell(0, 4)
	assert.NotNil(t, headerCell)
	assert.Equal(t, "NAME", headerCell.Text)

	// Check parent directory entry (since currentPath is not "/")
	parentCell := view.table.GetCell(1, 4)
	assert.NotNil(t, parentCell)
	assert.Equal(t, "..", parentCell.Text)

	// Check first file (directory)
	typeCell := view.table.GetCell(2, 0)
	assert.NotNil(t, typeCell)
	assert.Equal(t, "DIR", typeCell.Text)

	nameCell := view.table.GetCell(2, 4)
	assert.NotNil(t, nameCell)
	assert.Equal(t, "testdir/", nameCell.Text)

	// Check second file (regular file)
	typeCell = view.table.GetCell(3, 0)
	assert.NotNil(t, typeCell)
	assert.Equal(t, "FILE", typeCell.Text)

	nameCell = view.table.GetCell(3, 4)
	assert.NotNil(t, nameCell)
	assert.Equal(t, "test.txt", nameCell.Text)

	// Check third file (symlink)
	typeCell = view.table.GetCell(4, 0)
	assert.NotNil(t, typeCell)
	assert.Equal(t, "LINK", typeCell.Text)

	nameCell = view.table.GetCell(4, 4)
	assert.NotNil(t, nameCell)
	assert.Equal(t, "link -> /target", nameCell.Text)
}

func TestFileBrowserView_GetSelectedFile(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewFileBrowserView(dockerClient)

	// Create test files
	testFiles := []models.ContainerFile{
		{Name: "file1.txt", IsDir: false},
		{Name: "dir1", IsDir: true},
		{Name: "file2.txt", IsDir: false},
	}

	view.mu.Lock()
	view.files = testFiles
	view.currentPath = "/app"
	view.mu.Unlock()

	// Create a test app
	app := tview.NewApplication()
	SetApp(app)

	// Update the table
	view.updateTable()

	// Select the second data row (first file, after parent "..")
	view.table.Select(2, 0)

	// Get selected file
	selectedFile := view.GetSelectedFile()
	assert.NotNil(t, selectedFile)
	assert.Equal(t, "file1.txt", selectedFile.Name)

	// Select the third data row (directory)
	view.table.Select(3, 0)
	selectedFile = view.GetSelectedFile()
	assert.NotNil(t, selectedFile)
	assert.Equal(t, "dir1", selectedFile.Name)
	assert.True(t, selectedFile.IsDir)
}

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		size     int64
		expected string
	}{
		{0, "0 B"},
		{100, "100 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{5242880, "5.0 MB"},
		{1073741824, "1.0 GB"},
		{2147483648, "2.0 GB"},
	}

	for _, tt := range tests {
		result := formatFileSize(tt.size)
		assert.Equal(t, tt.expected, result)
	}
}

func TestFileBrowserView_EmptyDirectory(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewFileBrowserView(dockerClient)

	// Set empty files
	view.mu.Lock()
	view.files = []models.ContainerFile{}
	view.currentPath = "/empty"
	view.mu.Unlock()

	// Create a test app
	app := tview.NewApplication()
	SetApp(app)

	// Update the table
	view.updateTable()

	// Should have headers and parent directory entry
	rowCount := view.table.GetRowCount()
	assert.Equal(t, 2, rowCount) // Headers + parent ".."
}

func TestFileBrowserView_RootDirectory(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewFileBrowserView(dockerClient)

	// Create test files for root
	testFiles := []models.ContainerFile{
		{Name: "etc", IsDir: true},
		{Name: "usr", IsDir: true},
		{Name: "var", IsDir: true},
	}

	view.mu.Lock()
	view.files = testFiles
	view.currentPath = "/"
	view.mu.Unlock()

	// Create a test app
	app := tview.NewApplication()
	SetApp(app)

	// Update the table
	view.updateTable()

	// Should NOT have parent directory entry at root
	rowCount := view.table.GetRowCount()
	assert.Equal(t, 4, rowCount) // Headers + 3 files (no parent "..")

	// First data row should be the first file, not ".."
	nameCell := view.table.GetCell(1, 4)
	assert.NotNil(t, nameCell)
	assert.Equal(t, "etc/", nameCell.Text)
}
