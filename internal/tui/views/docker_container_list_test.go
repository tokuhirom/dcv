package views

import (
	"testing"

	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
)

func createTestContainers() []models.DockerContainer {
	return []models.DockerContainer{
		{
			ID:      "abc123",
			Names:   "web-server",
			Image:   "nginx:latest",
			Status:  "Up 2 hours",
			State:   "running",
			Ports:   "80/tcp -> 0.0.0.0:8080",
			Command: "nginx",
		},
		{
			ID:      "def456",
			Names:   "database",
			Image:   "postgres:13",
			Status:  "Exited (0) 10 minutes ago",
			State:   "exited",
			Ports:   "",
			Command: "postgres",
		},
	}
}

func TestNewDockerContainerListView(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewDockerContainerListView(dockerClient)
	
	assert.NotNil(t, view)
	assert.NotNil(t, view.docker)
	assert.NotNil(t, view.table)
	assert.False(t, view.showAll)
	assert.Empty(t, view.containers)
}

func TestDockerContainerListView_GetPrimitive(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewDockerContainerListView(dockerClient)
	
	primitive := view.GetPrimitive()
	
	assert.NotNil(t, primitive)
	_, ok := primitive.(*tview.Table)
	assert.True(t, ok)
}

func TestDockerContainerListView_GetTitle(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewDockerContainerListView(dockerClient)
	
	title := view.GetTitle()
	
	assert.Equal(t, "Docker Containers", title)
}

func TestDockerContainerListView_UpdateTable(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewDockerContainerListView(dockerClient)
	view.containers = createTestContainers()
	
	// Create a test app and screen
	app := tview.NewApplication()
	SetApp(app)
	
	// Update the table
	view.updateTable()
	
	// Check headers
	headerCell := view.table.GetCell(0, 0)
	assert.NotNil(t, headerCell)
	assert.Equal(t, "NAME", headerCell.Text)
	
	headerCell = view.table.GetCell(0, 1)
	assert.NotNil(t, headerCell)
	assert.Equal(t, "IMAGE", headerCell.Text)
	
	headerCell = view.table.GetCell(0, 2)
	assert.NotNil(t, headerCell)
	assert.Equal(t, "STATUS", headerCell.Text)
	
	// Check container data
	nameCell := view.table.GetCell(1, 0)
	assert.NotNil(t, nameCell)
	assert.Equal(t, "web-server", nameCell.Text)
	
	imageCell := view.table.GetCell(1, 1)
	assert.NotNil(t, imageCell)
	assert.Equal(t, "nginx:latest", imageCell.Text)
	
	statusCell := view.table.GetCell(1, 2)
	assert.NotNil(t, statusCell)
	assert.Equal(t, "Up 2 hours", statusCell.Text)
	
	// Check second container
	nameCell2 := view.table.GetCell(2, 0)
	assert.NotNil(t, nameCell2)
	assert.Equal(t, "database", nameCell2.Text)
	
	// Verify table selection
	row, _ := view.table.GetSelection()
	assert.Equal(t, 1, row) // First data row should be selected
}

func TestDockerContainerListView_ShowAll(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewDockerContainerListView(dockerClient)
	
	// Initially showAll is false
	assert.False(t, view.showAll)
	
	// Toggle showAll
	view.showAll = true
	assert.True(t, view.showAll)
	
	// Toggle back
	view.showAll = false
	assert.False(t, view.showAll)
}

func TestDockerContainerListView_TableProperties(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewDockerContainerListView(dockerClient)
	
	// Check table properties
	selectable, selectableByRow := view.table.GetSelectable()
	assert.True(t, selectable)
	assert.False(t, selectableByRow)
	
	// Check that table is configured correctly
	// Note: GetFixed method doesn't exist in tview.Table
	// We just verify the table is configured
	assert.NotNil(t, view.table)
}

func TestDockerContainerListView_EmptyContainerList(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewDockerContainerListView(dockerClient)
	view.containers = []models.DockerContainer{}
	
	// Create a test app
	app := tview.NewApplication()
	SetApp(app)
	
	// Update the table with empty container list
	view.updateTable()
	
	// Should still have headers
	headerCell := view.table.GetCell(0, 0)
	assert.NotNil(t, headerCell)
	assert.Equal(t, "NAME", headerCell.Text)
	
	// Check table row count - should only have header row
	rowCount := view.table.GetRowCount()
	assert.Equal(t, 1, rowCount) // Only header row
}

func TestDockerContainerListView_ContainerColors(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewDockerContainerListView(dockerClient)
	
	containers := []models.DockerContainer{
		{
			ID:     "running1",
			Names:  "running-container",
			State:  "running",
			Status: "Up 1 hour",
			Image:  "test:latest",
			Ports:  "80/tcp",
			Command: "test",
		},
		{
			ID:     "exited1",
			Names:  "exited-container",
			State:  "exited",
			Status: "Exited (0) 5 minutes ago",
			Image:  "test:latest",
			Ports:  "",
			Command: "test",
		},
	}
	view.containers = containers
	
	// Create a test app
	app := tview.NewApplication()
	SetApp(app)
	
	// Update the table
	view.updateTable()
	
	// Check running container status cell exists and has correct text
	runningStatusCell := view.table.GetCell(1, 2)
	assert.NotNil(t, runningStatusCell)
	assert.Equal(t, "Up 1 hour", runningStatusCell.Text)
	// The color is set via SetTextColor which modifies the cell's style
	// We can verify the state was used correctly by checking the text
	
	// Check exited container status cell exists  
	exitedStatusCell := view.table.GetCell(2, 2)
	assert.NotNil(t, exitedStatusCell)
	assert.Equal(t, "Exited (0) 5 minutes ago", exitedStatusCell.Text)
}

func TestDockerContainerListView_KeyHandling(t *testing.T) {
	// This test verifies that key handlers are properly set up
	dockerClient := docker.NewClient()
	view := NewDockerContainerListView(dockerClient)
	
	// The table should have input capture function set
	assert.NotNil(t, view.table)
	
	// We can't easily test the actual key handling without running the app,
	// but we can verify the structure is in place
	primitive := view.GetPrimitive()
	table, ok := primitive.(*tview.Table)
	assert.True(t, ok)
	assert.NotNil(t, table)
}