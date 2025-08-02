package models

// ComposeProject represents a Docker Compose project
type ComposeProject struct {
	Name        string `json:"Name"`
	Status      string `json:"Status"`
	ConfigFiles string `json:"ConfigFiles"`
}
