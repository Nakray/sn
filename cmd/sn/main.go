package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Nakray/sn/internal/config"
	"github.com/Nakray/sn/internal/database"
	"github.com/Nakray/sn/internal/monitoring"
	"github.com/Nakray/sn/internal/server"
)

func main() {
	configPath := flag.String("config", "config.json", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfgData, err := os.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	var cfg config.Config
	if err := json.Unmarshal(cfgData, &cfg); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	// Initialize database
	db, err := database.New(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Database connected successfully")

	// Initialize monitoring service
	monService := monitoring.NewService(db, &cfg)
	monService.Start()
	defer monService.Stop()

	log.Printf("Monitoring service started with %d workers\n", cfg.Monitoring.Workers)

	// Initialize and start HTTP server
	srv := server.New(db, monService, &cfg)
	go func() {
		log.Printf("Starting HTTP server on port %d\n", cfg.Server.Port)
		if err := srv.Start(); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down gracefully...")
}
