package docker

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
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

func (c *Client) GetContainerLogs(containerID string, follow bool) (*exec.Cmd, error) {
	args := []string{"logs", containerID, "--tail", "1000", "--timestamps"}
	if follow {
		args = append(args, "-f")
	}

	cmd := exec.Command("docker", args...)

	return cmd, nil
}

func (c *Client) ListDindContainers(containerID string) ([]models.Container, error) {
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

func (c *Client) parseDindPSJSON(output []byte) ([]models.Container, error) {
	containers := make([]models.Container, 0)

	// Docker ps outputs each container as a separate JSON object on its own line
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var container models.Container

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
func (c *Client) GetStats() (string, error) {
	output, err := c.executeCaptured("stats", "--no-stream", "--format", "json", "--all")
	if err != nil {
		return "", fmt.Errorf("failed to executeCaptured: %w", err)
	}

	return string(output), nil
}
