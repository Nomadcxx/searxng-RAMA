# SearXNG RAMA

A complete RAMA-branded SearXNG distribution with custom theme and automated installer.

## Features

- Custom RAMA color palette (space cadet blue #2b2d42, pantone red #ef233c)
- ASCII art-based logo converted to PNG
- Pre-built SearXNG with RAMA theme baked into static assets
- TUI installer for automated deployment

## Quick Start

Build the installer:
```bash
go build -o rama-installer ./cmd/rama-installer/
```

Run the installer (requires root):
```bash
sudo ./rama-installer
```

The installer presents a terminal interface with two options:
- Install RAMA SearXNG
- Uninstall RAMA SearXNG

Use arrow keys or k/j to navigate, Enter to select, Ctrl+C or q to quit.

## Installation Details

The installer performs the following tasks:
- Validates source installation at `/home/nomadx/searxng-custom`
- Creates installation directory at `/opt/searxng-rama`
- Copies all SearXNG files
- Sets up Python virtual environment and installs dependencies
- Configures settings (generates secure secret key, sets port 8855, enables external access)
- Creates systemd service
- Enables and starts the service

After installation, SearXNG RAMA will be accessible at `http://localhost:8855` or your configured domain.

## Uninstallation

The installer can also remove RAMA SearXNG:
- Stops and disables the systemd service
- Removes all installation files
- Removes systemd service file

## Development

Generate logo from ASCII art:
```bash
./scripts/generate_logo_png.py
```

Build static assets:
```bash
cd /home/nomadx/searxng-custom/client/simple
npx vite build
```

## Configuration

The installer automatically configures:
- Port: 8855
- Bind address: 0.0.0.0 (accessible externally)
- Base URL: https://searxng.archpcx.one
- Secret key: Automatically generated secure random key
- Center alignment: Enabled

Settings can be modified after installation in `/opt/searxng-rama/searx/settings.yml`.

## Requirements

- Go 1.21 or later
- Python 3.8 or later
- Root access for installation
- SearXNG source at `/home/nomadx/searxng-custom`

## License

This project customizes SearXNG, which is licensed under AGPLv3.
