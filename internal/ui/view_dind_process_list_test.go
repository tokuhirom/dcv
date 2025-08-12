package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
)

func TestDindProcessListViewModel_Rendering(t *testing.T) {
	tests := []struct {
		name      string
		viewModel DindProcessListViewModel
		height    int
		expected  []string
	}{
		{
			name: "displays no containers message when empty",
			viewModel: DindProcessListViewModel{
				dindContainers: []models.DockerContainer{},
				hostContainer:  docker.NewDindContainer("host-1", "host-container", "container-1", "test", "running"),
			},
			height:   20,
			expected: []string{"No containers running inside this dind container"},
		},
		{
			name: "displays container list table",
			viewModel: DindProcessListViewModel{
				dindContainers: []models.DockerContainer{
					{
						ID:     "abc123def456",
						Image:  "nginx:latest",
						State:  "running",
						Status: "Up 5 minutes",
						Names:  "web-server",
					},
					{
						ID:     "xyz789ghi012",
						Image:  "postgres:13",
						State:  "running",
						Status: "Up 10 minutes",
						Names:  "database",
					},
				},
				selectedDindContainer: 0,
				hostContainer:         docker.NewDindContainer("host-1", "host-container", "container-1", "test", "running"),
			},
			height: 20,
			expected: []string{
				"CONTAINER ID",
				"IMAGE",
				"STATE",
				"STATUS",
				"NAMES",
				"abc123def456", // First 12 chars
				"nginx:latest",
				"web-server",
			},
		},
		{
			name: "truncates long image names",
			viewModel: DindProcessListViewModel{
				dindContainers: []models.DockerContainer{
					{
						ID:     "abc123def456",
						Image:  "very-long-registry-url.example.com/organization/project/image:latest",
						State:  "running",
						Status: "Up 5 minutes",
						Names:  "test-container",
					},
				},
				selectedDindContainer: 0,
				hostContainer:         docker.NewDindContainer("host-1", "host-container", "container-1", "test", "running"),
			},
			height:   20,
			expected: []string{"very-long-registry-url..."},
		},
		{
			name: "displays port mappings",
			viewModel: DindProcessListViewModel{
				dindContainers: []models.DockerContainer{
					{
						ID:     "def456ghi789",
						Image:  "nginx:latest",
						State:  "running",
						Status: "Up 10 minutes",
						Ports:  "0.0.0.0:8080->80/tcp",
						Names:  "web-server",
					},
				},
				selectedDindContainer: 0,
				hostContainer:         docker.NewDindContainer("host-1", "host-container", "container-1", "test", "running"),
			},
			height:   20,
			expected: []string{"PORTS", "0.0.0.0:8080->80/tcp"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.viewModel.render(tt.height - 4)

			for _, expected := range tt.expected {
				assert.Contains(t, result, expected, "Expected to find '%s' in output", expected)
			}
		})
	}
}

func TestDindProcessListViewModel_Navigation(t *testing.T) {
	t.Run("HandleDown moves selection down", func(t *testing.T) {
		vm := &DindProcessListViewModel{
			dindContainers: []models.DockerContainer{
				{ID: "1", Names: "container-1"},
				{ID: "2", Names: "container-2"},
				{ID: "3", Names: "container-3"},
			},
			selectedDindContainer: 0,
		}

		cmd := vm.HandleDown()
		assert.Nil(t, cmd)
		assert.Equal(t, 1, vm.selectedDindContainer)

		// Test boundary
		vm.selectedDindContainer = 2
		cmd = vm.HandleDown()
		assert.Nil(t, cmd)
		assert.Equal(t, 2, vm.selectedDindContainer, "Should not go beyond last container")
	})

	t.Run("HandleUp moves selection up", func(t *testing.T) {
		vm := &DindProcessListViewModel{
			dindContainers: []models.DockerContainer{
				{ID: "1", Names: "container-1"},
				{ID: "2", Names: "container-2"},
				{ID: "3", Names: "container-3"},
			},
			selectedDindContainer: 2,
		}

		cmd := vm.HandleUp()
		assert.Nil(t, cmd)
		assert.Equal(t, 1, vm.selectedDindContainer)

		// Test boundary
		vm.selectedDindContainer = 0
		cmd = vm.HandleUp()
		assert.Nil(t, cmd)
		assert.Equal(t, 0, vm.selectedDindContainer, "Should not go below 0")
	})
}

func TestDindProcessListViewModel_Load(t *testing.T) {
	t.Run("Load initializes dind process list", func(t *testing.T) {
		model := &Model{
			dockerClient: docker.NewClient(),
			currentView:  ComposeProcessListView,
			loading:      false,
		}
		vm := &DindProcessListViewModel{}
		hostContainer := docker.NewDindContainer("host-1", "host-container", "container-1", "test", "running")

		cmd := vm.Load(model, hostContainer)
		assert.NotNil(t, cmd)
		assert.Equal(t, hostContainer, vm.hostContainer)
		assert.Equal(t, DindProcessListView, model.currentView)
		assert.True(t, model.loading)
	})
}

func TestDindProcessListViewModel_ToggleAll(t *testing.T) {
	t.Run("toggles showAll state and triggers reload", func(t *testing.T) {
		model := &Model{
			dockerClient: docker.NewClient(),
			loading:      false,
		}
		vm := &DindProcessListViewModel{
			showAll:       false,
			hostContainer: docker.NewDindContainer("host-1", "host-container", "container-1", "test", "running"),
		}

		// Test direct method call
		cmd := vm.HandleToggleAll(model)
		assert.True(t, vm.showAll, "showAll should be toggled to true")
		assert.NotNil(t, cmd, "Should return a command to trigger reload")
		assert.True(t, model.loading, "Should set loading to true")

		// Reset loading for next test
		model.loading = false

		// Toggle again
		cmd = vm.HandleToggleAll(model)
		assert.False(t, vm.showAll, "showAll should be toggled back to false")
		assert.NotNil(t, cmd, "Should return a command to trigger reload")
		assert.True(t, model.loading, "Should set loading to true")
	})

	t.Run("CmdToggleAll works with DinD view", func(t *testing.T) {
		model := &Model{
			currentView:  DindProcessListView,
			dockerClient: docker.NewClient(),
			loading:      false,
		}
		model.dindProcessListViewModel = DindProcessListViewModel{
			showAll:       false,
			hostContainer: docker.NewDindContainer("host-1", "host-container", "container-1", "test", "running"),
		}
		model.initializeKeyHandlers()

		// Test toggle via CmdToggleAll
		_, cmd := model.CmdToggleAll(tea.KeyMsg{})
		assert.True(t, model.dindProcessListViewModel.showAll, "showAll should be toggled to true")
		assert.NotNil(t, cmd, "Should return a command to trigger reload")
		assert.True(t, model.loading, "Should set loading to true")

		// Reset loading for next test
		model.loading = false

		// Toggle again
		_, cmd = model.CmdToggleAll(tea.KeyMsg{})
		assert.False(t, model.dindProcessListViewModel.showAll, "showAll should be toggled back to false")
		assert.NotNil(t, cmd, "Should return a command to trigger reload")
		assert.True(t, model.loading, "Should set loading to true")
	})
}

func TestDindProcessListViewModel_HandleBack(t *testing.T) {
	t.Run("HandleBack returns to previous view", func(t *testing.T) {
		model := &Model{
			currentView: DindProcessListView,
			viewHistory: []ViewType{ComposeProcessListView, DindProcessListView},
		}
		vm := &DindProcessListViewModel{}

		cmd := vm.HandleBack(model)
		assert.Nil(t, cmd)
		assert.Equal(t, ComposeProcessListView, model.currentView)
	})
}

func TestDindProcessListViewModel_Loaded(t *testing.T) {
	t.Run("Loaded updates container list", func(t *testing.T) {
		vm := &DindProcessListViewModel{
			selectedDindContainer: 10, // Out of bounds
		}

		containers := []models.DockerContainer{
			{ID: "1", Names: "container-1"},
			{ID: "2", Names: "container-2"},
		}

		vm.Loaded(containers)
		assert.Equal(t, containers, vm.dindContainers)
		assert.Equal(t, 0, vm.selectedDindContainer, "Should reset selection when out of bounds")
	})

	t.Run("Loaded preserves valid selection", func(t *testing.T) {
		vm := &DindProcessListViewModel{
			selectedDindContainer: 1,
		}

		containers := []models.DockerContainer{
			{ID: "1", Names: "container-1"},
			{ID: "2", Names: "container-2"},
			{ID: "3", Names: "container-3"},
		}

		vm.Loaded(containers)
		assert.Equal(t, containers, vm.dindContainers)
		assert.Equal(t, 1, vm.selectedDindContainer, "Should preserve valid selection")
	})
}

func TestDindProcessListViewModel_GetContainer(t *testing.T) {
	t.Run("GetContainer returns selected container", func(t *testing.T) {
		model := &Model{
			dockerClient: docker.NewClient(),
		}
		vm := &DindProcessListViewModel{
			dindContainers: []models.DockerContainer{
				{ID: "abc123", Names: "container-1"},
				{ID: "def456", Names: "container-2"},
			},
			selectedDindContainer: 1,
			hostContainer:         docker.NewDindContainer("host-1", "host-container", "container-1", "test", "running"),
		}

		container := vm.GetContainer(model)
		assert.NotNil(t, container)
		assert.Equal(t, "def456", container.GetContainerID())
	})

	t.Run("GetContainer returns nil when selection out of bounds", func(t *testing.T) {
		model := &Model{
			dockerClient: docker.NewClient(),
		}
		vm := &DindProcessListViewModel{
			dindContainers:        []models.DockerContainer{},
			selectedDindContainer: 0,
		}

		container := vm.GetContainer(model)
		assert.Nil(t, container)
	})
}

func TestDindProcessListViewModel_Title(t *testing.T) {
	t.Run("normal title without all", func(t *testing.T) {
		vm := &DindProcessListViewModel{
			showAll:       false,
			hostContainer: docker.NewDindContainer("host-1", "my-dind-container", "container-1", "test", "running"),
		}

		title := vm.Title()
		assert.Equal(t, "Docker in Docker: test", title)
	})

	t.Run("title with all indicator when showAll is true", func(t *testing.T) {
		vm := &DindProcessListViewModel{
			showAll:       true,
			hostContainer: docker.NewDindContainer("host-1", "my-dind-container", "container-1", "test", "running"),
		}

		title := vm.Title()
		assert.Equal(t, "Docker in Docker: test (all)", title)
	})
}

func TestDindProcessListViewModel_DoLoad(t *testing.T) {
	t.Run("DoLoad returns command to load containers", func(t *testing.T) {
		model := &Model{
			dockerClient: docker.NewClient(),
		}
		vm := &DindProcessListViewModel{
			hostContainer: docker.NewDindContainer("host-1", "host-container", "container-1", "test", "running"),
		}

		cmd := vm.DoLoad(model)
		assert.NotNil(t, cmd)
		assert.True(t, model.loading)
	})
}
