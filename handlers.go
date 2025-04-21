package main

import (
	"log"
	"net/http"
	"runtime"
	"time"
)

// Handle the root route - show status
func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Add cache control headers to prevent caching
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	online := isServerOnline()
	status := "Online"
	color := "#4caf50" // Material green
	if !online {
		status = "Offline"
		color = "#d32f2f" // Material red
	}

	// Get current schedule configuration
	scheduleConfig := GetScheduleConfig()

	data := StatusData{
		Server:          serverName,
		Status:          status,
		Color:           color,
		IsTestMode:      runtime.GOOS == "darwin",
		AskPassword:     false,
		ErrorMessage:    "",
		Schedule:        scheduleConfig,
		LastUpdated:     time.Now().Format("2006-01-02 15:04:05"),
		RefreshInterval: refreshInterval,
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Template render error: %v", err)
	}
}

// Handle boot request
func bootHandler(w http.ResponseWriter, r *http.Request) {
	if !isServerOnline() {
		// Boot the server using wakeonlan
		err := sendWakeOnLAN()
		if err != nil {
			log.Printf("Error booting server: %v", err)
		}

		// Display booting status
		data := StatusData{
			Server:          serverName,
			Status:          "Booting",
			Color:           "#607d8b", // Material blue-gray
			IsTestMode:      runtime.GOOS == "darwin",
			AskPassword:     false,
			Schedule:        GetScheduleConfig(),
			LastUpdated:     time.Now().Format("2006-01-02 15:04:05"),
			RefreshInterval: refreshInterval,
		}
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
			log.Printf("Template render error: %v", err)
		}
	} else {
		// Server is already online
		data := StatusData{
			Server:          serverName,
			Status:          "Online",
			Color:           "#4caf50", // Material green
			IsTestMode:      runtime.GOOS == "darwin",
			AskPassword:     false,
			Schedule:        GetScheduleConfig(),
			LastUpdated:     time.Now().Format("2006-01-02 15:04:05"),
			RefreshInterval: refreshInterval,
		}
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
			log.Printf("Template render error: %v", err)
		}
	}
}

// Handle shutdown confirmation request
func confirmShutdownHandler(w http.ResponseWriter, r *http.Request) {
	online := isServerOnline()

	if !online {
		// Server is already offline
		data := StatusData{
			Server:          serverName,
			Status:          "Offline",
			Color:           "#d32f2f", // Material red
			IsTestMode:      runtime.GOOS == "darwin",
			AskPassword:     false,
			Schedule:        GetScheduleConfig(),
			LastUpdated:     time.Now().Format("2006-01-02 15:04:05"),
			RefreshInterval: refreshInterval,
		}
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
			log.Printf("Template render error: %v", err)
		}
		return
	}

	// Check if shutdown password is set
	if shutdownPassword == "" {
		// Show error about missing password
		data := StatusData{
			Server:          serverName,
			Status:          "Online",
			Color:           "#4caf50", // Material green
			IsTestMode:      runtime.GOOS == "darwin",
			AskPassword:     false,
			ConfirmShutdown: false,
			ErrorMessage:    "SHUTDOWN_PASSWORD not set in environment. Please set it in the .env file.",
			Schedule:        GetScheduleConfig(),
			LastUpdated:     time.Now().Format("2006-01-02 15:04:05"),
			RefreshInterval: refreshInterval,
		}
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
			log.Printf("Template render error: %v", err)
		}
		return
	}

	// Show confirmation dialog - we'll use the password from .env
	data := StatusData{
		Server:          serverName,
		Status:          "Online",
		Color:           "#4caf50", // Material green
		IsTestMode:      runtime.GOOS == "darwin",
		ConfirmShutdown: true,
		AskPassword:     false, // Make sure we don't ask for password
		Schedule:        GetScheduleConfig(),
		LastUpdated:     time.Now().Format("2006-01-02 15:04:05"),
	}

	// Notify the user if password is not configured
	if shutdownPassword == "" {
		data.ErrorMessage = "SHUTDOWN_PASSWORD not set in environment. Shutdown may fail."
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Template render error: %v", err)
	}
}

// Handle shutdown confirmation without password
// enterPasswordHandler function removed - we now use the password from .env directly

// Handle actual shutdown request
func shutdownHandler(w http.ResponseWriter, r *http.Request) {
	// Only process POST requests for security
	if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Use the password from environment variable
	if shutdownPassword == "" {
		log.Printf("SHUTDOWN_PASSWORD not set in environment, cannot perform shutdown")
		// Show error message
		data := StatusData{
			Server:          serverName,
			Status:          "Online",
			Color:           "#4caf50",
			IsTestMode:      runtime.GOOS == "darwin",
			AskPassword:     false,
			ErrorMessage:    "SHUTDOWN_PASSWORD not set in environment",
			Schedule:        GetScheduleConfig(),
			LastUpdated:     time.Now().Format("2006-01-02 15:04:05"),
			RefreshInterval: refreshInterval,
		}
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
			log.Printf("Template render error: %v", err)
		}
		return
	}

	if isServerOnline() {
		// Shutdown the server using the password from .env file
		err := shutdownServer(shutdownPassword)
		if err != nil {
			log.Printf("Error shutting down server: %v", err)

			// Show error message
			data := StatusData{
				Server:          serverName,
				Status:          "Online",
				Color:           "#4caf50",
				IsTestMode:      runtime.GOOS == "darwin",
				AskPassword:     false, // No longer asking for password
				ErrorMessage:    "Failed to shutdown server. Please check the password in .env file.",
				Schedule:        GetScheduleConfig(),
				LastUpdated:     time.Now().Format("2006-01-02 15:04:05"),
				RefreshInterval: refreshInterval,
			}
			if err := tmpl.Execute(w, data); err != nil {
				http.Error(w, "Failed to render template", http.StatusInternalServerError)
				log.Printf("Template render error: %v", err)
			}
			return
		}

		// Display shutting down status
		data := StatusData{
			Server:          serverName,
			Status:          "Shutting down",
			Color:           "#5d4037", // Material brown
			IsTestMode:      runtime.GOOS == "darwin",
			AskPassword:     false,
			Schedule:        GetScheduleConfig(),
			LastUpdated:     time.Now().Format("2006-01-02 15:04:05"),
			RefreshInterval: refreshInterval,
		}
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
			log.Printf("Template render error: %v", err)
		}
	} else {
		// Server is already offline
		data := StatusData{
			Server:          serverName,
			Status:          "Offline",
			Color:           "#d32f2f", // Material red
			IsTestMode:      runtime.GOOS == "darwin",
			AskPassword:     false,
			Schedule:        GetScheduleConfig(),
			LastUpdated:     time.Now().Format("2006-01-02 15:04:05"),
			RefreshInterval: refreshInterval,
		}
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
			log.Printf("Template render error: %v", err)
		}
	}
}
