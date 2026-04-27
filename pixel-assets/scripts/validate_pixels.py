"""
Pixel art validator: checks rendered PNG against JSON spec constraints.

Usage:
    python scripts/validate_pixels.py specs/ output/
"""

import json, os, sys, argparse
from PIL import Image


def hex_to_rgb(hex_color: str) -> tuple:
    hex_color = hex_color.lstrip("#")
    return tuple(int(hex_color[i : i + 2], 16) for i in (0, 2, 4))


def validate(spec_path: str, png_path: str) -> list:
    errors = []

    if not os.path.exists(png_path):
        return [f"Missing PNG: {png_path}"]

    with open(spec_path) as f:
        spec = json.load(f)

    img = Image.open(png_path).convert("RGBA")
    w, h = img.size
    sw, sh = spec["size"]

    # Size check
    if w != sw or h != sh:
        errors.append(f"Size mismatch: expected {sw}x{sh}, got {w}x{h}")

    constraints = spec.get("constraints", {})
    max_pixels = constraints.get("maxPixels")
    allow_alpha = constraints.get("allowAlpha", False)

    # Read non-transparent pixels
    non_transparent = []
    unknown_colors = set()
    spec_colors_set = {hex_to_rgb(p["color"]) for p in spec["pixels"] if "color" in p}

    for y in range(h):
        for x in range(w):
            r, g, b, a = img.getpixel((x, y))
            if a > 0:
                non_transparent.append((x, y, (r, g, b), a))

    # Pixel count check
    if max_pixels is not None:
        if len(non_transparent) > max_pixels:
            errors.append(
                f"Pixel count {len(non_transparent)} exceeds max {max_pixels}"
            )

    # Expected count check (from spec pixels)
    expected_pixels = len(spec["pixels"])
    rendered_count = len([p for p in non_transparent if p[3] == 255])

    if rendered_count != expected_pixels:
        errors.append(
            f"Pixel count mismatch: spec has {expected_pixels}, "
            f"rendered has {rendered_count} non-alpha pixels"
        )

    # Check for non-opaque pixels (anti-aliasing detection)
    semi_transparent = [p for p in non_transparent if 0 < p[3] < 255]
    if semi_transparent and not allow_alpha:
        errors.append(
            f"Found {len(semi_transparent)} semi-transparent pixels "
            f"(anti-aliasing not allowed): "
            f"{[(x,y) for x,y,_,a in semi_transparent[:5]]}"
        )

    # Color usage check
    for x, y, rgb, a in non_transparent:
        if rgb not in spec_colors_set:
            unknown_colors.add(rgb)

    if unknown_colors:
        hex_colors = ["#%02x%02x%02x" % c for c in unknown_colors]
        errors.append(
            f"Found {len(unknown_colors)} color(s) not in spec: {hex_colors}"
        )

    # Bounds check
    for x, y, _, _ in non_transparent:
        if x == 0 or x == w - 1 or y == 0 or y == h - 1:
            errors.append(f"Pixel at edge ({x},{y}) — touches canvas boundary")

    return errors


def main():
    parser = argparse.ArgumentParser(
        description="Validate rendered pixel PNG against spec"
    )
    parser.add_argument("spec_dir", help="Directory containing JSON spec files")
    parser.add_argument("png_dir", help="Directory containing rendered PNGs")
    args = parser.parse_args()

    spec_files = sorted(f for f in os.listdir(args.spec_dir) if f.endswith(".json"))

    if not spec_files:
        print("No spec files found.")
        sys.exit(1)

    total_errors = 0
    for sf in spec_files:
        name = os.path.splitext(sf)[0]
        spec_path = os.path.join(args.spec_dir, sf)
        png_path = os.path.join(args.png_dir, f"{name}.png")

        errors = validate(spec_path, png_path)
        if errors:
            total_errors += len(errors)
            print(f"✗ {name}:")
            for e in errors:
                print(f"    {e}")
        else:
            print(f"✓ {name}")

    if total_errors:
        print(f"\n{total_errors} validation error(s) found.")
        sys.exit(1)
    else:
        print("\nAll validations passed.")


if __name__ == "__main__":
    main()
