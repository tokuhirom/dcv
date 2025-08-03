package models

// DockerNetworkList represents a Docker network from 'docker network ls --format json'
type DockerNetworkList struct {
	Name      string `json:"Name"`
	ID        string `json:"ID"`
	CreatedAt string `json:"CreatedAt"`
	Scope     string `json:"Scope"`
	Driver    string `json:"Driver"`
	IPv4      string `json:"IPv4"`
	IPv6      string `json:"IPv6"`
	Internal  string `json:"Internal"`
	Labels    string `json:"Labels"`
}

// ToDockerNetwork converts a DockerNetworkList to DockerNetwork
func (n DockerNetworkList) ToDockerNetwork() DockerNetwork {
	return DockerNetwork{
		Name:     n.Name,
		ID:       n.ID,
		Created:  n.CreatedAt,
		Scope:    n.Scope,
		Driver:   n.Driver,
		Internal: n.Internal == "true",
	}
}