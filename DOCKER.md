# SearXNG RAMA Edition - Docker Setup

This provides Docker support for SearXNG RAMA Edition with the custom theme and configuration.

## Quick Start

Build and run the container:

```bash
docker-compose up -d
```

Access SearXNG at http://localhost:8855

## How it works

The Docker setup:

1. Uses Ubuntu 22.04 as the base
2. Installs all required dependencies
3. Copies the SearXNG source code to `/home/nomadx/searxng-custom`
4. Runs a bootstrap script that replicates the essential functionality of the TUI installer:
   - Creates installation directory at `/opt/searxng-rama`
   - Copies SearXNG files from source to installation directory
   - Sets up a Python virtual environment
   - Installs Python dependencies
   - Generates a secret key and configures settings
   - Copies the RAMA theme files
5. Starts SearXNG directly on port 8855

## Customization

The container includes:
- The RAMA theme as the default theme
- Custom branding and settings
- Port 8855 (matching the native installation)

## Environment Variables

- `SEARXNG_BASE_URL`: Base URL for the instance (default: http://localhost:8855/)

## Notes

This Docker setup bypasses the systemd service creation since Docker containers don't use systemd. Instead, it starts SearXNG directly using the Python module.