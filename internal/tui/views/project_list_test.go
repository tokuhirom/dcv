package views

import (
	"testing"

	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
)

func createTestComposeProjects() []models.ComposeProject {
	return []models.ComposeProject{
		{
			Name:        "myapp",
			Status:      "running(3)",
			ConfigFiles: "/home/user/myapp/docker-compose.yml",
		},
		{
			Name:        "testproject",
			Status:      "exited(2)",
			ConfigFiles: "/home/user/test/docker-compose.yml,/home/user/test/docker-compose.override.yml",
		},
		{
			Name:        "devstack",
			Status:      "running(5)",
			ConfigFiles: "/opt/devstack/compose.yaml",
		},
	}
}

func TestNewProjectListView(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewProjectListView(dockerClient)

	assert.NotNil(t, view)
	assert.NotNil(t, view.docker)
	assert.NotNil(t, view.table)
	assert.Empty(t, view.projects)
	assert.Nil(t, view.onProjectSelected)
}

func TestProjectListView_GetPrimitive(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewProjectListView(dockerClient)

	primitive := view.GetPrimitive()

	assert.NotNil(t, primitive)
	_, ok := primitive.(*tview.Pages)
	assert.True(t, ok)
}

func TestProjectListView_GetTitle(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewProjectListView(dockerClient)

	title := view.GetTitle()
	assert.Equal(t, "Docker Compose Projects", title)
}

func TestProjectListView_UpdateTable(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewProjectListView(dockerClient)
	view.projects = createTestComposeProjects()

	// Create a test app
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
	assert.Equal(t, "STATUS", headerCell.Text)

	headerCell = view.table.GetCell(0, 2)
	assert.NotNil(t, headerCell)
	assert.Equal(t, "CONFIG FILES", headerCell.Text)

	// Check project data - first project (myapp)
	nameCell := view.table.GetCell(1, 0)
	assert.NotNil(t, nameCell)
	assert.Equal(t, "myapp", nameCell.Text)

	statusCell := view.table.GetCell(1, 1)
	assert.NotNil(t, statusCell)
	assert.Equal(t, "running(3)", statusCell.Text)

	configCell := view.table.GetCell(1, 2)
	assert.NotNil(t, configCell)
	assert.Equal(t, "/home/user/myapp/docker-compose.yml", configCell.Text)

	// Check second project (testproject)
	nameCell2 := view.table.GetCell(2, 0)
	assert.NotNil(t, nameCell2)
	assert.Equal(t, "testproject", nameCell2.Text)

	statusCell2 := view.table.GetCell(2, 1)
	assert.NotNil(t, statusCell2)
	assert.Equal(t, "exited(2)", statusCell2.Text)

	// Check third project (devstack)
	nameCell3 := view.table.GetCell(3, 0)
	assert.NotNil(t, nameCell3)
	assert.Equal(t, "devstack", nameCell3.Text)

	// Verify table selection
	row, _ := view.table.GetSelection()
	assert.Equal(t, 1, row) // First data row should be selected
}

func TestProjectListView_TableProperties(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewProjectListView(dockerClient)

	// Check table properties
	selectable, selectableByRow := view.table.GetSelectable()
	assert.True(t, selectable)
	assert.False(t, selectableByRow)

	// Check that table is configured correctly
	assert.NotNil(t, view.table)
}

func TestProjectListView_EmptyProjectList(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewProjectListView(dockerClient)
	view.projects = []models.ComposeProject{}

	// Create a test app
	app := tview.NewApplication()
	SetApp(app)

	// Update the table with empty project list
	view.updateTable()

	// Should still have headers
	headerCell := view.table.GetCell(0, 0)
	assert.NotNil(t, headerCell)
	assert.Equal(t, "NAME", headerCell.Text)

	// Check table row count - should only have header row
	rowCount := view.table.GetRowCount()
	assert.Equal(t, 1, rowCount) // Only header row
}

func TestProjectListView_GetSelectedProject(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewProjectListView(dockerClient)
	view.projects = createTestComposeProjects()

	// Create a test app
	app := tview.NewApplication()
	SetApp(app)

	// Update the table
	view.updateTable()

	// First row should be selected by default
	row, _ := view.table.GetSelection()
	assert.Equal(t, 1, row)

	// Get the selected project
	selectedProject := view.GetSelectedProject()
	assert.NotNil(t, selectedProject)
	assert.Equal(t, "myapp", selectedProject.Name)
	assert.Equal(t, "running(3)", selectedProject.Status)

	// Move selection down
	view.table.Select(2, 0)
	row, _ = view.table.GetSelection()
	assert.Equal(t, 2, row)

	// Get the new selected project
	selectedProject = view.GetSelectedProject()
	assert.NotNil(t, selectedProject)
	assert.Equal(t, "testproject", selectedProject.Name)
	assert.Equal(t, "exited(2)", selectedProject.Status)
}

func TestProjectListView_SearchProjects(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewProjectListView(dockerClient)
	view.projects = createTestComposeProjects()

	// Create a test app
	app := tview.NewApplication()
	SetApp(app)

	// Search for "myapp"
	view.SearchProjects("myapp")

	// Search for "running"
	view.SearchProjects("running")

	// Test empty search (should reset)
	view.SearchProjects("")

	// Verify original projects are still intact
	assert.Len(t, view.projects, 3)
}

func TestProjectListView_ProjectOperations(t *testing.T) {
	// Test that project operations don't crash
	// Note: These operations will fail in tests since we don't have actual projects,
	// but we're checking that the methods exist and don't panic

	projects := createTestComposeProjects()

	t.Run("StopProject", func(t *testing.T) {
		dockerClient := docker.NewClient()
		view := NewProjectListView(dockerClient)
		view.projects = projects

		assert.NotPanics(t, func() {
			view.stopProject(projects[0])
		})
	})

	t.Run("StartProject", func(t *testing.T) {
		dockerClient := docker.NewClient()
		view := NewProjectListView(dockerClient)
		view.projects = projects

		assert.NotPanics(t, func() {
			view.startProject(projects[0])
		})
	})

	t.Run("DownProject", func(t *testing.T) {
		dockerClient := docker.NewClient()
		view := NewProjectListView(dockerClient)
		view.projects = projects

		assert.NotPanics(t, func() {
			view.downProject(projects[0])
		})
	})

	t.Run("RemoveProject", func(t *testing.T) {
		dockerClient := docker.NewClient()
		view := NewProjectListView(dockerClient)
		view.projects = projects

		assert.NotPanics(t, func() {
			view.removeProject(projects[0])
		})
	})

	t.Run("UpProject", func(t *testing.T) {
		dockerClient := docker.NewClient()
		view := NewProjectListView(dockerClient)
		view.projects = projects

		assert.NotPanics(t, func() {
			view.upProject(projects[0])
		})
	})

	t.Run("UpProjectDetached", func(t *testing.T) {
		dockerClient := docker.NewClient()
		view := NewProjectListView(dockerClient)
		view.projects = projects

		assert.NotPanics(t, func() {
			view.upProjectDetached(projects[0])
		})
	})

	t.Run("RestartProject", func(t *testing.T) {
		dockerClient := docker.NewClient()
		view := NewProjectListView(dockerClient)
		view.projects = projects

		assert.NotPanics(t, func() {
			view.restartProject(projects[0])
		})
	})

	t.Run("PullProject", func(t *testing.T) {
		dockerClient := docker.NewClient()
		view := NewProjectListView(dockerClient)
		view.projects = projects

		assert.NotPanics(t, func() {
			view.pullProject(projects[0])
		})
	})

	t.Run("BuildProject", func(t *testing.T) {
		dockerClient := docker.NewClient()
		view := NewProjectListView(dockerClient)
		view.projects = projects

		assert.NotPanics(t, func() {
			view.buildProject(projects[0])
		})
	})
}

func TestProjectListView_KeyHandling(t *testing.T) {
	// This test verifies that key handlers are properly set up
	dockerClient := docker.NewClient()
	view := NewProjectListView(dockerClient)

	// The table should have input capture function set
	assert.NotNil(t, view.table)

	// We can't easily test the actual key handling without running the app,
	// but we can verify the structure is in place
	primitive := view.GetPrimitive()
	pages, ok := primitive.(*tview.Pages)
	assert.True(t, ok)
	assert.NotNil(t, pages)
}

func TestProjectListView_OnProjectSelectedCallback(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewProjectListView(dockerClient)

	// Test setting callback
	callbackCalled := false
	var selectedProject models.ComposeProject

	view.SetOnProjectSelected(func(project models.ComposeProject) {
		callbackCalled = true
		selectedProject = project
	})

	// The callback should be set
	assert.NotNil(t, view.onProjectSelected)

	// Test that callback gets called with correct project
	testProject := models.ComposeProject{
		Name:        "test",
		Status:      "running(1)",
		ConfigFiles: "test.yml",
	}
	view.onProjectSelected(testProject)

	assert.True(t, callbackCalled)
	assert.Equal(t, "test", selectedProject.Name)
}

func TestProjectListView_StatusColors(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewProjectListView(dockerClient)

	// Create projects with different statuses
	view.projects = []models.ComposeProject{
		{
			Name:        "running-project",
			Status:      "running(2)",
			ConfigFiles: "compose.yml",
		},
		{
			Name:        "exited-project",
			Status:      "exited(1)",
			ConfigFiles: "compose.yml",
		},
		{
			Name:        "stopped-project",
			Status:      "stopped",
			ConfigFiles: "compose.yml",
		},
	}

	// Create a test app
	app := tview.NewApplication()
	SetApp(app)

	// Update the table
	view.updateTable()

	// Check that status cells are created with appropriate colors
	// (We can't easily test the actual colors without more complex mocking,
	// but we can verify the cells exist)
	statusCell1 := view.table.GetCell(1, 1)
	assert.NotNil(t, statusCell1)
	assert.Equal(t, "running(2)", statusCell1.Text)

	statusCell2 := view.table.GetCell(2, 1)
	assert.NotNil(t, statusCell2)
	assert.Equal(t, "exited(1)", statusCell2.Text)

	statusCell3 := view.table.GetCell(3, 1)
	assert.NotNil(t, statusCell3)
	assert.Equal(t, "stopped", statusCell3.Text)
}

func TestProjectListView_GetProjectStatus(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewProjectListView(dockerClient)

	project := models.ComposeProject{
		Name:        "testproject",
		Status:      "running(2)",
		ConfigFiles: "compose.yml",
	}

	// This will likely fail in test environment but shouldn't panic
	_, err := view.GetProjectStatus(project)
	// We expect an error since the project doesn't exist
	_ = err
}

func TestProjectListView_GetProjectLogs(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewProjectListView(dockerClient)

	project := models.ComposeProject{
		Name:        "testproject",
		Status:      "running(2)",
		ConfigFiles: "compose.yml",
	}

	// Test getting logs with tail
	_, err := view.GetProjectLogs(project, 100)
	// We expect an error since the project doesn't exist
	_ = err

	// Test getting logs without tail
	_, err = view.GetProjectLogs(project, 0)
	// We expect an error since the project doesn't exist
	_ = err
}

func TestProjectListView_MultipleConfigFiles(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewProjectListView(dockerClient)

	// Project with multiple config files
	view.projects = []models.ComposeProject{
		{
			Name:        "multi-config",
			Status:      "running(3)",
			ConfigFiles: "docker-compose.yml,docker-compose.override.yml,docker-compose.prod.yml",
		},
	}

	// Create a test app
	app := tview.NewApplication()
	SetApp(app)

	// Update the table
	view.updateTable()

	// Check that the config files are displayed correctly
	configCell := view.table.GetCell(1, 2)
	assert.NotNil(t, configCell)
	assert.Equal(t, "docker-compose.yml,docker-compose.override.yml,docker-compose.prod.yml", configCell.Text)
}
