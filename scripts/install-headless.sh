#!/bin/bash
# Non-interactive SearXNG RAMA installation script

echo "Running SearXNG RAMA installer non-interactively..."

# Create the installation directory
mkdir -p /opt/searxng-rama

# Run the installer with automated input
# Option 1 = Install SearXNG (RAMA Edition)
# Enter to confirm
/app/searxng-RAMA/rama-installer <<< $'1\n'

echo "Installation process completed!"