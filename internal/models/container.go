package models

import "strings"

type Container struct {
	Name      string `json:"Name"`
	Image     string `json:"Image"`
	ID        string `json:"ID"`
	Status    string `json:"Status"`
	State     string `json:"State"`
	Service   string `json:"Service"`
	CreatedAt string `json:"CreatedAt"`
	Ports     string `json:"Ports"`
}

func (c Container) IsDind() bool {
	imageLower := strings.ToLower(c.Image)
	return strings.Contains(imageLower, "dind") || strings.Contains(imageLower, "docker:dind")
}
