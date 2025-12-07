# searxng-RAMA theme assets

Minimal RAMA theme assets for SearXNG.

## Contents
- `theme/rama/definitions.less` — RAMA color palette and UI variables.
- `brand/rama.svg` — Logo generated from `/home/nomadx/bit/RAMA.txt` (ASCII blocks to SVG).
- `scripts/generate_logo_svg.py` — Regenerate the SVG from the ASCII source.
- `cmd/rama-installer` — Go CLI to install the theme into a SearXNG checkout.

## Regenerate logo
```bash
./scripts/generate_logo_svg.py
```

## Installer
Build:
```bash
go build ./cmd/rama-installer
```
Run (example):
```bash
./rama-installer -searxng /path/to/searxng -set-default-theme
```
Flags:
- `-searxng` (required): path to SearXNG repo.
- `-theme-name` (default: `rama`): target theme name.
- `-set-default-theme`: patch `searx/settings.yml` to set the default theme.
- `-settings`: optional explicit settings.yml path.

What it does:
- Copies `theme/rama/definitions.less` to `<searxng>/client/simple/src/less/themes/<theme>/definitions.less`.
- Copies `brand/rama.svg` to `<searxng>/client/simple/src/brand/<theme>.svg`.
- If `-set-default-theme`, patches `searx/settings.yml` `theme:` key.

## Integrate manually (optional)
- Drop `theme/rama/definitions.less` into your SearXNG theme path and include it in the build.
- Replace the SearXNG logo asset reference with `brand/rama.svg`.
- Rebuild static assets per your SearXNG setup (e.g., `make themes.simple`).
