# Installation Guide

This guide covers all the different ways to install GoCat on your system.

## Release Archives

Every tagged release publishes tarballs per platform plus a Debian package
(see [Releases](https://github.com/realibrahimsql/Gocat/releases)).
Filenames carry the version, e.g. `gocat-1.0.0-linux-amd64.tar.gz`:

```bash
# Linux amd64, version 1.0.0 as example
wget https://github.com/realibrahimsql/Gocat/releases/download/v1.0.0/gocat-1.0.0-linux-amd64.tar.gz
tar -xzf gocat-1.0.0-linux-amd64.tar.gz
sudo install -m 0755 gocat-1.0.0-linux-amd64/gocat /usr/local/bin/gocat
```

---

## Package Managers

Build files for Homebrew (`pkg/Formula/gocat.rb`) and Arch (`pkg/arch/PKGBUILD`)
ship in the repo; no tap or AUR package is published yet, so install from
source or a release archive for now.

### Debian/Ubuntu

```bash
# Download the .deb package (version 1.0.0 as example)
wget https://github.com/realibrahimsql/Gocat/releases/download/v1.0.0/gocat-1.0.0_amd64.deb

# Install with dpkg
sudo dpkg -i gocat-1.0.0_amd64.deb

# Fix dependencies if needed
sudo apt-get install -f
```

---

## Manual Download

### Pre-built Binaries

Download the appropriate binary for your platform from our [releases page](https://github.com/realibrahimsql/Gocat/releases):

#### Linux

```bash
# x86_64, version 1.0.0 as example
wget https://github.com/realibrahimsql/Gocat/releases/download/v1.0.0/gocat-1.0.0-linux-amd64.tar.gz
tar -xzf gocat-1.0.0-linux-amd64.tar.gz
sudo install -m 0755 gocat-1.0.0-linux-amd64/gocat /usr/local/bin/gocat

# ARM64
wget https://github.com/realibrahimsql/Gocat/releases/download/v1.0.0/gocat-1.0.0-linux-arm64.tar.gz
tar -xzf gocat-1.0.0-linux-arm64.tar.gz
sudo install -m 0755 gocat-1.0.0-linux-arm64/gocat /usr/local/bin/gocat
```

#### macOS

```bash
# Intel Macs, version 1.0.0 as example
wget https://github.com/realibrahimsql/Gocat/releases/download/v1.0.0/gocat-1.0.0-darwin-amd64.tar.gz
tar -xzf gocat-1.0.0-darwin-amd64.tar.gz
sudo install -m 0755 gocat-1.0.0-darwin-amd64/gocat /usr/local/bin/gocat

# Apple Silicon (M1/M2)
wget https://github.com/realibrahimsql/Gocat/releases/download/v1.0.0/gocat-1.0.0-darwin-arm64.tar.gz
tar -xzf gocat-1.0.0-darwin-arm64.tar.gz
sudo install -m 0755 gocat-1.0.0-darwin-arm64/gocat /usr/local/bin/gocat
```

#### Windows

**PowerShell:**
```powershell
# x86_64, version 1.0.0 as example (zip archive)
Invoke-WebRequest -Uri "https://github.com/realibrahimsql/Gocat/releases/download/v1.0.0/gocat-1.0.0-windows-amd64.zip" -OutFile "gocat.zip"
Expand-Archive gocat.zip .
```

**Command Prompt:**
```cmd
curl -L -o gocat.zip https://github.com/realibrahimsql/Gocat/releases/download/v1.0.0/gocat-1.0.0-windows-amd64.zip
```

#### FreeBSD

```bash
# x86_64, version 1.0.0 as example
wget https://github.com/realibrahimsql/Gocat/releases/download/v1.0.0/gocat-1.0.0-freebsd-amd64.tar.gz
tar -xzf gocat-1.0.0-freebsd-amd64.tar.gz
sudo install -m 0755 gocat-1.0.0-freebsd-amd64/gocat /usr/local/bin/gocat
```

---

## Docker

No published image; build locally from the repo Dockerfile:

```bash
git clone https://github.com/realibrahimsql/Gocat.git
cd Gocat
docker build -t gocat .
docker run --rm -it gocat --help
docker run --rm -it --network host gocat listen 8080
```

---

## Build from Source

### Prerequisites

- **Go 1.24+**: [Download Go](https://golang.org/dl/)
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

## Post-Installation Setup

### Verify Installation

```bash
# Check version
gocat version

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

## Updating GoCat

### Package Managers

No tap, AUR, or apt repository is published yet; update by reinstalling
from a release archive or rebuilding from source.

### Manual Update

Re-download the tarball for your platform from
[Releases](https://github.com/realibrahimsql/Gocat/releases) and reinstall
the binary as above.

### Docker Update

```bash
docker build -t gocat .
```

---

## Uninstalling GoCat

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

## Troubleshooting

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

# Reinstall if needed: repeat the release-archive steps above
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

- [Report bugs](https://github.com/realibrahimsql/Gocat/issues/new?template=bug_report.yml)

---

## Next Steps

After installation:

1. Read the [User Guide](user-guide.md)
2. Try the [Quick Start](../README.md#quick-start) examples
3. [Contribute](../CONTRIBUTING.md) to the project

