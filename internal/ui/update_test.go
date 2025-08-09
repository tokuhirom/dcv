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
		checkFunc   func(t *testing.T, m *Model)
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
			checkFunc: func(t *testing.T, m *Model) {
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
			checkFunc: func(t *testing.T, m *Model) {
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
			checkFunc: func(t *testing.T, m *Model) {
				assert.Equal(t, "web-1", m.logViewModel.containerName)
				assert.False(t, m.logViewModel.isDindLog)
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
			checkFunc: func(t *testing.T, m *Model) {
				assert.Equal(t, "dind-1", m.dindProcessListViewModel.currentDindHostName)
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
				currentView: LogView,
				viewHistory: []ViewType{ComposeProcessListView},
				logViewModel: LogViewModel{
					containerName: "web-1",
				},
			},
			key:      tea.KeyMsg{Type: tea.KeyEsc},
			wantView: ComposeProcessListView,
		},
		{
			name: "scroll down in log view",
			model: Model{
				currentView: LogView,
				Height:      10,
				logViewModel: LogViewModel{
					logs:       []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
					logScrollY: 0,
				},
			},
			key:      tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")},
			wantView: LogView,
			checkFunc: func(t *testing.T, m *Model) {
				assert.Equal(t, 1, m.logViewModel.logScrollY)
			},
		},
		{
			name: "jump to end in log view",
			model: Model{
				currentView: LogView,
				Height:      5,
				logViewModel: LogViewModel{
					logs:       []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
					logScrollY: 0,
				},
			},
			key:      tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")},
			wantView: LogView,
			checkFunc: func(t *testing.T, m *Model) {
				assert.Equal(t, 9, m.logViewModel.logScrollY) // 10 logs - (5 Height - 4 ui elements) = 9
			},
		},
		{
			name: "enter search mode",
			model: Model{
				currentView: LogView,
			},
			key:      tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")},
			wantView: LogView,
			checkFunc: func(t *testing.T, m *Model) {
				assert.True(t, m.logViewModel.searchMode)
				assert.Equal(t, "", m.logViewModel.searchText)
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

	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			// Initialize key handlers for the test model
			tt.model.initializeKeyHandlers()

			newModel, _ := tt.model.handleKeyPress(tt.key)
			m := newModel.(*Model)

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
				currentView: LogView,
				logViewModel: LogViewModel{
					SearchViewModel: SearchViewModel{
						searchMode:      true,
						searchText:      "err",
						searchCursorPos: 3, // at end of "err"
					},
				},
			},
			key:            tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")},
			wantSearchMode: true,
			wantSearchText: "erro",
		},
		{
			name: "backspace in search mode",
			model: Model{
				currentView: LogView,
				logViewModel: LogViewModel{
					SearchViewModel: SearchViewModel{
						searchMode:      true,
						searchText:      "error",
						searchCursorPos: 5, // at end of "error"
					},
				},
			},
			key:            tea.KeyMsg{Type: tea.KeyBackspace},
			wantSearchMode: true,
			wantSearchText: "erro",
		},
		{
			name: "escape search mode",
			model: Model{
				currentView: LogView,
				logViewModel: LogViewModel{
					SearchViewModel: SearchViewModel{
						searchMode: true,
						searchText: "error",
					},
				},
			},
			key:            tea.KeyMsg{Type: tea.KeyEsc},
			wantSearchMode: false,
			wantSearchText: "",
		},
		{
			name: "enter to confirm search",
			model: Model{
				currentView: LogView,
				logViewModel: LogViewModel{
					SearchViewModel: SearchViewModel{
						searchMode: true,
						searchText: "error",
					},
				},
			},
			key:            tea.KeyMsg{Type: tea.KeyEnter},
			wantSearchMode: false,
			wantSearchText: "error",
		},
	}

	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			// Initialize key handlers for the test model
			tt.model.initializeKeyHandlers()

			newModel, _ := tt.model.handleSearchMode(tt.key, &tt.model.logViewModel.SearchViewModel)
			m := newModel.(*Model)

			assert.Equal(t, tt.wantSearchMode, m.logViewModel.searchMode)
			assert.Equal(t, tt.wantSearchText, m.logViewModel.searchText)
		})
	}
}

func TestHandleDindListKeys(t *testing.T) {
	model := Model{
		currentView: DindProcessListView,
		dindProcessListViewModel: DindProcessListViewModel{
			currentDindHostName: "dind-1",
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
	newModel, _ := model.handleViewKeys(tea.KeyMsg{Type: tea.KeyDown})
	m := newModel.(*Model)
	assert.Equal(t, 1, m.dindProcessListViewModel.selectedDindContainer)

	// Test entering log view
	newModel, cmd := m.handleViewKeys(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(*Model)
	assert.Equal(t, LogView, m.currentView)
	assert.Equal(t, "test-2", m.logViewModel.containerName)
	assert.Equal(t, "dind-1", m.logViewModel.hostContainer)
	assert.True(t, m.logViewModel.isDindLog)
	assert.NotNil(t, cmd)

	// Test escape - add viewHistory for proper navigation
	model.currentView = DindProcessListView
	model.viewHistory = []ViewType{ComposeProcessListView}
	newModel, _ = model.handleViewKeys(tea.KeyMsg{Type: tea.KeyEsc})
	m = newModel.(*Model)
	assert.Equal(t, ComposeProcessListView, m.currentView)
}

func TestUpdateMessages(t *testing.T) {
	model := NewModel(ComposeProcessListView)

	// Test window size message
	newModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m := newModel.(*Model)
	assert.Equal(t, 100, m.width)
	assert.Equal(t, 30, m.Height)

	// Test composeContainers loaded message
	processes := []models.ComposeContainer{
		{Name: "test-1"},
	}
	newModel, _ = m.Update(processesLoadedMsg{processes: processes})
	m = newModel.(*Model)
	assert.Equal(t, processes, m.composeProcessListViewModel.composeContainers)
	assert.False(t, m.loading)

	// Test error message
	testErr := errors.New("test error")
	newModel, _ = m.Update(errorMsg{err: testErr})
	m = newModel.(*Model)
	assert.Equal(t, testErr, m.err)
	assert.False(t, m.loading)

	// Test log lines message (status messages are now handled same as regular logs)
	m.currentView = LogView
	m.Height = 10
	newModel, cmd := m.Update(logLinesMsg{lines: []string{"[Log reader stopped]"}})
	m = newModel.(*Model)
	assert.Contains(t, m.logViewModel.logs, "[Log reader stopped]")
	assert.NotNil(t, cmd) // logLinesMsg always returns a command to continue polling

	// Test log lines message (for actual log streaming)
	m.logViewModel.logs = []string{} // Reset logs
	newModel, cmd = m.Update(logLinesMsg{lines: []string{"log line 1", "log line 2"}})
	m = newModel.(*Model)
	assert.Contains(t, m.logViewModel.logs, "log line 1")
	assert.Contains(t, m.logViewModel.logs, "log line 2")
	assert.NotNil(t, cmd) // Should continue streaming

	// Test dind composeContainers loaded
	containers := []models.DockerContainer{
		{ID: "abc123", Names: "test-container"},
	}
	newModel, _ = m.Update(dindContainersLoadedMsg{containers: containers})
	m = newModel.(*Model)
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
	m := newModel.(*Model)
	assert.Equal(t, 0, m.composeProcessListViewModel.selectedContainer) // Should stay at 0

	// Try to go down at the bottom
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = newModel.(*Model)
	assert.Equal(t, 0, m.composeProcessListViewModel.selectedContainer) // Should stay at 0 (only one item)

	// Test with empty list
	model.composeProcessListViewModel.composeContainers = []models.ComposeContainer{}
	newModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(*Model)
	assert.Equal(t, ComposeProcessListView, m.currentView) // Should stay in process list
}

func TestQuitBehaviorInDifferentViews(t *testing.T) {
	// From process list - should show quit confirmation
	model := Model{currentView: ComposeProcessListView}
	model.initializeKeyHandlers() // Initialize key handlers to register global 'q' handler
	newModel, cmd := model.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	m := newModel.(*Model)
	assert.True(t, m.quitConfirmation)
	assert.Nil(t, cmd) // No command yet, just showing confirmation

	// From log view - should now show quit confirmation
	model = Model{currentView: LogView}
	model.initializeKeyHandlers() // Initialize key handlers to register global 'q' handler
	newModel, cmd = model.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	m = newModel.(*Model)
	assert.Equal(t, LogView, m.currentView) // View should not change
	assert.True(t, m.quitConfirmation)
	assert.Nil(t, cmd)

	// From dind view - should now show quit confirmation
	model = Model{currentView: DindProcessListView}
	model.initializeKeyHandlers() // Initialize key handlers to register global 'q' handler
	newModel, cmd = model.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	m = newModel.(*Model)
	assert.Equal(t, DindProcessListView, m.currentView) // View should not change
	assert.True(t, m.quitConfirmation)
	assert.Nil(t, cmd)
}

func TestFileBrowserParentDirectory(t *testing.T) {
	// Test 'u' key to go to parent directory
	model := Model{
		currentView: FileBrowserView,
		fileBrowserViewModel: FileBrowserViewModel{
			browsingContainerID: "test-container",
			currentPath:         "/app/src",
			pathHistory:         []string{"/", "/app", "/app/src"},
			containerFiles: []models.ContainerFile{
				{Name: "..", IsDir: true},
				{Name: "file.txt", IsDir: false},
			},
		},
	}
	model.initializeKeyHandlers()

	// Press 'u' to go to parent directory
	newModel, cmd := model.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("u")})
	m := newModel.(*Model)

	assert.Equal(t, "/app", m.fileBrowserViewModel.currentPath)
	assert.Equal(t, 2, len(m.fileBrowserViewModel.pathHistory)) // Should have removed the last entry
	assert.True(t, m.loading)
	assert.Equal(t, 0, m.fileBrowserViewModel.selectedFile) // Should reset selection
	assert.NotNil(t, cmd)                                   // Should trigger loading parent directory files

	// Test at root directory - should not change
	model.fileBrowserViewModel.currentPath = "/"
	model.fileBrowserViewModel.pathHistory = []string{"/"}
	model.loading = false

	newModel, cmd = model.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("u")})
	m = newModel.(*Model)

	assert.Equal(t, "/", m.fileBrowserViewModel.currentPath) // Should stay at root
	assert.Equal(t, 1, len(m.fileBrowserViewModel.pathHistory))
	assert.False(t, m.loading) // Should not trigger loading
	assert.Nil(t, cmd)         // No command when already at root
}
