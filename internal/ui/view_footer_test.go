package ui

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestViewFooter(t *testing.T) {
	t.Run("shows action hint in process list views", func(t *testing.T) {
		model := NewModel(ComposeProcessListView)
		model.Init()
		model.width = 100
		model.Height = 30

		// Test ComposeProcessListView
		model.currentView = ComposeProcessListView
		footer := model.viewFooter()
		assert.Contains(t, footer, "Press ? for help")
		assert.Contains(t, footer, "Press x for actions")

		// Test DockerContainerListView
		model.currentView = DockerContainerListView
		footer = model.viewFooter()
		assert.Contains(t, footer, "Press ? for help")
		assert.Contains(t, footer, "Press x for actions")

		// Test DindProcessListView
		model.currentView = DindProcessListView
		footer = model.viewFooter()
		assert.Contains(t, footer, "Press ? for help")
		assert.Contains(t, footer, "Press x for actions")
	})

	t.Run("does not show action hint in non-process list views", func(t *testing.T) {
		model := NewModel(ComposeProcessListView)
		model.Init()
		model.width = 100
		model.Height = 30

		// Test LogView
		model.currentView = LogView
		footer := model.viewFooter()
		assert.NotContains(t, footer, "Press x for actions")

		// Test StatsView
		model.currentView = StatsView
		footer = model.viewFooter()
		assert.Contains(t, footer, "Press ? for help")
		assert.NotContains(t, footer, "Press x for actions")

		// Test ImageListView
		model.currentView = ImageListView
		footer = model.viewFooter()
		assert.Contains(t, footer, "Press ? for help")
		assert.NotContains(t, footer, "Press x for actions")

		// Test NetworkListView
		model.currentView = NetworkListView
		footer = model.viewFooter()
		assert.Contains(t, footer, "Press ? for help")
		assert.NotContains(t, footer, "Press x for actions")
	})

	t.Run("action hint appears with navbar hint when navbar is hidden", func(t *testing.T) {
		model := NewModel(ComposeProcessListView)
		model.Init()
		model.width = 100
		model.Height = 30

		// Hide navbar and check ComposeProcessListView
		model.navbarHidden = true
		model.currentView = ComposeProcessListView
		footer := model.viewFooter()
		assert.Contains(t, footer, "Press ? for help")
		assert.Contains(t, footer, "Press x for actions")
		assert.Contains(t, footer, "Press H to show navbar")

		// Check order is correct (help | actions | navbar)
		helpIdx := strings.Index(footer, "Press ? for help")
		actionsIdx := strings.Index(footer, "Press x for actions")
		navbarIdx := strings.Index(footer, "Press H to show navbar")
		assert.True(t, helpIdx < actionsIdx)
		assert.True(t, actionsIdx < navbarIdx)
	})

	t.Run("quit confirmation overrides other footer messages", func(t *testing.T) {
		model := NewModel(ComposeProcessListView)
		model.Init()
		model.width = 100
		model.Height = 30

		model.currentView = ComposeProcessListView
		model.quitConfirmation = true
		footer := model.viewFooter()
		assert.Contains(t, footer, "Really quit? (y/n)")
		assert.NotContains(t, footer, "Press ? for help")
		assert.NotContains(t, footer, "Press x for actions")
	})

	t.Run("help view shows back hint", func(t *testing.T) {
		model := NewModel(ComposeProcessListView)
		model.Init()
		model.width = 100
		model.Height = 30

		model.currentView = HelpView
		footer := model.viewFooter()
		assert.Contains(t, footer, "Press ESC or q to go back")
		assert.NotContains(t, footer, "Press ? for help")
		assert.NotContains(t, footer, "Press x for actions")
	})
}
