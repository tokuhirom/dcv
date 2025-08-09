package docker

import "fmt"

type Container interface {
	Inspect() ([]byte, error)
	GetName() string
	GetContainerID() string

	// Title returns a title for the container, used in UI
	Title() string
}

type ContainerImpl struct {
	containerID string
	client      *Client
	name        string
	title       string
}

func NewContainer(client *Client, containerID string, name string, title string) Container {
	return ContainerImpl{client: client, containerID: containerID, name: name, title: title}
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

func (c ContainerImpl) GetContainerID() string {
	return c.containerID
}

func (c ContainerImpl) Title() string {
	return c.title
}

type DindContainerImpl struct {
	client            *Client
	hostContainerName string
	hostContainerID   string
	containerID       string
	name              string
}

func NewDindContainer(client *Client, hostContainerID, hostContainerName, containerID, name string) Container {
	return DindContainerImpl{
		client:            client,
		hostContainerID:   hostContainerID,
		hostContainerName: hostContainerName,
		containerID:       containerID,
		name:              name}
}

func (c DindContainerImpl) GetContainerID() string {
	return c.containerID
}

func (c DindContainerImpl) Inspect() ([]byte, error) {
	captured, err := c.client.ExecuteCaptured("exec", c.hostContainerID, "docker", "inspect", c.containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute docker inspect: %w\nOutput: %s", err, string(captured))
	}
	return captured, nil
}

func (c DindContainerImpl) GetName() string {
	return c.name
}

func (c DindContainerImpl) Title() string {
	return fmt.Sprintf("DinD: %s (%s)", c.hostContainerID, c.name)
}
