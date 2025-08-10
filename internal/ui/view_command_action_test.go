package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/tokuhirom/dcv/internal/docker"
)

func TestCommandActionViewModel_Initialize(t *testing.T) {
	tests := []struct {
		name            string
		containerState  string
		expectedActions []string
	}{
		{
			name:           "running container shows stop/restart/kill/pause actions",
			containerState: "running",
			expectedActions: []string{
				"View Logs",
				"Inspect",
				"Browse Files",
				"Execute Shell",
				"Stop",
				"Restart",
				"Kill",
				"Pause",
			},
		},
		{
			name:           "paused container shows unpause action",
			containerState: "paused",
			expectedActions: []string{
				"View Logs",
				"Inspect",
				"Browse Files",
				"Execute Shell",
				"Unpause",
			},
		},
		{
			name:           "exited container shows start/delete actions",
			containerState: "exited",
			expectedActions: []string{
				"View Logs",
				"Inspect",
				"Browse Files",
				"Execute Shell",
				"Start",
				"Delete",
			},
		},
		{
			name:           "created container shows start/delete actions",
			containerState: "created",
			expectedActions: []string{
				"View Logs",
				"Inspect",
				"Browse Files",
				"Execute Shell",
				"Start",
				"Delete",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := &CommandActionViewModel{}
			container := docker.NewContainer("test-id", "test-container", "test-service", tt.containerState)

			vm.Initialize(container)

			// Check that we have the right actions
			assert.Equal(t, len(tt.expectedActions), len(vm.actions))
			for i, expectedName := range tt.expectedActions {
				assert.Equal(t, expectedName, vm.actions[i].Name)
			}

			// Check that the container is set
			assert.Equal(t, container, vm.targetContainer)
			assert.Equal(t, 0, vm.selectedAction)
		})
	}
}

func TestCommandActionViewModel_HandleUp(t *testing.T) {
	vm := &CommandActionViewModel{
		actions: []CommandAction{
			{Name: "Action1"},
			{Name: "Action2"},
			{Name: "Action3"},
		},
		selectedAction: 2,
	}

	// Move up from position 2 to 1
	cmd := vm.HandleUp()
	assert.Nil(t, cmd)
	assert.Equal(t, 1, vm.selectedAction)

	// Move up from position 1 to 0
	cmd = vm.HandleUp()
	assert.Nil(t, cmd)
	assert.Equal(t, 0, vm.selectedAction)

	// Try to move up from position 0 (should stay at 0)
	cmd = vm.HandleUp()
	assert.Nil(t, cmd)
	assert.Equal(t, 0, vm.selectedAction)
}

func TestCommandActionViewModel_HandleDown(t *testing.T) {
	vm := &CommandActionViewModel{
		actions: []CommandAction{
			{Name: "Action1"},
			{Name: "Action2"},
			{Name: "Action3"},
		},
		selectedAction: 0,
	}

	// Move down from position 0 to 1
	cmd := vm.HandleDown()
	assert.Nil(t, cmd)
	assert.Equal(t, 1, vm.selectedAction)

	// Move down from position 1 to 2
	cmd = vm.HandleDown()
	assert.Nil(t, cmd)
	assert.Equal(t, 2, vm.selectedAction)

	// Try to move down from position 2 (should stay at 2)
	cmd = vm.HandleDown()
	assert.Nil(t, cmd)
	assert.Equal(t, 2, vm.selectedAction)
}

func TestCommandActionViewModel_HandleSelect(t *testing.T) {
	t.Run("executes selected action and switches to previous view", func(t *testing.T) {
		model := &Model{
			currentView:  CommandActionView,
			viewHistory:  []ViewType{ComposeProcessListView, CommandActionView},
			dockerClient: docker.NewClient(),
		}
		model.initializeKeyHandlers()

		executedAction := false
		vm := &CommandActionViewModel{
			actions: []CommandAction{
				{
					Name: "Test Action",
					Handler: func(m *Model, c *docker.Container) tea.Cmd {
						executedAction = true
						return nil
					},
				},
			},
			selectedAction:  0,
			targetContainer: docker.NewContainer("test-id", "test-container", "test-service", "running"),
		}
		model.commandActionViewModel = *vm

		cmd := vm.HandleSelect(model)
		// The handler itself returns nil, so cmd will be nil
		assert.Nil(t, cmd)
		assert.True(t, executedAction)
		// View should have been switched back to previous
		assert.Equal(t, ComposeProcessListView, model.currentView)
	})

	t.Run("handles out of bounds selection", func(t *testing.T) {
		model := &Model{
			currentView:  CommandActionView,
			dockerClient: docker.NewClient(),
		}

		vm := &CommandActionViewModel{
			actions:        []CommandAction{},
			selectedAction: 0,
		}
		model.commandActionViewModel = *vm

		cmd := vm.HandleSelect(model)
		assert.Nil(t, cmd)
	})

	t.Run("handles negative selection", func(t *testing.T) {
		model := &Model{
			currentView:  CommandActionView,
			dockerClient: docker.NewClient(),
		}

		vm := &CommandActionViewModel{
			actions: []CommandAction{
				{Name: "Action1"},
			},
			selectedAction: -1,
		}
		model.commandActionViewModel = *vm

		cmd := vm.HandleSelect(model)
		assert.Nil(t, cmd)
	})
}

func TestCommandActionViewModel_HandleBack(t *testing.T) {
	model := &Model{
		currentView: CommandActionView,
		viewHistory: []ViewType{ComposeProcessListView, CommandActionView},
	}

	vm := &CommandActionViewModel{}
	model.commandActionViewModel = *vm

	cmd := vm.HandleBack(model)
	assert.Nil(t, cmd)
	assert.Equal(t, ComposeProcessListView, model.currentView)
}

func TestCommandActionViewModel_Render(t *testing.T) {
	tests := []struct {
		name              string
		container         *docker.Container
		actions           []CommandAction
		selectedAction    int
		expectContains    []string
		expectNotContains []string
	}{
		{
			name:      "renders with container info",
			container: docker.NewContainer("test-id", "test-container", "test-service", "running"),
			actions: []CommandAction{
				{Key: "S", Name: "Stop", Description: "Stop the container", Aggressive: true},
				{Key: "R", Name: "Restart", Description: "Restart the container", Aggressive: true},
				{Key: "I", Name: "Inspect", Description: "View details", Aggressive: false},
			},
			selectedAction: 1,
			expectContains: []string{
				"Select Action for test-container",
				"Container: test-container",
				"State: running",
				"[S] Stop - Stop the container",
				"> [R] Restart - Restart the container", // Selected with prefix
				"[I] Inspect - View details",
				"Use ↑/↓ to select, Enter to execute, Esc to cancel",
			},
		},
		{
			name:           "renders no container selected",
			container:      nil,
			actions:        []CommandAction{},
			selectedAction: 0,
			expectContains: []string{
				"No container selected",
			},
			expectNotContains: []string{
				"Select Action for",
				"Container:",
				"State:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &Model{
				width:  80,
				Height: 24,
			}

			vm := &CommandActionViewModel{
				targetContainer: tt.container,
				actions:         tt.actions,
				selectedAction:  tt.selectedAction,
			}

			output := vm.render(model)

			for _, expected := range tt.expectContains {
				assert.Contains(t, output, expected)
			}

			for _, notExpected := range tt.expectNotContains {
				assert.NotContains(t, output, notExpected)
			}
		})
	}
}

func TestCommandActionViewModel_ActionColoring(t *testing.T) {
	model := &Model{
		width:  80,
		Height: 24,
	}

	container := docker.NewContainer("test-id", "test-container", "test-service", "running")

	vm := &CommandActionViewModel{
		targetContainer: container,
		actions: []CommandAction{
			{Key: "S", Name: "Stop", Description: "Stop", Aggressive: true},
			{Key: "I", Name: "Inspect", Description: "Inspect", Aggressive: false},
		},
		selectedAction: 0,
	}

	output := vm.render(model)

	// The output should contain styled text, though we can't easily test
	// the exact ANSI codes. We can at least verify the structure is there.
	assert.Contains(t, output, "[S] Stop")
	assert.Contains(t, output, "[I] Inspect")
}

func TestCommandActionViewModel_Integration(t *testing.T) {
	t.Run("full flow from initialization to action execution", func(t *testing.T) {
		// Create a model with proper setup
		model := &Model{
			currentView:  DockerContainerListView,
			viewHistory:  []ViewType{DockerContainerListView},
			width:        80,
			Height:       24,
			dockerClient: docker.NewClient(),
		}
		model.initializeKeyHandlers()
		model.commandActionViewModel = CommandActionViewModel{}

		// Initialize with a running container
		container := docker.NewContainer("test-id", "test-container", "test-service", "running")
		model.commandActionViewModel.Initialize(container)

		// Should have actions for running container
		assert.Greater(t, len(model.commandActionViewModel.actions), 4)

		// Navigate down
		model.commandActionViewModel.HandleDown()
		assert.Equal(t, 1, model.commandActionViewModel.selectedAction)

		// Navigate up
		model.commandActionViewModel.HandleUp()
		assert.Equal(t, 0, model.commandActionViewModel.selectedAction)

		// Test render
		output := model.commandActionViewModel.render(model)
		assert.Contains(t, output, "test-container")
		assert.Contains(t, output, "running")

		// Test back navigation
		model.SwitchView(CommandActionView)
		model.commandActionViewModel.HandleBack(model)
		assert.Equal(t, DockerContainerListView, model.currentView)
	})
}

func TestCommandActionViewModel_ActionHandlers(t *testing.T) {
	t.Run("verify action handlers call correct Cmd methods", func(t *testing.T) {
		model := &Model{
			currentView:                  CommandActionView,
			viewHistory:                  []ViewType{ComposeProcessListView, CommandActionView},
			dockerClient:                 docker.NewClient(),
			composeProcessListViewModel:  ComposeProcessListViewModel{},
			dockerContainerListViewModel: DockerContainerListViewModel{},
			logViewModel:                 LogViewModel{},
			inspectViewModel:             InspectViewModel{},
			fileBrowserViewModel:         FileBrowserViewModel{},
			commandExecutionViewModel:    CommandExecutionViewModel{},
		}
		model.initializeKeyHandlers()

		container := docker.NewContainer("test-id", "test-container", "test-service", "running")

		vm := &CommandActionViewModel{}
		vm.Initialize(container)
		model.commandActionViewModel = *vm

		// Test that each action has a valid handler
		for _, action := range vm.actions {
			assert.NotNil(t, action.Handler, "Action %s should have a handler", action.Name)

			// Call the handler to ensure it doesn't panic
			// (though the actual commands may return nil in test environment)
			_ = action.Handler(model, container)
		}
	})
}

func TestCommandActionViewModel_StateSpecificActions(t *testing.T) {
	tests := []struct {
		state           string
		expectedActions map[string]bool
	}{
		{
			state: "running",
			expectedActions: map[string]bool{
				"Stop":    true,
				"Restart": true,
				"Kill":    true,
				"Pause":   true,
				"Start":   false,
				"Delete":  false,
			},
		},
		{
			state: "paused",
			expectedActions: map[string]bool{
				"Unpause": true,
				"Stop":    false,
				"Start":   false,
				"Delete":  false,
			},
		},
		{
			state: "exited",
			expectedActions: map[string]bool{
				"Start":  true,
				"Delete": true,
				"Stop":   false,
				"Pause":  false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.state+" container", func(t *testing.T) {
			vm := &CommandActionViewModel{}
			container := docker.NewContainer("test-id", "test-container", "test-service", tt.state)

			vm.Initialize(container)

			// Check which actions are present
			actionNames := make(map[string]bool)
			for _, action := range vm.actions {
				actionNames[action.Name] = true
			}

			for actionName, shouldExist := range tt.expectedActions {
				if shouldExist {
					assert.True(t, actionNames[actionName],
						"Action %s should exist for %s container", actionName, tt.state)
				} else {
					assert.False(t, actionNames[actionName],
						"Action %s should not exist for %s container", actionName, tt.state)
				}
			}
		})
	}
}

func TestCommandActionViewModel_EmptyActions(t *testing.T) {
	vm := &CommandActionViewModel{
		actions:        []CommandAction{},
		selectedAction: 0,
	}

	// HandleUp with empty actions
	cmd := vm.HandleUp()
	assert.Nil(t, cmd)
	assert.Equal(t, 0, vm.selectedAction)

	// HandleDown with empty actions
	cmd = vm.HandleDown()
	assert.Nil(t, cmd)
	assert.Equal(t, 0, vm.selectedAction)
}

func TestCommandActionViewModel_RenderFormattting(t *testing.T) {
	model := &Model{
		width:  80,
		Height: 24,
	}

	container := docker.NewContainer("very-long-container-id-that-should-not-break-rendering", "very-long-container-name-that-might-be-truncated", "service", "running")

	vm := &CommandActionViewModel{
		targetContainer: container,
		actions: []CommandAction{
			{
				Key:         "X",
				Name:        "Very Long Action Name That Should Still Render",
				Description: "This is a very long description that explains what this action does",
				Aggressive:  false,
			},
		},
		selectedAction: 0,
	}

	output := vm.render(model)

	// Should contain the container name
	assert.Contains(t, output, "very-long-container-name")

	// Should contain the action
	assert.Contains(t, output, "[X]")
	assert.Contains(t, output, "Very Long Action Name")

	// Should not panic or produce malformed output
	lines := strings.Split(output, "\n")
	assert.Greater(t, len(lines), 5) // Should have multiple lines of output
}
