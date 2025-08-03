package models

import "time"

// DockerVolume represents a Docker volume
type DockerVolume struct {
	Name       string            `json:"Name"`
	Driver     string            `json:"Driver"`
	Mountpoint string            `json:"Mountpoint"`
	CreatedAt  time.Time         `json:"CreatedAt"`
	Scope      string            `json:"Scope"`
	Labels     map[string]string `json:"Labels"`
	Options    map[string]string `json:"Options"`
	Size       int64             `json:"-"` // Size in bytes, populated separately
	RefCount   int               `json:"-"` // Reference count, populated separately
}

// GetLabel returns a label value by key, or empty string if not found
func (v DockerVolume) GetLabel(key string) string {
	if v.Labels == nil {
		return ""
	}
	return v.Labels[key]
}

// IsLocal returns true if the volume is using the local driver
func (v DockerVolume) IsLocal() bool {
	return v.Driver == "local"
}

// DockerVolumeSize represents size information from docker system df
type DockerVolumeSize struct {
	Name     string `json:"Name"`
	Size     int64  `json:"Size"`
	RefCount int    `json:"RefCount"`
}