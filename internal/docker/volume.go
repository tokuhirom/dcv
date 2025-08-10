package docker

import (
	"fmt"
	"log/slog"

	"github.com/tokuhirom/dcv/internal/models"
)

// ListVolumes lists all Docker volumes
func (c *Client) ListVolumes() ([]models.DockerVolume, error) {
	// Use docker volume ls with JSON format
	output, err := ExecuteCaptured([]string{"volume", "ls", "--format", "json"}...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute docker volume ls: %w\nOutput: %s", err, string(output))
	}

	// Handle empty output
	volumes, err := ParseVolumeJSON(output)
	if err != nil {
		return nil, err
	}

	if len(volumes) == 0 {
		slog.Info("No volumes found")
	}

	return volumes, nil
}
