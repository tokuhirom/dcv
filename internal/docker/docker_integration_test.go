package docker

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func skipIfNoDocker(t *testing.T) {
	t.Helper()
	if err := exec.Command("docker", "info").Run(); err != nil {
		t.Skip("Docker is not available, skipping integration test")
	}
}

func TestClient_ListVolumes_Integration(t *testing.T) {
	skipIfNoDocker(t)

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
