package main

import (
	"log"

	"github.com/tokuhirom/dcv/internal/ui"
)

func main() {
	app, err := ui.NewApp()
	if err != nil {
		log.Fatal(err)
	}

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}