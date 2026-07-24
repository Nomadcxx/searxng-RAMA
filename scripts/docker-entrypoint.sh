#!/bin/bash
# Container entrypoint for SearXNG RAMA Edition.
#
# Theme selection is declarative: set RAMA_THEME on the container
# (rama | google-light | google-dark) and the matching pre-built CSS bundle is
# published before SearXNG starts — no script or TUI to run inside the
# container, and switching theme is just redeploying with a different value.
set -euo pipefail

INSTALL_PATH="/opt/searxng-rama"
CSSDIR="$INSTALL_PATH/searx/static/themes/simple"
SETTINGS="$INSTALL_PATH/searx/settings.yml"

# --- theme -----------------------------------------------------------------
THEME="${RAMA_THEME:-rama}"
if [ ! -f "$CSSDIR/sxng-ltr.$THEME.min.css" ]; then
  echo "ERROR: unknown RAMA_THEME '$THEME'. Available variants:" >&2
  for f in "$CSSDIR"/sxng-ltr.*.min.css; do
    v="${f##*/sxng-ltr.}"
    echo "  - ${v%.min.css}" >&2
  done
  exit 1
fi
cp "$CSSDIR/sxng-ltr.$THEME.min.css" "$CSSDIR/sxng-ltr.min.css"
cp "$CSSDIR/sxng-rtl.$THEME.min.css" "$CSSDIR/sxng-rtl.min.css"
echo "RAMA theme: $THEME"

# --- secret key ------------------------------------------------------------
# Prefer SEARXNG_SECRET from the environment (SearXNG reads it natively). If it
# is not provided and settings.yml still carries the build placeholder, generate
# a per-container key so no two containers share one.
if [ -z "${SEARXNG_SECRET:-}" ] && grep -q 'secret_key: "ultrasecretkey"' "$SETTINGS"; then
  key="$("$INSTALL_PATH/venv/bin/python" -c 'import secrets; print(secrets.token_hex(32))')"
  sed -i "s/secret_key: \"ultrasecretkey\"/secret_key: \"${key}\"/" "$SETTINGS"
  echo "Generated per-container secret key."
fi

cd "$INSTALL_PATH"
exec "$INSTALL_PATH/venv/bin/python" -m searx.webapp "$@"
