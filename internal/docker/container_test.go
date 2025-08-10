package docker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewContainer(t *testing.T) {
	client := NewClient()
	containerID := "abc123"
	name := "web-server"
	title := "Web Server Container"
	state := "running"

	container := NewContainer(client, containerID, name, title, state)

	assert.NotNil(t, container)
	assert.Equal(t, containerID, container.containerID)
	assert.Equal(t, name, container.name)
	assert.Equal(t, title, container.title)
	assert.Equal(t, state, container.state)
	assert.False(t, container.isDind)
	assert.Empty(t, container.hostContainerID)
	assert.Empty(t, container.hostContainerName)
}

func TestNewDindContainer(t *testing.T) {
	client := NewClient()
	hostContainerID := "host123"
	hostContainerName := "docker-host"
	containerID := "inner456"
	name := "inner-web"
	state := "running"

	container := NewDindContainer(client, hostContainerID, hostContainerName, containerID, name, state)

	assert.NotNil(t, container)
	assert.Equal(t, containerID, container.containerID)
	assert.Equal(t, name, container.name)
	assert.Equal(t, "DinD: docker-host (inner-web)", container.title)
	assert.Equal(t, state, container.state)
	assert.Equal(t, hostContainerID, container.hostContainerID)
	assert.Equal(t, hostContainerName, container.hostContainerName)
	assert.True(t, container.isDind)
}

func TestContainer_ContainerID(t *testing.T) {
	client := NewClient()
	containerID := "test-container-123"
	container := NewContainer(client, containerID, "test", "Test Container", "running")

	result := container.ContainerID()
	assert.Equal(t, containerID, result)
}

func TestContainer_GetName(t *testing.T) {
	client := NewClient()
	name := "test-service"
	container := NewContainer(client, "abc123", name, "Test Service", "running")

	result := container.GetName()
	assert.Equal(t, name, result)
}

func TestContainer_GetContainerID(t *testing.T) {
	client := NewClient()
	containerID := "def456"
	container := NewContainer(client, containerID, "test", "Test", "stopped")

	result := container.GetContainerID()
	assert.Equal(t, containerID, result)
}

func TestContainer_GetState(t *testing.T) {
	tests := []struct {
		name  string
		state string
	}{
		{"running state", "running"},
		{"stopped state", "stopped"},
		{"paused state", "paused"},
		{"exited state", "exited"},
	}

	client := NewClient()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container := NewContainer(client, "abc123", "test", "Test", tt.state)
			result := container.GetState()
			assert.Equal(t, tt.state, result)
		})
	}
}

func TestContainer_Title(t *testing.T) {
	tests := []struct {
		name          string
		containerFunc func() *Container
		expectedTitle string
	}{
		{
			name: "regular container title",
			containerFunc: func() *Container {
				return NewContainer(NewClient(), "abc123", "web", "Web Server", "running")
			},
			expectedTitle: "Web Server",
		},
		{
			name: "dind container title",
			containerFunc: func() *Container {
				return NewDindContainer(NewClient(), "host123", "docker-host", "inner456", "nginx", "running")
			},
			expectedTitle: "DinD: docker-host (nginx)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container := tt.containerFunc()
			result := container.Title()
			assert.Equal(t, tt.expectedTitle, result)
		})
	}
}

func TestContainer_OperationArgs_RegularContainer(t *testing.T) {
	client := NewClient()
	container := NewContainer(client, "abc123", "web", "Web Server", "running")

	tests := []struct {
		name         string
		cmd          string
		extraArgs    []string
		expectedArgs []string
	}{
		{
			name:         "simple logs command",
			cmd:          "logs",
			extraArgs:    nil,
			expectedArgs: []string{"logs", "abc123"},
		},
		{
			name:         "logs with extra args",
			cmd:          "logs",
			extraArgs:    []string{"--follow", "--tail", "100"},
			expectedArgs: []string{"logs", "abc123", "--follow", "--tail", "100"},
		},
		{
			name:         "exec command",
			cmd:          "exec",
			extraArgs:    []string{"-it", "/bin/bash"},
			expectedArgs: []string{"exec", "abc123", "-it", "/bin/bash"},
		},
		{
			name:         "inspect command",
			cmd:          "inspect",
			extraArgs:    nil,
			expectedArgs: []string{"inspect", "abc123"},
		},
		{
			name:         "stats command with args",
			cmd:          "stats",
			extraArgs:    []string{"--no-stream"},
			expectedArgs: []string{"stats", "abc123", "--no-stream"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := container.OperationArgs(tt.cmd, tt.extraArgs...)
			assert.Equal(t, tt.expectedArgs, result)
		})
	}
}

func TestContainer_OperationArgs_DindContainer(t *testing.T) {
	client := NewClient()
	container := NewDindContainer(client, "host123", "docker-host", "inner456", "nginx", "running")

	tests := []struct {
		name         string
		cmd          string
		extraArgs    []string
		expectedArgs []string
	}{
		{
			name:         "simple logs command in dind",
			cmd:          "logs",
			extraArgs:    nil,
			expectedArgs: []string{"exec", "host123", "docker", "logs", "inner456"},
		},
		{
			name:         "logs with extra args in dind",
			cmd:          "logs",
			extraArgs:    []string{"--follow", "--tail", "100"},
			expectedArgs: []string{"exec", "host123", "docker", "logs", "inner456", "--follow", "--tail", "100"},
		},
		{
			name:         "exec command in dind",
			cmd:          "exec",
			extraArgs:    []string{"-it", "/bin/bash"},
			expectedArgs: []string{"exec", "host123", "docker", "exec", "inner456", "-it", "/bin/bash"},
		},
		{
			name:         "inspect command in dind",
			cmd:          "inspect",
			extraArgs:    nil,
			expectedArgs: []string{"exec", "host123", "docker", "inspect", "inner456"},
		},
		{
			name:         "top command in dind",
			cmd:          "top",
			extraArgs:    nil,
			expectedArgs: []string{"exec", "host123", "docker", "top", "inner456"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := container.OperationArgs(tt.cmd, tt.extraArgs...)
			assert.Equal(t, tt.expectedArgs, result)
		})
	}
}

func TestContainer_OperationArgs_EmptyExtraArgs(t *testing.T) {
	client := NewClient()

	t.Run("regular container with empty args", func(t *testing.T) {
		container := NewContainer(client, "abc123", "web", "Web", "running")
		result := container.OperationArgs("logs", []string{}...)
		expected := []string{"logs", "abc123"}
		assert.Equal(t, expected, result)
	})

	t.Run("dind container with empty args", func(t *testing.T) {
		container := NewDindContainer(client, "host123", "host", "inner456", "inner", "running")
		result := container.OperationArgs("logs", []string{}...)
		expected := []string{"exec", "host123", "docker", "logs", "inner456"}
		assert.Equal(t, expected, result)
	})
}

func TestContainer_OperationArgs_MultipleExtraArgs(t *testing.T) {
	client := NewClient()

	t.Run("regular container with multiple args", func(t *testing.T) {
		container := NewContainer(client, "abc123", "web", "Web", "running")
		result := container.OperationArgs("exec", "-i", "-t", "/bin/sh", "-c", "echo hello")
		expected := []string{"exec", "abc123", "-i", "-t", "/bin/sh", "-c", "echo hello"}
		assert.Equal(t, expected, result)
	})

	t.Run("dind container with multiple args", func(t *testing.T) {
		container := NewDindContainer(client, "host123", "host", "inner456", "inner", "running")
		result := container.OperationArgs("exec", "-i", "-t", "/bin/sh", "-c", "echo hello")
		expected := []string{"exec", "host123", "docker", "exec", "inner456", "-i", "-t", "/bin/sh", "-c", "echo hello"}
		assert.Equal(t, expected, result)
	})
}

func TestContainer_Getters_Consistency(t *testing.T) {
	client := NewClient()
	containerID := "test-container-789"
	name := "consistency-test"
	title := "Consistency Test Container"
	state := "running"

	container := NewContainer(client, containerID, name, title, state)

	// Test that getters return consistent values
	assert.Equal(t, container.ContainerID(), container.GetContainerID())
	assert.Equal(t, containerID, container.ContainerID())
	assert.Equal(t, containerID, container.GetContainerID())
	assert.Equal(t, name, container.GetName())
	assert.Equal(t, title, container.Title())
	assert.Equal(t, state, container.GetState())
}

func TestContainer_DindTitleFormatting(t *testing.T) {
	client := NewClient()

	tests := []struct {
		name              string
		hostContainerName string
		containerName     string
		expectedTitle     string
	}{
		{
			name:              "simple names",
			hostContainerName: "host",
			containerName:     "app",
			expectedTitle:     "DinD: host (app)",
		},
		{
			name:              "complex names with hyphens",
			hostContainerName: "docker-in-docker-host",
			containerName:     "web-application-server",
			expectedTitle:     "DinD: docker-in-docker-host (web-application-server)",
		},
		{
			name:              "names with numbers",
			hostContainerName: "host123",
			containerName:     "app456",
			expectedTitle:     "DinD: host123 (app456)",
		},
		{
			name:              "empty names",
			hostContainerName: "",
			containerName:     "",
			expectedTitle:     "DinD:  ()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container := NewDindContainer(client, "hostID", tt.hostContainerName, "innerID", tt.containerName, "running")
			assert.Equal(t, tt.expectedTitle, container.Title())
		})
	}
}
