package docker

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"github.com/tokuhirom/dcv/internal/models"
)

type Client struct {
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Compose(projectName string) *ComposeClient {
	return &ComposeClient{
		projectName: projectName,
	}
}

func (c *Client) Dind(dindHostContainerID string) *DindClient {
	return &DindClient{
		hostContainerID: dindHostContainerID,
	}
}

// ListComposeProjects lists all Docker Compose projects
func (c *Client) ListComposeProjects() ([]models.ComposeProject, error) {
	output, err := ExecuteCaptured("compose", "ls", "--format", "json")
	if err != nil {
		return nil, fmt.Errorf("failed to ExecuteCaptured docker compose ls: %w\nOutput: %s", err, string(output))
	}

	return ParseComposeProjectsJSON(output)
}

func (c *Client) Execute(args ...string) *exec.Cmd {
	return Execute(args...)
}

func (c *Client) ExecuteCaptured(args ...string) ([]byte, error) {
	return ExecuteCaptured(args...)
}

func (c *Client) ListContainers(showAll bool) ([]models.DockerContainer, error) {
	args := []string{"ps", "--format", "json", "--no-trunc"}
	if showAll {
		args = append(args, "--all")
	}

	output, err := ExecuteCaptured(args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute docker ps: %w\nOutput: %s", err, string(output))
	}

	containers, err := ParsePSJSON(output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse docker ps JSON output: %w\nOutput: %s", err, string(output))
	}

	return containers, nil
}

func (c *Client) ListImages(showAll bool) ([]models.DockerImage, error) {
	args := []string{"images", "--format", "json"}
	if showAll {
		args = append(args, "--all")
	}

	output, err := ExecuteCaptured(args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute docker images: %w\nOutput: %s", err, string(output))
	}

	return ParseImagesJSON(output)
}

func (c *Client) ListNetworks() ([]models.DockerNetwork, error) {
	output, err := ExecuteCaptured("network", "ls", "--format", "json")
	if err != nil {
		return nil, fmt.Errorf("failed to execute docker network ls: %w\nOutput: %s", err, string(output))
	}

	networks, err := ParseNetworkJSON(output)
	if err != nil {
		return nil, err
	}

	return networks, nil
}

func (c *Client) ListContainerFiles(containerID, path string) ([]models.ContainerFile, error) {
	// Use ls -la to get detailed file information
	output, err := ExecuteCaptured("exec", containerID, "ls", "-la", path)
	if err != nil {
		return nil, fmt.Errorf("failed to list files in container: %w\nOutput: %s", err, string(output))
	}

	files := models.ParseLsOutput(string(output))
	return files, nil
}

func (c *Client) ExecuteInteractive(containerID string, command []string) error {
	// Build docker exec command with -it flags for interactive session
	args := append([]string{"exec", "-it", containerID}, command...)
	cmd := exec.Command("docker", args...)

	// Connect to standard input/output/error
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Log the command
	slog.Info("Executing interactive command",
		slog.String("containerID", containerID),
		slog.String("command", strings.Join(command, " ")))

	// Run the command
	return cmd.Run()
}

// GetStats retrieves container statistics
func (c *Client) GetStats() ([]models.ContainerStats, error) {
	output, err := c.ExecuteCaptured("stats", "--no-stream", "--format", "json", "--all")
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	return ParseStatsJSON(output)
}

// Top shows the running processes of a container
func (c *Client) Top(containerID string) (string, error) {
	output, err := ExecuteCaptured("top", containerID)
	if err != nil {
		return "", fmt.Errorf("failed to execute docker top: %w\nOutput: %s", err, string(output))
	}
	return string(output), nil
}
