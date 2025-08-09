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

func TestCommandExecutionViewModel_GetConfirmationTarget(t *testing.T) {
	vm := &CommandExecutionViewModel{}

	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "container command with short ID",
			args:     []string{"stop", "abc123"},
			expected: "container abc123",
		},
		{
			name:     "container command with long ID",
			args:     []string{"stop", "abc123def456ghi789"},
			expected: "container abc123def456",
		},
		{
			name:     "compose command with project name",
			args:     []string{"compose", "-p", "myproject", "down"},
			expected: "project 'myproject'",
		},
		{
			name:     "compose command without project name",
			args:     []string{"compose", "down"},
			expected: "compose services",
		},
		{
			name:     "network removal",
			args:     []string{"network", "rm", "mynetwork"},
			expected: "network 'mynetwork'",
		},
		{
			name:     "volume removal",
			args:     []string{"volume", "rm", "myvolume"},
			expected: "volume 'myvolume'",
		},
		{
			name:     "image removal",
			args:     []string{"rmi", "nginx:latest"},
			expected: "image 'nginx:latest'",
		},
		{
			name:     "empty args",
			args:     []string{},
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vm.getConfirmationTarget(tt.args)
			assert.Equal(t, tt.expected, result, "getConfirmationTarget(%v) should return %v", tt.args, tt.expected)
		})
	}
}

func TestCommandExecutionViewModel_GetCommandDisplay(t *testing.T) {
	vm := &CommandExecutionViewModel{}

	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "stop command",
			args:     []string{"stop", "container_id"},
			expected: "stop",
		},
		{
			name:     "kill command",
			args:     []string{"kill", "container_id"},
			expected: "forcefully stop",
		},
		{
			name:     "rm command",
			args:     []string{"rm", "container_id"},
			expected: "remove",
		},
		{
			name:     "compose down",
			args:     []string{"compose", "-p", "project", "down"},
			expected: "stop and remove all services in",
		},
		{
			name:     "compose up",
			args:     []string{"compose", "-p", "project", "up", "-d"},
			expected: "start all services in",
		},
		{
			name:     "network rm",
			args:     []string{"network", "rm", "network_name"},
			expected: "remove network",
		},
		{
			name:     "volume rm",
			args:     []string{"volume", "rm", "volume_name"},
			expected: "remove volume",
		},
		{
			name:     "empty args",
			args:     []string{},
			expected: "execute command",
		},
		{
			name:     "unknown command",
			args:     []string{"unknown", "arg"},
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vm.getCommandDisplay(tt.args)
			assert.Equal(t, tt.expected, result, "getCommandDisplay(%v) should return %v", tt.args, tt.expected)
		})
	}
}

func TestCommandExecutionViewModel_HandleConfirmation(t *testing.T) {
	t.Run("confirm executes command", func(t *testing.T) {
		model := &Model{
			currentView: CommandExecutionView,
		}
		vm := &CommandExecutionViewModel{
			pendingConfirmation: true,
			pendingArgs:         []string{"stop", "container_id"},
			confirmationTarget:  "container container_id",
		}
		model.commandExecutionViewModel = *vm

		cmd := vm.HandleConfirmation(model, true)

		assert.NotNil(t, cmd)
		assert.False(t, vm.pendingConfirmation)
		assert.Nil(t, vm.pendingArgs)
		assert.Empty(t, vm.confirmationTarget)
	})

	t.Run("cancel returns to previous view", func(t *testing.T) {
		model := &Model{
			currentView: CommandExecutionView,
			viewHistory: []ViewType{ComposeProcessListView},
		}
		vm := &CommandExecutionViewModel{
			pendingConfirmation: true,
			pendingArgs:         []string{"stop", "container_id"},
			confirmationTarget:  "container container_id",
		}
		model.commandExecutionViewModel = *vm

		cmd := vm.HandleConfirmation(model, false)

		assert.Nil(t, cmd)
		assert.False(t, vm.pendingConfirmation)
		assert.Nil(t, vm.pendingArgs)
		assert.Empty(t, vm.confirmationTarget)
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
		confirmationTarget:  "container nginx",
	}

	result := vm.renderConfirmationDialog(model)

	// Check that the dialog contains key elements
	assert.Contains(t, result, "WARNING")
	assert.Contains(t, result, "Are you sure you want to stop")
	assert.Contains(t, result, "container nginx")
	assert.Contains(t, result, "Press 'y' to confirm, 'n' to cancel")
}
