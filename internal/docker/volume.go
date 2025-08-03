package docker

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/tokuhirom/dcv/internal/models"
)

// ListVolumes lists all Docker volumes
func (c *Client) ListVolumes() ([]models.DockerVolume, error) {
	// Use docker volume ls with JSON format
	output, err := c.executeCaptured("volume", "ls", "--format", "json")
	if err != nil {
		return nil, fmt.Errorf("failed to execute docker volume ls: %w\nOutput: %s", err, string(output))
	}

	// Handle empty output
	if len(output) == 0 || string(output) == "" || string(output) == "\n" {
		slog.Info("No volumes found")
		return []models.DockerVolume{}, nil
	}

	// Parse line-delimited JSON
	volumes := make([]models.DockerVolume, 0)
	scanner := bufio.NewScanner(bytes.NewReader(output))

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var volume models.DockerVolume
		if err := json.Unmarshal(line, &volume); err != nil {
			slog.Warn("Failed to parse volume JSON",
				slog.String("line", string(line)),
				slog.String("error", err.Error()))
			continue
		}

		volumes = append(volumes, volume)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning volume output: %w", err)
	}

	// Get size information using docker system df
	sizeInfo, err := c.getVolumeSizes()
	if err != nil {
		slog.Warn("Failed to get volume sizes", slog.String("error", err.Error()))
	} else {
		// Map sizes to volumes
		sizeMap := make(map[string]models.DockerVolumeSize)
		for _, size := range sizeInfo {
			sizeMap[size.Name] = size
		}

		for i := range volumes {
			if size, ok := sizeMap[volumes[i].Name]; ok {
				volumes[i].Size = size.Size
				volumes[i].RefCount = size.RefCount
			}
		}
	}

	return volumes, nil
}

// getVolumeSizes gets volume size information using docker system df
func (c *Client) getVolumeSizes() ([]models.DockerVolumeSize, error) {
	output, err := c.executeCaptured("system", "df", "--format", "json", "-v")
	if err != nil {
		return nil, fmt.Errorf("failed to execute docker system df: %w", err)
	}

	var systemDf struct {
		Volumes []models.DockerVolumeSize `json:"Volumes"`
	}

	if err := json.Unmarshal(output, &systemDf); err != nil {
		return nil, fmt.Errorf("failed to parse docker system df output: %w", err)
	}

	return systemDf.Volumes, nil
}

// RemoveVolume removes a Docker volume
func (c *Client) RemoveVolume(volumeName string, force bool) error {
	args := []string{"volume", "rm"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, volumeName)

	output, err := c.executeCaptured(args...)
	if err != nil {
		return fmt.Errorf("failed to remove volume %s: %w\nOutput: %s", volumeName, err, string(output))
	}

	slog.Info("Removed volume",
		slog.String("volumeName", volumeName),
		slog.String("output", string(output)))

	return nil
}

// InspectVolume inspects a Docker volume and returns the formatted JSON
func (c *Client) InspectVolume(volumeName string) (string, error) {
	output, err := c.executeCaptured("volume", "inspect", volumeName)
	if err != nil {
		return "", fmt.Errorf("failed to inspect volume %s: %w\nOutput: %s", volumeName, err, string(output))
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
