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

func TestParseVolumeJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    []models.DockerVolume
		wantErr bool
	}{
		{
			name:  "empty input",
			input: []byte(""),
			want:  []models.DockerVolume{},
		},
		{
			name:  "newline only",
			input: []byte("\n"),
			want:  []models.DockerVolume{},
		},
		{
			name: "parse docker volume ls JSON output",
			input: []byte(`{"Driver":"local","Labels":"","Mountpoint":"/var/lib/docker/volumes/my-volume/_data","Name":"my-volume","Scope":"local"}
{"Driver":"local","Labels":"com.example.label=value","Mountpoint":"/var/lib/docker/volumes/data-volume/_data","Name":"data-volume","Scope":"local"}`),
			want: []models.DockerVolume{
				{
					Name:       "my-volume",
					Driver:     "local",
					Scope:      "local",
					Labels:     "",
					Mountpoint: "/var/lib/docker/volumes/my-volume/_data",
				},
				{
					Name:       "data-volume",
					Driver:     "local",
					Scope:      "local",
					Labels:     "com.example.label=value",
					Mountpoint: "/var/lib/docker/volumes/data-volume/_data",
				},
			},
		},
		{
			name: "skip invalid JSON lines",
			input: []byte(`{"Name":"volume1","Driver":"local"}
invalid json line
{"Name":"volume2","Driver":"local"}`),
			want: []models.DockerVolume{
				{Name: "volume1", Driver: "local"},
				{Name: "volume2", Driver: "local"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseVolumeJSON(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseVolumeJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseVolumeJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseComposeProjectsJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    []models.ComposeProject
		wantErr bool
	}{
		{
			name:  "empty input",
			input: []byte(""),
			want:  []models.ComposeProject{},
		},
		{
			name:  "array format (newer docker compose)",
			input: []byte(`[{"Name":"myproject","Status":"running(3)","ConfigFiles":"/path/to/docker-compose.yml"},{"Name":"testproject","Status":"exited(1)","ConfigFiles":"/path/to/test.yml"}]`),
			want: []models.ComposeProject{
				{
					Name:        "myproject",
					Status:      "running(3)",
					ConfigFiles: "/path/to/docker-compose.yml",
				},
				{
					Name:        "testproject",
					Status:      "exited(1)",
					ConfigFiles: "/path/to/test.yml",
				},
			},
		},
		{
			name: "line-delimited format (older docker compose)",
			input: []byte(`{"Name":"project1","Status":"running(2)","ConfigFiles":"/app/compose.yml"}
{"Name":"project2","Status":"stopped","ConfigFiles":"/app/docker-compose.yml"}`),
			want: []models.ComposeProject{
				{
					Name:        "project1",
					Status:      "running(2)",
					ConfigFiles: "/app/compose.yml",
				},
				{
					Name:        "project2",
					Status:      "stopped",
					ConfigFiles: "/app/docker-compose.yml",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseComposeProjectsJSON(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseComposeProjectsJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseComposeProjectsJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseImagesJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    []models.DockerImage
		wantErr bool
	}{
		{
			name: "parse docker images JSON output",
			input: []byte(`{"ID":"sha256:abc123","Repository":"alpine","Tag":"latest","CreatedAt":"2024-01-01 12:00:00 +0000 UTC","Size":"5.61MB"}
{"ID":"sha256:def456","Repository":"nginx","Tag":"1.21","CreatedAt":"2024-01-02 13:00:00 +0000 UTC","Size":"142MB"}`),
			want: []models.DockerImage{
				{
					ID:         "sha256:abc123",
					Repository: "alpine",
					Tag:        "latest",
					CreatedAt:  "2024-01-01 12:00:00 +0000 UTC",
					Size:       "5.61MB",
				},
				{
					ID:         "sha256:def456",
					Repository: "nginx",
					Tag:        "1.21",
					CreatedAt:  "2024-01-02 13:00:00 +0000 UTC",
					Size:       "142MB",
				},
			},
		},
		{
			name:  "empty input",
			input: []byte(""),
			want:  []models.DockerImage{},
		},
		{
			name: "skip invalid JSON lines",
			input: []byte(`{"ID":"img1","Repository":"test"}
not valid json
{"ID":"img2","Repository":"test2"}`),
			want: []models.DockerImage{
				{ID: "img1", Repository: "test"},
				{ID: "img2", Repository: "test2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseImagesJSON(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseImagesJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseImagesJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}
