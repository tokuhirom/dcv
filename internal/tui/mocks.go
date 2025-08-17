package tui

import (
	"github.com/rivo/tview"
	
	"github.com/tokuhirom/dcv/internal/models"
	"github.com/tokuhirom/dcv/internal/tui/views"
)

// MockView is a mock implementation of the View interface
type MockView struct {
	Title          string
	RefreshCalled  int
	Primitive      tview.Primitive
}

// NewMockView creates a new mock view
func NewMockView(title string) *MockView {
	return &MockView{
		Title:     title,
		Primitive: tview.NewBox(),
	}
}

// GetPrimitive returns the mock primitive
func (m *MockView) GetPrimitive() tview.Primitive {
	return m.Primitive
}

// Refresh increments the refresh counter
func (m *MockView) Refresh() {
	m.RefreshCalled++
}

// GetTitle returns the mock title
func (m *MockView) GetTitle() string {
	return m.Title
}

// MockDockerClient is a mock implementation of Docker operations
type MockDockerClient struct {
	Containers       []models.DockerContainer
	Images           []models.DockerImage
	Networks         []models.DockerNetwork
	Volumes          []models.DockerVolume
	Projects         []models.ComposeProject
	ComposeContainers []models.ComposeContainer
	
	ListContainersCalled int
	StopCalled           int
	StartCalled          int
	KillCalled           int
	RemoveCalled         int
	
	LastStoppedID  string
	LastStartedID  string
	LastKilledID   string
	LastRemovedID  string
	
	ShouldError    bool
	ErrorMessage   string
}

// NewMockDockerClient creates a new mock Docker client
func NewMockDockerClient() *MockDockerClient {
	return &MockDockerClient{
		Containers: CreateTestContainers(),
		Images:     CreateTestImages(),
		Networks:   CreateTestNetworks(),
		Volumes:    CreateTestVolumes(),
		Projects:   CreateTestProjects(),
	}
}

// ListContainers returns mock containers
func (m *MockDockerClient) ListContainers(showAll bool) ([]models.DockerContainer, error) {
	m.ListContainersCalled++
	if m.ShouldError {
		return nil, &MockError{Message: m.ErrorMessage}
	}
	
	if !showAll {
		// Filter only running containers
		var running []models.DockerContainer
		for _, c := range m.Containers {
			if c.State == "running" {
				running = append(running, c)
			}
		}
		return running, nil
	}
	return m.Containers, nil
}

// StopContainer mocks stopping a container
func (m *MockDockerClient) StopContainer(containerID string) error {
	m.StopCalled++
	m.LastStoppedID = containerID
	if m.ShouldError {
		return &MockError{Message: m.ErrorMessage}
	}
	return nil
}

// StartContainer mocks starting a container
func (m *MockDockerClient) StartContainer(containerID string) error {
	m.StartCalled++
	m.LastStartedID = containerID
	if m.ShouldError {
		return &MockError{Message: m.ErrorMessage}
	}
	return nil
}

// KillContainer mocks killing a container
func (m *MockDockerClient) KillContainer(containerID string) error {
	m.KillCalled++
	m.LastKilledID = containerID
	if m.ShouldError {
		return &MockError{Message: m.ErrorMessage}
	}
	return nil
}

// RemoveContainer mocks removing a container
func (m *MockDockerClient) RemoveContainer(containerID string) error {
	m.RemoveCalled++
	m.LastRemovedID = containerID
	if m.ShouldError {
		return &MockError{Message: m.ErrorMessage}
	}
	return nil
}

// MockError is a mock error type
type MockError struct {
	Message string
}

func (e *MockError) Error() string {
	return e.Message
}

// MockApp is a mock implementation of the App
type MockApp struct {
	CurrentView      views.View
	ViewSwitchCount  int
	LastSwitchedView string
	StopCalled       bool
}

// NewMockApp creates a new mock app
func NewMockApp() *MockApp {
	return &MockApp{}
}

// SwitchView mocks view switching
func (m *MockApp) SwitchView(viewName string) {
	m.ViewSwitchCount++
	m.LastSwitchedView = viewName
}

// Stop mocks stopping the app
func (m *MockApp) Stop() {
	m.StopCalled = true
}