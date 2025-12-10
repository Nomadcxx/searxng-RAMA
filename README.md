![SearXNG RAMA Edition](brand/searxng.png)

A pre-configured SearXNG fork with custom theme and sensible privacy-focused defaults.

## Features

- Custom RAMA color palette (space cadet blue #2b2d42, pantone red #ef233c)
- ASCII art-based logo
- Pre-built static assets
- Automated TUI installer
- Secure defaults (auto-generated secret keys, center alignment)

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
- Install SearXNG (RAMA Edition)
- Uninstall SearXNG (RAMA Edition)

Use arrow keys or k/j to navigate, Enter to select, Ctrl+C or q to quit.

## Installation Details

The installer validates your source installation, then deploys all SearXNG files to `/opt/searxng-rama`. It sets up a Python virtual environment and installs dependencies, configures secure defaults including an auto-generated secret key and port 8855 for external access, then creates and enables a systemd service. After installation completes, SearXNG RAMA will be accessible at `http://localhost:8855`.

## Uninstallation

The uninstaller stops and disables the systemd service, removes all installation files from `/opt/searxng-rama`, and cleans up the systemd service configuration.

## Requirements

- Go 1.21 or later
- Python 3.8 or later
- Root access for installation
- SearXNG source at `/home/nomadx/searxng-custom`

## License

This project customizes SearXNG, which is licensed under AGPLv3.
