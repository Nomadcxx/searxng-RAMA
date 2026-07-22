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


WHITE = "#ffffff"


def palette_checks(name, p):
    """Standard text/surface pairs every theme variant must clear at AA (4.5)."""
    g, s, s2 = p["ground"], p["surface"], p["surface2"]
    return [
        (f"{name}: ink / ground",          p["ink"],   g,  4.5),
        (f"{name}: ink / surface",         p["ink"],   s,  4.5),
        (f"{name}: title-hover / surface", p["accent"], s, 4.5),
        (f"{name}: title-hover / ground",  p["accent"], g, 4.5),
        (f"{name}: muted / surface",       p["muted"], s,  4.5),
        (f"{name}: muted / ground",        p["muted"], g,  4.5),
        (f"{name}: engine badge / surf-2", p["muted"], s2, 4.5),
        (f"{name}: faint / ground",        p["faint"], g,  4.5),
        (f"{name}: faint / surface",       p["faint"], s,  4.5),
        (f"{name}: CTA white / fill",      WHITE, p["fill"],  4.5),
        (f"{name}: CTA white / fill-2",    WHITE, p["fill2"], 4.5),
        (f"{name}: success / surface",     p["ok"],   s,  4.5),
        (f"{name}: warning / surface",     p["warn"], s,  4.5),
    ]


# ---- one palette per switchable variant (docs/plan-of-record §7) ----
RAMA = dict(ground="#1e2030", surface="#282a3b", surface2="#313349",
            ink="#eef2f6", muted="#aab4c5", faint="#8d99ae",
            accent="#ff7282", fill="#e11235", fill2="#c30a29", ok="#5fd08a", warn="#e6c15a")
GOOGLE_LIGHT = dict(ground="#ffffff", surface="#ffffff", surface2="#f1f3f4",
                    ink="#202124", muted="#5f6368", faint="#6a6f74",
                    accent="#1a0dab", fill="#e11235", fill2="#c30a29", ok="#0d652d", warn="#b06000")
GOOGLE_DARK = dict(ground="#202124", surface="#303134", surface2="#3c3d40",
                   ink="#e8eaed", muted="#a8aeb4", faint="#979da3",
                   accent="#8ab4f8", fill="#e11235", fill2="#c30a29", ok="#81c995", warn="#fdd663")

CHECKS = (palette_checks("rama", RAMA)
          + palette_checks("google-light", GOOGLE_LIGHT)
          + palette_checks("google-dark", GOOGLE_DARK))

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
