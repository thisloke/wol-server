package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"

	"github.com/joho/godotenv"
)

// Default values
var (
	serverName = "server"           // Server to ping
	serverUser = "root"               // SSH username
	macAddress = "aa:aa:aa:aa:aa:aa"  // MAC address of the server
	port       = "8080"               // Port to listen on
)

func loadEnvVariables() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default values")
	}

	// Override defaults with environment variables if they exist
	if envServerName := os.Getenv("SERVER_NAME"); envServerName != "" {
		serverName = envServerName
	}

	if envServerUser := os.Getenv("SERVER_USER"); envServerUser != "" {
		serverUser = envServerUser
	}

	if envMacAddress := os.Getenv("MAC_ADDRESS"); envMacAddress != "" {
		macAddress = envMacAddress
	}

	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	log.Printf("Configuration loaded: SERVER_NAME=%s, SERVER_USER=%s, MAC_ADDRESS=%s, PORT=%s",
		serverName, serverUser, macAddress, port)
}

func main() {
	// Load environment variables
	loadEnvVariables()

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
