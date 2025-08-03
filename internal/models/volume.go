package models

import (
	"strings"
	"time"
)

// DockerVolume represents a Docker volume
type DockerVolume struct {
	Name       string            `json:"Name"`
	Driver     string            `json:"Driver"`
	Mountpoint string            `json:"Mountpoint"`
	CreatedAt  time.Time         `json:"CreatedAt"`
	Scope      string            `json:"Scope"`
	Labels     string            `json:"Labels"` // Raw labels string from docker
	Options    map[string]string `json:"Options"`
	Size       int64             `json:"-"` // Size in bytes, populated separately
	RefCount   int               `json:"-"` // Reference count, populated separately
	labelMap   map[string]string // Parsed labels, populated on demand
}

// GetLabel returns a label value by key, or empty string if not found
func (v *DockerVolume) GetLabel(key string) string {
	if v.Labels == "" {
		return ""
	}

	// Parse labels on demand
	if v.labelMap == nil {
		v.labelMap = parseLabels(v.Labels)
	}

	return v.labelMap[key]
}

// parseLabels parses a comma-separated label string into a map
func parseLabels(labels string) map[string]string {
	result := make(map[string]string)
	if labels == "" {
		return result
	}

	// Split by comma
	pairs := strings.Split(labels, ",")
	for _, pair := range pairs {
		// Split by equals sign
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}

	return result
}

// IsLocal returns true if the volume is using the local driver
func (v DockerVolume) IsLocal() bool {
	return v.Driver == "local"
}

// DockerVolumeSize represents size information from docker system df
type DockerVolumeSize struct {
	Name  string `json:"Name"`
	Size  string `json:"Size"`  // Size as string like "1.051GB"
	Links string `json:"Links"` // Number of links as string
}
