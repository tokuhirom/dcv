package ui

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/tokuhirom/dcv/internal/models"
)

func TestHandleKeyPress(t *testing.T) {
	tests := []struct {
		name        string
		model       Model
		key         tea.KeyMsg
		wantView    ViewType
		wantLoading bool
		checkFunc   func(t *testing.T, m Model)
	}{
		{
			name: "navigate down in process list",
			model: Model{
				currentView: ProcessListView,
				containers: []models.Container{
					{Name: "web-1"},
					{Name: "db-1"},
				},
				selectedContainer: 0,
			},
			key:      tea.KeyMsg{Type: tea.KeyDown},
			wantView: ProcessListView,
			checkFunc: func(t *testing.T, m Model) {
				assert.Equal(t, 1, m.selectedContainer)
			},
		},
		{
			name: "navigate up in process list",
			model: Model{
				currentView: ProcessListView,
				containers: []models.Container{
					{Name: "web-1"},
					{Name: "db-1"},
				},
				selectedContainer: 1,
			},
			key:      tea.KeyMsg{Type: tea.KeyUp},
			wantView: ProcessListView,
			checkFunc: func(t *testing.T, m Model) {
				assert.Equal(t, 0, m.selectedContainer)
			},
		},
		{
			name: "enter log view",
			model: Model{
				currentView: ProcessListView,
				containers: []models.Container{
					{Name: "web-1"},
				},
				selectedContainer: 0,
			},
			key:      tea.KeyMsg{Type: tea.KeyEnter},
			wantView: LogView,
			checkFunc: func(t *testing.T, m Model) {
				assert.Equal(t, "web-1", m.containerName)
				assert.False(t, m.isDindLog)
			},
		},
		{
			name: "enter dind view",
			model: Model{
				currentView: ProcessListView,
				containers: []models.Container{
					{
						Name:  "dind-1",
						Image: "docker:dind",
					},
				},
				selectedContainer: 0,
			},
			key:         tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")},
			wantView:    DindProcessListView,
			wantLoading: true,
			checkFunc: func(t *testing.T, m Model) {
				assert.Equal(t, "dind-1", m.currentDindHost)
			},
		},
		{
			name: "quit from process list",
			model: Model{
				currentView: ProcessListView,
			},
			key:      tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")},
			wantView: ProcessListView,
		},
		{
			name: "escape from log view",
			model: Model{
				currentView:   LogView,
				containerName: "web-1",
			},
			key:      tea.KeyMsg{Type: tea.KeyEsc},
			wantView: ProcessListView,
		},
		{
			name: "scroll down in log view",
			model: Model{
				currentView: LogView,
				height:      10,
				logs:        []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
				logScrollY:  0,
			},
			key:      tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")},
			wantView: LogView,
			checkFunc: func(t *testing.T, m Model) {
				assert.Equal(t, 1, m.logScrollY)
			},
		},
		{
			name: "jump to end in log view",
			model: Model{
				currentView: LogView,
				height:      5,
				logs:        []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
				logScrollY:  0,
			},
			key:      tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")},
			wantView: LogView,
			checkFunc: func(t *testing.T, m Model) {
				assert.Equal(t, 9, m.logScrollY) // 10 logs - (5 height - 4 ui elements) = 9
			},
		},
		{
			name: "enter search mode",
			model: Model{
				currentView: LogView,
			},
			key:      tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")},
			wantView: LogView,
			checkFunc: func(t *testing.T, m Model) {
				assert.True(t, m.searchMode)
				assert.Equal(t, "", m.searchText)
			},
		},
		{
			name: "refresh process list",
			model: Model{
				currentView: ProcessListView,
			},
			key:         tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")},
			wantView:    ProcessListView,
			wantLoading: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newModel, _ := tt.model.handleKeyPress(tt.key)
			m := newModel.(Model)

			assert.Equal(t, tt.wantView, m.currentView)
			if tt.wantLoading {
				assert.True(t, m.loading)
			}
			if tt.checkFunc != nil {
				tt.checkFunc(t, m)
			}
		})
	}
}

func TestHandleSearchMode(t *testing.T) {
	tests := []struct {
		name           string
		model          Model
		key            tea.KeyMsg
		wantSearchMode bool
		wantSearchText string
	}{
		{
			name: "type in search mode",
			model: Model{
				searchMode: true,
				searchText: "err",
			},
			key:            tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")},
			wantSearchMode: true,
			wantSearchText: "erro",
		},
		{
			name: "backspace in search mode",
			model: Model{
				searchMode: true,
				searchText: "error",
			},
			key:            tea.KeyMsg{Type: tea.KeyBackspace},
			wantSearchMode: true,
			wantSearchText: "erro",
		},
		{
			name: "escape search mode",
			model: Model{
				searchMode: true,
				searchText: "error",
			},
			key:            tea.KeyMsg{Type: tea.KeyEsc},
			wantSearchMode: false,
			wantSearchText: "",
		},
		{
			name: "enter to confirm search",
			model: Model{
				searchMode: true,
				searchText: "error",
			},
			key:            tea.KeyMsg{Type: tea.KeyEnter},
			wantSearchMode: false,
			wantSearchText: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newModel, _ := tt.model.handleSearchMode(tt.key)
			m := newModel.(Model)

			assert.Equal(t, tt.wantSearchMode, m.searchMode)
			assert.Equal(t, tt.wantSearchText, m.searchText)
		})
	}
}

func TestHandleDindListKeys(t *testing.T) {
	model := Model{
		currentView:     DindProcessListView,
		currentDindHost: "dind-1",
		dindContainers: []models.Container{
			{ID: "abc123", Name: "test-1"},
			{ID: "def456", Name: "test-2"},
		},
		selectedDindContainer: 0,
	}

	// Test navigation
	newModel, _ := model.handleDindListKeys(tea.KeyMsg{Type: tea.KeyDown})
	m := newModel.(Model)
	assert.Equal(t, 1, m.selectedDindContainer)

	// Test entering log view
	newModel, cmd := m.handleDindListKeys(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(Model)
	assert.Equal(t, LogView, m.currentView)
	assert.Equal(t, "test-2", m.containerName)
	assert.Equal(t, "dind-1", m.hostContainer)
	assert.True(t, m.isDindLog)
	assert.NotNil(t, cmd)

	// Test escape
	model.currentView = DindProcessListView
	newModel, _ = model.handleDindListKeys(tea.KeyMsg{Type: tea.KeyEsc})
	m = newModel.(Model)
	assert.Equal(t, ProcessListView, m.currentView)
}

func TestUpdateMessages(t *testing.T) {
	model := NewModel(ProcessListView, "")

	// Test window size message
	newModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m := newModel.(Model)
	assert.Equal(t, 100, m.width)
	assert.Equal(t, 30, m.height)

	// Test containers loaded message
	processes := []models.Container{
		{Name: "test-1"},
	}
	newModel, _ = m.Update(processesLoadedMsg{processes: processes})
	m = newModel.(Model)
	assert.Equal(t, processes, m.containers)
	assert.False(t, m.loading)

	// Test error message
	testErr := errors.New("test error")
	newModel, _ = m.Update(errorMsg{err: testErr})
	m = newModel.(Model)
	assert.Equal(t, testErr, m.err)
	assert.False(t, m.loading)

	// Test log line message (for status messages like "[Log reader stopped]")
	m.currentView = LogView
	m.height = 10
	newModel, cmd := m.Update(logLineMsg{line: "[Log reader stopped]"})
	m = newModel.(Model)
	assert.Contains(t, m.logs, "[Log reader stopped]")
	assert.Nil(t, cmd) // Status messages don't trigger continued polling

	// Test log lines message (for actual log streaming)
	m.logs = []string{} // Reset logs
	newModel, cmd = m.Update(logLinesMsg{lines: []string{"log line 1", "log line 2"}})
	m = newModel.(Model)
	assert.Contains(t, m.logs, "log line 1")
	assert.Contains(t, m.logs, "log line 2")
	assert.NotNil(t, cmd) // Should continue streaming

	// Test dind containers loaded
	containers := []models.Container{
		{ID: "abc123", Name: "test-container"},
	}
	newModel, _ = m.Update(dindContainersLoadedMsg{containers: containers})
	m = newModel.(Model)
	assert.Equal(t, containers, m.dindContainers)
	assert.False(t, m.loading)
}

func TestBoundaryConditions(t *testing.T) {
	// Test navigation at boundaries
	model := Model{
		currentView: ProcessListView,
		containers: []models.Container{
			{Name: "test-1"},
		},
		selectedContainer: 0,
	}

	// Try to go up at the top
	newModel, _ := model.handleKeyPress(tea.KeyMsg{Type: tea.KeyUp})
	m := newModel.(Model)
	assert.Equal(t, 0, m.selectedContainer) // Should stay at 0

	// Try to go down at the bottom
	newModel, _ = m.handleKeyPress(tea.KeyMsg{Type: tea.KeyDown})
	m = newModel.(Model)
	assert.Equal(t, 0, m.selectedContainer) // Should stay at 0 (only one item)

	// Test with empty list
	model.containers = []models.Container{}
	newModel, _ = model.handleKeyPress(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(Model)
	assert.Equal(t, ProcessListView, m.currentView) // Should stay in process list
}

func TestQuitBehaviorInDifferentViews(t *testing.T) {
	// From process list - should quit
	model := Model{currentView: ProcessListView}
	_, cmd := model.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	assert.NotNil(t, cmd)

	// From log view - should go back
	model = Model{currentView: LogView}
	newModel, cmd := model.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	m := newModel.(Model)
	assert.Equal(t, ProcessListView, m.currentView)
	assert.NotNil(t, cmd) // Should load containers

	// From dind view - should go back
	model = Model{currentView: DindProcessListView}
	newModel, cmd = model.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	m = newModel.(Model)
	assert.Equal(t, ProcessListView, m.currentView)
	assert.NotNil(t, cmd) // Should load containers
}
