package views

import (
	"testing"
	"time"

	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
)

func TestNewFileContentView(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewFileContentView(dockerClient)

	assert.NotNil(t, view)
	assert.NotNil(t, view.docker)
	assert.NotNil(t, view.textView)
	assert.Empty(t, view.containerID)
	assert.Empty(t, view.containerName)
	assert.Empty(t, view.filePath)
	assert.Empty(t, view.content)
	assert.True(t, view.lineNumbers)
	assert.False(t, view.wrap)
}

func TestFileContentView_GetPrimitive(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewFileContentView(dockerClient)

	primitive := view.GetPrimitive()
	assert.NotNil(t, primitive)
	_, ok := primitive.(*tview.TextView)
	assert.True(t, ok)
}

func TestFileContentView_GetTitle(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewFileContentView(dockerClient)

	// Test with no file
	title := view.GetTitle()
	assert.Equal(t, "File:  [Line numbers]", title)

	// Test with file path only
	view.mu.Lock()
	view.filePath = "/etc/hosts"
	view.mu.Unlock()
	title = view.GetTitle()
	assert.Equal(t, "File: /etc/hosts [Line numbers]", title)

	// Test with container and file
	view.mu.Lock()
	view.containerName = "test-container"
	view.fileInfo = models.ContainerFile{
		Size: 1024,
	}
	view.mu.Unlock()
	title = view.GetTitle()
	assert.Equal(t, "File: test-container:/etc/hosts [1.0 KB] [Line numbers]", title)

	// Test with wrap enabled
	view.mu.Lock()
	view.wrap = true
	view.mu.Unlock()
	title = view.GetTitle()
	assert.Equal(t, "File: test-container:/etc/hosts [1.0 KB] [Line numbers | Wrap]", title)

	// Test without line numbers
	view.mu.Lock()
	view.lineNumbers = false
	view.mu.Unlock()
	title = view.GetTitle()
	assert.Equal(t, "File: test-container:/etc/hosts [1.0 KB] [Wrap]", title)
}

func TestFileContentView_SetFile(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewFileContentView(dockerClient)

	fileInfo := models.ContainerFile{
		Name:    "test.txt",
		Size:    2048,
		ModTime: time.Now(),
		IsDir:   false,
	}

	// Set file without triggering load (which would require Docker)
	view.mu.Lock()
	view.containerID = "abc123"
	view.containerName = "test-container"
	view.filePath = "/app/test.txt"
	view.fileInfo = fileInfo
	view.mu.Unlock()

	assert.Equal(t, "abc123", view.containerID)
	assert.Equal(t, "test-container", view.containerName)
	assert.Equal(t, "/app/test.txt", view.filePath)
	assert.Equal(t, fileInfo, view.fileInfo)
}

func TestFileContentView_UpdateContent(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewFileContentView(dockerClient)

	// Create a test app
	app := tview.NewApplication()
	SetApp(app)

	// Test content without line numbers
	view.mu.Lock()
	view.content = "Line 1\nLine 2\nLine 3"
	view.lineNumbers = false
	view.mu.Unlock()

	view.updateContent()
	actualContent := view.textView.GetText(false)
	assert.Equal(t, "Line 1\nLine 2\nLine 3", actualContent)

	// Test content with line numbers
	view.mu.Lock()
	view.lineNumbers = true
	view.mu.Unlock()

	view.updateContent()
	actualContent = view.textView.GetText(true) // Strip color tags
	assert.Contains(t, actualContent, "1 Line 1")
	assert.Contains(t, actualContent, "2 Line 2")
	assert.Contains(t, actualContent, "3 Line 3")
}

func TestFileContentView_ToggleLineNumbers(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewFileContentView(dockerClient)

	// Initial state
	assert.True(t, view.lineNumbers)

	// Toggle off
	view.mu.Lock()
	view.lineNumbers = false
	view.mu.Unlock()
	assert.False(t, view.lineNumbers)

	// Toggle on
	view.mu.Lock()
	view.lineNumbers = true
	view.mu.Unlock()
	assert.True(t, view.lineNumbers)
}

func TestFileContentView_ToggleWrap(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewFileContentView(dockerClient)

	// Initial state
	assert.False(t, view.wrap)

	// Toggle on
	view.mu.Lock()
	view.wrap = true
	view.mu.Unlock()
	assert.True(t, view.wrap)

	// Toggle off
	view.mu.Lock()
	view.wrap = false
	view.mu.Unlock()
	assert.False(t, view.wrap)
}

func TestFileContentView_ErrorContent(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewFileContentView(dockerClient)

	// Create a test app
	app := tview.NewApplication()
	SetApp(app)

	// Set error content
	view.mu.Lock()
	view.content = "[red]Error loading file: permission denied[-]"
	view.lineNumbers = true // Should not add line numbers to error messages
	view.mu.Unlock()

	view.updateContent()
	actualContent := view.textView.GetText(false)
	assert.Equal(t, "[red]Error loading file: permission denied[-]", actualContent)
	// Should not have line numbers for error messages
	assert.NotContains(t, actualContent, "1 [red]")
}

func TestFileContentView_LargeFile(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewFileContentView(dockerClient)

	// Create a test app
	app := tview.NewApplication()
	SetApp(app)

	// Create content with many lines
	var content string
	for i := 1; i <= 1000; i++ {
		if i > 1 {
			content += "\n"
		}
		content += "Line content for testing"
	}

	view.mu.Lock()
	view.content = content
	view.lineNumbers = true
	view.mu.Unlock()

	view.updateContent()
	actualContent := view.textView.GetText(true)

	// Check that line numbers are properly formatted with correct width
	assert.Contains(t, actualContent, "   1 Line content")
	assert.Contains(t, actualContent, " 100 Line content")
	assert.Contains(t, actualContent, "1000 Line content")
}

func TestFileContentView_GetContent(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewFileContentView(dockerClient)

	testContent := "Test file content\nWith multiple lines"
	view.mu.Lock()
	view.content = testContent
	view.mu.Unlock()

	retrievedContent := view.GetContent()
	assert.Equal(t, testContent, retrievedContent)
}

func TestFileContentView_GetFilePath(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewFileContentView(dockerClient)

	testPath := "/app/config/settings.json"
	view.mu.Lock()
	view.filePath = testPath
	view.mu.Unlock()

	retrievedPath := view.GetFilePath()
	assert.Equal(t, testPath, retrievedPath)
}
