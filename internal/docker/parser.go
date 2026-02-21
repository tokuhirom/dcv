package docker

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/tokuhirom/dcv/internal/models"
)

func ParsePSJSON(output []byte) ([]models.DockerContainer, error) {
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

// ParseStatsJSON parses docker stats JSON output
func ParseStatsJSON(output []byte) ([]models.ContainerStats, error) {
	var stats []models.ContainerStats

	// Docker stats outputs each stat as a separate JSON object on its own line
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var stat models.ContainerStats
		if err := json.Unmarshal(line, &stat); err != nil {
			return nil, fmt.Errorf("failed to parse stats JSON: %w", err)
		}
		stats = append(stats, stat)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return stats, nil
}

func ParseNetworkJSON(output []byte) ([]models.DockerNetwork, error) {
	// Docker network ls outputs each network as a separate JSON object on its own line
	networks := make([]models.DockerNetwork, 0)
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var networkList models.DockerNetworkList
		if err := json.Unmarshal(line, &networkList); err != nil {
			// Skip invalid lines
			continue
		}

		// Convert to DockerNetwork
		network := networkList.ToDockerNetwork()
		networks = append(networks, network)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return networks, nil
}

// ParseComposePSJSON parses docker compose ps JSON output
func ParseComposePSJSON(output []byte) ([]models.ComposeContainer, error) {
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

// ParseVolumeJSON parses docker volume ls JSON output
func ParseVolumeJSON(output []byte) ([]models.DockerVolume, error) {
	// Handle empty output
	if len(output) == 0 || string(output) == "" || string(output) == "\n" {
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
			// Skip invalid lines
			continue
		}

		volumes = append(volumes, volume)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning volume output: %w", err)
	}

	return volumes, nil
}

// systemDfJSON represents the JSON output of docker system df -v --format json
type systemDfJSON struct {
	Volumes []models.DockerVolume `json:"Volumes"`
}

// ParseSystemDfVolumes parses docker system df -v --format json output and extracts volumes
func ParseSystemDfVolumes(output []byte) ([]models.DockerVolume, error) {
	if len(output) == 0 || string(output) == "" || string(output) == "\n" {
		return []models.DockerVolume{}, nil
	}

	var df systemDfJSON
	if err := json.Unmarshal(output, &df); err != nil {
		return nil, fmt.Errorf("error parsing system df JSON: %w", err)
	}

	if df.Volumes == nil {
		return []models.DockerVolume{}, nil
	}

	return df.Volumes, nil
}

// ParseComposeProjectsJSON parses docker compose ls JSON output
func ParseComposeProjectsJSON(output []byte) ([]models.ComposeProject, error) {
	// Handle empty output
	if len(output) == 0 || string(output) == "" || string(output) == "\n" {
		return []models.ComposeProject{}, nil
	}

	// Parse JSON output - docker compose ls returns an array
	var projects []models.ComposeProject
	if err := json.Unmarshal(output, &projects); err != nil {
		// Fallback to line-delimited JSON parsing for older versions
		scanner := bufio.NewScanner(bytes.NewReader(output))
		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}

			var project models.ComposeProject
			if err := json.Unmarshal(line, &project); err != nil {
				return nil, fmt.Errorf("failed to parse project JSON: %w", err)
			}
			projects = append(projects, project)
		}

		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}

	return projects, nil
}

// ParseImagesJSON parses docker images JSON output
func ParseImagesJSON(output []byte) ([]models.DockerImage, error) {
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
