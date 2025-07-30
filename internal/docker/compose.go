package docker

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/tokuhirom/dcv/internal/models"
)

type ComposeClient struct {
	workDir string
}

func NewComposeClient(workDir string) *ComposeClient {
	return &ComposeClient{
		workDir: workDir,
	}
}

func (c *ComposeClient) ListContainers() ([]models.Process, error) {
	// Always use JSON format for reliable parsing
	cmd := exec.Command("docker", "compose", "ps", "--format", "json")
	if c.workDir != "" {
		cmd.Dir = c.workDir
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if docker compose is available
		if execErr, ok := err.(*exec.ExitError); ok {
			if string(execErr.Stderr) != "" {
				return nil, fmt.Errorf("docker compose error: %s", execErr.Stderr)
			}
		}
		// Check if it's just empty (no containers)
		if len(output) == 0 || string(output) == "" {
			return []models.Process{}, nil
		}
		return nil, fmt.Errorf("failed to execute docker compose ps: %w\nOutput: %s", err, string(output))
	}

	// Handle empty output (no containers running)
	if len(output) == 0 || string(output) == "" || string(output) == "\n" {
		return []models.Process{}, nil
	}

	// Parse JSON format
	return c.parseComposePSJSON(output)
}

func (c *ComposeClient) parseComposePS(output []byte) ([]models.Process, error) {
	processes := []models.Process{}
	scanner := bufio.NewScanner(bytes.NewReader(output))

	// Skip header
	if scanner.Scan() {
		scanner.Text() // Skip the header line
	}

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Split by multiple spaces to handle columns properly
		// Expected format: NAME IMAGE COMMAND SERVICE CREATED STATUS PORTS
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}

		// Find STATUS field - it starts with "Up" or "Exited"
		statusIndex := -1
		for i := 4; i < len(fields); i++ {
			if strings.HasPrefix(fields[i], "Up") || strings.HasPrefix(fields[i], "Exited") || strings.HasPrefix(fields[i], "Created") {
				statusIndex = i
				break
			}
		}

		if statusIndex == -1 {
			continue
		}

		// Extract status (from statusIndex to the end or before PORTS)
		statusEnd := len(fields)
		for i := statusIndex + 1; i < len(fields); i++ {
			// Check if this might be a port (contains : or /)
			if strings.Contains(fields[i], ":") || strings.Contains(fields[i], "/") {
				statusEnd = i
				break
			}
		}

		process := models.Process{
			Container: models.Container{
				Name:    fields[0],
				Image:   fields[1],
				Service: fields[3],
				Status:  strings.Join(fields[statusIndex:statusEnd], " "),
			},
		}

		// Detect dind containers by image name
		imageLower := strings.ToLower(process.Image)
		if strings.Contains(imageLower, "dind") || strings.Contains(imageLower, "docker:dind") {
			process.IsDind = true
		}

		processes = append(processes, process)
	}

	return processes, scanner.Err()
}

func (c *ComposeClient) parseComposePSJSON(output []byte) ([]models.Process, error) {
	processes := []models.Process{}
	
	// Docker compose outputs each container as a separate JSON object on its own line
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var container struct {
			Name    string `json:"Name"`
			Image   string `json:"Image"`
			Status  string `json:"Status"`
			State   string `json:"State"`
			Service string `json:"Service"`
			ID      string `json:"ID"`
		}

		if err := json.Unmarshal(line, &container); err != nil {
			// Skip invalid lines
			continue
		}

		process := models.Process{
			Container: models.Container{
				Name:    container.Name,
				Image:   container.Image,
				Status:  container.Status,
				State:   container.State,
				Service: container.Service,
				ID:      container.ID,
			},
		}

		// Detect dind containers by image name
		imageLower := strings.ToLower(process.Image)
		if strings.Contains(imageLower, "dind") || strings.Contains(imageLower, "docker:dind") {
			process.IsDind = true
		}

		processes = append(processes, process)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return processes, nil
}

func (c *ComposeClient) GetContainerLogs(containerName string, follow bool) (*exec.Cmd, error) {
	args := []string{"compose", "logs"}
	if follow {
		args = append(args, "-f")
	}
	args = append(args, containerName)

	cmd := exec.Command("docker", args...)
	if c.workDir != "" {
		cmd.Dir = c.workDir
	}

	return cmd, nil
}

func (c *ComposeClient) ExecInContainer(containerName string, command []string) (*exec.Cmd, error) {
	args := []string{"compose", "exec", "-T", containerName}
	args = append(args, command...)

	cmd := exec.Command("docker", args...)
	if c.workDir != "" {
		cmd.Dir = c.workDir
	}

	return cmd, nil
}

func (c *ComposeClient) ListDindContainers(containerName string) ([]models.Container, error) {
	cmd, err := c.ExecInContainer(containerName, []string{"docker", "ps", "--format", "json"})
	if err != nil {
		return nil, err
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute docker ps in dind: %w", err)
	}

	return c.parseDindPSJSON(output)
}

func (c *ComposeClient) parseDindPS(output []byte) ([]models.Container, error) {
	containers := []models.Container{}
	scanner := bufio.NewScanner(bytes.NewReader(output))

	// Skip header
	if scanner.Scan() {
		scanner.Text() // Skip the header line
	}

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 7 {
			continue
		}

		// Parse the fields carefully
		// CONTAINER ID   IMAGE          COMMAND    CREATED         STATUS         PORTS     NAMES
		// a1b2c3d4e5f6   alpine:latest  "/bin/sh"  2 minutes ago   Up 2 minutes             test-container
		
		container := models.Container{
			ID:    fields[0],
			Image: fields[1],
			Name:  fields[len(fields)-1],
		}
		
		// Find CREATED and STATUS fields
		// They are usually "X time ago" and "Up X time" format
		createdIdx := -1
		statusIdx := -1
		
		for i := 3; i < len(fields); i++ {
			if fields[i] == "ago" && createdIdx == -1 {
				// Found end of CREATED field
				createdIdx = i
				container.CreatedAt = strings.Join(fields[3:i+1], " ")
			} else if strings.HasPrefix(fields[i], "Up") && statusIdx == -1 {
				// Found start of STATUS field
				statusIdx = i
				// Find the end of status (before PORTS or NAME)
				statusEnd := i + 2 // Usually "Up X minutes"
				if statusEnd < len(fields)-1 {
					container.Status = strings.Join(fields[i:statusEnd+1], " ")
				} else {
					container.Status = strings.Join(fields[i:len(fields)-1], " ")
				}
				break
			}
		}

		containers = append(containers, container)
	}

	return containers, scanner.Err()
}

func (c *ComposeClient) GetDindContainerLogs(hostContainer, targetContainer string, follow bool) (*exec.Cmd, error) {
	args := []string{"docker", "logs"}
	if follow {
		args = append(args, "-f")
	}
	args = append(args, targetContainer)

	return c.ExecInContainer(hostContainer, args)
}

func (c *ComposeClient) parseDindPSJSON(output []byte) ([]models.Container, error) {
	containers := []models.Container{}
	
	// Docker ps outputs each container as a separate JSON object on its own line
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var container struct {
			ID        string   `json:"ID"`
			Image     string   `json:"Image"`
			Names     string   `json:"Names"`
			Status    string   `json:"Status"`
			CreatedAt string   `json:"CreatedAt"`
			Ports     string   `json:"Ports"`
		}

		if err := json.Unmarshal(line, &container); err != nil {
			// Skip invalid lines
			continue
		}

		containers = append(containers, models.Container{
			ID:        container.ID,
			Image:     container.Image,
			Name:      container.Names,
			Status:    container.Status,
			CreatedAt: container.CreatedAt,
			Ports:     container.Ports,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return containers, nil
}