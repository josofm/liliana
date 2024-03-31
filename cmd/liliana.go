package main

import (
	"log"

	"github.com/josofm/liliana/config"
	"github.com/josofm/liliana/internal/app"
)

func main() {
	// Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// Run
	app.Run(cfg)
}
