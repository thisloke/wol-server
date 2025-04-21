# WOL Server - Wake-on-LAN Control Panel for Raspberry Pi

A lightweight web-based Wake-on-LAN control panel designed for Raspberry Pi that lets you remotely power on and shut down your network devices.

![WOL Server Screenshot](https://i.imgur.com/example.jpg)

## Features

- **Simple Web Interface**: Boot and shut down your server with a clean, responsive UI
- **Status Monitoring**: Check if your target device is online
- **Raspberry Pi Optimized**: Built specifically for ARM processors found in all Raspberry Pi models
- **Secure Shutdown**: Password-protected shutdown functionality
- **Lightweight**: Minimal resource usage ideal for running on even the oldest Pi models
- **Easy Setup**: Simple installation process with clear instructions

## Installation

### Prerequisites

- Raspberry Pi (any model) running Raspberry Pi OS
- Network connection
- Basic knowledge of SSH/terminal

### Option 1: One-Command Installation

1. **Download the latest release** on your local machine from the [Releases page](https://github.com/thisloke/wol-server/releases)

2. **Transfer the package to your Raspberry Pi** using SCP:
   ```bash
   scp wol-server.tar.gz pi@your-pi-ip:~/
   ```

3. **SSH into your Raspberry Pi**:
   ```bash
   ssh pi@your-pi-ip
   ```

4. **Install with a single command**:
   ```bash
   tar -xzf wol-server.tar.gz && ./install.sh
   ```

5. **Access the web interface** at:
   ```
   http://your-pi-ip:8080
   ```

### Option 2: Manual Installation

If you prefer a manual approach or encounter issues with the automated install:

1. **Create installation directory**:
   ```bash
   mkdir -p ~/wol-server/templates
   ```

2. **Transfer and install program files**:
   ```bash
   # Copy the executable
   cp wol-server-arm6 ~/wol-server/wol-server
   chmod +x ~/wol-server/wol-server

   # Copy template files
   cp templates/* ~/wol-server/templates/

   # Create .env file
   cat > ~/wol-server/.env << EOL
   SERVER_NAME=pippo
   SERVER_USER=root
   MAC_ADDRESS=aa:bb:cc:dd:ee:ff
   PORT=8080
   EOL
   ```

3. **Install service**:
   ```bash
   sudo cp wol-server.service /etc/systemd/system/
   sudo systemctl daemon-reload
   sudo systemctl enable wol-server
   sudo systemctl start wol-server
   ```

4. **Install required dependencies**:
   ```bash
   sudo apt-get update
   sudo apt-get install -y wakeonlan sshpass
   ```

## Configuration

The application can be configured by editing the `.env` file in the installation directory:

```bash
nano ~/wol-server/.env
```

### Available Configuration Options

| Setting | Description | Default |
|---------|-------------|---------|
| `SERVER_NAME` | Hostname/IP of target server | pippo |
| `SERVER_USER` | SSH username for shutdown | root |
| `MAC_ADDRESS` | MAC address for Wake-on-LAN | aa:bb:cc:dd:ee:ff |
| `PORT` | Web interface port | 8080 |

After changing configuration, restart the service:
```bash
sudo systemctl restart wol-server
```

## Usage

### Accessing the Interface

Open a web browser and navigate to:
```
http://your-pi-ip:8080
```

### Features

- **Status Checking**: The interface shows the current status (Online/Offline)
- **Booting**: Click the "Boot" button to send a WOL magic packet
- **Shutting Down**: Click "Shutdown" and enter your SSH password when prompted

## Maintenance

### Checking Service Status

```bash
sudo systemctl status wol-server
```

### Viewing Logs

```bash
sudo journalctl -u wol-server -f
```

### Updating

To update to a newer version:

1. Download and transfer the latest release
2. Stop the service:
   ```bash
   sudo systemctl stop wol-server
   ```
3. Extract the new files:
   ```bash
   tar -xzf wol-server.tar.gz
   ```
4. Run the install script:
   ```bash
   ./install.sh
   ```

## Troubleshooting

### Service Won't Start

Check for template errors:
```bash
ls -la ~/wol-server/templates/
```

Verify the .env file exists:
```bash
cat ~/wol-server/.env
```

### Boot Command Not Working

1. Ensure `wakeonlan` is installed:
   ```bash
   which wakeonlan || sudo apt-get install wakeonlan
   ```
2. Verify the MAC address is correct in your .env file
3. Make sure the target device is properly configured for Wake-on-LAN

### Shutdown Not Working

1. Verify `sshpass` is installed:
   ```bash
   which sshpass || sudo apt-get install sshpass
   ```
2. Check that the SERVER_USER setting in .env is correct
3. Ensure SSH access is working between your Pi and the target server

## Advanced Configuration

### Running on a Different Port

Edit the `.env` file:
```bash
echo "PORT=8181" >> ~/wol-server/.env
```

### Multiple Target Machines

To control multiple devices, you can install multiple instances:
```bash
# Create a second instance
mkdir -p ~/wol-server2/templates
cp -r ~/wol-server/templates/* ~/wol-server2/templates/
cp ~/wol-server/wol-server ~/wol-server2/

# Different config
cat > ~/wol-server2/.env << EOL
SERVER_NAME=server2
SERVER_USER=admin
MAC_ADDRESS=aa:bb:cc:dd:ee:ff
PORT=8081
EOL

# Create a new service
sudo cp /etc/systemd/system/wol-server.service /etc/systemd/system/wol-server2.service
sudo sed -i 's|/home/pi/wol-server|/home/pi/wol-server2|g' /etc/systemd/system/wol-server2.service
sudo systemctl daemon-reload
sudo systemctl enable wol-server2
sudo systemctl start wol-server2
```

## Project Information

Designed for use with Raspberry Pi to provide a simple way to manage servers and devices on your local network. The web interface makes it easy to power on and off machines without having to remember MAC addresses or commands.

### Contributing

Contributions are welcome! Feel free to submit pull requests or open issues to help improve this project.

### License

This project is licensed under the MIT License - see the LICENSE file for details.
