package docker

import "fmt"

type Container interface {
	Inspect() ([]byte, error)
	GetName() string
	GetContainerID() string
	GetState() string

	// Title returns a title for the container, used in UI
	Title() string

	Top() ([]byte, error)
	OperationArgs(cmd string) []string
}

type ContainerImpl struct {
	containerID string
	client      *Client
	name        string
	title       string
	state       string
}

func NewContainer(client *Client, containerID string, name string, title string, state string) Container {
	return ContainerImpl{
		client:      client,
		containerID: containerID,
		name:        name,
		title:       title,
		state:       state}
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

func (c ContainerImpl) GetState() string {
	return c.state
}

func (c ContainerImpl) Title() string {
	return c.title
}

func (c ContainerImpl) Top() ([]byte, error) {
	return c.client.ExecuteCaptured("top", c.containerID)
}

func (c ContainerImpl) OperationArgs(cmd string) []string {
	return []string{cmd, c.containerID}
}

type DindContainerImpl struct {
	client            *Client
	hostContainerName string
	hostContainerID   string
	containerID       string
	name              string
	state             string
}

func NewDindContainer(client *Client, hostContainerID, hostContainerName, containerID, name, state string) Container {
	return DindContainerImpl{
		client:            client,
		hostContainerID:   hostContainerID,
		hostContainerName: hostContainerName,
		containerID:       containerID,
		name:              name,
		state:             state,
	}
}

func (c DindContainerImpl) GetContainerID() string {
	return c.containerID
}

func (c DindContainerImpl) GetState() string {
	return c.state
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

func (c DindContainerImpl) Top() ([]byte, error) {
	return c.client.ExecuteCaptured("exec", c.hostContainerID, "docker", "top", c.containerID)
}

func (c DindContainerImpl) OperationArgs(op string) []string {
	return []string{"exec", c.hostContainerID, "docker", op, c.containerID}
}
