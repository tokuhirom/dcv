package models

// ComposeProject represents a Docker Compose project from `docker compose ls --format json`
type ComposeProject struct {
	Name        string `json:"Name"`
	Status      string `json:"Status"`
	ConfigFiles string `json:"ConfigFiles"`
}
