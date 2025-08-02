package ui

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tokuhirom/dcv/internal/models"
)

func TestView(t *testing.T) {
	tests := []struct {
		name     string
		model    Model
		contains []string
	}{
		{
			name: "loading state",
			model: Model{
				width:  0,
				height: 0,
			},
			contains: []string{"Loading..."},
		},
		{
			name: "process list with containers",
			model: Model{
				currentView: ProcessListView,
				width:       80,
				height:      24,
				loading:     false,
				containers: []models.Container{
					{
						Name:    "web-1",
						Image:   "nginx:latest",
						Service: "web",
						Status:  "Up 5 minutes",
					},
					{
						Name:    "dind-1",
						Image:   "docker:dind",
						Service: "dind",
						Status:  "Up 10 minutes",
					},
				},
			},
			contains: []string{
				"Docker Compose Processes",
				"NAME",
				"IMAGE",
				"SERVICE",
				"STATUS",
				"web-1",
				"nginx:latest",
				"dind-1",
				"docker:dind",
				"up:move up",
				"enter:view logs",
			},
		},
		{
			name: "process list with error",
			model: Model{
				currentView: ProcessListView,
				width:       80,
				height:      24,
				loading:     false,
				err:         assert.AnError,
			},
			contains: []string{
				"Error:",
				"Press 'q' to quit",
			},
		},
		{
			name: "process list with docker-compose.yml error",
			model: Model{
				currentView: ProcessListView,
				width:       80,
				height:      24,
				loading:     false,
				err:         &mockError{msg: "no configuration file provided"},
			},
			contains: []string{
				"No docker-compose.yml found",
				"Please run from a directory",
			},
		},
		{
			name: "log view",
			model: Model{
				currentView:   LogView,
				width:         80,
				height:        24,
				containerName: "web-1",
				logs: []string{
					"Starting web server...",
					"Listening on port 80",
					"Request received",
				},
			},
			contains: []string{
				"Logs: web-1",
				"Starting web server",
				"Listening on port 80",
				"up:scroll up",
				"/:search",
			},
		},
		{
			name: "log view in search mode",
			model: Model{
				currentView:   LogView,
				width:         80,
				height:        24,
				containerName: "web-1",
				searchMode:    true,
				searchText:    "error",
			},
			contains: []string{
				"Search: error",
			},
		},
		{
			name: "dind process list",
			model: Model{
				currentView:     DindProcessListView,
				width:           80,
				height:          24,
				loading:         false,
				currentDindHost: "dind-1",
				dindContainers: []models.Container{
					{
						ID:     "abc123def456",
						Image:  "alpine:latest",
						Name:   "test-container",
						Status: "Up 2 minutes",
					},
				},
			},
			contains: []string{
				"Docker in Docker: dind-1",
				"CONTAINER ID",
				"IMAGE",
				"STATUS",
				"NAME",
				"abc123def456"[:12],
				"alpine:latest",
				"test-container",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize key handlers for the test model
			tt.model.initializeKeyHandlers()
			
			view := tt.model.View()
			for _, expected := range tt.contains {
				assert.Contains(t, view, expected)
			}
		})
	}
}

func TestRenderProcessList(t *testing.T) {
	m := Model{
		currentView: ProcessListView,
		width:       80,
		height:      24,
		loading:     false,
		containers: []models.Container{
			{
				Name:    "web-1",
				Image:   "nginx:latest",
				Service: "web",
				Status:  "Up 5 minutes",
			},
		},
		selectedContainer: 0,
	}

	view := m.renderProcessList()

	// Check that the selected row is highlighted
	assert.Contains(t, view, "web-1")
	assert.Contains(t, view, "nginx:latest")

	// Check table structure
	assert.Contains(t, view, "│")
	assert.Contains(t, view, "─")
}

func TestRenderLogView(t *testing.T) {
	m := Model{
		currentView:   LogView,
		width:         80,
		height:        10,
		containerName: "web-1",
		logs: []string{
			"Line 1",
			"Line 2",
			"Line 3",
			"Line 4",
			"Line 5",
		},
		logScrollY: 0,
	}

	view := m.renderLogView()

	// Should show logs
	assert.Contains(t, view, "Line 1")

	// Test scrolling
	m.logScrollY = 2
	view = m.renderLogView()
	assert.NotContains(t, view, "Line 1")
	assert.Contains(t, view, "Line 3")
}

func TestRenderDindList(t *testing.T) {
	m := Model{
		currentView:     DindProcessListView,
		width:           80,
		height:          24,
		loading:         false,
		currentDindHost: "dind-1",
		dindContainers: []models.Container{
			{
				ID:     "abc123def456789",
				Image:  "alpine:latest",
				Name:   "test-1",
				Status: "Up 5 minutes",
			},
			{
				ID:     "def456ghi789012",
				Image:  "nginx:latest",
				Name:   "test-2",
				Status: "Up 3 minutes",
			},
		},
		selectedDindContainer: 1,
	}

	view := m.renderDindList()

	// Check title
	assert.Contains(t, view, "Docker in Docker: dind-1")

	// Check containers are listed
	assert.Contains(t, view, "abc123def456") // First 12 chars
	assert.Contains(t, view, "def456ghi789") // First 12 chars
	assert.Contains(t, view, "alpine:latest")
	assert.Contains(t, view, "nginx:latest")
}

func TestViewWithNoContainers(t *testing.T) {
	m := Model{
		currentView: ProcessListView,
		width:       80,
		height:      24,
		loading:     false,
		containers:  []models.Container{},
	}

	view := m.renderProcessList()
	assert.Contains(t, view, "No containers found")
	assert.Contains(t, view, "Press 'r' to refresh")
}

func TestTableRendering(t *testing.T) {
	m := Model{
		currentView: ProcessListView,
		width:       80,
		height:      24,
		loading:     false,
		containers: []models.Container{
			{
				Name:    "web-1",
				Image:   "nginx:latest",
				Service: "web",
				Status:  "Up 5 minutes",
			},
		},
	}

	view := m.renderProcessList()

	// Check for table borders
	lines := strings.Split(view, "\n")
	hasTopBorder := false
	hasBottomBorder := false
	hasVerticalBorder := false

	for _, line := range lines {
		if strings.Contains(line, "┌") || strings.Contains(line, "┐") {
			hasTopBorder = true
		}
		if strings.Contains(line, "└") || strings.Contains(line, "┘") {
			hasBottomBorder = true
		}
		if strings.Contains(line, "│") {
			hasVerticalBorder = true
		}
	}

	assert.True(t, hasTopBorder, "Table should have top border")
	assert.True(t, hasBottomBorder, "Table should have bottom border")
	assert.True(t, hasVerticalBorder, "Table should have vertical borders")
}

// mockError implements error interface for testing
type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}
