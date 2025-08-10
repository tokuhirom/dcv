package docker

import (
	"testing"
)

// TestNewClient tests the client creation
func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Error("NewClient() returned nil")
	}
}
