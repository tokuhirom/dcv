package tui

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	
	"github.com/tokuhirom/dcv/internal/ui"
)

func TestNewState(t *testing.T) {
	state := NewState()
	
	assert.NotNil(t, state)
	assert.NotNil(t, state.ViewHistory)
	assert.Equal(t, 0, len(state.ViewHistory))
	assert.False(t, state.ShowAll)
	assert.Equal(t, "", state.SearchText)
	assert.False(t, state.CommandMode)
	assert.Equal(t, "", state.CommandText)
	assert.Nil(t, state.Error)
	assert.False(t, state.Loading)
}

func TestState_PushView(t *testing.T) {
	tests := []struct {
		name        string
		initialView []ui.ViewType
		pushView    ui.ViewType
		expected    []ui.ViewType
	}{
		{
			name:        "push to empty history",
			initialView: []ui.ViewType{},
			pushView:    ui.DockerContainerListView,
			expected:    []ui.ViewType{ui.DockerContainerListView},
		},
		{
			name:        "push to existing history",
			initialView: []ui.ViewType{ui.DockerContainerListView},
			pushView:    ui.ComposeProcessListView,
			expected:    []ui.ViewType{ui.DockerContainerListView, ui.ComposeProcessListView},
		},
		{
			name:        "push multiple views",
			initialView: []ui.ViewType{ui.DockerContainerListView, ui.ComposeProcessListView},
			pushView:    ui.ImageListView,
			expected:    []ui.ViewType{ui.DockerContainerListView, ui.ComposeProcessListView, ui.ImageListView},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewState()
			state.ViewHistory = tt.initialView
			
			state.PushView(tt.pushView)
			
			assert.Equal(t, tt.expected, state.ViewHistory)
			assert.Equal(t, tt.pushView, state.CurrentView)
		})
	}
}

func TestState_PopView(t *testing.T) {
	tests := []struct {
		name            string
		initialHistory  []ui.ViewType
		expectedView    ui.ViewType
		expectedHistory []ui.ViewType
	}{
		{
			name:            "pop from empty history",
			initialHistory:  []ui.ViewType{},
			expectedView:    ui.ViewType(0),
			expectedHistory: []ui.ViewType{},
		},
		{
			name:            "pop from single item",
			initialHistory:  []ui.ViewType{ui.DockerContainerListView},
			expectedView:    ui.DockerContainerListView,
			expectedHistory: []ui.ViewType{ui.DockerContainerListView},
		},
		{
			name:            "pop from multiple items",
			initialHistory:  []ui.ViewType{ui.DockerContainerListView, ui.ComposeProcessListView, ui.ImageListView},
			expectedView:    ui.ComposeProcessListView,
			expectedHistory: []ui.ViewType{ui.DockerContainerListView, ui.ComposeProcessListView},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewState()
			state.ViewHistory = tt.initialHistory
			if len(tt.initialHistory) > 0 {
				state.CurrentView = tt.initialHistory[len(tt.initialHistory)-1]
			}
			
			result := state.PopView()
			
			assert.Equal(t, tt.expectedView, result)
			assert.Equal(t, tt.expectedHistory, state.ViewHistory)
			assert.Equal(t, tt.expectedView, state.CurrentView)
		})
	}
}

func TestState_ClearError(t *testing.T) {
	state := NewState()
	state.Error = errors.New("test error")
	
	state.ClearError()
	
	assert.Nil(t, state.Error)
}

func TestState_SetError(t *testing.T) {
	state := NewState()
	state.Loading = true
	testErr := errors.New("test error")
	
	state.SetError(testErr)
	
	assert.Equal(t, testErr, state.Error)
	assert.False(t, state.Loading)
}

func TestState_SetLoading(t *testing.T) {
	tests := []struct {
		name          string
		initialError  error
		loading       bool
		expectedError error
	}{
		{
			name:          "set loading true clears error",
			initialError:  errors.New("test error"),
			loading:       true,
			expectedError: nil,
		},
		{
			name:          "set loading false keeps error",
			initialError:  errors.New("test error"),
			loading:       false,
			expectedError: errors.New("test error"),
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewState()
			state.Error = tt.initialError
			
			state.SetLoading(tt.loading)
			
			assert.Equal(t, tt.loading, state.Loading)
			if tt.expectedError == nil {
				assert.Nil(t, state.Error)
			} else {
				assert.NotNil(t, state.Error)
			}
		})
	}
}

func TestState_ViewHistoryManagement(t *testing.T) {
	state := NewState()
	
	// Test pushing multiple views
	views := []ui.ViewType{
		ui.DockerContainerListView,
		ui.ComposeProcessListView,
		ui.ImageListView,
		ui.NetworkListView,
		ui.VolumeListView,
	}
	
	for _, view := range views {
		state.PushView(view)
	}
	
	assert.Equal(t, len(views), len(state.ViewHistory))
	assert.Equal(t, views[len(views)-1], state.CurrentView)
	
	// Test popping views
	for i := len(views) - 1; i > 0; i-- {
		result := state.PopView()
		assert.Equal(t, views[i-1], result)
		assert.Equal(t, i, len(state.ViewHistory))
	}
	
	// Test popping when only one view remains
	result := state.PopView()
	assert.Equal(t, views[0], result)
	assert.Equal(t, 1, len(state.ViewHistory))
}

func TestState_SearchAndCommand(t *testing.T) {
	state := NewState()
	
	// Test search text
	state.SearchText = "test search"
	assert.Equal(t, "test search", state.SearchText)
	
	// Test command mode
	state.CommandMode = true
	state.CommandText = ":quit"
	assert.True(t, state.CommandMode)
	assert.Equal(t, ":quit", state.CommandText)
	
	// Test selected items
	state.SelectedProject = "my-project"
	state.SelectedContainer = "container-123"
	assert.Equal(t, "my-project", state.SelectedProject)
	assert.Equal(t, "container-123", state.SelectedContainer)
	
	// Test show all toggle
	state.ShowAll = true
	assert.True(t, state.ShowAll)
}