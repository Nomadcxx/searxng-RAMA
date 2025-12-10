#!/bin/bash
# SearXNG RAMA Edition one-line installer
# Usage: curl -fsSL https://raw.githubusercontent.com/Nomadcxx/searxng-RAMA/main/install.sh | sudo bash

set -e

echo "SearXNG RAMA Edition Installer"
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Error: This script must be run as root"
    echo "Usage: curl -fsSL https://raw.githubusercontent.com/Nomadcxx/searxng-RAMA/main/install.sh | sudo bash"
    exit 1
fi

# Check for Go
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed"
    echo "Install Go first: https://go.dev/doc/install"
    exit 1
fi

# Check for git
if ! command -v git &> /dev/null; then
    echo "Error: git is not installed"
    echo "Please install git first"
    exit 1
fi

# Create temp directory
TEMP_DIR=$(mktemp -d)
cd "$TEMP_DIR"

echo "Cloning SearXNG RAMA..."
git clone --depth 1 https://github.com/Nomadcxx/searxng-RAMA.git
cd searxng-RAMA

echo "Building installer..."
go build -o rama-installer ./cmd/rama-installer/

echo "Running installer..."
./rama-installer < /dev/tty

# Cleanup
cd /
rm -rf "$TEMP_DIR"

echo ""
echo "Installation complete!"
