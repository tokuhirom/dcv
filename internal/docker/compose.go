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

type ComposeClient struct {
	projectName string
}

func (c *ComposeClient) executeCaptured(args ...string) ([]byte, error) {
	slog.Info("Executing docker command",
		slog.String("args", strings.Join(args, " ")))

	cmd := exec.Command("docker", args...)

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
		slog.Info("Executed command",
			slog.String("command", cmdStr),
			slog.Int("exitCode", exitCode),
			slog.String("error", errorStr),
			slog.Duration("duration", duration),
			slog.String("output", string(output)))
		return output, fmt.Errorf("failed to execute command '%s': %w\n\n%s", cmdStr, err, output)
	}

	slog.Info("Executed command",
		slog.String("command", cmdStr),
		slog.Duration("duration", duration),
		slog.String("output", string(output)))

	return output, nil
}

func (c *ComposeClient) ListContainers(showAll bool) ([]models.ComposeContainer, error) {
	// Always use JSON format for reliable parsing
	args := []string{"compose"}
	if c.projectName != "" {
		args = append(args, "-p", c.projectName)
	}
	args = append(args, "ps", "--format", "json", "--no-trunc")
	if showAll {
		args = append(args, "--all")
	}

	output, err := c.executeCaptured(args...)
	if err != nil {
		// Check if docker compose is available
		var execErr *exec.ExitError
		if errors.As(err, &execErr) {
			if string(execErr.Stderr) != "" {
				return nil, fmt.Errorf("docker compose error: %s", execErr.Stderr)
			}
		}
		// Check if it's just empty (no containers)
		if len(output) == 0 || string(output) == "" {
			return []models.ComposeContainer{}, nil
		}
		return nil, fmt.Errorf("failed to execute docker compose ps for project %s: %w\nOutput: %s", c.projectName, err, string(output))
	}

	// Handle empty output (no containers running)
	if len(output) == 0 || string(output) == "" || string(output) == "\n" {
		slog.Info("No containers found")
		return []models.ComposeContainer{}, nil
	}

	// Parse JSON format
	slog.Info("Parsing docker compose ps output")
	return c.parseComposePSJSON(output)
}

func (c *ComposeClient) parseComposePSJSON(output []byte) ([]models.ComposeContainer, error) {
	containers := make([]models.ComposeContainer, 0)

	// Docker compose outputs each container as a separate JSON object on its own line
	scanner := bufio.NewScanner(bytes.NewReader(output))
	hasValidJSON := false
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var container models.ComposeContainer
		if err := json.Unmarshal(line, &container); err != nil {
			// If we have content that's not valid JSON, return error
			if len(line) > 0 && !hasValidJSON {
				return nil, fmt.Errorf("invalid JSON: %v", err)
			}
			continue
		}
		hasValidJSON = true

		containers = append(containers, container)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return containers, nil
}

func (c *ComposeClient) GetContainerTop(serviceName string) (string, error) {
	args := []string{"compose"}
	if c.projectName != "" {
		args = append(args, "-p", c.projectName)
	}
	args = append(args, "top", serviceName)
	output, err := c.executeCaptured(args...)
	if err != nil {
		return "", fmt.Errorf("failed to execute docker compose top for service %s: %w\nOutput: %s", serviceName, err, string(output))
	}

	return string(output), nil
}

func (c *ComposeClient) Down() error {
	args := []string{"compose"}
	if c.projectName != "" {
		args = append(args, "-p", c.projectName)
	}
	args = append(args, "down")
	output, err := c.executeCaptured(args...)
	if err != nil {
		return fmt.Errorf("failed to execute docker compose down for project %s: %w\nOutput: %s", c.projectName, err, string(output))
	}

	slog.Info("Executed docker compose down",
		slog.String("output", string(output)))

	return nil
}

func (c *ComposeClient) UpService(serviceName string) error {
	args := []string{"compose"}
	if c.projectName != "" {
		args = append(args, "-p", c.projectName)
	}
	args = append(args, "up", "-d", serviceName)
	out, err := c.executeCaptured(args...)
	if err != nil {
		return fmt.Errorf("failed to execute docker compose up for service %s: %w\nOutput: %s", serviceName, err, string(out))
	}

	slog.Info("Executed docker compose up for service",
		slog.String("service", serviceName),
		slog.String("output", string(out)))

	return nil
}

func (c *ComposeClient) Up() error {
	args := []string{"compose"}
	if c.projectName != "" {
		args = append(args, "-p", c.projectName)
	}
	args = append(args, "up", "-d")
	out, err := c.executeCaptured(args...)
	if err != nil {
		return fmt.Errorf("failed to execute docker compose up for project %s: %w\nOutput: %s", c.projectName, err, string(out))
	}

	slog.Info("Executed docker compose up",
		slog.String("project", c.projectName),
		slog.String("output", string(out)))

	return nil
}
