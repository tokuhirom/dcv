package ui

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tokuhirom/dcv/internal/docker"
)

func TestTopViewModel_Rendering(t *testing.T) {
	tests := []struct {
		name      string
		viewModel TopViewModel
		height    int
		expected  []string
	}{
		{
			name: "displays no process info message when empty",
			viewModel: TopViewModel{
				content: "",
			},
			height:   20,
			expected: []string{"No process information available"},
		},
		{
			name: "displays process information",
			viewModel: TopViewModel{
				content: `USER       PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND
root         1  0.0  0.1   4188  3380 ?        Ss   10:00   0:00 /bin/sh
root        42  0.5  1.2  45678 12345 ?        S    10:01   0:15 node app.js
www-data   123  0.1  0.5  23456  5678 ?        S    10:02   0:05 nginx`,
			},
			height: 20,
			expected: []string{
				"USER",
				"PID",
				"%CPU",
				"COMMAND",
				"root",
				"/bin/sh",
				"node app.js",
				"nginx",
			},
		},
		{
			name: "truncates content when too tall",
			viewModel: TopViewModel{
				content: strings.Repeat("Line\n", 50),
			},
			height:   10, // Only 8 lines visible (height - 2)
			expected: []string{"Line"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.viewModel.render(tt.height)

			for _, expected := range tt.expected {
				assert.Contains(t, result, expected, "Expected to find '%s' in output", expected)
			}

			// Check truncation
			if tt.name == "truncates content when too tall" {
				lines := strings.Split(strings.TrimSpace(result), "\n")
				assert.LessOrEqual(t, len(lines), tt.height-2, "Should not exceed visible height")
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
		container := docker.NewContainer(
			docker.NewClient(), "test-container", "web-1", "web-1 (test-project)", "running",
		)

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
			container: docker.NewContainer(
				docker.NewClient(), "test-container", "web-1", "web-1 (test-project)", "running",
			),
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
	t.Run("Loaded updates content", func(t *testing.T) {
		vm := &TopViewModel{
			content: "",
		}

		output := `USER       PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND
root         1  0.0  0.1   4188  3380 ?        Ss   10:00   0:00 /bin/sh`

		vm.Loaded(output)
		assert.Equal(t, output, vm.content)
	})

	t.Run("Loaded replaces existing content", func(t *testing.T) {
		vm := &TopViewModel{
			content: "old content",
		}

		newOutput := "new process info"
		vm.Loaded(newOutput)
		assert.Equal(t, newOutput, vm.content)
	})
}

func TestTopViewModel_Title(t *testing.T) {
	tests := []struct {
		name      string
		container *docker.Container
		expected  string
	}{
		{
			name: "compose container title",
			container: docker.NewContainer(
				docker.NewClient(), "abc123", "web-1", "web-1 (myproject)", "running",
			),
			expected: "Process Info: web-1 (myproject)",
		},
		{
			name: "docker container title",
			container: docker.NewContainer(
				docker.NewClient(), "def456", "nginx-server", "nginx-server", "running",
			),
			expected: "Process Info: nginx-server",
		},
		{
			name: "dind container title",
			container: docker.NewDindContainer(
				docker.NewClient(), "host-1", "host-container",
				"inner-1", "inner-container", "running",
			),
			expected: "Process Info: DinD: host-container (inner-container)",
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

func TestTopViewModel_Integration(t *testing.T) {
	t.Run("Complete flow from load to display", func(t *testing.T) {
		// Setup
		model := &Model{
			dockerClient: docker.NewClient(),
			currentView:  ComposeProcessListView,
			loading:      false,
			Height:       20,
		}
		vm := &TopViewModel{}
		container := docker.NewContainer(
			docker.NewClient(), "test-container", "web-1", "web-1 (test-project)", "running",
		)

		// Load
		cmd := vm.Load(model, container)
		assert.NotNil(t, cmd)
		assert.Equal(t, TopView, model.currentView)

		// Simulate loading completion
		output := `USER       PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND
root         1  0.0  0.1   4188  3380 ?        Ss   10:00   0:00 /bin/sh`
		vm.Loaded(output)

		// Render
		rendered := vm.render(model.Height)
		assert.Contains(t, rendered, "USER")
		assert.Contains(t, rendered, "root")
		assert.Contains(t, rendered, "/bin/sh")

		// Check title
		title := vm.Title()
		assert.Equal(t, "Process Info: web-1 (test-project)", title)

		// Go back
		cmd = vm.HandleBack(model)
		assert.Nil(t, cmd)
		assert.Equal(t, ComposeProcessListView, model.currentView)
	})
}
