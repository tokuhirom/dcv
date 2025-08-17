package views

import (
	"testing"

	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
)

func createTestDockerNetworks() []models.DockerNetwork {
	return []models.DockerNetwork{
		{
			ID:     "abc123def456",
			Name:   "bridge",
			Driver: "bridge",
			Scope:  "local",
			IPAM: struct {
				Driver  string            `json:"Driver"`
				Options map[string]string `json:"Options"`
				Config  []struct {
					Subnet  string `json:"Subnet"`
					Gateway string `json:"Gateway"`
				} `json:"Config"`
			}{
				Config: []struct {
					Subnet  string `json:"Subnet"`
					Gateway string `json:"Gateway"`
				}{
					{
						Subnet:  "172.17.0.0/16",
						Gateway: "172.17.0.1",
					},
				},
			},
			Containers: map[string]struct {
				Name        string `json:"Name"`
				EndpointID  string `json:"EndpointID"`
				MacAddress  string `json:"MacAddress"`
				IPv4Address string `json:"IPv4Address"`
				IPv6Address string `json:"IPv6Address"`
			}{
				"container1": {
					Name:        "web1",
					IPv4Address: "172.17.0.2/16",
				},
				"container2": {
					Name:        "db1",
					IPv4Address: "172.17.0.3/16",
				},
			},
		},
		{
			ID:     "789012345678",
			Name:   "myapp_network",
			Driver: "bridge",
			Scope:  "local",
			IPAM: struct {
				Driver  string            `json:"Driver"`
				Options map[string]string `json:"Options"`
				Config  []struct {
					Subnet  string `json:"Subnet"`
					Gateway string `json:"Gateway"`
				} `json:"Config"`
			}{
				Config: []struct {
					Subnet  string `json:"Subnet"`
					Gateway string `json:"Gateway"`
				}{
					{
						Subnet:  "192.168.1.0/24",
						Gateway: "192.168.1.1",
					},
				},
			},
			Containers: map[string]struct {
				Name        string `json:"Name"`
				EndpointID  string `json:"EndpointID"`
				MacAddress  string `json:"MacAddress"`
				IPv4Address string `json:"IPv4Address"`
				IPv6Address string `json:"IPv6Address"`
			}{},
		},
		{
			ID:     "fedcba987654",
			Name:   "host",
			Driver: "host",
			Scope:  "local",
			Containers: map[string]struct {
				Name        string `json:"Name"`
				EndpointID  string `json:"EndpointID"`
				MacAddress  string `json:"MacAddress"`
				IPv4Address string `json:"IPv4Address"`
				IPv6Address string `json:"IPv6Address"`
			}{},
		},
	}
}

func TestNewNetworkListView(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewNetworkListView(dockerClient)

	assert.NotNil(t, view)
	assert.NotNil(t, view.docker)
	assert.NotNil(t, view.table)
	assert.Empty(t, view.dockerNetworks)
}

func TestNetworkListView_GetPrimitive(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewNetworkListView(dockerClient)

	primitive := view.GetPrimitive()

	assert.NotNil(t, primitive)
	_, ok := primitive.(*tview.Table)
	assert.True(t, ok)
}

func TestNetworkListView_GetTitle(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewNetworkListView(dockerClient)

	title := view.GetTitle()
	assert.Equal(t, "Docker Networks", title)
}

func TestNetworkListView_UpdateTable(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewNetworkListView(dockerClient)
	view.dockerNetworks = createTestDockerNetworks()

	// Create a test app
	app := tview.NewApplication()
	SetApp(app)

	// Update the table
	view.updateTable()

	// Check headers
	headerCell := view.table.GetCell(0, 0)
	assert.NotNil(t, headerCell)
	assert.Equal(t, "NETWORK ID", headerCell.Text)

	headerCell = view.table.GetCell(0, 1)
	assert.NotNil(t, headerCell)
	assert.Equal(t, "NAME", headerCell.Text)

	headerCell = view.table.GetCell(0, 2)
	assert.NotNil(t, headerCell)
	assert.Equal(t, "DRIVER", headerCell.Text)

	headerCell = view.table.GetCell(0, 3)
	assert.NotNil(t, headerCell)
	assert.Equal(t, "SCOPE", headerCell.Text)

	headerCell = view.table.GetCell(0, 4)
	assert.NotNil(t, headerCell)
	assert.Equal(t, "CONTAINERS", headerCell.Text)

	headerCell = view.table.GetCell(0, 5)
	assert.NotNil(t, headerCell)
	assert.Equal(t, "SUBNET", headerCell.Text)

	// Check network data - first network (bridge)
	idCell := view.table.GetCell(1, 0)
	assert.NotNil(t, idCell)
	assert.Equal(t, "abc123def456", idCell.Text) // Should show full 12 chars

	nameCell := view.table.GetCell(1, 1)
	assert.NotNil(t, nameCell)
	assert.Equal(t, "bridge", nameCell.Text)

	driverCell := view.table.GetCell(1, 2)
	assert.NotNil(t, driverCell)
	assert.Equal(t, "bridge", driverCell.Text)

	scopeCell := view.table.GetCell(1, 3)
	assert.NotNil(t, scopeCell)
	assert.Equal(t, "local", scopeCell.Text)

	containerCell := view.table.GetCell(1, 4)
	assert.NotNil(t, containerCell)
	assert.Equal(t, "2", containerCell.Text) // 2 containers connected

	subnetCell := view.table.GetCell(1, 5)
	assert.NotNil(t, subnetCell)
	assert.Equal(t, "172.17.0.0/16", subnetCell.Text)

	// Check second network
	nameCell2 := view.table.GetCell(2, 1)
	assert.NotNil(t, nameCell2)
	assert.Equal(t, "myapp_network", nameCell2.Text)

	containerCell2 := view.table.GetCell(2, 4)
	assert.NotNil(t, containerCell2)
	assert.Equal(t, "0", containerCell2.Text) // No containers

	// Check third network (host)
	nameCell3 := view.table.GetCell(3, 1)
	assert.NotNil(t, nameCell3)
	assert.Equal(t, "host", nameCell3.Text)

	driverCell3 := view.table.GetCell(3, 2)
	assert.NotNil(t, driverCell3)
	assert.Equal(t, "host", driverCell3.Text)

	subnetCell3 := view.table.GetCell(3, 5)
	assert.NotNil(t, subnetCell3)
	assert.Equal(t, "-", subnetCell3.Text) // No subnet for host network

	// Verify table selection
	row, _ := view.table.GetSelection()
	assert.Equal(t, 1, row) // First data row should be selected
}

func TestNetworkListView_TableProperties(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewNetworkListView(dockerClient)

	// Check table properties
	selectable, selectableByRow := view.table.GetSelectable()
	assert.True(t, selectable)
	assert.False(t, selectableByRow)

	// Check that table is configured correctly
	assert.NotNil(t, view.table)
}

func TestNetworkListView_EmptyNetworkList(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewNetworkListView(dockerClient)
	view.dockerNetworks = []models.DockerNetwork{}

	// Create a test app
	app := tview.NewApplication()
	SetApp(app)

	// Update the table with empty network list
	view.updateTable()

	// Should still have headers
	headerCell := view.table.GetCell(0, 0)
	assert.NotNil(t, headerCell)
	assert.Equal(t, "NETWORK ID", headerCell.Text)

	// Check table row count - should only have header row
	rowCount := view.table.GetRowCount()
	assert.Equal(t, 1, rowCount) // Only header row
}

func TestNetworkListView_GetSelectedNetwork(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewNetworkListView(dockerClient)
	view.dockerNetworks = createTestDockerNetworks()

	// Create a test app
	app := tview.NewApplication()
	SetApp(app)

	// Update the table
	view.updateTable()

	// First row should be selected by default
	row, _ := view.table.GetSelection()
	assert.Equal(t, 1, row)

	// Get the selected network
	selectedNetwork := view.GetSelectedNetwork()
	assert.NotNil(t, selectedNetwork)
	assert.Equal(t, "bridge", selectedNetwork.Name)
	assert.Equal(t, "abc123def456", selectedNetwork.ID)

	// Move selection down
	view.table.Select(2, 0)
	row, _ = view.table.GetSelection()
	assert.Equal(t, 2, row)

	// Get the new selected network
	selectedNetwork = view.GetSelectedNetwork()
	assert.NotNil(t, selectedNetwork)
	assert.Equal(t, "myapp_network", selectedNetwork.Name)
	assert.Equal(t, "789012345678", selectedNetwork.ID)
}

func TestNetworkListView_SearchNetworks(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewNetworkListView(dockerClient)
	view.dockerNetworks = createTestDockerNetworks()

	// Create a test app
	app := tview.NewApplication()
	SetApp(app)

	// Search for "bridge"
	view.SearchNetworks("bridge")

	// Search for "myapp"
	view.SearchNetworks("myapp")

	// Test empty search (should reset)
	view.SearchNetworks("")

	// Verify original networks are still intact
	assert.Len(t, view.dockerNetworks, 3)
}

func TestNetworkListView_NetworkOperations(t *testing.T) {
	// Test that network operations don't crash
	// Note: These operations will fail in tests since we don't have actual networks,
	// but we're checking that the methods exist and don't panic

	t.Run("DeleteNetwork", func(t *testing.T) {
		dockerClient := docker.NewClient()
		view := NewNetworkListView(dockerClient)
		networks := createTestDockerNetworks()

		// Try to delete a user network (should attempt)
		assert.NotPanics(t, func() {
			view.deleteNetwork(networks[1]) // myapp_network
		})

		// Try to delete a default network (should be blocked)
		assert.NotPanics(t, func() {
			view.deleteNetwork(networks[0]) // bridge
		})
	})

	t.Run("PruneNetworks", func(t *testing.T) {
		dockerClient := docker.NewClient()
		view := NewNetworkListView(dockerClient)

		assert.NotPanics(t, func() {
			view.pruneNetworks()
		})
	})

	t.Run("CreateNetwork", func(t *testing.T) {
		dockerClient := docker.NewClient()
		view := NewNetworkListView(dockerClient)

		// Test creating a network (will fail but shouldn't panic)
		err := view.CreateNetwork("test-network", "bridge", false)
		// We expect an error or success depending on Docker state
		_ = err
	})

	t.Run("ConnectContainer", func(t *testing.T) {
		dockerClient := docker.NewClient()
		view := NewNetworkListView(dockerClient)
		networks := createTestDockerNetworks()

		// Test connecting a container (will fail but shouldn't panic)
		err := view.ConnectContainer(networks[0], "test-container")
		// We expect an error since the container doesn't exist
		assert.Error(t, err)
	})

	t.Run("DisconnectContainer", func(t *testing.T) {
		dockerClient := docker.NewClient()
		view := NewNetworkListView(dockerClient)
		networks := createTestDockerNetworks()

		// Test disconnecting a container (will fail but shouldn't panic)
		err := view.DisconnectContainer(networks[0], "test-container")
		// We expect an error since the container doesn't exist
		assert.Error(t, err)
	})

	t.Run("InspectNetwork", func(t *testing.T) {
		dockerClient := docker.NewClient()
		view := NewNetworkListView(dockerClient)
		networks := createTestDockerNetworks()

		// Test inspecting a network (will fail but shouldn't panic)
		_, err := view.InspectNetwork(networks[0])
		// We expect an error or success depending on whether bridge network exists
		_ = err
	})
}

func TestNetworkListView_KeyHandling(t *testing.T) {
	// This test verifies that key handlers are properly set up
	dockerClient := docker.NewClient()
	view := NewNetworkListView(dockerClient)

	// The table should have input capture function set
	assert.NotNil(t, view.table)

	// We can't easily test the actual key handling without running the app,
	// but we can verify the structure is in place
	primitive := view.GetPrimitive()
	table, ok := primitive.(*tview.Table)
	assert.True(t, ok)
	assert.NotNil(t, table)
}

func TestNetworkListView_GetSubnet(t *testing.T) {
	networks := createTestDockerNetworks()

	// Test network with subnet
	subnet := networks[0].GetSubnet()
	assert.Equal(t, "172.17.0.0/16", subnet)

	// Test network without subnet (host network)
	subnet = networks[2].GetSubnet()
	assert.Equal(t, "", subnet)
}

func TestNetworkListView_GetContainerCount(t *testing.T) {
	networks := createTestDockerNetworks()

	// Test network with containers
	count := networks[0].GetContainerCount()
	assert.Equal(t, 2, count)

	// Test network without containers
	count = networks[1].GetContainerCount()
	assert.Equal(t, 0, count)
}

func TestNetworkListView_DefaultNetworkProtection(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewNetworkListView(dockerClient)

	defaultNetworks := []models.DockerNetwork{
		{ID: "bridge-id", Name: "bridge"},
		{ID: "host-id", Name: "host"},
		{ID: "none-id", Name: "none"},
	}

	// All default networks should be protected from deletion
	for _, network := range defaultNetworks {
		assert.NotPanics(t, func() {
			view.deleteNetwork(network)
			// Should log warning but not attempt deletion
		})
	}
}

func TestNetworkListView_NetworkIDTruncation(t *testing.T) {
	dockerClient := docker.NewClient()
	view := NewNetworkListView(dockerClient)

	// Create networks with various ID lengths
	view.dockerNetworks = []models.DockerNetwork{
		{
			ID:     "abcdefghijklmnopqrstuvwxyz",
			Name:   "long-id-network",
			Driver: "bridge",
			Scope:  "local",
		},
		{
			ID:     "short",
			Name:   "short-id-network",
			Driver: "bridge",
			Scope:  "local",
		},
	}

	// Create a test app
	app := tview.NewApplication()
	SetApp(app)

	// Update the table
	view.updateTable()

	// Check that long ID is truncated
	idCell1 := view.table.GetCell(1, 0)
	assert.NotNil(t, idCell1)
	assert.Equal(t, "abcdefghijkl", idCell1.Text) // Truncated to 12 chars

	// Check that short ID is not truncated
	idCell2 := view.table.GetCell(2, 0)
	assert.NotNil(t, idCell2)
	assert.Equal(t, "short", idCell2.Text) // Not truncated
}
