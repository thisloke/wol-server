package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

// ScheduleConfig holds the backup window schedule settings
type ScheduleConfig struct {
	Enabled           bool   `json:"enabled"`
	StartTime         string `json:"startTime"`         // Format: "HH:MM" (24-hour)
	EndTime           string `json:"endTime"`           // Format: "HH:MM" (24-hour)
	Frequency         string `json:"frequency"`         // "daily", "every2days", "weekly", "monthly"
	LastRun           string `json:"lastRun"`           // ISO8601 format - when the schedule last ran
	AutoShutdown      bool   `json:"autoShutdown"`      // Whether to automatically shut down at end time
	StartedBySchedule bool   `json:"startedBySchedule"` // Whether server was started by scheduler
}

// StatusData holds data for the HTML template
type StatusData struct {
	Server          string
	Status          string
	Color           string
	IsTestMode      bool
	ConfirmShutdown bool
	AskPassword     bool
	ErrorMessage    string
	Schedule        ScheduleConfig
	LastUpdated     string
	RefreshInterval int
}

var tmpl *template.Template
var scheduleConfig ScheduleConfig
var scheduleConfigPath = "schedule.json"
var shutdownPassword string // Will be loaded from environment

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

	// Load schedule config
	err = loadScheduleConfig()
	if err != nil {
		log.Printf("Warning: Failed to load schedule config: %v", err)
		// Continue with default (empty) schedule config
	}

	// Check for required system tools
	checkRequiredTools()

	// Load shutdown password from environment
	shutdownPassword = os.Getenv("SHUTDOWN_PASSWORD")
	if shutdownPassword == "" {
		log.Println("SHUTDOWN_PASSWORD not set in environment. Automatic shutdown will be disabled.")
	} else {
		log.Println("SHUTDOWN_PASSWORD loaded from environment")
	}

	return nil
}

// Check if required system tools are available
func checkRequiredTools() {
	// Check for wakeonlan command
	if _, err := exec.LookPath("wakeonlan"); err != nil {
		if _, err := exec.LookPath("etherwake"); err != nil {
			if _, err := exec.LookPath("wol"); err != nil {
				log.Printf("WARNING: No Wake-on-LAN tools found. Please install wakeonlan, etherwake, or wol package.")
				log.Printf("Installation instructions:")
				log.Printf("  - For Debian/Ubuntu: sudo apt-get install wakeonlan")
				log.Printf("  - For macOS: brew install wakeonlan")
				log.Printf("  - For Windows: Download from https://www.depicus.com/wake-on-lan/wake-on-lan-cmd")
			} else {
				log.Printf("Using 'wol' for Wake-on-LAN functionality")
			}
		} else {
			log.Printf("Using 'etherwake' for Wake-on-LAN functionality")
		}
	} else {
		log.Printf("Found 'wakeonlan' command for Wake-on-LAN functionality")
	}

	// Check for ping command (needed for server status checks)
	if _, err := exec.LookPath("ping"); err != nil {
		log.Printf("WARNING: 'ping' command not found. Server status checks may fail.")
	}
}

// Load schedule configuration from file
func loadScheduleConfig() error {
	// Check if config file exists
	if _, err := os.Stat(scheduleConfigPath); os.IsNotExist(err) {
		// Create default config
		scheduleConfig = ScheduleConfig{
			Enabled:           false,
			StartTime:         "",
			EndTime:           "",
			Frequency:         "daily",
			LastRun:           "",
			AutoShutdown:      false,
			StartedBySchedule: false,
		}
		// Save default config
		return saveScheduleConfig()
	}

	// Read the file
	data, err := os.ReadFile(scheduleConfigPath)
	if err != nil {
		return fmt.Errorf("failed to read schedule config file: %v", err)
	}

	// Unmarshal JSON data
	err = json.Unmarshal(data, &scheduleConfig)
	if err != nil {
		return fmt.Errorf("failed to parse schedule config: %v", err)
	}

	// Log loaded configuration for debugging
	log.Printf("Loaded schedule config: Enabled=%v, StartTime=%s, EndTime=%s, Frequency=%s",
		scheduleConfig.Enabled, scheduleConfig.StartTime, scheduleConfig.EndTime, scheduleConfig.Frequency)

	return nil
}

// Save schedule configuration to file
func saveScheduleConfig() error {
	data, err := json.MarshalIndent(scheduleConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal schedule config: %v", err)
	}

	err = os.WriteFile(scheduleConfigPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to save schedule config: %v", err)
	}

	return nil
}

// GetScheduleConfig returns the current schedule config
func GetScheduleConfig() ScheduleConfig {
	return scheduleConfig
}

// UpdateScheduleConfig updates the schedule configuration
func UpdateScheduleConfig(newConfig ScheduleConfig) error {
	scheduleConfig = newConfig
	return saveScheduleConfig()
}

// CheckSchedule checks if server should be on/off based on schedule
func CheckSchedule() (shouldBeOn bool) {
	// If schedule is not enabled, do nothing
	if !scheduleConfig.Enabled {
		return false
	}

	// If start time or end time is empty, the schedule is not properly configured
	if scheduleConfig.StartTime == "" || scheduleConfig.EndTime == "" {
		log.Printf("Schedule configuration incomplete: StartTime=%s, EndTime=%s",
			scheduleConfig.StartTime, scheduleConfig.EndTime)
		return false
	}

	now := time.Now()
	today := now.Format("2006-01-02")

	// Get current time as just hours and minutes for direct string comparison first
	currentTimeStr := now.Format("15:04")

	// Log the exact time comparison we're doing
	log.Printf("Schedule debug: Current=%s, Start=%s, End=%s, LastRun=%s",
		currentTimeStr, scheduleConfig.StartTime, scheduleConfig.EndTime, scheduleConfig.LastRun)

	// Parse start time with proper error handling
	startTime, err := time.Parse("2006-01-02 15:04", fmt.Sprintf("%s %s", today, scheduleConfig.StartTime))
	if err != nil {
		log.Printf("Error parsing start time '%s': %v", scheduleConfig.StartTime, err)
		return false
	}

	// Parse end time with proper error handling
	endTime, err := time.Parse("2006-01-02 15:04", fmt.Sprintf("%s %s", today, scheduleConfig.EndTime))
	if err != nil {
		log.Printf("Error parsing end time '%s': %v", scheduleConfig.EndTime, err)
		return false
	}

	// If end time is before start time, it means the window spans to the next day
	if endTime.Before(startTime) {
		endTime = endTime.AddDate(0, 0, 1)

		// Special case: if we're after midnight but before the end time
		// we need to adjust the start time to be from yesterday
		if now.Before(endTime) && now.After(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())) {
			startTime = startTime.AddDate(0, 0, -1)
		}
	}

	// Check if the schedule should run today based on frequency
	// Check if we're in the schedule window
	if !ShouldRunToday(now) {
		log.Printf("Schedule is active but not set to run today based on frequency: %s", scheduleConfig.Frequency)
		return false
	}

	// Check for auto shutdown at end time
	if currentTimeStr == scheduleConfig.EndTime {
		if scheduleConfig.AutoShutdown && shutdownPassword != "" && isServerOnline() {
			// Only shut down if the server was started by the scheduler
			if scheduleConfig.StartedBySchedule {
				log.Printf("Auto shutdown triggered at schedule end time %s", scheduleConfig.EndTime)

				// Try up to 3 times to shut down the server
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
			} else {
				log.Printf("Server was not started by scheduler, skipping auto shutdown")
			}
		}
	}

	// Check if current time is within the schedule window
	// Check if we're between start and end times
	if currentTimeStr == scheduleConfig.StartTime {
		log.Printf("Schedule match: Current time exactly matches start time")
		shouldBeOn = true
	} else if currentTimeStr == scheduleConfig.EndTime {
		log.Printf("Schedule end: Current time exactly matches end time")
		shouldBeOn = false

		// Check if auto shutdown is enabled
		if scheduleConfig.AutoShutdown && shutdownPassword != "" && isServerOnline() {
			// Only shut down if the server was started by the scheduler
			if scheduleConfig.StartedBySchedule {
				log.Printf("Auto shutdown is enabled - attempting to shut down server at end time")

				// Try up to 3 times to shut down the server
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
			} else {
				log.Printf("Server was not started by scheduler, skipping auto shutdown")
			}
		}
	} else {
		// ONLY consider the server should be on at EXACT start time or EXACT end time
		shouldBeOn = (currentTimeStr == scheduleConfig.StartTime)

		// Log that we're waiting for exact times for actions
		if currentTimeStr != scheduleConfig.StartTime && currentTimeStr != scheduleConfig.EndTime {
			log.Printf("Not at exact schedule times - no action needed until exact start/end time")
		}
	}

	log.Printf("Schedule window check: Current=%v, Start=%v, End=%v, ShouldBeOn=%v",
		now.Format("15:04:05"), startTime.Format("15:04:05"), endTime.Format("15:04:05"), shouldBeOn)

	// Explicitly check end time for better debugging
	if scheduleConfig.EndTime != "" && currentTimeStr == scheduleConfig.EndTime {
		log.Printf("EXACT END TIME MATCH! Current time %s equals end time - schedule window should close", currentTimeStr)
	}

	// If we're at exact start time, update the LastRun timestamp
	if currentTimeStr == scheduleConfig.StartTime && ShouldRunToday(now) {
		// We only track that we've seen the start time
		log.Printf("Exact start time reached - marking schedule run")

		// Don't automatically boot the server here - let the main scheduler handle it
		// We're just updating state information
		scheduleConfig.LastRun = now.Format(time.RFC3339)
		if err := saveScheduleConfig(); err != nil {
			log.Printf("Warning: Failed to save schedule config: %v", err)
		}
	}

	return shouldBeOn
}

// ShouldRunToday checks if the schedule should run today based on frequency
func ShouldRunToday(now time.Time) bool {
	// We no longer check for windows - we only check at exact times
	currentTimeStr := now.Format("15:04")
	today := now.Format("2006-01-02")

	startTime, startErr := time.Parse("2006-01-02 15:04", fmt.Sprintf("%s %s", today, scheduleConfig.StartTime))
	endTime, endErr := time.Parse("2006-01-02 15:04", fmt.Sprintf("%s %s", today, scheduleConfig.EndTime))

	if startErr == nil && endErr == nil {
		// If end time is before start time, it means the window spans to the next day
		if endTime.Before(startTime) {
			endTime = endTime.AddDate(0, 0, 1)
		}

		// Only log that we're at an exact schedule time
		if currentTimeStr == scheduleConfig.StartTime {
			log.Println("Currently at exact start time - schedule should be active")
			return true
		}
	}

	// If no previous run, allow it to run
	if scheduleConfig.LastRun == "" {
		log.Println("No previous run recorded, schedule can run today")
		return true
	}

	lastRun, err := time.Parse(time.RFC3339, scheduleConfig.LastRun)
	if err != nil {
		log.Printf("Error parsing last run date '%s': %v", scheduleConfig.LastRun, err)
		// If we can't parse the date, better to let it run than to block it
		return true
	}

	// Don't allow running multiple times on the same day unless
	// it's been reset explicitly (LastRun set to empty)
	if lastRun.Year() == now.Year() && lastRun.YearDay() == now.YearDay() {
		// Check if we've passed the end time today - if so, we can reset for next run
		if scheduleConfig.EndTime != "" && currentTimeStr > scheduleConfig.EndTime {
			log.Println("Current time is after end time - resetting for next run")
			scheduleConfig.LastRun = ""
			scheduleConfig.StartedBySchedule = false // Reset this flag too
			saveScheduleConfig()
			return false
		}

		log.Println("Schedule already ran today, skipping")
		return false
	}

	switch scheduleConfig.Frequency {
	case "daily":
		// Run every day
		log.Println("Daily schedule: allowed to run today")
		return true
	case "every2days":
		// Check if at least 2 days have passed
		elapsed := now.Sub(lastRun)
		eligible := elapsed >= 48*time.Hour
		log.Printf("Every 2 days schedule: %v hours elapsed, eligible=%v", elapsed.Hours(), eligible)
		return eligible
	case "weekly":
		// Check if at least 7 days have passed
		elapsed := now.Sub(lastRun)
		eligible := elapsed >= 7*24*time.Hour
		log.Printf("Weekly schedule: %v days elapsed, eligible=%v", elapsed.Hours()/24, eligible)
		return eligible
	case "monthly":
		// Check if last run was in a different month
		sameMonth := lastRun.Month() == now.Month() && lastRun.Year() == now.Year()
		log.Printf("Monthly schedule: eligible=%v", !sameMonth)
		return !sameMonth
	default:
		log.Printf("Unknown frequency '%s', defaulting to daily", scheduleConfig.Frequency)
		return true
	}
}

// Check if server is online
func isServerOnline() bool {
	var cmd *exec.Cmd
	var stdout, stderr bytes.Buffer

	// macOS and Linux have slightly different ping commands
	if runtime.GOOS == "darwin" {
		cmd = exec.Command("ping", "-c", "1", "-W", "1000", serverName)
	} else {
		cmd = exec.Command("ping", "-c", "1", "-W", "1", serverName)
	}

	// Capture output for debugging
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	log.Printf("Checking if server %s is online...", serverName)
	err := cmd.Run()

	if err != nil {
		// Only log the full error in debug mode to avoid spamming the logs
		if stderr.String() != "" {
			log.Printf("Server %s is offline: %v - %s", serverName, err, stderr.String())
		} else {
			log.Printf("Server %s is offline", serverName)
		}
		return false
	}

	log.Printf("Server %s is online", serverName)
	return true
}

// Send WOL packet
func sendWakeOnLAN() error {
	log.Printf("Sending WOL packet to %s (%s)", serverName, macAddress)

	// Check if wakeonlan command exists
	if _, err := exec.LookPath("wakeonlan"); err == nil {
		// Create the command
		cmd := exec.Command("wakeonlan", macAddress)

		// Capture both stdout and stderr
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		// Execute the command
		err := cmd.Run()

		// Log the result
		if err != nil {
			log.Printf("WOL command failed: %v - stderr: %s", err, stderr.String())
			return fmt.Errorf("WOL command failed: %v - %s", err, stderr.String())
		}

		output := stdout.String()
		if output != "" {
			log.Printf("WOL command output: %s", output)
		}

		log.Printf("WOL packet sent successfully to %s", macAddress)
		return nil
	} else {
		// wakeonlan command not found, try etherwake
		if _, err := exec.LookPath("etherwake"); err == nil {
			log.Printf("Using etherwake as wakeonlan alternative")
			cmd := exec.Command("etherwake", macAddress)

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			if err != nil {
				log.Printf("etherwake command failed: %v - stderr: %s", err, stderr.String())
				return fmt.Errorf("etherwake command failed: %v - %s", err, stderr.String())
			}

			log.Printf("WOL packet sent successfully via etherwake to %s", macAddress)
			return nil
		} else {
			// Try wol command as last resort
			if _, err := exec.LookPath("wol"); err == nil {
				log.Printf("Using wol as wakeonlan alternative")
				cmd := exec.Command("wol", macAddress)

				var stdout, stderr bytes.Buffer
				cmd.Stdout = &stdout
				cmd.Stderr = &stderr

				err := cmd.Run()
				if err != nil {
					log.Printf("wol command failed: %v - stderr: %s", err, stderr.String())
					return fmt.Errorf("wol command failed: %v - %s", err, stderr.String())
				}

				log.Printf("WOL packet sent successfully via wol to %s", macAddress)
				return nil
			} else {
				// Implement a fallback pure Go WOL solution
				log.Printf("No WOL tools found. Please install wakeonlan, etherwake, or wol package.")
				log.Printf("Installation instructions:")
				log.Printf("  - For Debian/Ubuntu: sudo apt-get install wakeonlan")
				log.Printf("  - For macOS: brew install wakeonlan")
				log.Printf("  - For Windows: Download from https://www.depicus.com/wake-on-lan/wake-on-lan-cmd")

				return fmt.Errorf("wakeonlan command not found in PATH. Please install wakeonlan tool")
			}
		}
	}
}

// Shutdown server with password
func shutdownServer(password string) error {
	log.Printf("Sending shutdown command to %s", serverName)

	var err error
	var stderr bytes.Buffer

	// First try using sshpass with password
	if _, err := exec.LookPath("sshpass"); err == nil {
		log.Println("Using sshpass for authentication")
		log.Printf("Password being used: %s", password)

		// Add more SSH options to handle potential issues
		cmd := exec.Command("sshpass", "-p", password, "ssh",
			"-o", "StrictHostKeyChecking=no",
			"-o", "UserKnownHostsFile=/dev/null",
			"-o", "LogLevel=ERROR",
			"-o", "ConnectTimeout=10",
			fmt.Sprintf("%s@%s", serverUser, serverName),
			"sudo", "-S", "shutdown", "-h", "now")

		// Capture stderr to log any error messages
		stderr.Reset()
		cmd.Stderr = &stderr
		cmd.Stdin = bytes.NewBufferString(password + "\n")

		err = cmd.Run()
		if err == nil {
			log.Println("SSH command executed successfully using sshpass")
			return nil
		}

		log.Printf("sshpass method failed: %v - %s", err, stderr.String())
	}

	// Try direct SSH with password via stdin
	log.Println("Trying direct SSH with password via stdin")
	log.Printf("Password being used: %s", password)

	cmd := exec.Command("ssh",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "LogLevel=ERROR",
		"-o", "ConnectTimeout=10",
		fmt.Sprintf("%s@%s", serverUser, serverName),
		"sudo", "-S", "shutdown", "-h", "now")

	stderr.Reset()
	cmd.Stderr = &stderr
	cmd.Stdin = bytes.NewBufferString(password + "\n")

	err = cmd.Run()
	if err == nil {
		log.Println("SSH command executed successfully using direct SSH")
		return nil
	}

	log.Printf("SSH Error details: %s", stderr.String())

	// Try a simpler shutdown command as a fallback
	log.Println("Trying simpler shutdown command as fallback")
	log.Printf("Password being used: %s", password)

	cmd = exec.Command("ssh",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "LogLevel=ERROR",
		fmt.Sprintf("%s@%s", serverUser, serverName),
		"sudo", "shutdown", "now")

	stderr.Reset()
	cmd.Stderr = &stderr
	cmd.Stdin = bytes.NewBufferString(password + "\n")

	err = cmd.Run()
	if err != nil {
		log.Printf("All shutdown attempts failed: %v - %s", err, stderr.String())
		return fmt.Errorf("SSH command failed: %v - %s", err, stderr.String())
	}

	return nil
}
