package docker

import (
	"bufio"
	"bytes"
	"encoding/json"
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
		return nil, fmt.Errorf("failed to executeCaptured docker compose ls: %w\nOutput: %s", err, string(output))
	}

	// Handle empty output
	if len(output) == 0 || string(output) == "" || string(output) == "\n" {
		return []models.ComposeProject{}, nil
	}

	// Parse JSON output - docker compose ls returns an array
	var projects []models.ComposeProject
	if err := json.Unmarshal(output, &projects); err != nil {
		// Fallback to line-delimited JSON parsing for older versions
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}

			var project models.ComposeProject
			if err := json.Unmarshal([]byte(line), &project); err != nil {
				return nil, fmt.Errorf("failed to parse project JSON: %w", err)
			}
			projects = append(projects, project)
		}
	}

	return projects, nil
}

func (c *Client) Execute(args ...string) *exec.Cmd {
	return Execute(args...)
}

func (c *Client) ExecuteCaptured(args ...string) ([]byte, error) {
	return ExecuteCaptured(args...)
}

func (c *Client) PauseContainer(containerID string) error {
	output, err := ExecuteCaptured("pause", containerID)
	if err != nil {
		return fmt.Errorf("failed to pause container: %w\nOutput: %s", err, string(output))
	}

	slog.Info("Paused container",
		slog.String("containerID", containerID),
		slog.String("output", string(output)))

	return nil
}

func (c *Client) UnpauseContainer(containerID string) error {
	output, err := ExecuteCaptured([]string{"unpause", containerID}...)
	if err != nil {
		return fmt.Errorf("failed to unpause container: %w\nOutput: %s", err, string(output))
	}

	slog.Info("Unpaused container",
		slog.String("containerID", containerID),
		slog.String("output", string(output)))

	return nil
}

func (c *Client) GetStats() (string, error) {
	output, err := ExecuteCaptured([]string{"stats", "--no-stream", "--format", "json", "--all"}...)
	if err != nil {
		return "", fmt.Errorf("failed to executeCaptured: %w", err)
	}

	return string(output), nil
}

func (c *Client) ListContainers(showAll bool) ([]models.DockerContainer, error) {
	args := []string{"ps", "--format", "json"}
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

	// Docker images outputs each image as a separate JSON object on its own line
	images := make([]models.DockerImage, 0)
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var image models.DockerImage
		if err := json.Unmarshal(line, &image); err != nil {
			// Skip invalid lines
			continue
		}

		images = append(images, image)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return images, nil
}

func (c *Client) RemoveImage(imageID string, force bool) error {
	args := []string{"rmi"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, imageID)

	output, err := ExecuteCaptured(args...)
	if err != nil {
		return fmt.Errorf("failed to remove image: %w\nOutput: %s", err, string(output))
	}

	slog.Info("Removed image",
		slog.String("imageID", imageID),
		slog.String("output", string(output)))

	return nil
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

func (c *Client) RemoveNetwork(networkID string) error {
	output, err := ExecuteCaptured([]string{"network", "rm", networkID}...)
	if err != nil {
		return fmt.Errorf("failed to remove network: %w\nOutput: %s", err, string(output))
	}

	slog.Info("Removed network",
		slog.String("networkID", networkID),
		slog.String("output", string(output)))

	return nil
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
	// Use cat to read file contents
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
