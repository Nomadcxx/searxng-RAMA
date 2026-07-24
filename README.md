![SearXNG RAMA Edition](brand/searxng.png)

SearXNG fork with a redesigned UI and privacy-first defaults out of the box.

## Features

- Full UI redesign: atmospheric dark home, Google-shaped results rail, workbench preferences (see `design.md` for the locked design system)
- Three switchable theme variants, all pre-built at install time — **rama** (dark, default), **google-light**, **google-dark**
- Self-hosted fonts (Inter + JetBrains Mono), no CDN calls
- Secure defaults: per-machine secret key, hardened systemd unit (`PrivateTmp`, `NoNewPrivileges`)
- WCAG AA contrast, gated in CI by `docs/redesign/check-contrast.py`

<div style="display: grid; grid-template-columns: repeat(2, 1fr); gap: 20px; margin: 20px 0;">
  <img src="brand/screenshot.png" alt="SearXNG RAMA Main" style="width: 100%; border-radius: 8px; box-shadow: 0 2px 8px rgba(0,0,0,0.1);"/>
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

Theme selection is declarative — set `RAMA_THEME` on the container and deploy;
there is nothing to run inside the container:

```bash
RAMA_THEME=google-dark docker compose up -d
# or plain docker:
docker build -t searxng-rama .
docker run -d --name searxng-rama -p 8855:8855 -e RAMA_THEME=rama searxng-rama
```

Variants: `rama` (default) · `google-light` · `google-dark`. A secret key is
generated per container unless you pin one with `-e SEARXNG_SECRET=<hex>`.

### Debian / Ubuntu / Fedora (bare metal)
```bash
curl -fsSL https://raw.githubusercontent.com/Nomadcxx/searxng-RAMA/main/install.sh | sudo bash
```

The script installs dependencies, builds every theme variant, and launches the
TUI installer. Re-run the installer any time to **switch theme** or uninstall.

### Manual build
```bash
git clone https://github.com/Nomadcxx/searxng-RAMA.git
cd searxng-RAMA
go build -o rama-installer ./cmd/rama-installer/
sudo ./rama-installer
```

## Installation details

Everything lands in `/opt/searxng-rama`: a Python virtual environment, the
SearXNG source with the RAMA theme compiled in, a generated secret key, and a
systemd service on port 8855. Once done, visit <http://localhost:8855>.

Theme switching never rebuilds anything — every variant ships as a pre-built
CSS bundle (`sxng-{ltr,rtl}.<variant>.min.css`) and switching just publishes
the chosen bundle and restarts the service.

## Uninstallation

- **AUR**: `sudo pacman -R searxng-rama`
- **Bare metal**: run the installer again and choose uninstall
- **Docker**: `docker compose down`

## Requirements

Handled automatically by each install path. For manual builds: Go ≥ 1.21,
Node ≥ 20 (distro/NodeSource package — the nodejs.org tarball's npm breaks the
web build), Python ≥ 3.10.
