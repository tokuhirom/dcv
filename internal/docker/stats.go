package docker

import (
	"fmt"

	"github.com/tokuhirom/dcv/internal/models"
)

// GetStats retrieves container statistics
func (c *Client) GetStats() ([]models.ContainerStats, error) {
	output, err := c.ExecuteCaptured("stats", "--no-stream", "--format", "json", "--all")
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	return ParseStatsJSON(output)
}
