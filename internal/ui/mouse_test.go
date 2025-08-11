package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestCalculateNavbarZones(t *testing.T) {
	m := NewModel(ComposeProcessListView)
	zones := m.calculateNavbarZones()

	// Check that we have the expected number of items
	assert.Equal(t, 6, len(zones.Items))

	// Check first item (Containers)
	assert.Equal(t, "Containers", zones.Items[0].Label)
	assert.Equal(t, DockerContainerListView, zones.Items[0].ViewType)
	assert.Equal(t, "1", zones.Items[0].Key)
	assert.Equal(t, 0, zones.Items[0].StartX)
	assert.Greater(t, zones.Items[0].EndX, zones.Items[0].StartX)

	// Check that items don't overlap
	for i := 1; i < len(zones.Items); i++ {
		assert.GreaterOrEqual(t, zones.Items[i].StartX, zones.Items[i-1].EndX+3) // +3 for separator
	}

	// Check hide button
	assert.Equal(t, "[H]ide navbar", zones.HideButton.Label)
	assert.Greater(t, zones.HideButton.EndX, zones.HideButton.StartX)
}

func TestHandleNavbarMouseClick(t *testing.T) {
	tests := []struct {
		name         string
		x            int
		y            int
		navbarHidden bool
		expectedView ViewType
		shouldChange bool
	}{
		{
			name:         "click on Containers",
			x:            5, // Somewhere in "[1] Containers"
			y:            0,
			navbarHidden: false,
			expectedView: DockerContainerListView,
			shouldChange: true,
		},
		{
			name:         "click below navbar",
			x:            5,
			y:            1,
			navbarHidden: false,
			expectedView: ComposeProcessListView,
			shouldChange: false,
		},
		{
			name:         "click when navbar hidden",
			x:            5,
			y:            0,
			navbarHidden: true,
			expectedView: ComposeProcessListView,
			shouldChange: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel(ComposeProcessListView)
			m.navbarHidden = tt.navbarHidden
			m.initializeKeyHandlers()

			model, _ := m.handleNavbarMouseClick(tt.x, tt.y)
			resultModel := model.(*Model)

			if tt.shouldChange {
				assert.Equal(t, tt.expectedView, resultModel.currentView)
			} else {
				assert.Equal(t, ComposeProcessListView, resultModel.currentView)
			}
		})
	}
}

func TestHandleNavigationKey(t *testing.T) {
	tests := []struct {
		key          string
		expectedView ViewType
	}{
		{"1", DockerContainerListView},
		{"2", ComposeProjectListView},
		{"3", ImageListView},
		{"4", NetworkListView},
		{"5", VolumeListView},
		{"6", StatsView},
		{"invalid", ComposeProcessListView}, // Should not change
	}

	for _, tt := range tests {
		t.Run("key_"+tt.key, func(t *testing.T) {
			m := NewModel(ComposeProcessListView)
			m.initializeKeyHandlers()

			model, _ := m.handleNavigationKey(tt.key)
			resultModel := model.(*Model)

			if tt.key != "invalid" {
				assert.Equal(t, tt.expectedView, resultModel.currentView)
			} else {
				assert.Equal(t, ComposeProcessListView, resultModel.currentView)
			}
		})
	}
}

func TestIsNavbarClick(t *testing.T) {
	tests := []struct {
		name     string
		msg      tea.MouseMsg
		expected bool
	}{
		{
			name: "left click on navbar",
			msg: tea.MouseMsg{
				X:      10,
				Y:      0,
				Action: tea.MouseActionPress,
				Button: tea.MouseButtonLeft,
			},
			expected: true,
		},
		{
			name: "left click below navbar",
			msg: tea.MouseMsg{
				X:      10,
				Y:      1,
				Action: tea.MouseActionPress,
				Button: tea.MouseButtonLeft,
			},
			expected: false,
		},
		{
			name: "right click on navbar",
			msg: tea.MouseMsg{
				X:      10,
				Y:      0,
				Action: tea.MouseActionPress,
				Button: tea.MouseButtonRight,
			},
			expected: false,
		},
		{
			name: "mouse motion on navbar",
			msg: tea.MouseMsg{
				X:      10,
				Y:      0,
				Action: tea.MouseActionMotion,
				Button: tea.MouseButtonNone,
			},
			expected: false,
		},
		{
			name: "mouse release on navbar",
			msg: tea.MouseMsg{
				X:      10,
				Y:      0,
				Action: tea.MouseActionRelease,
				Button: tea.MouseButtonLeft,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isNavbarClick(tt.msg)
			assert.Equal(t, tt.expected, result)
		})
	}
}
