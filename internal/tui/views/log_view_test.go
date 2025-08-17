package views

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tokuhirom/dcv/internal/docker"
)

func TestNewLogView(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewLogView(dockerClient)

	assert.NotNil(t, view)
	assert.NotNil(t, view.docker)
	assert.NotNil(t, view.textView)
	assert.NotNil(t, view.pages)
	assert.Empty(t, view.containerID)
	assert.Empty(t, view.logs)
	assert.True(t, view.autoScroll)
	assert.True(t, view.wrap)
	assert.True(t, view.follow)
	assert.False(t, view.timestamps)
	assert.Equal(t, "100", view.tail)
	assert.False(t, view.streaming)
}

func TestLogView_GetPrimitive(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewLogView(dockerClient)

	primitive := view.GetPrimitive()
	assert.NotNil(t, primitive)
}

func TestLogView_GetTitle(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewLogView(dockerClient)

	// Test with no container
	title := view.GetTitle()
	assert.Equal(t, "Container Logs [Following | Auto-scroll | Wrap]", title)

	// Test with container ID
	view.containerID = "abc123def456"
	title = view.GetTitle()
	assert.Contains(t, title, "Logs: abc123def456")
	assert.Contains(t, title, "Following")

	// Test with paused streaming
	view.follow = false
	title = view.GetTitle()
	assert.Contains(t, title, "Paused")

	// Test with search
	view.searchText = "error"
	title = view.GetTitle()
	assert.Contains(t, title, "Search: error")

	// Test with filter
	view.isFiltered = true
	view.filterText = "warning"
	title = view.GetTitle()
	assert.Contains(t, title, "Filter: warning")
}

func TestLogView_SetContainer(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewLogView(dockerClient)

	// Set container
	view.SetContainer("abc123", nil)
	assert.Equal(t, "abc123", view.containerID)
}

func TestLogView_AddLog(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewLogView(dockerClient)

	// Add a log line
	view.addLog("Test log line 1")
	assert.Len(t, view.logs, 1)
	assert.Equal(t, "Test log line 1", view.logs[0])

	// Add another log line
	view.addLog("Test log line 2")
	assert.Len(t, view.logs, 2)
	assert.Equal(t, "Test log line 2", view.logs[1])
}

func TestLogView_ClearLogs(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewLogView(dockerClient)

	// Add some logs
	view.addLog("Test log 1")
	view.addLog("Test log 2")
	assert.Len(t, view.logs, 2)

	// Clear logs
	view.clearLogs()
	assert.Empty(t, view.logs)
	assert.Empty(t, view.filteredLogs)
}

func TestLogView_Search(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewLogView(dockerClient)

	// Add some logs
	view.logs = []string{
		"Info: Starting application",
		"Error: Failed to connect",
		"Warning: Low memory",
		"Error: Timeout occurred",
		"Info: Application running",
	}

	// Perform search
	view.performSearch("Error")
	assert.Equal(t, "Error", view.searchText)
	assert.NotNil(t, view.searchRegex)
	assert.Len(t, view.searchResults, 2) // Should find 2 error lines
	assert.Equal(t, 1, view.searchResults[0])
	assert.Equal(t, 3, view.searchResults[1])
}

func TestLogView_Filter(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewLogView(dockerClient)

	// Add some logs
	view.logs = []string{
		"Info: Starting application",
		"Error: Failed to connect",
		"Warning: Low memory",
		"Error: Timeout occurred",
		"Info: Application running",
	}

	// Apply filter
	view.applyFilter("Error")
	assert.Equal(t, "Error", view.filterText)
	assert.NotNil(t, view.filterRegex)
	assert.True(t, view.isFiltered)
	assert.Len(t, view.filteredLogs, 2) // Should filter to 2 error lines
	assert.Equal(t, "Error: Failed to connect", view.filteredLogs[0])
	assert.Equal(t, "Error: Timeout occurred", view.filteredLogs[1])

	// Clear filter
	view.applyFilter("")
	assert.False(t, view.isFiltered)
	assert.Nil(t, view.filterRegex)
	assert.Nil(t, view.filteredLogs)
}

func TestLogView_BufferLimit(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewLogView(dockerClient)

	// Add more than 10000 logs
	for i := 0; i < 10100; i++ {
		view.addLog("Log line")
	}

	// Should be limited to 10000
	assert.Len(t, view.logs, 10000)
}

func TestLogView_ToggleSettings(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewLogView(dockerClient)

	// Test pause/resume
	assert.True(t, view.follow)
	view.pauseStreaming()
	assert.False(t, view.follow)
	view.resumeStreaming()
	assert.True(t, view.follow)

	// Test tail setting
	view.SetTailLines("50")
	assert.Equal(t, "50", view.tail)
}

func TestLogView_SearchAndFilter(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewLogView(dockerClient)

	// Add logs
	view.logs = []string{
		"2024-01-01 Info: Start",
		"2024-01-01 Error: Connection failed",
		"2024-01-02 Error: Timeout",
		"2024-01-02 Info: Retry",
	}

	// Apply filter first
	view.applyFilter("Error")
	assert.Len(t, view.filteredLogs, 2)

	// Then search within filtered results
	view.performSearch("Timeout")
	assert.Len(t, view.searchResults, 1)
}
