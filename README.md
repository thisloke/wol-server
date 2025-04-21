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

### Method 1: Pre-built Release (Recommended)

1. **Check which binary is right for your Raspberry Pi model**:
   ```bash
   uname -m
   ```
   - `armv6l`: Use `wol-server-arm6` (Pi Zero, Pi 1)
   - `armv7l`: Use `wol-server-arm7` (Pi 2, Pi 3 32-bit)
   - `aarch64`: Use `wol-server-arm64` (Pi 3/4 64-bit)

2. **Download the latest release**:

   Navigate to [Releases](https://github.com/thisloke/wol-server/releases) and download the appropriate files for your Pi, or use the following commands:

   ```bash
   # Create a directory for installation
   mkdir -p ~/wol-install
   cd ~/wol-install

   # Download the executable (replace v6 with the latest version number)
   wget https://github.com/thisloke/wol-server/releases/download/v6/wol-server-arm6

   # Download the service file
   wget https://github.com/thisloke/wol-server/releases/download/v6/wol-server.service

   # Download the deploy script
   wget https://github.com/thisloke/wol-server/releases/download/v6/deploy.sh

   # Make files executable
   chmod +x wol-server-arm6
   chmod +x deploy.sh
   ```

3. **Create a `.env` file for configuration**:
   ```bash
   cat > .env << EOL
   # Server Configuration
   SERVER_NAME=pippo
   SERVER_USER=root
   MAC_ADDRESS=aa:aa:aa:aa:aa:aa
   PORT=8080
   EOL
   ```

4. **Run the deployment script**:
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

### Method 3: Manual Installation

If you encounter issues with the deployment script:

```bash
# Create the application directory
mkdir -p ~/wol-server

# Copy the binary and rename it
cp wol-server-arm6 ~/wol-server/wol-server
chmod +x ~/wol-server/wol-server

# Copy the .env file
cp .env ~/wol-server/

# Install the service file
sudo cp wol-server.service /etc/systemd/system/
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

## Updating to a New Version

To update to a new version:

```bash
# Navigate to a temporary directory
mkdir -p ~/wol-update
cd ~/wol-update

# Download the latest release (replace v7 with the latest version number)
wget https://github.com/thisloke/wol-server/releases/download/v7/wol-server-arm6
chmod +x wol-server-arm6

# Stop the service
sudo systemctl stop wol-server

# Replace the binary
cp wol-server-arm6 ~/wol-server/wol-server

# Start the service
sudo systemctl start wol-server

# Clean up
cd ~
rm -rf ~/wol-update
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

### Can't Download Release Files

If you're having trouble downloading the release files directly, you can also:

1. Download the files on your computer
2. Transfer them to your Raspberry Pi using SCP, SFTP, or a USB drive
3. Continue with the installation steps as described above

## Contributing

Contributions are welcome! Feel free to submit pull requests or open issues for bugs and feature requests.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
