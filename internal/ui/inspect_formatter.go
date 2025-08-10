package ui

import (
	"encoding/json"

	"gopkg.in/yaml.v3"
)

// InspectFormatter formats JSON inspect data into a human-readable format
type InspectFormatter struct {
}

// NewInspectFormatter creates a new formatter
func NewInspectFormatter() *InspectFormatter {
	return &InspectFormatter{}
}

// FormatJSON converts JSON data to YAML format for better readability
func (f *InspectFormatter) FormatJSON(jsonData string) (string, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return "", err
	}

	// Handle array of objects (like docker inspect returns)
	if arr, ok := data.([]interface{}); ok && len(arr) > 0 {
		data = arr[0]
	}

	// Convert to YAML
	yamlBytes, err := yaml.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(yamlBytes), nil
}
