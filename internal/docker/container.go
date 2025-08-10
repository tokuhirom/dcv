package docker

import (
	"fmt"
)

// Container represents a Docker container with command execution capabilities
type Container struct {
	containerID string
	name        string
	title       string
	state       string
	// For DinD containers
	hostContainerName string
	hostContainerID   string
	isDind            bool
}

// NewContainer creates a regular container
func NewContainer(client *Client, containerID string, name string, title string, state string) *Container {
	return &Container{
		containerID: containerID,
		name:        name,
		title:       title,
		state:       state,
		isDind:      false,
	}
}

// NewDindContainer creates a Docker-in-Docker container
func NewDindContainer(client *Client, hostContainerID, hostContainerName, containerID, name, state string) *Container {
	return &Container{
		containerID:       containerID,
		name:              name,
		title:             fmt.Sprintf("DinD: %s (%s)", hostContainerName, name),
		state:             state,
		hostContainerName: hostContainerName,
		hostContainerID:   hostContainerID,
		isDind:            true,
	}
}

func (c *Container) ContainerID() string {
	return c.containerID
}

func (c *Container) GetName() string {
	return c.name
}

func (c *Container) GetContainerID() string {
	return c.containerID
}

func (c *Container) GetState() string {
	return c.state
}

func (c *Container) Title() string {
	return c.title
}

func (c *Container) OperationArgs(cmd string, extraArgs ...string) []string {
	if c.isDind {
		// For DinD containers, we need to exec into the host container first,
		// then run docker commands inside it
		// docker exec <host> docker <cmd> <container> <extraArgs>
		args := []string{"exec", c.hostContainerID, "docker", cmd, c.containerID}
		return append(args, extraArgs...)
	}
	// For regular containers: [cmd, containerID, extraArgs...]
	args := []string{cmd, c.containerID}
	return append(args, extraArgs...)
}
