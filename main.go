package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/tokuhirom/dcv/internal/models"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fujiwara/sloghandler"
	"github.com/tokuhirom/dcv/internal/ui"
)

func main() {
	// Parse command-line flags
	var projectName string
	var composeFile string
	var showProjects bool
	var debugLog string
	flag.StringVar(&projectName, "p", "", "Specify project name")
	flag.StringVar(&composeFile, "f", "", "Specify compose file")
	flag.BoolVar(&showProjects, "projects", false, "Show project list at startup")
	flag.StringVar(&debugLog, "debug", "", "enable debug logging to a file")
	flag.Parse()

	setupLog(debugLog)

	initialView, projectName := detectProject(projectName, composeFile, showProjects)

	slog.Info("Starting dcv",
		slog.Int("initialView", int(initialView)),
		slog.String("project", projectName),
		slog.String("composeFile", composeFile))

	// Create the initial model with options
	m := ui.NewModelWithOptions(initialView, projectName)

	// Create the program
	p := tea.NewProgram(m, tea.WithAltScreen())

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}

func detectProject(projectName string, composeFile string, showProjects bool) (ui.ViewType, string) {
	if showProjects {
		slog.Info("Showing project list at startup")
		return ui.ProjectListView, ""
	}

	if projectName != "" {
		slog.Info("Using specified project name",
			slog.String("projectName", projectName))
		return ui.ProcessListView, projectName
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
				return ui.ProcessListView, project.Name
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

		paths := filepath.SplitList(pwd)
		for i := len(paths) - 1; i >= 0; i-- {
			newPath := filepath.Join(paths[0:i]...)
			slog.Debug("Checking path",
				slog.String("basePath", pwd),
				slog.Any("paths", paths),
				slog.Int("depth", i),
				slog.String("path", newPath))
			for _, project := range projects {
				match, err := filepath.Match(fmt.Sprintf("%s/*", newPath), project.ConfigFiles)
				if err != nil {
					panic(fmt.Sprintf("Failed to match path: %v", err))
				}
				if match {
					slog.Info("Found Docker Compose project with config file",
						slog.String("composeFile", project.ConfigFiles),
						slog.String("projectName", project.Name))
					return ui.ProcessListView, project.Name
				}
			}
		}
	}
	slog.Info("No specific project found, showing project list")
	return ui.ProjectListView, ""
}

func setupLog(debugLog string) {
	if debugLog != "" {
		logFile, err := os.OpenFile(debugLog,
			os.O_CREATE|os.O_WRONLY|os.O_APPEND,
			0600)
		if err != nil {
			slog.Error("failed to open log file",
				slog.Any("error", err))
			os.Exit(1)
		}
		opts := &sloghandler.HandlerOptions{
			HandlerOptions: slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
			Color: true,
		}
		handler := sloghandler.NewLogHandler(logFile, opts)

		// カスタムハンドラを使用してロガーを作成
		logger := slog.New(handler)
		slog.SetDefault(logger)
	} else {
		// Set the default logger to discard if no debug log is specified
		slog.SetDefault(slog.New(slog.DiscardHandler))
	}
}
