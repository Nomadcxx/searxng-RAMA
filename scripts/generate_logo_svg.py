#!/usr/bin/env python3
"""Generate RAMA SVG logo from ASCII art source.

- Reads ASCII art from /home/nomadx/bit/RAMA.txt
- Outputs brand/rama.svg with dark background and red blocks
- Exposes tuning constants for cell size, margin, colors
"""
from pathlib import Path

ASCII_PATH = Path('/home/nomadx/bit/RAMA.txt')
OUT_PATH = Path(__file__).resolve().parents[1] / 'brand' / 'rama.svg'
CELL = 10
MARGIN = 6
BG = '#2b2d42'      # space cadet
FG = '#ef233c'      # pantone red
RADIUS = 2          # corner radius per cell
BG_RADIUS = 8       # background corner radius

def main():
    lines = ASCII_PATH.read_text().splitlines()
    max_cols = max(len(line) for line in lines)
    width = max_cols * CELL + 2 * MARGIN
    height = len(lines) * CELL + 2 * MARGIN

    parts = []
    parts.append(f'<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {width} {height}" role="img" aria-labelledby="title desc">')
    parts.append('  <title id="title">RAMA Logo</title>')
    parts.append(f'  <desc id="desc">Converted from ASCII art in {ASCII_PATH}</desc>')
    parts.append(f'  <rect width="{width}" height="{height}" fill="{BG}" rx="{BG_RADIUS}" />')
    for row, line in enumerate(lines):
        for col, ch in enumerate(line):
            if ch == '█':
                x = MARGIN + col * CELL
                y = MARGIN + row * CELL
                parts.append(f'  <rect x="{x}" y="{y}" width="{CELL}" height="{CELL}" fill="{FG}" rx="{RADIUS}" />')
    parts.append('</svg>')
    OUT_PATH.write_text('\n'.join(parts) + '\n')
    print(f"Wrote {OUT_PATH} ({width}x{height}) from {ASCII_PATH}")


if __name__ == '__main__':
    main()
