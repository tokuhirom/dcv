package models

// ContainerStats holds resource usage statistics for a container
type ContainerStats struct {
	Container string `json:"Container"`
	Name      string `json:"Name"`
	Service   string `json:"Service"`
	CPUPerc   string `json:"CPUPerc"`
	MemUsage  string `json:"MemUsage"`
	MemPerc   string `json:"MemPerc"`
	NetIO     string `json:"NetIO"`
	BlockIO   string `json:"BlockIO"`
	PIDs      string `json:"PIDs"`
}
