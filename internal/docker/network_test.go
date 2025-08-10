package docker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestParseNetworkJSON tests the network parsing logic
func TestParseNetworkJSON(t *testing.T) {
	testCases := []struct {
		name           string
		input          []byte
		expectedCount  int
		expectedFirst  string
		expectedDriver string
		expectedError  bool
	}{
		{
			name: "parse docker network ls JSON output",
			input: []byte(`{"CreatedAt":"2025-06-25 07:13:05.579113899 +0900 JST","Driver":"bridge","ID":"58c772315063","IPv4":"true","IPv6":"false","Internal":"false","Labels":"com.docker.compose.version=2.33.1,com.docker.compose.network=default","Name":"blog4_default","Scope":"local"}
{"CreatedAt":"2025-06-25 07:10:51.709317902 +0900 JST","Driver":"bridge","ID":"0e4dfb157d47","IPv4":"true","IPv6":"false","Internal":"false","Labels":"","Name":"bridge","Scope":"local"}
{"CreatedAt":"2025-06-22 21:41:39.32360742 +0900 JST","Driver":"host","ID":"4dc1e78a8da8","IPv4":"true","IPv6":"false","Internal":"false","Labels":"","Name":"host","Scope":"local"}`),
			expectedCount:  3,
			expectedFirst:  "blog4_default",
			expectedDriver: "bridge",
		},
		{
			name:          "empty output",
			input:         []byte(""),
			expectedCount: 0,
		},
		{
			name:           "single network",
			input:          []byte(`{"CreatedAt":"2025-06-25 07:10:51.709317902 +0900 JST","Driver":"bridge","ID":"0e4dfb157d47","IPv4":"true","IPv6":"false","Internal":"false","Labels":"","Name":"bridge","Scope":"local"}`),
			expectedCount:  1,
			expectedFirst:  "bridge",
			expectedDriver: "bridge",
		},
		{
			name:           "network with internal=true",
			input:          []byte(`{"CreatedAt":"2025-08-03 11:10:15.441860358 +0900 JST","Driver":"bridge","ID":"c65e5d9416a7","IPv4":"true","IPv6":"false","Internal":"true","Labels":"com.docker.compose.network=development","Name":"dcv-development","Scope":"local"}`),
			expectedCount:  1,
			expectedFirst:  "dcv-development",
			expectedDriver: "bridge",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			networks, err := ParseNetworkJSON(tc.input)

			if tc.expectedError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, networks, tc.expectedCount)

			if tc.expectedCount > 0 {
				assert.Equal(t, tc.expectedFirst, networks[0].Name)
				assert.Equal(t, tc.expectedDriver, networks[0].Driver)
			}
		})
	}
}

// TestClient_ListNetworks is an integration test
func TestClient_ListNetworks(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewClient()
	assert.NotNil(t, client)

	// This will succeed if docker is available, otherwise it will error
	// Both outcomes are acceptable for this integration test
	networks, err := client.ListNetworks()

	// We don't assert on the error because docker may not be available
	// in all test environments. This test primarily verifies the method doesn't panic
	if err == nil {
		// If no error, networks should be a valid slice (could be empty)
		if networks == nil {
			t.Error("ListNetworks() returned nil networks with no error")
		}
	}
}
