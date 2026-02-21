package ui

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandExecutionViewModel_ExecuteCommand(t *testing.T) {
	t.Run("aggressive command shows confirmation", func(t *testing.T) {
		model := &Model{
			currentView: DockerContainerListView,
		}
		vm := &CommandExecutionViewModel{}
		model.commandExecutionViewModel = *vm

		cmd := vm.ExecuteCommand(model, true, "stop", "container_id")

		assert.Nil(t, cmd, "Should return nil when showing confirmation")
		assert.True(t, vm.pendingConfirmation)
		assert.Equal(t, []string{"stop", "container_id"}, vm.pendingArgs)
		assert.Equal(t, CommandExecutionView, model.currentView)
	})

	t.Run("non-aggressive command executes immediately", func(t *testing.T) {
		model := &Model{
			currentView: DockerContainerListView,
		}
		vm := &CommandExecutionViewModel{}
		model.commandExecutionViewModel = *vm

		cmd := vm.ExecuteCommand(model, false, "logs", "container_id")

		assert.NotNil(t, cmd, "Should return command for immediate execution")
		assert.False(t, vm.pendingConfirmation)
		assert.Equal(t, CommandExecutionView, model.currentView)
	})
}

func TestCommandExecutionViewModel_HandleConfirmation(t *testing.T) {
	t.Run("confirm executes command", func(t *testing.T) {
		model := &Model{
			currentView: CommandExecutionView,
		}
		vm := &CommandExecutionViewModel{
			pendingConfirmation: true,
			pendingArgs:         []string{"stop", "container_id"},
		}
		model.commandExecutionViewModel = *vm

		cmd := vm.HandleConfirmation(model, true)

		assert.NotNil(t, cmd)
		assert.False(t, vm.pendingConfirmation)
		assert.Nil(t, vm.pendingArgs)
	})

	t.Run("cancel returns to previous view", func(t *testing.T) {
		model := &Model{
			currentView: CommandExecutionView,
			viewHistory: []ViewType{ComposeProcessListView},
		}
		vm := &CommandExecutionViewModel{
			pendingConfirmation: true,
			pendingArgs:         []string{"stop", "container_id"},
		}
		model.commandExecutionViewModel = *vm

		cmd := vm.HandleConfirmation(model, false)

		assert.Nil(t, cmd)
		assert.False(t, vm.pendingConfirmation)
		assert.Nil(t, vm.pendingArgs)
		assert.Equal(t, ComposeProcessListView, model.currentView)
	})

	t.Run("no-op when no pending confirmation", func(t *testing.T) {
		model := &Model{
			currentView: CommandExecutionView,
		}
		vm := &CommandExecutionViewModel{
			pendingConfirmation: false,
		}

		cmd := vm.HandleConfirmation(model, true)

		assert.Nil(t, cmd)
	})
}

func TestCommandExecutionViewModel_RenderConfirmationDialog(t *testing.T) {
	model := &Model{
		width:  80,
		Height: 24,
	}
	vm := &CommandExecutionViewModel{
		pendingConfirmation: true,
		pendingArgs:         []string{"stop", "container_id"},
	}

	result := vm.renderConfirmationDialog(model)

	// Check that the dialog contains key elements
	assert.Contains(t, result, "WARNING")
	assert.Contains(t, result, "Are you sure you want to execute:")
	assert.Contains(t, result, "docker stop container_id")
	assert.Contains(t, result, "Press 'y' to confirm, 'n' to cancel")
}

func TestCommandExecutionViewModel_LongStrings(t *testing.T) {
	t.Run("render with very long output lines", func(t *testing.T) {
		vm := &CommandExecutionViewModel{
			cmdString: "docker logs container123",
			output: func() []string {
				var lines []string
				for range 50 {
					lines = append(lines, strings.Repeat("A very long log line output ", 20))
				}
				return lines
			}(),
			done:     true,
			exitCode: 0,
		}
		model := &Model{width: 80, Height: 20}

		// Should not panic
		result := vm.render(model)
		assert.NotEmpty(t, result)
	})

	t.Run("render with very long command string", func(t *testing.T) {
		longArgs := strings.Repeat("--very-long-argument=value ", 30)
		vm := &CommandExecutionViewModel{
			cmdString: "docker " + longArgs,
			output:    []string{"output line 1"},
			done:      false,
		}
		model := &Model{width: 80, Height: 20}

		result := vm.render(model)
		assert.NotEmpty(t, result)
	})

	t.Run("render with narrow terminal", func(t *testing.T) {
		vm := &CommandExecutionViewModel{
			cmdString: "docker stop " + strings.Repeat("x", 200),
			output:    []string{strings.Repeat("output ", 100)},
			done:      true,
			exitCode:  1,
		}
		model := &Model{width: 30, Height: 20}

		result := vm.render(model)
		assert.NotEmpty(t, result)
	})

	t.Run("render with very small height", func(t *testing.T) {
		vm := &CommandExecutionViewModel{
			cmdString: "docker logs container",
			output: func() []string {
				var lines []string
				for range 100 {
					lines = append(lines, "log line")
				}
				return lines
			}(),
			done:     true,
			exitCode: 0,
		}
		model := &Model{width: 80, Height: 5}

		result := vm.render(model)
		assert.NotEmpty(t, result)
	})

	t.Run("confirmation dialog with very long args", func(t *testing.T) {
		longArgs := make([]string, 0, 30)
		for range 30 {
			longArgs = append(longArgs, strings.Repeat("long-arg-value-", 10))
		}
		vm := &CommandExecutionViewModel{
			pendingConfirmation: true,
			pendingArgs:         longArgs,
		}
		model := &Model{width: 80, Height: 24}

		result := vm.renderConfirmationDialog(model)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "WARNING")
	})

	t.Run("confirmation dialog with narrow terminal", func(t *testing.T) {
		vm := &CommandExecutionViewModel{
			pendingConfirmation: true,
			pendingArgs:         []string{"stop", strings.Repeat("container-id-", 20)},
		}
		model := &Model{width: 30, Height: 10}

		result := vm.renderConfirmationDialog(model)
		assert.NotEmpty(t, result)
	})

	t.Run("confirmation dialog with very small height", func(t *testing.T) {
		vm := &CommandExecutionViewModel{
			pendingConfirmation: true,
			pendingArgs:         []string{"kill", "container123"},
		}
		model := &Model{width: 80, Height: 5}

		result := vm.renderConfirmationDialog(model)
		assert.NotEmpty(t, result)
	})

	t.Run("render with zero dimensions", func(t *testing.T) {
		vm := &CommandExecutionViewModel{
			cmdString: "docker ps",
			output:    []string{"line1"},
			done:      true,
		}
		model := &Model{width: 0, Height: 0}

		result := vm.render(model)
		assert.Equal(t, "Loading...", result)
	})
}

func TestCommandExecutionViewModel_HandleBack(t *testing.T) {
	t.Run("returns RefreshMsg when going back", func(t *testing.T) {
		model := &Model{
			currentView: CommandExecutionView,
			viewHistory: []ViewType{ComposeProcessListView, CommandExecutionView},
		}
		vm := &CommandExecutionViewModel{
			done:     true,
			exitCode: 0,
		}
		model.commandExecutionViewModel = *vm

		cmd := vm.HandleBack(model)

		// Should switch to previous view
		assert.Equal(t, ComposeProcessListView, model.currentView)

		// Should return a command that generates RefreshMsg
		assert.NotNil(t, cmd, "Should return a command to trigger refresh")

		// Execute the command to verify it returns RefreshMsg
		msg := cmd()
		_, isRefreshMsg := msg.(RefreshMsg)
		assert.True(t, isRefreshMsg, "Command should return RefreshMsg")
	})
}
