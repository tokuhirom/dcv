package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tokuhirom/dcv/internal/ui"
)

func main() {
	// Parse command-line flags
	var projectName string
	var composeFile string
	var showProjects bool
	flag.StringVar(&projectName, "p", "", "Specify project name")
	flag.StringVar(&composeFile, "f", "", "Specify compose file")
	flag.BoolVar(&showProjects, "projects", false, "Show project list at startup")
	flag.Parse()

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
