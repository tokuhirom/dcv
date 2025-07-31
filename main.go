package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tokuhirom/dcv/internal/ui"
)

func main() {
	// Parse command line flags
	var workDir string
	flag.StringVar(&workDir, "C", "", "Run as if dcv was started in <path> instead of the current working directory")
	flag.StringVar(&workDir, "d", "", "Shorthand for -C")
	flag.Parse()

	// If work directory is specified, change to it
	if workDir != "" {
		if err := os.Chdir(workDir); err != nil {
			fmt.Printf("Failed to change directory to %s: %v\n", workDir, err)
			os.Exit(1)
		}
	}

	// Create the initial model
	m := ui.NewModel()

	// Create the program
	p := tea.NewProgram(m, tea.WithAltScreen())

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}