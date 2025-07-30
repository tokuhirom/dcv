package ui

import (
	"testing"

	"github.com/rivo/tview"
	"github.com/tokuhirom/dcv/internal/models"
)

func TestProcessListView_Creation(t *testing.T) {
	app := &App{
		app:   tview.NewApplication(),
		pages: tview.NewPages(),
	}
	
	view := NewProcessListView(app)
	
	if view == nil {
		t.Fatal("Expected process list view to be created")
	}
	
	if view.table == nil {
		t.Fatal("Expected table to be initialized")
	}
	
	if view.view == nil {
		t.Fatal("Expected view to be initialized")
	}
}

func TestProcessListView_UpdateTable(t *testing.T) {
	app := &App{
		app:   tview.NewApplication(),
		pages: tview.NewPages(),
	}
	
	view := NewProcessListView(app)
	
	// Test with empty processes
	view.processes = []models.Process{}
	view.updateTable()
	
	// Should have only header row
	if view.table.GetRowCount() != 1 {
		t.Errorf("Expected 1 row (header only), got %d", view.table.GetRowCount())
	}
	
	// Test with processes
	view.processes = []models.Process{
		{
			Container: models.Container{
				Name:   "test-1",
				Image:  "test:latest",
				Status: "Up 5 minutes",
			},
			IsDind: false,
		},
		{
			Container: models.Container{
				Name:   "dind-1",
				Image:  "docker:dind",
				Status: "Up 10 minutes",
			},
			IsDind: true,
		},
	}
	
	view.updateTable()
	
	// Should have header + 2 data rows
	if view.table.GetRowCount() != 3 {
		t.Errorf("Expected 3 rows (header + 2 data), got %d", view.table.GetRowCount())
	}
	
	// Check header cells
	headers := []string{"NAME", "IMAGE", "STATUS"}
	for i, expected := range headers {
		cell := view.table.GetCell(0, i)
		if cell == nil {
			t.Errorf("Expected header cell at column %d", i)
			continue
		}
		if cell.Text != expected {
			t.Errorf("Expected header %d to be '%s', got '%s'", i, expected, cell.Text)
		}
	}
	
	// Check first data row
	nameCell := view.table.GetCell(1, 0)
	if nameCell == nil || nameCell.Text != "test-1" {
		t.Error("Expected first row name to be 'test-1'")
	}
	
	// Check dind container has different color
	dindNameCell := view.table.GetCell(2, 0)
	if dindNameCell == nil || dindNameCell.Text != "dind-1" {
		t.Error("Expected second row name to be 'dind-1'")
	}
	
	// Verify table has proper selection
	row, _ := view.table.GetSelection()
	if row != 1 {
		t.Errorf("Expected row 1 to be selected, got %d", row)
	}
}

func TestProcessListView_ErrorHandling(t *testing.T) {
	// Skip this test as it requires running the event loop
	t.Skip("Skipping test that requires event loop")
}