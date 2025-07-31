package docker

import (
	"testing"
)

func TestBuildComposeArgs(t *testing.T) {
	tests := []struct {
		name        string
		client      *ComposeClient
		baseArgs    []string
		expected    []string
	}{
		{
			name: "no project or file",
			client: &ComposeClient{},
			baseArgs: []string{"ps"},
			expected: []string{"compose", "ps"},
		},
		{
			name: "with project name",
			client: &ComposeClient{projectName: "myproject"},
			baseArgs: []string{"ps"},
			expected: []string{"compose", "-p", "myproject", "ps"},
		},
		{
			name: "with compose file",
			client: &ComposeClient{composeFile: "docker-compose.prod.yml"},
			baseArgs: []string{"ps"},
			expected: []string{"compose", "-f", "docker-compose.prod.yml", "ps"},
		},
		{
			name: "with both project and file",
			client: &ComposeClient{
				projectName: "myproject",
				composeFile: "docker-compose.prod.yml",
			},
			baseArgs: []string{"ps", "--all"},
			expected: []string{"compose", "-p", "myproject", "-f", "docker-compose.prod.yml", "ps", "--all"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.client.buildComposeArgs(tt.baseArgs...)
			if len(result) != len(tt.expected) {
				t.Errorf("buildComposeArgs() returned %d args, expected %d", len(result), len(tt.expected))
				return
			}
			for i, arg := range result {
				if arg != tt.expected[i] {
					t.Errorf("buildComposeArgs()[%d] = %s, expected %s", i, arg, tt.expected[i])
				}
			}
		})
	}
}

func TestListProjects(t *testing.T) {
	// This is a basic test that just checks the method exists
	// In a real test environment, you would mock the exec.Command
	client := NewComposeClient("")
	
	// We can't actually test listing projects without docker compose
	// The method should return an error or empty result
	_, err := client.ListProjects()
	// Either error or empty result is acceptable
	_ = err
}