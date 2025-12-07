# searxng-RAMA theme assets

Minimal RAMA theme assets for SearXNG.

## Contents
- `theme/rama/definitions.less` — RAMA color palette and UI variables.
- `brand/rama.svg` — Logo generated from `/home/nomadx/bit/RAMA.txt` (ASCII blocks to SVG).
- `scripts/generate_logo_svg.py` — Regenerate the SVG from the ASCII source.

## Regenerate logo
```bash
./scripts/generate_logo_svg.py
```

## Integrate with SearXNG
- Drop `theme/rama/definitions.less` into your SearXNG theme path and include it in the build (e.g., `client/simple/src/less/definitions.less` override or custom theme include).
- Replace the SearXNG logo asset reference with `brand/rama.svg` (e.g., under `client/simple/src/brand/`).
- Rebuild static assets per your SearXNG setup (e.g., `make themes.simple`).
