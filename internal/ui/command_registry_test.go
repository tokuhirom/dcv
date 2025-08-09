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
		{"CmdBack", "cmd-back"},
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
	assert.NotNil(t, model.viewCommandRegistry)
	assert.Greater(t, len(model.viewCommandRegistry), 0)

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
			_, exists := model.viewCommandRegistry[ComposeProcessListView][cmd]
			assert.True(t, exists, "Command %s should exist in registry", cmd)
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
