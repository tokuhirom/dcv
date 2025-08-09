package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tokuhirom/dcv/internal/models"
)

func TestImageListViewModel_Rendering(t *testing.T) {
	tests := []struct {
		name      string
		viewModel ImageListViewModel
		height    int
		expected  []string
	}{
		{
			name: "displays no images message when empty",
			viewModel: ImageListViewModel{
				dockerImages:        []models.DockerImage{},
				selectedDockerImage: 0,
			},
			height:   20,
			expected: []string{"No images found"},
		},
		{
			name: "displays image list table",
			viewModel: ImageListViewModel{
				dockerImages: []models.DockerImage{
					{
						Repository:   "nginx",
						Tag:          "latest",
						ID:           "sha256:1234567890abcdef",
						CreatedSince: "24 hours ago",
						Size:         "128MB",
					},
					{
						Repository:   "postgres",
						Tag:          "15-alpine",
						ID:           "sha256:abcdef1234567890",
						CreatedSince: "2 days ago",
						Size:         "256MB",
					},
				},
				selectedDockerImage: 0,
			},
			height: 20,
			expected: []string{
				"REPOSITORY",
				"TAG",
				"IMAGE ID",
				"CREATED",
				"SIZE",
				"nginx",
				"latest",
				"postgres",
				"15-alpine",
			},
		},
		{
			name: "truncates long repository names",
			viewModel: ImageListViewModel{
				dockerImages: []models.DockerImage{
					{
						Repository:   "very-long-repository-name-that-should-be-truncated-in-the-display",
						Tag:          "latest",
						ID:           "sha256:1234567890abcdef",
						CreatedSince: "1 hour ago",
						Size:         "1KB",
					},
				},
				selectedDockerImage: 0,
			},
			height:   20,
			expected: []string{"very-long-reposit..."},
		},
		{
			name: "shows <none> for empty repository",
			viewModel: ImageListViewModel{
				dockerImages: []models.DockerImage{
					{
						Repository:   "<none>",
						Tag:          "<none>",
						ID:           "sha256:1234567890abcdef",
						CreatedSince: "1 hour ago",
						Size:         "1KB",
					},
				},
				selectedDockerImage: 0,
			},
			height:   20,
			expected: []string{"<none>"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &Model{
				imageListViewModel: tt.viewModel,
				width:              100,
				Height:             tt.height,
			}

			result := tt.viewModel.render(model, tt.height-4)

			for _, expected := range tt.expected {
				assert.Contains(t, result, expected, "Expected to find '%s' in output", expected)
			}
		})
	}
}

func TestImageListViewModel_Navigation(t *testing.T) {
	t.Run("HandleDown moves selection down", func(t *testing.T) {
		vm := &ImageListViewModel{
			dockerImages: []models.DockerImage{
				{Repository: "image1", Tag: "latest"},
				{Repository: "image2", Tag: "latest"},
				{Repository: "image3", Tag: "latest"},
			},
			selectedDockerImage: 0,
		}

		cmd := vm.HandleDown()
		assert.Nil(t, cmd)
		assert.Equal(t, 1, vm.selectedDockerImage)

		// Test boundary
		vm.selectedDockerImage = 2
		cmd = vm.HandleDown()
		assert.Nil(t, cmd)
		assert.Equal(t, 2, vm.selectedDockerImage, "Should not go beyond last image")
	})

	t.Run("HandleUp moves selection up", func(t *testing.T) {
		vm := &ImageListViewModel{
			dockerImages: []models.DockerImage{
				{Repository: "image1", Tag: "latest"},
				{Repository: "image2", Tag: "latest"},
				{Repository: "image3", Tag: "latest"},
			},
			selectedDockerImage: 2,
		}

		cmd := vm.HandleUp()
		assert.Nil(t, cmd)
		assert.Equal(t, 1, vm.selectedDockerImage)

		// Test boundary
		vm.selectedDockerImage = 0
		cmd = vm.HandleUp()
		assert.Nil(t, cmd)
		assert.Equal(t, 0, vm.selectedDockerImage, "Should not go below 0")
	})
}

func TestImageListViewModel_Operations(t *testing.T) {
	t.Run("Show switches to image list view", func(t *testing.T) {
		model := &Model{
			currentView: ComposeProcessListView,
			loading:     false,
		}
		vm := &ImageListViewModel{}

		cmd := vm.Show(model)

		assert.Equal(t, ImageListView, model.currentView)
		assert.True(t, model.loading)
		assert.NotNil(t, cmd)
	})

	t.Run("HandleToggleAll toggles showAll flag", func(t *testing.T) {
		model := &Model{loading: false}
		vm := &ImageListViewModel{showAll: false}

		cmd := vm.HandleToggleAll(model)

		assert.True(t, vm.showAll)
		assert.True(t, model.loading)
		assert.NotNil(t, cmd)

		// Toggle back
		model.loading = false
		_ = vm.HandleToggleAll(model)
		assert.False(t, vm.showAll)
		assert.True(t, model.loading)
	})

	t.Run("HandleDelete removes selected image", func(t *testing.T) {
		model := &Model{
			loading:                   false,
			commandExecutionViewModel: CommandExecutionViewModel{},
			currentView:               ImageListView,
		}
		vm := &ImageListViewModel{
			dockerImages: []models.DockerImage{
				{ID: "image1", Repository: "test", Tag: "latest"},
				{ID: "image2", Repository: "test2", Tag: "v1"},
			},
			selectedDockerImage: 1,
		}

		cmd := vm.HandleDelete(model)

		// HandleDelete now shows confirmation dialog, so returns nil
		assert.Nil(t, cmd)
		// Should switch to CommandExecutionView
		assert.Equal(t, CommandExecutionView, model.currentView)
		// Should set pending confirmation
		assert.True(t, model.commandExecutionViewModel.pendingConfirmation)
		// Should have the correct pending args
		assert.Equal(t, []string{"rmi", "test2:v1"}, model.commandExecutionViewModel.pendingArgs)
	})

	t.Run("HandleDelete does nothing when no images", func(t *testing.T) {
		model := &Model{loading: false}
		vm := &ImageListViewModel{
			dockerImages:        []models.DockerImage{},
			selectedDockerImage: 0,
		}

		cmd := vm.HandleDelete(model)

		assert.False(t, model.loading)
		assert.Nil(t, cmd)
	})

	t.Run("HandleInspect shows image inspection", func(t *testing.T) {
		model := &Model{
			currentView: ImageListView,
			loading:     false,
		}
		vm := &ImageListViewModel{
			dockerImages: []models.DockerImage{
				{ID: "image1"},
			},
			selectedDockerImage: 0,
		}

		cmd := vm.HandleInspect(model)

		// The view change happens when inspectLoadedMsg is processed, not immediately
		assert.Equal(t, ImageListView, model.currentView)
		assert.Equal(t, "image1", model.inspectViewModel.inspectImageID)
		assert.True(t, model.loading)
		assert.NotNil(t, cmd)
	})

	t.Run("HandleBack returns to previous view", func(t *testing.T) {
		model := &Model{
			currentView: ImageListView,
			viewHistory: []ViewType{ComposeProcessListView},
		}
		vm := &ImageListViewModel{}

		cmd := vm.HandleBack(model)

		assert.Equal(t, ComposeProcessListView, model.currentView)
		assert.Nil(t, cmd) // HandleBack now returns nil
	})
}

func TestImageListViewModel_Messages(t *testing.T) {
	t.Run("Loaded updates image list", func(t *testing.T) {
		vm := &ImageListViewModel{
			selectedDockerImage: 5, // Out of bounds
		}

		images := []models.DockerImage{
			{Repository: "nginx", Tag: "latest"},
			{Repository: "postgres", Tag: "15"},
		}

		vm.Loaded(images)

		assert.Equal(t, images, vm.dockerImages)
		assert.Equal(t, 0, vm.selectedDockerImage, "Should reset selection when out of bounds")
	})

	t.Run("Loaded keeps selection in bounds", func(t *testing.T) {
		vm := &ImageListViewModel{
			selectedDockerImage: 1,
		}

		images := []models.DockerImage{
			{Repository: "nginx", Tag: "latest"},
			{Repository: "postgres", Tag: "15"},
			{Repository: "redis", Tag: "alpine"},
		}

		vm.Loaded(images)

		assert.Equal(t, 1, vm.selectedDockerImage, "Should keep selection when in bounds")
	})
}

func TestImageListViewModel_HandleRefresh(t *testing.T) {
	model := &Model{
		loading: false,
	}
	vm := &ImageListViewModel{
		showAll: true,
	}

	cmd := vm.HandleRefresh(model)

	assert.True(t, model.loading)
	assert.NotNil(t, cmd)

	// Execute the command to verify it's loadDockerImages
	msg := cmd()
	_, ok := msg.(dockerImagesLoadedMsg)
	assert.True(t, ok, "Should return dockerImagesLoadedMsg")
}

func TestImageListViewModel_EmptySelection(t *testing.T) {
	t.Run("operations handle empty image list gracefully", func(t *testing.T) {
		model := &Model{}
		vm := &ImageListViewModel{
			dockerImages:        []models.DockerImage{},
			selectedDockerImage: 0,
		}

		// Test all operations that depend on selection
		assert.Nil(t, vm.HandleDelete(model))
		assert.Nil(t, vm.HandleInspect(model))

		// Navigation should not crash
		assert.Nil(t, vm.HandleUp())
		assert.Nil(t, vm.HandleDown())
	})
}

func TestImageListViewModel_KeyHandlers(t *testing.T) {
	model := NewModel(ImageListView)
	model.initializeKeyHandlers()

	// Verify key handlers are registered
	handlers := model.imageListViewHandlers
	assert.Greater(t, len(handlers), 0, "Should have registered key handlers")

	// Check specific handlers exist
	expectedKeys := []string{"up", "down", "a", "D", "i", "r"} // 'q' is now a global handler
	registeredKeys := make(map[string]bool)

	for _, h := range handlers {
		for _, key := range h.Keys {
			registeredKeys[key] = true
		}
	}

	for _, key := range expectedKeys {
		assert.True(t, registeredKeys[key], "Key %s should be registered", key)
	}
}

func TestImageListViewModel_Update(t *testing.T) {
	t.Run("handles dockerImagesLoadedMsg success", func(t *testing.T) {
		model := &Model{
			currentView: ImageListView,
			loading:     true,
		}
		vm := &ImageListViewModel{}
		model.imageListViewModel = *vm

		images := []models.DockerImage{
			{Repository: "nginx", Tag: "latest"},
		}

		msg := dockerImagesLoadedMsg{
			images: images,
			err:    nil,
		}

		newModel, cmd := model.Update(msg)
		m := newModel.(*Model)

		assert.False(t, m.loading)
		assert.Nil(t, m.err)
		assert.Equal(t, images, m.imageListViewModel.dockerImages)
		assert.Nil(t, cmd)
	})

	t.Run("handles dockerImagesLoadedMsg error", func(t *testing.T) {
		model := &Model{
			currentView: ImageListView,
			loading:     true,
		}

		testErr := assert.AnError
		msg := dockerImagesLoadedMsg{
			images: nil,
			err:    testErr,
		}

		newModel, cmd := model.Update(msg)
		m := newModel.(*Model)

		assert.False(t, m.loading)
		assert.Equal(t, testErr, m.err)
		assert.Nil(t, cmd)
	})
}

// Test that image formatting displays correctly
func TestImageListViewModel_ImageFormatting(t *testing.T) {
	t.Run("formats image with repository and tag", func(t *testing.T) {
		image := models.DockerImage{
			Repository: "nginx",
			Tag:        "latest",
		}
		assert.Equal(t, "nginx:latest", image.GetRepoTag())
	})

	t.Run("formats image with <none> repository", func(t *testing.T) {
		image := models.DockerImage{
			Repository: "<none>",
			Tag:        "<none>",
			ID:         "sha256:abcd",
		}
		assert.Equal(t, "sha256:abcd", image.GetRepoTag())
	})
}
