package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fujiwara/sloghandler"

	"github.com/tokuhirom/dcv/internal/config"
	"github.com/tokuhirom/dcv/internal/ui"
)

func main() {
	// Parse command-line flags
	var debugLog string
	flag.StringVar(&debugLog, "debug", "", "enable debug logging to a file")
	flag.Parse()

	setupLog(debugLog)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Determine initial view from config
	var initialView ui.ViewType
	switch cfg.General.InitialView {
	case "compose":
		initialView = ui.ComposeProcessListView
	case "projects":
		initialView = ui.ProjectListView
	case "docker":
		initialView = ui.DockerContainerListView
	case "":
		// Empty is valid, use default
		initialView = ui.DockerContainerListView
	default:
		slog.Warn("Unknown initial_view in config, using default",
			slog.String("initial_view", cfg.General.InitialView),
			slog.String("valid_values", "docker, compose, projects"))
		initialView = ui.DockerContainerListView
	}

	slog.Info("Starting dcv", slog.String("initial_view", cfg.General.InitialView))

	// Create the initial model with configured view
	m := ui.NewModel(initialView, "")

	// Create the program
	p := tea.NewProgram(&m, tea.WithAltScreen())

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
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
