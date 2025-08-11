package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/tokuhirom/dcv/internal/ui"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: generate-docs <output-file>")
		os.Exit(1)
	}

	outputFile := os.Args[1]

	// Generate the documentation
	doc := generateKeymapDocumentation()

	// Write to file
	err := os.WriteFile(outputFile, []byte(doc), 0644)
	if err != nil {
		log.Fatalf("Failed to write documentation: %v", err)
	}

	fmt.Printf("Documentation generated successfully: %s\n", outputFile)
}

func generateKeymapDocumentation() string {
	var sb strings.Builder

	sb.WriteString("# DCV Keyboard Shortcuts and Commands\n\n")
	sb.WriteString("This document lists all keyboard shortcuts and commands available in DCV (Docker Container Viewer).\n\n")
	sb.WriteString("## Global Shortcuts\n\n")
	sb.WriteString("These shortcuts work across all views:\n\n")
	sb.WriteString("| Key | Description | Command |\n")
	sb.WriteString("|-----|-------------|----------|\n")

	// Create a temporary model to extract keymaps
	model := ui.NewModel(ui.ComposeProcessListView)
	model.Init()

	// Get global handlers
	globalHandlers := getGlobalHandlers()
	for _, handler := range globalHandlers {
		keys := strings.Join(handler.Keys, ", ")
		sb.WriteString(fmt.Sprintf("| `%s` | %s | :%s |\n", keys, handler.Description, handler.Command))
	}

	sb.WriteString("\n## View-Specific Shortcuts\n\n")

	// Define views to document
	views := []struct {
		Type        ui.ViewType
		Name        string
		Description string
	}{
		{ui.ComposeProcessListView, "Docker Compose Process List", "View and manage Docker Compose containers"},
		{ui.DockerContainerListView, "Docker Container List", "View and manage all Docker containers"},
		{ui.LogView, "Log View", "View container logs"},
		{ui.TopView, "Top View", "View container process information"},
		{ui.StatsView, "Stats View", "View container resource statistics"},
		{ui.ComposeProjectListView, "Project List", "View and select Docker Compose projects"},
		{ui.ImageListView, "Image List", "View and manage Docker images"},
		{ui.NetworkListView, "Network List", "View and manage Docker networks"},
		{ui.VolumeListView, "Volume List", "View and manage Docker volumes"},
		{ui.FileBrowserView, "File Browser", "Browse files inside containers"},
		{ui.FileContentView, "File Content", "View file contents from containers"},
		{ui.InspectView, "Inspect View", "View detailed container/image/network/volume information"},
		{ui.HelpView, "Help View", "View help information"},
		{ui.DindProcessListView, "Docker in Docker", "View containers inside dind containers"},
		{ui.CommandExecutionView, "Command Execution", "View command execution output"},
	}

	for _, view := range views {
		handlers := model.GetViewKeyHandlers(view.Type)
		if len(handlers) == 0 {
			continue
		}

		sb.WriteString(fmt.Sprintf("### %s\n\n", view.Name))
		sb.WriteString(fmt.Sprintf("%s\n\n", view.Description))
		sb.WriteString("| Key | Description | Command |\n")
		sb.WriteString("|-----|-------------|----------|\n")

		for _, handler := range handlers {
			keys := strings.Join(handler.Keys, ", ")
			command := getCommandName(handler)
			sb.WriteString(fmt.Sprintf("| `%s` | %s | :%s |\n", keys, handler.Description, command))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Command Mode\n\n")
	sb.WriteString("Enter command mode by pressing `:`. Available commands:\n\n")
	sb.WriteString("| Command | Description |\n")
	sb.WriteString("|---------|-------------|\n")
	sb.WriteString("| `:q` or `:quit` | Quit DCV |\n")
	sb.WriteString("| `:q!` or `:quit!` | Force quit without confirmation |\n")
	sb.WriteString("| `:help commands` | List all available commands |\n")
	sb.WriteString("| `:set all` | Show all containers (including stopped) |\n")
	sb.WriteString("| `:set noall` | Hide stopped containers |\n")
	sb.WriteString("\n")

	sb.WriteString("## Tips\n\n")
	sb.WriteString("- Most views support vim-style navigation (`j`/`k` for down/up)\n")
	sb.WriteString("- Press `?` in any view to see context-specific help\n")
	sb.WriteString("- Press `ESC` or `q` to go back to the previous view\n")
	sb.WriteString("- Press `H` to toggle the navigation bar visibility\n")
	sb.WriteString("- In process list views, press `x` to see available actions for a container\n")
	sb.WriteString("\n")

	sb.WriteString("---\n")
	sb.WriteString("*This document is auto-generated. Do not edit manually.*\n")

	return sb.String()
}

// getGlobalHandlers returns the global key handlers
func getGlobalHandlers() []KeyHandler {
	return []KeyHandler{
		{Keys: []string{"q"}, Description: "Quit", Command: "quit"},
		{Keys: []string{":"}, Description: "Command mode", Command: "command-mode"},
		{Keys: []string{"H"}, Description: "Toggle navbar", Command: "toggle-navbar"},
		{Keys: []string{"1"}, Description: "Docker container list", Command: "ps"},
		{Keys: []string{"2"}, Description: "Project list", Command: "compose-ls"},
		{Keys: []string{"3"}, Description: "Docker images", Command: "images"},
		{Keys: []string{"4"}, Description: "Docker networks", Command: "network-ls"},
		{Keys: []string{"5"}, Description: "Docker volumes", Command: "volume-ls"},
		{Keys: []string{"6"}, Description: "Container stats", Command: "stats"},
	}
}

// KeyHandler represents a keyboard shortcut
type KeyHandler struct {
	Keys        []string
	Description string
	Command     string
}

// getCommandName extracts the command name from a handler
func getCommandName(handler ui.KeyConfig) string {
	// This is a simplified version - in reality we'd need to extract
	// the actual command name from the handler function
	desc := strings.ToLower(handler.Description)
	desc = strings.ReplaceAll(desc, " ", "-")
	return desc
}
