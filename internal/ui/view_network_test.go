package ui

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tokuhirom/dcv/internal/models"
)

func TestNetworkListView(t *testing.T) {
	testCases := []struct {
		name     string
		model    Model
		expected []string
	}{
		{
			name: "network list with multiple networks",
			model: Model{
				currentView: NetworkListView,
				networkListViewModel: NetworkListViewModel{
					dockerNetworks: []models.DockerNetwork{
						{
							ID:       "58c772315063",
							Name:     "blog4_default",
							Driver:   "bridge",
							Scope:    "local",
							Internal: false,
						},
						{
							ID:       "0e4dfb157d47",
							Name:     "bridge",
							Driver:   "bridge",
							Scope:    "local",
							Internal: false,
						},
						{
							ID:       "c65e5d9416a7",
							Name:     "dcv-development",
							Driver:   "bridge",
							Scope:    "local",
							Internal: true,
						},
					},
					TableViewModel: TableViewModel{Cursor: 1},
				},
				width:  120,
				Height: 30,
			},
			expected: []string{
				"NETWORK ID",
				"NAME",
				"DRIVER",
				"SCOPE",
				"CONTAINERS",
				"58c772315063", // First network ID (should be dimmed)
				"blog4_default",
				"bridge", // selected row
				"dcv-development",
			},
		},
		{
			name: "empty network list",
			model: Model{
				currentView: NetworkListView,
				networkListViewModel: NetworkListViewModel{
					dockerNetworks: []models.DockerNetwork{},
				},
				width:  120,
				Height: 30,
			},
			expected: []string{
				"No networks found",
			},
		},
	}

	for i := range testCases {
		tc := &testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			// Initialize handlers to avoid nil map panic
			tc.model.initializeKeyHandlers()

			// Build rows for table
			tc.model.networkListViewModel.SetRows(
				tc.model.networkListViewModel.buildRows(),
				tc.model.ViewHeight(),
			)

			// RenderTable the view
			view := tc.model.View()

			// Check that expected strings are present
			for _, expected := range tc.expected {
				assert.Contains(t, view, expected, "View should contain: %s", expected)
			}
		})
	}
}

func TestNetworkListView_LongStrings(t *testing.T) {
	longName := strings.Repeat("my-very-long-network-name-", 10)
	longDriver := strings.Repeat("custom-driver-", 20)
	longID := strings.Repeat("abcdef", 20)

	tests := []struct {
		name     string
		networks []models.DockerNetwork
		width    int
		height   int
	}{
		{
			name: "very long network name",
			networks: []models.DockerNetwork{
				{ID: "abc123", Name: longName, Driver: "bridge", Scope: "local"},
			},
			width: 80, height: 20,
		},
		{
			name: "very long driver name",
			networks: []models.DockerNetwork{
				{ID: "abc123", Name: "net", Driver: longDriver, Scope: "local"},
			},
			width: 80, height: 20,
		},
		{
			name: "very long network ID",
			networks: []models.DockerNetwork{
				{ID: longID, Name: "net", Driver: "bridge", Scope: "local"},
			},
			width: 80, height: 20,
		},
		{
			name: "all fields long simultaneously",
			networks: []models.DockerNetwork{
				{ID: longID, Name: longName, Driver: longDriver, Scope: strings.Repeat("scope-", 20)},
			},
			width: 60, height: 20,
		},
		{
			name: "narrow terminal",
			networks: []models.DockerNetwork{
				{ID: "abc123", Name: longName, Driver: "bridge", Scope: "local"},
			},
			width: 30, height: 20,
		},
		{
			name: "very small height with many networks",
			networks: func() []models.DockerNetwork {
				var nets []models.DockerNetwork
				for range 20 {
					nets = append(nets, models.DockerNetwork{
						ID: longID, Name: longName, Driver: "bridge", Scope: "local",
					})
				}
				return nets
			}(),
			width: 80, height: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := &NetworkListViewModel{
				dockerNetworks: tt.networks,
				TableViewModel: TableViewModel{Cursor: 0},
			}
			model := &Model{width: tt.width, Height: tt.height}
			vm.SetRows(vm.buildRows(), model.ViewHeight())

			// Should not panic
			result := vm.render(model, tt.height-4)
			assert.NotEmpty(t, result)
		})
	}
}

func TestRenderNetworkList(t *testing.T) {
	vm := &NetworkListViewModel{
		dockerNetworks: []models.DockerNetwork{
			{
				ID:       "abc123",
				Name:     "test-network",
				Driver:   "bridge",
				Scope:    "local",
				Internal: false,
			},
		},
		TableViewModel: TableViewModel{Cursor: 0},
	}

	// Test rendering with sufficient Height
	model := &Model{Height: 20, width: 100}
	// Build rows for table
	vm.SetRows(vm.buildRows(), model.ViewHeight())
	output := vm.render(model, 10)

	assert.Contains(t, output, "NETWORK ID")
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "DRIVER")
	assert.Contains(t, output, "SCOPE")
	assert.Contains(t, output, "CONTAINERS")
	assert.Contains(t, output, "abc123")
	assert.Contains(t, output, "test-network")
	assert.Contains(t, output, "bridge")
	assert.Contains(t, output, "local")
	assert.Contains(t, output, "0") // container count
}
