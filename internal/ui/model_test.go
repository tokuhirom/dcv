package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/tokuhirom/dcv/internal/models"
)

func TestNewModel(t *testing.T) {
	m := NewModel(ComposeProcessListView, "")

	assert.Equal(t, ComposeProcessListView, m.currentView)
	assert.NotNil(t, m.dockerClient)
	assert.True(t, m.loading)
	assert.Empty(t, m.composeProcessListViewModel.composeContainers)
	assert.Equal(t, 0, m.composeProcessListViewModel.selectedContainer)
}

func TestModelInit(t *testing.T) {
	m := NewModel(ComposeProcessListView, "")
	cmd := m.Init()

	// Init should return a batch command
	assert.NotNil(t, cmd)
}

func TestProcessesLoadedMsg(t *testing.T) {
	m := NewModel(ComposeProcessListView, "")

	// Test successful load
	containers := []models.ComposeContainer{
		{
			Name:    "web-1",
			Command: "/docker-entrypoint.sh nginx -g 'daemon off;'",
			Service: "web",
			State:   "running",
		},
		{
			Name:    "dind-1",
			Command: "dockerd",
			Service: "dind",
			State:   "running",
		},
	}

	msg := processesLoadedMsg{
		processes: containers,
		err:       nil,
	}

	newModel, cmd := m.Update(msg)
	m = *newModel.(*Model)

	assert.False(t, m.loading)
	assert.Nil(t, m.err)
	assert.Equal(t, 2, len(m.composeProcessListViewModel.composeContainers))
	assert.Equal(t, "web-1", m.composeProcessListViewModel.composeContainers[0].Name)
	assert.Equal(t, "dind-1", m.composeProcessListViewModel.composeContainers[1].Name)
	assert.True(t, m.composeProcessListViewModel.composeContainers[1].IsDind())
	assert.Nil(t, cmd)
}

func TestWindowSizeMsg(t *testing.T) {
	m := NewModel(ComposeProcessListView, "")

	msg := tea.WindowSizeMsg{
		Width:  80,
		Height: 24,
	}

	newModel, cmd := m.Update(msg)
	m = *newModel.(*Model)

	assert.Equal(t, 80, m.width)
	assert.Equal(t, 24, m.Height)
	assert.Nil(t, cmd)
}

func TestKeyNavigation(t *testing.T) {
	m := NewModel(ComposeProcessListView, "")
	m.Init() // Initialize key handlers
	m.loading = false
	m.composeProcessListViewModel.composeContainers = []models.ComposeContainer{
		{Name: "web-1"},
		{Name: "db-1"},
		{Name: "redis-1"},
	}

	tests := []struct {
		name     string
		key      string
		expected int
	}{
		{"down arrow", "down", 1},
		{"j key", "j", 1},
		{"up arrow", "up", 0},
		{"k key", "k", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.composeProcessListViewModel.selectedContainer = 0

			// Move down
			if tt.expected > 0 {
				msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
				if tt.key == "down" {
					msg = tea.KeyMsg{Type: tea.KeyDown}
				}
				newModel, _ := m.Update(msg)
				m = *newModel.(*Model)
			}

			assert.Equal(t, tt.expected, m.composeProcessListViewModel.selectedContainer)
		})
	}
}

func TestViewSwitching(t *testing.T) {
	m := NewModel(ComposeProcessListView, "")
	m.Init() // Initialize key handlers
	m.loading = false
	m.composeProcessListViewModel.composeContainers = []models.ComposeContainer{
		{
			Name: "web-1",
		},
		{
			Name:    "dind-1",
			Command: "dockerd",
		},
	}

	// Test entering log view
	m.composeProcessListViewModel.selectedContainer = 0
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd := m.Update(msg)
	m = *newModel.(*Model)

	assert.Equal(t, LogView, m.currentView)
	assert.Equal(t, "web-1", m.logViewModel.containerName)
	assert.False(t, m.logViewModel.isDindLog)
	assert.NotNil(t, cmd)

	// Test going back with ESC
	msg = tea.KeyMsg{Type: tea.KeyEsc}
	newModel, cmd = m.Update(msg)
	m = *newModel.(*Model)

	assert.Equal(t, ComposeProcessListView, m.currentView)
	assert.NotNil(t, cmd)

	// Test entering dind view
	m.composeProcessListViewModel.selectedContainer = 1
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")}
	newModel, cmd = m.Update(msg)
	m = *newModel.(*Model)

	assert.Equal(t, DindProcessListView, m.currentView)
	assert.Equal(t, "dind-1", m.dindProcessListViewModel.currentDindHost)
	assert.True(t, m.loading)
	assert.NotNil(t, cmd)
}

func TestSearchMode(t *testing.T) {
	m := NewModel(ComposeProcessListView, "")
	m.Init() // Initialize key handlers
	m.currentView = LogView
	m.logViewModel.logs = []string{"line 1", "line 2", "error occurred", "line 4"}

	// Enter search mode
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")}
	newModel, _ := m.Update(msg)
	m = *newModel.(*Model)

	assert.True(t, m.logViewModel.searchMode)
	assert.Equal(t, "", m.logViewModel.searchText)

	// Type search text
	for _, r := range "error" {
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		newModel, _ = m.Update(msg)
		m = *newModel.(*Model)
	}

	assert.Equal(t, "error", m.logViewModel.searchText)

	// Exit search mode with ESC
	msg = tea.KeyMsg{Type: tea.KeyEsc}
	newModel, _ = m.Update(msg)
	m = *newModel.(*Model)

	assert.False(t, m.logViewModel.searchMode)
}

func TestErrorHandling(t *testing.T) {
	m := NewModel(ComposeProcessListView, "")

	// Test error message
	msg := errorMsg{err: assert.AnError}
	newModel, _ := m.Update(msg)
	m = *newModel.(*Model)

	assert.NotNil(t, m.err)
	assert.False(t, m.loading)
}

func TestQuitBehavior(t *testing.T) {
	tests := []struct {
		name        string
		currentView ViewType
		expectQuit  bool
		expectView  ViewType
	}{
		{
			name:        "quit from process list shows confirmation",
			currentView: ComposeProcessListView,
			expectQuit:  false,
			expectView:  ComposeProcessListView,
		},
		{
			name:        "quit from log view shows confirmation",
			currentView: LogView,
			expectQuit:  false,
			expectView:  LogView, // View should not change
		},
		{
			name:        "quit from dind view shows confirmation",
			currentView: DindProcessListView,
			expectQuit:  false,
			expectView:  DindProcessListView, // View should not change
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel(ComposeProcessListView, "")
			m.initializeKeyHandlers() // Initialize key handlers to register global 'q' handler
			m.currentView = tt.currentView
			m.loading = false

			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
			newModel, cmd := m.Update(msg)
			m = *newModel.(*Model)

			assert.Equal(t, tt.expectView, m.currentView)

			// All views now show quit confirmation when 'q' is pressed
			assert.True(t, m.quitConfirmation)
			assert.Equal(t, tt.currentView, m.currentView) // View should not change
			assert.Nil(t, cmd)
		})
	}
}
