package docker

import (
	"fmt"
)

type Container interface {
	GetName() string
	GetContainerID() string
	GetState() string

	// Title returns a formatted title string for UI display.
	// For regular containers: returns the container name or title
	// For Compose containers: includes project name (e.g., "project-name/service-name")
	// For DinD containers: includes host container info (e.g., "DinD: host-container (nested-container)")
	Title() string

	// OperationArgs returns Docker command arguments for operations on this container.
	// For regular containers: [cmd, containerID]
	// For DinD containers: [exec, hostID, docker, cmd, containerID]
	OperationArgs(cmd string) []string

	// FileOperationArgs returns Docker command arguments for file operations.
	// For regular containers: [exec, containerID, fileCmd...]
	// For DinD containers: [exec, hostID, docker, exec, containerID, fileCmd...]
	FileOperationArgs(fileCmd ...string) []string
}

type ContainerImpl struct {
	containerID string
	name        string
	title       string
	state       string
}

func NewContainer(client *Client, containerID string, name string, title string, state string) Container {
	return ContainerImpl{
		containerID: containerID,
		name:        name,
		title:       title,
		state:       state}
}

func (c ContainerImpl) ContainerID() string {
	return c.containerID
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

func (c ContainerImpl) OperationArgs(cmd string) []string {
	return []string{cmd, c.containerID}
}

func (c ContainerImpl) FileOperationArgs(fileCmd ...string) []string {
	args := []string{"exec", c.containerID}
	return append(args, fileCmd...)
}

type DindContainerImpl struct {
	hostContainerName string
	hostContainerID   string
	containerID       string
	name              string
	state             string
}

func NewDindContainer(client *Client, hostContainerID, hostContainerName, containerID, name, state string) Container {
	return DindContainerImpl{
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

func (c DindContainerImpl) GetName() string {
	return c.name
}

func (c DindContainerImpl) Title() string {
	return fmt.Sprintf("DinD: %s (%s)", c.hostContainerName, c.name)
}

func (c DindContainerImpl) OperationArgs(op string) []string {
	return []string{"exec", c.hostContainerID, "docker", op, c.containerID}
}

func (c DindContainerImpl) FileOperationArgs(fileCmd ...string) []string {
	args := []string{"exec", c.hostContainerID, "docker", "exec", c.containerID}
	return append(args, fileCmd...)
}
