// Package testutil provides shared test helpers.
package testutil

import (
	"os"
	"os/exec"
	"testing"
)

// SkipIfNoDocker skips the test when Docker is not available locally,
// but fails the test in CI (where Docker is expected to be present).
// GitHub Actions sets the CI environment variable automatically.
func SkipIfNoDocker(t *testing.T) {
	t.Helper()
	if err := exec.Command("docker", "info").Run(); err != nil {
		if os.Getenv("CI") != "" {
			t.Fatalf("Docker is not available in CI environment: %v", err)
		}
		t.Skip("Docker is not available, skipping integration test")
	}
}
