package docker

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/tokuhirom/dcv/internal/models"
)

type ComposeClient struct {
	workDir     string
	projectName string
	composeFile string
	commandLogs []CommandLog
}

// CommandLog represents a command execution log entry
type CommandLog struct {
	Timestamp time.Time
	Command   string
	ExitCode  int
	Output    string
	Error     string
	Duration  time.Duration
}

func NewComposeClient(workDir string) *ComposeClient {
	return &ComposeClient{
		workDir: workDir,
	}
}

func NewComposeClientWithOptions(workDir, projectName, composeFile string) *ComposeClient {
	return &ComposeClient{
		workDir:     workDir,
		projectName: projectName,
		composeFile: composeFile,
	}
}

// ListProjects lists all Docker Compose projects
func (c *ComposeClient) ListProjects() ([]models.ComposeProject, error) {
	cmd := exec.Command("docker", "compose", "ls", "--format", "json")
	if c.workDir != "" {
		cmd.Dir = c.workDir
	}
	
	output, err := c.executeAndLog(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to execute docker compose ls: %w\nOutput: %s", err, string(output))
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

// GetCommandLogs returns the command execution logs
func (c *ComposeClient) GetCommandLogs() []CommandLog {
	return c.commandLogs
}

// executeAndLog executes a command and logs the result
func (c *ComposeClient) executeAndLog(cmd *exec.Cmd) ([]byte, error) {
	startTime := time.Now()
	cmdStr := strings.Join(cmd.Args, " ")
	
	output, err := cmd.CombinedOutput()
	duration := time.Since(startTime)
	
	exitCode := 0
	errorStr := ""
	
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
		errorStr = err.Error()
	}
	
	// Add to command logs
	log := CommandLog{
		Timestamp: startTime,
		Command:   cmdStr,
		ExitCode:  exitCode,
		Output:    string(output),
		Error:     errorStr,
		Duration:  duration,
	}
	
	c.commandLogs = append(c.commandLogs, log)
	
	// Keep only last 100 commands
	if len(c.commandLogs) > 100 {
		c.commandLogs = c.commandLogs[len(c.commandLogs)-100:]
	}
	
	return output, err
}

// buildComposeArgs builds the docker compose command arguments with project and file options
func (c *ComposeClient) buildComposeArgs(baseArgs ...string) []string {
	args := []string{"compose"}
	
	if c.projectName != "" {
		args = append(args, "-p", c.projectName)
	}
	
	if c.composeFile != "" {
		args = append(args, "-f", c.composeFile)
	}
	
	args = append(args, baseArgs...)
	return args
}

func (c *ComposeClient) ListContainers(showAll bool) ([]models.Process, error) {
	// Always use JSON format for reliable parsing
	baseArgs := []string{"ps", "--format", "json"}
	if showAll {
		baseArgs = append(baseArgs, "--all")
	}
	
	args := c.buildComposeArgs(baseArgs...)
	cmd := exec.Command("docker", args...)
	if c.workDir != "" {
		cmd.Dir = c.workDir
	}

	output, err := c.executeAndLog(cmd)
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

		// Use a more robust parsing approach - split by multiple spaces
		// Expected format: NAME IMAGE SERVICE STATUS PORTS
		// Split preserving the column alignment
		parts := strings.Fields(line)
		if len(parts) < 4 {
			continue
		}

		// Find where STATUS starts (contains "Up", "Exited", etc.)
		statusStartIdx := -1
		for i := 2; i < len(parts); i++ {
			if strings.HasPrefix(parts[i], "Up") || strings.HasPrefix(parts[i], "Exited") || strings.HasPrefix(parts[i], "Created") {
				statusStartIdx = i
				break
			}
		}

		if statusStartIdx == -1 || statusStartIdx < 3 {
			continue
		}

		// Extract fields based on position
		name := parts[0]
		image := parts[1]
		service := parts[statusStartIdx-1]
		
		// Build status from statusStartIdx onward, stopping at ports
		statusParts := []string{}
		for i := statusStartIdx; i < len(parts); i++ {
			// Stop if we hit a port (contains "/" for tcp/udp or ":" for port mapping)
			if strings.Contains(parts[i], "/") || strings.Contains(parts[i], ":") {
				break
			}
			statusParts = append(statusParts, parts[i])
		}
		status := strings.Join(statusParts, " ")

		process := models.Process{
			Container: models.Container{
				Name:    name,
				Image:   image,
				Service: service,
				Status:  status,
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
	hasValidJSON := false
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
			// If we have content that's not valid JSON, return error
			if len(line) > 0 && !hasValidJSON {
				return nil, fmt.Errorf("invalid JSON: %v", err)
			}
			continue
		}
		hasValidJSON = true

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

func (c *ComposeClient) GetContainerLogs(serviceName string, follow bool) (*exec.Cmd, error) {
	baseArgs := []string{"logs", "--tail", "1000", "--timestamps"}
	if follow {
		baseArgs = append(baseArgs, "-f")
	}
	baseArgs = append(baseArgs, serviceName)

	args := c.buildComposeArgs(baseArgs...)
	cmd := exec.Command("docker", args...)
	if c.workDir != "" {
		cmd.Dir = c.workDir
	}

	return cmd, nil
}

func (c *ComposeClient) ExecInContainer(containerName string, command []string) (*exec.Cmd, error) {
	baseArgs := []string{"exec", "-T", containerName}
	baseArgs = append(baseArgs, command...)

	args := c.buildComposeArgs(baseArgs...)
	cmd := exec.Command("docker", args...)
	if c.workDir != "" {
		cmd.Dir = c.workDir
	}

	return cmd, nil
}

func (c *ComposeClient) ListDindContainers(containerName string) ([]models.Container, error) {
	// First check if docker daemon is ready
	checkCmd, err := c.ExecInContainer(containerName, []string{"docker", "info"})
	if err != nil {
		return nil, err
	}
	
	checkOutput, checkErr := checkCmd.CombinedOutput()
	if checkErr != nil {
		// Docker daemon not ready
		cmdStr := strings.Join(checkCmd.Args, " ")
		return nil, fmt.Errorf("docker daemon not ready in container %s\nCommand: %s\nOutput: %s\nError: %w", 
			containerName, cmdStr, string(checkOutput), checkErr)
	}

	// First try JSON format
	cmd, err := c.ExecInContainer(containerName, []string{"docker", "ps", "--format", "json"})
	if err != nil {
		return nil, err
	}

	output, err := c.executeAndLog(cmd)
	if err != nil {
		// If JSON format fails, try table format
		cmd, err = c.ExecInContainer(containerName, []string{"docker", "ps"})
		if err != nil {
			return nil, err
		}
		
		output, err = c.executeAndLog(cmd)
		if err != nil {
			// Include the command and output in error message for debugging
			cmdStr := strings.Join(cmd.Args, " ")
			return nil, fmt.Errorf("failed to execute docker ps in dind\nCommand: %s\nOutput: %s\nError: %w", cmdStr, string(output), err)
		}
		
		// Parse table format
		return c.parseDindPS(output)
	}

	// Try to parse as JSON first
	containers, err := c.parseDindPSJSON(output)
	if err != nil {
		// If JSON parsing fails, try table format
		return c.parseDindPS(output)
	}
	
	return containers, nil
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
	args := []string{"docker", "logs", "--tail", "1000", "--timestamps"}
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

func (c *ComposeClient) GetContainerTop(serviceName string) (string, error) {
	args := c.buildComposeArgs("top", serviceName)
	cmd := exec.Command("docker", args...)
	if c.workDir != "" {
		cmd.Dir = c.workDir
	}

	output, err := c.executeAndLog(cmd)
	if err != nil {
		return "", fmt.Errorf("failed to execute docker compose top: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}

func (c *ComposeClient) KillService(serviceName string) error {
	args := c.buildComposeArgs("kill", serviceName)
	cmd := exec.Command("docker", args...)
	if c.workDir != "" {
		cmd.Dir = c.workDir
	}

	output, err := c.executeAndLog(cmd)
	if err != nil {
		return fmt.Errorf("failed to execute docker compose kill: %w\nOutput: %s", err, string(output))
	}

	return nil
}

func (c *ComposeClient) StopService(serviceName string) error {
	args := c.buildComposeArgs("stop", serviceName)
	cmd := exec.Command("docker", args...)
	if c.workDir != "" {
		cmd.Dir = c.workDir
	}

	output, err := c.executeAndLog(cmd)
	if err != nil {
		return fmt.Errorf("failed to execute docker compose stop: %w\nOutput: %s", err, string(output))
	}

	return nil
}

func (c *ComposeClient) StartService(serviceName string) error {
	args := c.buildComposeArgs("start", serviceName)
	cmd := exec.Command("docker", args...)
	if c.workDir != "" {
		cmd.Dir = c.workDir
	}

	output, err := c.executeAndLog(cmd)
	if err != nil {
		return fmt.Errorf("failed to execute docker compose start: %w\nOutput: %s", err, string(output))
	}

	return nil
}

func (c *ComposeClient) RestartService(serviceName string) error {
	args := c.buildComposeArgs("restart", serviceName)
	cmd := exec.Command("docker", args...)
	if c.workDir != "" {
		cmd.Dir = c.workDir
	}

	output, err := c.executeAndLog(cmd)
	if err != nil {
		return fmt.Errorf("failed to execute docker compose restart: %w\nOutput: %s", err, string(output))
	}

	return nil
}

func (c *ComposeClient) RemoveService(serviceName string) error {
	args := c.buildComposeArgs("rm", "-f", serviceName)
	cmd := exec.Command("docker", args...)
	if c.workDir != "" {
		cmd.Dir = c.workDir
	}

	output, err := c.executeAndLog(cmd)
	if err != nil {
		return fmt.Errorf("failed to execute docker compose rm: %w\nOutput: %s", err, string(output))
	}

	return nil
}

func (c *ComposeClient) UpService(serviceName string) error {
	args := c.buildComposeArgs("up", "-d", serviceName)
	cmd := exec.Command("docker", args...)
	if c.workDir != "" {
		cmd.Dir = c.workDir
	}

	output, err := c.executeAndLog(cmd)
	if err != nil {
		return fmt.Errorf("failed to execute docker compose up: %w\nOutput: %s", err, string(output))
	}

	return nil
}
func (c *ComposeClient) GetStats() (string, error) {
	args := c.buildComposeArgs("stats", "--format", "json", "--no-stream", "--all")
	cmd := exec.Command("docker", args...)
	if c.workDir != "" {
		cmd.Dir = c.workDir
	}

	output, err := c.executeAndLog(cmd)
	if err != nil {
		return "", fmt.Errorf("failed to execute docker compose stats: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}