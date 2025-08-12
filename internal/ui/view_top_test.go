package ui

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
)

func TestTopViewModel_Rendering(t *testing.T) {
	tests := []struct {
		name        string
		viewModel   TopViewModel
		height      int
		expected    []string
		notExpected []string
	}{
		{
			name: "displays no process info message when empty",
			viewModel: TopViewModel{
				processes: nil,
			},
			height:   20,
			expected: []string{"No process information available"},
		},
		{
			name: "displays process information with stats",
			viewModel: TopViewModel{
				processes: []models.Process{
					{UID: "root", PID: "1", PPID: "0", CPUPerc: 5.5, MemPerc: 2.3, STIME: "10:00", TIME: "00:01:30", CMD: "/bin/sh"},
					{UID: "www", PID: "42", PPID: "1", CPUPerc: 25.0, MemPerc: 10.5, STIME: "10:01", TIME: "00:05:00", CMD: "nginx"},
				},
				containerStats: &models.ContainerStats{
					CPUPerc:  "30.5%",
					MemUsage: "512MiB / 4GiB",
					MemPerc:  "12.8%",
					PIDs:     "10",
				},
				sortField:   SortByCPU,
				sortReverse: true,
			},
			height: 20,
			expected: []string{
				"CPU", "Memory", "PIDs", // Stats header
				"Sort: CPU%", "(desc)", // Sort info
				"PID", "CPU%", "MEM%", "COMMAND", // Column headers
				"root", "/bin/sh",
				"nginx",
			},
		},
		{
			name: "displays sorting shortcuts",
			viewModel: TopViewModel{
				processes: []models.Process{
					{PID: "1", CMD: "test"},
				},
			},
			height: 20,
			expected: []string{
				"[c]PU [m]EM [p]ID [t]IME [n]ame [r]everse",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &Model{width: 100}
			result := tt.viewModel.render(model, tt.height)

			for _, expected := range tt.expected {
				assert.Contains(t, result, expected, "Expected to find '%s' in output", expected)
			}

			for _, notExpected := range tt.notExpected {
				assert.NotContains(t, result, notExpected, "Should not find '%s' in output", notExpected)
			}
		})
	}
}

func TestTopViewModel_Load(t *testing.T) {
	t.Run("Load switches to top view and initiates loading", func(t *testing.T) {
		model := &Model{
			dockerClient: docker.NewClient(),
			currentView:  ComposeProcessListView,
			loading:      false,
		}
		vm := &TopViewModel{}
		container := docker.NewContainer("test-container", "web-1", "web-1 (test-project)", "running")

		cmd := vm.Load(model, container)
		assert.NotNil(t, cmd)
		assert.Equal(t, container, vm.container)
		assert.Equal(t, TopView, model.currentView)
		assert.True(t, model.loading)
	})
}

func TestTopViewModel_DoLoad(t *testing.T) {
	t.Run("DoLoad returns command to load process info", func(t *testing.T) {
		model := &Model{
			dockerClient: docker.NewClient(),
			loading:      false,
		}
		vm := &TopViewModel{
			container: docker.NewContainer("test-container", "web-1", "web-1 (test-project)", "running"),
		}

		cmd := vm.DoLoad(model)
		assert.NotNil(t, cmd)
		assert.True(t, model.loading)
	})
}

func TestTopViewModel_HandleBack(t *testing.T) {
	t.Run("HandleBack returns to previous view", func(t *testing.T) {
		model := &Model{
			currentView: TopView,
			viewHistory: []ViewType{ComposeProcessListView, TopView},
		}
		vm := &TopViewModel{}

		cmd := vm.HandleBack(model)
		assert.Nil(t, cmd)
		assert.Equal(t, ComposeProcessListView, model.currentView)
	})
}

func TestTopViewModel_Loaded(t *testing.T) {
	t.Run("Loaded updates processes and stats", func(t *testing.T) {
		vm := &TopViewModel{
			scrollY: 5,
		}

		processes := []models.Process{
			{PID: "100", CMD: "process1"},
			{PID: "200", CMD: "process2"},
		}
		stats := &models.ContainerStats{
			Name:    "test-container",
			CPUPerc: "25.5%",
			MemPerc: "10.2%",
		}

		vm.Loaded(processes, stats)
		assert.Equal(t, processes, vm.processes)
		assert.Equal(t, stats, vm.containerStats)
		assert.Equal(t, 0, vm.scrollY) // Should reset scroll position
	})
}

func TestTopViewModel_Title(t *testing.T) {
	tests := []struct {
		name      string
		container *docker.Container
		expected  string
	}{
		{
			name:      "compose container title",
			container: docker.NewContainer("abc123", "web-1", "web-1 (myproject)", "running"),
			expected:  "Process Info: web-1 (myproject)",
		},
		{
			name:      "docker container title",
			container: docker.NewContainer("def456", "nginx-server", "nginx-server", "running"),
			expected:  "Process Info: nginx-server",
		},
		{
			name:      "dind container title",
			container: docker.NewDindContainer("host-1", "host-container", "inner-1", "inner-container", "running"),
			expected:  "Process Info: DinD: inner-container (host-container)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := &TopViewModel{
				container: tt.container,
			}
			result := vm.Title()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTopViewModel_ParseProcesses(t *testing.T) {
	vm := &TopViewModel{}

	t.Run("parses docker top output correctly", func(t *testing.T) {
		output := `UID                 PID                 PPID                C                   STIME               TTY                 TIME                CMD
root                1234                1000                5                   10:00               ?                   00:01:30            /usr/bin/process1
www-data            5678                1234                10                  10:05               pts/0               00:00:45            nginx: worker process
999                 9012                1234                2                   10:10               ?                   00:02:15            redis-server *:6379`

		processes := vm.parseProcesses(output)

		assert.Len(t, processes, 3)

		// Check first process
		assert.Equal(t, "root", processes[0].UID)
		assert.Equal(t, "1234", processes[0].PID)
		assert.Equal(t, "1000", processes[0].PPID)
		assert.Equal(t, "5", processes[0].C)
		assert.Equal(t, "10:00", processes[0].STIME)
		assert.Equal(t, "?", processes[0].TTY)
		assert.Equal(t, "00:01:30", processes[0].TIME)
		assert.Equal(t, "/usr/bin/process1", processes[0].CMD)
	})

	t.Run("handles empty output", func(t *testing.T) {
		processes := vm.parseProcesses("")
		assert.Nil(t, processes)
	})
}

func TestTopViewModel_Sorting(t *testing.T) {
	vm := &TopViewModel{
		processes: []models.Process{
			{PID: "100", CMD: "process1", CPUPerc: 50.0, MemPerc: 20.0, TIME: "00:10:00"},
			{PID: "200", CMD: "process2", CPUPerc: 30.0, MemPerc: 40.0, TIME: "00:05:00"},
			{PID: "150", CMD: "process3", CPUPerc: 70.0, MemPerc: 10.0, TIME: "00:15:00"},
		},
	}

	t.Run("sort by CPU descending", func(t *testing.T) {
		vm.sortField = SortByCPU
		vm.sortReverse = true
		vm.sortProcesses()

		assert.Equal(t, 70.0, vm.processes[0].CPUPerc)
		assert.Equal(t, 50.0, vm.processes[1].CPUPerc)
		assert.Equal(t, 30.0, vm.processes[2].CPUPerc)
	})

	t.Run("sort by memory descending", func(t *testing.T) {
		vm.sortField = SortByMem
		vm.sortReverse = true
		vm.sortProcesses()

		assert.Equal(t, 40.0, vm.processes[0].MemPerc)
		assert.Equal(t, 20.0, vm.processes[1].MemPerc)
		assert.Equal(t, 10.0, vm.processes[2].MemPerc)
	})
}

func TestTopViewModel_SortHandlers(t *testing.T) {
	vm := &TopViewModel{}

	t.Run("HandleSortByCPU toggles correctly", func(t *testing.T) {
		// First call sets CPU sort with reverse=true
		vm.HandleSortByCPU()
		assert.Equal(t, SortByCPU, vm.sortField)
		assert.True(t, vm.sortReverse)

		// Second call toggles reverse
		vm.HandleSortByCPU()
		assert.Equal(t, SortByCPU, vm.sortField)
		assert.False(t, vm.sortReverse)
	})

	t.Run("HandleSortByMem defaults to descending", func(t *testing.T) {
		vm.sortField = SortByPID
		vm.HandleSortByMem()
		assert.Equal(t, SortByMem, vm.sortField)
		assert.True(t, vm.sortReverse)
	})

	t.Run("HandleReverseSort toggles order", func(t *testing.T) {
		vm.sortReverse = false
		vm.HandleReverseSort()
		assert.True(t, vm.sortReverse)

		vm.HandleReverseSort()
		assert.False(t, vm.sortReverse)
	})
}

func TestTopViewModel_Navigation(t *testing.T) {
	vm := &TopViewModel{
		processes: []models.Process{
			{PID: "100"},
			{PID: "200"},
			{PID: "300"},
		},
		scrollY: 1,
	}

	t.Run("HandleUp scrolls up", func(t *testing.T) {
		vm.HandleUp()
		assert.Equal(t, 0, vm.scrollY)

		// Shouldn't go below 0
		vm.HandleUp()
		assert.Equal(t, 0, vm.scrollY)
	})

	t.Run("HandleDown scrolls down", func(t *testing.T) {
		vm.scrollY = 0
		vm.HandleDown()
		assert.Equal(t, 1, vm.scrollY)

		vm.HandleDown()
		assert.Equal(t, 2, vm.scrollY)

		// Shouldn't go beyond last process
		vm.HandleDown()
		assert.Equal(t, 2, vm.scrollY)
	})
}

func TestTopViewModel_AutoRefresh(t *testing.T) {
	vm := &TopViewModel{}

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

	t.Run("Load enables auto-refresh by default", func(t *testing.T) {
		model := &Model{
			dockerClient: docker.NewClient(),
			currentView:  ComposeProcessListView,
		}
		vm := &TopViewModel{}
		container := docker.NewContainer("test-container", "web-1", "web-1 (test-project)", "running")

		cmd := vm.Load(model, container)
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
		vm := &TopViewModel{
			container: docker.NewContainer("test-container", "web-1", "web-1 (test-project)", "running"),
		}

		cmd := vm.DoLoadSilent(model)
		assert.NotNil(t, cmd)
		assert.False(t, model.loading) // Should remain false
	})
}
