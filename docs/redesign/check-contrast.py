#!/usr/bin/env python3
"""RAMA theme WCAG contrast gate.

Objective acceptance test for the redesign: every text/background pair used by
the theme must clear its WCAG target. Run after editing any theme's colours.

Usage:  python3 docs/redesign/check-contrast.py
Exit 0 = all pass, exit 1 = one or more fail (prints the failures).

Targets: normal text >= 4.5:1 (AA), large/bold >= 3.0:1 (AA-large).
Add a row for every new coloured text-on-surface pair you introduce.
"""
import sys


def _lin(c):
    c /= 255
    return c / 12.92 if c <= 0.03928 else ((c + 0.055) / 1.055) ** 2.4


def _lum(hexs):
    h = hexs.lstrip("#")
    r, g, b = int(h[0:2], 16), int(h[2:4], 16), int(h[4:6], 16)
    return 0.2126 * _lin(r) + 0.7152 * _lin(g) + 0.0722 * _lin(b)


def ratio(a, b):
    la, lb = _lum(a), _lum(b)
    hi, lo = max(la, lb), min(la, lb)
    return (hi + 0.05) / (lo + 0.05)


# ---- locked RAMA dark palette (docs/plan-of-record §7) ----
GROUND, SURFACE, SURFACE2 = "#1e2030", "#282a3b", "#313349"
INK, MUTED, FAINT = "#eef2f6", "#aab4c5", "#8d99ae"
ACCENT, ACCENT_FILL, ACCENT_FILL2 = "#ff7282", "#e11235", "#c30a29"
OK, WARN, ERROR_BG = "#5fd08a", "#e6c15a", "#c30a29"
WHITE = "#ffffff"

# (label, fg, bg, minimum-ratio)
CHECKS = [
    ("body ink / ground",               INK,    GROUND,       4.5),
    ("body ink / surface",              INK,    SURFACE,      4.5),
    ("result title (hover) / surface",   ACCENT, SURFACE,      4.5),
    ("result title (hover) / ground",    ACCENT, GROUND,       4.5),
    ("url + meta muted / surface",       MUTED,  SURFACE,      4.5),
    ("url + meta muted / ground",        MUTED,  GROUND,       4.5),
    ("engine badge muted / surface2",    MUTED,  SURFACE2,     4.5),
    ("CTA label white / accent-fill",    WHITE,  ACCENT_FILL,  4.5),
    ("CTA label white / accent-fill-2",  WHITE,  ACCENT_FILL2, 4.5),
    ("faint text / ground",              FAINT,  GROUND,       4.5),
    ("faint text / surface",             FAINT,  SURFACE,      4.5),
    ("accent / surface-2",              ACCENT, SURFACE2,     4.5),
    ("success text / surface",           OK,     SURFACE,      4.5),
    ("success text / ground",            OK,     GROUND,       4.5),
    ("warning text / surface",            WARN,   SURFACE,      4.5),
    ("warning text / ground",             WARN,   GROUND,       4.5),
    ("error text (white) / error-bg",    WHITE,  ERROR_BG,     4.5),
]

fails = []
print(f"{'ratio':>6}  {'target':>6}  result  pair")
for label, fg, bg, need in CHECKS:
    r = ratio(fg, bg)
    ok = r >= need
    print(f"{r:6.2f}  {need:6.1f}  {'PASS ' if ok else 'FAIL!'}  {label}")
    if not ok:
        fails.append((label, r, need))

if fails:
    print(f"\n{len(fails)} contrast failure(s):")
    for label, r, need in fails:
        print(f"  - {label}: {r:.2f} < {need}")
    sys.exit(1)
print("\nAll contrast checks pass.")
