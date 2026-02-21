package ui

import (
	"strings"
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
				dockerVolumes:  []models.DockerVolume{},
				TableViewModel: TableViewModel{Cursor: 0},
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
						Size:   "183.8kB",
					},
					{
						Name:   "redis-data",
						Driver: "local",
						Scope:  "local",
						Size:   "1.234GB",
					},
					{
						Name:   "app-logs",
						Driver: "overlay",
						Scope:  "global",
						Size:   "0B",
					},
				},
				TableViewModel: TableViewModel{Cursor: 0},
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
				"Size",
				"postgres-data",
				"redis-data",
				"app-logs",
				"local",
				"overlay",
				"183.8kB",
				"1.234GB",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build rows for table
			tt.viewModel.SetRows(tt.viewModel.buildRows(), tt.model.ViewHeight())
			result := tt.viewModel.render(tt.model, tt.height-4)

			for _, expected := range tt.expected {
				assert.Contains(t, result, expected, "Expected to find '%s' in output", expected)
			}
		})
	}
}

func TestVolumeListViewModel_LongStrings(t *testing.T) {
	longName := strings.Repeat("my-very-long-volume-name-", 20)
	longDriver := strings.Repeat("custom-driver-", 20)
	longSize := strings.Repeat("999.9GB", 20)

	tests := []struct {
		name    string
		volumes []models.DockerVolume
		width   int
		height  int
	}{
		{
			name: "very long volume name",
			volumes: []models.DockerVolume{
				{Name: longName, Driver: "local", Scope: "local", Size: "100MB"},
			},
			width: 80, height: 20,
		},
		{
			name: "very long driver name",
			volumes: []models.DockerVolume{
				{Name: "vol", Driver: longDriver, Scope: "local", Size: "100MB"},
			},
			width: 80, height: 20,
		},
		{
			name: "very long size string",
			volumes: []models.DockerVolume{
				{Name: "vol", Driver: "local", Scope: "local", Size: longSize},
			},
			width: 80, height: 20,
		},
		{
			name: "all fields long simultaneously",
			volumes: []models.DockerVolume{
				{Name: longName, Driver: longDriver, Scope: strings.Repeat("scope-", 20), Size: longSize},
			},
			width: 60, height: 20,
		},
		{
			name: "narrow terminal",
			volumes: []models.DockerVolume{
				{Name: longName, Driver: "local", Scope: "local", Size: "100MB"},
			},
			width: 30, height: 20,
		},
		{
			name: "very small height with many volumes",
			volumes: func() []models.DockerVolume {
				var vols []models.DockerVolume
				for range 20 {
					vols = append(vols, models.DockerVolume{
						Name: longName, Driver: "local", Scope: "local", Size: "100MB",
					})
				}
				return vols
			}(),
			width: 80, height: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := &VolumeListViewModel{
				dockerVolumes:  tt.volumes,
				TableViewModel: TableViewModel{Cursor: 0},
			}
			model := &Model{width: tt.width, Height: tt.height}
			vm.SetRows(vm.buildRows(), model.ViewHeight())

			// Should not panic
			result := vm.render(model, tt.height-4)
			assert.NotEmpty(t, result)
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
			TableViewModel: TableViewModel{Cursor: 0},
		}

		model := &Model{Height: 20}
		vm.SetRows(vm.buildRows(), model.ViewHeight())
		cmd := vm.HandleDown(model)
		assert.Nil(t, cmd)
		assert.Equal(t, 1, vm.Cursor)

		// Test boundary
		vm.Cursor = 2
		cmd = vm.HandleDown(model)
		assert.Nil(t, cmd)
		assert.Equal(t, 2, vm.Cursor, "Should not go beyond last volume")
	})

	t.Run("HandleUp moves selection up", func(t *testing.T) {
		vm := &VolumeListViewModel{
			dockerVolumes: []models.DockerVolume{
				{Name: "volume-1"},
				{Name: "volume-2"},
				{Name: "volume-3"},
			},
			TableViewModel: TableViewModel{Cursor: 2},
		}

		model := &Model{Height: 20}
		vm.SetRows(vm.buildRows(), model.ViewHeight())
		cmd := vm.HandleUp(model)
		assert.Nil(t, cmd)
		assert.Equal(t, 1, vm.Cursor)

		// Test boundary
		vm.Cursor = 0
		cmd = vm.HandleUp(model)
		assert.Nil(t, cmd)
		assert.Equal(t, 0, vm.Cursor, "Should not go below 0")
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
			TableViewModel: TableViewModel{Cursor: 5},
		}

		cmd := vm.Show(model)
		assert.NotNil(t, cmd)
		assert.Equal(t, VolumeListView, model.currentView)
		assert.Equal(t, 0, vm.Cursor, "Should reset selection")
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
			dockerVolumes:  []models.DockerVolume{},
			TableViewModel: TableViewModel{Cursor: 0},
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
			TableViewModel: TableViewModel{Cursor: 5}, // Out of bounds
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
			TableViewModel: TableViewModel{Cursor: 1},
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
			dockerVolumes:  []models.DockerVolume{},
			TableViewModel: TableViewModel{Cursor: 0},
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
			TableViewModel: TableViewModel{Cursor: 5}, // Out of bounds
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
			TableViewModel: TableViewModel{Cursor: 1},
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
			TableViewModel: TableViewModel{Cursor: 0},
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
			TableViewModel: TableViewModel{Cursor: 10}, // Out of bounds
		}

		volumes := []models.DockerVolume{
			{Name: "volume-1"},
			{Name: "volume-2"},
		}

		model := &Model{Height: 20}
		vm.Loaded(model, volumes)
		assert.Equal(t, volumes, vm.dockerVolumes)
		assert.Equal(t, 0, vm.Cursor, "Should reset selection when out of bounds")
	})

	t.Run("Loaded preserves valid selection", func(t *testing.T) {
		vm := &VolumeListViewModel{
			TableViewModel: TableViewModel{Cursor: 1},
		}

		volumes := []models.DockerVolume{
			{Name: "volume-1"},
			{Name: "volume-2"},
			{Name: "volume-3"},
		}

		model := &Model{Height: 20}
		vm.Loaded(model, volumes)
		assert.Equal(t, volumes, vm.dockerVolumes)
		assert.Equal(t, 1, vm.Cursor, "Should preserve valid selection")
	})

	t.Run("Loaded handles empty volume list", func(t *testing.T) {
		vm := &VolumeListViewModel{
			TableViewModel: TableViewModel{Cursor: 5},
		}

		volumes := []models.DockerVolume{}

		model := &Model{Height: 20}
		vm.Loaded(model, volumes)
		assert.Equal(t, volumes, vm.dockerVolumes)
		assert.Equal(t, 0, vm.Cursor, "Selection reset when list is empty")
	})
}
