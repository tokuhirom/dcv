package docker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseVolumeSize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{"Empty string", "", 0},
		{"N/A", "N/A", 0},
		{"Plain bytes", "1024", 1024},
		{"Bytes with unit", "100B", 100},
		{"Kilobytes", "2KB", 2048},
		{"Megabytes", "1.5MB", 1572864},
		{"Gigabytes", "1.051GB", 1128502657}, // 1.051 * 1024 * 1024 * 1024
		{"Terabytes", "2TB", 2199023255552},
		{"KibiBytes", "4KiB", 4096},
		{"MebiBytes", "2MiB", 2097152},
		{"GibiBytes", "1GiB", 1073741824},
		{"Decimal gigabytes", "10.5GB", 11274289152},
		{"With spaces", " 100MB ", 104857600},
		{"Invalid format", "abc", 0},
		{"No unit number", "GB", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseVolumeSize(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
