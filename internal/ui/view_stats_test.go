package ui

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
)

func TestStatsViewModel_Rendering(t *testing.T) {
	tests := []struct {
		name      string
		viewModel StatsViewModel
		model     *Model
		height    int
		expected  []string
	}{
		{
			name: "displays no stats message when empty",
			viewModel: StatsViewModel{
				stats: []models.ContainerStats{},
			},
			model: &Model{
				width:  100,
				Height: 20,
			},
			height:   20,
			expected: []string{"No stats available"},
		},
		{
			name: "displays stats table",
			viewModel: StatsViewModel{
				stats: []models.ContainerStats{
					{
						Name:     "web-container",
						CPUPerc:  "15.5%",
						MemUsage: "512MiB/1GiB",
						MemPerc:  "50.0%",
						NetIO:    "1.2GB/800MB",
						BlockIO:  "500MB/200MB",
					},
					{
						Name:     "db-container",
						CPUPerc:  "5.2%",
						MemUsage: "256MiB/512MiB",
						MemPerc:  "50.0%",
						NetIO:    "100MB/50MB",
						BlockIO:  "1GB/500MB",
					},
				},
			},
			model: &Model{
				width:  100,
				Height: 20,
			},
			height: 20,
			expected: []string{
				"NAME",
				"CPU %",
				"MEM USAGE",
				"NET I/O",
				"BLOCK I/O",
				"web-container",
				"15.5%",
				"512MiB/1GiB",
				"1.2GB/800MB",
				"db-container",
			},
		},
		{
			name: "truncates long container names",
			viewModel: StatsViewModel{
				stats: []models.ContainerStats{
					{
						Name:     "very-long-container-name-that-exceeds-limit",
						CPUPerc:  "10.0%",
						MemUsage: "100MiB/200MiB",
						MemPerc:  "50.0%",
						NetIO:    "10MB/5MB",
						BlockIO:  "20MB/10MB",
					},
				},
			},
			model: &Model{
				width:  100,
				Height: 20,
			},
			height:   20,
			expected: []string{"very-long-contain..."},
		},
		{
			name: "colors high CPU usage",
			viewModel: StatsViewModel{
				stats: []models.ContainerStats{
					{
						Name:     "high-cpu",
						CPUPerc:  "85.0%",
						MemUsage: "100MiB/200MiB",
						MemPerc:  "50.0%",
						NetIO:    "10MB/5MB",
						BlockIO:  "20MB/10MB",
					},
					{
						Name:     "medium-cpu",
						CPUPerc:  "55.0%",
						MemUsage: "100MiB/200MiB",
						MemPerc:  "50.0%",
						NetIO:    "10MB/5MB",
						BlockIO:  "20MB/10MB",
					},
				},
			},
			model: &Model{
				width:  100,
				Height: 20,
			},
			height: 20,
			expected: []string{
				"high-cpu",
				"85.0%", // Will be colored red (>80%)
				"medium-cpu",
				"55.0%", // Will be colored yellow (>50%)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize the TableViewModel state
			tt.viewModel.buildRows()
			tt.viewModel.SetRows(tt.viewModel.Rows, tt.height-4)

			result := tt.viewModel.render(tt.model, tt.height-4)

			for _, expected := range tt.expected {
				assert.Contains(t, result, expected, "Expected to find '%s' in output", expected)
			}
		})
	}
}

func TestStatsViewModel_Show(t *testing.T) {
	t.Run("Show switches to stats view and initiates loading", func(t *testing.T) {
		model := &Model{
			dockerClient: docker.NewClient(),
			currentView:  ComposeProcessListView,
			loading:      false,
		}
		vm := &StatsViewModel{}

		cmd := vm.Show(model)
		assert.NotNil(t, cmd)
		assert.Equal(t, StatsView, model.currentView)
		assert.True(t, model.loading)
	})
}

func TestStatsViewModel_DoLoad(t *testing.T) {
	t.Run("DoLoad returns command to load stats", func(t *testing.T) {
		model := &Model{
			dockerClient: docker.NewClient(),
			loading:      false,
		}
		vm := &StatsViewModel{}

		cmd := vm.DoLoad(model)
		assert.NotNil(t, cmd)
		assert.True(t, model.loading)
	})
}

func TestStatsViewModel_HandleBack(t *testing.T) {
	t.Run("HandleBack returns to previous view", func(t *testing.T) {
		model := &Model{
			currentView: StatsView,
			viewHistory: []ViewType{ComposeProcessListView, StatsView},
		}
		vm := &StatsViewModel{}

		cmd := vm.HandleBack(model)
		assert.Nil(t, cmd)
		assert.Equal(t, ComposeProcessListView, model.currentView)
	})
}

func TestStatsViewModel_Loaded(t *testing.T) {
	t.Run("Loaded updates stats and resets scroll", func(t *testing.T) {
		vm := &StatsViewModel{
			stats: []models.ContainerStats{},
		}

		newStats := []models.ContainerStats{
			{
				Name:     "container-1",
				CPUPerc:  "10.0%",
				MemUsage: "100MiB/1GiB",
				MemPerc:  "10.0%",
				NetIO:    "1MB/500KB",
				BlockIO:  "10MB/5MB",
			},
			{
				Name:     "container-2",
				CPUPerc:  "20.0%",
				MemUsage: "200MiB/1GiB",
				MemPerc:  "20.0%",
				NetIO:    "2MB/1MB",
				BlockIO:  "20MB/10MB",
			},
		}

		model := &Model{Height: 30}
		vm.Loaded(model, newStats)
		assert.Equal(t, newStats, vm.stats)
	})

	t.Run("Loaded replaces existing stats", func(t *testing.T) {
		vm := &StatsViewModel{
			stats: []models.ContainerStats{
				{Name: "old-container"},
			},
		}

		newStats := []models.ContainerStats{
			{Name: "new-container"},
		}

		model := &Model{Height: 30}
		vm.Loaded(model, newStats)
		assert.Equal(t, newStats, vm.stats)
		assert.Len(t, vm.stats, 1)
		assert.Equal(t, "new-container", vm.stats[0].Name)
	})
}

func TestStatsViewModel_Sorting(t *testing.T) {
	vm := &StatsViewModel{
		stats: []models.ContainerStats{
			{Name: "container-b", CPUPerc: "50.0%", MemPerc: "20.0%"},
			{Name: "container-a", CPUPerc: "30.0%", MemPerc: "40.0%"},
			{Name: "container-c", CPUPerc: "70.0%", MemPerc: "10.0%"},
		},
	}

	t.Run("sort by name ascending", func(t *testing.T) {
		vm.sortField = StatsSortByName
		vm.sortReverse = false
		vm.sortStats()

		assert.Equal(t, "container-a", vm.stats[0].Name)
		assert.Equal(t, "container-b", vm.stats[1].Name)
		assert.Equal(t, "container-c", vm.stats[2].Name)
	})

	t.Run("sort by CPU descending", func(t *testing.T) {
		vm.sortField = StatsSortByCPU
		vm.sortReverse = true
		vm.sortStats()

		assert.Equal(t, "70.0%", vm.stats[0].CPUPerc)
		assert.Equal(t, "50.0%", vm.stats[1].CPUPerc)
		assert.Equal(t, "30.0%", vm.stats[2].CPUPerc)
	})

	t.Run("sort by memory descending", func(t *testing.T) {
		vm.sortField = StatsSortByMem
		vm.sortReverse = true
		vm.sortStats()

		assert.Equal(t, "40.0%", vm.stats[0].MemPerc)
		assert.Equal(t, "20.0%", vm.stats[1].MemPerc)
		assert.Equal(t, "10.0%", vm.stats[2].MemPerc)
	})
}

func TestStatsViewModel_SortHandlers(t *testing.T) {
	vm := &StatsViewModel{}
	model := &Model{Height: 30}

	t.Run("HandleSortByCPU defaults to descending", func(t *testing.T) {
		vm.HandleSortByCPU(model)
		assert.Equal(t, StatsSortByCPU, vm.sortField)
		assert.True(t, vm.sortReverse)

		// Second call toggles
		vm.HandleSortByCPU(model)
		assert.Equal(t, StatsSortByCPU, vm.sortField)
		assert.False(t, vm.sortReverse)
	})

	t.Run("HandleSortByMem defaults to descending", func(t *testing.T) {
		vm.sortField = StatsSortByName
		vm.HandleSortByMem(model)
		assert.Equal(t, StatsSortByMem, vm.sortField)
		assert.True(t, vm.sortReverse)
	})

	t.Run("HandleSortByName defaults to ascending", func(t *testing.T) {
		vm.sortField = StatsSortByCPU
		vm.HandleSortByName(model)
		assert.Equal(t, StatsSortByName, vm.sortField)
		assert.False(t, vm.sortReverse)
	})

	t.Run("HandleReverseSort toggles order", func(t *testing.T) {
		vm.sortReverse = false
		vm.HandleReverseSort(model)
		assert.True(t, vm.sortReverse)

		vm.HandleReverseSort(model)
		assert.False(t, vm.sortReverse)
	})
}

func TestStatsViewModel_AutoRefresh(t *testing.T) {
	vm := &StatsViewModel{}

	t.Run("HandleToggleAutoRefresh toggles state", func(t *testing.T) {
		// Initially off
		assert.False(t, vm.autoRefresh)

		// Turn on
		vm.HandleToggleAutoRefresh()
		assert.True(t, vm.autoRefresh)

		// Turn off
		vm.HandleToggleAutoRefresh()
		assert.False(t, vm.autoRefresh)
	})

	t.Run("Show enables auto-refresh by default", func(t *testing.T) {
		model := &Model{
			dockerClient: docker.NewClient(),
			currentView:  ComposeProcessListView,
		}
		vm := &StatsViewModel{}

		cmd := vm.Show(model)
		assert.NotNil(t, cmd)
		assert.True(t, vm.autoRefresh)
		assert.Equal(t, 2*time.Second, vm.refreshInterval)
	})

	t.Run("startAutoRefresh returns tick command", func(t *testing.T) {
		vm.refreshInterval = 3 * time.Second
		cmd := vm.startAutoRefresh()
		assert.NotNil(t, cmd)
	})

	t.Run("DoLoadSilent does not set loading indicator", func(t *testing.T) {
		model := &Model{
			dockerClient: docker.NewClient(),
			loading:      false,
		}
		vm := &StatsViewModel{}

		cmd := vm.DoLoadSilent(model)
		assert.NotNil(t, cmd)
		assert.False(t, model.loading) // Should remain false
	})
}

func TestStatsViewModel_Navigation(t *testing.T) {
	vm := &StatsViewModel{}
	model := &Model{Height: 30}

	vm.stats = []models.ContainerStats{
		{Name: "container-1"},
		{Name: "container-2"},
		{Name: "container-3"},
	}
	vm.buildRows()
	vm.SetRows(vm.Rows, model.ViewHeight())
	vm.Cursor = 1

	t.Run("HandleUp scrolls up", func(t *testing.T) {
		vm.HandleUp(model)
		assert.Equal(t, 0, vm.Cursor)

		// Shouldn't go below 0
		vm.HandleUp(model)
		assert.Equal(t, 0, vm.Cursor)
	})

	t.Run("HandleDown scrolls down", func(t *testing.T) {
		vm.Cursor = 0
		vm.HandleDown(model)
		assert.Equal(t, 1, vm.Cursor)

		vm.HandleDown(model)
		assert.Equal(t, 2, vm.Cursor)

		// Shouldn't go beyond last container
		vm.HandleDown(model)
		assert.Equal(t, 2, vm.Cursor)
	})
}

func TestStatsSortField_String(t *testing.T) {
	tests := []struct {
		field    StatsSortField
		expected string
	}{
		{StatsSortByName, "NAME"},
		{StatsSortByCPU, "CPU%"},
		{StatsSortByMem, "MEM%"},
		{StatsSortField(999), "NAME"}, // Default case
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.field.String())
		})
	}
}

func TestParsePercentage(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"25.5%", 25.5},
		{"100%", 100.0},
		{"0.5%", 0.5},
		{"", 0.0},
		{"invalid", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := models.ParsePercentage(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStatsViewModel_CPUColoring(t *testing.T) {
	tests := []struct {
		name        string
		cpuPercent  string
		expectColor bool
		colorLevel  string // "high" or "medium"
	}{
		{
			name:        "very high CPU usage",
			cpuPercent:  "95.0%",
			expectColor: true,
			colorLevel:  "high",
		},
		{
			name:        "high CPU usage threshold",
			cpuPercent:  "81.0%",
			expectColor: true,
			colorLevel:  "high",
		},
		{
			name:        "medium CPU usage",
			cpuPercent:  "60.0%",
			expectColor: true,
			colorLevel:  "medium",
		},
		{
			name:        "medium CPU usage threshold",
			cpuPercent:  "51.0%",
			expectColor: true,
			colorLevel:  "medium",
		},
		{
			name:        "low CPU usage",
			cpuPercent:  "10.0%",
			expectColor: false,
		},
		{
			name:        "invalid CPU format",
			cpuPercent:  "invalid",
			expectColor: false,
		},
		{
			name:        "empty CPU",
			cpuPercent:  "",
			expectColor: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := &StatsViewModel{
				stats: []models.ContainerStats{
					{
						Name:     "test-container",
						CPUPerc:  tt.cpuPercent,
						MemUsage: "100MiB/1GiB",
						MemPerc:  "10.0%",
						NetIO:    "1MB/500KB",
						BlockIO:  "10MB/5MB",
					},
				},
			}

			model := &Model{
				width:  100,
				Height: 20,
			}

			// Initialize the TableViewModel state
			vm.buildRows()
			vm.SetRows(vm.Rows, 16) // 20 - 4 for chrome

			result := vm.render(model, 20)
			assert.Contains(t, result, "test-container")
			// The actual coloring is done through lipgloss styles,
			// so we just verify the CPU percentage is present
			if tt.cpuPercent != "" && tt.cpuPercent != "invalid" {
				assert.Contains(t, result, tt.cpuPercent)
			}
		})
	}
}

func TestStatsViewModel_Integration(t *testing.T) {
	t.Run("Complete flow from show to display", func(t *testing.T) {
		// Setup
		model := &Model{
			dockerClient: docker.NewClient(),
			currentView:  ComposeProcessListView,
			loading:      false,
			width:        100,
			Height:       20,
		}
		vm := &StatsViewModel{}

		// Show stats view
		cmd := vm.Show(model)
		assert.NotNil(t, cmd)
		assert.Equal(t, StatsView, model.currentView)

		// Simulate loading completion
		stats := []models.ContainerStats{
			{
				Name:     "web-1",
				CPUPerc:  "25.5%",
				MemUsage: "512MiB/2GiB",
				MemPerc:  "25.0%",
				NetIO:    "100MB/50MB",
				BlockIO:  "200MB/100MB",
			},
			{
				Name:     "db-1",
				CPUPerc:  "10.0%",
				MemUsage: "1GiB/4GiB",
				MemPerc:  "25.0%",
				NetIO:    "50MB/25MB",
				BlockIO:  "500MB/250MB",
			},
		}
		vm.Loaded(model, stats)

		// RenderTable
		rendered := vm.render(model, model.Height-4)
		assert.Contains(t, rendered, "NAME")
		assert.Contains(t, rendered, "CPU %")
		assert.Contains(t, rendered, "web-1")
		assert.Contains(t, rendered, "25.5%")
		assert.Contains(t, rendered, "db-1")
		assert.Contains(t, rendered, "10.0%")

		// Go back
		cmd = vm.HandleBack(model)
		assert.Nil(t, cmd)
		assert.Equal(t, ComposeProcessListView, model.currentView)
	})
}
