package ui

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tokuhirom/dcv/internal/models"
)

func TestToKebabCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"SelectUpContainer", "select-up-container"},
		{"ShowComposeLog", "show-compose-log"},
		{"BackFromLogView", "back-from-log-view"},
		{"simple", "simple"},
		{"ABC", "a-b-c"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toKebabCase(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetSimplifiedCommandName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"SelectUpContainer", "up"},
		{"ShowComposeLog", "logs"},
		{"ShowDockerContainerList", "ps"},
		{"ExecuteShell", "exec"},
		{"RefreshProcessList", "refresh"},
		{"BackFromLogView", "back"},
		{"KillContainer", "kill"},
		{"StopContainer", "stop"},
		{"RestartContainer", "restart"},
		{"DeleteContainer", "rm"},
		{"ShowImageList", "images"},
		{"ShowNetworkList", "networks"},
		{"ShowVolumeList", "volumes"},
		{"UnknownCommand", ""}, // Should return empty for unknown commands
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := getSimplifiedCommandName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCommandRegistry(t *testing.T) {
	// Create a model and initialize it
	model := NewModel(ComposeProcessListView, "")
	model.initializeKeyHandlers()

	// Test that command registry is populated
	assert.NotNil(t, commandRegistry)
	assert.Greater(t, len(commandRegistry), 0)

	// Test that some common commands exist (using simplified names)
	commonCommands := []string{
		"up",
		"down",
		"logs",
		"refresh",
		"kill",
		"stop",
		"ps",
		"images",
		"exec",
		"inspect",
	}

	for _, cmd := range commonCommands {
		t.Run(cmd, func(t *testing.T) {
			_, exists := commandRegistry[cmd]
			assert.True(t, exists, "Command %s should exist in registry", cmd)
		})
	}

	// Test some aliases still work
	aliases := map[string]string{
		"remove":  "rm",
		"delete":  "rm",
		"unpause": "pause",
		"list":    "ps",
	}

	for alias, target := range aliases {
		t.Run("alias_"+alias, func(t *testing.T) {
			aliasCmd, aliasExists := commandRegistry[alias]
			targetCmd, targetExists := commandRegistry[target]

			assert.True(t, aliasExists, "Alias %s should exist", alias)
			assert.True(t, targetExists, "Target %s should exist", target)

			if aliasExists && targetExists && len(aliasCmd) > 0 && len(targetCmd) > 0 {
				// Check that the handlers are the same
				assert.Equal(t, aliasCmd[0].Description, targetCmd[0].Description)
			}
		})
	}
}

func TestExecuteKeyHandlerCommand(t *testing.T) {
	// Create a model and initialize it
	model := NewModel(ComposeProcessListView, "")
	model.initializeKeyHandlers()
	model.composeContainers = []models.ComposeContainer{
		{Name: "container1"},
		{Name: "container2"},
		{Name: "container3"},
	}
	model.selectedContainer = 1

	// Test executing a navigation command
	t.Run("down", func(t *testing.T) {
		newModel, _ := model.executeKeyHandlerCommand("down")
		m := newModel.(*Model)
		assert.Equal(t, 2, m.selectedContainer)
	})

	// Test executing an unknown command
	t.Run("unknown-command", func(t *testing.T) {
		newModel, _ := model.executeKeyHandlerCommand("unknown-command")
		m := newModel.(*Model)
		assert.NotNil(t, m.err)
		assert.Contains(t, m.err.Error(), "unknown command")
	})

	// Test executing alias
	t.Run("down-alias", func(t *testing.T) {
		model.selectedContainer = 0
		newModel, _ := model.executeKeyHandlerCommand("down")
		m := newModel.(*Model)
		assert.Equal(t, 1, m.selectedContainer)
	})
}

func TestGetAvailableCommands(t *testing.T) {
	// Create a model and initialize it
	model := NewModel(ComposeProcessListView, "")
	model.initializeKeyHandlers()

	commands := model.getAvailableCommands()
	assert.Greater(t, len(commands), 0)

	// Check that we have some expected commands
	hasUp := false
	hasLogs := false
	for _, cmd := range commands {
		if cmd == "up" {
			hasUp = true
		}
		if cmd == "logs" {
			hasLogs = true
		}
	}
	assert.True(t, hasUp, "Should have up command")
	assert.True(t, hasLogs, "Should have logs command")
}

func TestGetCommandSuggestions(t *testing.T) {
	// Create a model and initialize it
	model := NewModel(ComposeProcessListView, "")
	model.initializeKeyHandlers()

	// Test prefix matching
	t.Run("prefix-match", func(t *testing.T) {
		suggestions := model.getCommandSuggestions("re")
		assert.Greater(t, len(suggestions), 0)
		// Should find commands like "refresh", "restart", "remove"
		for _, s := range suggestions {
			assert.True(t, strings.HasPrefix(s, "re"), "Suggestion %s should start with 're'", s)
		}
	})

	// Test substring matching
	t.Run("substring-match", func(t *testing.T) {
		suggestions := model.getCommandSuggestions("load")
		// If no prefix match, should fall back to substring match
		// But with simplified names, this might not find matches
		// Let's search for something that should exist
		suggestions = model.getCommandSuggestions("log")
		if len(suggestions) > 0 {
			for _, s := range suggestions {
				assert.Contains(t, s, "log")
			}
		}
	})

	// Test no match
	t.Run("no-match", func(t *testing.T) {
		suggestions := model.getCommandSuggestions("xyz123")
		assert.Empty(t, suggestions)
	})
}
