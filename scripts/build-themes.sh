#!/usr/bin/env bash
# Canonical RAMA theme build.
#
# Builds every switchable theme variant's CSS into a SearXNG checkout, installs
# the self-hosted fonts, applies the template forks and RAMA branding. This is
# the single source of truth for "how the RAMA theme is built" — the PKGBUILD
# inlines the same steps for the AUR path; the cross-distro install.sh calls
# this script directly.
#
# Usage: build-themes.sh <searxng_checkout> <rama_repo_root>
set -euo pipefail

SXNG="${1:?searxng checkout dir required}"
RAMA="${2:?rama repo root required}"
THEME="$RAMA/theme"
CSSDIR="$SXNG/searx/static/themes/simple"

cd "$SXNG"

# --- 1. RAMA (default) palette + shared, theme-neutral redesign layer ---
cp "$THEME/rama/definitions.less" "client/simple/src/less/definitions.less"
mkdir -p "client/simple/src/less/themes/rama"
cp "$THEME/rama/rama.less"  "client/simple/src/less/themes/rama/rama.less"
cp "$THEME/rama/fonts.less" "client/simple/src/less/themes/rama/fonts.less"
grep -q 'themes/rama/rama.less' "client/simple/src/less/style.less" \
  || echo '@import "themes/rama/rama.less";' >> "client/simple/src/less/style.less"

# --- 2. branding assets vite consumes (must exist before the build) ---
mkdir -p "client/simple/src/brand" "client/simple/src/svg"
cat > "client/simple/src/brand/searxng.svg" <<'EOF'
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 20"><text x="5" y="15" font-family="monospace" font-size="12" fill="#e11235">SEARXNG</text></svg>
EOF
[ -f "$RAMA/assets/favicon.svg" ] && cp "$RAMA/assets/favicon.svg" "client/simple/src/brand/searxng-wordmark.svg"
[ -f "$RAMA/assets/empty_favicon.svg" ] && cp "$RAMA/assets/empty_favicon.svg" "client/simple/src/svg/empty_favicon.svg"

# --- 3. build every variant; stash CSS OUTSIDE the vite output dir ---
# vite's emptyOutDir wipes the output dir on each build, so bundles must be
# copied out and restored after the final build.
STASH="$(mktemp -d)"
(
  cd client/simple
  npm install --no-audit --no-fund --ignore-scripts
  npx vite build   # first build = RAMA (default)
  cp "$CSSDIR/sxng-ltr.min.css" "$STASH/sxng-ltr.rama.min.css"
  cp "$CSSDIR/sxng-rtl.min.css" "$STASH/sxng-rtl.rama.min.css"
  for variant in google-light google-dark; do
    python3 "$THEME/gen-variant.py" "$THEME/google/definitions.less" "${variant#google-}" \
      > src/less/definitions.less
    npx vite build
    cp "$CSSDIR/sxng-ltr.min.css" "$STASH/sxng-ltr.${variant}.min.css"
    cp "$CSSDIR/sxng-rtl.min.css" "$STASH/sxng-rtl.${variant}.min.css"
  done
)
cp "$STASH"/*.min.css "$CSSDIR/"
cp "$STASH/sxng-ltr.rama.min.css" "$CSSDIR/sxng-ltr.min.css"
cp "$STASH/sxng-rtl.rama.min.css" "$CSSDIR/sxng-rtl.min.css"
rm -rf "$STASH"

# --- 4. fonts + template forks + branding (after builds, so nothing wipes them) ---
mkdir -p "$CSSDIR/fonts"
cp "$THEME/rama/fonts/"*.woff2 "$CSSDIR/fonts/" 2>/dev/null || true
cp "$THEME/rama/templates/index.html"   "searx/templates/simple/index.html"
cp "$THEME/rama/templates/results.html" "searx/templates/simple/results.html"
[ -f "$RAMA/brand/searxng.png"  ] && cp "$RAMA/brand/searxng.png"  "$CSSDIR/img/searxng.png"
[ -f "$RAMA/assets/favicon.svg" ] && cp "$RAMA/assets/favicon.svg" "$CSSDIR/img/favicon.svg"
[ -f "$RAMA/assets/favicon.png" ] && cp "$RAMA/assets/favicon.png" "$CSSDIR/img/favicon.png"

echo "RAMA theme build complete. Variant bundles:"
ls -1 "$CSSDIR"/sxng-ltr.*.min.css
