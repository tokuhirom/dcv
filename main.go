package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

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

	slog.Info("Starting dcv",
		slog.String("project", projectName),
		slog.String("composeFile", composeFile))

	// Create the initial model with options
	m := ui.NewModelWithOptions(projectName, composeFile, showProjects)

	// Create the program
	p := tea.NewProgram(m, tea.WithAltScreen())

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
