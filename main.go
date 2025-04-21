package main

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
)

const (
	serverName = "delemaco"  // Server to ping
	serverUser = "root"      // SSH username
	macAddress = "b8:cb:29:a1:f3:88"  // MAC address of the server
	port       = "8080"  // Port to listen on
)

func main() {
	// Setup template
	if err := setupTemplate(); err != nil {
		log.Fatalf("Failed to setup template: %v", err)
	}

	// Register route handlers
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/boot", bootHandler)
	http.HandleFunc("/confirm-shutdown", confirmShutdownHandler)
	http.HandleFunc("/enter-password", enterPasswordHandler)
	http.HandleFunc("/shutdown", shutdownHandler)

	// Start the server
	listenAddr := fmt.Sprintf(":%s", port)
	log.Printf("Starting WOL Server on http://localhost%s", listenAddr)

	if runtime.GOOS == "darwin" {
		log.Println("Running on macOS - commands will be executed using the provided password")
	}

	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
