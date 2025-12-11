# SearXNG RAMA Edition Dockerfile
# This creates a container with SearXNG RAMA Edition bootstrapped for Docker

FROM ubuntu:22.04

# Avoid prompts from apt
ENV DEBIAN_FRONTEND=noninteractive

# Install dependencies
RUN apt-get update && apt-get install -y \
    curl \
    git \
    python3 \
    python3-pip \
    python3-venv \
    openssl \
    && rm -rf /var/lib/apt/lists/*

# Create the expected source directory
RUN mkdir -p /home/nomadx/searxng-custom

# Copy the SearXNG source
COPY searxng-custom /home/nomadx/searxng-custom

# Copy the RAMA project files
COPY . /app/searxng-RAMA

# Run the Docker bootstrap script
RUN /app/searxng-RAMA/scripts/bootstrap-docker.sh

# Verify installation
RUN test -d /opt/searxng-rama && \
    test -f /opt/searxng-rama/searx/settings.yml && \
    test -d /opt/searxng-rama/searx/static/themes/rama

# Expose the port
EXPOSE 8855

# Start SearXNG directly
WORKDIR /opt/searxng-rama
CMD ["./venv/bin/python", "-m", "searx.webapp"]