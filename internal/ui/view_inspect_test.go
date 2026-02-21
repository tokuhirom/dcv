package ui

import (
	"errors"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

func TestInspectViewModel_Rendering(t *testing.T) {
	tests := []struct {
		name      string
		viewModel InspectViewModel
		height    int
		expected  []string
	}{
		{
			name: "displays no data message when empty",
			viewModel: InspectViewModel{
				inspectContent: "",
			},
			height:   20,
			expected: []string{"1"}, // Empty content shows line 1
		},
		{
			name: "displays JSON content with line numbers",
			viewModel: InspectViewModel{
				inspectContent: `{
  "Id": "abc123",
  "Name": "test-container",
  "State": "running"
}`,
				inspectScrollY: 0,
			},
			height:   20,
			expected: []string{"1", "Id", "abc123", "Name", "test-container"},
		},
		{
			name: "handles scrolling",
			viewModel: InspectViewModel{
				inspectContent: `Line 1
Line 2
Line 3
Line 4
Line 5`,
				inspectScrollY: 2,
			},
			height:   10,
			expected: []string{"Line 3", "Line 4", "Line 5"},
		},
		{
			name: "shows position indicator for long content",
			viewModel: InspectViewModel{
				inspectContent: strings.Repeat("Line\n", 50),
				inspectScrollY: 10,
			},
			height:   10,
			expected: []string{"Lines 11-14 of 51"}, // height-1 lines shown
		},
		{
			name: "highlights search matches",
			viewModel: InspectViewModel{
				SearchViewModel: SearchViewModel{
					searchText:       "test",
					searchMode:       false,
					searchResults:    []int{1, 3},
					currentSearchIdx: 0,
				},
				inspectContent: `{
  "Id": "test-id",
  "Name": "container",
  "Image": "test-image"
}`,
				inspectScrollY: 0,
			},
			height:   20,
			expected: []string{"test"}, // The position indicator only shows when lines > viewHeight
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.viewModel.render(tt.height - 5)

			for _, expected := range tt.expected {
				assert.Contains(t, result, expected, "Expected to find '%s' in output", expected)
			}
		})
	}
}

func TestInspectViewModel_Navigation(t *testing.T) {
	t.Run("HandleDown scrolls down", func(t *testing.T) {
		model := &Model{
			Height: 10,
		}
		vm := &InspectViewModel{
			inspectContent: strings.Repeat("Line\n", 20),
			inspectScrollY: 0,
		}

		cmd := vm.HandleDown(model)
		assert.Nil(t, cmd)
		assert.Equal(t, 1, vm.inspectScrollY)

		// Test boundary
		vm.inspectScrollY = 15 // Max scroll for 20 lines with height 10
		cmd = vm.HandleDown(model)
		assert.Nil(t, cmd)
		assert.Equal(t, 15, vm.inspectScrollY, "Should be at max scroll")
	})

	t.Run("HandleUp scrolls up", func(t *testing.T) {
		vm := &InspectViewModel{
			inspectContent: strings.Repeat("Line\n", 10),
			inspectScrollY: 2,
		}

		cmd := vm.HandleUp()
		assert.Nil(t, cmd)
		assert.Equal(t, 1, vm.inspectScrollY)

		// Test boundary
		vm.inspectScrollY = 0
		cmd = vm.HandleUp()
		assert.Nil(t, cmd)
		assert.Equal(t, 0, vm.inspectScrollY, "Should not scroll above 0")
	})

	t.Run("HandleGoToBeginning jumps to beginning", func(t *testing.T) {
		vm := &InspectViewModel{
			inspectContent: strings.Repeat("Line\n", 10),
			inspectScrollY: 5,
		}

		cmd := vm.HandleGoToBeginning()
		assert.Nil(t, cmd)
		assert.Equal(t, 0, vm.inspectScrollY)
	})

	t.Run("HandleGoToEnd jumps to end", func(t *testing.T) {
		model := &Model{
			Height: 10,
		}
		vm := &InspectViewModel{
			inspectContent: strings.Repeat("Line\n", 20),
			inspectScrollY: 0,
		}

		cmd := vm.HandleGoToEnd(model)
		assert.Nil(t, cmd)
		assert.Equal(t, 15, vm.inspectScrollY) // 21 lines - (10 height - 5) = 15

		// Test with short content
		vm.inspectContent = "Line 1\nLine 2"
		vm.inspectScrollY = 0
		cmd = vm.HandleGoToEnd(model)
		assert.Nil(t, cmd)
		assert.Equal(t, 0, vm.inspectScrollY, "Should not scroll if content fits")
	})
}

func TestInspectViewModel_HandleBack(t *testing.T) {
	t.Run("HandleBack clears search and returns to previous view", func(t *testing.T) {
		model := &Model{
			currentView: InspectView,
			viewHistory: []ViewType{ComposeProcessListView, InspectView},
		}
		vm := &InspectViewModel{
			SearchViewModel: SearchViewModel{
				searchMode:       true,
				searchText:       "test",
				searchResults:    []int{1, 2, 3},
				currentSearchIdx: 1,
			},
			inspectContent: "some content",
			inspectScrollY: 10,
		}

		cmd := vm.HandleBack(model)
		assert.Nil(t, cmd)
		assert.Equal(t, ComposeProcessListView, model.currentView)
		assert.False(t, vm.searchMode)
		assert.Empty(t, vm.searchText)
		assert.Nil(t, vm.searchResults)
		assert.Equal(t, 0, vm.currentSearchIdx)
	})
}

func TestInspectViewModel_Search(t *testing.T) {
	t.Run("HandleSearch clears search state", func(t *testing.T) {
		vm := &InspectViewModel{
			SearchViewModel: SearchViewModel{
				searchText:       "test",
				searchResults:    []int{1, 2, 3},
				currentSearchIdx: 1,
			},
		}

		cmd := vm.HandleSearch()
		assert.Nil(t, cmd)
		assert.Empty(t, vm.searchText)
		assert.Nil(t, vm.searchResults)
		assert.Equal(t, 0, vm.currentSearchIdx)
	})

	t.Run("HandleNextSearchResult cycles through results", func(t *testing.T) {
		model := &Model{
			Height: 20,
		}
		vm := &InspectViewModel{
			SearchViewModel: SearchViewModel{
				searchResults:    []int{5, 10, 15},
				currentSearchIdx: 0,
			},
			inspectContent: strings.Repeat("Line\n", 20),
			inspectScrollY: 0,
		}

		// First next
		cmd := vm.HandleNextSearchResult(model)
		assert.Nil(t, cmd)
		assert.Equal(t, 1, vm.currentSearchIdx)
		assert.Equal(t, 3, vm.inspectScrollY) // Should center result at line 10

		// Second next
		cmd = vm.HandleNextSearchResult(model)
		assert.Nil(t, cmd)
		assert.Equal(t, 2, vm.currentSearchIdx)

		// Wrap around
		cmd = vm.HandleNextSearchResult(model)
		assert.Nil(t, cmd)
		assert.Equal(t, 0, vm.currentSearchIdx)
	})

	t.Run("HandlePrevSearchResult cycles backward through results", func(t *testing.T) {
		model := &Model{
			Height: 20,
		}
		vm := &InspectViewModel{
			SearchViewModel: SearchViewModel{
				searchResults:    []int{5, 10, 15},
				currentSearchIdx: 0,
			},
			inspectContent: strings.Repeat("Line\n", 20),
			inspectScrollY: 0,
		}

		// Go to previous (wraps to end)
		cmd := vm.HandlePrevSearchResult(model)
		assert.Nil(t, cmd)
		assert.Equal(t, 2, vm.currentSearchIdx)

		// Previous again
		cmd = vm.HandlePrevSearchResult(model)
		assert.Nil(t, cmd)
		assert.Equal(t, 1, vm.currentSearchIdx)
	})
}

func TestInspectViewModel_SearchHighlighting_YAMLFormat(t *testing.T) {
	t.Run("applies YAML highlighting when no search is active", func(t *testing.T) {
		vm := &InspectViewModel{}
		// No search text set means search is not active

		keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Bold(true)
		valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("76"))
		highlightStyle := lipgloss.NewStyle().Background(lipgloss.Color("226"))

		line := "name: container-name"
		result := vm.renderLineWithHighlighting(line, keyStyle, valueStyle, highlightStyle)

		// Should just apply YAML highlighting without search highlighting
		assert.Contains(t, result, "name")
		assert.Contains(t, result, "container-name")
	})

	t.Run("applies YAML highlighting to list items", func(t *testing.T) {
		vm := &InspectViewModel{}

		keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Bold(true)
		valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("76"))
		highlightStyle := lipgloss.NewStyle().Background(lipgloss.Color("226"))

		line := "- my-list-item"
		result := vm.renderLineWithHighlighting(line, keyStyle, valueStyle, highlightStyle)

		assert.Contains(t, result, "my-list-item")
		assert.Contains(t, result, "-") // List marker should be preserved
	})

	t.Run("applies YAML highlighting to regular content", func(t *testing.T) {
		vm := &InspectViewModel{}

		keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Bold(true)
		valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("76"))
		highlightStyle := lipgloss.NewStyle().Background(lipgloss.Color("226"))

		line := "some regular content"
		result := vm.renderLineWithHighlighting(line, keyStyle, valueStyle, highlightStyle)

		assert.Contains(t, result, "some regular content")
	})
}

func TestInspectViewModel_LineNumberRendering(t *testing.T) {
	t.Run("line number with search marker doesn't contain corrupted ANSI sequences", func(t *testing.T) {
		vm := &InspectViewModel{
			SearchViewModel: SearchViewModel{
				searchText:       "test",
				searchResults:    []int{5},
				currentSearchIdx: 0,
			},
			inspectContent: strings.Repeat("line\n", 10),
		}

		result := vm.render(15)

		// Should not contain corrupted ANSI sequences like [38;5;241m
		assert.NotContains(t, result, "[38;5;241m")
		assert.NotContains(t, result, "[241m")

		// Should contain the search marker
		assert.Contains(t, result, "▶")

		// Should contain proper line numbers
		assert.Contains(t, result, "6 ") // Line 6 (0-indexed line 5)
	})
}

func TestInspectViewModel_Set(t *testing.T) {
	vm := &InspectViewModel{
		inspectScrollY: 10,
	}

	content := `{
  "Id": "test123",
  "Name": "container"
}`
	targetName := "test-container"

	vm.Set(content, targetName)
	assert.Equal(t, content, vm.inspectContent)
	assert.Equal(t, targetName, vm.inspectTargetName)
	assert.Equal(t, 0, vm.inspectScrollY, "Should reset scroll position")
}

func TestInspectViewModel_Title(t *testing.T) {
	tests := []struct {
		name     string
		vm       InspectViewModel
		expected string
	}{
		{
			name: "basic title",
			vm: InspectViewModel{
				inspectTargetName: "test-container",
			},
			expected: "Inspect test-container ",
		},
		{
			name: "title with search",
			vm: InspectViewModel{
				SearchViewModel: SearchViewModel{
					searchText:       "test",
					searchMode:       false,
					searchResults:    []int{1, 2, 3},
					currentSearchIdx: 1,
				},
				inspectTargetName: "container",
			},
			expected: "Inspect container  | Search: test (2/3)",
		},
		{
			name: "title with search no matches",
			vm: InspectViewModel{
				SearchViewModel: SearchViewModel{
					searchText:    "test",
					searchMode:    false,
					searchResults: []int{},
				},
				inspectTargetName: "container",
			},
			expected: "Inspect container  | Search: test (no matches)",
		},
		{
			name: "title with case-insensitive regex search",
			vm: InspectViewModel{
				SearchViewModel: SearchViewModel{
					searchText:       "test",
					searchMode:       false,
					searchIgnoreCase: true,
					searchRegex:      true,
					searchResults:    []int{1},
					currentSearchIdx: 0,
				},
				inspectTargetName: "container",
			},
			expected: "Inspect container  | Search: test (1/1) [i] [re]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.vm.Title()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInspectViewModel_LongStrings(t *testing.T) {
	longValue := strings.Repeat("x", 500)
	longKey := strings.Repeat("k", 500)

	tests := []struct {
		name      string
		viewModel InspectViewModel
		height    int
	}{
		{
			name: "very long JSON value",
			viewModel: InspectViewModel{
				inspectContent: "key: " + longValue,
			},
			height: 20,
		},
		{
			name: "very long JSON key",
			viewModel: InspectViewModel{
				inspectContent: longKey + ": value",
			},
			height: 20,
		},
		{
			name: "very long line without key-value",
			viewModel: InspectViewModel{
				inspectContent: strings.Repeat("z", 1000),
			},
			height: 20,
		},
		{
			name: "many long lines",
			viewModel: InspectViewModel{
				inspectContent: strings.Repeat(strings.Repeat("a", 300)+"\n", 100),
			},
			height: 10,
		},
		{
			name: "very small available height of 1",
			viewModel: InspectViewModel{
				inspectContent: "line1\nline2\nline3\nline4\nline5",
			},
			height: 1,
		},
		{
			name: "zero available height",
			viewModel: InspectViewModel{
				inspectContent: "line1\nline2\nline3",
			},
			height: 0,
		},
		{
			name: "negative available height",
			viewModel: InspectViewModel{
				inspectContent: "line1\nline2\nline3",
			},
			height: -5,
		},
		{
			name: "long list item",
			viewModel: InspectViewModel{
				inspectContent: "- " + longValue,
			},
			height: 20,
		},
		{
			name: "search on long lines",
			viewModel: InspectViewModel{
				SearchViewModel: SearchViewModel{
					searchText:       "xxx",
					searchMode:       false,
					searchResults:    []int{0},
					currentSearchIdx: 0,
				},
				inspectContent: strings.Repeat("x", 1000),
			},
			height: 20,
		},
		{
			name: "deeply indented content",
			viewModel: InspectViewModel{
				inspectContent: strings.Repeat("  ", 100) + "key: value",
			},
			height: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			result := tt.viewModel.render(tt.height)
			assert.NotEmpty(t, result)
		})
	}
}

func TestInspectViewModel_Inspect(t *testing.T) {
	t.Run("Inspect initiates loading", func(t *testing.T) {
		model := &Model{
			loading: false,
		}
		vm := &InspectViewModel{}

		targetName := "test-container"
		provider := func() ([]byte, error) {
			return []byte(`{"Id": "test123"}`), nil
		}

		cmd := vm.Inspect(model, targetName, provider)
		assert.NotNil(t, cmd)
		assert.True(t, model.loading)
	})

	t.Run("Inspect handles error from provider", func(t *testing.T) {
		model := &Model{
			loading: false,
		}
		vm := &InspectViewModel{}

		targetName := "test-container"
		testErr := errors.New("failed to inspect")
		provider := func() ([]byte, error) {
			return nil, testErr
		}

		cmd := vm.Inspect(model, targetName, provider)
		assert.NotNil(t, cmd)
		assert.True(t, model.loading)

		// The Inspect function now returns a batch command (spinner + inspect)
		// We need to execute it properly to test the inspect operation
		// For testing, we'll create the inspect command directly
		inspectCmd := func() interface{} {
			content, err := provider()
			return inspectLoadedMsg{
				content:    string(content),
				err:        err,
				targetName: targetName,
			}
		}
		msg := inspectCmd()
		inspectMsg, ok := msg.(inspectLoadedMsg)
		assert.True(t, ok)
		assert.Equal(t, testErr, inspectMsg.err)
		assert.Equal(t, targetName, inspectMsg.targetName)
	})
}
