package docker

import (
	"os/exec"

	"github.com/tokuhirom/dcv/internal/models"
)

// DockerClient defines the interface for Docker operations
type DockerClient interface {
	Compose(projectName string) ComposeClientInterface
	ListComposeProjects() ([]models.ComposeProject, error)
	GetContainerLogs(containerID string, follow bool) (*exec.Cmd, error)
	ListDindContainers(containerID string) ([]models.DockerContainer, error)
	GetDindContainerLogs(hostContainerID, targetContainerID string, follow bool) (*exec.Cmd, error)
	KillContainer(containerID string) error
	StopContainer(containerID string) error
	StartContainer(containerID string) error
	RestartContainer(containerID string) error
	GetStats() ([]models.ContainerStats, error)
	ListAllContainers(showAll bool) ([]models.DockerContainer, error)
	ListImages(showAll bool) ([]models.DockerImage, error)
	ListNetworks() ([]models.DockerNetwork, error)
	ListContainerFiles(containerID, path string) ([]models.ContainerFile, error)
	ReadContainerFile(containerID, path string) (string, error)
	ExecuteInteractive(containerID string, command []string) error
	ListVolumes() ([]models.DockerVolume, error)
}

// ComposeClientInterface defines the interface for Docker Compose operations
type ComposeClientInterface interface {
	ListContainers(showAll bool) ([]models.ComposeContainer, error)
	KillService(serviceName string) error
	StopService(serviceName string) error
	StartService(serviceName string) error
	RestartService(serviceName string) error
	Top() (string, error)
	Up(detach bool) error
	Down() error
}
