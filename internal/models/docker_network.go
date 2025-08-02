package models

// DockerNetwork represents a Docker network
type DockerNetwork struct {
	Name       string `json:"Name"`
	ID         string `json:"ID"`
	Created    string `json:"Created"`
	Scope      string `json:"Scope"`
	Driver     string `json:"Driver"`
	EnableIPv6 bool   `json:"EnableIPv6"`
	IPAM       struct {
		Driver  string            `json:"Driver"`
		Options map[string]string `json:"Options"`
		Config  []struct {
			Subnet  string `json:"Subnet"`
			Gateway string `json:"Gateway"`
		} `json:"Config"`
	} `json:"IPAM"`
	Internal   bool `json:"Internal"`
	Attachable bool `json:"Attachable"`
	Ingress    bool `json:"Ingress"`
	ConfigFrom struct {
		Network string `json:"Network"`
	} `json:"ConfigFrom"`
	ConfigOnly bool `json:"ConfigOnly"`
	Containers map[string]struct {
		Name        string `json:"Name"`
		EndpointID  string `json:"EndpointID"`
		MacAddress  string `json:"MacAddress"`
		IPv4Address string `json:"IPv4Address"`
		IPv6Address string `json:"IPv6Address"`
	} `json:"Containers"`
	Options map[string]string `json:"Options"`
	Labels  map[string]string `json:"Labels"`
}

// GetSubnet returns the first subnet if available
func (n DockerNetwork) GetSubnet() string {
	if len(n.IPAM.Config) > 0 {
		return n.IPAM.Config[0].Subnet
	}
	return ""
}

// GetContainerCount returns the number of containers attached to the network
func (n DockerNetwork) GetContainerCount() int {
	return len(n.Containers)
}
