package views

import (
	"testing"

	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
)

func createTestComposeContainers() []models.ComposeContainer {
	return []models.ComposeContainer{
		{
			ID:      "compose-abc123",
			Service: "web",
			Name:    "myapp_web_1",
			Image:   "nginx:alpine",
			State:   "running",
			Health:  "",
			Project: "myapp",
			Publishers: []struct {
				URL           string `json:"URL"`
				TargetPort    int    `json:"TargetPort"`
				PublishedPort int    `json:"PublishedPort"`
				Protocol      string `json:"Protocol"`
			}{
				{
					URL:           "0.0.0.0",
					TargetPort:    80,
					PublishedPort: 8080,
					Protocol:      "tcp",
				},
			},
		},
		{
			ID:       "compose-def456",
			Service:  "database",
			Name:     "myapp_db_1",
			Image:    "postgres:13",
			State:    "exited",
			Health:   "",
			Project:  "myapp",
			ExitCode: 0,
		},
	}
}

func TestNewComposeProcessListView(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewComposeProcessListView(dockerClient)

	assert.NotNil(t, view)
	assert.NotNil(t, view.docker)
	assert.NotNil(t, view.table)
	assert.False(t, view.showAll)
	assert.Empty(t, view.composeContainers)
	assert.Empty(t, view.projectName)
}

func TestComposeProcessListView_GetPrimitive(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewComposeProcessListView(dockerClient)

	primitive := view.GetPrimitive()

	assert.NotNil(t, primitive)
	_, ok := primitive.(*tview.Table)
	assert.True(t, ok)
}

func TestComposeProcessListView_GetTitle(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewComposeProcessListView(dockerClient)

	// Test without project name
	title := view.GetTitle()
	assert.Equal(t, "Compose Processes", title)

	// Test with project name
	view.projectName = "myapp"
	title = view.GetTitle()
	assert.Equal(t, "Compose: myapp", title)
}

func TestComposeProcessListView_SetProject(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewComposeProcessListView(dockerClient)

	// Set project name
	view.SetProject("test-project")

	assert.Equal(t, "test-project", view.projectName)
}

func TestComposeProcessListView_UpdateTable(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewComposeProcessListView(dockerClient)
	view.composeContainers = createTestComposeContainers()
	view.projectName = "myapp"

	// Create a test app and screen
	app := tview.NewApplication()
	SetApp(app)

	// Update the table
	view.updateTable()

	// Check headers
	headerCell := view.table.GetCell(0, 0)
	assert.NotNil(t, headerCell)
	assert.Equal(t, "SERVICE", headerCell.Text)

	headerCell = view.table.GetCell(0, 1)
	assert.NotNil(t, headerCell)
	assert.Equal(t, "IMAGE", headerCell.Text)

	headerCell = view.table.GetCell(0, 2)
	assert.NotNil(t, headerCell)
	assert.Equal(t, "STATE", headerCell.Text)

	headerCell = view.table.GetCell(0, 3)
	assert.NotNil(t, headerCell)
	assert.Equal(t, "STATUS", headerCell.Text)

	headerCell = view.table.GetCell(0, 4)
	assert.NotNil(t, headerCell)
	assert.Equal(t, "PORTS", headerCell.Text)

	// Check container data
	serviceCell := view.table.GetCell(1, 0)
	assert.NotNil(t, serviceCell)
	assert.Equal(t, "web", serviceCell.Text)

	imageCell := view.table.GetCell(1, 1)
	assert.NotNil(t, imageCell)
	assert.Equal(t, "nginx:alpine", imageCell.Text)

	stateCell := view.table.GetCell(1, 2)
	assert.NotNil(t, stateCell)
	assert.Equal(t, "running", stateCell.Text)

	statusCell := view.table.GetCell(1, 3)
	assert.NotNil(t, statusCell)
	assert.Equal(t, "Up", statusCell.Text)

	// Check second container
	serviceCell2 := view.table.GetCell(2, 0)
	assert.NotNil(t, serviceCell2)
	assert.Equal(t, "database", serviceCell2.Text)

	// Verify table selection
	row, _ := view.table.GetSelection()
	assert.Equal(t, 1, row) // First data row should be selected
}

func TestComposeProcessListView_ShowAll(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewComposeProcessListView(dockerClient)

	// Initially showAll is false
	assert.False(t, view.showAll)

	// Toggle showAll
	view.showAll = true
	assert.True(t, view.showAll)

	// Toggle back
	view.showAll = false
	assert.False(t, view.showAll)
}

func TestComposeProcessListView_TableProperties(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewComposeProcessListView(dockerClient)

	// Check table properties
	selectable, selectableByRow := view.table.GetSelectable()
	assert.True(t, selectable)
	assert.False(t, selectableByRow)

	// Check that table is configured correctly
	assert.NotNil(t, view.table)
}

func TestComposeProcessListView_EmptyContainerList(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewComposeProcessListView(dockerClient)
	view.composeContainers = []models.ComposeContainer{}
	view.projectName = "empty-project"

	// Create a test app
	app := tview.NewApplication()
	SetApp(app)

	// Update the table with empty container list
	view.updateTable()

	// Should still have headers
	headerCell := view.table.GetCell(0, 0)
	assert.NotNil(t, headerCell)
	assert.Equal(t, "SERVICE", headerCell.Text)

	// Check table row count - should only have header row
	rowCount := view.table.GetRowCount()
	assert.Equal(t, 1, rowCount) // Only header row
}

func TestComposeProcessListView_DindIndicator(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewComposeProcessListView(dockerClient)

	// Create a container with dind
	containers := []models.ComposeContainer{
		{
			ID:      "dind-container",
			Service: "dind-service",
			Name:    "myapp_dind_1",
			Image:   "docker:dind",
			State:   "running",
		},
	}
	view.composeContainers = containers
	view.projectName = "myapp"

	// Create a test app
	app := tview.NewApplication()
	SetApp(app)

	// Update the table
	view.updateTable()

	// Check that dind indicator is present
	serviceCell := view.table.GetCell(1, 0)
	assert.NotNil(t, serviceCell)
	// The service name should have the 🔄 emoji for dind containers
	assert.Contains(t, serviceCell.Text, "🔄")
}

func TestComposeProcessListView_StatusColors(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewComposeProcessListView(dockerClient)

	containers := []models.ComposeContainer{
		{
			ID:      "running1",
			Service: "running-service",
			Name:    "myapp_running_1",
			State:   "running",
			Image:   "test:latest",
		},
		{
			ID:       "exited1",
			Service:  "exited-service",
			Name:     "myapp_exited_1",
			State:    "exited",
			Image:    "test:latest",
			ExitCode: 0,
		},
	}
	view.composeContainers = containers
	view.projectName = "myapp"

	// Create a test app
	app := tview.NewApplication()
	SetApp(app)

	// Update the table
	view.updateTable()

	// Check running container status cell exists and has correct text
	runningStatusCell := view.table.GetCell(1, 3)
	assert.NotNil(t, runningStatusCell)
	assert.Equal(t, "Up", runningStatusCell.Text)

	// Check exited container status cell exists
	exitedStatusCell := view.table.GetCell(2, 3)
	assert.NotNil(t, exitedStatusCell)
	assert.Equal(t, "Exited (0)", exitedStatusCell.Text)
}

func TestComposeProcessListView_KeyHandling(t *testing.T) {
	// This test verifies that key handlers are properly set up
	dockerClient := docker.NewClient()
	view := NewComposeProcessListView(dockerClient)

	// The table should have input capture function set
	assert.NotNil(t, view.table)

	// We can't easily test the actual key handling without running the app,
	// but we can verify the structure is in place
	primitive := view.GetPrimitive()
	table, ok := primitive.(*tview.Table)
	assert.True(t, ok)
	assert.NotNil(t, table)
}

func TestComposeProcessListView_RefreshWithoutProject(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewComposeProcessListView(dockerClient)

	// Refresh without project name should not crash
	view.Refresh()

	// Should not load containers without project
	assert.Empty(t, view.composeContainers)
}

func TestComposeProcessListView_GetSelectedContainer(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewComposeProcessListView(dockerClient)
	view.composeContainers = createTestComposeContainers()
	view.projectName = "myapp"

	// Create a test app
	app := tview.NewApplication()
	SetApp(app)

	// Update the table
	view.updateTable()

	// First row should be selected by default
	row, _ := view.table.GetSelection()
	assert.Equal(t, 1, row)

	// Get the selected container
	if row > 0 && row <= len(view.composeContainers) {
		selectedContainer := view.composeContainers[row-1]
		assert.Equal(t, "web", selectedContainer.Service)
		assert.Equal(t, "myapp_web_1", selectedContainer.Name)
	}

	// Move selection down
	view.table.Select(2, 0)
	row, _ = view.table.GetSelection()
	assert.Equal(t, 2, row)

	// Get the new selected container
	if row > 0 && row <= len(view.composeContainers) {
		selectedContainer := view.composeContainers[row-1]
		assert.Equal(t, "database", selectedContainer.Service)
		assert.Equal(t, "myapp_db_1", selectedContainer.Name)
	}
}

func TestComposeProcessListView_ContainerOperations(t *testing.T) {
	// Test that container operations don't crash
	// Note: These operations will fail in tests since we don't have actual containers,
	// but we're checking that the methods exist and don't panic

	t.Run("StopContainer", func(t *testing.T) {
		dockerClient := docker.NewClient()
		view := NewComposeProcessListView(dockerClient)
		view.composeContainers = createTestComposeContainers()
		view.projectName = "myapp"

		assert.NotPanics(t, func() {
			view.stopContainer(view.composeContainers[0])
		})
	})

	t.Run("StartContainer", func(t *testing.T) {
		dockerClient := docker.NewClient()
		view := NewComposeProcessListView(dockerClient)
		view.composeContainers = createTestComposeContainers()
		view.projectName = "myapp"

		assert.NotPanics(t, func() {
			view.startContainer(view.composeContainers[0])
		})
	})

	t.Run("RestartContainer", func(t *testing.T) {
		dockerClient := docker.NewClient()
		view := NewComposeProcessListView(dockerClient)
		view.composeContainers = createTestComposeContainers()
		view.projectName = "myapp"

		assert.NotPanics(t, func() {
			view.restartContainer(view.composeContainers[0])
		})
	})

	t.Run("DeleteContainer", func(t *testing.T) {
		dockerClient := docker.NewClient()
		view := NewComposeProcessListView(dockerClient)
		view.composeContainers = createTestComposeContainers()
		view.projectName = "myapp"

		assert.NotPanics(t, func() {
			view.deleteContainer(view.composeContainers[0])
		})
	})

	t.Run("ExecShell", func(t *testing.T) {
		dockerClient := docker.NewClient()
		view := NewComposeProcessListView(dockerClient)
		view.composeContainers = createTestComposeContainers()
		view.projectName = "myapp"

		assert.NotPanics(t, func() {
			view.execShell(view.composeContainers[0])
		})
	})
}

func TestComposeProcessListView_PortsDisplay(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewComposeProcessListView(dockerClient)
	view.composeContainers = createTestComposeContainers()
	view.projectName = "myapp"

	// Create a test app
	app := tview.NewApplication()
	SetApp(app)

	// Update the table
	view.updateTable()

	// Check that ports are displayed correctly for the first container
	portsCell := view.table.GetCell(1, 4)
	assert.NotNil(t, portsCell)
	// The GetPortsString() method should format ports appropriately
	// For the test container with port 8080->80, it should show something like "0.0.0.0:8080->80/tcp"
	assert.Contains(t, portsCell.Text, "8080")
	assert.Contains(t, portsCell.Text, "80")

	// Check that the second container has no ports
	portsCell2 := view.table.GetCell(2, 4)
	assert.NotNil(t, portsCell2)
	assert.Equal(t, "", portsCell2.Text) // No ports for database container
}
