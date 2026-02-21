package docker

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"github.com/tokuhirom/dcv/internal/models"
)

type Client struct {
	fileOps *FileOperations
}

func NewClient() *Client {
	return &Client{
		fileOps: NewFileOperations(nil),
	}
}

// ListContainerFiles lists files in a container directory
func (c *Client) ListContainerFiles(containerID, path string) ([]models.ContainerFile, error) {
	container := NewContainer(containerID, "", "", "")
	return c.fileOps.ListFiles(context.TODO(), container, path)
}

// GetFileContent retrieves file content from a container
func (c *Client) GetFileContent(containerID, filePath string) (string, error) {
	return c.fileOps.GetFileContent(context.TODO(), containerID, filePath)
}

// ListComposeContainers lists containers for a Docker Compose project
func (c *Client) ListComposeContainers(projectName string, showAll bool) ([]models.ComposeContainer, error) {
	// Always use JSON format for reliable parsing
	args := []string{"compose", "-p", projectName, "ps", "--format", "json", "--no-trunc"}
	if showAll {
		args = append(args, "--all")
	}

	output, err := ExecuteCaptured(args...)
	if err != nil {
		// Check if it's just empty (no containers)
		if len(output) == 0 || string(output) == "" {
			return []models.ComposeContainer{}, nil
		}
		return nil, fmt.Errorf("failed to ExecuteCaptured docker compose ps: %w\nOutput: %s", err, string(output))
	}

	// Parse JSON format
	slog.Info("Parsing docker compose ps output")
	return ParseComposePSJSON(output)
}

// ListDindContainers lists containers inside a Docker-in-Docker container
func (c *Client) ListDindContainers(hostContainerID string, showAll bool) ([]models.DockerContainer, error) {
	args := []string{"exec", hostContainerID, "docker", "ps", "--format", "json", "--no-trunc"}
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

func (c *Client) ListVolumes() ([]models.DockerVolume, error) {
	// Use docker system df -v to get volume sizes
	output, err := ExecuteCaptured([]string{"system", "df", "-v", "--format", "json"}...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute docker system df -v: %w\nOutput: %s", err, string(output))
	}

	volumes, err := ParseSystemDfVolumes(output)
	if err != nil {
		return nil, err
	}

	if len(volumes) == 0 {
		slog.Info("No volumes found")
	}

	return volumes, nil
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
func (c *Client) GetStats(all bool) ([]models.ContainerStats, error) {
	args := []string{"stats", "--no-stream", "--format", "json"}
	if all {
		args = append(args, "--all")
	}
	output, err := c.ExecuteCaptured(args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	return ParseStatsJSON(output)
}
