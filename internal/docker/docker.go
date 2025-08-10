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

// composeClientAdapter wraps Client methods to implement ComposeClientInterface
type composeClientAdapter struct {
	client      *Client
	projectName string
}

func (c *Client) Compose(projectName string) ComposeClientInterface {
	return &composeClientAdapter{
		client:      c,
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

func (c *Client) ReadContainerFile(containerID, path string) (string, error) {
	// Read file content using cat
	output, err := ExecuteCaptured("exec", containerID, "cat", path)
	if err != nil {
		return "", fmt.Errorf("failed to read file in container: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
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

// ListComposeContainers lists containers for a specific Docker Compose project
func (c *Client) ListComposeContainers(projectName string, showAll bool) ([]models.ComposeContainer, error) {
	// Always use JSON format for reliable parsing
	args := []string{"compose", "-p", projectName, "ps", "--format", "json", "--no-trunc"}
	if showAll {
		args = append(args, "--all")
	}

	output, err := c.ExecuteCaptured(args...)
	if err != nil {
		// Check if it's just empty (no containers)
		if len(output) == 0 || string(output) == "" {
			return []models.ComposeContainer{}, nil
		}
		return nil, fmt.Errorf("failed to execute docker compose ps: %w\nOutput: %s", err, string(output))
	}

	// Parse JSON format
	slog.Info("Parsing docker compose ps output")
	return ParseComposePSJSON(output)
}

// Adapter methods for ComposeClientInterface
func (a *composeClientAdapter) ListContainers(showAll bool) ([]models.ComposeContainer, error) {
	return a.client.ListComposeContainers(a.projectName, showAll)
}

// The following methods are required by ComposeClientInterface but not used in the codebase
func (a *composeClientAdapter) Top() (string, error) {
	return "", fmt.Errorf("Top not implemented")
}

func (a *composeClientAdapter) KillService(serviceName string) error {
	return fmt.Errorf("KillService not implemented")
}

func (a *composeClientAdapter) StopService(serviceName string) error {
	return fmt.Errorf("StopService not implemented")
}

func (a *composeClientAdapter) StartService(serviceName string) error {
	return fmt.Errorf("StartService not implemented")
}

func (a *composeClientAdapter) RestartService(serviceName string) error {
	return fmt.Errorf("RestartService not implemented")
}

func (a *composeClientAdapter) Up(detach bool) error {
	return fmt.Errorf("Up not implemented")
}

func (a *composeClientAdapter) Down() error {
	return fmt.Errorf("Down not implemented")
}
