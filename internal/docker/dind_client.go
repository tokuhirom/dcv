package docker

import (
	"fmt"
	"os/exec"

	"github.com/tokuhirom/dcv/internal/models"
)

type DindClient struct {
	hostContainerID string
}

func (d *DindClient) ListContainers(showAll bool) ([]models.DockerContainer, error) {
	args := []string{"exec", d.hostContainerID, "docker", "ps", "--format", "json", "--no-trunc"}
	if showAll {
		args = append(args, "--all")
	}

	output, err := ExecuteCaptured(args...)
	if err != nil {
		return nil, fmt.Errorf("failed to ExecuteCaptured docker ps: %w\nOutput: %s", err, string(output))
	}

	// Try to parse as JSON first
	containers, err := ParsePSJSON(output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse docker ps JSON output: %w\nOutput: %s", err, string(output))
	}

	return containers, nil
}

func (d *DindClient) Execute(args ...string) *exec.Cmd {
	return Execute(append([]string{"exec", d.hostContainerID, "docker"}, args...)...)
}

func (d *DindClient) ListContainerFiles(containerID, path string) ([]models.ContainerFile, error) {
	// Use ls -la to get detailed file information through the host container
	// docker exec <host-container> docker exec <nested-container> ls -la <path>
	args := []string{"exec", d.hostContainerID, "docker", "exec", containerID, "ls", "-la", path}
	output, err := ExecuteCaptured(args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list files in dind container: %w\nOutput: %s", err, string(output))
	}

	files := models.ParseLsOutput(string(output))
	return files, nil
}

func (d *DindClient) ReadContainerFile(containerID, path string) (string, error) {
	// Read file content through the host container
	// docker exec <host-container> docker exec <nested-container> cat <path>
	args := []string{"exec", d.hostContainerID, "docker", "exec", containerID, "cat", path}
	output, err := ExecuteCaptured(args...)
	if err != nil {
		return "", fmt.Errorf("failed to read file in dind container: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}
