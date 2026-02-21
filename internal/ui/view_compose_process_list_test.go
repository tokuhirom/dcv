package ui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
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
		m.composeProcessListViewModel.Cursor = 0
		m.width = 120
		m.Height = 30
		// Build rows for table rendering
		m.composeProcessListViewModel.SetRows(m.composeProcessListViewModel.buildRows(), m.ViewHeight())

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
		// Build rows for table rendering
		m.composeProcessListViewModel.SetRows(m.composeProcessListViewModel.buildRows(), m.ViewHeight())

		output := m.composeProcessListViewModel.render(m, 20)

		// Check for dind indicator (🔄 emoji)
		assert.Contains(t, output, "🔄")
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
		// Build rows for table rendering
		m.composeProcessListViewModel.SetRows(m.composeProcessListViewModel.buildRows(), m.ViewHeight())

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
		// Build rows for table rendering
		m.composeProcessListViewModel.SetRows(m.composeProcessListViewModel.buildRows(), m.ViewHeight())

		// The render function applies different styles to Up vs Exited containers
		output := m.composeProcessListViewModel.render(m, 20)
		assert.Contains(t, output, "Up")         // GetStatus() returns "Up" for running
		assert.Contains(t, output, "Exited (0)") // GetStatus() returns "Exited (0)" for exited with code 0
	})
}

func TestComposeProcessListView_LongStrings(t *testing.T) {
	longImage := strings.Repeat("registry.example.com/org/", 20) + "my-image:latest"
	longService := strings.Repeat("my-service-", 20)

	tests := []struct {
		name       string
		containers []models.ComposeContainer
		width      int
		height     int
	}{
		{
			name: "very long image name",
			containers: []models.ComposeContainer{
				{Service: "web", Image: longImage, State: "running"},
			},
			width: 80, height: 20,
		},
		{
			name: "very long service name",
			containers: []models.ComposeContainer{
				{Service: longService, Image: "nginx:latest", State: "running"},
			},
			width: 80, height: 20,
		},
		{
			name: "many ports",
			containers: []models.ComposeContainer{
				{
					Service: "web", Image: "nginx:latest", State: "running",
					Publishers: func() []struct {
						URL           string `json:"URL"`
						TargetPort    int    `json:"TargetPort"`
						PublishedPort int    `json:"PublishedPort"`
						Protocol      string `json:"Protocol"`
					} {
						var pubs []struct {
							URL           string `json:"URL"`
							TargetPort    int    `json:"TargetPort"`
							PublishedPort int    `json:"PublishedPort"`
							Protocol      string `json:"Protocol"`
						}
						for i := range 50 {
							pubs = append(pubs, struct {
								URL           string `json:"URL"`
								TargetPort    int    `json:"TargetPort"`
								PublishedPort int    `json:"PublishedPort"`
								Protocol      string `json:"Protocol"`
							}{PublishedPort: 8000 + i, TargetPort: 80 + i, Protocol: "tcp"})
						}
						return pubs
					}(),
				},
			},
			width: 80, height: 20,
		},
		{
			name: "all fields long simultaneously",
			containers: []models.ComposeContainer{
				{
					Service: longService, Image: longImage,
					State: "running", Name: strings.Repeat("name-", 50),
				},
			},
			width: 60, height: 20,
		},
		{
			name: "narrow terminal",
			containers: []models.ComposeContainer{
				{Service: "web", Image: longImage, State: "running"},
			},
			width: 30, height: 20,
		},
		{
			name: "very small height with many containers",
			containers: func() []models.ComposeContainer {
				var cs []models.ComposeContainer
				for i := range 20 {
					cs = append(cs, models.ComposeContainer{
						Service: fmt.Sprintf("service-%d-%s", i, strings.Repeat("x", 50)),
						Image:   longImage, State: "running",
					})
				}
				return cs
			}(),
			width: 80, height: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := createTestModel(ComposeProcessListView)
			m.width = tt.width
			m.Height = tt.height
			m.composeProcessListViewModel.composeContainers = tt.containers
			m.composeProcessListViewModel.SetRows(m.composeProcessListViewModel.buildRows(), m.ViewHeight())

			// Should not panic
			result := m.composeProcessListViewModel.render(m, tt.height-4)
			assert.NotEmpty(t, result)
		})
	}
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
		m.composeProcessListViewModel.SetRows(m.composeProcessListViewModel.buildRows(), m.ViewHeight())
		m.composeProcessListViewModel.Cursor = 0
		m.initializeKeyHandlers()

		// Test moving down
		m.composeProcessListViewModel.HandleDown(m)
		assert.Equal(t, 1, m.composeProcessListViewModel.Cursor)

		// Test moving down again
		m.composeProcessListViewModel.HandleDown(m)
		assert.Equal(t, 2, m.composeProcessListViewModel.Cursor)

		// Test moving down at the end (should stay at 2)
		m.composeProcessListViewModel.HandleDown(m)
		assert.Equal(t, 2, m.composeProcessListViewModel.Cursor)

		// Test moving up
		m.composeProcessListViewModel.HandleUp(m)
		assert.Equal(t, 1, m.composeProcessListViewModel.Cursor)

		// Test moving up again
		m.composeProcessListViewModel.HandleUp(m)
		assert.Equal(t, 0, m.composeProcessListViewModel.Cursor)

		// Test moving up at the beginning (should stay at 0)
		m.composeProcessListViewModel.HandleUp(m)
		assert.Equal(t, 0, m.composeProcessListViewModel.Cursor)
	})
}

func TestComposeProcessListView_KeyHandlers(t *testing.T) {
	t.Run("key handler registration", func(t *testing.T) {
		m := createTestModel(ComposeProcessListView)
		m.initializeKeyHandlers()

		// Check that compose process view handlers are registered
		assert.NotEmpty(t, m.composeProcessListViewHandlers)
		assert.NotEmpty(t, m.composeProcessListViewKeymap)

		// Check specific handlers exist
		hasKillHandler := false
		hasToggleHandler := false
		for _, config := range m.composeProcessListViewHandlers {
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
		m := createTestModel(ComposeProcessListView)
		m.initializeKeyHandlers()

		// Check view switching handlers exist in global handlers
		hasDockerSwitch := false
		hasProjectSwitch := false
		hasImageSwitch := false
		for _, config := range m.globalHandlers {
			if strings.Contains(config.Description, "docker ps") {
				hasDockerSwitch = true
			}
			if strings.Contains(config.Description, "project") {
				hasProjectSwitch = true
			}
			if strings.Contains(config.Description, "images") {
				hasImageSwitch = true
			}
		}
		// Check dind handler in view-specific handlers
		hasDindSwitch := false
		for _, config := range m.composeProcessListViewHandlers {
			if strings.Contains(strings.ToLower(config.Description), "dind") {
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
		m.composeProcessListViewModel.SetRows(m.composeProcessListViewModel.buildRows(), m.ViewHeight())
		m.composeProcessListViewModel.Cursor = 0
		m.initializeKeyHandlers()

		// Try to move up from first item
		cmd := m.composeProcessListViewModel.HandleUp(m)
		assert.Nil(t, cmd)
		assert.Equal(t, 0, m.composeProcessListViewModel.Cursor)

		// Move to last item
		m.composeProcessListViewModel.Cursor = 1

		// Try to move down from last item
		cmd = m.composeProcessListViewModel.HandleDown(m)
		assert.Nil(t, cmd)
		assert.Equal(t, 1, m.composeProcessListViewModel.Cursor)
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
		m.composeProcessListViewModel.projectName = "test-project"
		m.loading = false
		// Build rows for table rendering
		m.composeProcessListViewModel.SetRows(m.composeProcessListViewModel.buildRows(), m.ViewHeight())
		m.initializeKeyHandlers()

		// Test the View() method directly instead of using teatest
		output := m.View()

		// Check that the output contains expected content
		assert.Contains(t, output, "test-project")
		assert.Contains(t, output, "SERVICE")
		assert.Contains(t, output, "IMAGE")
		assert.Contains(t, output, "nginx:latest")
		assert.Contains(t, output, "web")
		assert.Contains(t, output, "Up") // Status shows as "Up" not "running"
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
