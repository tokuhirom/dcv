package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandExecutionViewModel_NeedsConfirmation(t *testing.T) {
	vm := &CommandExecutionViewModel{}

	tests := []struct {
		name     string
		args     []string
		expected bool
	}{
		{
			name:     "stop command needs confirmation",
			args:     []string{"stop", "container_id"},
			expected: true,
		},
		{
			name:     "start command needs confirmation",
			args:     []string{"start", "container_id"},
			expected: true,
		},
		{
			name:     "restart command needs confirmation",
			args:     []string{"restart", "container_id"},
			expected: true,
		},
		{
			name:     "kill command needs confirmation",
			args:     []string{"kill", "container_id"},
			expected: true,
		},
		{
			name:     "pause command needs confirmation",
			args:     []string{"pause", "container_id"},
			expected: true,
		},
		{
			name:     "unpause command needs confirmation",
			args:     []string{"unpause", "container_id"},
			expected: true,
		},
		{
			name:     "rm command needs confirmation",
			args:     []string{"rm", "container_id"},
			expected: true,
		},
		{
			name:     "rmi command needs confirmation",
			args:     []string{"rmi", "image_id"},
			expected: true,
		},
		{
			name:     "compose down needs confirmation",
			args:     []string{"compose", "-p", "project", "down"},
			expected: true,
		},
		{
			name:     "compose stop needs confirmation",
			args:     []string{"compose", "-p", "project", "stop"},
			expected: true,
		},
		{
			name:     "compose restart needs confirmation",
			args:     []string{"compose", "-p", "project", "restart", "service"},
			expected: true,
		},
		{
			name:     "network rm needs confirmation",
			args:     []string{"network", "rm", "network_name"},
			expected: true,
		},
		{
			name:     "volume rm needs confirmation",
			args:     []string{"volume", "rm", "volume_name"},
			expected: true,
		},
		{
			name:     "logs command does not need confirmation",
			args:     []string{"logs", "container_id"},
			expected: false,
		},
		{
			name:     "ps command does not need confirmation",
			args:     []string{"ps"},
			expected: false,
		},
		{
			name:     "inspect command does not need confirmation",
			args:     []string{"inspect", "container_id"},
			expected: false,
		},
		{
			name:     "exec command does not need confirmation",
			args:     []string{"exec", "-it", "container_id", "/bin/sh"},
			expected: false,
		},
		{
			name:     "compose up does not need confirmation",
			args:     []string{"compose", "-p", "project", "up", "-d"},
			expected: false,
		},
		{
			name:     "empty args does not need confirmation",
			args:     []string{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vm.needsConfirmation(tt.args)
			assert.Equal(t, tt.expected, result, "needsConfirmation(%v) should return %v", tt.args, tt.expected)
		})
	}
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
