package docker

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"

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

// executeCaptured executes a command and logs the result
func (c *Client) executeCaptured(args ...string) ([]byte, error) {
	cmd := c.execute(args...)

	startTime := time.Now()
	cmdStr := strings.Join(cmd.Args, " ")

	output, err := cmd.CombinedOutput()
	duration := time.Since(startTime)

	exitCode := 0
	errorStr := ""

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		}
		errorStr = err.Error()
	}

	slog.Info("Executed command",
		slog.String("command", cmdStr),
		slog.Int("exitCode", exitCode),
		slog.String("error", errorStr),
		slog.Duration("duration", duration),
		slog.String("output", string(output)))

	return output, err
}

func (c *Client) execute(args ...string) *exec.Cmd {
	slog.Info("Executing docker command",
		slog.String("args", strings.Join(args, " ")))

	return exec.Command("docker", args...)
}

// ListComposeProjects lists all Docker Compose projects
func (c *Client) ListComposeProjects() ([]models.ComposeProject, error) {
	output, err := c.executeCaptured("compose", "ls", "--format", "json")
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

func (c *Client) GetContainerLogs(containerID string, follow bool) (*exec.Cmd, error) {
	args := []string{"logs", containerID, "--tail", "1000", "--timestamps"}
	if follow {
		args = append(args, "-f")
	}

	cmd := exec.Command("docker", args...)

	return cmd, nil
}

func (c *Client) ListDindContainers(containerID string) ([]models.DockerContainer, error) {
	output, err := c.executeCaptured("exec", containerID, "docker", "ps", "--format", "json")
	if err != nil {
		return nil, fmt.Errorf("failed to executeCaptured docker ps: %w\nOutput: %s", err, string(output))
	}

	// Try to parse as JSON first
	containers, err := c.parseDindPSJSON(output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse docker ps JSON output: %w\nOutput: %s", err, string(output))
	}

	return containers, nil
}

func (c *Client) GetDindContainerLogs(hostContainerID, targetContainerID string, follow bool) (*exec.Cmd, error) {
	args := []string{"logs", hostContainerID, "docker", "logs", targetContainerID, "--tail", "1000", "--timestamps"}
	if follow {
		args = append(args, "-f")
	}
	cmd := c.execute(args...)

	return cmd, nil
}

func (c *Client) parseDindPSJSON(output []byte) ([]models.DockerContainer, error) {
	containers := make([]models.DockerContainer, 0)

	// Docker ps outputs each container as a separate JSON object on its own line
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var container models.DockerContainer

		if err := json.Unmarshal(line, &container); err != nil {
			// Skip invalid lines
			continue
		}

		containers = append(containers, container)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return containers, nil
}

func (c *Client) KillContainer(containerID string) error {
	output, err := c.executeCaptured("kill", containerID)
	if err != nil {
		return fmt.Errorf("failed to executeCaptured docker compose kill: %w\nOutput: %s", err, string(output))
	}

	return nil
}

func (c *Client) StopContainer(containerID string) error {
	output, err := c.executeCaptured("stop", containerID)
	if err != nil {
		return fmt.Errorf("failed to executeCaptured docker compose stop: %w\nOutput: %s", err, string(output))
	}

	return nil
}

func (c *Client) StartContainer(containerID string) error {
	output, err := c.executeCaptured("start", containerID)
	if err != nil {
		return fmt.Errorf("failed to executeCaptured docker compose start: %w\nOutput: %s", err, string(output))
	}

	slog.Info("Started container",
		slog.String("containerID", containerID),
		slog.String("output", string(output)))

	return nil
}

func (c *Client) RestartContainer(containerID string) error {
	output, err := c.executeCaptured("restart", containerID)
	if err != nil {
		return fmt.Errorf("failed to executeCaptured docker compose restart: %w\nOutput: %s", err, string(output))
	}

	return nil
}

func (c *Client) RemoveContainer(containerID string) error {
	output, err := c.executeCaptured("rm", "-f", containerID)
	if err != nil {
		return fmt.Errorf("failed to executeCaptured docker compose rm: %w\nOutput: %s", err, string(output))
	}

	slog.Info("Removed container",
		slog.String("containerID", containerID),
		slog.String("output", string(output)))

	return nil
}

func (c *ComposeClient) UpService(serviceName string) error {
	out, err := c.executeCaptured("up", "-d", serviceName)
	if err != nil {
		return fmt.Errorf("failed to executeCaptured: %w", err)
	}

	slog.Info("Executed docker compose up",
		slog.String("output", string(out)))

	// TODO: show the result of the up command

	return nil
}

func (c *ComposeClient) Up() error {
	out, err := c.executeCaptured("up", "-d")
	if err != nil {
		return fmt.Errorf("failed to executeCaptured: %w", err)
	}

	slog.Info("Executed docker compose up",
		slog.String("output", string(out)))

	// TODO: show the result/progress of the up command

	return nil
}

func (c *Client) GetStats() (string, error) {
	output, err := c.executeCaptured("stats", "--no-stream", "--format", "json", "--all")
	if err != nil {
		return "", fmt.Errorf("failed to executeCaptured: %w", err)
	}

	return string(output), nil
}

func (c *Client) ListAllContainers(showAll bool) ([]models.DockerContainer, error) {
	args := []string{"ps", "--format", "json"}
	if showAll {
		args = append(args, "--all")
	}

	output, err := c.executeCaptured(args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute docker ps: %w\nOutput: %s", err, string(output))
	}

	// Docker ps outputs each container as a separate JSON object on its own line
	containers := make([]models.DockerContainer, 0)
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var container models.DockerContainer
		if err := json.Unmarshal(line, &container); err != nil {
			// Skip invalid lines
			continue
		}

		containers = append(containers, container)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return containers, nil
}

func (c *Client) ListImages(showAll bool) ([]models.DockerImage, error) {
	args := []string{"images", "--format", "json"}
	if showAll {
		args = append(args, "--all")
	}

	output, err := c.executeCaptured(args...)
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

	output, err := c.executeCaptured(args...)
	if err != nil {
		return fmt.Errorf("failed to remove image: %w\nOutput: %s", err, string(output))
	}

	slog.Info("Removed image",
		slog.String("imageID", imageID),
		slog.String("output", string(output)))

	return nil
}

func (c *Client) ListNetworks() ([]models.DockerNetwork, error) {
	output, err := c.executeCaptured("network", "ls", "--format", "json")
	if err != nil {
		return nil, fmt.Errorf("failed to execute docker network ls: %w\nOutput: %s", err, string(output))
	}

	// Docker network ls outputs each network as a separate JSON object on its own line
	networks := make([]models.DockerNetwork, 0)
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var network models.DockerNetwork
		if err := json.Unmarshal(line, &network); err != nil {
			// Skip invalid lines
			continue
		}

		networks = append(networks, network)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return networks, nil
}

func (c *Client) RemoveNetwork(networkID string) error {
	output, err := c.executeCaptured("network", "rm", networkID)
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
	output, err := c.executeCaptured("exec", containerID, "ls", "-la", path)
	if err != nil {
		return nil, fmt.Errorf("failed to list files in container: %w\nOutput: %s", err, string(output))
	}

	files := models.ParseLsOutput(string(output))
	return files, nil
}

func (c *Client) ReadContainerFile(containerID, path string) (string, error) {
	// Use cat to read file contents
	output, err := c.executeCaptured("exec", containerID, "cat", path)
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

func (c *Client) InspectContainer(containerID string) (string, error) {
	output, err := c.executeCaptured("inspect", containerID)
	if err != nil {
		return "", fmt.Errorf("failed to inspect container: %w\nOutput: %s", err, string(output))
	}
	
	// Pretty format the JSON output
	var jsonData interface{}
	if err := json.Unmarshal(output, &jsonData); err != nil {
		// If we can't parse it, return raw output
		return string(output), nil
	}
	
	prettyJSON, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		// If we can't pretty print, return raw output
		return string(output), nil
	}
	
	return string(prettyJSON), nil
}
