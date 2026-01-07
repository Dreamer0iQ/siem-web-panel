package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"siem-project/backend/pkg/api"
	"siem-project/backend/pkg/storage"
)

func main() {
	port := flag.Int("port", 8080, "Port to run the API server on")
	dataDir := flag.String("data", "./data", "Directory for data storage")
	flag.Parse()

	log.Println("Starting SIEM API Server...")
	log.Printf("Data directory: %s", *dataDir)
	log.Printf("Port: %d", *port)

	store, err := storage.NewStorage(*dataDir)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	server := api.NewServer(store, *port)

	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	log.Printf("SIEM API Server running on http://localhost:%d", *port)
	log.Println("Dashboard: http://localhost:%d/", *port)
	log.Println("API endpoint: http://localhost:%d/api", *port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down server...")
	server.Stop()
	fmt.Println("Server stopped")
}
