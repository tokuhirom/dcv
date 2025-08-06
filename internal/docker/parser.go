package docker

import (
	"bufio"
	"bytes"
	"encoding/json"

	"github.com/tokuhirom/dcv/internal/models"
)

func ParsePSJSON(output []byte) ([]models.DockerContainer, error) {
	containers := make([]models.DockerContainer, 0)

	// Docker ps outputs each container as a separate JSON object on its own line
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var container models.DockerContainer

		if err := json.Unmarshal(line, &container); err != nil {
			// Skip invalid lines
			continue
		}

		containers = append(containers, container)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return containers, nil
}
