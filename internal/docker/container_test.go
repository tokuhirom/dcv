package docker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewContainer(t *testing.T) {
	container := NewContainer("abc123", "test-container", "Test Container", "running")

	assert.NotNil(t, container)
	assert.Equal(t, "abc123", container.GetContainerID())
	assert.Equal(t, "test-container", container.GetName())
	assert.Equal(t, "Test Container", container.Title())
	assert.Equal(t, "running", container.GetState())
}

func TestNewDindContainer(t *testing.T) {
	container := NewDindContainer("host123", "host-container", "dind456", "dind-container", "running")

	assert.NotNil(t, container)
	assert.Equal(t, "dind456", container.GetContainerID())
	assert.Equal(t, "dind-container", container.GetName())
	assert.Equal(t, "DinD: dind-container (host-container)", container.Title())
	assert.Equal(t, "running", container.GetState())
}

func TestContainer_ContainerID(t *testing.T) {
	container := NewContainer("abc123", "test-container", "Test Container", "running")
	assert.Equal(t, "abc123", container.ContainerID())
}

func TestContainer_GetName(t *testing.T) {
	container := NewContainer("abc123", "test-container", "Test Container", "running")
	assert.Equal(t, "test-container", container.GetName())
}

func TestContainer_GetContainerID(t *testing.T) {
	container := NewContainer("abc123", "test-container", "Test Container", "running")
	assert.Equal(t, "abc123", container.GetContainerID())
}

func TestContainer_GetState(t *testing.T) {
	container := NewContainer("abc123", "test-container", "Test Container", "running")
	assert.Equal(t, "running", container.GetState())
}

func TestContainer_Title(t *testing.T) {
	t.Run("regular container", func(t *testing.T) {
		container := NewContainer("abc123", "test-container", "Test Container", "running")
		assert.Equal(t, "Test Container", container.Title())
	})

	t.Run("dind container", func(t *testing.T) {
		container := NewDindContainer("host123", "host-container", "dind456", "dind-container", "running")
		assert.Equal(t, "DinD: dind-container (host-container)", container.Title())
	})
}

func TestContainer_OperationArgs(t *testing.T) {
	t.Run("regular container", func(t *testing.T) {
		container := NewContainer("abc123", "test-container", "Test Container", "running")

		args := container.OperationArgs("stop")
		assert.Equal(t, []string{"stop", "abc123"}, args)

		args = container.OperationArgs("logs", "--tail", "100")
		assert.Equal(t, []string{"logs", "abc123", "--tail", "100"}, args)
	})

	t.Run("dind container", func(t *testing.T) {
		container := NewDindContainer("host123", "host-container", "dind456", "dind-container", "running")

		args := container.OperationArgs("stop")
		assert.Equal(t, []string{"exec", "host123", "docker", "stop", "dind456"}, args)

		args = container.OperationArgs("logs", "--tail", "100")
		assert.Equal(t, []string{"exec", "host123", "docker", "logs", "dind456", "--tail", "100"}, args)
	})
}
