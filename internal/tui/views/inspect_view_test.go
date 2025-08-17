package views

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"

	"github.com/tokuhirom/dcv/internal/docker"
)

func TestNewInspectView(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewInspectView(dockerClient)

	assert.NotNil(t, view)
	assert.NotNil(t, view.docker)
	assert.NotNil(t, view.textView)
	assert.NotNil(t, view.searchInput)
	assert.Empty(t, view.content)
	assert.Empty(t, view.targetName)
	assert.Empty(t, view.targetType)
	assert.False(t, view.isSearchMode)
}

func TestInspectView_GetPrimitive(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewInspectView(dockerClient)

	// Test normal mode
	primitive := view.GetPrimitive()
	assert.NotNil(t, primitive)
	_, ok := primitive.(*tview.TextView)
	assert.True(t, ok)

	// Test search mode
	view.isSearchMode = true
	primitive = view.GetPrimitive()
	assert.NotNil(t, primitive)
	_, ok = primitive.(*tview.Flex)
	assert.True(t, ok)
}

func TestInspectView_GetTitle(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewInspectView(dockerClient)

	// Test basic title
	view.targetName = "my-container"
	view.targetType = "container"
	title := view.GetTitle()
	assert.Equal(t, "Inspect: my-container (container)", title)

	// Test with search (no matches)
	view.searchText = "test"
	view.searchResults = []int{}
	title = view.GetTitle()
	assert.Contains(t, title, "Search: test (no matches)")

	// Test with search (with matches)
	view.searchResults = []int{1, 3, 5}
	view.currentSearchIdx = 1
	title = view.GetTitle()
	assert.Contains(t, title, "Search: test (2/3)")
}

func TestInspectView_FormatJSON(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewInspectView(dockerClient)

	tests := []struct {
		name     string
		input    string
		wantYAML bool
	}{
		{
			name:     "Simple JSON object",
			input:    `{"name": "test", "value": 123}`,
			wantYAML: true,
		},
		{
			name:     "JSON array with single object",
			input:    `[{"name": "test", "value": 123}]`,
			wantYAML: true,
		},
		{
			name:     "Invalid JSON",
			input:    `{invalid json}`,
			wantYAML: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := view.formatJSON(tt.input)
			if tt.wantYAML {
				assert.NoError(t, err)
				assert.NotEmpty(t, result)
				// Check that it's YAML format (should contain "name:" not "name":")
				assert.Contains(t, result, "name:")
				assert.NotContains(t, result, "\"name\"")
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestInspectView_SetContent(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewInspectView(dockerClient)

	content := `name: test-container
id: abc123
status: running
ports:
  - 80:8080
  - 443:8443`

	view.SetContent(content, "test-container", "container")

	assert.Equal(t, content, view.content)
	assert.Equal(t, "test-container", view.targetName)
	assert.Equal(t, "container", view.targetType)
	assert.NotEmpty(t, view.lines)
	assert.Equal(t, 6, len(view.lines))
}

func TestInspectView_ApplySyntaxHighlighting(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewInspectView(dockerClient)

	tests := []struct {
		name     string
		line     string
		isMatch  bool
		expected string
	}{
		{
			name:     "Key-value pair",
			line:     "name: test",
			isMatch:  false,
			expected: "[blue]name:[white] [white]test[white]",
		},
		{
			name:     "Key-value with string",
			line:     "name: \"test string\"",
			isMatch:  false,
			expected: "[blue]name:[white] [green]\"test string\"[white]",
		},
		{
			name:     "Key-value with number",
			line:     "port: 8080",
			isMatch:  false,
			expected: "[blue]port:[white] [yellow]8080[white]",
		},
		{
			name:     "Key-value with boolean",
			line:     "enabled: true",
			isMatch:  false,
			expected: "[blue]enabled:[white] [red]true[white]",
		},
		{
			name:     "List item",
			line:     "- item1",
			isMatch:  false,
			expected: "[white]- [green]item1[white]",
		},
		{
			name:     "Indented key-value",
			line:     "  name: test",
			isMatch:  false,
			expected: "  [blue]name:[white] [white]test[white]",
		},
		{
			name:     "Key-value with search match",
			line:     "name: test",
			isMatch:  true,
			expected: "[blue]name:[white] [white]test[white]", // Background colors only applied when searchText is set
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := view.applySyntaxHighlighting(tt.line, tt.isMatch)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInspectView_Search(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewInspectView(dockerClient)

	// Create a test app
	app := tview.NewApplication()
	SetApp(app)

	// Set test content
	content := `name: test-container
id: abc123
status: running
image: nginx:latest
ports:
  - 80:8080
  - 443:8443
environment:
  - PATH=/usr/bin
  - HOME=/root`

	view.SetContent(content, "test-container", "container")

	// Test search for "nginx"
	view.performSearch("nginx")
	assert.Equal(t, "nginx", view.searchText)
	assert.NotEmpty(t, view.searchResults)
	assert.Contains(t, view.searchResults, 3) // Line 4 (0-indexed)

	// Test search for "8080"
	view.performSearch("8080")
	assert.Equal(t, "8080", view.searchText)
	assert.NotEmpty(t, view.searchResults)
	assert.Contains(t, view.searchResults, 5) // Line 6

	// Test search with no matches
	view.performSearch("notfound")
	assert.Equal(t, "notfound", view.searchText)
	assert.Empty(t, view.searchResults)

	// Test clear search
	view.clearSearch()
	assert.Empty(t, view.searchText)
	assert.Empty(t, view.searchResults)
	assert.Equal(t, 0, view.currentSearchIdx)
	assert.False(t, view.isSearchMode)
}

func TestInspectView_SearchNavigation(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewInspectView(dockerClient)

	// Set test content with multiple matches
	content := `test: value1
data: test
info: something
test: value2
result: test`

	view.SetContent(content, "test", "container")

	// Search for "test" - should find 4 matches
	view.performSearch("test")
	assert.Len(t, view.searchResults, 4)
	assert.Equal(t, 0, view.currentSearchIdx)

	// Navigate to next result
	view.nextSearchResult()
	assert.Equal(t, 1, view.currentSearchIdx)

	// Navigate to next result again
	view.nextSearchResult()
	assert.Equal(t, 2, view.currentSearchIdx)

	// Navigate to previous result
	view.prevSearchResult()
	assert.Equal(t, 1, view.currentSearchIdx)

	// Wrap around to last result
	view.currentSearchIdx = 0
	view.prevSearchResult()
	assert.Equal(t, 3, view.currentSearchIdx)

	// Wrap around to first result
	view.currentSearchIdx = 3
	view.nextSearchResult()
	assert.Equal(t, 0, view.currentSearchIdx)
}

func TestInspectView_SearchCaseInsensitive(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewInspectView(dockerClient)

	content := `Name: Test
name: test
NAME: TEST`

	view.SetContent(content, "test", "container")

	// Case-sensitive search
	view.searchIgnoreCase = false
	view.performSearch("name")
	assert.Len(t, view.searchResults, 1) // Only "name: test"

	// Case-insensitive search
	view.searchIgnoreCase = true
	view.performSearch("name")
	assert.Len(t, view.searchResults, 3) // All three lines
}

func TestInspectView_SearchRegex(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewInspectView(dockerClient)

	content := `port: 8080
port: 8081
port: 9000
host: localhost`

	view.SetContent(content, "test", "container")

	// Normal search
	view.searchRegex = false
	view.performSearch("808")
	assert.Len(t, view.searchResults, 2) // 8080 and 8081

	// Regex search
	view.searchRegex = true
	view.performSearch("80[0-9]+")
	assert.Len(t, view.searchResults, 2) // 8080 and 8081

	// Invalid regex should fall back to normal search
	view.performSearch("[invalid")
	assert.False(t, view.searchRegex) // Should be disabled due to invalid regex
}

func TestInspectView_LoadInspectData(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewInspectView(dockerClient)

	// Test with unsupported type
	err := view.LoadInspectData("test", "unsupported")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported inspect type")

	// Note: Testing actual Docker commands would require Docker to be running
	// and would be integration tests rather than unit tests
}

func TestInspectView_UpdateDisplay(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewInspectView(dockerClient)

	// Test with empty content
	view.lines = []string{}
	view.updateDisplay()
	text := view.textView.GetText(false)
	assert.Contains(t, text, "No inspect data available")

	// Test with content
	view.lines = []string{"line1", "line2", "line3"}
	view.updateDisplay()
	text = view.textView.GetText(false)
	assert.Contains(t, text, "line1")
	assert.Contains(t, text, "line2")
	assert.Contains(t, text, "line3")
}

func TestInspectView_JSONHandling(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewInspectView(dockerClient)

	// Test with Docker inspect-style JSON array
	jsonData := `[{
		"Id": "abc123",
		"Name": "test-container",
		"State": {
			"Status": "running",
			"Running": true,
			"Paused": false
		},
		"Config": {
			"Image": "nginx:latest",
			"Env": ["PATH=/usr/bin", "HOME=/root"]
		}
	}]`

	formatted, err := view.formatJSON(jsonData)
	assert.NoError(t, err)
	assert.NotEmpty(t, formatted)

	// Should be converted to YAML format
	assert.Contains(t, formatted, "Id:")
	assert.Contains(t, formatted, "Name:")
	assert.Contains(t, formatted, "Status:")
	assert.NotContains(t, formatted, "\"Id\"")
	assert.NotContains(t, formatted, "\"Name\"")
}

func TestInspectView_ComplexJSON(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewInspectView(dockerClient)

	// Test with complex nested structure
	complexJSON := `{
		"name": "test",
		"nested": {
			"level1": {
				"level2": {
					"value": 123
				}
			}
		},
		"array": [1, 2, 3],
		"mixed": [
			{"key": "value1"},
			{"key": "value2"}
		]
	}`

	// Parse to ensure it's valid JSON
	var data interface{}
	err := json.Unmarshal([]byte(complexJSON), &data)
	assert.NoError(t, err)

	formatted, err := view.formatJSON(complexJSON)
	assert.NoError(t, err)
	assert.NotEmpty(t, formatted)

	// Check YAML format
	assert.Contains(t, formatted, "name:")
	assert.Contains(t, formatted, "nested:")
	assert.Contains(t, formatted, "level1:")
	assert.Contains(t, formatted, "array:")
}

func TestInspectView_SearchMode(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewInspectView(dockerClient)

	// Test entering search mode
	assert.False(t, view.isSearchMode)
	view.startSearch()
	assert.True(t, view.isSearchMode)

	// Test that GetPrimitive returns flex in search mode
	primitive := view.GetPrimitive()
	_, ok := primitive.(*tview.Flex)
	assert.True(t, ok)

	// Test exiting search mode
	view.clearSearch()
	assert.False(t, view.isSearchMode)
}

func TestInspectView_LongContent(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewInspectView(dockerClient)

	// Create long content
	var lines []string
	for i := 0; i < 1000; i++ {
		lines = append(lines, fmt.Sprintf("line%d: value%d", i, i))
	}
	content := strings.Join(lines, "\n")

	view.SetContent(content, "test", "container")
	assert.Equal(t, 1000, len(view.lines))

	// Test search in long content
	view.performSearch("line500")
	assert.NotEmpty(t, view.searchResults)
	assert.Contains(t, view.searchResults, 500)
}

func TestInspectView_EmptySearch(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewInspectView(dockerClient)

	view.SetContent("test content", "test", "container")

	// Perform empty search
	view.performSearch("")
	assert.Empty(t, view.searchText)
	assert.Empty(t, view.searchResults)
}
