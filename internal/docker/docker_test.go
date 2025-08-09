package docker

import (
	"reflect"
	"testing"

	"github.com/tokuhirom/dcv/internal/models"
)

func TestParseComposePSJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    []models.ComposeContainer
		wantErr bool
	}{
		{
			name: "parse docker compose ps JSON output (line-delimited)",
			input: []byte(`{"Name": "web-1", "Command": "/docker-entrypoint.sh nginx -g 'daemon off;'", "State": "running", "Service": "web", "ID": "abc123", "Project": "myproject", "Health": "", "ExitCode": 0, "Publishers": null}
{"Name": "dind-1", "Command": "dockerd", "State": "running", "Service": "dind", "ID": "def456", "Project": "myproject", "Health": "", "ExitCode": 0, "Publishers": null}`),
			want: []models.ComposeContainer{
				{
					Name:     "web-1",
					Command:  "/docker-entrypoint.sh nginx -g 'daemon off;'",
					State:    "running",
					Service:  "web",
					ID:       "abc123",
					Project:  "myproject",
					Health:   "",
					ExitCode: 0,
				},
				{
					Name:     "dind-1",
					Command:  "dockerd",
					State:    "running",
					Service:  "dind",
					ID:       "def456",
					Project:  "myproject",
					Health:   "",
					ExitCode: 0,
				},
			},
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			input:   []byte("not json"),
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseComposePSJSON(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseComposePSJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseComposePSJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}
