//go:build screenshots
// +build screenshots

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	ansitoimg "github.com/pavelpatrin/go-ansi-to-image"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
	"github.com/tokuhirom/dcv/internal/ui"
)

const (
	width  = 120
	height = 30
)

type screenshot struct {
	name       string
	filename   string
	viewType   ui.ViewType
	setupModel func(*ui.Model)
}

func main() {
	// Create output directory
	outputDir := "docs/screenshots"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Define screenshots to generate
	screenshots := []screenshot{
		{
			name:       "Compose Process List",
			filename:   "compose-process-list.png",
			viewType:   ui.ComposeProcessListView,
			setupModel: setupComposeProcessList,
		},
		{
			name:       "Docker Container List",
			filename:   "docker-container-list.png",
			viewType:   ui.DockerContainerListView,
			setupModel: setupDockerContainerList,
		},
		{
			name:       "Image List",
			filename:   "image-list.png",
			viewType:   ui.ImageListView,
			setupModel: setupImageList,
		},
		{
			name:       "Network List",
			filename:   "network-list.png",
			viewType:   ui.NetworkListView,
			setupModel: setupNetworkList,
		},
		{
			name:       "Volume List",
			filename:   "volume-list.png",
			viewType:   ui.VolumeListView,
			setupModel: setupVolumeList,
		},
		{
			name:       "Project List",
			filename:   "project-list.png",
			viewType:   ui.ComposeProjectListView,
			setupModel: setupProjectList,
		},
		{
			name:       "Log View",
			filename:   "log-view.png",
			viewType:   ui.LogView,
			setupModel: setupLogView,
		},
		{
			name:       "Help View",
			filename:   "help-view.png",
			viewType:   ui.HelpView,
			setupModel: nil, // Help view generates its content automatically
		},
		{
			name:       "Stats View",
			filename:   "stats-view.png",
			viewType:   ui.StatsView,
			setupModel: setupStatsView,
		},
		{
			name:       "Top View",
			filename:   "top-view.png",
			viewType:   ui.TopView,
			setupModel: setupTopView,
		},
	}

	fmt.Println("Generating screenshots...")

	for _, ss := range screenshots {
		fmt.Printf("  Generating %s...\n", ss.name)
		if err := generateScreenshot(ss, outputDir); err != nil {
			log.Printf("Failed to generate %s: %v", ss.filename, err)
		}
	}

	fmt.Println("Screenshots generated successfully!")
}

func generateScreenshot(ss screenshot, outputDir string) error {
	// Create model with the appropriate view
	model := ui.NewModel(ss.viewType)

	// Initialize the model
	model.Init()

	// Set window size
	updatedModel, _ := model.Update(tea.WindowSizeMsg{
		Width:  width,
		Height: height,
	})
	model = updatedModel.(*ui.Model)

	// Setup mock data for the specific view
	if ss.setupModel != nil {
		ss.setupModel(model)
	}

	// Render view
	view := model.View()

	// Create ANSI to image converter with custom config
	config := ansitoimg.DefaultConfig
	config.PageCols = width
	config.PageRows = height
	config.Padding = 20

	converter, err := ansitoimg.NewConverter(config)
	if err != nil {
		return fmt.Errorf("failed to create converter: %w", err)
	}

	// Parse ANSI text
	if err := converter.Parse(view); err != nil {
		return fmt.Errorf("failed to parse ANSI text: %w", err)
	}

	// Convert to PNG
	pngBytes, err := converter.ToPNG()
	if err != nil {
		return fmt.Errorf("failed to convert to PNG: %w", err)
	}

	// Write to file
	outputPath := filepath.Join(outputDir, ss.filename)
	if err := os.WriteFile(outputPath, pngBytes, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// Setup functions for each view - use public methods to set mock data

func setupComposeProcessList(m *ui.Model) {
	vm := m.GetComposeProcessListViewModel()
	vm.SetProjectName("myapp")
	vm.SetComposeContainers([]models.ComposeContainer{
		{
			Service: "web",
			Image:   "myapp/web:latest",
			State:   "running",
			Name:    "myapp-web-1",
			ID:      "abc123def456",
			Publishers: []struct {
				URL           string `json:"URL"`
				TargetPort    int    `json:"TargetPort"`
				PublishedPort int    `json:"PublishedPort"`
				Protocol      string `json:"Protocol"`
			}{
				{URL: "0.0.0.0", TargetPort: 3000, PublishedPort: 3000, Protocol: "tcp"},
			},
		},
		{
			Service: "db",
			Image:   "postgres:15",
			State:   "running",
			Name:    "myapp-db-1",
			ID:      "def456ghi789",
			Publishers: []struct {
				URL           string `json:"URL"`
				TargetPort    int    `json:"TargetPort"`
				PublishedPort int    `json:"PublishedPort"`
				Protocol      string `json:"Protocol"`
			}{
				{TargetPort: 5432, Protocol: "tcp"},
			},
		},
		{
			Service: "redis",
			Image:   "redis:7-alpine",
			State:   "running",
			Name:    "myapp-redis-1",
			ID:      "ghi789jkl012",
			Publishers: []struct {
				URL           string `json:"URL"`
				TargetPort    int    `json:"TargetPort"`
				PublishedPort int    `json:"PublishedPort"`
				Protocol      string `json:"Protocol"`
			}{
				{TargetPort: 6379, Protocol: "tcp"},
			},
		},
		{
			Service:  "worker",
			Image:    "myapp/worker:latest",
			State:    "exited",
			ExitCode: 0,
			Name:     "myapp-worker-1",
			ID:       "jkl012mno345",
		},
	})
}

func setupDockerContainerList(m *ui.Model) {
	vm := m.GetDockerContainerListViewModel()
	vm.SetDockerContainers([]models.DockerContainer{
		{
			ID:         "nginx123abc",
			Image:      "nginx:latest",
			Command:    "nginx -g daemon off;",
			CreatedAt:  "2 hours ago",
			RunningFor: "2 hours",
			Status:     "Up 2 hours",
			Ports:      "0.0.0.0:80->80/tcp",
			Names:      "nginx-server",
			State:      "running",
		},
		{
			ID:         "mysql456def",
			Image:      "mysql:8.0",
			Command:    "docker-entrypoint.sh mysqld",
			CreatedAt:  "3 hours ago",
			RunningFor: "3 hours",
			Status:     "Up 3 hours",
			Ports:      "3306/tcp",
			Names:      "mysql-db",
			State:      "running",
		},
		{
			ID:         "redis789ghi",
			Image:      "redis:alpine",
			Command:    "redis-server",
			CreatedAt:  "1 day ago",
			RunningFor: "24 hours",
			Status:     "Up 24 hours",
			Ports:      "6379/tcp",
			Names:      "redis-cache",
			State:      "running",
		},
		{
			ID:         "app012jkl",
			Image:      "myapp:v1.2.3",
			Command:    "python app.py",
			CreatedAt:  "5 minutes ago",
			RunningFor: "2 minutes",
			Status:     "Exited (1) 2 minutes ago",
			Ports:      "",
			Names:      "myapp-failed",
			State:      "exited",
		},
		{
			ID:         "dind345mno",
			Image:      "docker:dind",
			Command:    "dockerd-entrypoint.sh",
			CreatedAt:  "1 hour ago",
			RunningFor: "1 hour",
			Status:     "Up 1 hour",
			Ports:      "2375/tcp, 2376/tcp",
			Names:      "docker-in-docker",
			State:      "running",
		},
	})
}

func setupImageList(m *ui.Model) {
	vm := m.GetImageListViewModel()
	vm.SetImages([]models.DockerImage{
		{
			Repository:   "nginx",
			Tag:          "latest",
			ID:           "sha256:abc123def456",
			CreatedSince: "2 weeks ago",
			Size:         "187MB",
		},
		{
			Repository:   "postgres",
			Tag:          "15",
			ID:           "sha256:def456ghi789",
			CreatedSince: "3 weeks ago",
			Size:         "379MB",
		},
		{
			Repository:   "redis",
			Tag:          "7-alpine",
			ID:           "sha256:ghi789jkl012",
			CreatedSince: "1 month ago",
			Size:         "40.2MB",
		},
		{
			Repository:   "myapp/web",
			Tag:          "latest",
			ID:           "sha256:jkl012mno345",
			CreatedSince: "1 hour ago",
			Size:         "823MB",
		},
		{
			Repository:   "myapp/worker",
			Tag:          "latest",
			ID:           "sha256:mno345pqr678",
			CreatedSince: "1 hour ago",
			Size:         "756MB",
		},
		{
			Repository:   "<none>",
			Tag:          "<none>",
			ID:           "sha256:pqr678stu901",
			CreatedSince: "3 days ago",
			Size:         "1.24GB",
		},
	})
}

func setupNetworkList(m *ui.Model) {
	vm := m.GetNetworkListViewModel()
	vm.SetNetworks([]models.DockerNetwork{
		{
			ID:     "abc123def456",
			Name:   "bridge",
			Driver: "bridge",
			Scope:  "local",
			Containers: map[string]struct {
				Name        string `json:"Name"`
				EndpointID  string `json:"EndpointID"`
				MacAddress  string `json:"MacAddress"`
				IPv4Address string `json:"IPv4Address"`
				IPv6Address string `json:"IPv6Address"`
			}{
				"nginx123abc": {Name: "nginx-server", EndpointID: "abc123", MacAddress: "02:42:ac:11:00:02", IPv4Address: "172.17.0.2/16"},
				"mysql456def": {Name: "mysql-db", EndpointID: "def456", MacAddress: "02:42:ac:11:00:03", IPv4Address: "172.17.0.3/16"},
				"redis789ghi": {Name: "redis-cache", EndpointID: "ghi789", MacAddress: "02:42:ac:11:00:04", IPv4Address: "172.17.0.4/16"},
			},
		},
		{
			ID:     "def456ghi789",
			Name:   "host",
			Driver: "host",
			Scope:  "local",
			Containers: map[string]struct {
				Name        string `json:"Name"`
				EndpointID  string `json:"EndpointID"`
				MacAddress  string `json:"MacAddress"`
				IPv4Address string `json:"IPv4Address"`
				IPv6Address string `json:"IPv6Address"`
			}{},
		},
		{
			ID:     "ghi789jkl012",
			Name:   "myapp_default",
			Driver: "bridge",
			Scope:  "local",
			Containers: map[string]struct {
				Name        string `json:"Name"`
				EndpointID  string `json:"EndpointID"`
				MacAddress  string `json:"MacAddress"`
				IPv4Address string `json:"IPv4Address"`
				IPv6Address string `json:"IPv6Address"`
			}{
				"abc123def456": {Name: "myapp-web-1", EndpointID: "web123", MacAddress: "02:42:ac:12:00:02", IPv4Address: "172.18.0.2/16"},
				"def456ghi789": {Name: "myapp-db-1", EndpointID: "db456", MacAddress: "02:42:ac:12:00:03", IPv4Address: "172.18.0.3/16"},
				"ghi789jkl012": {Name: "myapp-redis-1", EndpointID: "redis789", MacAddress: "02:42:ac:12:00:04", IPv4Address: "172.18.0.4/16"},
				"jkl012mno345": {Name: "myapp-worker-1", EndpointID: "worker012", MacAddress: "02:42:ac:12:00:05", IPv4Address: "172.18.0.5/16"},
			},
		},
		{
			ID:     "jkl012mno345",
			Name:   "none",
			Driver: "null",
			Scope:  "local",
			Containers: map[string]struct {
				Name        string `json:"Name"`
				EndpointID  string `json:"EndpointID"`
				MacAddress  string `json:"MacAddress"`
				IPv4Address string `json:"IPv4Address"`
				IPv6Address string `json:"IPv6Address"`
			}{},
		},
	})
}

func setupVolumeList(m *ui.Model) {
	vm := m.GetVolumeListViewModel()
	vm.SetVolumes([]models.DockerVolume{
		{
			Driver: "local",
			Name:   "myapp_db_data",
			Scope:  "local",
		},
		{
			Driver: "local",
			Name:   "myapp_redis_data",
			Scope:  "local",
		},
		{
			Driver: "local",
			Name:   "nginx_config",
			Scope:  "local",
		},
		{
			Driver: "local",
			Name:   "jenkins_home",
			Scope:  "local",
		},
		{
			Driver: "local",
			Name:   "postgres_backup",
			Scope:  "local",
		},
	})
}

func setupProjectList(m *ui.Model) {
	vm := m.GetComposeProjectListViewModel()
	vm.SetProjects([]models.ComposeProject{
		{
			Name:        "myapp",
			Status:      "running(4)",
			ConfigFiles: "/home/user/projects/myapp/docker-compose.yml",
		},
		{
			Name:        "frontend",
			Status:      "running(2)",
			ConfigFiles: "/home/user/projects/frontend/docker-compose.yml",
		},
		{
			Name:        "backend",
			Status:      "exited(3)",
			ConfigFiles: "/home/user/projects/backend/docker-compose.yml",
		},
		{
			Name:        "monitoring",
			Status:      "running(5)",
			ConfigFiles: "/home/user/infrastructure/monitoring/docker-compose.yml",
		},
		{
			Name:        "testing",
			Status:      "created(2)",
			ConfigFiles: "/home/user/test/docker-compose.test.yml",
		},
	})
}

func setupLogView(m *ui.Model) {
	vm := m.GetLogViewModel()
	// Create a mock container
	container := docker.NewContainer("abc123def456", "myapp-web-1", "myapp-web-1", "running")
	vm.SetContainer(container)
	logContent := `2024-01-10 10:23:45 [INFO] Starting application server...
2024-01-10 10:23:46 [INFO] Loading configuration from /app/config.yml
2024-01-10 10:23:46 [INFO] Database connection established
2024-01-10 10:23:47 [INFO] Redis cache connected
2024-01-10 10:23:47 [INFO] Starting HTTP server on port 3000
2024-01-10 10:23:48 [INFO] Server is ready to accept connections
2024-01-10 10:24:15 [INFO] GET /api/health 200 15ms
2024-01-10 10:24:30 [INFO] POST /api/users 201 125ms
2024-01-10 10:24:45 [WARN] Slow query detected: SELECT * FROM products (523ms)
2024-01-10 10:25:00 [INFO] GET /api/products 200 531ms
2024-01-10 10:25:15 [ERROR] Failed to process payment: connection timeout
2024-01-10 10:25:16 [INFO] Retrying payment processing...
2024-01-10 10:25:17 [INFO] Payment processed successfully on retry
2024-01-10 10:25:30 [INFO] Background job completed: email_notifications
2024-01-10 10:25:45 [INFO] Cache hit ratio: 87.3%
2024-01-10 10:26:00 [INFO] Active connections: 42
2024-01-10 10:26:15 [DEBUG] Memory usage: 256MB / 512MB
2024-01-10 10:26:30 [INFO] GET /api/dashboard 200 89ms`
	vm.SetLogContent(logContent)
}

func setupStatsView(m *ui.Model) {
	vm := m.GetStatsViewModel()
	vm.SetStats([]models.ContainerStats{
		{
			Container: "abc123def456",
			Name:      "myapp-web-1",
			Service:   "web",
			CPUPerc:   "12.45%",
			MemUsage:  "256.8MiB / 2GiB",
			MemPerc:   "12.54%",
			NetIO:     "1.2MB / 856KB",
			BlockIO:   "45.2MB / 12.3MB",
			PIDs:      "15",
		},
		{
			Container: "def456ghi789",
			Name:      "myapp-db-1",
			Service:   "db",
			CPUPerc:   "3.21%",
			MemUsage:  "512.3MiB / 4GiB",
			MemPerc:   "12.51%",
			NetIO:     "856KB / 2.1MB",
			BlockIO:   "123MB / 456MB",
			PIDs:      "8",
		},
		{
			Container: "ghi789jkl012",
			Name:      "myapp-redis-1",
			Service:   "redis",
			CPUPerc:   "0.52%",
			MemUsage:  "48.2MiB / 512MiB",
			MemPerc:   "9.41%",
			NetIO:     "523KB / 412KB",
			BlockIO:   "0B / 0B",
			PIDs:      "4",
		},
	})
}

func setupTopView(m *ui.Model) {
	vm := m.GetTopViewModel()
	// Set container for the title
	container := docker.NewContainer("abc123def456", "myapp-web-1", "myapp-web-1", "running")
	vm.SetContainer(container)
	vm.SetProcesses([]models.Process{
		{
			UID:   "root",
			PID:   "1",
			PPID:  "0",
			C:     "0",
			STIME: "10:23",
			TTY:   "?",
			TIME:  "00:00:02",
			CMD:   "python app.py",
		},
		{
			UID:   "root",
			PID:   "15",
			PPID:  "1",
			C:     "0",
			STIME: "10:23",
			TTY:   "?",
			TIME:  "00:00:00",
			CMD:   "/usr/local/bin/gunicorn",
		},
		{
			UID:   "root",
			PID:   "16",
			PPID:  "15",
			C:     "2",
			STIME: "10:23",
			TTY:   "?",
			TIME:  "00:00:15",
			CMD:   "/usr/local/bin/gunicorn: worker",
		},
	})
}
