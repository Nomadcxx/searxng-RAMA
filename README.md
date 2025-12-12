![SearXNG RAMA Edition](brand/searxng.png)

A pre-configured SearXNG fork with custom theme and sensible privacy-focused defaults.

## Features

- Custom RAMA color palette (space cadet blue #2b2d42, pantone red #ef233c)
- Pre-built static assets
- Automated TUI installer
- Secure defaults (auto-generated secret keys, center alignment)
  
## Installation

Installer:
```bash
curl -fsSL https://raw.githubusercontent.com/Nomadcxx/searxng-RAMA/main/install.sh | sudo bash
```

build and run manually:
```bash
git clone https://github.com/Nomadcxx/searxng-RAMA.git
cd searxng-RAMA
go build -o rama-installer ./cmd/rama-installer/
sudo ./rama-installer
```

### Docker Installation
```bash
docker run -d --name searxng-rama -p 8855:8855 ghcr.io/nomadcxx/searxng-rama:latest
```

### Docker Compose Installation
```bash
docker-compose up -d
```

## Installation Details

The installer validates your source installation, then deploys all SearXNG files to `/opt/searxng-rama`. It sets up a virtual environment, nstalls dependencies, configures secure default, generates secret key and port 8855 for external access, then starts/enables a systemd service. After installation, should will be accessible at <http://localhost:8855>.

## Uninstallation

The TUI provides an uninstaller which will stop/disable the systemd service, remove installation files from `/opt/searxng-rama`

## Requirements

- Go 1.21 or later
