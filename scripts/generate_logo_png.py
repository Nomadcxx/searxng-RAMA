#!/usr/bin/env python3
"""
Generate RAMA SearXNG logo PNG from ASCII art.
Converts /home/nomadx/bit/SEARXNG.txt to brand/searxng.png
"""

from PIL import Image, ImageDraw, ImageFont
import os

# RAMA color palette
BG_COLOR = "#2b2d42"      # Space cadet blue
TEXT_COLOR = "#ef233c"    # Pantone red

# Read ASCII art
ascii_file = "/home/nomadx/bit/SEARXNG.txt"
with open(ascii_file, 'r') as f:
    lines = f.readlines()

# Clean up lines (remove line numbers if present)
ascii_lines = []
for line in lines:
    # Remove leading spaces and line numbers
    if '→' in line:
        line = line.split('→', 1)[1]
    ascii_lines.append(line.rstrip())

# Calculate dimensions
# Use a monospace font size that will make the logo readable
font_size = 20
char_width = font_size * 0.6  # Approximate monospace character width
char_height = font_size * 1.2  # Line height

max_line_length = max(len(line) for line in ascii_lines)
img_width = int(max_line_length * char_width) + 40  # Add padding
img_height = int(len(ascii_lines) * char_height) + 40  # Add padding

# Create image
img = Image.new('RGB', (img_width, img_height), BG_COLOR)
draw = ImageDraw.Draw(img)

# Try to use a monospace font
try:
    font = ImageFont.truetype("/usr/share/fonts/TTF/DejaVuSansMono.ttf", font_size)
except:
    try:
        font = ImageFont.truetype("/usr/share/fonts/truetype/dejavu/DejaVuSansMono.ttf", font_size)
    except:
        # Fallback to default font
        font = ImageFont.load_default()

# Draw ASCII art
y_offset = 20
for line in ascii_lines:
    draw.text((20, y_offset), line, fill=TEXT_COLOR, font=font)
    y_offset += char_height

# Save PNG
output_file = "/home/nomadx/searxng-custom/client/simple/src/brand/searxng.png"
img.save(output_file, 'PNG')
print(f"✓ Generated {output_file}")
print(f"  Dimensions: {img_width}x{img_height}px")
print(f"  Colors: {BG_COLOR} (bg), {TEXT_COLOR} (text)")
