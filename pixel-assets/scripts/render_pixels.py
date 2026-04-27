"""
Pixel art renderer: reads JSON spec → draws PNG.

Usage:
    python scripts/render_pixels.py specs/slash-green-0.json -o output/

JSON format:
{
    "name": "slash-green-0",
    "size": [48, 48],
    "constraints": {
        "maxPixels": 10,
        "allowAlpha": false
    },
    "pixels": [
        {"x": 20, "y": 24, "color": "#2bb873"},
        {"x": 24, "y": 26, "color": "#4ade80"}
    ]
}
"""

import json, sys, os, argparse
from PIL import Image


def hex_to_rgb(hex_color: str) -> tuple:
    hex_color = hex_color.lstrip("#")
    return tuple(int(hex_color[i : i + 2], 16) for i in (0, 2, 4))


def render(spec_path: str, output_dir: str, dry_run: bool = False) -> str:
    with open(spec_path) as f:
        spec = json.load(f)

    w, h = spec["size"]
    img = Image.new("RGBA", (w, h), (0, 0, 0, 0))
    pixels_data = spec["pixels"]

    for p in pixels_data:
        x, y = p["x"], p["y"]
        color = hex_to_rgb(p.get("color", "#000000"))
        alpha = p.get("alpha", 255)
        if x < 0 or x >= w or y < 0 or y >= h:
            print(f"  ⚠ Pixel ({x},{y}) out of bounds, skipping")
            continue
        img.putpixel((x, y), (color[0], color[1], color[2], alpha))

    if not dry_run:
        os.makedirs(output_dir, exist_ok=True)
        name = spec.get("name", os.path.splitext(os.path.basename(spec_path))[0])
        out_path = os.path.join(output_dir, f"{name}.png")
        img.save(out_path)
        print(f"  ✓ {name}.png — {len(pixels_data)} pixels")
        return out_path
    else:
        print(f"  ~ {spec.get('name')} — {len(pixels_data)} pixels (dry run)")
        return ""


def main():
    parser = argparse.ArgumentParser(description="Render pixel art JSON specs to PNG")
    parser.add_argument("input", nargs="+", help="JSON spec file(s) or directory")
    parser.add_argument("-o", "--output", default="output", help="Output directory")
    parser.add_argument("--dry-run", action="store_true", help="Validate only, no file output")
    args = parser.parse_args()

    spec_files = []
    for inp in args.input:
        if os.path.isdir(inp):
            for f in sorted(os.listdir(inp)):
                if f.endswith(".json"):
                    spec_files.append(os.path.join(inp, f))
        elif os.path.isfile(inp):
            spec_files.append(inp)

    if not spec_files:
        print("No spec files found.")
        sys.exit(1)

    print(f"Rendering {len(spec_files)} spec(s)...")
    for sf in spec_files:
        render(sf, args.output, args.dry_run)

    print("Done.")


if __name__ == "__main__":
    main()
