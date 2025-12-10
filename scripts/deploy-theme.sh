#!/bin/bash
# Deploy RAMA theme from searxng-custom to live installation

set -e

echo "Deploying RAMA theme to live installation..."

# Sync static assets
echo "Syncing static assets..."
sudo rsync -av --delete /home/nomadx/searxng-custom/searx/static/themes/simple/ /opt/searxng-rama/searx/static/themes/simple/

# Restart service
echo "Restarting SearXNG service..."
sudo systemctl restart searxng-rama

# Wait for service to start
sleep 2

# Verify service is running
if systemctl is-active --quiet searxng-rama; then
    echo "SearXNG RAMA is running"
    echo "Access at: http://localhost:8855"
else
    echo "SearXNG RAMA failed to start"
    echo "Check logs with: sudo journalctl -u searxng-rama -n 50"
    exit 1
fi
