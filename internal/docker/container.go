package docker

import "fmt"

type Container interface {
	Inspect() ([]byte, error)
	ContainerID() string
	GetName() string
}

type ContainerImpl struct {
	containerID string
	client      *Client
	name        string
}

func NewContainer(client *Client, containerID string, name string) Container {
	return ContainerImpl{client: client, containerID: containerID, name: name}
}

func (c ContainerImpl) ContainerID() string {
	return c.containerID
}

func (c ContainerImpl) Inspect() ([]byte, error) {
	captured, err := c.client.ExecuteCaptured("inspect", c.containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute docker inspect: %w\nOutput: %s", err, string(captured))
	}
	return captured, nil
}

func (c ContainerImpl) GetName() string {
	return c.name
}
