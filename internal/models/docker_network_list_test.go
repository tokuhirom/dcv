package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDockerNetworkList_ToDockerNetwork(t *testing.T) {
	testCases := []struct {
		name     string
		input    DockerNetworkList
		expected DockerNetwork
	}{
		{
			name: "basic conversion",
			input: DockerNetworkList{
				Name:      "test-network",
				ID:        "abc123",
				CreatedAt: "2025-06-25 07:13:05.579113899 +0900 JST",
				Scope:     "local",
				Driver:    "bridge",
				IPv4:      "true",
				IPv6:      "false",
				Internal:  "false",
				Labels:    "com.docker.compose.network=default",
			},
			expected: DockerNetwork{
				Name:     "test-network",
				ID:       "abc123",
				Created:  "2025-06-25 07:13:05.579113899 +0900 JST",
				Scope:    "local",
				Driver:   "bridge",
				Internal: false,
			},
		},
		{
			name: "internal network",
			input: DockerNetworkList{
				Name:      "internal-net",
				ID:        "def456",
				CreatedAt: "2025-08-03 11:10:15.441860358 +0900 JST",
				Scope:     "local",
				Driver:    "bridge",
				IPv4:      "true",
				IPv6:      "false",
				Internal:  "true",
				Labels:    "",
			},
			expected: DockerNetwork{
				Name:     "internal-net",
				ID:       "def456",
				Created:  "2025-08-03 11:10:15.441860358 +0900 JST",
				Scope:    "local",
				Driver:   "bridge",
				Internal: true,
			},
		},
		{
			name: "host network",
			input: DockerNetworkList{
				Name:      "host",
				ID:        "4dc1e78a8da8",
				CreatedAt: "2025-06-22 21:41:39.32360742 +0900 JST",
				Scope:     "local",
				Driver:    "host",
				IPv4:      "true",
				IPv6:      "false",
				Internal:  "false",
				Labels:    "",
			},
			expected: DockerNetwork{
				Name:     "host",
				ID:       "4dc1e78a8da8",
				Created:  "2025-06-22 21:41:39.32360742 +0900 JST",
				Scope:    "local",
				Driver:   "host",
				Internal: false,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.input.ToDockerNetwork()

			assert.Equal(t, tc.expected.Name, result.Name)
			assert.Equal(t, tc.expected.ID, result.ID)
			assert.Equal(t, tc.expected.Created, result.Created)
			assert.Equal(t, tc.expected.Scope, result.Scope)
			assert.Equal(t, tc.expected.Driver, result.Driver)
			assert.Equal(t, tc.expected.Internal, result.Internal)
		})
	}
}

func TestDockerNetworkList_JSONUnmarshal(t *testing.T) {
	testCases := []struct {
		name     string
		json     string
		expected DockerNetworkList
	}{
		{
			name: "typical docker network ls output",
			json: `{"CreatedAt":"2025-06-25 07:13:05.579113899 +0900 JST","Driver":"bridge","ID":"58c772315063","IPv4":"true","IPv6":"false","Internal":"false","Labels":"com.docker.compose.version=2.33.1","Name":"blog4_default","Scope":"local"}`,
			expected: DockerNetworkList{
				Name:      "blog4_default",
				ID:        "58c772315063",
				CreatedAt: "2025-06-25 07:13:05.579113899 +0900 JST",
				Scope:     "local",
				Driver:    "bridge",
				IPv4:      "true",
				IPv6:      "false",
				Internal:  "false",
				Labels:    "com.docker.compose.version=2.33.1",
			},
		},
		{
			name: "internal network",
			json: `{"CreatedAt":"2025-08-03 11:10:15.441860358 +0900 JST","Driver":"bridge","ID":"c65e5d9416a7","IPv4":"true","IPv6":"false","Internal":"true","Labels":"","Name":"dcv-development","Scope":"local"}`,
			expected: DockerNetworkList{
				Name:      "dcv-development",
				ID:        "c65e5d9416a7",
				CreatedAt: "2025-08-03 11:10:15.441860358 +0900 JST",
				Scope:     "local",
				Driver:    "bridge",
				IPv4:      "true",
				IPv6:      "false",
				Internal:  "true",
				Labels:    "",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var result DockerNetworkList
			err := json.Unmarshal([]byte(tc.json), &result)

			assert.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}
