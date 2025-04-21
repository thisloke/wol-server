![Server Status Screenshot](https://placeholder-for-screenshot.png)

## Features

- **Status Monitoring**: Real-time status checks to determine if your server is online or offline
- **Wake-on-LAN**: Boot your server remotely with the click of a button
- **Remote Shutdown**: Safely shut down your server when it's not needed
- **Responsive UI**: Simple, mobile-friendly interface with color-coded status indicators
- **Lightweight**: Built with Go for minimal resource usage, perfect for Raspberry Pi Zero

## Requirements

- Raspberry Pi (Zero, 2, 3, 4, etc.)
- Go (version 1.16 or higher)
- wakeonlan utility
- SSH access to the target server (for shutdown functionality)
- Target server configured for Wake-on-LAN

## Installation

### 1. Install Dependencies

```bash
sudo apt update
sudo apt install golang-go wakeonlan
```

### 2. Clone the Repository

```bash
git clone https://github.com/yourusername/wol-server.git
cd wol-server
```

### 3. Configure the Application

Edit the constants in `main.go` to match your server:

```go
const (
    serverName = "yourserver"     // Hostname or IP address of your server
    macAddress = "xx:xx:xx:xx:xx:xx"  // MAC address of your server's network interface
    port       = "8080"           // Port to run the web application on
)
```

### 4. Build the Application

```bash
go build -o wol-server
```

### 5. Set Up as a System Service

Create a systemd service file:

```bash
sudo nano /etc/systemd/system/wol-server.service
```

Add the following content (adjust paths if needed):

```
[Unit]
Description=WOL Server Go Application
After=network.target

[Service]
User=pi
WorkingDirectory=/home/pi/wol-server
ExecStart=/home/pi/wol-server/wol-server
Restart=always

[Install]
WantedBy=multi-user.target
```

Enable and start the service:

```bash
sudo systemctl enable wol-server
sudo systemctl start wol-server
```

## SSH Configuration for Remote Shutdown

For the shutdown functionality to work, you need to set up password-less SSH:

1. Generate an SSH key on your Raspberry Pi:
   ```bash
   ssh-keygen -t rsa
   ```

2. Copy the key to your target server:
   ```bash
   ssh-copy-id user@yourserver
   ```

3. Configure sudo on the target server to allow password-less shutdown:
   ```bash
   # On the target server, run:
   sudo visudo

   # Add this line (replacing 'user' with your username):
   user ALL=(ALL) NOPASSWD: /sbin/shutdown
   ```

## Usage

Access the web interface by navigating to `http://raspberry-pi-ip:8080` in your browser.

The interface provides:

- Current server status (Online/Offline)
- Boot button - sends a Wake-on-LAN magic packet to your server
- Shutdown button - safely shuts down your server via SSH
- Refresh button - manually updates the status display

## Troubleshooting

- **Server won't boot**:
  - Verify that Wake-on-LAN is enabled in your server's BIOS/UEFI
  - Confirm the MAC address is correct
  - Check if your router blocks Wake-on-LAN packets

- **Shutdown doesn't work**:
  - Verify SSH key setup
  - Check sudo configuration on target server
  - Test manual SSH command: `ssh user@server sudo shutdown -h now`

- **Web interface not accessible**:
  - Ensure the service is running: `sudo systemctl status wol-server`
  - Check for firewall rules blocking port 8080
  - Verify the Raspberry Pi is connected to the network

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Inspired by the need for a simple, lightweight server power management tool
- Thanks to the Go community for the excellent standard library that makes web development straightforward
