package docker

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

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

// parseVolumeSize parses a size string like "1.051GB" into bytes
func parseVolumeSize(sizeStr string) int64 {
	if sizeStr == "" || sizeStr == "N/A" {
		return 0
	}

	// Remove any spaces
	sizeStr = strings.TrimSpace(sizeStr)

	// Map of unit suffixes to multipliers
	units := map[string]int64{
		"B":   1,
		"KB":  1024,
		"MB":  1024 * 1024,
		"GB":  1024 * 1024 * 1024,
		"TB":  1024 * 1024 * 1024 * 1024,
		"PB":  1024 * 1024 * 1024 * 1024 * 1024,
		"KiB": 1024,
		"MiB": 1024 * 1024,
		"GiB": 1024 * 1024 * 1024,
		"TiB": 1024 * 1024 * 1024 * 1024,
		"PiB": 1024 * 1024 * 1024 * 1024 * 1024,
	}

	// Try to find a matching unit suffix
	for unit, multiplier := range units {
		if strings.HasSuffix(sizeStr, unit) {
			// Extract the numeric part
			numStr := strings.TrimSuffix(sizeStr, unit)
			if value, err := strconv.ParseFloat(numStr, 64); err == nil {
				return int64(value * float64(multiplier))
			}
		}
	}

	// If no unit found, try to parse as raw number
	if value, err := strconv.ParseInt(sizeStr, 10, 64); err == nil {
		return value
	}

	return 0
}
