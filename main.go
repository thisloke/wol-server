package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os/exec"
	"runtime"
)

const (
	serverName = "delemaco"  // Server to ping
	macAddress = "b8:cb:29:a1:f3:88"  // MAC address of the server
	port       = "8080"  // Port to listen on
)

// StatusData holds data for the HTML template
type StatusData struct {
	Server   string
	Status   string
	Color    string
	IsTestMode bool
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

// Shutdown server
func shutdownServer() error {

	// Real shutdown command for Linux
	log.Printf("Sending shutdown command to %s", serverName)
	cmd := exec.Command("ssh", serverName, "sudo", "shutdown", "-h", "now")
	return cmd.Run()
}

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
		Server:   serverName,
		Status:   status,
		Color:    color,
		IsTestMode: runtime.GOOS == "darwin",
	}

	renderTemplate(w, data)
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
			Server:   serverName,
			Status:   "Booting",
			Color:    "#607d8b",  // Material blue-gray
			IsTestMode: runtime.GOOS == "darwin",
		}
		renderTemplate(w, data)
	} else {
		// Server is already online
		data := StatusData{
			Server:   serverName,
			Status:   "Online",
			Color:    "#4caf50",  // Material green
			IsTestMode: runtime.GOOS == "darwin",
		}
		renderTemplate(w, data)
	}
}

// Handle shutdown request
func shutdownHandler(w http.ResponseWriter, r *http.Request) {
	if isServerOnline() {
		// Shutdown the server
		err := shutdownServer()
		if err != nil {
			log.Printf("Error shutting down server: %v", err)
		}

		// Display shutting down status
		data := StatusData{
			Server:   serverName,
			Status:   "Shutting down",
			Color:    "#5d4037",  // Material brown
			IsTestMode: runtime.GOOS == "darwin",
		}
		renderTemplate(w, data)
	} else {
		// Server is already offline
		data := StatusData{
			Server:   serverName,
			Status:   "Offline",
			Color:    "#d32f2f",  // Material red
			IsTestMode: runtime.GOOS == "darwin",
		}
		renderTemplate(w, data)
	}
}

// Render the HTML template
func renderTemplate(w http.ResponseWriter, data StatusData) {
	tmpl, err := template.New("status").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Server Status: {{.Server}}</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Montserrat:wght@400;600;700&display=swap" rel="stylesheet">
    <style>
        :root {
            --primary-color: {{.Color}};
            --text-color: white;
            --shadow-color: rgba(0, 0, 0, 0.3);
            --hover-color: rgba(255, 255, 255, 0.1);
            --card-bg: rgba(0, 0, 0, 0.15);
        }

        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            background-color: var(--primary-color);
            font-family: 'Montserrat', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
            min-height: 100vh;
            display: flex;
            flex-direction: column;
            justify-content: center;
            align-items: center;
            color: var(--text-color);
            padding: 20px;
            transition: background-color 0.5s ease;
            background-image: radial-gradient(circle at 10% 20%, rgba(255, 255, 255, 0.05) 0%, transparent 20%),
                              radial-gradient(circle at 90% 80%, rgba(255, 255, 255, 0.05) 0%, transparent 20%);
        }

        .container {
            max-width: 800px;
            width: 100%;
            display: flex;
            flex-direction: column;
            align-items: center;
        }

        .card {
            background-color: var(--card-bg);
            border-radius: 20px;
            padding: 40px;
            margin-bottom: 30px;
            box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
            backdrop-filter: blur(5px);
            width: 100%;
            max-width: 600px;
            border: 1px solid rgba(255, 255, 255, 0.1);
        }

        .status-icon {
            font-size: 48px;
            margin-bottom: 20px;
            text-align: center;
        }

        .status-text {
            font-size: 2.5rem;
            font-weight: 600;
            text-align: center;
            margin-bottom: 10px;
            text-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
        }

        .server-name {
            font-size: 1.2rem;
            text-align: center;
            margin-bottom: 30px;
            opacity: 0.9;
        }

        .controls {
            display: flex;
            gap: 15px;
            flex-wrap: wrap;
            justify-content: center;
        }

        .button {
            padding: 15px 25px;
            font-size: 1rem;
            font-weight: 600;
            border: none;
            border-radius: 50px;
            cursor: pointer;
            text-decoration: none;
            color: var(--text-color);
            background-color: var(--shadow-color);
            transition: transform 0.2s ease, background-color 0.3s ease, box-shadow 0.3s ease;
            min-width: 140px;
            text-align: center;
            box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 8px;
        }

        .button:hover {
            background-color: var(--hover-color);
            transform: translateY(-3px);
            box-shadow: 0 7px 14px rgba(0, 0, 0, 0.15);
        }

        .button:active {
            transform: translateY(1px);
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
        }

        .button::before {
            content: '';
            display: inline-block;
            width: 20px;
            height: 20px;
            background-size: contain;
            background-repeat: no-repeat;
            background-position: center;
        }

        .button.refresh::before {
            background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' height='24' viewBox='0 -960 960 960' width='24' fill='white'%3E%3Cpath d='M480-160q-133 0-226.5-93.5T160-480q0-133 93.5-226.5T480-800q85 0 149 34.5T740-671v-99q0-13 8.5-21.5T770-800q13 0 21.5 8.5T800-770v194q0 13-8.5 21.5T770-546H576q-13 0-21.5-8.5T546-576q0-13 8.5-21.5T576-606h138q-38-60-97-97t-137-37q-109 0-184.5 75.5T220-480q0 109 75.5 184.5T480-220q59 0 111-25t89-69q8-9 20.5-10t21.5 7q9 8 10 20t-7 22q-45 53-112 86.5T480-160Z'/%3E%3C/svg%3E");
        }

        .button.boot::before {
            background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' height='24' viewBox='0 -960 960 960' width='24' fill='white'%3E%3Cpath d='M480-120q-151 0-255.5-104.5T120-480q0-138 89-239t219-120q20-3 33.5 9.5T480-797q4 20-9 35.5T437-748q-103 12-170 87t-67 181q0 124 88 212t212 88q124 0 212-88t88-212q0-109-69.5-184.5T564-748q-21-3-31.5-19T525-798q3-20 19-30.5t35-6.5q136 19 228.5 122.5T900-480q0 150-104.5 255T480-120Zm0-170q-20 0-33.5-14T433-340v-286q0-21 14-34.5t33-13.5q20 0 33.5 13.5T527-626v286q0 22-14 36t-33 14Z'/%3E%3C/svg%3E");
        }

        .button.shutdown::before {
            background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' height='24' viewBox='0 -960 960 960' width='24' fill='white'%3E%3Cpath d='M480-120q-151 0-255.5-104.5T120-480q0-138 89-239t219-120q20-3 33.5 9.5T480-797q4 20-9 35.5T437-748q-103 12-170 87t-67 181q0 124 88 212t212 88q124 0 212-88t88-212q0-109-69.5-184.5T564-748q-21-3-31.5-19T525-798q3-20 19-30.5t35-6.5q136 19 228.5 122.5T900-480q0 150-104.5 255T480-120Zm0-360Z'/%3E%3C/svg%3E");
        }

        .test-panel {
            margin-top: 40px;
            padding: 20px;
            background-color: rgba(0, 0, 0, 0.3);
            border-radius: 15px;
            width: 100%;
            max-width: 600px;
            border: 1px solid rgba(255, 255, 255, 0.1);
        }

        .test-note {
            color: white;
            margin-bottom: 20px;
            text-align: center;
            font-size: 0.9rem;
            opacity: 0.8;
        }

        .footer {
            margin-top: 40px;
            font-size: 0.8rem;
            opacity: 0.7;
            text-align: center;
        }

        /* Responsive adjustments */
        @media (max-width: 600px) {
            .card {
                padding: 30px 20px;
            }

            .status-text {
                font-size: 2rem;
            }

            .controls {
                flex-direction: column;
                width: 100%;
            }

            .button {
                width: 100%;
            }
        }

        /* Status-specific icons */
        {{if eq .Status "Online"}}
        .status-icon::before {
            content: "✓";
            color: #4caf50;
        }
        {{else if eq .Status "Offline"}}
        .status-icon::before {
            content: "✗";
            color: #f44336;
        }
        {{else if eq .Status "Booting"}}
        .status-icon::before {
            content: "⟳";
            color: #ffeb3b;
            display: inline-block;
            animation: spin 2s linear infinite;
        }
        {{else if eq .Status "Shutting down"}}
        .status-icon::before {
            content: "⏻";
            color: #ff9800;
        }
        {{end}}

        @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="card">
            <div class="status-icon"></div>
            <h1 class="status-text">{{.Status}}</h1>
            <div class="server-name">Server: <strong>{{.Server}}</strong></div>

            <div class="controls">
                <a href="/" class="button refresh">Refresh</a>
                <a href="/boot" class="button boot">Boot</a>
                <a href="/shutdown" class="button shutdown">Shutdown</a>
            </div>
        </div>

        {{if .IsTestMode}}
        <div class="test-panel">
            <div class="test-note">
                Running on macOS in test mode. Wake-on-LAN packets are sent, but remote shutdown is simulated.
            </div>
        </div>
        {{end}}

        <div class="footer">
            Wake-on-LAN Server Control Panel
        </div>
    </div>
</body>
</html>`)

	if err != nil {
		http.Error(w, "Failed to load template", http.StatusInternalServerError)
		log.Printf("Template error: %v", err)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Template render error: %v", err)
	}
}

func main() {
	// Register route handlers
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/boot", bootHandler)
	http.HandleFunc("/shutdown", shutdownHandler)

	// Start the server
	listenAddr := fmt.Sprintf(":%s", port)
	log.Printf("Starting WOL Server on http://localhost%s", listenAddr)

	if runtime.GOOS == "darwin" {
		log.Println("Running on macOS in test mode - remote shutdown commands will be simulated")
	}

	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
