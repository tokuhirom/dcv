package views

import (
	"testing"

	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
)

func createTestDockerVolumes() []models.DockerVolume {
	return []models.DockerVolume{
		{
			Name:       "myapp_data",
			Driver:     "local",
			Scope:      "local",
			Mountpoint: "/var/lib/docker/volumes/myapp_data/_data",
			Labels:     "com.docker.compose.project=myapp,com.docker.compose.volume=data",
		},
		{
			Name:       "postgres_volume",
			Driver:     "local",
			Scope:      "local",
			Mountpoint: "/var/lib/docker/volumes/postgres_volume/_data",
			Labels:     "",
		},
		{
			Name:       "nfs_share",
			Driver:     "nfs",
			Scope:      "global",
			Mountpoint: "server:/path/to/share",
			Labels:     "backup=daily",
			Options: map[string]string{
				"type":   "nfs",
				"device": "server:/path/to/share",
			},
		},
	}
}

func TestNewVolumeListView(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewVolumeListView(dockerClient)

	assert.NotNil(t, view)
	assert.NotNil(t, view.docker)
	assert.NotNil(t, view.table)
	assert.Empty(t, view.dockerVolumes)
}

func TestVolumeListView_GetPrimitive(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewVolumeListView(dockerClient)

	primitive := view.GetPrimitive()

	assert.NotNil(t, primitive)
	_, ok := primitive.(*tview.Pages)
	assert.True(t, ok)
}

func TestVolumeListView_GetTitle(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewVolumeListView(dockerClient)

	title := view.GetTitle()
	assert.Equal(t, "Docker Volumes", title)
}

func TestVolumeListView_UpdateTable(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewVolumeListView(dockerClient)
	view.dockerVolumes = createTestDockerVolumes()

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
	assert.Equal(t, "DRIVER", headerCell.Text)

	headerCell = view.table.GetCell(0, 2)
	assert.NotNil(t, headerCell)
	assert.Equal(t, "SCOPE", headerCell.Text)

	headerCell = view.table.GetCell(0, 3)
	assert.NotNil(t, headerCell)
	assert.Equal(t, "MOUNTPOINT", headerCell.Text)

	headerCell = view.table.GetCell(0, 4)
	assert.NotNil(t, headerCell)
	assert.Equal(t, "LABELS", headerCell.Text)

	// Check volume data - first volume (myapp_data)
	nameCell := view.table.GetCell(1, 0)
	assert.NotNil(t, nameCell)
	assert.Equal(t, "myapp_data", nameCell.Text)

	driverCell := view.table.GetCell(1, 1)
	assert.NotNil(t, driverCell)
	assert.Equal(t, "local", driverCell.Text)

	scopeCell := view.table.GetCell(1, 2)
	assert.NotNil(t, scopeCell)
	assert.Equal(t, "local", scopeCell.Text)

	mountpointCell := view.table.GetCell(1, 3)
	assert.NotNil(t, mountpointCell)
	// Mountpoint is truncated to 50 chars
	assert.Contains(t, mountpointCell.Text, "myapp_data/_data")

	labelsCell := view.table.GetCell(1, 4)
	assert.NotNil(t, labelsCell)
	assert.Contains(t, labelsCell.Text, "project=myapp")
	assert.Contains(t, labelsCell.Text, "volume=data")

	// Check second volume (postgres_volume)
	nameCell2 := view.table.GetCell(2, 0)
	assert.NotNil(t, nameCell2)
	assert.Equal(t, "postgres_volume", nameCell2.Text)

	labelsCell2 := view.table.GetCell(2, 4)
	assert.NotNil(t, labelsCell2)
	assert.Equal(t, "", labelsCell2.Text) // No labels

	// Check third volume (nfs_share)
	nameCell3 := view.table.GetCell(3, 0)
	assert.NotNil(t, nameCell3)
	assert.Equal(t, "nfs_share", nameCell3.Text)

	driverCell3 := view.table.GetCell(3, 1)
	assert.NotNil(t, driverCell3)
	assert.Equal(t, "nfs", driverCell3.Text)

	scopeCell3 := view.table.GetCell(3, 2)
	assert.NotNil(t, scopeCell3)
	assert.Equal(t, "global", scopeCell3.Text)

	// Verify table selection
	row, _ := view.table.GetSelection()
	assert.Equal(t, 1, row) // First data row should be selected
}

func TestVolumeListView_TableProperties(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewVolumeListView(dockerClient)

	// Check table properties
	selectable, selectableByRow := view.table.GetSelectable()
	assert.True(t, selectable)
	assert.False(t, selectableByRow)

	// Check that table is configured correctly
	assert.NotNil(t, view.table)
}

func TestVolumeListView_EmptyVolumeList(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewVolumeListView(dockerClient)
	view.dockerVolumes = []models.DockerVolume{}

	// Create a test app
	app := tview.NewApplication()
	SetApp(app)

	// Update the table with empty volume list
	view.updateTable()

	// Should still have headers
	headerCell := view.table.GetCell(0, 0)
	assert.NotNil(t, headerCell)
	assert.Equal(t, "NAME", headerCell.Text)

	// Check table row count - should only have header row
	rowCount := view.table.GetRowCount()
	assert.Equal(t, 1, rowCount) // Only header row
}

func TestVolumeListView_GetSelectedVolume(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewVolumeListView(dockerClient)
	view.dockerVolumes = createTestDockerVolumes()

	// Create a test app
	app := tview.NewApplication()
	SetApp(app)

	// Update the table
	view.updateTable()

	// First row should be selected by default
	row, _ := view.table.GetSelection()
	assert.Equal(t, 1, row)

	// Get the selected volume
	selectedVolume := view.GetSelectedVolume()
	assert.NotNil(t, selectedVolume)
	assert.Equal(t, "myapp_data", selectedVolume.Name)
	assert.Equal(t, "local", selectedVolume.Driver)

	// Move selection down
	view.table.Select(2, 0)
	row, _ = view.table.GetSelection()
	assert.Equal(t, 2, row)

	// Get the new selected volume
	selectedVolume = view.GetSelectedVolume()
	assert.NotNil(t, selectedVolume)
	assert.Equal(t, "postgres_volume", selectedVolume.Name)
}

func TestVolumeListView_SearchVolumes(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewVolumeListView(dockerClient)
	view.dockerVolumes = createTestDockerVolumes()

	// Create a test app
	app := tview.NewApplication()
	SetApp(app)

	// Search for "myapp"
	view.SearchVolumes("myapp")

	// Search for "nfs"
	view.SearchVolumes("nfs")

	// Test empty search (should reset)
	view.SearchVolumes("")

	// Verify original volumes are still intact
	assert.Len(t, view.dockerVolumes, 3)
}

func TestVolumeListView_VolumeOperations(t *testing.T) {
	// Test that volume operations don't crash
	// Note: These operations will fail in tests since we don't have actual volumes,
	// but we're checking that the methods exist and don't panic

	volumes := createTestDockerVolumes()

	t.Run("DeleteVolume", func(t *testing.T) {
		dockerClient := docker.NewClient()
		view := NewVolumeListView(dockerClient)
		view.dockerVolumes = volumes

		assert.NotPanics(t, func() {
			view.deleteVolume(volumes[0], false)
		})
	})

	t.Run("ForceDeleteVolume", func(t *testing.T) {
		dockerClient := docker.NewClient()
		view := NewVolumeListView(dockerClient)
		view.dockerVolumes = volumes

		assert.NotPanics(t, func() {
			view.deleteVolume(volumes[0], true)
		})
	})

	t.Run("PruneVolumes", func(t *testing.T) {
		dockerClient := docker.NewClient()
		view := NewVolumeListView(dockerClient)

		assert.NotPanics(t, func() {
			view.pruneVolumes()
		})
	})

	t.Run("PruneAllVolumes", func(t *testing.T) {
		dockerClient := docker.NewClient()
		view := NewVolumeListView(dockerClient)

		assert.NotPanics(t, func() {
			view.pruneAllVolumes()
		})
	})

	t.Run("CreateVolume", func(t *testing.T) {
		dockerClient := docker.NewClient()
		view := NewVolumeListView(dockerClient)

		// Test creating a volume (will fail but shouldn't panic)
		labels := map[string]string{
			"test": "true",
		}
		err := view.CreateVolume("test-volume", "local", labels)
		// We expect success or error depending on Docker state
		_ = err
	})

	t.Run("InspectVolume", func(t *testing.T) {
		dockerClient := docker.NewClient()
		view := NewVolumeListView(dockerClient)

		// Test inspecting a volume (will fail but shouldn't panic)
		_, err := view.InspectVolume(volumes[0])
		// We expect an error since the volume doesn't exist
		_ = err
	})

	t.Run("GetVolumeSize", func(t *testing.T) {
		dockerClient := docker.NewClient()
		view := NewVolumeListView(dockerClient)

		// Test getting volume size (will fail but shouldn't panic)
		_, err := view.GetVolumeSize(volumes[0])
		// We expect an error since the volume doesn't exist
		_ = err

		// Test non-local volume
		size, err := view.GetVolumeSize(volumes[2]) // nfs volume
		assert.NoError(t, err)
		assert.Equal(t, "N/A", size)
	})
}

func TestVolumeListView_KeyHandling(t *testing.T) {
	// This test verifies that key handlers are properly set up
	dockerClient := docker.NewClient()
	view := NewVolumeListView(dockerClient)

	// The table should have input capture function set
	assert.NotNil(t, view.table)

	// We can't easily test the actual key handling without running the app,
	// but we can verify the structure is in place
	primitive := view.GetPrimitive()
	pages, ok := primitive.(*tview.Pages)
	assert.True(t, ok)
	assert.NotNil(t, pages)
}

func TestVolumeListView_VolumeLabels(t *testing.T) {
	volumes := createTestDockerVolumes()

	// Test GetLabel for compose volume
	projectName := volumes[0].GetLabel("com.docker.compose.project")
	assert.Equal(t, "myapp", projectName)

	volumeName := volumes[0].GetLabel("com.docker.compose.volume")
	assert.Equal(t, "data", volumeName)

	// Test GetLabel for volume without labels
	label := volumes[1].GetLabel("any.label")
	assert.Equal(t, "", label)

	// Test GetLabel for volume with custom labels
	backupLabel := volumes[2].GetLabel("backup")
	assert.Equal(t, "daily", backupLabel)
}

func TestVolumeListView_IsLocal(t *testing.T) {
	volumes := createTestDockerVolumes()

	// Test local volumes
	assert.True(t, volumes[0].IsLocal())
	assert.True(t, volumes[1].IsLocal())

	// Test non-local volume
	assert.False(t, volumes[2].IsLocal())
}

func TestVolumeListView_MountpointTruncation(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewVolumeListView(dockerClient)

	// Create volume with very long mountpoint
	view.dockerVolumes = []models.DockerVolume{
		{
			Name:       "test_volume",
			Driver:     "local",
			Scope:      "local",
			Mountpoint: "/very/long/path/to/docker/volumes/that/exceeds/fifty/characters/test_volume/_data",
			Labels:     "",
		},
	}

	// Create a test app
	app := tview.NewApplication()
	SetApp(app)

	// Update the table
	view.updateTable()

	// Check that mountpoint is truncated
	mountpointCell := view.table.GetCell(1, 3)
	assert.NotNil(t, mountpointCell)
	assert.True(t, len(mountpointCell.Text) <= 50)
	assert.Contains(t, mountpointCell.Text, "...")
}
