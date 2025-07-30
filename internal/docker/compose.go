package docker

import (
	"bufio"
	"bytes"
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
	cmd := exec.Command("docker", "compose", "ps", "--format", "table")
	if c.workDir != "" {
		cmd.Dir = c.workDir
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute docker compose ps: %w", err)
	}

	return c.parseComposePS(output)
}

func (c *ComposeClient) parseComposePS(output []byte) ([]models.Process, error) {
	var processes []models.Process
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
		if len(fields) < 4 {
			continue
		}

		process := models.Process{
			Container: models.Container{
				Name:   fields[0],
				Image:  fields[1],
				Status: strings.Join(fields[3:], " "),
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
	cmd, err := c.ExecInContainer(containerName, []string{"docker", "ps", "--format", "table"})
	if err != nil {
		return nil, err
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute docker ps in dind: %w", err)
	}

	return c.parseDindPS(output)
}

func (c *ComposeClient) parseDindPS(output []byte) ([]models.Container, error) {
	var containers []models.Container
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

		container := models.Container{
			ID:        fields[0],
			Image:     fields[1],
			CreatedAt: fields[3] + " " + fields[4],
			Status:    fields[5] + " " + fields[6],
			Name:      fields[len(fields)-1],
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