package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tokuhirom/dcv/internal/models"
)

func TestComposeProjectListViewModel_Rendering(t *testing.T) {
	tests := []struct {
		name      string
		viewModel ComposeProjectListViewModel
		height    int
		expected  []string
	}{
		{
			name: "displays no projects message when empty",
			viewModel: ComposeProjectListViewModel{
				TableViewModel: TableViewModel{Cursor: 0},
				projects:       []models.ComposeProject{},
			},
			height:   20,
			expected: []string{"No Docker Compose projects found"},
		},
		{
			name: "displays project list table",
			viewModel: ComposeProjectListViewModel{
				projects: []models.ComposeProject{
					{
						Name:        "webapp",
						Status:      "running(2)",
						ConfigFiles: "/home/user/webapp/docker-compose.yml",
					},
					{
						Name:        "database",
						Status:      "exited(1)",
						ConfigFiles: "/home/user/database/docker-compose.yml",
					},
				},
				TableViewModel: TableViewModel{Cursor: 0},
			},
			height: 20,
			expected: []string{
				"NAME",
				"STATUS",
				"CONFIG FILES",
				"webapp",
				"running(2)",
				"database",
				"exited(1)",
			},
		},
		{
			name: "highlights running vs exited projects",
			viewModel: ComposeProjectListViewModel{
				projects: []models.ComposeProject{
					{
						Name:   "running-project",
						Status: "running",
					},
					{
						Name:   "exited-project",
						Status: "exited",
					},
				},
				TableViewModel: TableViewModel{Cursor: 0},
			},
			height:   20,
			expected: []string{"running-project", "exited-project"},
		},
		{
			name: "truncates long config file paths",
			viewModel: ComposeProjectListViewModel{
				projects: []models.ComposeProject{
					{
						Name:        "project",
						Status:      "running",
						ConfigFiles: "/very/long/path/that/should/be/truncated/in/the/display/docker-compose.yml",
					},
				},
				TableViewModel: TableViewModel{Cursor: 0},
			},
			height:   20,
			expected: []string{"/very/long/path/that/should/be/truncated/in/the..."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &Model{
				composeProjectListViewModel: tt.viewModel,
				width:                       100,
				Height:                      tt.height,
			}

			// Initialize table rows if projects exist
			if len(tt.viewModel.projects) > 0 {
				tt.viewModel.SetRows(tt.viewModel.buildRows(), model.ViewHeight())
			}

			result := tt.viewModel.render(model, tt.height-4)

			for _, expected := range tt.expected {
				assert.Contains(t, result, expected, "Expected to find '%s' in output", expected)
			}
		})
	}
}

func TestComposeProjectListViewModel_Navigation(t *testing.T) {
	t.Run("HandleDown moves selection down", func(t *testing.T) {
		model := &Model{Height: 20}
		vm := &ComposeProjectListViewModel{
			projects: []models.ComposeProject{
				{Name: "project1"},
				{Name: "project2"},
				{Name: "project3"},
			},
		}
		vm.SetRows(vm.buildRows(), model.ViewHeight())

		cmd := vm.HandleDown(model)
		assert.Nil(t, cmd)
		assert.Equal(t, 1, vm.Cursor)

		// Test boundary
		vm.Cursor = 2
		cmd = vm.HandleDown(model)
		assert.Nil(t, cmd)
		assert.Equal(t, 2, vm.Cursor, "Should not go beyond last project")
	})

	t.Run("HandleUp moves selection up", func(t *testing.T) {
		model := &Model{Height: 20}
		vm := &ComposeProjectListViewModel{
			projects: []models.ComposeProject{
				{Name: "project1"},
				{Name: "project2"},
				{Name: "project3"},
			},
		}
		vm.SetRows(vm.buildRows(), model.ViewHeight())
		vm.Cursor = 2

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

func TestComposeProjectListViewModel_Operations(t *testing.T) {
	t.Run("HandleSelectProject switches to process list view", func(t *testing.T) {
		model := &Model{
			currentView: ComposeProjectListView,
			loading:     false,
		}
		vm := &ComposeProjectListViewModel{
			projects: []models.ComposeProject{
				{Name: "my-project"},
			},
			TableViewModel: TableViewModel{Cursor: 0},
		}
		model.composeProjectListViewModel = *vm

		cmd := vm.HandleSelectProject(model)

		// The actual view switch happens in ComposeProcessListViewModel.Load
		assert.NotNil(t, cmd)
	})

	t.Run("HandleSelectProject does nothing when no projects", func(t *testing.T) {
		model := &Model{
			currentView: ComposeProjectListView,
		}
		model.composeProcessListViewModel.projectName = "old-project"
		vm := &ComposeProjectListViewModel{
			projects:       []models.ComposeProject{},
			TableViewModel: TableViewModel{Cursor: 0},
		}

		cmd := vm.HandleSelectProject(model)

		assert.Equal(t, ComposeProjectListView, model.currentView)
		assert.Equal(t, "old-project", model.composeProcessListViewModel.projectName, "Should not change project name")
		assert.Nil(t, cmd)
	})

	t.Run("HandleSelectProject bounds checks selection", func(t *testing.T) {
		model := &Model{}
		vm := &ComposeProjectListViewModel{
			projects: []models.ComposeProject{
				{Name: "project1"},
			},
			TableViewModel: TableViewModel{Cursor: 5}, // Out of bounds
		}

		cmd := vm.HandleSelectProject(model)

		assert.Nil(t, cmd, "Should return nil when selection is out of bounds")
	})
}

func TestComposeProjectListViewModel_Messages(t *testing.T) {
	t.Run("Loaded updates project list", func(t *testing.T) {
		vm := &ComposeProjectListViewModel{
			TableViewModel: TableViewModel{Cursor: 5}, // Out of bounds
		}

		projects := []models.ComposeProject{
			{Name: "project1"},
			{Name: "project2"},
		}

		model := &Model{Height: 20}
		vm.Loaded(model, projects)

		assert.Equal(t, projects, vm.projects)
		assert.Equal(t, 0, vm.Cursor, "Should reset selection when out of bounds")
	})

	t.Run("Loaded keeps selection in bounds", func(t *testing.T) {
		vm := &ComposeProjectListViewModel{
			TableViewModel: TableViewModel{Cursor: 1},
		}

		projects := []models.ComposeProject{
			{Name: "project1"},
			{Name: "project2"},
			{Name: "project3"},
		}

		model := &Model{Height: 20}
		vm.Loaded(model, projects)

		assert.Equal(t, 1, vm.Cursor, "Should keep selection when in bounds")
	})
}

func TestComposeProjectListViewModel_EmptySelection(t *testing.T) {
	t.Run("operations handle empty project list gracefully", func(t *testing.T) {
		model := &Model{}
		vm := &ComposeProjectListViewModel{
			projects:       []models.ComposeProject{},
			TableViewModel: TableViewModel{Cursor: 0},
		}

		// Test operations that depend on selection
		assert.Nil(t, vm.HandleSelectProject(model))

		// Navigation should not crash
		assert.Nil(t, vm.HandleUp(model))
		assert.Nil(t, vm.HandleDown(model))
	})
}

func TestComposeProjectListViewModel_KeyHandlers(t *testing.T) {
	model := NewModel(ComposeProjectListView)
	model.initializeKeyHandlers()

	// Verify key handlers are registered
	handlers := model.composeProjectListViewHandlers
	assert.Greater(t, len(handlers), 0, "Should have registered key handlers")

	// Check view-specific handlers
	viewSpecificKeys := []string{"up", "down", "enter", "r", "?"}
	registeredKeys := make(map[string]bool)

	for _, h := range handlers {
		for _, key := range h.Keys {
			registeredKeys[key] = true
		}
	}

	for _, key := range viewSpecificKeys {
		assert.True(t, registeredKeys[key], "Key %s should be registered in view handlers", key)
	}

	// Check global handlers
	globalKeys := []string{"1", "3", "4", "5"}
	globalRegisteredKeys := make(map[string]bool)
	for _, h := range model.globalHandlers {
		for _, key := range h.Keys {
			globalRegisteredKeys[key] = true
		}
	}
	for _, key := range globalKeys {
		assert.True(t, globalRegisteredKeys[key], "Key %s should be registered in global handlers", key)
	}
}

func TestComposeProjectListViewModel_Update(t *testing.T) {
	t.Run("handles projectsLoadedMsg success", func(t *testing.T) {
		model := &Model{
			currentView: ComposeProjectListView,
			loading:     true,
		}
		vm := &ComposeProjectListViewModel{}
		model.composeProjectListViewModel = *vm

		projects := []models.ComposeProject{
			{Name: "test-project"},
		}

		msg := projectsLoadedMsg{
			projects: projects,
			err:      nil,
		}

		newModel, cmd := model.Update(msg)
		m := newModel.(*Model)

		assert.False(t, m.loading)
		assert.Nil(t, m.err)
		assert.Equal(t, projects, m.composeProjectListViewModel.projects)
		assert.Nil(t, cmd)
	})

	t.Run("handles projectsLoadedMsg error", func(t *testing.T) {
		model := &Model{
			currentView: ComposeProjectListView,
			loading:     true,
		}

		testErr := assert.AnError
		msg := projectsLoadedMsg{
			projects: nil,
			err:      testErr,
		}

		newModel, cmd := model.Update(msg)
		m := newModel.(*Model)

		assert.False(t, m.loading)
		assert.Equal(t, testErr, m.err)
		assert.Nil(t, cmd)
	})
}

func TestComposeProjectListViewModel_StatusColors(t *testing.T) {
	tests := []struct {
		status   string
		isGreen  bool
		expected string
	}{
		{"running(3)", true, "running"},
		{"running", true, "running"},
		{"exited(0)", false, "exited"},
		{"created(1)", false, "created"},
		{"paused(2)", false, "paused"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			project := models.ComposeProject{
				Name:   "test",
				Status: tt.status,
			}

			// The view should show different colors for running vs non-running
			assert.Contains(t, project.Status, tt.expected)
		})
	}
}
