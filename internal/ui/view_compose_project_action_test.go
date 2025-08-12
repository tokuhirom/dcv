package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/tokuhirom/dcv/internal/models"
)

func TestComposeProjectActionViewModel_Initialize(t *testing.T) {
	tests := []struct {
		name           string
		projectStatus  string
		expectedKeys   []string
		unexpectedKeys []string
	}{
		{
			name:           "running project shows down/stop/restart actions",
			projectStatus:  "running(2)",
			expectedKeys:   []string{"D", "S", "R"},
			unexpectedKeys: []string{"U"},
		},
		{
			name:           "stopped project shows up/start actions",
			projectStatus:  "exited(2)",
			expectedKeys:   []string{"U", "S"},
			unexpectedKeys: []string{"D", "R"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := &ComposeProjectActionViewModel{}
			project := &models.ComposeProject{
				Name:        "test-project",
				Status:      tt.projectStatus,
				ConfigFiles: "/path/to/docker-compose.yml",
			}

			vm.Initialize(project)

			// Check that target project is set
			if vm.targetProject != project {
				t.Errorf("target project not set correctly")
			}

			// Check that selection is reset
			if vm.selectedAction != 0 {
				t.Errorf("selectedAction = %d, want 0", vm.selectedAction)
			}

			// Check expected actions exist
			for _, key := range tt.expectedKeys {
				found := false
				for _, action := range vm.actions {
					if action.Key == key {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected action with key %s not found", key)
				}
			}

			// Check unexpected actions don't exist
			for _, key := range tt.unexpectedKeys {
				for _, action := range vm.actions {
					if action.Key == key {
						t.Errorf("unexpected action with key %s found", key)
					}
				}
			}

			// Always available actions
			expectedAlways := []string{"Enter", "B", "P", "L"}
			for _, key := range expectedAlways {
				found := false
				for _, action := range vm.actions {
					if action.Key == key {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected always-available action with key %s not found", key)
				}
			}
		})
	}
}

func TestComposeProjectActionViewModel_HandleUp(t *testing.T) {
	vm := &ComposeProjectActionViewModel{
		actions: []ComposeProjectAction{
			{Key: "1", Name: "Action 1"},
			{Key: "2", Name: "Action 2"},
			{Key: "3", Name: "Action 3"},
		},
		selectedAction: 2,
	}

	vm.HandleUp()
	if vm.selectedAction != 1 {
		t.Errorf("HandleUp() selectedAction = %d, want 1", vm.selectedAction)
	}

	vm.HandleUp()
	if vm.selectedAction != 0 {
		t.Errorf("HandleUp() selectedAction = %d, want 0", vm.selectedAction)
	}

	// Should not go below 0
	vm.HandleUp()
	if vm.selectedAction != 0 {
		t.Errorf("HandleUp() selectedAction = %d, want 0", vm.selectedAction)
	}
}

func TestComposeProjectActionViewModel_HandleDown(t *testing.T) {
	vm := &ComposeProjectActionViewModel{
		actions: []ComposeProjectAction{
			{Key: "1", Name: "Action 1"},
			{Key: "2", Name: "Action 2"},
			{Key: "3", Name: "Action 3"},
		},
		selectedAction: 0,
	}

	vm.HandleDown()
	if vm.selectedAction != 1 {
		t.Errorf("HandleDown() selectedAction = %d, want 1", vm.selectedAction)
	}

	vm.HandleDown()
	if vm.selectedAction != 2 {
		t.Errorf("HandleDown() selectedAction = %d, want 2", vm.selectedAction)
	}

	// Should not go beyond last action
	vm.HandleDown()
	if vm.selectedAction != 2 {
		t.Errorf("HandleDown() selectedAction = %d, want 2", vm.selectedAction)
	}
}

func TestComposeProjectActionViewModel_HandleSelect(t *testing.T) {
	tests := []struct {
		name           string
		selectedAction int
		actionsCount   int
		expectCall     bool
	}{
		{
			name:           "executes selected action and switches to previous view",
			selectedAction: 0,
			actionsCount:   3,
			expectCall:     true,
		},
		{
			name:           "handles out of bounds selection",
			selectedAction: 5,
			actionsCount:   3,
			expectCall:     false,
		},
		{
			name:           "handles negative selection",
			selectedAction: -1,
			actionsCount:   3,
			expectCall:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			model := &Model{}
			model.initializeKeyHandlers()
			model.viewHistory = []ViewType{ComposeProjectListView}
			model.currentView = ComposeProjectActionView

			vm := &ComposeProjectActionViewModel{
				selectedAction: tt.selectedAction,
				targetProject: &models.ComposeProject{
					Name:        "test-project",
					Status:      "running(2)",
					ConfigFiles: "/path/to/docker-compose.yml",
				},
			}

			for i := 0; i < tt.actionsCount; i++ {
				vm.actions = append(vm.actions, ComposeProjectAction{
					Key:  string(rune('A' + i)),
					Name: "Test Action",
					Handler: func(m *Model, p models.ComposeProject) tea.Cmd {
						called = true
						return nil
					},
				})
			}

			vm.HandleSelect(model)

			if called != tt.expectCall {
				t.Errorf("Handler called = %v, want %v", called, tt.expectCall)
			}

			if tt.expectCall && model.currentView != ComposeProjectListView {
				t.Errorf("Did not switch to previous view")
			}
		})
	}
}

func TestComposeProjectActionViewModel_HandleBack(t *testing.T) {
	model := &Model{
		viewHistory: []ViewType{ComposeProjectListView},
		currentView: ComposeProjectActionView,
	}

	vm := &ComposeProjectActionViewModel{}
	vm.HandleBack(model)

	if model.currentView != ComposeProjectListView {
		t.Errorf("HandleBack() currentView = %v, want ComposeProjectListView", model.currentView)
	}
}

func TestComposeProjectActionViewModel_Render(t *testing.T) {
	tests := []struct {
		name            string
		targetProject   *models.ComposeProject
		expectedContent []string
	}{
		{
			name: "renders with project info",
			targetProject: &models.ComposeProject{
				Name:        "test-project",
				Status:      "running(2)",
				ConfigFiles: "/path/to/docker-compose.yml",
			},
			expectedContent: []string{
				"Select Action for test-project",
				"Project: test-project",
				"Status: running(2)",
				"Config Files: /path/to/docker-compose.yml",
				"Available Actions:",
			},
		},
		{
			name:            "renders no project selected",
			targetProject:   nil,
			expectedContent: []string{"No project selected"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &Model{width: 80, Height: 24}
			vm := &ComposeProjectActionViewModel{
				targetProject: tt.targetProject,
			}

			if tt.targetProject != nil {
				vm.Initialize(tt.targetProject)
			}

			output := vm.render(model)

			for _, expected := range tt.expectedContent {
				if !strings.Contains(output, expected) {
					t.Errorf("render() output missing expected content: %s", expected)
				}
			}
		})
	}
}

func TestComposeProjectActionViewModel_ActionColoring(t *testing.T) {
	model := &Model{width: 80, Height: 24}
	project := &models.ComposeProject{
		Name:        "test-project",
		Status:      "running(2)",
		ConfigFiles: "/path/to/docker-compose.yml",
	}

	vm := &ComposeProjectActionViewModel{}
	vm.Initialize(project)

	output := vm.render(model)

	// Check that aggressive actions have different styling
	// The actual ANSI color codes will be in the output
	if !strings.Contains(output, "[D] Down") {
		t.Error("Down action not found in output")
	}
	if !strings.Contains(output, "[Enter] View Containers") {
		t.Error("View Containers action not found in output")
	}
}
