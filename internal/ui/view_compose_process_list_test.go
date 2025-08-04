package ui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/stretchr/testify/assert"

	"github.com/tokuhirom/dcv/internal/models"
)

func TestComposeProcessListView_Rendering(t *testing.T) {
	t.Run("displays no containers message when empty", func(t *testing.T) {
		// Create model with empty container list
		m := createTestModel(ComposeProcessListView)
		m.composeProcessListViewModel.composeContainers = []models.ComposeContainer{}

		// Test the render function directly
		output := m.composeProcessListViewModel.render(m, 20)
		assert.Contains(t, output, "No containers found")
		assert.Contains(t, output, "Press u to start services")
		assert.Contains(t, output, "or p to switch to project list")
	})

	t.Run("displays container list table", func(t *testing.T) {
		// Create model with test containers
		m := createTestModel(ComposeProcessListView)
		m.composeProcessListViewModel.composeContainers = []models.ComposeContainer{
			{
				Service: "web",
				Image:   "nginx:latest",
				State:   "running",
				Publishers: []struct {
					URL           string `json:"URL"`
					TargetPort    int    `json:"TargetPort"`
					PublishedPort int    `json:"PublishedPort"`
					Protocol      string `json:"Protocol"`
				}{
					{
						PublishedPort: 8080,
						TargetPort:    80,
						Protocol:      "tcp",
					},
				},
			},
			{
				Service: "db",
				Image:   "postgres:13",
				State:   "running",
				Publishers: []struct {
					URL           string `json:"URL"`
					TargetPort    int    `json:"TargetPort"`
					PublishedPort int    `json:"PublishedPort"`
					Protocol      string `json:"Protocol"`
				}{
					{
						TargetPort: 5432,
						Protocol:   "tcp",
					},
				},
			},
		}
		m.composeProcessListViewModel.selectedContainer = 0
		m.width = 120
		m.Height = 30

		// Test the render function
		output := m.composeProcessListViewModel.render(m, 20)

		// Check for table headers
		assert.Contains(t, output, "SERVICE")
		assert.Contains(t, output, "IMAGE")
		assert.Contains(t, output, "STATUS")
		assert.Contains(t, output, "PORTS")

		// Check for container data
		assert.Contains(t, output, "web")
		assert.Contains(t, output, "nginx:latest")
		assert.Contains(t, output, "db")
		assert.Contains(t, output, "postgres:13")
	})

	t.Run("displays dind indicator", func(t *testing.T) {
		m := createTestModel(ComposeProcessListView)
		m.composeProcessListViewModel.composeContainers = []models.ComposeContainer{
			{
				Service: "dind",
				Name:    "project-dind-1",
				Image:   "docker:dind",
				State:   "running",
			},
		}
		m.width = 120

		output := m.composeProcessListViewModel.render(m, 20)

		// Check for dind indicator (⬢ symbol)
		assert.Contains(t, output, "⬢")
		assert.Contains(t, output, "dind")
	})

	t.Run("truncates long values", func(t *testing.T) {
		m := createTestModel(ComposeProcessListView)
		m.composeProcessListViewModel.composeContainers = []models.ComposeContainer{
			{
				Service: "service",
				Image:   "verylongimagenamethatneedstruncationbecauseitistoolong",
				State:   "running",
				Publishers: []struct {
					URL           string `json:"URL"`
					TargetPort    int    `json:"TargetPort"`
					PublishedPort int    `json:"PublishedPort"`
					Protocol      string `json:"Protocol"`
				}{
					{
						PublishedPort: 8080,
						TargetPort:    80,
						Protocol:      "tcp",
					},
					{
						PublishedPort: 8081,
						TargetPort:    81,
						Protocol:      "tcp",
					},
					{
						PublishedPort: 8082,
						TargetPort:    82,
						Protocol:      "tcp",
					},
					{
						PublishedPort: 8083,
						TargetPort:    83,
						Protocol:      "tcp",
					},
				},
			},
		}
		m.width = 120

		output := m.composeProcessListViewModel.render(m, 20)

		// Check that image is truncated with ellipsis
		assert.Contains(t, output, "...")
		assert.NotContains(t, output, "verylongimagenamethatneedstruncationbecauseitistoolong")
	})

	t.Run("highlights running vs stopped containers", func(t *testing.T) {
		m := createTestModel(ComposeProcessListView)
		m.composeProcessListViewModel.composeContainers = []models.ComposeContainer{
			{
				Service: "running-service",
				Image:   "nginx:latest",
				State:   "running",
			},
			{
				Service:  "stopped-service",
				Image:    "postgres:13",
				State:    "exited",
				ExitCode: 0,
			},
		}
		m.width = 120

		// The render function applies different styles to Up vs Exited containers
		output := m.composeProcessListViewModel.render(m, 20)
		assert.Contains(t, output, "Up")         // GetStatus() returns "Up" for running
		assert.Contains(t, output, "Exited (0)") // GetStatus() returns "Exited (0)" for exited with code 0
	})
}

func TestComposeProcessListView_Navigation(t *testing.T) {
	t.Run("navigation with direct key handler calls", func(t *testing.T) {
		// Create model with multiple containers
		m := createTestModel(ComposeProcessListView)
		m.composeProcessListViewModel.composeContainers = []models.ComposeContainer{
			{Service: "service1", Image: "image1", State: "running"},
			{Service: "service2", Image: "image2", State: "running"},
			{Service: "service3", Image: "image3", State: "running"},
		}
		m.composeProcessListViewModel.selectedContainer = 0
		m.initializeKeyHandlers()

		// Test moving down
		m.composeProcessListViewModel.HandleDown()
		assert.Equal(t, 1, m.composeProcessListViewModel.selectedContainer)

		// Test moving down again
		m.composeProcessListViewModel.HandleDown()
		assert.Equal(t, 2, m.composeProcessListViewModel.selectedContainer)

		// Test moving down at the end (should stay at 2)
		m.composeProcessListViewModel.HandleDown()
		assert.Equal(t, 2, m.composeProcessListViewModel.selectedContainer)

		// Test moving up
		m.composeProcessListViewModel.HandleUp()
		assert.Equal(t, 1, m.composeProcessListViewModel.selectedContainer)

		// Test moving up again
		m.composeProcessListViewModel.HandleUp()
		assert.Equal(t, 0, m.composeProcessListViewModel.selectedContainer)

		// Test moving up at the beginning (should stay at 0)
		m.composeProcessListViewModel.HandleUp()
		assert.Equal(t, 0, m.composeProcessListViewModel.selectedContainer)
	})
}

func TestComposeProcessListView_KeyHandlers(t *testing.T) {
	t.Run("key handler registration", func(t *testing.T) {
		m := createTestModel(ComposeProcessListView)
		m.initializeKeyHandlers()

		// Check that compose process view handlers are registered
		assert.NotEmpty(t, m.processListViewHandlers)
		assert.NotEmpty(t, m.processListViewKeymap)

		// Check specific handlers exist
		hasKillHandler := false
		hasToggleHandler := false
		hasUpHandler := false
		hasDownHandler := false
		for _, config := range m.processListViewHandlers {
			if config.Description == "kill" {
				hasKillHandler = true
			}
			if config.Description == "toggle all" {
				hasToggleHandler = true
			}
			if config.Description == "up -d" {
				hasUpHandler = true
			}
			if config.Description == "down" {
				hasDownHandler = true
			}
		}
		assert.True(t, hasKillHandler, "Should have kill handler")
		assert.True(t, hasToggleHandler, "Should have toggle all handler")
		assert.True(t, hasUpHandler, "Should have up -d handler")
		assert.True(t, hasDownHandler, "Should have down handler")
	})

	t.Run("view switching handlers", func(t *testing.T) {
		m := createTestModel(ComposeProcessListView)
		m.initializeKeyHandlers()

		// Check view switching handlers exist
		hasDockerSwitch := false
		hasProjectSwitch := false
		hasImageSwitch := false
		hasDindSwitch := false
		for _, config := range m.processListViewHandlers {
			if strings.Contains(config.Description, "docker ps") {
				hasDockerSwitch = true
			}
			if strings.Contains(config.Description, "project") {
				hasProjectSwitch = true
			}
			if strings.Contains(config.Description, "images") {
				hasImageSwitch = true
			}
			if strings.Contains(config.Description, "dind") {
				hasDindSwitch = true
			}
		}
		assert.True(t, hasDockerSwitch, "Should have docker switch handler")
		assert.True(t, hasProjectSwitch, "Should have project switch handler")
		assert.True(t, hasImageSwitch, "Should have image switch handler")
		assert.True(t, hasDindSwitch, "Should have dind switch handler")
	})
}

func TestComposeProcessListView_Update(t *testing.T) {
	t.Run("handles container selection bounds", func(t *testing.T) {
		m := createTestModel(ComposeProcessListView)
		m.composeProcessListViewModel.composeContainers = []models.ComposeContainer{
			{Service: "service1"},
			{Service: "service2"},
		}
		m.composeProcessListViewModel.selectedContainer = 0
		m.initializeKeyHandlers()

		// Try to move up from first item
		cmd := m.composeProcessListViewModel.HandleUp()
		assert.Nil(t, cmd)
		assert.Equal(t, 0, m.composeProcessListViewModel.selectedContainer)

		// Move to last item
		m.composeProcessListViewModel.selectedContainer = 1

		// Try to move down from last item
		cmd = m.composeProcessListViewModel.HandleDown()
		assert.Nil(t, cmd)
		assert.Equal(t, 1, m.composeProcessListViewModel.selectedContainer)
	})

	t.Run("handles empty container list", func(t *testing.T) {
		m := createTestModel(ComposeProcessListView)
		m.composeProcessListViewModel.composeContainers = []models.ComposeContainer{}
		m.initializeKeyHandlers()

		// Try operations on empty list
		_, cmd := m.CmdKill(tea.KeyMsg{})
		assert.Nil(t, cmd) // Should not crash
	})
}

func TestComposeProcessListView_FullOutput(t *testing.T) {
	t.Run("renders complete view", func(t *testing.T) {
		m := createTestModel(ComposeProcessListView)
		m.composeProcessListViewModel.composeContainers = []models.ComposeContainer{
			{
				Service: "web",
				Image:   "nginx:latest",
				State:   "running",
			},
		}
		m.width = 120
		m.Height = 30
		m.projectName = "test-project"
		m.initializeKeyHandlers()

		tm := teatest.NewTestModel(
			t, m,
			teatest.WithInitialTermSize(120, 30),
		)

		// Wait for render
		teatest.WaitFor(
			t, tm.Output(),
			func(bts []byte) bool {
				output := string(bts)
				// The view should show the project name and the container
				return strings.Contains(output, "test-project") &&
					strings.Contains(output, "nginx:latest")
			},
			teatest.WithCheckInterval(time.Millisecond*50),
			teatest.WithDuration(time.Second),
		)

		// Send quit
		tm.Send(tea.QuitMsg{})
		tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
	})
}

func TestComposeProcessListView_ServiceOperations(t *testing.T) {
	t.Run("toggle show all containers", func(t *testing.T) {
		m := createTestModel(ComposeProcessListView)
		m.composeProcessListViewModel.showAll = false
		m.initializeKeyHandlers()

		// Toggle show all
		_, cmd := m.CmdToggleAll(tea.KeyMsg{})
		assert.NotNil(t, cmd) // Should return a command to refresh
		assert.True(t, m.composeProcessListViewModel.showAll)

		// Toggle back
		_, cmd = m.CmdToggleAll(tea.KeyMsg{})
		assert.NotNil(t, cmd)
		assert.False(t, m.composeProcessListViewModel.showAll)
	})

	t.Run("dind container detection", func(t *testing.T) {
		m := createTestModel(ComposeProcessListView)
		m.composeProcessListViewModel.composeContainers = []models.ComposeContainer{
			{
				Service: "regular",
				Name:    "project-regular-1",
				Image:   "nginx:latest",
			},
			{
				Service: "dind",
				Name:    "project-dind-1",
				Image:   "docker:dind",
			},
			{
				Service: "another-dind",
				Name:    "project-another-dind-1",
				Image:   "docker:20-dind",
			},
		}

		// Check that IsDind method works correctly
		assert.False(t, m.composeProcessListViewModel.composeContainers[0].IsDind())
		assert.True(t, m.composeProcessListViewModel.composeContainers[1].IsDind())
		assert.True(t, m.composeProcessListViewModel.composeContainers[2].IsDind())
	})
}

func TestComposeProcessListView_PortsDisplay(t *testing.T) {
	t.Run("formats ports correctly", func(t *testing.T) {
		container := models.ComposeContainer{
			Publishers: []struct {
				URL           string `json:"URL"`
				TargetPort    int    `json:"TargetPort"`
				PublishedPort int    `json:"PublishedPort"`
				Protocol      string `json:"Protocol"`
			}{
				{
					PublishedPort: 8080,
					TargetPort:    80,
					Protocol:      "tcp",
				},
				{
					PublishedPort: 8443,
					TargetPort:    443,
					Protocol:      "tcp",
				},
			},
		}

		portsStr := container.GetPortsString()
		assert.Contains(t, portsStr, "8080->80/tcp")
		assert.Contains(t, portsStr, "8443->443/tcp")
	})

	t.Run("handles no published port", func(t *testing.T) {
		container := models.ComposeContainer{
			Publishers: []struct {
				URL           string `json:"URL"`
				TargetPort    int    `json:"TargetPort"`
				PublishedPort int    `json:"PublishedPort"`
				Protocol      string `json:"Protocol"`
			}{
				{
					TargetPort: 3306,
					Protocol:   "tcp",
				},
			},
		}

		portsStr := container.GetPortsString()
		assert.Equal(t, "3306/tcp", portsStr)
	})

	t.Run("handles empty publishers", func(t *testing.T) {
		container := models.ComposeContainer{
			Publishers: []struct {
				URL           string `json:"URL"`
				TargetPort    int    `json:"TargetPort"`
				PublishedPort int    `json:"PublishedPort"`
				Protocol      string `json:"Protocol"`
			}{},
		}

		portsStr := container.GetPortsString()
		assert.Equal(t, "", portsStr)
	})
}
