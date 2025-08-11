package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPSCommandRegistration(t *testing.T) {
	// Create a model and initialize it
	model := NewModel(ComposeProcessListView)
	model.initializeKeyHandlers()

	t.Run("PS command is registered correctly", func(t *testing.T) {
		// The command name for CmdPS should be "ps" (not "p-s")
		cmdName := model.getCommandForHandler(model.CmdPS)
		assert.Equal(t, "ps", cmdName)
	})

	t.Run("ComposeLS command is registered correctly", func(t *testing.T) {
		// The command name for CmdComposeLS should be "compose-ls" (not "compose-l-s")
		cmdName := model.getCommandForHandler(model.CmdComposeLS)
		assert.Equal(t, "compose-ls", cmdName)
	})

	t.Run("Commands can be executed", func(t *testing.T) {
		// Test that :ps command exists and can be executed
		psExists := false
		composeLsExists := false

		// Check all commands to verify our special cases are registered
		for cmd := range model.allCommands {
			if cmd == "ps" {
				psExists = true
			}
			if cmd == "compose-ls" {
				composeLsExists = true
			}
		}

		assert.True(t, psExists, ":ps command should be registered")
		assert.True(t, composeLsExists, ":compose-ls command should be registered")
	})
}
