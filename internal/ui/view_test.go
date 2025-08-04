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
			name: "process list with composeContainers",
			model: Model{
				currentView: ComposeProcessListView,
				width:       80,
				height:      24,
				loading:     false,
				composeProcessListViewModel: ComposeProcessListViewModel{
					composeContainers: []models.ComposeContainer{
						{
							Name:    "web-1",
							Command: "/docker-entrypoint.sh nginx -g 'daemon off;'",
							Service: "web",
							State:   "running",
						},
						{
							Name:    "dind-1",
							Command: "dockerd",
							Service: "dind",
							State:   "running",
						},
					},
				},
			},
			contains: []string{
				"Docker Compose",
				"SERVICE",
				"IMAGE",
				"STATUS",
				"PORTS",
				"web",
				"dind",
				"Press ? for help",
			},
		},
		{
			name: "process list with error",
			model: Model{
				currentView: ComposeProcessListView,
				width:       80,
				height:      24,
				loading:     false,
				err:         assert.AnError,
			},
			contains: []string{
				"Error:",
			},
		},
		{
			name: "process list with docker-compose.yml error",
			model: Model{
				currentView: ComposeProcessListView,
				width:       80,
				height:      24,
				loading:     false,
				err:         &mockError{msg: "no configuration file provided"},
			},
			contains: []string{
				"Error: no configuration file provided",
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
				"Press ? for help",
			},
		},
		{
			name: "log view in search mode",
			model: Model{
				currentView:     LogView,
				width:           80,
				height:          24,
				containerName:   "web-1",
				searchMode:      true,
				searchText:      "error",
				searchCursorPos: 5,
			},
			contains: []string{
				"/error", // Search prompt in footer
			},
		},
		{
			name: "dind process list",
			model: Model{
				currentView: DindProcessListView,
				width:       80,
				height:      24,
				loading:     false,
				dindProcessListViewModel: DindProcessListViewModel{
					currentDindHost: "dind-1",
					dindContainers: []models.DockerContainer{
						{
							ID:     "abc123def456",
							Image:  "alpine:latest",
							Names:  "test-container",
							Status: "Up 2 minutes",
						},
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
		currentView: ComposeProcessListView,
		width:       80,
		height:      24,
		loading:     false,
		composeProcessListViewModel: ComposeProcessListViewModel{
			composeContainers: []models.ComposeContainer{
				{
					Name:    "web-1",
					Command: "/docker-entrypoint.sh nginx -g 'daemon off;'",
					Service: "web",
					State:   "running",
				},
			},
			selectedContainer: 0,
		},
	}

	// Calculate available height (height - title - footer)
	availableHeight := m.height - 2
	view := m.composeProcessListViewModel.render(&m, availableHeight)

	// Check that the selected row is highlighted
	assert.Contains(t, view, "web")

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

	// Calculate available height
	availableHeight := m.height - 2
	view := m.renderLogView(availableHeight)

	// Should show logs
	assert.Contains(t, view, "Line 1")

	// Test scrolling
	m.logScrollY = 2
	view = m.renderLogView(availableHeight)
	assert.NotContains(t, view, "Line 1")
	assert.Contains(t, view, "Line 3")
}

func TestRenderDindList(t *testing.T) {
	m := Model{
		currentView: DindProcessListView,
		width:       80,
		height:      24,
		loading:     false,
		dindProcessListViewModel: DindProcessListViewModel{
			currentDindHost: "dind-1",
			dindContainers: []models.DockerContainer{
				{
					ID:     "abc123def456789",
					Image:  "alpine:latest",
					Names:  "test-1",
					Status: "Up 5 minutes",
				},
				{
					ID:     "def456ghi789012",
					Image:  "nginx:latest",
					Names:  "test-2",
					Status: "Up 3 minutes",
				},
			},
			selectedDindContainer: 1,
		},
	}

	// Calculate available height
	availableHeight := m.height - 2
	view := m.dindProcessListViewModel.render(&m, availableHeight)

	// The title is in viewTitle(), not renderDindList()
	// Check that dind containers are listed correctly

	// Check composeContainers are listed
	assert.Contains(t, view, "abc123def456") // First 12 chars
	assert.Contains(t, view, "def456ghi789") // First 12 chars
	assert.Contains(t, view, "alpine:latest")
	assert.Contains(t, view, "nginx:latest")
}

func TestViewWithNoContainers(t *testing.T) {
	m := Model{
		currentView: ComposeProcessListView,
		width:       80,
		height:      24,
		loading:     false,
		composeProcessListViewModel: ComposeProcessListViewModel{
			composeContainers: []models.ComposeContainer{},
		},
	}

	// Calculate available height
	availableHeight := m.height - 2
	view := m.composeProcessListViewModel.render(&m, availableHeight)
	assert.Contains(t, view, "No containers found")
	assert.Contains(t, view, "Press u to start services or p to switch to project list")
}

func TestTableRendering(t *testing.T) {
	m := Model{
		currentView: ComposeProcessListView,
		width:       80,
		height:      24,
		loading:     false,
		composeProcessListViewModel: ComposeProcessListViewModel{
			composeContainers: []models.ComposeContainer{
				{
					Name:    "web-1",
					Command: "/docker-entrypoint.sh nginx -g 'daemon off;'",
					Service: "web",
					State:   "running",
				},
			},
		},
	}

	// Calculate available height
	availableHeight := m.height - 2
	view := m.composeProcessListViewModel.render(&m, availableHeight)

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

func TestDockerContainerListView(t *testing.T) {
	t.Run("docker_list_with_containers", func(t *testing.T) {
		m := &Model{
			width:       80,
			height:      24,
			currentView: DockerContainerListView,
			dockerContainerListViewModel: DockerContainerListViewModel{
				selectedDockerContainer: 0,
				dockerContainers: []models.DockerContainer{
					{ID: "abc123def456", Names: "nginx", Image: "nginx:latest", Status: "Up 2 hours", Ports: "80/tcp"},
					{ID: "789012345678", Names: "redis", Image: "redis:alpine", Status: "Exited (0) 1 hour ago", Ports: ""},
				},
			},
		}
		m.initializeKeyHandlers()

		view := m.View()

		// Check title
		assert.Contains(t, view, "Docker Containers")

		// Check for table headers
		assert.Contains(t, view, "CONTAINER ID")
		assert.Contains(t, view, "IMAGE")
		assert.Contains(t, view, "STATUS")
		assert.Contains(t, view, "PORTS")
		assert.Contains(t, view, "NAMES")

		// Check for data
		assert.Contains(t, view, "abc123def456")
		assert.Contains(t, view, "nginx")
		assert.Contains(t, view, "nginx:latest")
		assert.Contains(t, view, "Up 2 hours")
		assert.Contains(t, view, "80/tcp")
	})

	t.Run("docker_list_show_all", func(t *testing.T) {
		m := &Model{
			width:       80,
			height:      24,
			currentView: DockerContainerListView,
			dockerContainerListViewModel: DockerContainerListViewModel{
				dockerContainers: []models.DockerContainer{},
				showAll:          true,
			},
		}
		m.initializeKeyHandlers()

		view := m.View()

		// Check title includes (all)
		assert.Contains(t, view, "Docker Containers (all)")
	})

	t.Run("docker_list_empty", func(t *testing.T) {
		m := &Model{
			width:       80,
			height:      24,
			currentView: DockerContainerListView,
			dockerContainerListViewModel: DockerContainerListViewModel{
				dockerContainers: []models.DockerContainer{},
			},
		}
		m.initializeKeyHandlers()

		view := m.View()

		assert.Contains(t, view, "No containers found")
	})
}

func TestFileBrowserTableView(t *testing.T) {
	t.Run("file_browser_with_files", func(t *testing.T) {
		m := &Model{
			width:                 80,
			height:                24,
			currentView:           FileBrowserView,
			browsingContainerName: "web-1",
			currentPath:           "/app",
			selectedFile:          1,
			containerFiles: []models.ContainerFile{
				{Name: "Dockerfile", Permissions: "-rw-r--r--", IsDir: false},
				{Name: "src", Permissions: "drwxr-xr-x", IsDir: true},
				{Name: "README.md", Permissions: "-rw-r--r--", IsDir: false},
				{Name: "link", Permissions: "lrwxrwxrwx", IsDir: false, LinkTarget: "/etc/hosts"},
			},
		}
		m.initializeKeyHandlers()

		view := m.View()

		// Check title
		assert.Contains(t, view, "File Browser: web-1 [/app]")

		// Check for table headers
		assert.Contains(t, view, "PERMISSIONS")
		assert.Contains(t, view, "NAME")

		// Check for file data
		assert.Contains(t, view, "-rw-r--r--")
		assert.Contains(t, view, "drwxr-xr-x")
		assert.Contains(t, view, "Dockerfile")
		assert.Contains(t, view, "src/")
		assert.Contains(t, view, "README.md")
		assert.Contains(t, view, "link -> /etc/hosts")

		// Check for path at bottom
		assert.Contains(t, view, "Path: /app")

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
	})

	t.Run("file_browser_empty_directory", func(t *testing.T) {
		m := &Model{
			width:                 80,
			height:                24,
			currentView:           FileBrowserView,
			browsingContainerName: "web-1",
			currentPath:           "/empty",
			containerFiles:        []models.ContainerFile{},
		}
		m.initializeKeyHandlers()

		view := m.View()

		assert.Contains(t, view, "No files found or directory is empty")
	})
}

// mockError implements error interface for testing
type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}
