package project

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/tokuhirom/dcv/internal/models"
	"github.com/tokuhirom/dcv/internal/ui"
)

// DetectProject detects the Docker Compose project based on the provided parameters.
// This method panics if it encounters an error during execution or parsing.
func DetectProject(projectName string, composeFile string, showProjects bool) (ui.ViewType, string) {
	if showProjects {
		slog.Info("Showing project list at startup")
		return ui.ProjectListView, ""
	}

	if projectName != "" {
		slog.Info("Using specified project name",
			slog.String("projectName", projectName))
		return ui.ComposeProcessListView, projectName
	}

	// execute `docker compose ls --format=json` to find projects
	slog.Info("Detecting Docker Compose project")
	cmd := exec.Command("docker", "compose", "ls", "--format=json")
	output, err := cmd.Output()
	if err != nil {
		panic(fmt.Sprintf("Failed to execute Docker Compose ls: %v", err))
	}
	var projects []models.ComposeProject
	if err := json.Unmarshal(output, &projects); err != nil {
		panic(fmt.Sprintf("Failed to parse Docker Compose projects JSON: %v", err))
	}

	if composeFile != "" {
		for _, project := range projects {
			if project.ConfigFiles == composeFile {
				slog.Info("Found Docker Compose project with config file",
					slog.String("composeFile", composeFile),
					slog.String("projectName", project.Name))
				return ui.ComposeProcessListView, project.Name
			}
		}
		panic(fmt.Sprintf("Failed to find Docker Compose project with config file: %s", composeFile))
	} else {
		// Run `docker compose ls --format=json` to list projects.
		// If there's a compose file in the current directory or a parent directory,
		// we can use the project name from that file.
		// If no compose file is found, we show the project list.
		pwd, err := os.Getwd()
		if err != nil {
			panic(fmt.Sprintf("Failed to get current working directory: %v", err))
		}

		// Check current directory and all parent directories
		currentPath := pwd
		for {
			slog.Debug("Checking path",
				slog.String("currentPath", currentPath))

			// Check if any project's config file is in this directory
			for _, project := range projects {
				projectDir := filepath.Dir(project.ConfigFiles)
				if projectDir == currentPath {
					slog.Info("Found Docker Compose project with config file",
						slog.String("composeFile", project.ConfigFiles),
						slog.String("projectName", project.Name))
					return ui.ComposeProcessListView, project.Name
				}
			}

			// Move to parent directory
			parent := filepath.Dir(currentPath)
			if parent == currentPath {
				// Reached root directory
				break
			}
			currentPath = parent
		}
	}
	slog.Info("No specific project found, showing project list")
	return ui.ProjectListView, ""
}
