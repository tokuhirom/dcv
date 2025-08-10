package terminal

import (
	"runtime"
	"testing"
)

func TestDetectTerminal(t *testing.T) {
	term, err := DetectTerminal()

	switch runtime.GOOS {
	case "darwin":
		if err != nil {
			t.Errorf("DetectTerminal failed on macOS: %v", err)
		}
		if term == nil {
			t.Error("DetectTerminal returned nil on macOS")
		} else if term.Name != "Terminal" && term.Name != "iTerm2" {
			t.Errorf("Unexpected terminal on macOS: %s", term.Name)
		}
	case "linux":
		// On Linux, it might not find a terminal in CI environment
		if err != nil && err.Error() != "no supported terminal emulator found" {
			t.Errorf("Unexpected error on Linux: %v", err)
		}
	case "windows":
		if err != nil {
			t.Errorf("DetectTerminal failed on Windows: %v", err)
		}
		if term == nil {
			t.Error("DetectTerminal returned nil on Windows")
		} else if term.Name != "Windows Terminal" && term.Name != "Command Prompt" {
			t.Errorf("Unexpected terminal on Windows: %s", term.Name)
		}
	default:
		if err == nil {
			t.Errorf("Expected error for unsupported OS %s", runtime.GOOS)
		}
	}
}

func TestBuildDockerCommand(t *testing.T) {
	tests := []struct {
		name        string
		containerID string
		command     []string
		expected    string
	}{
		{
			name:        "simple command",
			containerID: "abc123",
			command:     []string{"/bin/sh"},
			expected:    "docker exec -it abc123 /bin/sh",
		},
		{
			name:        "command with spaces",
			containerID: "def456",
			command:     []string{"/bin/bash", "-c", "echo hello"},
			expected:    "docker exec -it def456 /bin/bash -c 'echo hello'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildDockerCommand(tt.containerID, tt.command)
			if result != tt.expected {
				t.Errorf("buildDockerCommand() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEscapeAppleScript(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{`with "quotes"`, `with \"quotes\"`},
		{`with \backslash`, `with \\backslash`},
		{`both "quotes" and \backslash`, `both \"quotes\" and \\backslash`},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := escapeAppleScript(tt.input)
			if result != tt.expected {
				t.Errorf("escapeAppleScript(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
