#!/bin/bash
# SearXNG RAMA Edition — cross-distro bare-metal installer.
#
# For users who are NOT on Arch (use the AUR package: `yay -S searxng-rama`)
# and who do NOT want Docker (`docker compose up`). Supported here:
#   - Debian / Ubuntu  (apt)
#   - Fedora           (dnf)
#
#   curl -fsSL https://raw.githubusercontent.com/Nomadcxx/searxng-RAMA/main/install.sh | sudo bash
set -euo pipefail

RAMA_REPO="https://github.com/Nomadcxx/searxng-RAMA.git"
SEARXNG_REPO="https://github.com/searxng/searxng.git"
INSTALL_PATH="/opt/searxng-rama"

echo "SearXNG RAMA Edition — cross-distro installer"
echo ""

if [ "${EUID:-$(id -u)}" -ne 0 ]; then
  echo "Error: this script must be run as root (sudo)." >&2
  echo "Usage: curl -fsSL .../install.sh | sudo bash" >&2
  exit 1
fi

# --- distro detection + dependency install --------------------------------
install_deps() {
  if command -v apt-get >/dev/null 2>&1; then
    echo "Detected apt (Debian/Ubuntu) — installing dependencies..."
    export DEBIAN_FRONTEND=noninteractive
    apt-get update -qq
    apt-get install -y --no-install-recommends \
      git curl ca-certificates python3 python3-venv python3-dev \
      openssl gcc make libffi-dev golang-go
  elif command -v dnf >/dev/null 2>&1; then
    echo "Detected dnf (Fedora) — installing dependencies..."
    dnf install -y \
      git curl ca-certificates python3 python3-virtualenv python3-devel \
      nodejs npm openssl gcc make libffi-devel golang
  else
    echo "Unsupported distribution: need apt (Debian/Ubuntu) or dnf (Fedora)." >&2
    echo "  - On Arch Linux, install from the AUR:  yay -S searxng-rama" >&2
    echo "  - Or run the container image:           docker compose up" >&2
    exit 1
  fi
}

# go.mod targets a recent Go; Go >= 1.21 can auto-fetch the required toolchain
# (GOTOOLCHAIN=auto). If the distro Go is older/missing, install the official
# tarball to /usr/local/go.
ensure_go() {
  local cur=0
  if command -v go >/dev/null 2>&1; then
    cur=$(go version | sed -nE 's/.*go1\.([0-9]+).*/\1/p' || echo 0)
  fi
  if [ "${cur:-0}" -ge 21 ] 2>/dev/null; then
    return 0
  fi
  echo "Distro Go missing or too old — installing an official Go toolchain..."
  local arch; arch=$(uname -m)
  case "$arch" in
    x86_64) arch=amd64 ;;
    aarch64|arm64) arch=arm64 ;;
    *) echo "Error: unsupported architecture '$arch' for the Go tarball." >&2; exit 1 ;;
  esac
  local ver="go1.23.4"
  curl -fsSL "https://go.dev/dl/${ver}.linux-${arch}.tar.gz" -o /tmp/go.tgz
  rm -rf /usr/local/go
  tar -C /usr/local -xzf /tmp/go.tgz
  rm -f /tmp/go.tgz
  export PATH="/usr/local/go/bin:$PATH"
}

# SearXNG's build toolchain (rolldown/vite) needs Node >= 20.12, and — verified
# the hard way — it must be a *distro-integrated* Node: the nodejs.org tarball's
# npm silently skips rolldown's prebuilt native binding, so the build fails at
# runtime. Debian/Ubuntu ship Node 18, so install Node 20 from NodeSource (which
# packages it as a proper apt/dnf package). Fedora 40+ already ships Node 20+.
ensure_node() {
  local cur=0
  if command -v node >/dev/null 2>&1; then
    cur=$(node --version | sed -nE 's/^v([0-9]+).*/\1/p' || echo 0)
  fi
  if [ "${cur:-0}" -ge 20 ] 2>/dev/null; then
    return 0
  fi
  echo "Node.js missing or older than 20 — installing Node 20 via NodeSource..."
  if command -v apt-get >/dev/null 2>&1; then
    curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
    apt-get install -y nodejs
  elif command -v dnf >/dev/null 2>&1; then
    curl -fsSL https://rpm.nodesource.com/setup_20.x | bash -
    dnf install -y nodejs
  else
    echo "Error: cannot install a recent Node.js on this distro." >&2
    exit 1
  fi
}

install_deps
ensure_go
ensure_node
export GOTOOLCHAIN=auto

for t in git curl go python3 node npm openssl; do
  command -v "$t" >/dev/null 2>&1 || { echo "Error: required tool '$t' not available after install." >&2; exit 1; }
done

# --- fetch sources, build theme, run installer ----------------------------
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT
cd "$TMP"

echo "Cloning RAMA..."
git clone --depth 1 "$RAMA_REPO" rama
echo "Cloning SearXNG (upstream)..."
git clone --depth 1 "$SEARXNG_REPO" searxng

echo "Building the RAMA theme (compiles CSS for every variant; may take a few minutes)..."
if ! bash rama/scripts/build-themes.sh "$TMP/searxng" "$TMP/rama"; then
  echo "" >&2
  echo "ERROR: the SearXNG web build failed. Most common cause: a Node.js not" >&2
  echo "installed from a distro/NodeSource package (the nodejs.org tarball's npm" >&2
  echo "silently skips the bundler's native binding)." >&2
  echo "" >&2
  echo "Verified working: Debian 12+, Ubuntu 24.04+, Fedora 40+ with the Node 20" >&2
  echo "this script installs. Fallback that works anywhere:  docker compose up" >&2
  exit 1
fi

echo "Building the installer..."
( cd rama && go build -o "$TMP/rama-installer" ./cmd/rama-installer/ )

echo "Launching the installer..."
# Point the installer at the freshly-built SearXNG checkout (source) to copy into
# /opt/searxng-rama (install). Switch/Uninstall modes operate on the install path.
RAMA_SOURCE_PATH="$TMP/searxng" RAMA_INSTALL_PATH="$INSTALL_PATH" \
  "$TMP/rama-installer" < /dev/tty

echo ""
echo "Installation complete."
echo "  Status:  systemctl status searxng-rama"
echo "  Logs:    journalctl -u searxng-rama -f"
echo "  Search:  http://localhost:8855"
