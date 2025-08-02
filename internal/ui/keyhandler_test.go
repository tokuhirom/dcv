package ui

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderHelpText(t *testing.T) {
	tests := []struct {
		name     string
		configs  []KeyConfig
		width    int
		expected string
		contains []string
	}{
		{
			name:     "empty configs",
			configs:  []KeyConfig{},
			width:    80,
			expected: "",
		},
		{
			name: "single key config",
			configs: []KeyConfig{
				{Keys: []string{"q"}, Description: "quit"},
			},
			width:    80,
			expected: "q:quit",
		},
		{
			name: "multiple key configs that fit on one line",
			configs: []KeyConfig{
				{Keys: []string{"up", "k"}, Description: "move up"},
				{Keys: []string{"down", "j"}, Description: "move down"},
				{Keys: []string{"enter"}, Description: "view logs"},
			},
			width:    80,
			expected: "up:move up | down:move down | enter:view logs",
		},
		{
			name: "very narrow screen",
			configs: []KeyConfig{
				{Keys: []string{"up", "k"}, Description: "move up"},
				{Keys: []string{"down", "j"}, Description: "move down"},
			},
			width:    15,
			expected: "Press q to quit",
		},
		{
			name: "configs with multiple keys uses first key",
			configs: []KeyConfig{
				{Keys: []string{"up", "k", "↑"}, Description: "move up"},
			},
			width:    80,
			expected: "up:move up",
		},
		{
			name: "long help text that needs wrapping",
			configs: []KeyConfig{
				{Keys: []string{"up"}, Description: "move up"},
				{Keys: []string{"down"}, Description: "move down"},
				{Keys: []string{"enter"}, Description: "view logs"},
				{Keys: []string{"d"}, Description: "dind composeContainers"},
				{Keys: []string{"r"}, Description: "refresh"},
			},
			width:    50,
			contains: []string{"up:move up", "down:move down"},
		},
		{
			name: "extremely long single item gets truncated",
			configs: []KeyConfig{
				{Keys: []string{"verylongkeyname"}, Description: "extremely long description that will definitely not fit on any reasonable screen width"},
			},
			width:    30,
			contains: []string{"verylongkeyname:extremely"},
		},
		{
			name: "no keys in config",
			configs: []KeyConfig{
				{Keys: []string{}, Description: "no key"},
			},
			width:    80,
			expected: "",
		},
		{
			name: "realistic process list view config",
			configs: []KeyConfig{
				{Keys: []string{"up", "k"}, Description: "move up"},
				{Keys: []string{"down", "j"}, Description: "move down"},
				{Keys: []string{"enter"}, Description: "view logs"},
				{Keys: []string{"d"}, Description: "dind composeContainers"},
				{Keys: []string{"r"}, Description: "refresh"},
				{Keys: []string{"a"}, Description: "toggle all"},
				{Keys: []string{"s"}, Description: "stats"},
				{Keys: []string{"t"}, Description: "top"},
				{Keys: []string{"K"}, Description: "kill"},
				{Keys: []string{"S"}, Description: "stop"},
				{Keys: []string{"U"}, Description: "start"},
				{Keys: []string{"R"}, Description: "restart"},
				{Keys: []string{"D"}, Description: "remove"},
			},
			width:    120,
			contains: []string{"up:move up", "enter:view logs", "d:dind composeContainers"},
		},
		{
			name: "wrapping at exact boundary",
			configs: []KeyConfig{
				{Keys: []string{"a"}, Description: "aa"},
				{Keys: []string{"b"}, Description: "bb"},
				{Keys: []string{"c"}, Description: "cc"},
			},
			width:    24,                   // Will have 20 chars available (24-4), minimum required
			expected: "a:aa | b:bb | c:cc", // This is 18 chars, fits in 20
		},
		{
			name: "wrapping when exceeds width",
			configs: []KeyConfig{
				{Keys: []string{"up"}, Description: "move up"},
				{Keys: []string{"down"}, Description: "move down"},
				{Keys: []string{"enter"}, Description: "select"},
			},
			width: 30, // Will have 26 chars available (30-4)
			// "up:move up | down:move down | enter:select" is 43 chars, won't fit
			// First line should be "up:move up | down:move down" (28 chars, too long)
			// So it will be just "up:move up"
			expected: "up:move up",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderHelpText(tt.configs, tt.width)

			if tt.expected != "" {
				assert.Equal(t, tt.expected, result)
			}

			for _, substring := range tt.contains {
				assert.Contains(t, result, substring)
			}
		})
	}
}

func TestGetHelpText(t *testing.T) {
	m := NewModel(ComposeProcessListView, "")
	m.Init() // Initialize key handlers
	m.width = 80

	tests := []struct {
		name    string
		view    ViewType
		hasHelp bool
	}{
		{
			name:    "process list view",
			view:    ComposeProcessListView,
			hasHelp: true,
		},
		{
			name:    "log view",
			view:    LogView,
			hasHelp: true,
		},
		{
			name:    "dind process list view",
			view:    DindComposeProcessListView,
			hasHelp: true,
		},
		{
			name:    "top view",
			view:    TopView,
			hasHelp: true,
		},
		{
			name:    "stats view",
			view:    StatsView,
			hasHelp: true,
		},
		{
			name:    "project list view",
			view:    ProjectListView,
			hasHelp: true,
		},
		{
			name:    "invalid view type",
			view:    ViewType(999),
			hasHelp: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.currentView = tt.view
			helpText := m.GetHelpText()

			if tt.hasHelp {
				assert.NotEmpty(t, helpText)
				assert.Contains(t, helpText, ":")
			} else {
				assert.Empty(t, helpText)
			}
		})
	}
}

func TestGetStyledHelpText(t *testing.T) {
	m := NewModel(ComposeProcessListView, "")
	m.Init()
	m.width = 80

	// Test with a view that has help text
	m.currentView = ComposeProcessListView
	styledHelp := m.GetStyledHelpText()
	assert.NotEmpty(t, styledHelp)
	// The styled help should contain the help text content
	plainHelp := m.GetHelpText()
	assert.Contains(t, styledHelp, "up:move up") // Check for actual content
	// The length should be different due to styling (padding adds spaces)
	assert.True(t, len(styledHelp) > len(plainHelp))
	// Should have some padding spaces
	assert.True(t, strings.Contains(styledHelp, " up:move up") || strings.Contains(styledHelp, "up:move up "))

	// Test with no help text
	m.currentView = ViewType(999)
	styledHelp = m.GetStyledHelpText()
	assert.Empty(t, styledHelp)
}
