# WOL-Server

A lightweight Wake-on-LAN (WOL) server application designed specifically for Raspberry Pi devices. This tool allows you to remotely power on your network devices using magic packets through a simple, user-friendly interface.

## Features

- **Cross-Platform Compatibility**: Optimized for various Raspberry Pi models (Zero, 1, 2, 3, 4)
- **Easy Installation**: Automated deployment script for quick setup
- **Systemd Integration**: Runs as a system service for reliability
- **Simple Web Interface**: Control your devices through an intuitive web UI
- **Configurable**: Easily customize settings through environment variables or `.env` file
- **Low Resource Usage**: Minimal footprint to run efficiently on any Pi

## Installation

### Method 1: Download Individual Files (Recommended)

1. Download the necessary files directly to your Raspberry Pi:

```bash
# Create a directory for installation
mkdir -p ~/wol-install
cd ~/wol-install

# Download the executable, service file, and deployment script
wget https://github.com/thisloke/wol-server/releases/download/v6/wol-server-arm6
wget https://github.com/thisloke/wol-server/releases/download/v6/wol-server.service
wget https://github.com/thisloke/wol-server/releases/download/v6/deploy.sh

# Make files executable
chmod +x wol-server-arm6
chmod +x deploy.sh
```

2. Create a `.env` file for configuration:
```bash
cat > .env << EOL
# Server Configuration
SERVER_NAME=pippo
SERVER_USER=root
MAC_ADDRESS=aa:aa:aa:aa:aa:aa
PORT=8080
EOL
```

3. Run the deployment script:
```bash
./deploy.sh
```

### Method 2: Build from Source

If you prefer to build the application from source:

```bash
# Install Go (if not already installed)
sudo apt update
sudo apt install golang-go

# Clone the repository
git clone https://github.com/thisloke/wol-server.git
cd wol-server

# Install dependencies
go get github.com/joho/godotenv

# Create a .env file for configuration
cat > .env << EOL
# Server Configuration
SERVER_NAME=pippo
SERVER_USER=root
MAC_ADDRESS=aa:aa:aa:aa:aa:aa
PORT=8080
EOL

# Build the application
go build -o wol-server

# Create installation directory
mkdir -p ~/wol-server

# Copy the binary and config
cp wol-server ~/wol-server/
cp .env ~/wol-server/
chmod +x ~/wol-server/wol-server

# Create and install systemd service
sudo bash -c 'cat > /etc/systemd/system/wol-server.service << EOL
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
EOL'

# Enable and start the service
sudo systemctl daemon-reload
sudo systemctl enable wol-server
sudo systemctl start wol-server
```

## Configuration

WOL-server can be configured using environment variables or a `.env` file in the application directory.

### Available Configuration Options

| Environment Variable | Description | Default Value |
|----------------------|-------------|---------------|
| `SERVER_NAME` | Name of the server to ping/wake | `pippo` |
| `SERVER_USER` | SSH username for remote commands | `root` |
| `MAC_ADDRESS` | MAC address of the target server | `aa:aa:aa:aa:aa:aa` |
| `PORT` | The port number for the web server | `8080` |

### Customizing Your Configuration

You can edit the `.env` file to modify the application's behavior:

```bash
# Navigate to the installation directory
cd ~/wol-server

# Edit the .env file
nano .env
```

After modifying the configuration, restart the service:

```bash
sudo systemctl restart wol-server
```

## Usage

Once installed, the WOL server will be accessible at:
```
http://your-pi-ip:8080
```
(or whatever port you've configured)

### Using the Web Interface

1. **Wake Server**: Click the "Boot" button to send a Wake-on-LAN magic packet to the configured MAC address
2. **Shut Down Server**: Click the "Shutdown" button, confirm, and enter your password if required

## Service Management

Control the WOL server using standard systemd commands:

```bash
# Check service status
sudo systemctl status wol-server

# Stop the service
sudo systemctl stop wol-server

# Start the service
sudo systemctl start wol-server

# Restart the service
sudo systemctl restart wol-server

# View logs
journalctl -u wol-server -f
```

## Troubleshooting

### Checking Your Raspberry Pi Architecture

If you need to verify which version of the application you should use:

```bash
uname -m
```

This will output your Pi's architecture:
- `armv6l`: Use the `wol-server-arm6` binary (Pi Zero, Pi 1)
- `armv7l`: Use the `wol-server-arm7` binary (Pi 2, Pi 3)
- `aarch64`: Use the `wol-server-arm64` binary (64-bit Pi 3, Pi 4)

### Service Not Starting

If the service doesn't start properly, check the logs:

```bash
journalctl -u wol-server -e
```

### Checking Configuration

Verify that your configuration is being properly loaded:

```bash
# View the environment variables being used
sudo systemctl status wol-server
```

Look for a line in the output that shows the loaded configuration values.

### Permission Issues

Make sure the binary has execute permissions:

```bash
chmod +x ~/wol-server/wol-server
```

## Contributing

Contributions are welcome! Feel free to submit pull requests or open issues for bugs and feature requests.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

---

*WOL-Server - Simple Wake-on-LAN management for Raspberry Pi*
