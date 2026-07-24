![SearXNG RAMA Edition](brand/searxng.png)

SearXNG fork with a redesigned UI and privacy-first defaults.

## Features

- Redesigned UI: atmospheric dark home, Google-shaped results rail, workbench preferences (`design.md` holds the locked design system)
- Three switchable theme variants, built at install time: **rama** (dark, default), **google-light**, **google-dark**
- Self-hosted fonts (Inter + JetBrains Mono), no CDN calls
- Secure defaults: per-machine secret key, hardened systemd unit (`PrivateTmp`, `NoNewPrivileges`)
- WCAG AA contrast on every text/surface pair, enforced by `docs/redesign/check-contrast.py`

<img src="brand/screenshot.png" alt="SearXNG RAMA home" style="width: 100%; border-radius: 8px;"/>

<div style="display: grid; grid-template-columns: repeat(2, 1fr); gap: 20px; margin: 20px 0;">
  <img src="brand/screenshot_search.png" alt="SearXNG RAMA results" style="width: 100%; border-radius: 8px;"/>
  <img src="brand/screenshot_preferences.png" alt="SearXNG RAMA preferences" style="width: 100%; border-radius: 8px;"/>
</div>

## Installation

### Arch Linux (AUR)
```bash
yay -S searxng-rama
# or
paru -S searxng-rama

sudo systemctl enable --now searxng-rama.service
```

The package installs with the RAMA theme active. To switch variants:
```bash
searxng-rama-theme              # list variants, show which is active
sudo searxng-rama-theme google-dark
```

### Docker / Docker Compose

Pick the theme with `RAMA_THEME` when you deploy. The container applies it at
start:

```bash
RAMA_THEME=google-dark docker compose up -d
# or plain docker:
docker build -t searxng-rama .
docker run -d --name searxng-rama -p 8855:8855 -e RAMA_THEME=rama searxng-rama
```

Variants: `rama` (default) · `google-light` · `google-dark`. Each container
generates its own secret key unless you pin one with `-e SEARXNG_SECRET=<hex>`.

### Debian / Ubuntu / Fedora (bare metal)
```bash
curl -fsSL https://raw.githubusercontent.com/Nomadcxx/searxng-RAMA/main/install.sh | sudo bash
```

The script installs dependencies, builds every theme variant, and launches the
TUI installer. Re-run the installer to switch theme or uninstall.

### Manual build
```bash
git clone https://github.com/Nomadcxx/searxng-RAMA.git
cd searxng-RAMA
go build -o rama-installer ./cmd/rama-installer/
sudo ./rama-installer
```

## Installation details

The installer puts everything in `/opt/searxng-rama`: a Python virtual
environment, the SearXNG source with the RAMA theme compiled in, a secret key,
and a systemd service on port 8855. Then visit <http://localhost:8855>.

Each variant ships as a pre-built CSS bundle
(`sxng-{ltr,rtl}.<variant>.min.css`). Switching copies the chosen bundle over
the served one and restarts the service; nothing recompiles.

## Uninstallation

- **AUR**: `sudo pacman -R searxng-rama`
- **Bare metal**: run the installer again and choose uninstall
- **Docker**: `docker compose down`

## Requirements

Each install path installs its own dependencies. For manual builds: Go ≥ 1.21,
Node ≥ 20 from a distro or NodeSource package (the nodejs.org tarball's npm
breaks the web build), Python ≥ 3.10.
