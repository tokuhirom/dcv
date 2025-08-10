package ui

import (
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
