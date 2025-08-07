package ui

import (
	"fmt"
	"strings"
	"testing"

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
				logViewModel: LogViewModel{
					logs:       []string{"Line 1", "Line 2", "Line 3"},
					logScrollY: 0,
				},
				width:  100,
				Height: 10,
			},
			height:   10,
			expected: []string{"Line 1", "Line 2", "Line 3"},
		},
		{
			name: "displays empty log message",
			model: &Model{
				logViewModel: LogViewModel{
					logs:       []string{},
					logScrollY: 0,
				},
				width:  100,
				Height: 10,
			},
			height:   10,
			expected: []string{"No logs available"},
		},
		{
			name: "handles scrolling",
			model: &Model{
				logViewModel: LogViewModel{
					logs:       []string{"Line 1", "Line 2", "Line 3", "Line 4", "Line 5"},
					logScrollY: 1,
				},
				width:  100,
				Height: 5,
			},
			height:   3,
			expected: []string{"Line 2", "Line 3", "Line 4"},
		},
		{
			name: "shows loading message",
			model: &Model{
				logViewModel: LogViewModel{
					logs: []string{},
				},
				loading: true,
				width:   100,
				Height:  10,
			},
			height:   10,
			expected: []string{"Loading logs..."},
		},
		{
			name: "highlights search matches",
			model: &Model{
				logViewModel: LogViewModel{
					logs:       []string{"This is an error message", "This is info", "Another error here"},
					logScrollY: 0,
					SearchViewModel: SearchViewModel{
						searchText: "error",
					},
				},
				width:  100,
				Height: 10,
			},
			height:   10,
			expected: []string{"error", "info", "error"},
		},
		{
			name: "filters logs when filter is active",
			model: &Model{
				logViewModel: LogViewModel{
					logs:       []string{"Error: something went wrong", "Info: all good", "Error: another issue"},
					logScrollY: 0,
					FilterViewModel: FilterViewModel{
						filterText:   "Error",
						filterMode:   true,
						filteredLogs: []string{"Error: something went wrong", "Error: another issue"},
					},
				},
				width:  100,
				Height: 10,
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
			result := tt.model.logViewModel.render(tt.model, tt.height)

			for _, expected := range tt.expected {
				assert.Contains(t, result, expected, "Expected to find '%s' in output", expected)
			}
		})
	}
}

func TestLogView_Navigation(t *testing.T) {
	t.Run("HandleDown moves down one line", func(t *testing.T) {
		model := &Model{
			logViewModel: LogViewModel{
				logs:       []string{"Line 1", "Line 2", "Line 3", "Line 4", "Line 5", "Line 6", "Line 7", "Line 8", "Line 9", "Line 10"},
				logScrollY: 0,
			},
			Height: 8, // 8 - 4 = 4 visible lines, maxScroll = 10 - 4 = 6
		}

		cmd := model.logViewModel.HandleDown(model)
		assert.Nil(t, cmd)
		assert.Equal(t, 1, model.logViewModel.logScrollY)

		// Test boundary - should not scroll beyond maxScroll
		model.logViewModel.logScrollY = 5
		cmd = model.logViewModel.HandleDown(model)
		assert.Nil(t, cmd)
		assert.Equal(t, 6, model.logViewModel.logScrollY) // Can scroll to maxScroll

		// Should not scroll beyond maxScroll
		cmd = model.logViewModel.HandleDown(model)
		assert.Nil(t, cmd)
		assert.Equal(t, 6, model.logViewModel.logScrollY)
	})

	t.Run("HandleUp moves up one line", func(t *testing.T) {
		model := &Model{
			logViewModel: LogViewModel{
				logs:       []string{"Line 1", "Line 2", "Line 3"},
				logScrollY: 2,
			},
			Height: 10,
		}

		cmd := model.logViewModel.HandleUp()
		assert.Nil(t, cmd)
		assert.Equal(t, 1, model.logViewModel.logScrollY)

		// Test boundary
		model.logViewModel.logScrollY = 0
		cmd = model.logViewModel.HandleUp()
		assert.Nil(t, cmd)
		assert.Equal(t, 0, model.logViewModel.logScrollY, "Should not scroll above 0")
	})

	t.Run("HandleGoToStart jumps to beginning", func(t *testing.T) {
		model := &Model{
			logViewModel: LogViewModel{
				logScrollY: 100,
			},
		}

		cmd := model.logViewModel.HandleGoToStart()
		assert.Nil(t, cmd)
		assert.Equal(t, 0, model.logViewModel.logScrollY)
	})

	t.Run("HandleGoToEnd jumps to last line", func(t *testing.T) {
		model := &Model{
			logViewModel: LogViewModel{
				logs:       []string{"Line 1", "Line 2", "Line 3", "Line 4", "Line 5"},
				logScrollY: 0,
			},
			Height: 7, // View Height (7 - 4 = 3 visible lines)
		}

		cmd := model.logViewModel.HandleGoToEnd(model)
		assert.Nil(t, cmd)
		// Should position so last line is visible
		// maxScroll = 5 logs - (7 Height - 4 ui) = 5 - 3 = 2
		assert.Equal(t, 2, model.logViewModel.logScrollY)
	})

	t.Run("HandleBack returns to process list", func(t *testing.T) {
		model := &Model{
			currentView: LogView,
			viewHistory: ComposeProcessListView,
			logViewModel: LogViewModel{
				isDindLog: false,
			},
		}

		cmd := model.logViewModel.HandleBack(model)
		assert.Nil(t, cmd) // Now returns nil as it just switches view
		assert.Equal(t, ComposeProcessListView, model.currentView)
	})

	t.Run("HandleBack returns to dind list for dind logs", func(t *testing.T) {
		model := &Model{
			currentView: LogView,
			viewHistory: DindProcessListView,
			logViewModel: LogViewModel{
				isDindLog: true,
			},
			dindProcessListViewModel: DindProcessListViewModel{
				currentDindContainerID: "dind-container",
			},
		}

		cmd := model.logViewModel.HandleBack(model)
		assert.Nil(t, cmd) // Now returns nil as it just switches view
		assert.Equal(t, DindProcessListView, model.currentView)
	})
}

func TestLogView_Search(t *testing.T) {
	t.Run("search finds matches", func(t *testing.T) {
		model := &Model{
			logViewModel: LogViewModel{
				logs:       []string{"Error on line 1", "Info on line 2", "Error on line 3"},
				logScrollY: 0,
				SearchViewModel: SearchViewModel{
					searchText: "Error",
				},
			},
			width:  100,
			Height: 10,
		}

		// Verify search would highlight Error
		result := model.logViewModel.render(model, 10)
		assert.Contains(t, result, "Error")
	})

	t.Run("HandleNextSearchResult navigates to next result", func(t *testing.T) {
		model := &Model{
			logViewModel: LogViewModel{
				logs: []string{"Line 1", "Error here", "Line 3", "Error there"},
				SearchViewModel: SearchViewModel{
					searchText:       "Error",
					searchResults:    []int{1, 3}, // Lines with matches
					currentSearchIdx: 0,
				},
				logScrollY: 0,
			},
			Height: 9, // 9 - 4 = 5 visible lines
		}

		cmd := model.logViewModel.HandleNextSearchResult(model)
		assert.Nil(t, cmd)
		assert.Equal(t, 1, model.logViewModel.currentSearchIdx)
		// Should scroll to center line 3: targetLine - Height/2 + 3 = 3 - 9/2 + 3 = 3 - 4 + 3 = 2
		assert.Equal(t, 2, model.logViewModel.logScrollY)

		// Wrap around
		cmd = model.logViewModel.HandleNextSearchResult(model)
		assert.Nil(t, cmd)
		assert.Equal(t, 0, model.logViewModel.currentSearchIdx)
	})

	t.Run("HandlePrevSearchResult navigates to previous result", func(t *testing.T) {
		model := &Model{
			logViewModel: LogViewModel{
				logs: []string{"Line 1", "Error here", "Line 3", "Error there"},
				SearchViewModel: SearchViewModel{
					searchText:       "Error",
					searchResults:    []int{1, 3},
					currentSearchIdx: 1,
				},
				logScrollY: 0,
			},
			Height: 9,
		}

		cmd := model.logViewModel.HandlePrevSearchResult(model)
		assert.Nil(t, cmd)
		assert.Equal(t, 0, model.logViewModel.currentSearchIdx)
		assert.Equal(t, 0, model.logViewModel.logScrollY)
	})
}

func TestLogView_AutoScroll(t *testing.T) {
	t.Run("auto-scrolls to bottom when new logs arrive", func(t *testing.T) {
		model := &Model{
			logViewModel: LogViewModel{
				logs:       []string{"Line 1", "Line 2", "Line 3"},
				logScrollY: 0,
			},
			Height:      10,
			currentView: LogView,
		}

		// Simulate receiving new log lines
		msg := logLinesMsg{
			lines: []string{"Line 4", "Line 5", "Line 6", "Line 7", "Line 8", "Line 9", "Line 10"},
		}

		newModel, _ := model.Update(msg)
		m := newModel.(*Model)

		// Should auto-scroll to bottom
		// 10 total lines - (10 Height - 4) = 10 - 6 = 4
		assert.Equal(t, 4, m.logViewModel.logScrollY)
	})
}

func TestLogViewModel_ShowMethods(t *testing.T) {
	t.Run("StreamComposeLogs sets up log view", func(t *testing.T) {
		model := &Model{
			currentView: ComposeProcessListView,
		}
		process := models.ComposeContainer{
			ID:   "container1",
			Name: "test-container",
		}

		cmd := model.logViewModel.StreamComposeLogs(model, process)

		assert.Equal(t, LogView, model.currentView)
		assert.Equal(t, "test-container", model.logViewModel.containerName)
		assert.False(t, model.logViewModel.isDindLog)
		assert.Equal(t, 0, model.logViewModel.logScrollY)
		assert.Equal(t, 0, len(model.logViewModel.logs))
		assert.NotNil(t, cmd)
	})

	t.Run("StreamLogsDind sets up dind log view", func(t *testing.T) {
		model := &Model{
			currentView: DindProcessListView,
			dindProcessListViewModel: DindProcessListViewModel{
				currentDindHost: "host-container",
			},
		}
		container := models.DockerContainer{
			ID:    "docker1",
			Names: "/docker-test",
		}

		cmd := model.logViewModel.StreamLogsDind(model, "dind-container-id", container)

		assert.Equal(t, LogView, model.currentView)
		assert.Equal(t, "/docker-test", model.logViewModel.containerName)
		assert.Equal(t, "host-container", model.logViewModel.hostContainer)
		assert.True(t, model.logViewModel.isDindLog)
		assert.Equal(t, 0, model.logViewModel.logScrollY)
		assert.NotNil(t, cmd)
	})

	t.Run("SwitchToLogView resets log view state", func(t *testing.T) {
		model := &Model{
			currentView: ComposeProcessListView,
			logViewModel: LogViewModel{
				logs:          []string{"old", "logs"},
				logScrollY:    5,
				isDindLog:     true,
				containerName: "old-container",
			},
		}

		model.logViewModel.SwitchToLogView(model, "new-container")

		assert.Equal(t, LogView, model.currentView)
		assert.Equal(t, "new-container", model.logViewModel.containerName)
		assert.False(t, model.logViewModel.isDindLog)
		assert.Equal(t, 0, model.logViewModel.logScrollY)
		assert.Equal(t, 0, len(model.logViewModel.logs))
	})
}

func TestLogView_Update(t *testing.T) {
	t.Run("handles logLinesMsg success", func(t *testing.T) {
		model := &Model{
			currentView: LogView,
			loading:     true,
			Height:      10,
			logViewModel: LogViewModel{
				logScrollY: 0,
				logs:       []string{},
			},
		}

		lines := []string{"Line 1", "Line 2", "Line 3", "Line 4", "Line 5", "Line 6", "Line 7", "Line 8", "Line 9", "Line 10", "Line 11", "Line 12"}
		msg := logLinesMsg{
			lines: lines,
		}

		newModel, cmd := model.Update(msg)
		m := newModel.(*Model)

		assert.True(t, m.loading) // logLinesMsg doesn't change loading state
		assert.Nil(t, m.err)
		assert.Equal(t, 12, len(m.logViewModel.logs))
		// Should auto-scroll to end
		// maxScroll = 12 lines - (10 Height - 4) = 12 - 6 = 6
		assert.Equal(t, 6, m.logViewModel.logScrollY)
		assert.NotNil(t, cmd) // Should return pollLogsContinue command
	})

	t.Run("handles logLineMsg", func(t *testing.T) {
		model := &Model{
			currentView: LogView,
			Height:      10,
			logViewModel: LogViewModel{
				logScrollY: 0,
				logs:       []string{"Line 1", "Line 2"},
			},
		}

		msg := logLineMsg{
			line: "Line 3",
		}

		newModel, cmd := model.Update(msg)
		m := newModel.(*Model)

		assert.Equal(t, 3, len(m.logViewModel.logs))
		assert.Equal(t, "Line 3", m.logViewModel.logs[2])
		assert.Nil(t, cmd) // logLineMsg doesn't continue polling
	})
}

func TestLogView_FilterMode(t *testing.T) {
	t.Run("filters logs based on filterText", func(t *testing.T) {
		model := &Model{
			logViewModel: LogViewModel{
				logs:       []string{"ERROR: Database connection failed", "INFO: Server started", "ERROR: Invalid request", "DEBUG: Processing"},
				logScrollY: 0,
				FilterViewModel: FilterViewModel{
					filterText:   "ERROR",
					filterMode:   true,
					filteredLogs: []string{"ERROR: Database connection failed", "ERROR: Invalid request"},
				},
			},
			width:  100,
			Height: 10,
		}

		result := model.logViewModel.render(model, 10)

		// Should only show ERROR lines
		assert.Contains(t, result, "ERROR: Database connection failed")
		assert.Contains(t, result, "ERROR: Invalid request")
		assert.NotContains(t, result, "INFO: Server started")
		assert.NotContains(t, result, "DEBUG: Processing")
	})

	t.Run("shows no match message", func(t *testing.T) {
		model := &Model{
			logViewModel: LogViewModel{
				logs:       []string{"Info: something", "Debug: another"},
				logScrollY: 0,
				FilterViewModel: FilterViewModel{
					filterText:   "error",
					filterMode:   true,
					filteredLogs: []string{},
				},
			},
			width:  100,
			Height: 10,
		}

		result := model.logViewModel.render(model, 10)

		assert.Contains(t, result, "No logs match the filter")
	})
}

func TestLogView_EmptyLogs(t *testing.T) {
	t.Run("handles empty logs gracefully", func(t *testing.T) {
		model := &Model{
			logViewModel: LogViewModel{
				logs:       []string{},
				logScrollY: 0,
			},
			width:  100,
			Height: 10,
		}

		// Test all navigation operations
		cmd := model.logViewModel.HandleDown(model)
		assert.Nil(t, cmd)
		assert.Equal(t, 0, model.logViewModel.logScrollY)

		cmd = model.logViewModel.HandleUp()
		assert.Nil(t, cmd)
		assert.Equal(t, 0, model.logViewModel.logScrollY)

		cmd = model.logViewModel.HandleGoToEnd(model)
		assert.Nil(t, cmd)
		assert.Equal(t, 0, model.logViewModel.logScrollY)

		cmd = model.logViewModel.HandleGoToStart()
		assert.Nil(t, cmd)
		assert.Equal(t, 0, model.logViewModel.logScrollY)

		// Should show empty message
		result := model.logViewModel.render(model, 10)
		assert.Contains(t, result, "No logs available")
	})
}

func TestLogView_ScrollIndicator(t *testing.T) {
	t.Run("shows scroll indicator when logs exceed Height", func(t *testing.T) {
		logs := make([]string, 20)
		for i := 0; i < 20; i++ {
			logs[i] = fmt.Sprintf("Log line %d", i+1)
		}

		model := &Model{
			logViewModel: LogViewModel{
				logs:       logs,
				logScrollY: 5,
			},
			width:  100,
			Height: 10, // 10 - 4 = 6 visible lines
		}

		result := model.logViewModel.render(model, 6)

		// Should show scroll indicator: [6-11/20]
		assert.Contains(t, result, "[6-11/20]")
	})

	t.Run("shows filtered count in scroll indicator", func(t *testing.T) {
		model := &Model{
			logViewModel: LogViewModel{
				logs:       []string{"Error 1", "Info 1", "Error 2", "Info 2"},
				logScrollY: 0,
				FilterViewModel: FilterViewModel{
					filteredLogs: []string{"Error 1", "Error 2"},
					filterMode:   true,
					filterText:   "Error",
				},
			},
			width:  100,
			Height: 5, // 5 - 4 = 1 visible line
		}

		result := model.logViewModel.render(model, 1)

		// Should show filtered indicator
		assert.Contains(t, result, "(filtered from 4)")
	})
}

func TestLogView_SearchHighlighting(t *testing.T) {
	t.Run("highlights search text", func(t *testing.T) {
		model := &Model{
			logViewModel: LogViewModel{
				logs:       []string{"This contains ERROR in the text"},
				logScrollY: 0,
				SearchViewModel: SearchViewModel{
					searchText: "ERROR",
					searchMode: false,
				},
				FilterViewModel: FilterViewModel{
					filterMode: false,
				},
			},
			width:  100,
			Height: 10,
		}

		result := model.logViewModel.render(model, 10)

		// Should contain the line with ERROR
		assert.Contains(t, result, "ERROR")
	})

	t.Run("marks current search result", func(t *testing.T) {
		model := &Model{
			logViewModel: LogViewModel{
				logs:       []string{"Line 1", "Error here", "Line 3"},
				logScrollY: 0,
				SearchViewModel: SearchViewModel{
					searchText:       "Error",
					searchResults:    []int{1},
					currentSearchIdx: 0,
					searchMode:       false,
				},
				FilterViewModel: FilterViewModel{
					filterMode: false,
				},
			},
			width:  100,
			Height: 10,
		}

		result := model.logViewModel.render(model, 10)

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
			Height:      20,
		}

		// Show logs for a container
		process := models.ComposeContainer{
			ID:   "test-container",
			Name: "my-app",
		}
		cmd := model.logViewModel.StreamComposeLogs(model, process)
		assert.NotNil(t, cmd)
		assert.Equal(t, LogView, model.currentView)
		assert.Equal(t, "my-app", model.logViewModel.containerName)

		// Simulate receiving logs
		msg := logLinesMsg{
			lines: []string{"Log 1", "Log 2", "Log 3"},
		}
		_, _ = model.Update(msg)
		assert.Equal(t, 3, len(model.logViewModel.logs))

		// Go back to process list
		_ = model.logViewModel.HandleBack(model)
		assert.Equal(t, ComposeProcessListView, model.currentView)
	})
}
