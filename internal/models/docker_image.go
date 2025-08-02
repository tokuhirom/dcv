package models

// DockerImage represents an image from `docker images --format json`
type DockerImage struct {
	Containers   string `json:"Containers"`
	CreatedAt    string `json:"CreatedAt"`
	CreatedSince string `json:"CreatedSince"`
	Digest       string `json:"Digest"`
	ID           string `json:"ID"`
	Repository   string `json:"Repository"`
	SharedSize   string `json:"SharedSize"`
	Size         string `json:"Size"`
	Tag          string `json:"Tag"`
	UniqueSize   string `json:"UniqueSize"`
	VirtualSize  string `json:"VirtualSize"`
}

// GetRepoTag returns the repository:tag string
func (i DockerImage) GetRepoTag() string {
	if i.Repository == "<none>" {
		return i.ID
	}
	if i.Tag == "<none>" {
		return i.Repository
	}
	return i.Repository + ":" + i.Tag
}
