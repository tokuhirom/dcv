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
	RemoveContainer(containerID string) error
	GetStats() (string, error)
	ListAllContainers(showAll bool) ([]models.DockerContainer, error)
	ListImages(showAll bool) ([]models.DockerImage, error)
	RemoveImage(imageID string, force bool) error
	ListNetworks() ([]models.DockerNetwork, error)
	RemoveNetwork(networkID string) error
	ListContainerFiles(containerID, path string) ([]models.ContainerFile, error)
	ReadContainerFile(containerID, path string) (string, error)
	ExecuteInteractive(containerID string, command []string) error
	InspectContainer(containerID string) (string, error)
	InspectImage(imageID string) (string, error)
	InspectNetwork(networkID string) (string, error)
	ListVolumes() ([]models.DockerVolume, error)
	RemoveVolume(volumeName string, force bool) error
	InspectVolume(volumeName string) (string, error)
}

// ComposeClientInterface defines the interface for Docker Compose operations
type ComposeClientInterface interface {
	ListContainers(showAll bool) ([]models.ComposeContainer, error)
	KillService(serviceName string) error
	StopService(serviceName string) error
	StartService(serviceName string) error
	RestartService(serviceName string) error
	RemoveService(serviceName string) error
	Top() (string, error)
	Up(detach bool) error
	Down() error
}
