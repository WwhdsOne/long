#!/usr/bin/env python3
"""Generate pixel art JSON specs for magic talent nodes."""
import json
import os

SPEC_DIR = "/Users/Learning/web/long/pixel-assets/specs"
os.makedirs(SPEC_DIR, exist_ok=True)


def write_spec(name, pixels, size=48, max_pixels=14):
    spec = {
        "name": name,
        "size": [size, size],
        "constraints": {"maxPixels": max_pixels, "allowAlpha": False},
        "pixels": pixels,
    }
    path = os.path.join(SPEC_DIR, f"{name}.json")
    with open(path, "w") as f:
        json.dump(spec, f, indent=2, ensure_ascii=False)
    print(f"  {name}.json ({len(pixels)}px)")


C = 24

# 奥术蓝系
MB = "#3b82f6"  # base blue
MH = "#7dd3fc"  # highlight
MD = "#1e3a8a"  # dark outline
MW = "#dbeafe"  # white-blue core

print("=== 奥术潮汐 ===")

# magic_core - 魔力涌流: arcane orb with converging tides
write_spec("talent-magic_core", [
    {"x": C, "y": 10, "color": MH},
    {"x": C - 6, "y": 16, "color": MB},
    {"x": C + 6, "y": 16, "color": MB},
    {"x": C - 9, "y": C, "color": MD},
    {"x": C + 9, "y": C, "color": MD},
    {"x": C - 6, "y": 32, "color": MB},
    {"x": C + 6, "y": 32, "color": MB},
    {"x": C, "y": 38, "color": MD},
    {"x": C - 3, "y": C - 3, "color": MH},
    {"x": C + 3, "y": C - 3, "color": MH},
    {"x": C, "y": C, "color": MW},
    {"x": C, "y": C + 5, "color": MB},
], max_pixels=12)

# magic_amp - 法术增幅: rising arcane spire
write_spec("talent-magic_amp", [
    {"x": C, "y": 8, "color": MW},
    {"x": C - 2, "y": 12, "color": MH},
    {"x": C + 2, "y": 12, "color": MH},
    {"x": C - 4, "y": 17, "color": MB},
    {"x": C + 4, "y": 17, "color": MB},
    {"x": C - 2, "y": 22, "color": MB},
    {"x": C + 2, "y": 22, "color": MB},
    {"x": C, "y": 26, "color": MD},
    {"x": C - 6, "y": 28, "color": MH},
    {"x": C + 6, "y": 28, "color": MH},
    {"x": C - 4, "y": 33, "color": MB},
    {"x": C + 4, "y": 33, "color": MB},
    {"x": C, "y": 37, "color": MD},
], max_pixels=13)

# magic_resonance - 法术共鸣: triple inward rings
write_spec("talent-magic_resonance", [
    {"x": C - 8, "y": C - 4, "color": MB},
    {"x": C + 8, "y": C - 4, "color": MB},
    {"x": C - 8, "y": C + 4, "color": MD},
    {"x": C + 8, "y": C + 4, "color": MD},
    {"x": C - 4, "y": C - 7, "color": MH},
    {"x": C + 4, "y": C - 7, "color": MH},
    {"x": C - 4, "y": C + 7, "color": MB},
    {"x": C + 4, "y": C + 7, "color": MB},
    {"x": C - 2, "y": C - 2, "color": MH},
    {"x": C + 2, "y": C - 2, "color": MH},
    {"x": C, "y": C, "color": MW},
    {"x": C, "y": C + 4, "color": MD},
], max_pixels=12)

# magic_splash - 余波扩散: twin side bursts around a core rune
write_spec("talent-magic_splash", [
    {"x": C, "y": C - 6, "color": MW},
    {"x": C, "y": C, "color": MH},
    {"x": C, "y": C + 6, "color": MD},
    {"x": C - 10, "y": C - 2, "color": MB},
    {"x": C - 7, "y": C - 5, "color": MH},
    {"x": C - 6, "y": C + 3, "color": MB},
    {"x": C + 10, "y": C - 2, "color": MB},
    {"x": C + 7, "y": C - 5, "color": MH},
    {"x": C + 6, "y": C + 3, "color": MB},
    {"x": C - 3, "y": C + 8, "color": MB},
    {"x": C + 3, "y": C + 8, "color": MB},
    {"x": C - 2, "y": C - 2, "color": MH},
    {"x": C + 2, "y": C - 2, "color": MH},
], max_pixels=13)

# magic_focus - 奥能聚焦: diamond lens with condensed center
write_spec("talent-magic_focus", [
    {"x": C, "y": 10, "color": MH},
    {"x": C - 5, "y": 16, "color": MB},
    {"x": C + 5, "y": 16, "color": MB},
    {"x": C - 8, "y": C, "color": MD},
    {"x": C + 8, "y": C, "color": MD},
    {"x": C - 5, "y": 32, "color": MB},
    {"x": C + 5, "y": 32, "color": MB},
    {"x": C, "y": 38, "color": MD},
    {"x": C - 2, "y": C - 3, "color": MH},
    {"x": C + 2, "y": C - 3, "color": MH},
    {"x": C - 2, "y": C + 3, "color": MB},
    {"x": C + 2, "y": C + 3, "color": MB},
    {"x": C, "y": C, "color": MW},
], max_pixels=13)

# magic_echo_mark - 回响刻印: marked target rune
write_spec("talent-magic_echo_mark", [
    {"x": C, "y": 8, "color": MH},
    {"x": C, "y": 40, "color": MD},
    {"x": 8, "y": C, "color": MB},
    {"x": 40, "y": C, "color": MB},
    {"x": C - 5, "y": C - 5, "color": MH},
    {"x": C + 5, "y": C - 5, "color": MH},
    {"x": C - 5, "y": C + 5, "color": MB},
    {"x": C + 5, "y": C + 5, "color": MB},
    {"x": C - 2, "y": C - 2, "color": MW},
    {"x": C + 2, "y": C - 2, "color": MW},
    {"x": C, "y": C + 2, "color": MD},
    {"x": C - 9, "y": 14, "color": MB},
    {"x": C + 9, "y": 14, "color": MB},
], max_pixels=13)

# magic_static_flux - 静电外溢: branching static discharge
write_spec("talent-magic_static_flux", [
    {"x": C - 3, "y": 10, "color": MH},
    {"x": C + 1, "y": 14, "color": MW},
    {"x": C - 2, "y": 18, "color": MB},
    {"x": C + 3, "y": 21, "color": MB},
    {"x": C - 1, "y": 24, "color": MH},
    {"x": C + 4, "y": 28, "color": MD},
    {"x": C, "y": 32, "color": MD},
    {"x": C - 6, "y": 17, "color": MH},
    {"x": C - 8, "y": 21, "color": MB},
    {"x": C + 7, "y": 18, "color": MH},
    {"x": C + 9, "y": 22, "color": MB},
    {"x": C - 5, "y": 34, "color": MB},
    {"x": C + 5, "y": 36, "color": MB},
], max_pixels=13)

# magic_pierce - 秘法穿透: narrow crystal lance
write_spec("talent-magic_pierce", [
    {"x": C - 6, "y": 14, "color": MH},
    {"x": C - 3, "y": 17, "color": MH},
    {"x": C, "y": 20, "color": MW},
    {"x": C + 3, "y": 23, "color": MB},
    {"x": C + 6, "y": 26, "color": MB},
    {"x": C + 9, "y": 29, "color": MD},
    {"x": C - 8, "y": 18, "color": MB},
    {"x": C - 5, "y": 21, "color": MB},
    {"x": C - 2, "y": 24, "color": MH},
    {"x": C + 1, "y": 27, "color": MB},
    {"x": C + 4, "y": 30, "color": MD},
    {"x": C + 7, "y": 33, "color": MD},
], max_pixels=12)

# magic_chain_bound - 连锁约束: linked runic nodes
write_spec("talent-magic_chain_bound", [
    {"x": C - 9, "y": C - 2, "color": MB},
    {"x": C - 6, "y": C - 5, "color": MH},
    {"x": C - 3, "y": C - 2, "color": MB},
    {"x": C, "y": C, "color": MW},
    {"x": C + 3, "y": C + 2, "color": MB},
    {"x": C + 6, "y": C + 5, "color": MH},
    {"x": C + 9, "y": C + 2, "color": MD},
    {"x": C - 2, "y": C - 8, "color": MH},
    {"x": C + 2, "y": C - 8, "color": MH},
    {"x": C - 2, "y": C + 8, "color": MB},
    {"x": C + 2, "y": C + 8, "color": MB},
    {"x": C - 6, "y": C + 8, "color": MD},
    {"x": C + 6, "y": C - 8, "color": MD},
], max_pixels=13)

# magic_ultimate - 星陨潮爆: starfall sigil
write_spec("talent-magic_ultimate", [
    {"x": C, "y": 6, "color": MW},
    {"x": C - 4, "y": 12, "color": MH},
    {"x": C + 4, "y": 12, "color": MH},
    {"x": C - 8, "y": 18, "color": MB},
    {"x": C + 8, "y": 18, "color": MB},
    {"x": C - 10, "y": C, "color": MH},
    {"x": C + 10, "y": C, "color": MH},
    {"x": C - 8, "y": 30, "color": MB},
    {"x": C + 8, "y": 30, "color": MB},
    {"x": C - 4, "y": 36, "color": MD},
    {"x": C + 4, "y": 36, "color": MD},
    {"x": C, "y": 42, "color": MD},
    {"x": C - 2, "y": C - 2, "color": MH},
    {"x": C + 2, "y": C - 2, "color": MH},
    {"x": C, "y": C + 3, "color": MW},
], max_pixels=15)
