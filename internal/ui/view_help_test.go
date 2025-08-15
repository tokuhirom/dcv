package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestHelpViewModel_Rendering(t *testing.T) {
	tests := []struct {
		name      string
		viewModel HelpViewModel
		model     *Model
		height    int
		expected  []string
	}{
		{
			name: "displays help table with headers",
			viewModel: HelpViewModel{
				TableViewModel: TableViewModel{Cursor: 0},
				parentView:     ComposeProcessListView,
			},
			model: &Model{
				width:  100,
				Height: 20,
				composeProcessListViewHandlers: []KeyConfig{
					{
						Keys:        []string{"Enter"},
						KeyHandler:  nil,
						Description: "View logs",
					},
					{
						Keys:        []string{"r"},
						KeyHandler:  nil,
						Description: "Refresh",
					},
				},
				globalHandlers: []KeyConfig{
					{
						Keys:        []string{"q", "Esc"},
						KeyHandler:  nil,
						Description: "Quit",
					},
				},
			},
			height: 20,
			expected: []string{
				"Key",
				"Command",
				"Description",
				"Enter",
				"View logs",
				"r",
				"Refresh",
				"q/Esc",
				"Quit",
			},
		},
		{
			name: "handles scrolling",
			viewModel: HelpViewModel{
				TableViewModel: TableViewModel{Cursor: 2},
				parentView:     LogView,
			},
			model: &Model{
				width:  100,
				Height: 20,
				logViewHandlers: []KeyConfig{
					{Keys: []string{"j"}, Description: "Scroll down"},
					{Keys: []string{"k"}, Description: "Scroll up"},
					{Keys: []string{"G"}, Description: "Jump to end"},
					{Keys: []string{"g"}, Description: "Jump to start"},
					{Keys: []string{"/"}, Description: "Search"},
				},
			},
			height:   20,
			expected: []string{"G", "Jump to end"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build rows for table
			tt.viewModel.SetRows(tt.viewModel.buildRows(tt.model), tt.model.ViewHeight())
			result := tt.viewModel.render(tt.model, tt.height)

			for _, expected := range tt.expected {
				assert.Contains(t, result, expected, "Expected to find '%s' in output", expected)
			}
		})
	}
}

func TestHelpViewModel_BuildRows(t *testing.T) {
	t.Run("builds rows from view and global handlers", func(t *testing.T) {
		model := &Model{
			composeProcessListViewHandlers: []KeyConfig{
				{
					Keys:        []string{"Enter"},
					KeyHandler:  nil,
					Description: "View logs",
				},
				{
					Keys:        []string{"d"},
					KeyHandler:  nil,
					Description: "View dind containers",
				},
			},
			globalHandlers: []KeyConfig{
				{
					Keys:        []string{"q"},
					KeyHandler:  nil,
					Description: "Quit",
				},
				{
					Keys:        []string{"?"},
					KeyHandler:  nil,
					Description: "Show help",
				},
			},
		}

		vm := &HelpViewModel{
			parentView: ComposeProcessListView,
		}

		rows := vm.buildRows(model)

		// Check view-specific rows
		assert.Greater(t, len(rows), 0)
		assert.Equal(t, "Enter", rows[0][0])
		assert.Equal(t, "View logs", rows[0][2])
		assert.Equal(t, "d", rows[1][0])
		assert.Equal(t, "View dind containers", rows[1][2])

		// Check separator row
		assert.Equal(t, "", rows[2][0])
		assert.Equal(t, "", rows[2][1])
		assert.Equal(t, "", rows[2][2])

		// Check global rows
		assert.Equal(t, "q", rows[3][0])
		assert.Equal(t, "Quit", rows[3][2])
		assert.Equal(t, "?", rows[4][0])
		assert.Equal(t, "Show help", rows[4][2])
	})

	t.Run("handles multiple keys for single handler", func(t *testing.T) {
		model := &Model{
			logViewHandlers: []KeyConfig{
				{
					Keys:        []string{"q", "Esc"},
					KeyHandler:  nil,
					Description: "Back to previous view",
				},
			},
			globalHandlers: []KeyConfig{},
		}

		vm := &HelpViewModel{
			parentView: LogView,
		}

		rows := vm.buildRows(model)
		assert.Equal(t, "q/Esc", rows[0][0])
		assert.Equal(t, "Back to previous view", rows[0][2])
	})

	t.Run("includes command names when available", func(t *testing.T) {
		handler := func(msg tea.KeyMsg) (tea.Model, tea.Cmd) { return nil, nil }

		model := &Model{
			composeProcessListViewHandlers: []KeyConfig{
				{
					Keys:        []string{"r"},
					KeyHandler:  handler,
					Description: "Refresh",
				},
			},
			globalHandlers: []KeyConfig{},
		}

		vm := &HelpViewModel{
			parentView: ComposeProcessListView,
		}

		rows := vm.buildRows(model)
		assert.Equal(t, "r", rows[0][0])
		// Command name might be an auto-generated name like ":1" for anonymous functions
		// We just check it starts with ":"
		assert.True(t, strings.HasPrefix(rows[0][1], ":") || rows[0][1] == "")
		assert.Equal(t, "Refresh", rows[0][2])
	})
}

func TestHelpViewModel_Navigation(t *testing.T) {
	t.Run("HandleDown scrolls down", func(t *testing.T) {
		model := &Model{
			composeProcessListViewHandlers: []KeyConfig{
				{Keys: []string{"1"}, Description: "Item 1"},
				{Keys: []string{"2"}, Description: "Item 2"},
				{Keys: []string{"3"}, Description: "Item 3"},
			},
			globalHandlers: []KeyConfig{},
		}

		vm := &HelpViewModel{
			TableViewModel: TableViewModel{Cursor: 0},
			parentView:     ComposeProcessListView,
		}

		vm.SetRows(vm.buildRows(model), model.ViewHeight())
		cmd := vm.HandleDown(model)
		assert.Nil(t, cmd)
		assert.Equal(t, 1, vm.Cursor)

		// Test boundary - should not scroll beyond last row
		vm.Cursor = 2 // Last item (3 items total)
		cmd = vm.HandleDown(model)
		assert.Nil(t, cmd)
		assert.Equal(t, 2, vm.Cursor, "Should not scroll beyond last item")
	})

	t.Run("HandleUp scrolls up", func(t *testing.T) {
		vm := &HelpViewModel{
			TableViewModel: TableViewModel{Cursor: 2},
			parentView:     ComposeProcessListView,
		}

		model := &Model{
			Height: 20,
			composeProcessListViewHandlers: []KeyConfig{
				{Keys: []string{"1"}, Description: "Item 1"},
				{Keys: []string{"2"}, Description: "Item 2"},
				{Keys: []string{"3"}, Description: "Item 3"},
			},
			globalHandlers: []KeyConfig{},
		}
		vm.SetRows(vm.buildRows(model), model.ViewHeight())
		cmd := vm.HandleUp(model)
		assert.Nil(t, cmd)
		assert.Equal(t, 1, vm.Cursor)

		// Test boundary
		vm.Cursor = 0
		cmd = vm.HandleUp(model)
		assert.Nil(t, cmd)
		assert.Equal(t, 0, vm.Cursor, "Should not scroll above 0")
	})
}

func TestHelpViewModel_Show(t *testing.T) {
	t.Run("Show switches to help view", func(t *testing.T) {
		model := &Model{
			currentView: ComposeProcessListView,
		}
		vm := &HelpViewModel{
			TableViewModel: TableViewModel{Cursor: 10},
		}

		cmd := vm.Show(model, ComposeProcessListView)
		assert.Nil(t, cmd)
		assert.Equal(t, HelpView, model.currentView)
		assert.Equal(t, ComposeProcessListView, vm.parentView)
		assert.Equal(t, 0, vm.Cursor, "Should reset scroll position")
	})
}

func TestHelpViewModel_HandleBack(t *testing.T) {
	t.Run("HandleBack returns to previous view and resets scroll", func(t *testing.T) {
		model := &Model{
			currentView: HelpView,
			viewHistory: []ViewType{ComposeProcessListView, HelpView},
		}
		vm := &HelpViewModel{
			parentView:     ComposeProcessListView,
			TableViewModel: TableViewModel{Cursor: 5},
		}

		cmd := vm.HandleBack(model)
		assert.Nil(t, cmd)
		assert.Equal(t, ComposeProcessListView, model.currentView)
		assert.Equal(t, 0, vm.Cursor, "Should reset scroll position")
	})
}

func TestKeyHandlersToTableRows(t *testing.T) {
	tests := []struct {
		name        string
		keyHandlers []KeyConfig
		expected    [][]string
	}{
		{
			name: "single key per handler",
			keyHandlers: []KeyConfig{
				{
					Keys:        []string{"Enter"},
					Description: "Select item",
				},
				{
					Keys:        []string{"r"},
					Description: "Refresh",
				},
			},
			expected: [][]string{
				{"Enter", "", "Select item"},
				{"r", "", "Refresh"},
			},
		},
		{
			name: "multiple keys per handler",
			keyHandlers: []KeyConfig{
				{
					Keys:        []string{"q", "Esc", "Ctrl+C"},
					Description: "Exit",
				},
			},
			expected: [][]string{
				{"q/Esc/Ctrl+C", "", "Exit"},
			},
		},
		{
			name:        "empty handlers",
			keyHandlers: []KeyConfig{},
			expected:    [][]string{},
		},
		{
			name: "handler with no keys",
			keyHandlers: []KeyConfig{
				{
					Keys:        []string{},
					Description: "No key",
				},
			},
			expected: [][]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &Model{}

			rows := keyHandlersToTableRows(model, tt.keyHandlers)

			assert.Equal(t, len(tt.expected), len(rows))
			for i, expectedRow := range tt.expected {
				assert.Equal(t, expectedRow[0], rows[i][0], "Key mismatch")
				assert.Equal(t, expectedRow[1], rows[i][1], "Command mismatch")
				assert.Equal(t, expectedRow[2], rows[i][2], "Description mismatch")
			}
		})
	}
}

func TestHelpViewModel_Integration(t *testing.T) {
	t.Run("Complete flow from show to back", func(t *testing.T) {
		// Setup model with handlers
		model := &Model{
			currentView: ComposeProcessListView,
			width:       100,
			Height:      20,
			composeProcessListViewHandlers: []KeyConfig{
				{Keys: []string{"Enter"}, Description: "View logs"},
				{Keys: []string{"r"}, Description: "Refresh"},
				{Keys: []string{"d"}, Description: "View dind"},
			},
			logViewHandlers: []KeyConfig{
				{Keys: []string{"j"}, Description: "Scroll down"},
				{Keys: []string{"k"}, Description: "Scroll up"},
			},
			globalHandlers: []KeyConfig{
				{Keys: []string{"?"}, Description: "Show help"},
				{Keys: []string{"q"}, Description: "Quit"},
			},
		}

		vm := &HelpViewModel{}

		// Show help from compose process list view
		cmd := vm.Show(model, ComposeProcessListView)
		assert.Nil(t, cmd)
		assert.Equal(t, HelpView, model.currentView)
		assert.Equal(t, ComposeProcessListView, vm.parentView)

		// RenderTable and verify content
		rendered := vm.render(model, model.Height)
		assert.Contains(t, rendered, "Key")
		assert.Contains(t, rendered, "Command")
		assert.Contains(t, rendered, "Description")
		assert.Contains(t, rendered, "Enter")
		assert.Contains(t, rendered, "View logs")
		assert.Contains(t, rendered, "?")
		assert.Contains(t, rendered, "Show help")

		// Navigate down
		cmd = vm.HandleDown(model)
		assert.Nil(t, cmd)
		assert.Equal(t, 1, vm.Cursor)

		// Navigate up
		cmd = vm.HandleUp(model)
		assert.Nil(t, cmd)
		assert.Equal(t, 0, vm.Cursor)

		// Go back to previous view
		model.viewHistory = []ViewType{ComposeProcessListView, HelpView}
		cmd = vm.HandleBack(model)
		assert.Nil(t, cmd)
		assert.Equal(t, ComposeProcessListView, model.currentView)
		assert.Equal(t, 0, vm.Cursor)
	})
}
