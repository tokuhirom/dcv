package models

import (
	"fmt"
	"strings"
)

// ComposeContainer represents a container from `docker compose ps --format json`
type ComposeContainer struct {
	ID         string `json:"ID"`
	Name       string `json:"Name"`
	Command    string `json:"Command"`
	Project    string `json:"Project"`
	Service    string `json:"Service"`
	State      string `json:"State"`
	Health     string `json:"Health"`
	ExitCode   int    `json:"ExitCode"`
	Publishers []struct {
		URL           string `json:"URL"`
		TargetPort    int    `json:"TargetPort"`
		PublishedPort int    `json:"PublishedPort"`
		Protocol      string `json:"Protocol"`
	} `json:"Publishers"`
}

func (c ComposeContainer) IsDind() bool {
	// Since compose containers don't have an Image field in JSON output,
	// we check the container name for dind patterns
	nameLower := strings.ToLower(c.Name)
	return strings.Contains(nameLower, "dind") || strings.Contains(c.Command, "dockerd")
}

// GetPortsString returns a formatted string of the container's ports
func (c ComposeContainer) GetPortsString() string {
	if len(c.Publishers) == 0 {
		return ""
	}
	
	var ports []string
	for _, p := range c.Publishers {
		if p.PublishedPort > 0 {
			ports = append(ports, fmt.Sprintf("%d->%d/%s", p.PublishedPort, p.TargetPort, p.Protocol))
		} else {
			ports = append(ports, fmt.Sprintf("%d/%s", p.TargetPort, p.Protocol))
		}
	}
	return strings.Join(ports, ", ")
}

// GetStatus returns a status string for the container
func (c ComposeContainer) GetStatus() string {
	if c.State == "running" {
		return "Up"
	} else if c.State == "exited" {
		if c.ExitCode == 0 {
			return fmt.Sprintf("Exited (%d)", c.ExitCode)
		}
		return fmt.Sprintf("Exited (%d)", c.ExitCode)
	}
	return c.State
}