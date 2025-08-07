package models

import "strings"

// DockerContainer represents a container from `docker ps --format json`
type DockerContainer struct {
	Command      string `json:"Command"`
	CreatedAt    string `json:"CreatedAt"`
	ID           string `json:"ID"`
	Image        string `json:"Image"`
	Labels       string `json:"Labels"`
	LocalVolumes string `json:"LocalVolumes"`
	Mounts       string `json:"Mounts"`
	Names        string `json:"Names"`
	Networks     string `json:"Networks"`
	Ports        string `json:"Ports"`
	RunningFor   string `json:"RunningFor"`
	Size         string `json:"Size"`
	State        string `json:"State"`
	Status       string `json:"Status"`
}

func (c DockerContainer) IsDind() bool {
	// Since compose containers don't have an Image field in JSON output,
	// we check the container name for dind patterns
	nameLower := strings.ToLower(c.Names)
	return strings.Contains(nameLower, "dind") || strings.Contains(c.Command, "dockerd")
}
