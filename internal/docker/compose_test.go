package docker

import (
	"reflect"
	"testing"

	"github.com/tokuhirom/dcv/internal/models"
)

func TestComposeClient_parseComposePS(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    []models.Process
		wantErr bool
	}{
		{
			name: "parse docker compose ps table output",
			input: []byte(`NAME                IMAGE               SERVICE             STATUS              PORTS
web-1               nginx:latest        web                 Up 5 minutes        80/tcp
db-1                postgres:13         db                  Up 5 minutes        5432/tcp
dind-1              docker:dind         dind                Up 5 minutes        2375/tcp, 2376/tcp`),
			want: []models.Process{
				{
					Container: models.Container{
						Name:    "web-1",
						Image:   "nginx:latest",
						Service: "web",
						Status:  "Up 5 minutes",
					},
					IsDind: false,
				},
				{
					Container: models.Container{
						Name:    "db-1",
						Image:   "postgres:13",
						Service: "db",
						Status:  "Up 5 minutes",
					},
					IsDind: false,
				},
				{
					Container: models.Container{
						Name:    "dind-1",
						Image:   "docker:dind",
						Service: "dind",
						Status:  "Up 5 minutes",
					},
					IsDind: true,
				},
			},
			wantErr: false,
		},
		{
			name:    "empty output",
			input:   []byte(""),
			want:    []models.Process{},
			wantErr: false,
		},
	}

	c := &ComposeClient{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.parseComposePS(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseComposePS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseComposePS() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComposeClient_parseComposePSJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    []models.Process
		wantErr bool
	}{
		{
			name: "parse docker compose ps JSON output (line-delimited)",
			input: []byte(`{"Name": "web-1", "Image": "nginx:latest", "Status": "Up 5 minutes", "State": "running", "Service": "web", "ID": "abc123"}
{"Name": "dind-1", "Image": "docker:dind", "Status": "Up 5 minutes", "State": "running", "Service": "dind", "ID": "def456"}`),
			want: []models.Process{
				{
					Container: models.Container{
						Name:    "web-1",
						Image:   "nginx:latest",
						Status:  "Up 5 minutes",
						State:   "running",
						Service: "web",
						ID:      "abc123",
					},
					IsDind: false,
				},
				{
					Container: models.Container{
						Name:    "dind-1",
						Image:   "docker:dind",
						Status:  "Up 5 minutes",
						State:   "running",
						Service: "dind",
						ID:      "def456",
					},
					IsDind: true,
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

	c := &ComposeClient{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.parseComposePSJSON(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseComposePSJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseComposePSJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComposeClient_parseDindPS(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    []models.Container
		wantErr bool
	}{
		{
			name: "parse docker ps table output inside dind",
			input: []byte(`CONTAINER ID   IMAGE          COMMAND    CREATED         STATUS         PORTS     NAMES
a1b2c3d4e5f6   alpine:latest  "/bin/sh"  2 minutes ago   Up 2 minutes             test-container
b2c3d4e5f6g7   nginx:latest   nginx      5 minutes ago   Up 5 minutes   80/tcp    web-server`),
			want: []models.Container{
				{
					ID:        "a1b2c3d4e5f6",
					Image:     "alpine:latest",
					CreatedAt: "2 minutes ago",
					Status:    "Up 2 minutes",
					Name:      "test-container",
				},
				{
					ID:        "b2c3d4e5f6g7",
					Image:     "nginx:latest",
					CreatedAt: "5 minutes ago",
					Status:    "Up 5 minutes",
					Name:      "web-server",
				},
			},
			wantErr: false,
		},
	}

	c := &ComposeClient{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.parseDindPS(tt.input)
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