#!/usr/bin/env python3
"""Flatten a SearXNG theme palette into an unconditional :root palette.

The theme-switch feature serves one pre-built CSS bundle per variant. To build a
variant, the active palette must apply unconditionally (on plain :root), not under
a theme-light/theme-dark class. This script produces that flat palette:

  - A theme whose palette already lives on plain :root (rama) needs no flattening;
    passing "flat" just copies it through.
  - A theme with :root.theme-light / :root.theme-dark blocks (google) gets the
    requested block flattened onto :root, plus the shared plain-:root block
    (fonts / type scale / spacing).

Usage:
  gen-variant.py <definitions.less> light  > google-light.less
  gen-variant.py <definitions.less> dark   > google-dark.less
  gen-variant.py <definitions.less> flat   > rama.less        (pass-through)
"""
import re
import sys


def block_body(text, header_regex):
    """Return the content between the braces of the first block whose header matches."""
    m = re.search(header_regex, text)
    if not m:
        return None
    open_idx = text.index("{", m.start())
    depth = 0
    for j in range(open_idx, len(text)):
        if text[j] == "{":
            depth += 1
        elif text[j] == "}":
            depth -= 1
            if depth == 0:
                return text[open_idx + 1 : j]
    return None


def main():
    if len(sys.argv) != 3 or sys.argv[2] not in ("light", "dark", "flat"):
        sys.exit("usage: gen-variant.py <definitions.less> <light|dark|flat>")

    src = open(sys.argv[1], encoding="utf-8").read()
    which = sys.argv[2]

    if which == "flat":
        # Palette already on plain :root — emit unchanged.
        sys.stdout.write(src)
        return

    shared = block_body(src, r"(?m)^:root\s*\{")
    palette = block_body(src, r":root\.theme-" + which)
    if palette is None:
        sys.exit(f"error: no :root.theme-{which} block found in {sys.argv[1]}")

    # Top-level LESS variable declarations + imports (@icon-font-path, breakpoints,
    # @select-*-svg-path, …). These are theme-independent structural values that the
    # theme LESS needs at compile time — they live outside the :root blocks, so carry
    # them over verbatim or the build fails on an undefined @variable.
    less_vars = re.findall(r"(?m)^@(?!media\b)[\w-]+\s*:[^\n]*", src)
    imports = re.findall(r"(?m)^@import[^\n]*", src)

    out = ["// GENERATED flat palette — do not edit; edit theme/google/definitions.less"]
    out.extend(imports)
    out.append(":root {")
    if shared is not None:
        out.append(shared.rstrip())
    out.append(palette.rstrip())
    out.append("}")
    out.extend(less_vars)
    sys.stdout.write("\n".join(out) + "\n")


if __name__ == "__main__":
    main()
