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
				currentView: ComposeProcessListView,
				composeProcessListViewModel: ComposeProcessListViewModel{
					composeContainers: []models.ComposeContainer{
						{Name: "web-1"},
						{Name: "db-1"},
					},
					selectedContainer: 0,
				},
			},
			key:      tea.KeyMsg{Type: tea.KeyDown},
			wantView: ComposeProcessListView,
			checkFunc: func(t *testing.T, m Model) {
				assert.Equal(t, 1, m.composeProcessListViewModel.selectedContainer)
			},
		},
		{
			name: "navigate up in process list",
			model: Model{
				currentView: ComposeProcessListView,
				composeProcessListViewModel: ComposeProcessListViewModel{
					composeContainers: []models.ComposeContainer{
						{Name: "web-1"},
						{Name: "db-1"},
					},
					selectedContainer: 1,
				},
			},
			key:      tea.KeyMsg{Type: tea.KeyUp},
			wantView: ComposeProcessListView,
			checkFunc: func(t *testing.T, m Model) {
				assert.Equal(t, 0, m.composeProcessListViewModel.selectedContainer)
			},
		},
		{
			name: "enter log view",
			model: Model{
				currentView: ComposeProcessListView,
				composeProcessListViewModel: ComposeProcessListViewModel{
					composeContainers: []models.ComposeContainer{
						{Name: "web-1"},
					},
					selectedContainer: 0,
				},
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
				currentView: ComposeProcessListView,
				composeProcessListViewModel: ComposeProcessListViewModel{
					composeContainers: []models.ComposeContainer{
						{
							Name:    "dind-1",
							Command: "dockerd",
						},
					},
					selectedContainer: 0,
				},
			},
			key:         tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")},
			wantView:    DindProcessListView,
			wantLoading: true,
			checkFunc: func(t *testing.T, m Model) {
				assert.Equal(t, "dind-1", m.dindProcessListViewModel.currentDindHost)
			},
		},
		{
			name: "quit from process list",
			model: Model{
				currentView: ComposeProcessListView,
			},
			key:      tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")},
			wantView: ComposeProcessListView,
		},
		{
			name: "escape from log view",
			model: Model{
				currentView:   LogView,
				containerName: "web-1",
			},
			key:      tea.KeyMsg{Type: tea.KeyEsc},
			wantView: ComposeProcessListView,
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
				currentView: ComposeProcessListView,
			},
			key:         tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")},
			wantView:    ComposeProcessListView,
			wantLoading: false, // Loading is set in Update when handling RefreshMsg, not in the key handler
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize key handlers for the test model
			tt.model.initializeKeyHandlers()

			newModel, _ := tt.model.handleKeyPress(tt.key)
			m := *newModel.(*Model)

			assert.Equal(t, tt.wantView, m.currentView)
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
				searchMode:      true,
				searchText:      "err",
				searchCursorPos: 3, // at end of "err"
			},
			key:            tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")},
			wantSearchMode: true,
			wantSearchText: "erro",
		},
		{
			name: "backspace in search mode",
			model: Model{
				searchMode:      true,
				searchText:      "error",
				searchCursorPos: 5, // at end of "error"
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
			// Initialize key handlers for the test model
			tt.model.initializeKeyHandlers()

			newModel, _ := tt.model.handleSearchMode(tt.key)
			m := *newModel.(*Model)

			assert.Equal(t, tt.wantSearchMode, m.searchMode)
			assert.Equal(t, tt.wantSearchText, m.searchText)
		})
	}
}

func TestHandleDindListKeys(t *testing.T) {
	model := Model{
		currentView: DindProcessListView,
		dindProcessListViewModel: DindProcessListViewModel{
			currentDindHost: "dind-1",
			dindContainers: []models.DockerContainer{
				{ID: "abc123", Names: "test-1"},
				{ID: "def456", Names: "test-2"},
			},
			selectedDindContainer: 0,
		},
	}
	// Initialize key handlers
	model.initializeKeyHandlers()

	// Test navigation
	newModel, _ := model.handleDindListKeys(tea.KeyMsg{Type: tea.KeyDown})
	m := *newModel.(*Model)
	assert.Equal(t, 1, m.dindProcessListViewModel.selectedDindContainer)

	// Test entering log view
	newModel, cmd := m.handleDindListKeys(tea.KeyMsg{Type: tea.KeyEnter})
	m = *newModel.(*Model)
	assert.Equal(t, LogView, m.currentView)
	assert.Equal(t, "test-2", m.containerName)
	assert.Equal(t, "dind-1", m.hostContainer)
	assert.True(t, m.isDindLog)
	assert.NotNil(t, cmd)

	// Test escape
	model.currentView = DindProcessListView
	newModel, _ = model.handleDindListKeys(tea.KeyMsg{Type: tea.KeyEsc})
	m = *newModel.(*Model)
	assert.Equal(t, ComposeProcessListView, m.currentView)
}

func TestUpdateMessages(t *testing.T) {
	model := NewModel(ComposeProcessListView, "")

	// Test window size message
	newModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m := *newModel.(*Model)
	assert.Equal(t, 100, m.width)
	assert.Equal(t, 30, m.height)

	// Test composeContainers loaded message
	processes := []models.ComposeContainer{
		{Name: "test-1"},
	}
	newModel, _ = m.Update(processesLoadedMsg{processes: processes})
	m = *newModel.(*Model)
	assert.Equal(t, processes, m.composeProcessListViewModel.composeContainers)
	assert.False(t, m.loading)

	// Test error message
	testErr := errors.New("test error")
	newModel, _ = m.Update(errorMsg{err: testErr})
	m = *newModel.(*Model)
	assert.Equal(t, testErr, m.err)
	assert.False(t, m.loading)

	// Test log line message (for status messages like "[Log reader stopped]")
	m.currentView = LogView
	m.height = 10
	newModel, cmd := m.Update(logLineMsg{line: "[Log reader stopped]"})
	m = *newModel.(*Model)
	assert.Contains(t, m.logs, "[Log reader stopped]")
	assert.Nil(t, cmd) // Status messages don't trigger continued polling

	// Test log lines message (for actual log streaming)
	m.logs = []string{} // Reset logs
	newModel, cmd = m.Update(logLinesMsg{lines: []string{"log line 1", "log line 2"}})
	m = *newModel.(*Model)
	assert.Contains(t, m.logs, "log line 1")
	assert.Contains(t, m.logs, "log line 2")
	assert.NotNil(t, cmd) // Should continue streaming

	// Test dind composeContainers loaded
	containers := []models.DockerContainer{
		{ID: "abc123", Names: "test-container"},
	}
	newModel, _ = m.Update(dindContainersLoadedMsg{containers: containers})
	m = *newModel.(*Model)
	assert.Equal(t, containers, m.dindProcessListViewModel.dindContainers)
	assert.False(t, m.loading)
}

func TestBoundaryConditions(t *testing.T) {
	// Test navigation at boundaries
	model := Model{
		currentView: ComposeProcessListView,
		composeProcessListViewModel: ComposeProcessListViewModel{
			composeContainers: []models.ComposeContainer{
				{Name: "test-1"},
			},
			selectedContainer: 0,
		},
	}
	model.Init() // Initialize key handlers

	// Try to go up at the top
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyUp})
	m := *newModel.(*Model)
	assert.Equal(t, 0, m.composeProcessListViewModel.selectedContainer) // Should stay at 0

	// Try to go down at the bottom
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = *newModel.(*Model)
	assert.Equal(t, 0, m.composeProcessListViewModel.selectedContainer) // Should stay at 0 (only one item)

	// Test with empty list
	model.composeProcessListViewModel.composeContainers = []models.ComposeContainer{}
	newModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = *newModel.(*Model)
	assert.Equal(t, ComposeProcessListView, m.currentView) // Should stay in process list
}

func TestQuitBehaviorInDifferentViews(t *testing.T) {
	// From process list - should show quit confirmation
	model := Model{currentView: ComposeProcessListView}
	newModel, cmd := model.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	m := *newModel.(*Model)
	assert.True(t, m.quitConfirmation)
	assert.Nil(t, cmd) // No command yet, just showing confirmation

	// From log view - should go back
	model = Model{currentView: LogView}
	newModel, cmd = model.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	m = *newModel.(*Model)
	assert.Equal(t, ComposeProcessListView, m.currentView)
	assert.NotNil(t, cmd) // Should load composeContainers

	// From dind view - should go back
	model = Model{currentView: DindProcessListView}
	newModel, cmd = model.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	m = *newModel.(*Model)
	assert.Equal(t, ComposeProcessListView, m.currentView)
	assert.NotNil(t, cmd) // Should load composeContainers
}

func TestFileBrowserParentDirectory(t *testing.T) {
	// Test 'u' key to go to parent directory
	model := Model{
		currentView:         FileBrowserView,
		browsingContainerID: "test-container",
		currentPath:         "/app/src",
		pathHistory:         []string{"/", "/app", "/app/src"},
		containerFiles: []models.ContainerFile{
			{Name: "..", IsDir: true},
			{Name: "file.txt", IsDir: false},
		},
	}
	model.initializeKeyHandlers()

	// Press 'u' to go to parent directory
	newModel, cmd := model.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("u")})
	m := *newModel.(*Model)

	assert.Equal(t, "/app", m.currentPath)
	assert.Equal(t, 2, len(m.pathHistory)) // Should have removed the last entry
	assert.True(t, m.loading)
	assert.Equal(t, 0, m.selectedFile) // Should reset selection
	assert.NotNil(t, cmd)              // Should trigger loading parent directory files

	// Test at root directory - should not change
	model.currentPath = "/"
	model.pathHistory = []string{"/"}
	model.loading = false

	newModel, cmd = model.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("u")})
	m = *newModel.(*Model)

	assert.Equal(t, "/", m.currentPath) // Should stay at root
	assert.Equal(t, 1, len(m.pathHistory))
	assert.False(t, m.loading) // Should not trigger loading
	assert.Nil(t, cmd)         // No command when already at root
}
