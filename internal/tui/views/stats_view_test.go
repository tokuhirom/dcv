package views

import (
	"testing"
	"time"

	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
)

func createTestContainerStats() []models.ContainerStats {
	return []models.ContainerStats{
		{
			Container: "abc123",
			Name:      "web-server",
			Service:   "web",
			CPUPerc:   "45.23%",
			MemUsage:  "512MB / 1GB",
			MemPerc:   "50.00%",
			NetIO:     "1.2MB / 3.4MB",
			BlockIO:   "100MB / 200MB",
			PIDs:      "5",
		},
		{
			Container: "def456",
			Name:      "database",
			Service:   "db",
			CPUPerc:   "80.50%",
			MemUsage:  "1.5GB / 2GB",
			MemPerc:   "75.00%",
			NetIO:     "5.6MB / 7.8MB",
			BlockIO:   "500MB / 600MB",
			PIDs:      "12",
		},
		{
			Container: "ghi789",
			Name:      "cache",
			Service:   "redis",
			CPUPerc:   "10.00%",
			MemUsage:  "256MB / 512MB",
			MemPerc:   "50.00%",
			NetIO:     "100KB / 200KB",
			BlockIO:   "10MB / 20MB",
			PIDs:      "2",
		},
	}
}

func TestNewStatsView(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewStatsView(dockerClient)

	assert.NotNil(t, view)
	assert.NotNil(t, view.docker)
	assert.NotNil(t, view.table)
	assert.Empty(t, view.stats)
	assert.Equal(t, StatsSortByCPU, view.sortField)
	assert.True(t, view.sortReverse)
	assert.True(t, view.autoRefresh)
	assert.Equal(t, 2*time.Second, view.refreshInterval)
	assert.False(t, view.showAll)

	// Stop auto-refresh for tests
	view.Stop()
}

func TestStatsView_GetPrimitive(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewStatsView(dockerClient)
	defer view.Stop()

	primitive := view.GetPrimitive()

	assert.NotNil(t, primitive)
	_, ok := primitive.(*tview.Table)
	assert.True(t, ok)
}

func TestStatsView_GetTitle(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewStatsView(dockerClient)
	defer view.Stop()

	// Test with auto-refresh enabled (default)
	title := view.GetTitle()
	assert.Contains(t, title, "Container Statistics")
	assert.Contains(t, title, "[Auto-refresh: 2s]")
	assert.Contains(t, title, "[Running]")

	// Test with auto-refresh disabled
	view.autoRefresh = false
	title = view.GetTitle()
	assert.Contains(t, title, "[Auto-refresh: OFF]")

	// Test with show all enabled
	view.showAll = true
	title = view.GetTitle()
	assert.Contains(t, title, "[All]")

	// Test with different refresh interval
	view.autoRefresh = true
	view.refreshInterval = 5 * time.Second
	title = view.GetTitle()
	assert.Contains(t, title, "[Auto-refresh: 5s]")
}

func TestStatsView_UpdateTable(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewStatsView(dockerClient)
	defer view.Stop()
	view.stats = createTestContainerStats()

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
	assert.Contains(t, headerCell.Text, "CPU %") // Contains arrow due to sorting

	headerCell = view.table.GetCell(0, 2)
	assert.NotNil(t, headerCell)
	assert.Equal(t, "MEM USAGE", headerCell.Text)

	headerCell = view.table.GetCell(0, 3)
	assert.NotNil(t, headerCell)
	assert.Equal(t, "MEM %", headerCell.Text)

	headerCell = view.table.GetCell(0, 4)
	assert.NotNil(t, headerCell)
	assert.Equal(t, "NET I/O", headerCell.Text)

	headerCell = view.table.GetCell(0, 5)
	assert.NotNil(t, headerCell)
	assert.Equal(t, "BLOCK I/O", headerCell.Text)

	headerCell = view.table.GetCell(0, 6)
	assert.NotNil(t, headerCell)
	assert.Equal(t, "PIDS", headerCell.Text)

	// Check stats data - should be sorted by CPU descending
	// database (80.50%) should be first
	nameCell := view.table.GetCell(1, 0)
	assert.NotNil(t, nameCell)
	assert.Equal(t, "database", nameCell.Text)

	cpuCell := view.table.GetCell(1, 1)
	assert.NotNil(t, cpuCell)
	assert.Equal(t, "80.50%", cpuCell.Text)

	memUsageCell := view.table.GetCell(1, 2)
	assert.NotNil(t, memUsageCell)
	assert.Equal(t, "1.5GB / 2GB", memUsageCell.Text)

	memPercCell := view.table.GetCell(1, 3)
	assert.NotNil(t, memPercCell)
	assert.Equal(t, "75.00%", memPercCell.Text)

	netIOCell := view.table.GetCell(1, 4)
	assert.NotNil(t, netIOCell)
	assert.Equal(t, "5.6MB / 7.8MB", netIOCell.Text)

	blockIOCell := view.table.GetCell(1, 5)
	assert.NotNil(t, blockIOCell)
	assert.Equal(t, "500MB / 600MB", blockIOCell.Text)

	pidsCell := view.table.GetCell(1, 6)
	assert.NotNil(t, pidsCell)
	assert.Equal(t, "12", pidsCell.Text)

	// Verify table selection
	row, _ := view.table.GetSelection()
	assert.Equal(t, 1, row) // First data row should be selected
}

func TestStatsView_Sorting(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewStatsView(dockerClient)
	defer view.Stop()
	view.stats = createTestContainerStats()

	// Create a test app
	app := tview.NewApplication()
	SetApp(app)

	tests := []struct {
		name          string
		sortField     StatsSortField
		sortReverse   bool
		expectedFirst string
	}{
		{
			name:          "Sort by CPU descending",
			sortField:     StatsSortByCPU,
			sortReverse:   true,
			expectedFirst: "database", // 80.50%
		},
		{
			name:          "Sort by CPU ascending",
			sortField:     StatsSortByCPU,
			sortReverse:   false,
			expectedFirst: "cache", // 10.00%
		},
		{
			name:          "Sort by Memory descending",
			sortField:     StatsSortByMem,
			sortReverse:   true,
			expectedFirst: "database", // 75.00%
		},
		{
			name:          "Sort by Name ascending",
			sortField:     StatsSortByName,
			sortReverse:   false,
			expectedFirst: "cache",
		},
		{
			name:          "Sort by Name descending",
			sortField:     StatsSortByName,
			sortReverse:   true,
			expectedFirst: "web-server",
		},
		{
			name:          "Sort by Network I/O descending",
			sortField:     StatsSortByNetIO,
			sortReverse:   true,
			expectedFirst: "database", // 5.6MB + 7.8MB
		},
		{
			name:          "Sort by Block I/O descending",
			sortField:     StatsSortByBlockIO,
			sortReverse:   true,
			expectedFirst: "database", // 500MB + 600MB
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view.sortField = tt.sortField
			view.sortReverse = tt.sortReverse
			view.updateTable()

			// Check that the first data row contains the expected container
			nameCell := view.table.GetCell(1, 0)
			assert.NotNil(t, nameCell)
			assert.Equal(t, tt.expectedFirst, nameCell.Text)
		})
	}
}

func TestStatsView_SetSortField(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewStatsView(dockerClient)
	defer view.Stop()

	// Test setting a new sort field
	view.setSortField(StatsSortByName)
	assert.Equal(t, StatsSortByName, view.sortField)
	assert.False(t, view.sortReverse) // Default to ascending for names

	// Test toggling the same field
	view.setSortField(StatsSortByName)
	assert.Equal(t, StatsSortByName, view.sortField)
	assert.True(t, view.sortReverse) // Should toggle

	// Test setting numeric field
	view.setSortField(StatsSortByCPU)
	assert.Equal(t, StatsSortByCPU, view.sortField)
	assert.True(t, view.sortReverse) // Default to descending for numeric
}

func TestStatsView_ParseFunctions(t *testing.T) {
	// Test parsePercentage
	assert.Equal(t, 45.23, parsePercentage("45.23%"))
	assert.Equal(t, 100.0, parsePercentage("100%"))
	assert.Equal(t, 0.0, parsePercentage(""))
	assert.Equal(t, 0.0, parsePercentage("invalid"))

	// Test parseSizeString
	assert.Equal(t, 1024.0, parseSizeString("1KB"))
	assert.Equal(t, 1024.0*1024, parseSizeString("1MB"))
	assert.Equal(t, 1024.0*1024*1024, parseSizeString("1GB"))
	assert.Equal(t, 512.0*1024*1024, parseSizeString("512MB"))
	assert.Equal(t, 0.0, parseSizeString("--"))
	assert.Equal(t, 0.0, parseSizeString(""))

	// Test parseIOBytes
	assert.Equal(t, (1.2+3.4)*1024*1024, parseIOBytes("1.2MB / 3.4MB"))
	assert.Equal(t, (100.0+200.0)*1024, parseIOBytes("100KB / 200KB"))
	assert.Equal(t, 0.0, parseIOBytes("--"))
	assert.Equal(t, 0.0, parseIOBytes(""))
}

func TestStatsView_AutoRefresh(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewStatsView(dockerClient)
	defer view.Stop()

	// Test initial state
	assert.True(t, view.autoRefresh)

	// Test toggle off
	view.toggleAutoRefresh()
	assert.False(t, view.autoRefresh)

	// Test toggle on
	view.toggleAutoRefresh()
	assert.True(t, view.autoRefresh)

	// Test refresh interval changes
	originalInterval := view.refreshInterval
	view.refreshInterval = 3 * time.Second
	assert.Equal(t, 3*time.Second, view.refreshInterval)
	assert.NotEqual(t, originalInterval, view.refreshInterval)
}

func TestStatsView_EmptyStatsList(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewStatsView(dockerClient)
	defer view.Stop()
	view.stats = []models.ContainerStats{}

	// Create a test app
	app := tview.NewApplication()
	SetApp(app)

	// Update the table with empty stats list
	view.updateTable()

	// Should still have headers
	headerCell := view.table.GetCell(0, 0)
	assert.NotNil(t, headerCell)
	assert.Equal(t, "NAME", headerCell.Text)

	// Check table row count - should only have header row
	rowCount := view.table.GetRowCount()
	assert.Equal(t, 1, rowCount) // Only header row
}

func TestStatsView_GetSelectedStat(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewStatsView(dockerClient)
	defer view.Stop()
	view.stats = createTestContainerStats()

	// Create a test app
	app := tview.NewApplication()
	SetApp(app)

	// Update the table
	view.updateTable()

	// First row should be selected by default
	row, _ := view.table.GetSelection()
	assert.Equal(t, 1, row)

	// Get the selected stat (should be database due to CPU sort)
	selectedStat := view.GetSelectedStat()
	assert.NotNil(t, selectedStat)
	assert.Equal(t, "database", selectedStat.Name)
	assert.Equal(t, "80.50%", selectedStat.CPUPerc)

	// Move selection down
	view.table.Select(2, 0)
	row, _ = view.table.GetSelection()
	assert.Equal(t, 2, row)

	// Get the new selected stat
	selectedStat = view.GetSelectedStat()
	assert.NotNil(t, selectedStat)
	assert.Equal(t, "web-server", selectedStat.Name)
}

func TestStatsView_TableProperties(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewStatsView(dockerClient)
	defer view.Stop()

	// Check table properties
	selectable, selectableByRow := view.table.GetSelectable()
	assert.True(t, selectable)
	assert.False(t, selectableByRow)

	// Check that table is configured correctly
	assert.NotNil(t, view.table)
}

func TestStatsView_KeyHandling(t *testing.T) {
	// This test verifies that key handlers are properly set up
	dockerClient := docker.NewClient()
	view := NewStatsView(dockerClient)
	defer view.Stop()

	// The table should have input capture function set
	assert.NotNil(t, view.table)

	// We can't easily test the actual key handling without running the app,
	// but we can verify the structure is in place
	primitive := view.GetPrimitive()
	table, ok := primitive.(*tview.Table)
	assert.True(t, ok)
	assert.NotNil(t, table)
}

func TestStatsView_ShowAll(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewStatsView(dockerClient)
	defer view.Stop()

	// Test initial state
	assert.False(t, view.showAll)

	// Test toggle
	view.showAll = true
	assert.True(t, view.showAll)

	// Verify it affects the title
	title := view.GetTitle()
	assert.Contains(t, title, "[All]")

	view.showAll = false
	title = view.GetTitle()
	assert.Contains(t, title, "[Running]")
}

func TestStatsView_LongNameTruncation(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewStatsView(dockerClient)
	defer view.Stop()

	// Create stats with a very long name
	view.stats = []models.ContainerStats{
		{
			Container: "abc123",
			Name:      "very-long-container-name-that-exceeds-twenty-characters",
			CPUPerc:   "10.00%",
			MemUsage:  "256MB / 512MB",
			MemPerc:   "50.00%",
			NetIO:     "100KB / 200KB",
			BlockIO:   "10MB / 20MB",
			PIDs:      "2",
		},
	}

	// Create a test app
	app := tview.NewApplication()
	SetApp(app)

	// Update the table
	view.updateTable()

	// Check that name is truncated
	nameCell := view.table.GetCell(1, 0)
	assert.NotNil(t, nameCell)
	assert.Equal(t, "very-long-contain...", nameCell.Text)
	assert.True(t, len(nameCell.Text) <= 20)
}
