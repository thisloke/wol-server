package main

import (
	"log"
	"net/http"
	"runtime"
)

// Handle the root route - show status
func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	online := isServerOnline()
	status := "Online"
	color := "#4caf50"  // Material green
	if !online {
		status = "Offline"
		color = "#d32f2f"  // Material red
	}

	data := StatusData{
		Server:       serverName,
		Status:       status,
		Color:        color,
		IsTestMode:   runtime.GOOS == "darwin",
		AskPassword:  false,
		ErrorMessage: "",
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
			Server:      serverName,
			Status:      "Booting",
			Color:       "#607d8b",  // Material blue-gray
			IsTestMode:  runtime.GOOS == "darwin",
			AskPassword: false,
		}
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
			log.Printf("Template render error: %v", err)
		}
	} else {
		// Server is already online
		data := StatusData{
			Server:      serverName,
			Status:      "Online",
			Color:       "#4caf50",  // Material green
			IsTestMode:  runtime.GOOS == "darwin",
			AskPassword: false,
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
			Server:      serverName,
			Status:      "Offline",
			Color:       "#d32f2f",  // Material red
			IsTestMode:  runtime.GOOS == "darwin",
			AskPassword: false,
		}
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
			log.Printf("Template render error: %v", err)
		}
		return
	}

	// Show confirmation dialog
	data := StatusData{
		Server:          serverName,
		Status:          "Online",
		Color:           "#4caf50",  // Material green
		IsTestMode:      runtime.GOOS == "darwin",
		ConfirmShutdown: true,
		AskPassword:     false,
	}
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Template render error: %v", err)
	}
}

// Handle password entry for shutdown
func enterPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if !isServerOnline() {
		// Server is already offline, redirect to home
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Show password entry dialog
	data := StatusData{
		Server:      serverName,
		Status:      "Online",
		Color:       "#4caf50",  // Material green
		IsTestMode:  runtime.GOOS == "darwin",
		AskPassword: true,
	}
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Template render error: %v", err)
	}
}

// Handle actual shutdown request
func shutdownHandler(w http.ResponseWriter, r *http.Request) {
	// Only process POST requests for security
	if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Parse form data to get password
	if err := r.ParseForm(); err != nil {
		log.Printf("Error parsing form: %v", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Get password from form
	password := r.FormValue("password")

	if password == "" {
		// Show password form again with error
		data := StatusData{
			Server:       serverName,
			Status:       "Online",
			Color:        "#4caf50",
			IsTestMode:   runtime.GOOS == "darwin",
			AskPassword:  true,
			ErrorMessage: "Password cannot be empty",
		}
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
			log.Printf("Template render error: %v", err)
		}
		return
	}

	if isServerOnline() {
		// Shutdown the server
		err := shutdownServer(password)
		if err != nil {
			log.Printf("Error shutting down server: %v", err)

			// Show password form again with error
			data := StatusData{
				Server:       serverName,
				Status:       "Online",
				Color:        "#4caf50",
				IsTestMode:   runtime.GOOS == "darwin",
				AskPassword:  true,
				ErrorMessage: "Failed to shutdown server. Please check your password.",
			}
			if err := tmpl.Execute(w, data); err != nil {
				http.Error(w, "Failed to render template", http.StatusInternalServerError)
				log.Printf("Template render error: %v", err)
			}
			return
		}

		// Display shutting down status
		data := StatusData{
			Server:      serverName,
			Status:      "Shutting down",
			Color:       "#5d4037",  // Material brown
			IsTestMode:  runtime.GOOS == "darwin",
			AskPassword: false,
		}
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
			log.Printf("Template render error: %v", err)
		}
	} else {
		// Server is already offline
		data := StatusData{
			Server:      serverName,
			Status:      "Offline",
			Color:       "#d32f2f",  // Material red
			IsTestMode:  runtime.GOOS == "darwin",
			AskPassword: false,
		}
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
			log.Printf("Template render error: %v", err)
		}
	}
}
