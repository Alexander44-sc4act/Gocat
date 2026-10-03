# 📦 Installation Guide

This guide covers all the different ways to install GoCat on your system.

## 🚀 Quick Installation

### One-Line Install Script

The fastest way to get GoCat up and running:

```bash
curl -sSL https://raw.githubusercontent.com/ibrahmsql/gocat/main/pkg/install.sh | bash
```

This script will:
- 🔍 Detect your operating system and architecture
- 📥 Download the latest release
- 📁 Install to `/usr/local/bin/gocat`
- 🔗 Create a `nc` symlink for compatibility
- ✅ Verify the installation

---

## 📦 Package Managers

### 🍺 Homebrew (macOS/Linux)

```bash
# Add our tap
brew tap ibrahmsql/gocat

# Install GoCat
brew install gocat

# Update to latest version
brew upgrade gocat
```

### 🐧 Arch Linux (AUR)

```bash
# Using yay
yay -S gocat

# Using paru
paru -S gocat

# Manual installation
git clone https://aur.archlinux.org/gocat.git
cd gocat
makepkg -si
```

### 📦 Debian/Ubuntu

```bash
# Download the .deb package
wget https://github.com/realibrahimsql/Gocat/releases/latest/download/gocat_amd64.deb

# Install with dpkg
sudo dpkg -i gocat_amd64.deb

# Fix dependencies if needed
sudo apt-get install -f
```

### 🎩 RPM (RHEL/CentOS/Fedora)

```bash
# Download the .rpm package
wget https://github.com/realibrahimsql/Gocat/releases/latest/download/gocat_amd64.rpm

# Install with rpm
sudo rpm -i gocat_amd64.rpm

# Or with dnf (Fedora)
sudo dnf install gocat_amd64.rpm

# Or with yum (RHEL/CentOS)
sudo yum install gocat_amd64.rpm
```

---

## 📥 Manual Download

### Pre-built Binaries

Download the appropriate binary for your platform from our [releases page](https://github.com/realibrahimsql/Gocat/releases):

#### 🐧 Linux

```bash
# x86_64
wget https://github.com/realibrahimsql/Gocat/releases/latest/download/gocat-linux-amd64
chmod +x gocat-linux-amd64
sudo mv gocat-linux-amd64 /usr/local/bin/gocat

# ARM64
wget https://github.com/realibrahimsql/Gocat/releases/latest/download/gocat-linux-arm64
chmod +x gocat-linux-arm64
sudo mv gocat-linux-arm64 /usr/local/bin/gocat

# ARM v7
wget https://github.com/realibrahimsql/Gocat/releases/latest/download/gocat-linux-armv7
chmod +x gocat-linux-armv7
sudo mv gocat-linux-armv7 /usr/local/bin/gocat
```

#### 🍎 macOS

```bash
# Intel Macs
wget https://github.com/realibrahimsql/Gocat/releases/latest/download/gocat-darwin-amd64
chmod +x gocat-darwin-amd64
sudo mv gocat-darwin-amd64 /usr/local/bin/gocat

# Apple Silicon (M1/M2)
wget https://github.com/realibrahimsql/Gocat/releases/latest/download/gocat-darwin-arm64
chmod +x gocat-darwin-arm64
sudo mv gocat-darwin-arm64 /usr/local/bin/gocat
```

#### 🪟 Windows

**PowerShell:**
```powershell
# x86_64
Invoke-WebRequest -Uri "https://github.com/realibrahimsql/Gocat/releases/latest/download/gocat-windows-amd64.exe" -OutFile "gocat.exe"

# ARM64
Invoke-WebRequest -Uri "https://github.com/realibrahimsql/Gocat/releases/latest/download/gocat-windows-arm64.exe" -OutFile "gocat.exe"
```

**Command Prompt:**
```cmd
curl -L -o gocat.exe https://github.com/realibrahimsql/Gocat/releases/latest/download/gocat-windows-amd64.exe
```

#### 🔥 FreeBSD

```bash
# x86_64
wget https://github.com/realibrahimsql/Gocat/releases/latest/download/gocat-freebsd-amd64
chmod +x gocat-freebsd-amd64
sudo mv gocat-freebsd-amd64 /usr/local/bin/gocat
```

---

## 🐳 Docker

### Using Docker Hub

```bash
# Pull the latest image
docker pull ghcr.io/ibrahmsql/gocat:latest

# Run GoCat in a container
docker run --rm -it ghcr.io/ibrahmsql/gocat:latest

# Run with network access
docker run --rm -it --network host ghcr.io/ibrahmsql/gocat:latest listen 8080
```

### Using Docker Compose

```bash
# Clone the repository
git clone https://github.com/realibrahimsql/Gocat.git
cd gocat

# Run with docker-compose
docker-compose up

# Run specific service
docker-compose up gocat

# Development mode
docker-compose --profile dev up
```

### Available Docker Tags

- `latest` - Latest stable release
- `v1.0.0` - Specific version
- `dev` - Development build
- `alpine` - Alpine-based minimal image

---

## 🛠️ Build from Source

### Prerequisites

- **Go 1.21+**: [Download Go](https://golang.org/dl/)
- **Git**: [Install Git](https://git-scm.com/downloads)
- **Make**: Usually pre-installed on Unix systems

### Build Steps

```bash
# Clone the repository
git clone https://github.com/realibrahimsql/Gocat.git
cd gocat

# Install dependencies
make deps

# Build for your platform
make build

# Or build for all platforms
make build-all

# Install to system
sudo make install
```

### Custom Build Options

```bash
# Build with debug symbols
make build-debug

# Build with race detection
make build-race

# Build for specific platform
GOOS=linux GOARCH=amd64 make build

# Build with custom version
VERSION=1.0.0-custom make build
```

---

## 🔧 Post-Installation Setup

### Verify Installation

```bash
# Check version
gocat --version

# Test basic functionality
gocat --help

# Test network connectivity
gocat connect google.com 80 <<< "GET / HTTP/1.0\r\n\r\n"
```

### Shell Completion

#### Bash

```bash
# Generate completion script
gocat completion bash > /etc/bash_completion.d/gocat

# Or for user-only
gocat completion bash > ~/.bash_completion.d/gocat
source ~/.bash_completion.d/gocat
```

#### Zsh

```bash
# Generate completion script
gocat completion zsh > "${fpath[1]}/_gocat"

# Reload completions
compinit
```

#### Fish

```bash
# Generate completion script
gocat completion fish > ~/.config/fish/completions/gocat.fish
```

### Create Symlink for Netcat Compatibility

```bash
# Create nc symlink
sudo ln -sf /usr/local/bin/gocat /usr/local/bin/nc

# Verify
nc --version
```

### Configuration

Create a configuration file for default settings:

```bash
# Create config directory
mkdir -p ~/.config/gocat

# Create basic config
cat > ~/.config/gocat/config.yaml << EOF
defaults:
  timeout: 30s
  retry: 3
  keep_alive: true
  
logging:
  level: info
  format: text
  
network:
  ipv6: false
  buffer_size: 4096
EOF
```

---

## 🔄 Updating GoCat

### Package Managers

```bash
# Homebrew
brew upgrade gocat

# Arch Linux
yay -Syu gocat

# Debian/Ubuntu
sudo apt update && sudo apt upgrade gocat
```

### Manual Update

```bash
# Using install script
curl -sSL https://raw.githubusercontent.com/ibrahmsql/gocat/main/pkg/install.sh | bash

# Or download latest binary manually
wget https://github.com/realibrahimsql/Gocat/releases/latest/download/gocat-linux-amd64
chmod +x gocat-linux-amd64
sudo mv gocat-linux-amd64 /usr/local/bin/gocat
```

### Docker Update

```bash
# Pull latest image
docker pull ghcr.io/ibrahmsql/gocat:latest

# Update docker-compose
docker-compose pull
docker-compose up -d
```

---

## 🗑️ Uninstalling GoCat

### Package Managers

```bash
# Homebrew
brew uninstall gocat

# Arch Linux
yay -R gocat

# Debian/Ubuntu
sudo apt remove gocat

# RPM
sudo rpm -e gocat
```

### Manual Removal

```bash
# Remove binary
sudo rm -f /usr/local/bin/gocat

# Remove symlink
sudo rm -f /usr/local/bin/nc

# Remove configuration
rm -rf ~/.config/gocat

# Remove completion scripts
sudo rm -f /etc/bash_completion.d/gocat
rm -f ~/.bash_completion.d/gocat
```

---

## 🆘 Troubleshooting

### Common Issues

#### Permission Denied

```bash
# Make sure the binary is executable
chmod +x gocat

# Check if /usr/local/bin is in PATH
echo $PATH

# Add to PATH if needed
export PATH="/usr/local/bin:$PATH"
```

#### Command Not Found

```bash
# Check if gocat is installed
which gocat

# Check installation location
find /usr -name "gocat" 2>/dev/null

# Reinstall if needed
curl -sSL https://raw.githubusercontent.com/ibrahmsql/gocat/main/pkg/install.sh | bash
```

#### Network Issues

```bash
# Test with verbose output
gocat -v connect google.com 80

# Check firewall settings
sudo ufw status

# Test with different port
gocat connect google.com 443
```

### Getting Help

If you encounter issues:

- 📖 Check our [documentation](https://docs.gocat.dev)
- 🐛 [Report bugs](https://github.com/realibrahimsql/Gocat/issues/new?template=bug_report.yml)
- 💬 [Join our Discord](https://discord.gg/gocat)
- 📧 [Email support](mailto:support@gocat.dev)

---

## 🎯 Next Steps

After installation:

1. 📖 Read the [User Guide](user-guide.md)
2. 🎯 Try the [Quick Start](../README.md#quick-start) examples
3. 🔧 Explore [Advanced Usage](advanced-usage.md)
4. 🤝 [Contribute](../CONTRIBUTING.md) to the project

**Happy networking with GoCat!** 🚀