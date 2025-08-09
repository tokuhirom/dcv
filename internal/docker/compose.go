package docker

import (
	"errors"
	"fmt"
	"log/slog"
	"os/exec"

	"github.com/tokuhirom/dcv/internal/models"
)

type ComposeClient struct {
	projectName string
}

func (c *ComposeClient) ExecuteCaptured(args ...string) ([]byte, error) {
	return ExecuteCaptured(args...)
}

func (c *ComposeClient) ListContainers(showAll bool) ([]models.ComposeContainer, error) {
	// Always use JSON format for reliable parsing
	args := []string{"compose", "-p", c.projectName, "ps", "--format", "json", "--no-trunc"}
	if showAll {
		args = append(args, "--all")
	}

	output, err := c.ExecuteCaptured(args...)
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
		return nil, fmt.Errorf("failed to ExecuteCaptured docker compose ps: %w\nOutput: %s", err, string(output))
	}

	// Parse JSON format
	slog.Info("Parsing docker compose ps output")
	return ParseComposePSJSON(output)
}

func (c *ComposeClient) Top(serviceName string) (string, error) {
	output, err := c.ExecuteCaptured("compose", "-p", c.projectName, "top", serviceName)
	if err != nil {
		return "", fmt.Errorf("failed to ExecuteCaptured docker compose top: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}
