# Installation Guide

## Installing Docker (Required)

**This system requires Docker to run.** Docker Desktop is the easiest way to get Docker on macOS.

### macOS - Install Docker Desktop

**Option 1: Using Homebrew (Recommended)**
```bash
brew install --cask docker
```

Then start Docker Desktop:
- Open Applications → Docker
- Or run: `open -a Docker`

**Option 2: Direct Download**
1. Visit https://www.docker.com/products/docker-desktop/
2. Download Docker Desktop for Mac (Apple Silicon or Intel)
3. Open the `.dmg` file and drag Docker to Applications
4. Open Docker from Applications
5. Follow the setup wizard

### Verify Docker Installation

After Docker Desktop is running, verify:
```bash
docker --version
docker compose version
```

You should see version numbers. If you get "command not found", make sure:
- Docker Desktop is running (check the menu bar for the Docker icon)
- Restart your terminal after installation

## Quick Start (After Docker is Installed)

Once Docker is installed and running:

```bash
# Start all services (Docker will build and download dependencies)
make up
```

The Dockerfiles will automatically:
1. Download Go dependencies during build
2. Compile the services
3. Run everything

## Installing Go (For Local Development)

If you want to:
- Run services locally (outside Docker)
- Run tests locally
- Develop/debug code
- Generate `go.sum` file

Then you need Go installed.

### macOS

**Option 1: Using Homebrew (Recommended)**
```bash
brew install go
```

**Option 2: Direct Download**
1. Visit https://go.dev/dl/
2. Download the macOS installer
3. Run the installer
4. Verify: `go version`

### Linux

**Option 1: Using Package Manager**
```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install golang-go

# Fedora
sudo dnf install golang

# Arch
sudo pacman -S go
```

**Option 2: Direct Download**
```bash
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

### Windows

1. Visit https://go.dev/dl/
2. Download the Windows installer (.msi)
3. Run the installer
4. Verify: Open PowerShell and run `go version`

### Verify Installation

After installing, verify it works:
```bash
go version
# Should output: go version go1.22.x ...
```

### Generate go.sum

Once Go is installed:
```bash
go mod tidy
```

This will:
- Download all dependencies
- Generate `go.sum` with checksums
- Update `go.mod` if needed

## Troubleshooting

### "command not found: go"
- Go is not installed or not in PATH
- Install Go using instructions above
- Restart your terminal after installation

### "command not found: docker"
- Docker is not installed
- Install Docker Desktop: https://www.docker.com/products/docker-desktop/

### Permission Errors
- On Linux, you may need to add your user to the docker group:
  ```bash
  sudo usermod -aG docker $USER
  # Then log out and back in
  ```

