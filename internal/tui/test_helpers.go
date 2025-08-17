package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	
	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
)

// TestHelpers provides helper functions for testing tview components

// NewTestApplication creates a test application with simulation screen
func NewTestApplication() (*tview.Application, tcell.SimulationScreen) {
	simScreen := tcell.NewSimulationScreen("UTF-8")
	if err := simScreen.Init(); err != nil {
		panic(err)
	}
	simScreen.SetSize(80, 24)
	
	app := tview.NewApplication()
	app.SetScreen(simScreen)
	
	return app, simScreen
}

// CreateTestDockerClient creates a mock Docker client for testing
func CreateTestDockerClient() *docker.Client {
	return docker.NewClient()
}

// CreateTestContainers creates sample containers for testing
func CreateTestContainers() []models.DockerContainer {
	return []models.DockerContainer{
		{
			ID:      "abc123def456",
			Names:   "web-server",
			Image:   "nginx:latest",
			Status:  "Up 2 hours",
			State:   "running",
			Ports:   "80/tcp -> 0.0.0.0:8080",
			Command: "nginx -g daemon off;",
		},
		{
			ID:      "def456ghi789",
			Names:   "database",
			Image:   "postgres:13",
			Status:  "Up 3 hours",
			State:   "running", 
			Ports:   "5432/tcp -> 0.0.0.0:5432",
			Command: "postgres",
		},
		{
			ID:      "ghi789jkl012",
			Names:   "redis-cache",
			Image:   "redis:alpine",
			Status:  "Exited (0) 10 minutes ago",
			State:   "exited",
			Ports:   "",
			Command: "redis-server",
		},
	}
}

// CreateTestImages creates sample images for testing
func CreateTestImages() []models.DockerImage {
	return []models.DockerImage{
		{
			ID:         "sha256:abc123",
			Repository: "nginx",
			Tag:        "latest",
			CreatedAt:  "2024-01-15",
			Size:       "142MB",
		},
		{
			ID:         "sha256:def456",
			Repository: "postgres",
			Tag:        "13",
			CreatedAt:  "2024-01-10",
			Size:       "374MB",
		},
	}
}

// CreateTestNetworks creates sample networks for testing
func CreateTestNetworks() []models.DockerNetwork {
	return []models.DockerNetwork{
		{
			ID:     "net123",
			Name:   "bridge",
			Driver: "bridge",
			Scope:  "local",
		},
		{
			ID:     "net456", 
			Name:   "host",
			Driver: "host",
			Scope:  "local",
		},
	}
}

// CreateTestVolumes creates sample volumes for testing
func CreateTestVolumes() []models.DockerVolume {
	return []models.DockerVolume{
		{
			Name:   "data-volume",
			Driver: "local",
			Scope:  "local",
		},
		{
			Name:   "config-volume",
			Driver: "local", 
			Scope:  "local",
		},
	}
}

// CreateTestProjects creates sample Docker Compose projects for testing
func CreateTestProjects() []models.ComposeProject {
	return []models.ComposeProject{
		{
			Name:   "web-app",
			Status: "running(3)",
		},
		{
			Name:   "backend-services",
			Status: "exited(2)",
		},
	}
}

// SimulateKeyPress simulates a key press event on the screen
func SimulateKeyPress(screen tcell.SimulationScreen, key tcell.Key, ch rune) {
	ev := tcell.NewEventKey(key, ch, tcell.ModNone)
	screen.PostEvent(ev)
}

// GetScreenContent gets the content at a specific position on the screen
func GetScreenContent(screen tcell.SimulationScreen, x, y int) (rune, tcell.Style) {
	ch, _, style, _ := screen.GetContent(x, y)
	return ch, style
}

// GetScreenLine gets an entire line of text from the screen
func GetScreenLine(screen tcell.SimulationScreen, y int) string {
	width, _ := screen.Size()
	var line []rune
	for x := 0; x < width; x++ {
		ch, _, _, _ := screen.GetContent(x, y)
		if ch != 0 && ch != ' ' {
			line = append(line, ch)
		} else if len(line) > 0 && ch == ' ' {
			line = append(line, ch)
		}
	}
	// Trim trailing spaces
	result := string(line)
	for len(result) > 0 && result[len(result)-1] == ' ' {
		result = result[:len(result)-1]
	}
	return result
}