#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
批量重做 pixel specs：提升体块厚度与可读性。

用法：
  python scripts/thicken_specs.py specs/
  python scripts/thicken_specs.py specs/ --dry-run
"""

from __future__ import annotations

import argparse
import json
import math
from collections import Counter
from pathlib import Path


def hex_to_rgb(color: str) -> tuple[int, int, int]:
    value = color.strip().lstrip("#")
    return int(value[0:2], 16), int(value[2:4], 16), int(value[4:6], 16)


def rgb_to_hex(rgb: tuple[int, int, int]) -> str:
    return f"#{rgb[0]:02x}{rgb[1]:02x}{rgb[2]:02x}"


def luminance(rgb: tuple[int, int, int]) -> float:
    r, g, b = rgb
    return 0.2126 * r + 0.7152 * g + 0.0722 * b


def neighbors4(x: int, y: int) -> list[tuple[int, int]]:
    return [(x - 1, y), (x + 1, y), (x, y - 1), (x, y + 1)]


def neighbors8(x: int, y: int) -> list[tuple[int, int]]:
    points: list[tuple[int, int]] = []
    for dx in (-1, 0, 1):
        for dy in (-1, 0, 1):
            if dx == 0 and dy == 0:
                continue
            points.append((x + dx, y + dy))
    return points


def in_inner_canvas(x: int, y: int, w: int, h: int, margin: int = 1) -> bool:
    return margin <= x < (w - margin) and margin <= y < (h - margin)


def classify_target_multiplier(pixel_count: int) -> float:
    if pixel_count <= 10:
        return 2.8
    if pixel_count <= 16:
        return 2.5
    if pixel_count <= 32:
        return 2.1
    if pixel_count <= 64:
        return 1.8
    return 1.45


def recommended_max_pixels(w: int, h: int, current_pixels: int) -> int:
    longest = max(w, h)
    if longest <= 16:
        floor = 24
    elif longest <= 24:
        floor = 32
    elif longest <= 48:
        floor = 42
    else:
        floor = 72
    return max(floor, int(current_pixels * 1.15) + 4)


def build_chunky_shape(orig: set[tuple[int, int]], w: int, h: int, target_count: int) -> set[tuple[int, int]]:
    if not orig:
        return set()

    # 计算中心，扩张时让结构朝中心聚合，避免离散薄片。
    cx = sum(x for x, _ in orig) / len(orig)
    cy = sum(y for _, y in orig) / len(orig)

    shape = set(orig)
    frontier: set[tuple[int, int]] = set()
    for x, y in shape:
        for nx, ny in neighbors8(x, y):
            if (nx, ny) in shape:
                continue
            if in_inner_canvas(nx, ny, w, h):
                frontier.add((nx, ny))

    while len(shape) < target_count and frontier:
        best = None
        best_score = -1e18
        for px, py in frontier:
            adj = sum((nx, ny) in shape for nx, ny in neighbors8(px, py))
            # 越靠近中心越优先，避免扩张成细长丝。
            dist = math.hypot(px - cx, py - cy)
            score = adj * 10.0 - dist * 0.25
            if score > best_score:
                best_score = score
                best = (px, py)

        if best is None:
            break

        shape.add(best)
        frontier.remove(best)
        bx, by = best
        for nx, ny in neighbors8(bx, by):
            if (nx, ny) in shape:
                continue
            if in_inner_canvas(nx, ny, w, h):
                frontier.add((nx, ny))

    return shape


def build_color_roles(colors: list[str]) -> tuple[str, str, str]:
    if not colors:
        return "#666666", "#888888", "#bbbbbb"

    uniq = sorted(set(c.lower() for c in colors))
    rgb_list = [(c, hex_to_rgb(c)) for c in uniq]
    rgb_list.sort(key=lambda item: luminance(item[1]))

    dark = rgb_to_hex(rgb_list[0][1])
    bright = rgb_to_hex(rgb_list[-1][1])

    if len(rgb_list) >= 3:
        base = rgb_to_hex(rgb_list[len(rgb_list) // 2][1])
    else:
        base = bright
    return dark, base, bright


def build_pixels(
    orig_pixels: list[dict],
    shape: set[tuple[int, int]],
    dark: str,
    base: str,
    bright: str,
) -> list[dict]:
    orig_map: dict[tuple[int, int], str] = {}
    for p in orig_pixels:
        orig_map[(int(p["x"]), int(p["y"]))] = str(p.get("color", base)).lower()

    # 边界层（用于暗边）
    boundary: set[tuple[int, int]] = set()
    for x, y in shape:
        for nx, ny in neighbors4(x, y):
            if (nx, ny) not in shape:
                boundary.add((x, y))
                break

    # 生成像素：原始像素优先保留颜色，其次按边界/主体分配颜色。
    entries: list[dict] = []
    for x, y in sorted(shape, key=lambda pos: (pos[1], pos[0])):
        if (x, y) in orig_map:
            color = orig_map[(x, y)]
        elif (x, y) in boundary:
            color = dark
        else:
            color = base
        entries.append({"x": x, "y": y, "color": color})

    # 在内部补一点高亮，增强体积感。
    inner = [e for e in entries if (e["x"], e["y"]) not in boundary and (e["x"], e["y"]) not in orig_map]
    highlight_n = max(1, int(len(entries) * 0.08))
    inner.sort(key=lambda e: (e["x"] + e["y"]))  # 左上方向光照
    for e in inner[:highlight_n]:
        e["color"] = bright

    return entries


def transform_spec(spec: dict) -> tuple[dict, dict]:
    w, h = int(spec["size"][0]), int(spec["size"][1])
    src_pixels = spec.get("pixels", [])
    if not src_pixels:
        return spec, {"before": 0, "after": 0}

    before = len(src_pixels)
    mult = classify_target_multiplier(before)
    target_count = int(round(before * mult))
    target_count = max(target_count, before + 6)
    target_count = min(target_count, (w - 2) * (h - 2))

    orig = {(int(p["x"]), int(p["y"])) for p in src_pixels}
    shape = build_chunky_shape(orig, w, h, target_count)

    colors = [str(p.get("color", "")).lower() for p in src_pixels if p.get("color")]
    dark, base, bright = build_color_roles(colors)
    next_pixels = build_pixels(src_pixels, shape, dark=dark, base=base, bright=bright)

    next_spec = dict(spec)
    next_constraints = dict(spec.get("constraints", {}))
    next_constraints["allowAlpha"] = False
    next_constraints["maxPixels"] = max(
        int(next_constraints.get("maxPixels", 0) or 0),
        recommended_max_pixels(w, h, len(next_pixels)),
    )
    next_spec["constraints"] = next_constraints
    next_spec["pixels"] = next_pixels

    return next_spec, {
        "before": before,
        "after": len(next_pixels),
        "target": target_count,
        "maxPixels": next_constraints["maxPixels"],
    }


def main() -> int:
    parser = argparse.ArgumentParser(description="批量加厚 specs，提升像素特效可读性")
    parser.add_argument("spec_dir", help="spec 目录，例如 specs/")
    parser.add_argument("--dry-run", action="store_true", help="仅预览，不写回")
    args = parser.parse_args()

    spec_dir = Path(args.spec_dir).resolve()
    if not spec_dir.exists() or not spec_dir.is_dir():
        print(f"无效目录：{spec_dir}")
        return 1

    spec_files = sorted(spec_dir.glob("*.json"))
    if not spec_files:
        print(f"未找到 spec：{spec_dir}")
        return 1

    report: dict[str, dict] = {}
    total_before = 0
    total_after = 0

    for file in spec_files:
        data = json.loads(file.read_text(encoding="utf-8"))
        next_data, meta = transform_spec(data)
        report[file.name] = meta
        total_before += int(meta["before"])
        total_after += int(meta["after"])
        print(f"✓ {file.name}: {meta['before']} -> {meta['after']} (max={meta['maxPixels']})")
        if not args.dry_run:
            file.write_text(json.dumps(next_data, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")

    report_path = spec_dir.parent / "thicken-report.json"
    full_report = {
        "specCount": len(spec_files),
        "totalBeforePixels": total_before,
        "totalAfterPixels": total_after,
        "dryRun": args.dry_run,
        "files": report,
    }
    if not args.dry_run:
        report_path.write_text(json.dumps(full_report, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
        print(f"\n报告已写入：{report_path}")

    print(f"\n总像素数：{total_before} -> {total_after}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

