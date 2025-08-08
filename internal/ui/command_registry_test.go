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

func TestCommandRegistry(t *testing.T) {
	// Create a model and initialize it
	model := NewModel(ComposeProcessListView, "")
	model.initializeKeyHandlers()

	// Test that command registry is populated
	assert.NotNil(t, commandRegistry)
	assert.Greater(t, len(commandRegistry), 0)

	// Test that some common commands exist
	commonCommands := []string{
		"up",
		"down",
		"log",
		"refresh",
		"kill",
		"stop",
	}

	for _, cmd := range commonCommands {
		t.Run(cmd, func(t *testing.T) {
			_, exists := commandRegistry[cmd]
			assert.True(t, exists, "Command %s should exist in registry", cmd)
		})
	}

	// Test aliases
	aliases := map[string]string{
		"logs": "log",
		"exec": "shell",
		"rm":   "remove",
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
	model.composeProcessListViewModel.composeContainers = []models.ComposeContainer{
		{Name: "container1"},
		{Name: "container2"},
		{Name: "container3"},
	}
	model.composeProcessListViewModel.selectedContainer = 1

	// Test executing a navigation command
	t.Run("down", func(t *testing.T) {
		newModel, _ := model.commandViewModel.executeKeyHandlerCommand(&model, "down")
		m := newModel.(*Model)
		assert.Equal(t, 2, m.composeProcessListViewModel.selectedContainer)
	})

	// Test executing an unknown command
	t.Run("unknown-command", func(t *testing.T) {
		newModel, _ := model.commandViewModel.executeKeyHandlerCommand(&model, "unknown-command")
		m := newModel.(*Model)
		assert.NotNil(t, m.err)
		assert.Contains(t, m.err.Error(), "unknown command")
	})

	// Test executing alias
	t.Run("down-alias", func(t *testing.T) {
		// Create a fresh model for this test
		testModel := NewModel(ComposeProcessListView, "")
		testModel.initializeKeyHandlers()
		testModel.composeProcessListViewModel.composeContainers = []models.ComposeContainer{
			{Name: "container1"},
			{Name: "container2"},
			{Name: "container3"},
		}
		testModel.composeProcessListViewModel.selectedContainer = 0

		// Check that "down" command exists and what handler it has
		if cmd, exists := commandRegistry["down"]; exists {
			t.Logf("Found 'down' command with description: %s, ViewMask: %v", cmd.Description, cmd.ViewMask)
			// List all commands that have "down" in their name
			for name, regCmd := range commandRegistry {
				if strings.Contains(name, "down") {
					t.Logf("Command '%s' has ViewMask: %v", name, regCmd.ViewMask)
				}
			}
		}

		t.Logf("Before: selectedContainer=%d, numContainers=%d",
			testModel.composeProcessListViewModel.selectedContainer,
			len(testModel.composeProcessListViewModel.composeContainers))

		newModel, cmd := testModel.commandViewModel.executeKeyHandlerCommand(&testModel, "down")
		m := newModel.(*Model)

		t.Logf("After: selectedContainer=%d, cmd=%v",
			m.composeProcessListViewModel.selectedContainer, cmd)

		assert.Equal(t, 1, m.composeProcessListViewModel.selectedContainer)
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
	hasLog := false
	for _, cmd := range commands {
		if cmd == "up" {
			hasUp = true
		}
		if cmd == "log" {
			hasLog = true
		}
	}
	assert.True(t, hasUp, "Should have up command")
	assert.True(t, hasLog, "Should have log command")
}

func TestGetCommandSuggestions(t *testing.T) {
	// Create a model and initialize it
	model := NewModel(ComposeProcessListView, "")
	model.initializeKeyHandlers()

	// Test prefix matching
	t.Run("prefix-match", func(t *testing.T) {
		suggestions := model.getCommandSuggestions("lo")
		assert.Greater(t, len(suggestions), 0)
		for _, s := range suggestions {
			assert.True(t, strings.HasPrefix(s, "lo") || strings.Contains(s, "lo"))
		}
	})

	// Test substring matching
	t.Run("substring-match", func(t *testing.T) {
		suggestions := model.getCommandSuggestions("log")
		assert.Greater(t, len(suggestions), 0)
		for _, s := range suggestions {
			assert.Contains(t, s, "log")
		}
	})

	// Test no match
	t.Run("no-match", func(t *testing.T) {
		suggestions := model.getCommandSuggestions("xyz123")
		assert.Empty(t, suggestions)
	})
}
