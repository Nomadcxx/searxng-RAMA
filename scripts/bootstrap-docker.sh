#!/bin/bash
# Docker bootstrap script for SearXNG RAMA Edition
# This script replicates the essential functionality of the TUI installer
# but works in Docker containers without TTY requirements

set -e

echo "Bootstraping SearXNG RAMA Edition for Docker..."

# Configuration
INSTALL_PATH="/opt/searxng-rama"
SOURCE_PATH="/home/nomadx/searxng-custom"
VENV_PATH="$INSTALL_PATH/venv"

# Create installation directory
echo "Creating installation directory..."
mkdir -p "$INSTALL_PATH"

# Copy SearXNG files
echo "Copying SearXNG files..."
cp -r "$SOURCE_PATH/searx" "$INSTALL_PATH/"
cp -r "$SOURCE_PATH/dockerfiles" "$INSTALL_PATH/" 2>/dev/null || echo "No dockerfiles directory"
cp -r "$SOURCE_PATH/docs" "$INSTALL_PATH/" 2>/dev/null || echo "No docs directory"
cp -r "$SOURCE_PATH/utils" "$INSTALL_PATH/" 2>/dev/null || echo "No utils directory"

# Copy essential files
for file in Makefile manage requirements.txt requirements-dev.txt setup.py babel.cfg .git; do
  if [ -e "$SOURCE_PATH/$file" ]; then
    cp -r "$SOURCE_PATH/$file" "$INSTALL_PATH/"
  fi
done

# Set up Python virtual environment
echo "Setting up Python virtual environment..."
cd "$INSTALL_PATH"
python3 -m venv "$VENV_PATH"

# Install dependencies
echo "Installing Python dependencies..."
"$VENV_PATH/bin/pip" install -r requirements.txt

# Configure settings
echo "Configuring SearXNG settings..."
SETTINGS_PATH="$INSTALL_PATH/searx/settings.yml"

if [ -f "$SETTINGS_PATH" ]; then
  # Generate secret key
  SECRET_KEY=$(openssl rand -hex 32)
  
  # Modify settings
  sed -i "s/secret_key: \"ultrasecretkey\"/secret_key: \"$SECRET_KEY\"/g" "$SETTINGS_PATH"
  sed -i "s/port: 8888/port: 8855/g" "$SETTINGS_PATH"
  sed -i "s/bind_address: \"127.0.0.1\"/bind_address: \"0.0.0.0\"/g" "$SETTINGS_PATH"
  sed -i "s/instance_name: \"SearXNG\"/instance_name: \"SearXNG RAMA Edition\"/g" "$SETTINGS_PATH"
  
  # Enable image proxy and disable limiter
  sed -i "s/image_proxy: false/image_proxy: true/g" "$SETTINGS_PATH"
  sed -i "s/limiter: true/limiter: false/g" "$SETTINGS_PATH"
  
  # Set the theme properly by finding and replacing the default theme setting
  sed -i "s/default_theme: \"simple\"/default_theme: \"rama\"/g" "$SETTINGS_PATH"
  
  # Ensure center alignment is set
  if grep -q "center_alignment:" "$SETTINGS_PATH"; then
    sed -i "s/center_alignment: .*/center_alignment: true/g" "$SETTINGS_PATH"
  else
    # Add center_alignment to the ui section
    sed -i "/ui:/a\  center_alignment: true" "$SETTINGS_PATH"
  fi
else
  echo "Warning: settings.yml not found, creating basic configuration..."
  cat > "$SETTINGS_PATH" << EOF
use_default_settings: true

server:
  secret_key: "$SECRET_KEY"
  limiter: false
  image_proxy: true
  port: 8855
  bind_address: "0.0.0.0"

general:
  debug: false
  instance_name: "SearXNG RAMA Edition"

ui:
  default_theme: rama
  center_alignment: true
EOF
fi

# Copy RAMA theme - create a proper theme structure
echo "Setting up RAMA theme..."
mkdir -p "$INSTALL_PATH/searx/static/themes/rama"

# Copy the RAMA definitions
cp /app/searxng-RAMA/theme/rama/definitions.less "$INSTALL_PATH/searx/static/themes/rama/"

# Create a basic theme structure by copying from simple theme
cp -r "$INSTALL_PATH/searx/static/themes/simple"/* "$INSTALL_PATH/searx/static/themes/rama/" 2>/dev/null || echo "Could not copy simple theme files"

# Create theme metadata
cat > "$INSTALL_PATH/searx/static/themes/rama/manifest.json" << EOF
{
    "name": "rama",
    "description": "SearXNG RAMA Edition theme",
    "author": "SearXNG RAMA Edition",
    "css": [
        "sxng-ltr.min.css"
    ],
    "js": [
        "sxng-core.min.js"
    ],
    "img": "img/favicon.png"
}
EOF

echo "Bootstrap complete!"
echo "Installation directory: $INSTALL_PATH"
echo "To start SearXNG, run: $VENV_PATH/bin/python -m searx.webapp"