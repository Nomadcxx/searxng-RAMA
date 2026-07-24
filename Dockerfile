# SearXNG RAMA Edition — container image.
#
# Multi-stage: the builder carries the whole web toolchain (Node 20 via
# NodeSource — verified: the nodejs.org tarball's npm silently skips rolldown's
# prebuilt native binding, NodeSource works); the runtime image is just Python.
#
# Theme selection is declarative via RAMA_THEME (rama | google-light |
# google-dark) — every variant is pre-built into the image and the entrypoint
# publishes the chosen one at start. See docker-compose.yaml.

# --- build stage -----------------------------------------------------------
FROM ubuntu:24.04 AS builder

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y --no-install-recommends \
        git curl ca-certificates python3 python3-venv python3-dev \
        gcc make libffi-dev \
    && curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y nodejs \
    && rm -rf /var/lib/apt/lists/*

# upstream SearXNG source
RUN git clone --depth 1 https://github.com/searxng/searxng.git /build/searxng

# RAMA theme sources + build script
COPY . /build/rama

# build every theme variant's CSS, install fonts, apply template forks/branding
RUN bash /build/rama/scripts/build-themes.sh /build/searxng /build/rama

# assemble the install tree. version_frozen pins the version so searx never
# shells out to git at import time (the runtime image has neither git nor .git)
RUN mkdir -p /opt/searxng-rama \
    && cp -r /build/searxng/searx /opt/searxng-rama/ \
    && cp /build/searxng/requirements.txt /opt/searxng-rama/ \
    && printf '%s\n' \
       'VERSION_STRING = "1.0.0-RAMA"' \
       'VERSION_TAG = "1.0.0-RAMA"' \
       'DOCKER_TAG = "1.0.0-RAMA"' \
       'GIT_URL = "https://github.com/Nomadcxx/searxng-RAMA"' \
       'GIT_BRANCH = "main"' \
       > /opt/searxng-rama/searx/version_frozen.py

# venv + python deps
RUN python3 -m venv /opt/searxng-rama/venv \
    && /opt/searxng-rama/venv/bin/pip install --no-cache-dir --upgrade pip wheel \
    && /opt/searxng-rama/venv/bin/pip install --no-cache-dir -r /opt/searxng-rama/requirements.txt

# static settings (the secret key stays a placeholder — generated per container
# by the entrypoint so no two containers share one)
RUN set -e; s=/opt/searxng-rama/searx/settings.yml; \
    grep -q 'secret_key: "ultrasecretkey"' "$s"; \
    sed -i 's/port: 8888/port: 8855/' "$s"; \
    sed -i 's/bind_address: "127.0.0.1"/bind_address: "0.0.0.0"/' "$s"; \
    sed -i 's/instance_name: "SearXNG"/instance_name: "SearXNG RAMA Edition"/' "$s"; \
    sed -i 's/image_proxy: false/image_proxy: true/' "$s"; \
    sed -i 's/limiter: true/limiter: false/' "$s"

# --- runtime stage ---------------------------------------------------------
FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y --no-install-recommends python3 \
    && rm -rf /var/lib/apt/lists/* \
    && useradd --system --home-dir /opt/searxng-rama --shell /usr/sbin/nologin searxng

COPY --from=builder --chown=searxng:searxng /opt/searxng-rama /opt/searxng-rama
COPY --chown=root:root --chmod=755 scripts/docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh

ENV SEARXNG_SETTINGS_PATH=/opt/searxng-rama/searx/settings.yml
# rama | google-light | google-dark
ENV RAMA_THEME=rama

USER searxng
EXPOSE 8855
HEALTHCHECK --interval=30s --timeout=5s --start-period=20s \
    CMD /opt/searxng-rama/venv/bin/python -c "import urllib.request; urllib.request.urlopen('http://127.0.0.1:8855/healthz')" || exit 1

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
