package docker

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tokuhirom/dcv/internal/testutil"
)

func TestClient_ListVolumes_Integration(t *testing.T) {
	testutil.SkipIfNoDocker(t)

	volumeName := "dcv-test-volume-" + t.Name()

	// Create a test volume
	out, err := exec.Command("docker", "volume", "create", volumeName).CombinedOutput()
	require.NoError(t, err, "failed to create test volume: %s", string(out))
	t.Cleanup(func() {
		_ = exec.Command("docker", "volume", "rm", "-f", volumeName).Run()
	})

	// Write some data into the volume so it has a non-zero size
	out, err = exec.Command("docker", "run", "--rm",
		"-v", volumeName+":/data",
		"alpine:latest",
		"sh", "-c", "dd if=/dev/urandom of=/data/testfile bs=1024 count=10",
	).CombinedOutput()
	require.NoError(t, err, "failed to write data to volume: %s", string(out))

	client := NewClient()
	volumes, err := client.ListVolumes()
	require.NoError(t, err)

	// Find our test volume
	var found bool
	for _, v := range volumes {
		if v.Name == volumeName {
			found = true
			assert.Equal(t, "local", v.Driver)
			assert.Equal(t, "local", v.Scope)
			assert.NotEmpty(t, v.Size, "Size should not be empty")
			// Size should not be "N/A" since we used system df
			assert.False(t, strings.EqualFold(v.Size, "N/A"), "Size should not be N/A, got: %s", v.Size)
			break
		}
	}
	assert.True(t, found, "test volume %s not found in ListVolumes result", volumeName)
}

func TestClient_ListContainers_Integration(t *testing.T) {
	testutil.SkipIfNoDocker(t)

	containerName := "dcv-test-list-containers"

	// Start a test container
	out, err := exec.Command("docker", "run", "-d", "--name", containerName,
		"alpine:latest", "sleep", "300").CombinedOutput()
	require.NoError(t, err, "failed to start test container: %s", string(out))
	t.Cleanup(func() {
		_ = exec.Command("docker", "rm", "-f", containerName).Run()
	})

	client := NewClient()
	containers, err := client.ListContainers(false)
	require.NoError(t, err)

	// Find our test container
	var found bool
	for _, c := range containers {
		if c.Names == containerName {
			found = true
			assert.Equal(t, "alpine:latest", c.Image)
			assert.Contains(t, c.Status, "Up")
			assert.Equal(t, "running", c.State)
			assert.NotEmpty(t, c.ID)
			break
		}
	}
	assert.True(t, found, "test container %s not found in ListContainers result", containerName)
}

func TestClient_ListContainers_ShowAll_Integration(t *testing.T) {
	testutil.SkipIfNoDocker(t)

	containerName := "dcv-test-list-containers-all"

	// Create a stopped container
	out, err := exec.Command("docker", "create", "--name", containerName,
		"alpine:latest", "echo", "hello").CombinedOutput()
	require.NoError(t, err, "failed to create test container: %s", string(out))
	t.Cleanup(func() {
		_ = exec.Command("docker", "rm", "-f", containerName).Run()
	})

	client := NewClient()

	// Without showAll, stopped container should not appear
	containers, err := client.ListContainers(false)
	require.NoError(t, err)
	var foundWithoutAll bool
	for _, c := range containers {
		if c.Names == containerName {
			foundWithoutAll = true
			break
		}
	}
	assert.False(t, foundWithoutAll, "stopped container should not appear without showAll")

	// With showAll, stopped container should appear
	containers, err = client.ListContainers(true)
	require.NoError(t, err)
	var foundWithAll bool
	for _, c := range containers {
		if c.Names == containerName {
			foundWithAll = true
			assert.Equal(t, "created", c.State)
			break
		}
	}
	assert.True(t, foundWithAll, "stopped container should appear with showAll")
}

func TestClient_ListImages_Integration(t *testing.T) {
	testutil.SkipIfNoDocker(t)

	// Ensure alpine:latest is available (it should be from other tests)
	out, err := exec.Command("docker", "pull", "alpine:latest").CombinedOutput()
	require.NoError(t, err, "failed to pull alpine: %s", string(out))

	client := NewClient()
	images, err := client.ListImages(false)
	require.NoError(t, err)

	// Find alpine image
	var found bool
	for _, img := range images {
		if img.Repository == "alpine" && img.Tag == "latest" {
			found = true
			assert.NotEmpty(t, img.ID)
			assert.NotEmpty(t, img.Size)
			assert.NotEmpty(t, img.CreatedSince)
			break
		}
	}
	assert.True(t, found, "alpine:latest not found in ListImages result")
}

func TestClient_ListNetworks_Integration(t *testing.T) {
	testutil.SkipIfNoDocker(t)

	networkName := "dcv-test-network"

	// Create a test network
	out, err := exec.Command("docker", "network", "create", networkName).CombinedOutput()
	require.NoError(t, err, "failed to create test network: %s", string(out))
	t.Cleanup(func() {
		_ = exec.Command("docker", "network", "rm", networkName).Run()
	})

	client := NewClient()
	networks, err := client.ListNetworks()
	require.NoError(t, err)

	// Find our test network
	var found bool
	for _, n := range networks {
		if n.Name == networkName {
			found = true
			assert.Equal(t, "bridge", n.Driver)
			assert.Equal(t, "local", n.Scope)
			assert.NotEmpty(t, n.ID)
			break
		}
	}
	assert.True(t, found, "test network %s not found in ListNetworks result", networkName)

	// Default networks should also be present
	var hasBridge bool
	for _, n := range networks {
		if n.Name == "bridge" {
			hasBridge = true
			break
		}
	}
	assert.True(t, hasBridge, "default bridge network should be present")
}

func TestClient_GetStats_Integration(t *testing.T) {
	testutil.SkipIfNoDocker(t)

	containerName := "dcv-test-stats"

	// Start a test container
	out, err := exec.Command("docker", "run", "-d", "--name", containerName,
		"alpine:latest", "sleep", "300").CombinedOutput()
	require.NoError(t, err, "failed to start test container: %s", string(out))
	t.Cleanup(func() {
		_ = exec.Command("docker", "rm", "-f", containerName).Run()
	})

	client := NewClient()
	stats, err := client.GetStats(false)
	require.NoError(t, err)

	// Find our test container's stats
	var found bool
	for _, s := range stats {
		if s.Name == containerName {
			found = true
			assert.NotEmpty(t, s.Container)
			assert.NotEmpty(t, s.CPUPerc)
			assert.NotEmpty(t, s.MemUsage)
			assert.NotEmpty(t, s.MemPerc)
			assert.NotEmpty(t, s.NetIO)
			assert.NotEmpty(t, s.BlockIO)
			assert.NotEmpty(t, s.PIDs)
			break
		}
	}
	assert.True(t, found, "test container %s not found in GetStats result", containerName)
}
