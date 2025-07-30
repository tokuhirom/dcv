package main

import (
	"flag"
	"log"
	"os"

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
			log.Fatalf("Failed to change directory to %s: %v", workDir, err)
		}
	}

	app, err := ui.NewApp()
	if err != nil {
		log.Fatal(err)
	}

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}