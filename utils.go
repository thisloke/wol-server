package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// StatusData holds data for the HTML template
type StatusData struct {
	Server          string
	Status          string
	Color           string
	IsTestMode      bool
	ConfirmShutdown bool
	AskPassword     bool
	ErrorMessage    string
}

var tmpl *template.Template

// Setup the HTML template
func setupTemplate() error {
	// Check if templates directory exists, create if not
	if _, err := os.Stat("templates"); os.IsNotExist(err) {
		if err := os.Mkdir("templates", 0755); err != nil {
			return fmt.Errorf("failed to create templates directory: %v", err)
		}
	}

	// Path to the template file
	templatePath := filepath.Join("templates", "status.html")

	// Check if the template file exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		log.Printf("Template file not found at %s. Please create it.", templatePath)
		return fmt.Errorf("template file not found: %s", templatePath)
	}

	// Parse the template from the file
	var err error
	tmpl, err = template.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}

	return nil
}

// Check if server is online
func isServerOnline() bool {
	var cmd *exec.Cmd

	// macOS and Linux have slightly different ping commands
	if runtime.GOOS == "darwin" {
		cmd = exec.Command("ping", "-c", "1", "-W", "1000", serverName)
	} else {
		cmd = exec.Command("ping", "-c", "1", "-W", "1", serverName)
	}

	err := cmd.Run()
	return err == nil
}

// Send WOL packet
func sendWakeOnLAN() error {
	log.Printf("Sending WOL packet to %s (%s)", serverName, macAddress)
	cmd := exec.Command("wakeonlan", macAddress)
	return cmd.Run()
}

// Shutdown server with password
func shutdownServer(password string) error {
    log.Printf("Sending shutdown command to %s", serverName)

    // Add more SSH options to handle potential issues
    cmd := exec.Command("sshpass", "-p", password, "ssh",
        "-o", "StrictHostKeyChecking=no",
        "-o", "UserKnownHostsFile=/dev/null",
        "-o", "LogLevel=ERROR",
        fmt.Sprintf("%s@%s", serverUser, serverName),
        "sudo", "-S", "shutdown", "-h", "now")

    // Capture stderr to log any error messages
    var stderr bytes.Buffer
    cmd.Stderr = &stderr

    err := cmd.Run()
    if err != nil {
        log.Printf("SSH Error details: %s", stderr.String())
        return fmt.Errorf("SSH command failed: %v - %s", err, stderr.String())
    }

    return nil
}
