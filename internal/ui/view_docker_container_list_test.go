package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
)

// Helper function to create a test model
func createTestModel(currentView ViewType) *Model {
	return &Model{
		currentView:          currentView,
		dockerClient:         docker.NewClient(),
		width:                80,
		Height:               24,
		dockerListViewKeymap: make(map[string]KeyHandler),
	}
}

func TestDockerContainerListView_Rendering(t *testing.T) {
	t.Run("displays no containers message when empty", func(t *testing.T) {
		// Create model with empty container list
		m := createTestModel(DockerContainerListView)
		m.dockerContainerListViewModel.dockerContainers = []models.DockerContainer{}

		// Test the render function directly
		output := m.dockerContainerListViewModel.renderDockerList(m, 20)
		assert.Contains(t, output, "No containers found")
	})

	t.Run("displays container list table", func(t *testing.T) {
		// Create model with test containers
		m := createTestModel(DockerContainerListView)
		m.dockerContainerListViewModel.dockerContainers = []models.DockerContainer{
			{
				ID:     "abc123def456",
				Image:  "nginx:latest",
				Status: "Up 2 hours",
				Ports:  "80/tcp -> 0.0.0.0:8080",
				Names:  "web-server",
			},
			{
				ID:     "def456ghi789",
				Image:  "postgres:13",
				Status: "Up 3 hours",
				Ports:  "5432/tcp",
				Names:  "database",
			},
		}
		m.dockerContainerListViewModel.selectedDockerContainer = 0

		// Test the render function
		output := m.dockerContainerListViewModel.renderDockerList(m, 20)

		// Check for table headers
		assert.Contains(t, output, "CONTAINER ID")
		assert.Contains(t, output, "IMAGE")
		assert.Contains(t, output, "STATUS")
		assert.Contains(t, output, "PORTS")
		assert.Contains(t, output, "NAMES")

		// Check for container data
		assert.Contains(t, output, "abc123def456")
		assert.Contains(t, output, "nginx:latest")
		assert.Contains(t, output, "web-server")
		assert.Contains(t, output, "postgres:13")
		assert.Contains(t, output, "database")
	})

	t.Run("truncates long values", func(t *testing.T) {
		m := createTestModel(DockerContainerListView)
		m.dockerContainerListViewModel.dockerContainers = []models.DockerContainer{
			{
				ID:     "verylongcontaineridthatneedstruncation",
				Image:  "verylongimagenamethatneedstruncationbecauseitistoolong",
				Status: "Up",
				Ports:  "verylongportstringthatneedstruncationbecauseitistoolong",
				Names:  "container",
			},
		}

		output := m.dockerContainerListViewModel.renderDockerList(m, 20)

		// Check that ID is truncated to 12 chars
		assert.Contains(t, output, "verylongcont")
		assert.NotContains(t, output, "verylongcontaineridthatneedstruncation")

		// Check that image is truncated with ellipsis (Unicode character)
		assert.Contains(t, output, "…")
	})

	t.Run("highlights running containers", func(t *testing.T) {
		m := createTestModel(DockerContainerListView)
		m.dockerContainerListViewModel.dockerContainers = []models.DockerContainer{
			{
				ID:     "abc123def456",
				Image:  "nginx:latest",
				Status: "Up 2 hours",
				Names:  "running",
			},
			{
				ID:     "def456ghi789",
				Image:  "postgres:13",
				Status: "Exited (0) 1 hour ago",
				Names:  "stopped",
			},
		}

		// The render function applies different styles to Up vs Exited containers
		output := m.dockerContainerListViewModel.renderDockerList(m, 20)
		assert.Contains(t, output, "Up 2 hours")
		assert.Contains(t, output, "Exited")
	})
}

func TestDockerContainerListView_Navigation(t *testing.T) {
	t.Run("navigation with direct key handler calls", func(t *testing.T) {
		// Create model with multiple containers
		m := createTestModel(DockerContainerListView)
		m.dockerContainerListViewModel.dockerContainers = []models.DockerContainer{
			{ID: "container1", Image: "image1", Status: "Up", Names: "name1"},
			{ID: "container2", Image: "image2", Status: "Up", Names: "name2"},
			{ID: "container3", Image: "image3", Status: "Up", Names: "name3"},
		}
		m.dockerContainerListViewModel.selectedDockerContainer = 0
		m.initializeKeyHandlers()

		// Test moving down
		_, _ = m.CmdDown(tea.KeyMsg{Type: tea.KeyDown})
		assert.Equal(t, 1, m.dockerContainerListViewModel.selectedDockerContainer)

		// Test moving down again
		_, _ = m.CmdDown(tea.KeyMsg{Type: tea.KeyDown})
		assert.Equal(t, 2, m.dockerContainerListViewModel.selectedDockerContainer)

		// Test moving down at the end (should stay at 2)
		_, _ = m.CmdDown(tea.KeyMsg{Type: tea.KeyDown})
		assert.Equal(t, 2, m.dockerContainerListViewModel.selectedDockerContainer)

		// Test moving up
		_, _ = m.CmdUp(tea.KeyMsg{Type: tea.KeyUp})
		assert.Equal(t, 1, m.dockerContainerListViewModel.selectedDockerContainer)

		// Test moving up again
		_, _ = m.CmdUp(tea.KeyMsg{Type: tea.KeyUp})
		assert.Equal(t, 0, m.dockerContainerListViewModel.selectedDockerContainer)

		// Test moving up at the beginning (should stay at 0)
		_, _ = m.CmdUp(tea.KeyMsg{Type: tea.KeyUp})
		assert.Equal(t, 0, m.dockerContainerListViewModel.selectedDockerContainer)
	})
}

func TestDockerContainerListView_KeyHandlers(t *testing.T) {
	t.Run("key handler registration", func(t *testing.T) {
		m := createTestModel(DockerContainerListView)
		m.initializeKeyHandlers()

		// Check that docker container view handlers are registered
		assert.NotEmpty(t, m.dockerContainerListViewHandlers)
		assert.NotEmpty(t, m.dockerListViewKeymap)

		// Check specific handlers exist
		hasKillHandler := false
		hasToggleHandler := false
		for _, config := range m.dockerContainerListViewHandlers {
			if config.Description == "kill" {
				hasKillHandler = true
			}
			if config.Description == "toggle all" {
				hasToggleHandler = true
			}
		}
		assert.True(t, hasKillHandler, "Should have kill handler")
		assert.True(t, hasToggleHandler, "Should have toggle all handler")
	})

	t.Run("view switching handlers", func(t *testing.T) {
		m := createTestModel(DockerContainerListView)
		m.initializeKeyHandlers()

		// Check view-specific handlers
		hasBackSwitch := false
		for _, config := range m.dockerContainerListViewHandlers {
			if config.Description == "back" {
				hasBackSwitch = true
			}
		}
		// Check global handlers for view switching
		hasImageSwitch := false
		hasProjectSwitch := false
		for _, config := range m.globalHandlers {
			if strings.Contains(config.Description, "images") {
				hasImageSwitch = true
			}
			if strings.Contains(config.Description, "project") {
				hasProjectSwitch = true
			}
		}
		assert.True(t, hasBackSwitch, "Should have back handler")
		assert.True(t, hasImageSwitch, "Should have image switch handler")
		assert.True(t, hasProjectSwitch, "Should have project switch handler")
	})
}

func TestDockerContainerListView_Update(t *testing.T) {
	t.Run("handles container selection bounds", func(t *testing.T) {
		m := createTestModel(DockerContainerListView)
		m.dockerContainerListViewModel.dockerContainers = []models.DockerContainer{
			{ID: "container1", Names: "test1"},
			{ID: "container2", Names: "test2"},
		}
		m.dockerContainerListViewModel.selectedDockerContainer = 0
		m.initializeKeyHandlers()

		// Try to move up from first item
		_, cmd := m.CmdUp(tea.KeyMsg{})
		assert.Nil(t, cmd)
		assert.Equal(t, 0, m.dockerContainerListViewModel.selectedDockerContainer)

		// Move to last item
		m.dockerContainerListViewModel.selectedDockerContainer = 1

		// Try to move down from last item
		_, cmd = m.CmdDown(tea.KeyMsg{})
		assert.Nil(t, cmd)
		assert.Equal(t, 1, m.dockerContainerListViewModel.selectedDockerContainer)
	})

	t.Run("handles empty container list", func(t *testing.T) {
		m := createTestModel(DockerContainerListView)
		m.dockerContainerListViewModel.dockerContainers = []models.DockerContainer{}
		m.initializeKeyHandlers()

		// Try operations on empty list
		_, cmd := m.CmdKill(tea.KeyMsg{})
		assert.Nil(t, cmd) // Should not crash
	})
}

// Test the view content directly without teatest to avoid Docker daemon calls
func TestDockerContainerListView_FullOutput(t *testing.T) {
	t.Run("renders complete view", func(t *testing.T) {
		m := createTestModel(DockerContainerListView)
		m.dockerContainerListViewModel.dockerContainers = []models.DockerContainer{
			{
				ID:     "abc123def456",
				Image:  "nginx:latest",
				Status: "Up 2 hours",
				Ports:  "80/tcp",
				Names:  "web",
			},
		}
		m.width = 120
		m.Height = 30
		m.loading = false
		m.initializeKeyHandlers()

		// Test the View() method directly instead of using teatest
		output := m.View()

		// Check that the output contains expected content
		assert.Contains(t, output, "Docker Containers")
		assert.Contains(t, output, "CONTAINER ID")
		assert.Contains(t, output, "IMAGE")
		assert.Contains(t, output, "nginx:latest")
		assert.Contains(t, output, "abc123def456")
		assert.Contains(t, output, "web")
		assert.Contains(t, output, "Up 2 hours")
	})
}
