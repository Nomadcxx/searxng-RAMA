#!/bin/bash
# Build script for SearXNG RAMA Docker image

echo "Building SearXNG RAMA Docker image..."

# Check if searxng-custom directory exists
if [ -d "/home/nomadx/searxng-custom" ]; then
    echo "Using existing SearXNG source from /home/nomadx/searxng-custom"
    
    # Create a temporary directory for the build context
    BUILD_DIR=$(mktemp -d)
    echo "Using build directory: $BUILD_DIR"
    
    # Copy the project files
    cp -r /home/nomadx/searxng-RAMA/* "$BUILD_DIR/"
    
    # Copy the SearXNG source
    cp -r /home/nomadx/searxng-custom "$BUILD_DIR/"
    
    # Build the Docker image
    cd "$BUILD_DIR"
    docker build -t searxng-rama:latest .
    
    # Clean up
    cd /
    rm -rf "$BUILD_DIR"
else
    echo "No local SearXNG source found, Dockerfile will clone it during build"
    
    # Build from current directory
    docker build -t searxng-rama:latest .
fi

echo "Build complete!"
echo "To run the container:"
echo "docker run -d --name searxng-rama -p 8855:8855 searxng-rama:latest"