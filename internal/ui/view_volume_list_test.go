package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
)

func TestVolumeListViewModel_Rendering(t *testing.T) {
	tests := []struct {
		name      string
		viewModel VolumeListViewModel
		model     *Model
		height    int
		expected  []string
	}{
		{
			name: "displays no volumes message when empty",
			viewModel: VolumeListViewModel{
				dockerVolumes:        []models.DockerVolume{},
				selectedDockerVolume: 0,
			},
			model: &Model{
				width:  100,
				Height: 20,
			},
			height:   20,
			expected: []string{"No volumes found"},
		},
		{
			name: "displays volume list table",
			viewModel: VolumeListViewModel{
				dockerVolumes: []models.DockerVolume{
					{
						Name:   "postgres-data",
						Driver: "local",
						Scope:  "local",
					},
					{
						Name:   "redis-data",
						Driver: "local",
						Scope:  "local",
					},
					{
						Name:   "app-logs",
						Driver: "overlay",
						Scope:  "global",
					},
				},
				selectedDockerVolume: 0,
			},
			model: &Model{
				width:  100,
				Height: 20,
			},
			height: 20,
			expected: []string{
				"Name",
				"Driver",
				"Scope",
				"postgres-data",
				"redis-data",
				"app-logs",
				"local",
				"overlay",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.viewModel.render(tt.model, tt.height-4)

			for _, expected := range tt.expected {
				assert.Contains(t, result, expected, "Expected to find '%s' in output", expected)
			}
		})
	}
}

func TestVolumeListViewModel_Navigation(t *testing.T) {
	t.Run("HandleDown moves selection down", func(t *testing.T) {
		vm := &VolumeListViewModel{
			dockerVolumes: []models.DockerVolume{
				{Name: "volume-1"},
				{Name: "volume-2"},
				{Name: "volume-3"},
			},
			selectedDockerVolume: 0,
		}

		cmd := vm.HandleDown()
		assert.Nil(t, cmd)
		assert.Equal(t, 1, vm.selectedDockerVolume)

		// Test boundary
		vm.selectedDockerVolume = 2
		cmd = vm.HandleDown()
		assert.Nil(t, cmd)
		assert.Equal(t, 2, vm.selectedDockerVolume, "Should not go beyond last volume")
	})

	t.Run("HandleUp moves selection up", func(t *testing.T) {
		vm := &VolumeListViewModel{
			dockerVolumes: []models.DockerVolume{
				{Name: "volume-1"},
				{Name: "volume-2"},
				{Name: "volume-3"},
			},
			selectedDockerVolume: 2,
		}

		cmd := vm.HandleUp()
		assert.Nil(t, cmd)
		assert.Equal(t, 1, vm.selectedDockerVolume)

		// Test boundary
		vm.selectedDockerVolume = 0
		cmd = vm.HandleUp()
		assert.Nil(t, cmd)
		assert.Equal(t, 0, vm.selectedDockerVolume, "Should not go below 0")
	})
}

func TestVolumeListViewModel_Show(t *testing.T) {
	t.Run("Show switches to volume list view", func(t *testing.T) {
		model := &Model{
			dockerClient: docker.NewClient(),
			currentView:  ComposeProcessListView,
			loading:      false,
		}
		vm := &VolumeListViewModel{
			dockerVolumes: []models.DockerVolume{
				{Name: "existing-volume"},
			},
			selectedDockerVolume: 5,
		}

		cmd := vm.Show(model)
		assert.NotNil(t, cmd)
		assert.Equal(t, VolumeListView, model.currentView)
		assert.Equal(t, 0, vm.selectedDockerVolume, "Should reset selection")
		assert.Empty(t, vm.dockerVolumes, "Should reset volume list")
		assert.Nil(t, model.err, "Should clear error")
		assert.True(t, model.loading)
	})
}

func TestVolumeListViewModel_HandleInspect(t *testing.T) {
	t.Run("HandleInspect returns nil when no volumes", func(t *testing.T) {
		model := &Model{
			dockerClient: docker.NewClient(),
		}
		vm := &VolumeListViewModel{
			dockerVolumes:        []models.DockerVolume{},
			selectedDockerVolume: 0,
		}

		cmd := vm.HandleInspect(model)
		assert.Nil(t, cmd)
	})

	t.Run("HandleInspect returns nil when selection out of bounds", func(t *testing.T) {
		model := &Model{
			dockerClient: docker.NewClient(),
		}
		vm := &VolumeListViewModel{
			dockerVolumes: []models.DockerVolume{
				{Name: "volume-1"},
			},
			selectedDockerVolume: 5, // Out of bounds
		}

		cmd := vm.HandleInspect(model)
		assert.Nil(t, cmd)
	})

	t.Run("HandleInspect initiates inspection for selected volume", func(t *testing.T) {
		model := &Model{
			dockerClient:     docker.NewClient(),
			inspectViewModel: InspectViewModel{},
			loading:          false,
		}
		vm := &VolumeListViewModel{
			dockerVolumes: []models.DockerVolume{
				{Name: "volume-1"},
				{Name: "volume-2"},
			},
			selectedDockerVolume: 1,
		}

		cmd := vm.HandleInspect(model)
		assert.NotNil(t, cmd)
		assert.True(t, model.loading)
		assert.Nil(t, model.err)
	})
}

func TestVolumeListViewModel_HandleDelete(t *testing.T) {
	t.Run("HandleDelete returns nil when no volumes", func(t *testing.T) {
		model := &Model{
			dockerClient: docker.NewClient(),
		}
		vm := &VolumeListViewModel{
			dockerVolumes:        []models.DockerVolume{},
			selectedDockerVolume: 0,
		}

		cmd := vm.HandleDelete(model, false)
		assert.Nil(t, cmd)
	})

	t.Run("HandleDelete returns nil when selection out of bounds", func(t *testing.T) {
		model := &Model{
			dockerClient: docker.NewClient(),
		}
		vm := &VolumeListViewModel{
			dockerVolumes: []models.DockerVolume{
				{Name: "volume-1"},
			},
			selectedDockerVolume: 5, // Out of bounds
		}

		cmd := vm.HandleDelete(model, false)
		assert.Nil(t, cmd)
	})

	t.Run("HandleDelete initiates deletion for selected volume", func(t *testing.T) {
		model := &Model{
			dockerClient:              docker.NewClient(),
			commandExecutionViewModel: CommandExecutionViewModel{},
		}
		vm := &VolumeListViewModel{
			dockerVolumes: []models.DockerVolume{
				{Name: "volume-1"},
				{Name: "volume-2"},
			},
			selectedDockerVolume: 1,
		}

		cmd := vm.HandleDelete(model, false)
		// Returns nil because it needs confirmation first (aggressive command)
		assert.Nil(t, cmd)
		assert.True(t, model.commandExecutionViewModel.pendingConfirmation)
	})

	t.Run("HandleDelete with force flag", func(t *testing.T) {
		model := &Model{
			dockerClient:              docker.NewClient(),
			commandExecutionViewModel: CommandExecutionViewModel{},
		}
		vm := &VolumeListViewModel{
			dockerVolumes: []models.DockerVolume{
				{Name: "volume-1"},
			},
			selectedDockerVolume: 0,
		}

		cmd := vm.HandleDelete(model, true) // Force delete
		// Still returns nil because it needs confirmation (aggressive command)
		assert.Nil(t, cmd)
		assert.True(t, model.commandExecutionViewModel.pendingConfirmation)
	})
}

func TestVolumeListViewModel_HandleBack(t *testing.T) {
	t.Run("HandleBack returns to previous view", func(t *testing.T) {
		model := &Model{
			currentView: VolumeListView,
			viewHistory: []ViewType{ComposeProcessListView, VolumeListView},
		}
		vm := &VolumeListViewModel{}

		cmd := vm.HandleBack(model)
		assert.Nil(t, cmd)
		assert.Equal(t, ComposeProcessListView, model.currentView)
	})
}

func TestVolumeListViewModel_DoLoad(t *testing.T) {
	t.Run("DoLoad returns command to load volumes", func(t *testing.T) {
		model := &Model{
			dockerClient: docker.NewClient(),
			loading:      false,
		}
		vm := &VolumeListViewModel{}

		cmd := vm.DoLoad(model)
		assert.NotNil(t, cmd)
		assert.True(t, model.loading)
	})
}

func TestVolumeListViewModel_Loaded(t *testing.T) {
	t.Run("Loaded updates volume list", func(t *testing.T) {
		vm := &VolumeListViewModel{
			selectedDockerVolume: 10, // Out of bounds
		}

		volumes := []models.DockerVolume{
			{Name: "volume-1"},
			{Name: "volume-2"},
		}

		vm.Loaded(volumes)
		assert.Equal(t, volumes, vm.dockerVolumes)
		assert.Equal(t, 0, vm.selectedDockerVolume, "Should reset selection when out of bounds")
	})

	t.Run("Loaded preserves valid selection", func(t *testing.T) {
		vm := &VolumeListViewModel{
			selectedDockerVolume: 1,
		}

		volumes := []models.DockerVolume{
			{Name: "volume-1"},
			{Name: "volume-2"},
			{Name: "volume-3"},
		}

		vm.Loaded(volumes)
		assert.Equal(t, volumes, vm.dockerVolumes)
		assert.Equal(t, 1, vm.selectedDockerVolume, "Should preserve valid selection")
	})

	t.Run("Loaded handles empty volume list", func(t *testing.T) {
		vm := &VolumeListViewModel{
			selectedDockerVolume: 5,
		}

		volumes := []models.DockerVolume{}

		vm.Loaded(volumes)
		assert.Equal(t, volumes, vm.dockerVolumes)
		assert.Equal(t, 5, vm.selectedDockerVolume, "Selection unchanged when list is empty")
	})
}
