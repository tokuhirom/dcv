package docker

import (
	"reflect"
	"testing"

	"github.com/tokuhirom/dcv/internal/models"
)

func TestParsePSJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    []models.DockerContainer
		wantErr bool
	}{
		{
			name: "parse docker ps JSON output inside dind",
			input: []byte(`{"ID":"a1b2c3d4e5f6","Image":"alpine:latest","Names":"test-container","Status":"Up 2 minutes","CreatedAt":"2 minutes ago"}
{"ID":"b2c3d4e5f6g7","Image":"nginx:latest","Names":"web-server","Status":"Up 5 minutes","CreatedAt":"5 minutes ago"}`),
			want: []models.DockerContainer{
				{
					ID:        "a1b2c3d4e5f6",
					Image:     "alpine:latest",
					CreatedAt: "2 minutes ago",
					Status:    "Up 2 minutes",
					Names:     "test-container",
				},
				{
					ID:        "b2c3d4e5f6g7",
					Image:     "nginx:latest",
					CreatedAt: "5 minutes ago",
					Status:    "Up 5 minutes",
					Names:     "web-server",
				},
			},
			wantErr: false,
		},
	}

	c := &Client{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePSJSON(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDindPS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseDindPS() = %v, want %v", got, tt.want)
			}
		})
	}
}
