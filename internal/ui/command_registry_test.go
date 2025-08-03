package ui

import (
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

func TestCommandRegistry(t *testing.T) {
	// Create a model and initialize it
	model := NewModel(ComposeProcessListView, "")
	model.initializeKeyHandlers()

	// Test that command registry is populated
	assert.NotNil(t, commandRegistry)
	assert.Greater(t, len(commandRegistry), 0)

	// Test that some common commands exist
	commonCommands := []string{
		"select-up-container",
		"select-down-container",
		"show-compose-log",
		"refresh-process-list",
		"kill-container",
		"stop-container",
	}

	for _, cmd := range commonCommands {
		t.Run(cmd, func(t *testing.T) {
			_, exists := commandRegistry[cmd]
			assert.True(t, exists, "Command %s should exist in registry", cmd)
		})
	}

	// Test aliases
	aliases := map[string]string{
		"up":     "select-up-container",
		"down":   "select-down-container",
		"logs":   "show-compose-log",
		"refresh": "refresh-process-list",
	}

	for alias, target := range aliases {
		t.Run("alias_"+alias, func(t *testing.T) {
			aliasCmd, aliasExists := commandRegistry[alias]
			targetCmd, targetExists := commandRegistry[target]
			
			assert.True(t, aliasExists, "Alias %s should exist", alias)
			assert.True(t, targetExists, "Target %s should exist", target)
			
			if aliasExists && targetExists {
				// Check that the handlers are the same
				assert.Equal(t, aliasCmd.Description, targetCmd.Description)
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
	t.Run("select-down-container", func(t *testing.T) {
		newModel, _ := model.executeKeyHandlerCommand("select-down-container")
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
	hasSelectUp := false
	hasShowLog := false
	for _, cmd := range commands {
		if cmd == "select-up-container" {
			hasSelectUp = true
		}
		if cmd == "show-compose-log" {
			hasShowLog = true
		}
	}
	assert.True(t, hasSelectUp, "Should have select-up-container command")
	assert.True(t, hasShowLog, "Should have show-compose-log command")
}

func TestGetCommandSuggestions(t *testing.T) {
	// Create a model and initialize it
	model := NewModel(ComposeProcessListView, "")
	model.initializeKeyHandlers()

	// Test prefix matching
	t.Run("prefix-match", func(t *testing.T) {
		suggestions := model.getCommandSuggestions("select-")
		assert.Greater(t, len(suggestions), 0)
		for _, s := range suggestions {
			assert.Contains(t, s, "select-")
		}
	})

	// Test substring matching
	t.Run("substring-match", func(t *testing.T) {
		suggestions := model.getCommandSuggestions("container")
		assert.Greater(t, len(suggestions), 0)
		for _, s := range suggestions {
			assert.Contains(t, s, "container")
		}
	})

	// Test no match
	t.Run("no-match", func(t *testing.T) {
		suggestions := model.getCommandSuggestions("xyz123")
		assert.Empty(t, suggestions)
	})
}