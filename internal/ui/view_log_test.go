package ui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/tokuhirom/dcv/internal/models"
)

func TestLogView_Rendering(t *testing.T) {
	tests := []struct {
		name     string
		model    *Model
		height   int
		expected []string
	}{
		{
			name: "displays logs",
			model: &Model{
				logs:       []string{"Line 1", "Line 2", "Line 3"},
				logScrollY: 0,
				width:      100,
				height:     10,
			},
			height:   10,
			expected: []string{"Line 1", "Line 2", "Line 3"},
		},
		{
			name: "displays empty log message",
			model: &Model{
				logs:       []string{},
				logScrollY: 0,
				width:      100,
				height:     10,
			},
			height:   10,
			expected: []string{"No logs available"},
		},
		{
			name: "handles scrolling",
			model: &Model{
				logs:       []string{"Line 1", "Line 2", "Line 3", "Line 4", "Line 5"},
				logScrollY: 1,
				width:      100,
				height:     5,
			},
			height:   3,
			expected: []string{"Line 2", "Line 3", "Line 4"},
		},
		{
			name: "shows loading message",
			model: &Model{
				logs:    []string{},
				loading: true,
				width:   100,
				height:  10,
			},
			height:   10,
			expected: []string{"Loading logs..."},
		},
		{
			name: "highlights search matches",
			model: &Model{
				logs:       []string{"This is an error message", "This is info", "Another error here"},
				searchText: "error",
				logScrollY: 0,
				width:      100,
				height:     10,
			},
			height:   10,
			expected: []string{"error", "info", "error"},
		},
		{
			name: "filters logs when filter is active",
			model: &Model{
				logs:         []string{"Error: something went wrong", "Info: all good", "Error: another issue"},
				filterText:   "Error",
				filterMode:   true,
				filteredLogs: []string{"Error: something went wrong", "Error: another issue"},
				logScrollY:   0,
				width:        100,
				height:       10,
			},
			height: 10,
			expected: []string{
				"Error: something went wrong",
				"Error: another issue",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.model.renderLogView(tt.height)

			for _, expected := range tt.expected {
				assert.Contains(t, result, expected, "Expected to find '%s' in output", expected)
			}
		})
	}
}

func TestLogView_Navigation(t *testing.T) {
	t.Run("ScrollLogDown moves down one line", func(t *testing.T) {
		model := &Model{
			logs:       []string{"Line 1", "Line 2", "Line 3", "Line 4", "Line 5", "Line 6", "Line 7", "Line 8", "Line 9", "Line 10"},
			logScrollY: 0,
			height:     8, // 8 - 4 = 4 visible lines, maxScroll = 10 - 4 = 6
		}

		_, cmd := model.ScrollLogDown(tea.KeyMsg{})
		assert.Nil(t, cmd)
		assert.Equal(t, 1, model.logScrollY)

		// Test boundary - should not scroll beyond maxScroll
		model.logScrollY = 5
		_, cmd = model.ScrollLogDown(tea.KeyMsg{})
		assert.Nil(t, cmd)
		assert.Equal(t, 6, model.logScrollY) // Can scroll to maxScroll

		// Should not scroll beyond maxScroll
		_, cmd = model.ScrollLogDown(tea.KeyMsg{})
		assert.Nil(t, cmd)
		assert.Equal(t, 6, model.logScrollY)
	})

	t.Run("ScrollLogUp moves up one line", func(t *testing.T) {
		model := &Model{
			logs:       []string{"Line 1", "Line 2", "Line 3"},
			logScrollY: 2,
			height:     10,
		}

		_, cmd := model.ScrollLogUp(tea.KeyMsg{})
		assert.Nil(t, cmd)
		assert.Equal(t, 1, model.logScrollY)

		// Test boundary
		model.logScrollY = 0
		_, cmd = model.ScrollLogUp(tea.KeyMsg{})
		assert.Nil(t, cmd)
		assert.Equal(t, 0, model.logScrollY, "Should not scroll above 0")
	})

	t.Run("GoToLogStart jumps to beginning", func(t *testing.T) {
		model := &Model{
			logScrollY: 100,
		}

		_, cmd := model.GoToLogStart(tea.KeyMsg{})
		assert.Nil(t, cmd)
		assert.Equal(t, 0, model.logScrollY)
	})

	t.Run("GoToLogEnd jumps to last line", func(t *testing.T) {
		model := &Model{
			logs:       []string{"Line 1", "Line 2", "Line 3", "Line 4", "Line 5"},
			logScrollY: 0,
			height:     7, // View height (7 - 4 = 3 visible lines)
		}

		_, cmd := model.GoToLogEnd(tea.KeyMsg{})
		assert.Nil(t, cmd)
		// Should position so last line is visible
		// maxScroll = 5 logs - (7 height - 4 ui) = 5 - 3 = 2
		assert.Equal(t, 2, model.logScrollY)
	})

	t.Run("BackFromLogView returns to process list", func(t *testing.T) {
		model := &Model{
			currentView: LogView,
			isDindLog:   false,
		}

		_, cmd := model.BackFromLogView(tea.KeyMsg{})
		assert.NotNil(t, cmd) // Returns loadProcesses command
		assert.Equal(t, ComposeProcessListView, model.currentView)
	})

	t.Run("BackFromLogView returns to dind list for dind logs", func(t *testing.T) {
		model := &Model{
			currentView: LogView,
			isDindLog:   true,
			dindProcessListViewModel: DindProcessListViewModel{
				currentDindContainerID: "dind-container",
			},
		}

		_, cmd := model.BackFromLogView(tea.KeyMsg{})
		assert.NotNil(t, cmd) // Returns loadDindContainers command
		assert.Equal(t, DindProcessListView, model.currentView)
	})
}

func TestLogView_Search(t *testing.T) {
	t.Run("search finds matches", func(t *testing.T) {
		model := &Model{
			logs:       []string{"Error on line 1", "Info on line 2", "Error on line 3"},
			searchText: "Error",
			logScrollY: 0,
			width:      100,
			height:     10,
		}

		// Verify search would highlight Error
		result := model.renderLogView(10)
		assert.Contains(t, result, "Error")
	})

	t.Run("NextSearchResult navigates to next result", func(t *testing.T) {
		model := &Model{
			logs:             []string{"Line 1", "Error here", "Line 3", "Error there"},
			searchText:       "Error",
			searchResults:    []int{1, 3}, // Lines with matches
			currentSearchIdx: 0,
			logScrollY:       0,
			height:           9, // 9 - 4 = 5 visible lines
		}

		_, cmd := model.NextSearchResult(tea.KeyMsg{})
		assert.Nil(t, cmd)
		assert.Equal(t, 1, model.currentSearchIdx)
		// Should scroll to center line 3: targetLine - height/2 + 3 = 3 - 9/2 + 3 = 3 - 4 + 3 = 2
		assert.Equal(t, 2, model.logScrollY)

		// Wrap around
		_, cmd = model.NextSearchResult(tea.KeyMsg{})
		assert.Nil(t, cmd)
		assert.Equal(t, 0, model.currentSearchIdx)
	})

	t.Run("PrevSearchResult navigates to previous result", func(t *testing.T) {
		model := &Model{
			logs:             []string{"Line 1", "Error here", "Line 3", "Error there"},
			searchText:       "Error",
			searchResults:    []int{1, 3},
			currentSearchIdx: 1,
			logScrollY:       0,
			height:           9,
		}

		_, cmd := model.PrevSearchResult(tea.KeyMsg{})
		assert.Nil(t, cmd)
		assert.Equal(t, 0, model.currentSearchIdx)
		assert.Equal(t, 0, model.logScrollY)
	})
}

func TestLogView_AutoScroll(t *testing.T) {
	t.Run("auto-scrolls to bottom when new logs arrive", func(t *testing.T) {
		model := &Model{
			logs:        []string{"Line 1", "Line 2", "Line 3"},
			logScrollY:  0,
			height:      10,
			currentView: LogView,
		}

		// Simulate receiving new log lines
		msg := logLinesMsg{
			lines: []string{"Line 4", "Line 5", "Line 6", "Line 7", "Line 8", "Line 9", "Line 10"},
		}

		newModel, _ := model.Update(msg)
		m := newModel.(*Model)

		// Should auto-scroll to bottom
		// 10 total lines - (10 height - 4) = 10 - 6 = 4
		assert.Equal(t, 4, m.logScrollY)
	})
}

func TestLogViewModel_ShowMethods(t *testing.T) {
	t.Run("StreamLogs sets up log view", func(t *testing.T) {
		model := &Model{
			currentView: ComposeProcessListView,
		}
		vm := &LogViewModel{}
		process := models.ComposeContainer{
			ID:   "container1",
			Name: "test-container",
		}

		cmd := vm.StreamLogs(model, process, false, "")

		assert.Equal(t, LogView, model.currentView)
		assert.Equal(t, "test-container", model.containerName)
		assert.False(t, model.isDindLog)
		assert.Equal(t, 0, model.logScrollY)
		assert.Equal(t, 0, len(model.logs))
		assert.NotNil(t, cmd)
	})

	t.Run("ShowDindLog sets up dind log view", func(t *testing.T) {
		model := &Model{
			currentView: DindProcessListView,
			dindProcessListViewModel: DindProcessListViewModel{
				currentDindHost: "host-container",
			},
		}
		vm := &LogViewModel{}
		container := models.DockerContainer{
			ID:    "docker1",
			Names: "/docker-test",
		}

		cmd := vm.ShowDindLog(model, "dind-container-id", container)

		assert.Equal(t, LogView, model.currentView)
		assert.Equal(t, "/docker-test", model.containerName)
		assert.Equal(t, "host-container", model.hostContainer)
		assert.True(t, model.isDindLog)
		assert.Equal(t, 0, model.logScrollY)
		assert.NotNil(t, cmd)
	})

	t.Run("Clear resets log view state", func(t *testing.T) {
		model := &Model{
			currentView:   ComposeProcessListView,
			logs:          []string{"old", "logs"},
			logScrollY:    5,
			isDindLog:     true,
			containerName: "old-container",
		}
		vm := &LogViewModel{}

		vm.Clear(model, "new-container")

		assert.Equal(t, LogView, model.currentView)
		assert.Equal(t, "new-container", model.containerName)
		assert.False(t, model.isDindLog)
		assert.Equal(t, 0, model.logScrollY)
		assert.Equal(t, 0, len(model.logs))
	})
}

func TestLogView_KeyHandlers(t *testing.T) {
	model := NewModel(LogView, "")
	model.initializeKeyHandlers()

	// Verify key handlers are registered
	handlers := model.logViewHandlers
	assert.Greater(t, len(handlers), 0, "Should have registered key handlers")

	// Check specific handlers exist
	expectedKeys := []string{"up", "down", "g", "G", "/", "f", "n", "N", "esc", "?"}
	registeredKeys := make(map[string]bool)

	for _, h := range handlers {
		for _, key := range h.Keys {
			registeredKeys[key] = true
		}
	}

	for _, key := range expectedKeys {
		assert.True(t, registeredKeys[key], "Key %s should be registered", key)
	}
}

func TestLogView_Update(t *testing.T) {
	t.Run("handles logLinesMsg success", func(t *testing.T) {
		model := &Model{
			currentView: LogView,
			loading:     true,
			height:      10,
			logScrollY:  0,
			logs:        []string{},
		}

		lines := []string{"Line 1", "Line 2", "Line 3", "Line 4", "Line 5", "Line 6", "Line 7", "Line 8", "Line 9", "Line 10", "Line 11", "Line 12"}
		msg := logLinesMsg{
			lines: lines,
		}

		newModel, cmd := model.Update(msg)
		m := newModel.(*Model)

		assert.True(t, m.loading) // logLinesMsg doesn't change loading state
		assert.Nil(t, m.err)
		assert.Equal(t, 12, len(m.logs))
		// Should auto-scroll to end
		// maxScroll = 12 lines - (10 height - 4) = 12 - 6 = 6
		assert.Equal(t, 6, m.logScrollY)
		assert.NotNil(t, cmd) // Should return pollLogsContinue command
	})

	t.Run("handles logLineMsg", func(t *testing.T) {
		model := &Model{
			currentView: LogView,
			height:      10,
			logScrollY:  0,
			logs:        []string{"Line 1", "Line 2"},
		}

		msg := logLineMsg{
			line: "Line 3",
		}

		newModel, cmd := model.Update(msg)
		m := newModel.(*Model)

		assert.Equal(t, 3, len(m.logs))
		assert.Equal(t, "Line 3", m.logs[2])
		assert.Nil(t, cmd) // logLineMsg doesn't continue polling
	})
}

func TestLogView_FilterMode(t *testing.T) {
	t.Run("filters logs based on filterText", func(t *testing.T) {
		model := &Model{
			logs:         []string{"ERROR: Database connection failed", "INFO: Server started", "ERROR: Invalid request", "DEBUG: Processing"},
			filterText:   "ERROR",
			filterMode:   true,
			filteredLogs: []string{"ERROR: Database connection failed", "ERROR: Invalid request"},
			logScrollY:   0,
			width:        100,
			height:       10,
		}

		result := model.renderLogView(10)

		// Should only show ERROR lines
		assert.Contains(t, result, "ERROR: Database connection failed")
		assert.Contains(t, result, "ERROR: Invalid request")
		assert.NotContains(t, result, "INFO: Server started")
		assert.NotContains(t, result, "DEBUG: Processing")
	})

	t.Run("shows no match message", func(t *testing.T) {
		model := &Model{
			logs:         []string{"Info: something", "Debug: another"},
			filterText:   "error",
			filterMode:   true,
			filteredLogs: []string{},
			logScrollY:   0,
			width:        100,
			height:       10,
		}

		result := model.renderLogView(10)

		assert.Contains(t, result, "No logs match the filter")
	})
}

func TestLogView_EmptyLogs(t *testing.T) {
	t.Run("handles empty logs gracefully", func(t *testing.T) {
		model := &Model{
			logs:       []string{},
			logScrollY: 0,
			width:      100,
			height:     10,
		}

		// Test all navigation operations
		_, cmd := model.ScrollLogDown(tea.KeyMsg{})
		assert.Nil(t, cmd)
		assert.Equal(t, 0, model.logScrollY)

		_, cmd = model.ScrollLogUp(tea.KeyMsg{})
		assert.Nil(t, cmd)
		assert.Equal(t, 0, model.logScrollY)

		_, cmd = model.GoToLogEnd(tea.KeyMsg{})
		assert.Nil(t, cmd)
		assert.Equal(t, 0, model.logScrollY)

		_, cmd = model.GoToLogStart(tea.KeyMsg{})
		assert.Nil(t, cmd)
		assert.Equal(t, 0, model.logScrollY)

		// Should show empty message
		result := model.renderLogView(10)
		assert.Contains(t, result, "No logs available")
	})
}

func TestLogView_ScrollIndicator(t *testing.T) {
	t.Run("shows scroll indicator when logs exceed height", func(t *testing.T) {
		logs := make([]string, 20)
		for i := 0; i < 20; i++ {
			logs[i] = fmt.Sprintf("Log line %d", i+1)
		}

		model := &Model{
			logs:       logs,
			logScrollY: 5,
			width:      100,
			height:     10, // 10 - 4 = 6 visible lines
		}

		result := model.renderLogView(6)

		// Should show scroll indicator: [6-11/20]
		assert.Contains(t, result, "[6-11/20]")
	})

	t.Run("shows filtered count in scroll indicator", func(t *testing.T) {
		model := &Model{
			logs:         []string{"Error 1", "Info 1", "Error 2", "Info 2"},
			filteredLogs: []string{"Error 1", "Error 2"},
			filterMode:   true,
			filterText:   "Error",
			logScrollY:   0,
			width:        100,
			height:       5, // 5 - 4 = 1 visible line
		}

		result := model.renderLogView(1)

		// Should show filtered indicator
		assert.Contains(t, result, "(filtered from 4)")
	})
}

func TestLogView_SearchHighlighting(t *testing.T) {
	t.Run("highlights search text", func(t *testing.T) {
		model := &Model{
			logs:       []string{"This contains ERROR in the text"},
			searchText: "ERROR",
			searchMode: false,
			filterMode: false,
			logScrollY: 0,
			width:      100,
			height:     10,
		}

		result := model.renderLogView(10)

		// Should contain the line with ERROR
		assert.Contains(t, result, "ERROR")
	})

	t.Run("marks current search result", func(t *testing.T) {
		model := &Model{
			logs:             []string{"Line 1", "Error here", "Line 3"},
			searchText:       "Error",
			searchResults:    []int{1},
			currentSearchIdx: 0,
			searchMode:       false,
			filterMode:       false,
			logScrollY:       0,
			width:            100,
			height:           10,
		}

		result := model.renderLogView(10)

		// Should mark the current search result with >
		lines := strings.Split(result, "\n")
		foundMarker := false
		for _, line := range lines {
			if strings.HasPrefix(line, "> ") && strings.Contains(line, "Error here") {
				foundMarker = true
				break
			}
		}
		assert.True(t, foundMarker, "Should mark current search result with >")
	})
}

func TestLogView_Integration(t *testing.T) {
	t.Run("complete log view workflow", func(t *testing.T) {
		// Start with process list view
		model := &Model{
			currentView: ComposeProcessListView,
			width:       100,
			height:      20,
		}
		vm := &LogViewModel{}

		// Show logs for a container
		process := models.ComposeContainer{
			ID:   "test-container",
			Name: "my-app",
		}
		cmd := vm.StreamLogs(model, process, false, "")
		assert.NotNil(t, cmd)
		assert.Equal(t, LogView, model.currentView)
		assert.Equal(t, "my-app", model.containerName)

		// Simulate receiving logs
		msg := logLinesMsg{
			lines: []string{"Log 1", "Log 2", "Log 3"},
		}
		_, _ = model.Update(msg)
		assert.Equal(t, 3, len(model.logs))

		// Go back to process list
		_, _ = model.BackFromLogView(tea.KeyMsg{})
		assert.Equal(t, ComposeProcessListView, model.currentView)
	})
}
