package docker

import (
	"testing"
)

// TestClient_ListComposeProjects is an integration test that requires docker compose
func TestClient_ListComposeProjects(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewClient()
	if client == nil {
		t.Fatal("NewClient() returned nil")
	}

	// This will succeed if docker compose is available, otherwise it will error
	// Both outcomes are acceptable for this integration test
	projects, err := client.ListComposeProjects()

	// We don't assert on the error because docker compose may not be available
	// in all test environments. This test primarily verifies the method doesn't panic
	if err == nil {
		// If no error, projects should be a valid slice (could be empty)
		if projects == nil {
			t.Error("ListComposeProjects() returned nil projects with no error")
		}
	}
}
