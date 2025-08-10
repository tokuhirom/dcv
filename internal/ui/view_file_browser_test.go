package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
)

func TestFileBrowserViewModel_HistoryManagement(t *testing.T) {
	t.Run("pushHistory adds path to history", func(t *testing.T) {
		vm := &FileBrowserViewModel{
			pathHistory: []string{"/"},
			currentPath: "/",
		}

		vm.pushHistory("/usr")
		assert.Equal(t, []string{"/", "/usr"}, vm.pathHistory)
		assert.Equal(t, "/usr", vm.currentPath)

		vm.pushHistory("/usr/local")
		assert.Equal(t, []string{"/", "/usr", "/usr/local"}, vm.pathHistory)
		assert.Equal(t, "/usr/local", vm.currentPath)
	})

	t.Run("popHistory removes last path and returns to previous", func(t *testing.T) {
		vm := &FileBrowserViewModel{
			pathHistory: []string{"/", "/usr", "/usr/local"},
			currentPath: "/usr/local",
		}

		result := vm.popHistory()
		assert.True(t, result)
		assert.Equal(t, []string{"/", "/usr"}, vm.pathHistory)
		assert.Equal(t, "/usr", vm.currentPath)

		result = vm.popHistory()
		assert.True(t, result)
		assert.Equal(t, []string{"/"}, vm.pathHistory)
		assert.Equal(t, "/", vm.currentPath)

		// Can't pop when only one item left
		result = vm.popHistory()
		assert.False(t, result)
		assert.Equal(t, []string{"/"}, vm.pathHistory)
		assert.Equal(t, "/", vm.currentPath)
	})

	t.Run("popHistory returns false when empty", func(t *testing.T) {
		vm := &FileBrowserViewModel{
			pathHistory: []string{},
			currentPath: "",
		}

		result := vm.popHistory()
		assert.False(t, result)
		assert.Equal(t, []string{}, vm.pathHistory)
	})
}

func TestFileBrowserViewModel_Rendering(t *testing.T) {
	tests := []struct {
		name      string
		viewModel FileBrowserViewModel
		model     *Model
		height    int
		expected  []string
	}{
		{
			name: "displays no files message when empty",
			viewModel: FileBrowserViewModel{
				containerFiles: []models.ContainerFile{},
				selectedFile:   0,
				currentPath:    "/",
			},
			model: &Model{
				width:  100,
				Height: 20,
			},
			height:   20,
			expected: []string{"No files found or directory is empty"},
		},
		{
			name: "displays file list table",
			viewModel: FileBrowserViewModel{
				containerFiles: []models.ContainerFile{
					{
						Name:        ".",
						Permissions: "drwxr-xr-x",
						IsDir:       true,
					},
					{
						Name:        "..",
						Permissions: "drwxr-xr-x",
						IsDir:       true,
					},
					{
						Name:        "app",
						Permissions: "drwxr-xr-x",
						IsDir:       true,
					},
					{
						Name:        "config.json",
						Permissions: "-rw-r--r--",
						IsDir:       false,
					},
				},
				selectedFile: 0,
				currentPath:  "/usr/local",
			},
			model: &Model{
				width:  100,
				Height: 20,
			},
			height: 20,
			expected: []string{
				"PERMISSIONS",
				"NAME",
				"drwxr-xr-x",
				"app/",
				"config.json",
			},
		},
		{
			name: "highlights directories with color",
			viewModel: FileBrowserViewModel{
				containerFiles: []models.ContainerFile{
					{
						Name:        "bin",
						Permissions: "drwxr-xr-x",
						IsDir:       true,
					},
					{
						Name:        "file.txt",
						Permissions: "-rw-r--r--",
						IsDir:       false,
					},
				},
				selectedFile: 0,
				currentPath:  "/",
			},
			model: &Model{
				width:  100,
				Height: 20,
			},
			height:   20,
			expected: []string{"bin/", "file.txt"},
		},
		{
			name: "shows symlinks with arrow",
			viewModel: FileBrowserViewModel{
				containerFiles: []models.ContainerFile{
					{
						Name:        "link",
						Permissions: "lrwxrwxrwx",
						IsDir:       false,
						LinkTarget:  "/target/path",
					},
				},
				selectedFile: 0,
				currentPath:  "/",
			},
			model: &Model{
				width:  100,
				Height: 20,
			},
			height:   20,
			expected: []string{"link -> /target/path"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.model.fileBrowserViewModel = tt.viewModel
			result := tt.viewModel.render(tt.model, tt.height-4)

			for _, expected := range tt.expected {
				assert.Contains(t, result, expected, "Expected to find '%s' in output", expected)
			}
		})
	}
}

func TestFileBrowserViewModel_Navigation(t *testing.T) {
	t.Run("HandleDown moves selection down", func(t *testing.T) {
		vm := &FileBrowserViewModel{
			containerFiles: []models.ContainerFile{
				{Name: "file1.txt"},
				{Name: "file2.txt"},
				{Name: "file3.txt"},
			},
			selectedFile: 0,
		}

		cmd := vm.HandleDown()
		assert.Nil(t, cmd)
		assert.Equal(t, 1, vm.selectedFile)

		// Test boundary
		vm.selectedFile = 2
		cmd = vm.HandleDown()
		assert.Nil(t, cmd)
		assert.Equal(t, 2, vm.selectedFile, "Should not go beyond last file")
	})

	t.Run("HandleUp moves selection up", func(t *testing.T) {
		vm := &FileBrowserViewModel{
			containerFiles: []models.ContainerFile{
				{Name: "file1.txt"},
				{Name: "file2.txt"},
				{Name: "file3.txt"},
			},
			selectedFile: 2,
		}

		cmd := vm.HandleUp()
		assert.Nil(t, cmd)
		assert.Equal(t, 1, vm.selectedFile)

		// Test boundary
		vm.selectedFile = 0
		cmd = vm.HandleUp()
		assert.Nil(t, cmd)
		assert.Equal(t, 0, vm.selectedFile, "Should not go below 0")
	})
}

func TestFileBrowserViewModel_DirectoryNavigation(t *testing.T) {
	t.Run("HandleGoToParentDirectory goes up one level", func(t *testing.T) {
		model := &Model{
			dockerClient: docker.NewClient(),
			loading:      false,
		}
		vm := &FileBrowserViewModel{
			currentPath: "/usr/local/bin",
			pathHistory: []string{"/", "/usr", "/usr/local", "/usr/local/bin"},
		}

		cmd := vm.HandleGoToParentDirectory(model)
		assert.NotNil(t, cmd)
		assert.Equal(t, "/usr/local", vm.currentPath)
		// pushHistory appends the parent path
		assert.Equal(t, []string{"/", "/usr", "/usr/local", "/usr/local/bin", "/usr/local"}, vm.pathHistory)
		assert.Equal(t, 0, vm.selectedFile)
	})

	t.Run("HandleGoToParentDirectory does nothing at root", func(t *testing.T) {
		model := &Model{
			dockerClient: docker.NewClient(),
			loading:      false,
		}
		vm := &FileBrowserViewModel{
			currentPath: "/",
			pathHistory: []string{"/"},
		}

		cmd := vm.HandleGoToParentDirectory(model)
		assert.Nil(t, cmd)
		assert.Equal(t, "/", vm.currentPath)
		assert.Equal(t, []string{"/"}, vm.pathHistory)
		assert.False(t, model.loading)
	})
}

func TestFileBrowserViewModel_FileOperations(t *testing.T) {
	t.Run("HandleOpenFileOrDirectory navigates into directory", func(t *testing.T) {
		model := &Model{
			dockerClient: docker.NewClient(),
			loading:      false,
		}
		vm := &FileBrowserViewModel{
			containerFiles: []models.ContainerFile{
				{Name: "subdir", IsDir: true},
			},
			selectedFile: 0,
			currentPath:  "/usr",
			pathHistory:  []string{"/", "/usr"},
		}

		cmd := vm.HandleOpenFileOrDirectory(model)
		assert.NotNil(t, cmd)
		assert.Equal(t, "/usr/subdir", vm.currentPath)
		assert.Equal(t, []string{"/", "/usr", "/usr/subdir"}, vm.pathHistory)
		assert.Equal(t, 0, vm.selectedFile)
	})

	t.Run("HandleOpenFileOrDirectory opens file content", func(t *testing.T) {
		model := &Model{
			dockerClient:         docker.NewClient(),
			loading:              false,
			fileContentViewModel: FileContentViewModel{},
		}
		vm := &FileBrowserViewModel{
			containerFiles: []models.ContainerFile{
				{Name: "file.txt", IsDir: false},
			},
			selectedFile:      0,
			currentPath:       "/usr",
			browsingContainer: docker.NewContainer(model.dockerClient, "container123", "test-container", "test-container", "running"),
		}

		cmd := vm.HandleOpenFileOrDirectory(model)
		assert.NotNil(t, cmd)
		// File path should be constructed and passed to file content view
		assert.Equal(t, "/usr", vm.currentPath) // Path should not change
	})

	t.Run("HandleOpenFileOrDirectory handles .. directory", func(t *testing.T) {
		model := &Model{
			dockerClient: docker.NewClient(),
			loading:      false,
		}
		vm := &FileBrowserViewModel{
			containerFiles: []models.ContainerFile{
				{Name: "..", IsDir: true},
			},
			selectedFile: 0,
			currentPath:  "/usr/local",
			pathHistory:  []string{"/", "/usr", "/usr/local"},
		}

		cmd := vm.HandleOpenFileOrDirectory(model)
		assert.NotNil(t, cmd)
		assert.Equal(t, "/usr", vm.currentPath)
		assert.Equal(t, []string{"/", "/usr", "/usr/local", "/usr"}, vm.pathHistory)
		assert.Equal(t, 0, vm.selectedFile)
	})

	t.Run("HandleOpenFileOrDirectory ignores . directory", func(t *testing.T) {
		model := &Model{
			dockerClient: docker.NewClient(),
			loading:      false,
		}
		vm := &FileBrowserViewModel{
			containerFiles: []models.ContainerFile{
				{Name: ".", IsDir: true},
			},
			selectedFile: 0,
			currentPath:  "/usr",
		}

		cmd := vm.HandleOpenFileOrDirectory(model)
		assert.Nil(t, cmd)
		assert.Equal(t, "/usr", vm.currentPath) // Path should not change
	})
}

func TestFileBrowserViewModel_LoadContainer(t *testing.T) {
	t.Run("LoadContainer initializes file browser", func(t *testing.T) {
		model := &Model{
			dockerClient: docker.NewClient(),
			currentView:  ComposeProcessListView,
			loading:      false,
		}
		vm := &FileBrowserViewModel{}
		container := docker.NewContainer(model.dockerClient, "container123", "test-container", "test-container", "running")

		cmd := vm.LoadContainer(model, container)
		assert.NotNil(t, cmd)
		assert.Equal(t, container, vm.browsingContainer)
		assert.Equal(t, "/", vm.currentPath)
		assert.Equal(t, []string{"/"}, vm.pathHistory) // pushHistory adds the initial path
		assert.Equal(t, FileBrowserView, model.currentView)
	})
}

func TestFileBrowserViewModel_HandleBack(t *testing.T) {
	t.Run("HandleBack navigates through path history", func(t *testing.T) {
		model := &Model{
			dockerClient: docker.NewClient(),
			currentView:  FileBrowserView,
			viewHistory:  []ViewType{ComposeProcessListView, FileBrowserView},
		}
		vm := &FileBrowserViewModel{
			currentPath: "/usr/local",
			pathHistory: []string{"/", "/usr", "/usr/local"},
		}

		cmd := vm.HandleBack(model)
		assert.NotNil(t, cmd)
		assert.Equal(t, "/usr", vm.currentPath)
		assert.Equal(t, []string{"/", "/usr"}, vm.pathHistory)
		assert.Equal(t, FileBrowserView, model.currentView) // Should stay in file browser
	})

	t.Run("HandleBack switches to previous view when no more history", func(t *testing.T) {
		model := &Model{
			currentView: FileBrowserView,
			viewHistory: []ViewType{ComposeProcessListView, FileBrowserView},
		}
		vm := &FileBrowserViewModel{
			currentPath: "/",
			pathHistory: []string{"/"},
		}

		cmd := vm.HandleBack(model)
		assert.Nil(t, cmd)
		assert.Equal(t, ComposeProcessListView, model.currentView)
	})
}

func TestFileBrowserViewModel_Loaded(t *testing.T) {
	t.Run("Loaded updates container files", func(t *testing.T) {
		vm := &FileBrowserViewModel{
			selectedFile: 10, // Out of bounds
		}

		files := []models.ContainerFile{
			{Name: "file1.txt"},
			{Name: "file2.txt"},
		}

		vm.Loaded(files)
		assert.Equal(t, files, vm.containerFiles)
		assert.Equal(t, 0, vm.selectedFile, "Should reset selection when out of bounds")
	})

	t.Run("Loaded preserves valid selection", func(t *testing.T) {
		vm := &FileBrowserViewModel{
			selectedFile: 1,
		}

		files := []models.ContainerFile{
			{Name: "file1.txt"},
			{Name: "file2.txt"},
			{Name: "file3.txt"},
		}

		vm.Loaded(files)
		assert.Equal(t, files, vm.containerFiles)
		assert.Equal(t, 1, vm.selectedFile, "Should preserve valid selection")
	})
}

func TestFileBrowserViewModel_Title(t *testing.T) {
	dockerClient := docker.NewClient()
	container := docker.NewContainer(dockerClient, "test123", "test-container", "test-container", "running")
	vm := &FileBrowserViewModel{
		browsingContainer: container,
		currentPath:       "/usr/local",
	}

	title := vm.Title()
	assert.Equal(t, "File Browser: test-container [/usr/local]", title)
}

func TestFileBrowserViewModel_DoLoad(t *testing.T) {
	t.Run("DoLoad returns command to load files", func(t *testing.T) {
		model := &Model{
			dockerClient: docker.NewClient(),
		}
		container := docker.NewContainer(model.dockerClient, "container123", "test", "test", "running")
		vm := &FileBrowserViewModel{
			browsingContainer: container,
			currentPath:       "/usr/local",
		}

		cmd := vm.DoLoad(model)
		assert.NotNil(t, cmd)

		// Command should be created to load files
		// We don't execute it in tests as it would require a real container
	})
}
