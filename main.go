package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Default values
var (
	serverName      = "server"            // Server to ping
	serverUser      = "root"              // SSH username
	macAddress      = "aa:aa:aa:aa:aa:aa" // MAC address of the server
	port            = "8080"              // Port to listen on
	refreshInterval = 60                  // UI refresh interval in seconds
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

	// Load refresh interval if set
	if envRefresh := os.Getenv("REFRESH_INTERVAL"); envRefresh != "" {
		if val, err := strconv.Atoi(envRefresh); err == nil && val > 0 {
			refreshInterval = val
		}
	}

	log.Printf("Configuration loaded: SERVER_NAME=%s, SERVER_USER=%s, MAC_ADDRESS=%s, PORT=%s, REFRESH=%d",
		serverName, serverUser, macAddress, port, refreshInterval)
}

func main() {
	// Load environment variables
	loadEnvVariables()

	// Setup template
	if err := setupTemplate(); err != nil {
		log.Fatalf("Failed to setup template: %v", err)
	}

	// Verify schedule configuration and clean up stale schedule data if needed
	verifyScheduleConfig()

	// Setup a ticker to check schedule and perform actions
	go runScheduleChecker()

	// Register route handlers
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/boot", bootHandler)
	http.HandleFunc("/confirm-shutdown", confirmShutdownHandler)
	// Password is now taken directly from .env file
	http.HandleFunc("/shutdown", shutdownHandler)

	// Schedule API endpoints
	http.HandleFunc("/api/schedule", scheduleHandler)
	// API shutdown endpoint
	http.HandleFunc("/api/shutdown", apiShutdownHandler)

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

// API Shutdown handler - shuts down the server with password from environment
func apiShutdownHandler(w http.ResponseWriter, r *http.Request) {
	// Add cache control headers to prevent caching
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	// Set content type for JSON response
	w.Header().Set("Content-Type", "application/json")

	// Only allow POST requests
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Method not allowed. Use POST.",
		})
		return
	}

	// Check if shutdown password is available in environment
	if shutdownPassword == "" {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "SHUTDOWN_PASSWORD not set in environment",
		})
		return
	}

	// Check if server is online before attempting shutdown
	if !isServerOnline() {
		// Server is already offline
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Server is already offline",
		})
		return
	}

	// Try to shut down the server using the password from environment
	err := shutdownServer(shutdownPassword)
	if err != nil {
		// Shutdown command failed
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to shutdown server: " + err.Error(),
		})
		log.Printf("API shutdown failed: %v", err)
		return
	}

	// Shutdown initiated successfully
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Server shutdown initiated",
	})
	log.Printf("API shutdown successful")
}

// Handle schedule API requests
func scheduleHandler(w http.ResponseWriter, r *http.Request) {
	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Handle GET request - return current schedule
	if r.Method == "GET" {
		data, err := json.Marshal(GetScheduleConfig())
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "Failed to marshal schedule data: %v"}`, err), http.StatusInternalServerError)
			return
		}
		w.Write(data)
		return
	}

	// Handle POST request - update schedule
	if r.Method == "POST" {
		var newConfig ScheduleConfig
		err := json.NewDecoder(r.Body).Decode(&newConfig)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "Failed to parse request body: %v"}`, err), http.StatusBadRequest)
			return
		}

		// Validate the schedule data
		if newConfig.Enabled {
			// Validate time format (HH:MM)
			_, err = time.Parse("15:04", newConfig.StartTime)
			if err != nil {
				http.Error(w, `{"error": "Invalid start time format. Use 24-hour format (HH:MM)"}`, http.StatusBadRequest)
				return
			}

			_, err = time.Parse("15:04", newConfig.EndTime)
			if err != nil {
				http.Error(w, `{"error": "Invalid end time format. Use 24-hour format (HH:MM)"}`, http.StatusBadRequest)
				return
			}

			// Validate frequency
			validFrequencies := map[string]bool{
				"daily":      true,
				"every2days": true,
				"weekly":     true,
				"monthly":    true,
			}

			if !validFrequencies[newConfig.Frequency] {
				http.Error(w, `{"error": "Invalid frequency. Use 'daily', 'every2days', 'weekly', or 'monthly'"}`, http.StatusBadRequest)
				return
			}

			// Reset lastRun if it wasn't set
			if newConfig.LastRun == "" {
				newConfig.LastRun = ""
			}

			// If auto shutdown is enabled, make sure we have a password in env
			if newConfig.AutoShutdown && shutdownPassword == "" {
				http.Error(w, `{"error": "SHUTDOWN_PASSWORD not set in environment. Please set it before enabling auto-shutdown"}`, http.StatusBadRequest)
				return
			}

			// Check if SSH connection can be established with the password
			if newConfig.AutoShutdown && shutdownPassword != "" {
				log.Printf("Testing SSH connection to %s with provided password", serverName)

				// We'll just check if the server is reachable first
				if !isServerOnline() {
					log.Printf("Server %s is not online, can't test SSH connection", serverName)
				} else {
					// Try to run a harmless command to test SSH connection
					cmd := exec.Command("sshpass", "-p", shutdownPassword, "ssh",
						"-o", "StrictHostKeyChecking=no",
						"-o", "UserKnownHostsFile=/dev/null",
						"-o", "LogLevel=ERROR",
						"-o", "ConnectTimeout=5",
						fmt.Sprintf("%s@%s", serverUser, serverName),
						"echo", "SSH connection test successful")

					var stderr bytes.Buffer
					cmd.Stderr = &stderr

					if err := cmd.Run(); err != nil {
						log.Printf("SSH connection test failed: %v - %s", err, stderr.String())
						// We don't prevent saving the config even if test fails
						// Just log a warning for now
						log.Printf("WARNING: Auto shutdown may not work with the provided password")
					} else {
						log.Printf("SSH connection test successful")
					}
				}
			}
		}

		// Save the new configuration
		err = UpdateScheduleConfig(newConfig)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "Failed to save schedule config: %v"}`, err), http.StatusInternalServerError)
			return
		}

		// Return the updated config
		data, err := json.Marshal(GetScheduleConfig())
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "Failed to marshal schedule data: %v"}`, err), http.StatusInternalServerError)
			return
		}
		w.Write(data)
		return
	}

	// Method not allowed
	http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
}

// Verify and clean up schedule configuration
func verifyScheduleConfig() {
	// If schedule is enabled, validate all required fields
	if scheduleConfig.Enabled {
		log.Println("Verifying schedule configuration...")
		log.Printf("Current config: StartTime=%s, EndTime=%s, Frequency=%s, AutoShutdown=%v",
			scheduleConfig.StartTime, scheduleConfig.EndTime, scheduleConfig.Frequency, scheduleConfig.AutoShutdown)

		// Check for valid time formats
		_, startErr := time.Parse("15:04", scheduleConfig.StartTime)
		_, endErr := time.Parse("15:04", scheduleConfig.EndTime)

		if startErr != nil || endErr != nil || scheduleConfig.StartTime == "" || scheduleConfig.EndTime == "" {
			log.Println("Warning: Invalid time format in schedule configuration, disabling schedule")
			scheduleConfig.Enabled = false
			UpdateScheduleConfig(scheduleConfig)
			return
		}

		// Immediately check if we need to boot ONLY at exact start time
		now := time.Now()
		currentTimeStr := now.Format("15:04")

		// Check ONLY if current time EXACTLY matches start time
		if currentTimeStr == scheduleConfig.StartTime && ShouldRunToday(now) {
			log.Printf("STARTUP MATCH: Current time %s matches start time EXACTLY, attempting boot", currentTimeStr)
			if !isServerOnline() {
				sendWakeOnLAN()
				// Mark that the server was started by the scheduler
				scheduleConfig.StartedBySchedule = true
				scheduleConfig.LastRun = now.Format(time.RFC3339)
				UpdateScheduleConfig(scheduleConfig)
			}
		}

		// Check for valid frequency
		validFrequencies := map[string]bool{
			"daily":      true,
			"every2days": true,
			"weekly":     true,
			"monthly":    true,
		}

		if !validFrequencies[scheduleConfig.Frequency] {
			log.Println("Warning: Invalid frequency in schedule configuration, setting to daily")
			scheduleConfig.Frequency = "daily"
			UpdateScheduleConfig(scheduleConfig)
		}

		log.Printf("Schedule configuration verified: Start=%s, End=%s, Frequency=%s",
			scheduleConfig.StartTime, scheduleConfig.EndTime, scheduleConfig.Frequency)
	}
}

// Run a periodic check of schedule and take appropriate actions
func runScheduleChecker() {
	// Define the checkScheduleOnce function
	checkScheduleOnce := func() {
		// Only check exact times for schedule actions, don't use window logic
		now := time.Now()
		currentTimeStr := now.Format("15:04")
		serverIsOn := isServerOnline()

		// Log schedule status (debug level)
		if scheduleConfig.Enabled {
			log.Printf("Schedule check: Current=%s, Start=%s, End=%s, LastRun=%s",
				currentTimeStr, scheduleConfig.StartTime, scheduleConfig.EndTime, scheduleConfig.LastRun)

			// Only act at exact start or end times
			// EXACT START TIME MATCH - Try to boot server
			if currentTimeStr == scheduleConfig.StartTime && !serverIsOn && ShouldRunToday(now) {
				log.Println("EXACT START TIME: Initiating boot sequence...")

				// Try multiple times to boot with small delays between attempts
				for attempt := 1; attempt <= 3; attempt++ {
					log.Printf("Boot attempt %d/3", attempt)
					err := sendWakeOnLAN()
					if err != nil {
						log.Printf("Error booting server from schedule: %v", err)
					} else {
						log.Println("Schedule: Boot command sent successfully")
						// Mark that server was started by scheduler
						scheduleConfig.StartedBySchedule = true
						scheduleConfig.LastRun = now.Format(time.RFC3339)
						UpdateScheduleConfig(scheduleConfig)
					}

					// Check if server came online
					time.Sleep(3 * time.Second) // Extended wait time for boot check
					if isServerOnline() {
						log.Println("Server successfully booted!")
						break
					}

					// Short delay before next attempt
					if attempt < 3 {
						time.Sleep(1 * time.Second)
					}
				}
				// EXACT END TIME MATCH - Try to shutdown server
			} else if currentTimeStr == scheduleConfig.EndTime && serverIsOn {
				// Check if auto-shutdown is enabled
				if scheduleConfig.AutoShutdown && shutdownPassword != "" && scheduleConfig.StartedBySchedule {
					log.Println("EXACT END TIME: Attempting auto-shutdown")

					// Try multiple times to shut down the server
					var shutdownSuccessful bool
					for attempt := 1; attempt <= 3; attempt++ {
						log.Printf("Auto shutdown attempt %d/3", attempt)
						err := shutdownServer(shutdownPassword)
						if err != nil {
							log.Printf("Auto shutdown attempt %d failed: %v", attempt, err)
							if attempt < 3 {
								time.Sleep(3 * time.Second)
							}
						} else {
							log.Printf("Auto shutdown initiated successfully on attempt %d", attempt)
							shutdownSuccessful = true
							break
						}
					}

					if !shutdownSuccessful {
						log.Printf("All auto shutdown attempts failed")
					}
				}
			} else {
				// No action at non-exact times, just log status
				if serverIsOn && scheduleConfig.StartedBySchedule && currentTimeStr > scheduleConfig.EndTime {
					log.Printf("Server is still online after end time %s - waiting for next exact end time match", scheduleConfig.EndTime)
				}
			}
		}

		// Update last run timestamp if we've passed the end time
		// This helps track when the schedule was last active
		currentConfig := GetScheduleConfig()
		nowTime := time.Now()
		currentTimeString := nowTime.Format("15:04")
		if currentConfig.Enabled && currentTimeString > currentConfig.EndTime && currentConfig.LastRun != "" {
			lastRun, err := time.Parse(time.RFC3339, currentConfig.LastRun)
			if err == nil {
				// If it's been more than a day since the last update, reset the timestamp
				// This allows the schedule to run again based on frequency
				if time.Since(lastRun) > 24*time.Hour {
					log.Println("Schedule: Resetting last run timestamp for next scheduled run")
					currentConfig.LastRun = ""
					UpdateScheduleConfig(currentConfig)
				}
			}
		}
	}

	// Use a slightly shorter interval for more responsive scheduling
	// First check immediately at startup
	checkScheduleOnce()

	// Then set up regular checks
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	log.Println("Schedule checker started - checking every 5 seconds")
	log.Printf("Current schedule: enabled=%v, startTime=%s, endTime=%s, frequency=%s, autoShutdown=%v",
		scheduleConfig.Enabled, scheduleConfig.StartTime, scheduleConfig.EndTime, scheduleConfig.Frequency, scheduleConfig.AutoShutdown)

	for {
		func() {
			// Recover from any panics that might occur in the schedule checker
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Recovered from panic in schedule checker: %v", r)
				}
			}()

			checkScheduleOnce()
		}()

		// Wait for next tick
		<-ticker.C
	}
}
