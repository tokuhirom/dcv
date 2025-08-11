package ui

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tokuhirom/dcv/internal/docker"
)

func TestFileContentViewModel_Rendering(t *testing.T) {
	tests := []struct {
		name      string
		viewModel FileContentViewModel
		model     *Model
		expected  []string
	}{
		{
			name: "displays file content",
			viewModel: FileContentViewModel{
				content:     "Line 1\nLine 2\nLine 3\nLine 4\nLine 5",
				contentPath: "/etc/config.conf",
				scrollY:     0,
			},
			model: &Model{
				width:  100,
				Height: 20,
				err:    nil,
			},
			expected: []string{"Line 1", "Line 2", "Line 3"},
		},
		{
			name: "displays error when present",
			viewModel: FileContentViewModel{
				content:     "",
				contentPath: "/etc/config.conf",
			},
			model: &Model{
				width:  100,
				Height: 20,
				err:    assert.AnError,
			},
			expected: []string{"Error:"},
		},
		{
			name: "handles scrolling",
			viewModel: FileContentViewModel{
				content:     "Line 1\nLine 2\nLine 3\nLine 4\nLine 5",
				contentPath: "/etc/config.conf",
				scrollY:     2,
			},
			model: &Model{
				width:  100,
				Height: 20,
				err:    nil,
			},
			expected: []string{"Line 3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.model.fileContentViewModel = tt.viewModel
			result := tt.viewModel.render(tt.model)

			for _, expected := range tt.expected {
				assert.Contains(t, result, expected, "Expected to find '%s' in output", expected)
			}
		})
	}
}

func TestFileContentViewModel_Navigation(t *testing.T) {
	t.Run("HandleDown scrolls down", func(t *testing.T) {
		vm := &FileContentViewModel{
			content: "Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6\nLine 7\nLine 8\nLine 9\nLine 10",
			scrollY: 0,
		}

		cmd := vm.HandleDown(10) // height = 10, visible lines = 5
		assert.Nil(t, cmd)
		assert.Equal(t, 1, vm.scrollY)

		// Test boundary
		vm.scrollY = 5 // Max scroll for 10 lines with height 10
		cmd = vm.HandleDown(10)
		assert.Nil(t, cmd)
		assert.Equal(t, 5, vm.scrollY, "Should not scroll beyond content")
	})

	t.Run("HandleUp scrolls up", func(t *testing.T) {
		vm := &FileContentViewModel{
			content: "Line 1\nLine 2\nLine 3\nLine 4\nLine 5",
			scrollY: 2,
		}

		cmd := vm.HandleUp()
		assert.Nil(t, cmd)
		assert.Equal(t, 1, vm.scrollY)

		// Test boundary
		vm.scrollY = 0
		cmd = vm.HandleUp()
		assert.Nil(t, cmd)
		assert.Equal(t, 0, vm.scrollY, "Should not scroll above 0")
	})

	t.Run("HandleGoToBeginning jumps to beginning", func(t *testing.T) {
		vm := &FileContentViewModel{
			content: "Line 1\nLine 2\nLine 3",
			scrollY: 2,
		}

		cmd := vm.HandleGoToBeginning()
		assert.Nil(t, cmd)
		assert.Equal(t, 0, vm.scrollY)
	})

	t.Run("HandleGoToEnd jumps to end", func(t *testing.T) {
		content := make([]string, 20)
		for i := range content {
			content[i] = "Line"
		}
		vm := &FileContentViewModel{
			content: strings.Join(content, "\n"),
			scrollY: 0,
		}

		cmd := vm.HandleGoToEnd(10) // height = 10, visible lines = 5
		assert.Nil(t, cmd)
		assert.Equal(t, 15, vm.scrollY) // 20 lines - 5 visible = 15

		// Test with short content
		vm.content = "Line 1\nLine 2"
		vm.scrollY = 0
		cmd = vm.HandleGoToEnd(10)
		assert.Nil(t, cmd)
		assert.Equal(t, 0, vm.scrollY, "Should not scroll if content fits")
	})
}

func TestFileContentViewModel_LoadContainer(t *testing.T) {
	t.Run("LoadContainer initializes file content view", func(t *testing.T) {
		model := &Model{
			dockerClient: docker.NewClient(),
			currentView:  FileBrowserView,
			loading:      false,
		}
		container := docker.NewContainer("container123", "test-container", "test-container", "running")
		vm := &FileContentViewModel{}

		cmd := vm.LoadContainer(model, container, "/etc/config.conf")
		assert.NotNil(t, cmd)
		assert.Equal(t, container, vm.container)
		assert.Equal(t, 0, vm.scrollY)
		assert.Equal(t, FileContentView, model.currentView)
		assert.True(t, model.loading)

		// Command should be created to load file content
		// We don't execute it in tests as it would require a real container
	})
}

func TestFileContentViewModel_HandleBack(t *testing.T) {
	t.Run("HandleBack returns to previous view", func(t *testing.T) {
		model := &Model{
			currentView: FileContentView,
			viewHistory: []ViewType{FileBrowserView, FileContentView},
		}
		container := docker.NewContainer("test123", "test-container", "test-container", "running")
		vm := &FileContentViewModel{
			content:     "some content",
			contentPath: "/file.txt",
			container:   container,
			scrollY:     5,
		}

		cmd := vm.HandleBack(model)
		assert.Nil(t, cmd)
		assert.Equal(t, FileBrowserView, model.currentView)
		assert.Empty(t, vm.content)
		assert.Empty(t, vm.contentPath)
		assert.Equal(t, 0, vm.scrollY)
	})
}

func TestFileContentViewModel_Loaded(t *testing.T) {
	t.Run("Loaded updates content and path", func(t *testing.T) {
		vm := &FileContentViewModel{
			scrollY: 10,
		}

		content := "New file content\nLine 2\nLine 3"
		path := "/new/path.txt"

		vm.Loaded(content, path)
		assert.Equal(t, content, vm.content)
		assert.Equal(t, path, vm.contentPath)
		assert.Equal(t, 0, vm.scrollY, "Should reset scroll position")
	})
}

func TestFileContentViewModel_Title(t *testing.T) {
	container := docker.NewContainer("test123", "test-container", "test-container", "running")
	vm := &FileContentViewModel{
		content:     "Line 1\nLine 2\nLine 3\nLine 4\nLine 5",
		contentPath: "/etc/config.conf",
		container:   container,
		scrollY:     2,
	}

	title := vm.Title()
	assert.Contains(t, title, "File:")
	assert.Contains(t, title, "[2/5]") // scrollY / total lines
	assert.Contains(t, title, "/etc/config.conf")
	assert.Contains(t, title, "[test-container]")
}

func TestFileContentViewModel_Integration(t *testing.T) {
	t.Run("Complete navigation flow from file browser", func(t *testing.T) {
		// Setup model with file browser and file content views
		model := &Model{
			dockerClient:         docker.NewClient(),
			currentView:          FileBrowserView,
			viewHistory:          []ViewType{ComposeProcessListView, FileBrowserView},
			loading:              false,
			width:                100,
			Height:               20,
			fileContentViewModel: FileContentViewModel{},
		}

		// Load file content from file browser
		container := docker.NewContainer("container123", "test-container", "test-container", "running")
		vm := &FileContentViewModel{}
		cmd := vm.LoadContainer(model, container, "/etc/config.conf")
		assert.NotNil(t, cmd)
		assert.Equal(t, FileContentView, model.currentView)

		// Simulate loading completion
		vm.Loaded("File content here", "/etc/config.conf")
		assert.Equal(t, "File content here", vm.content)

		// Navigate back should return to file browser
		cmd = vm.HandleBack(model)
		assert.Nil(t, cmd)
		assert.Equal(t, FileBrowserView, model.currentView)
	})

	t.Run("Navigation flow from other views", func(t *testing.T) {
		// Setup model starting from compose process list
		model := &Model{
			dockerClient:         docker.NewClient(),
			currentView:          ComposeProcessListView,
			viewHistory:          []ViewType{DockerContainerListView, ComposeProcessListView},
			loading:              false,
			fileContentViewModel: FileContentViewModel{},
		}

		// Setup model with FileContentView
		vm := &FileContentViewModel{}
		model.currentView = FileContentView
		model.viewHistory = []ViewType{ComposeProcessListView, FileContentView}

		// Navigate back should return to previous view (compose process list)
		cmd := vm.HandleBack(model)
		assert.Nil(t, cmd)
		assert.Equal(t, ComposeProcessListView, model.currentView)
	})
}

// Test helper functions
func TestFileContentViewModel_ScrollBoundaries(t *testing.T) {
	t.Run("Scroll boundaries with varying content sizes", func(t *testing.T) {
		testCases := []struct {
			name         string
			contentLines int
			height       int
			expectedMax  int
		}{
			{"Short content", 3, 10, 0},
			{"Exact fit", 5, 10, 0},
			{"Long content", 20, 10, 15},
			{"Very long content", 100, 10, 95},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				lines := make([]string, tc.contentLines)
				for i := range lines {
					lines[i] = "Line"
				}
				vm := &FileContentViewModel{
					content: strings.Join(lines, "\n"),
					scrollY: 0,
				}

				// Jump to end
				vm.HandleGoToEnd(tc.height)
				assert.Equal(t, tc.expectedMax, vm.scrollY)

				// Try to scroll beyond
				vm.HandleDown(tc.height)
				assert.Equal(t, tc.expectedMax, vm.scrollY, "Should not scroll beyond max")

				// Jump to start
				vm.HandleGoToBeginning()
				assert.Equal(t, 0, vm.scrollY)

				// Try to scroll above start
				vm.HandleUp()
				assert.Equal(t, 0, vm.scrollY, "Should not scroll above 0")
			})
		}
	})
}

func TestFileContentViewModel_EmptyContent(t *testing.T) {
	t.Run("Handles empty content gracefully", func(t *testing.T) {
		vm := &FileContentViewModel{
			content:     "",
			contentPath: "/empty/file",
			scrollY:     0,
		}

		// Should not crash on navigation
		vm.HandleDown(10)
		assert.Equal(t, 0, vm.scrollY)

		vm.HandleUp()
		assert.Equal(t, 0, vm.scrollY)

		vm.HandleGoToEnd(10)
		assert.Equal(t, 0, vm.scrollY)

		vm.HandleGoToBeginning()
		assert.Equal(t, 0, vm.scrollY)
	})
}

// Test message types are already defined in model.go
