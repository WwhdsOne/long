#!/usr/bin/env python3
"""Generate pixel art JSON specs for filler talent nodes and tier rewards."""
import json, os

SPEC_DIR = "/Users/Learning/web/long/pixel-assets/specs"
os.makedirs(SPEC_DIR, exist_ok=True)

def write_spec(name, pixels, size=48, max_pixels=14, allow_alpha=False):
    spec = {
        "name": name,
        "size": [size, size],
        "constraints": {"maxPixels": max_pixels, "allowAlpha": allow_alpha},
        "pixels": pixels
    }
    path = os.path.join(SPEC_DIR, f"{name}.json")
    with open(path, 'w') as f:
        json.dump(spec, f, indent=2, ensure_ascii=False)
    print(f"  {name}.json ({len(pixels)}px)")

C = 24  # center of 48x48

# ===== Palette helpers =====
def G(bright):  # green shades
    return {"#4ade80": (0, 1), "#2bb873": (1, 0), "#166534": (2, 0), "#bbf7d0": (3, 0)}.get(bright, bright)

def K(bright):  # gold/amber shades
    return {"#fbbf24": (0, 1), "#c48a33": (1, 0), "#78350f": (2, 0), "#fde68a": (3, 0)}.get(bright, bright)

def R(bright):  # red shades
    return {"#ef4444": (0, 1), "#ca3e59": (1, 0), "#7f1d1d": (2, 0), "#f87171": (3, 0)}.get(bright, bright)

# ===== 均衡攻势 (normal) filler icons =====
print("Normal filler icons:")

# 锐锋 - upward arrow
write_spec("talent-normal-t1a", [
    {"x": C, "y": 12, "color": "#4ade80"},
    {"x": C-2, "y": 16, "color": "#4ade80"},
    {"x": C+2, "y": 16, "color": "#4ade80"},
    {"x": C-4, "y": 20, "color": "#2bb873"},
    {"x": C+4, "y": 20, "color": "#2bb873"},
    {"x": C-1, "y": 22, "color": "#2bb873"},
    {"x": C+1, "y": 22, "color": "#2bb873"},
    {"x": C, "y": 28, "color": "#166534"},
    {"x": C-2, "y": 30, "color": "#166534"},
    {"x": C+2, "y": 30, "color": "#166534"},
])

# 乱舞 - expanding rings (concentric arcs)
write_spec("talent-normal-t1b", [
    {"x": C-6, "y": C-6, "color": "#4ade80"},
    {"x": C+6, "y": C-6, "color": "#4ade80"},
    {"x": C-6, "y": C+6, "color": "#2bb873"},
    {"x": C+6, "y": C+6, "color": "#2bb873"},
    {"x": C-10, "y": C, "color": "#166534"},
    {"x": C+10, "y": C, "color": "#166534"},
    {"x": C, "y": C-10, "color": "#2bb873"},
    {"x": C, "y": C+10, "color": "#2bb873"},
    {"x": C-3, "y": C-3, "color": "#bbf7d0"},
    {"x": C+3, "y": C+3, "color": "#bbf7d0"},
])

# 追猎 - three diagonal lines
write_spec("talent-normal-t2a", [
    {"x": C-8, "y": C-6, "color": "#4ade80"},
    {"x": C-6, "y": C-4, "color": "#4ade80"},
    {"x": C-4, "y": C-2, "color": "#4ade80"},
    {"x": C-2, "y": C, "color": "#2bb873"},
    {"x": C, "y": C+2, "color": "#2bb873"},
    {"x": C+2, "y": C+4, "color": "#2bb873"},
    {"x": C+4, "y": C+6, "color": "#bbf7d0"},
    {"x": C+6, "y": C+8, "color": "#bbf7d0"},
    {"x": C-10, "y": C-2, "color": "#166534"},
    {"x": C-12, "y": C, "color": "#166534"},
    {"x": C+8, "y": C-2, "color": "#2bb873"},
    {"x": C+10, "y": C-4, "color": "#2bb873"},
])

# 穿刺 - concentric rings
write_spec("talent-normal-t2b", [
    {"x": C-4, "y": C-8, "color": "#4ade80"},
    {"x": C+4, "y": C-8, "color": "#4ade80"},
    {"x": C-4, "y": C+8, "color": "#2bb873"},
    {"x": C+4, "y": C+8, "color": "#2bb873"},
    {"x": C-8, "y": C-4, "color": "#166534"},
    {"x": C+8, "y": C-4, "color": "#166534"},
    {"x": C-8, "y": C+4, "color": "#166534"},
    {"x": C+8, "y": C+4, "color": "#166534"},
    {"x": C-2, "y": C-6, "color": "#bbf7d0"},
    {"x": C+2, "y": C-6, "color": "#bbf7d0"},
    {"x": C-2, "y": C+6, "color": "#bbf7d0"},
    {"x": C+2, "y": C+6, "color": "#bbf7d0"},
])

# 狩猎 - crown + sword
write_spec("talent-normal-t3a", [
    {"x": C-6, "y": 14, "color": "#4ade80"},
    {"x": C-3, "y": 12, "color": "#4ade80"},
    {"x": C+3, "y": 12, "color": "#4ade80"},
    {"x": C+6, "y": 14, "color": "#4ade80"},
    {"x": C, "y": 11, "color": "#bbf7d0"},
    {"x": C-8, "y": 18, "color": "#2bb873"},
    {"x": C+8, "y": 18, "color": "#2bb873"},
    {"x": C, "y": 22, "color": "#2bb873"},
    {"x": C, "y": 27, "color": "#166534"},
    {"x": C, "y": 32, "color": "#166534"},
    {"x": C-1, "y": 35, "color": "#2bb873"},
    {"x": C+1, "y": 35, "color": "#2bb873"},
])

# 铁腕 - fist
write_spec("talent-normal-t3b", [
    {"x": C-5, "y": 16, "color": "#4ade80"},
    {"x": C-3, "y": 14, "color": "#4ade80"},
    {"x": C+1, "y": 14, "color": "#4ade80"},
    {"x": C+4, "y": 16, "color": "#2bb873"},
    {"x": C-6, "y": 20, "color": "#2bb873"},
    {"x": C-3, "y": 20, "color": "#2bb873"},
    {"x": C+2, "y": 20, "color": "#2bb873"},
    {"x": C+5, "y": 20, "color": "#2bb873"},
    {"x": C-4, "y": 24, "color": "#166534"},
    {"x": C+2, "y": 24, "color": "#166534"},
    {"x": C-1, "y": 28, "color": "#166534"},
    {"x": C+3, "y": 28, "color": "#166534"},
    {"x": C, "y": 31, "color": "#2bb873"},
])

# ===== 碎盾攻坚 (armor) filler icons =====
print("\nArmor filler icons:")

# 破岩 - small hammer
write_spec("talent-armor-t1a", [
    {"x": C-6, "y": 14, "color": "#fbbf24"},
    {"x": C-4, "y": 14, "color": "#fbbf24"},
    {"x": C-2, "y": 14, "color": "#fbbf24"},
    {"x": C, "y": 14, "color": "#fbbf24"},
    {"x": C+2, "y": 14, "color": "#fbbf24"},
    {"x": C+4, "y": 14, "color": "#fbbf24"},
    {"x": C-1, "y": 18, "color": "#c48a33"},
    {"x": C+1, "y": 18, "color": "#c48a33"},
    {"x": C, "y": 22, "color": "#c48a33"},
    {"x": C, "y": 28, "color": "#78350f"},
    {"x": C, "y": 32, "color": "#78350f"},
    {"x": C-1, "y": 35, "color": "#c48a33"},
    {"x": C+1, "y": 35, "color": "#c48a33"},
])

# 凿裂 - arrow through plate
write_spec("talent-armor-t1b", [
    {"x": C-8, "y": C-4, "color": "#fbbf24"},
    {"x": C-6, "y": C-2, "color": "#fbbf24"},
    {"x": C-4, "y": C, "color": "#fbbf24"},
    {"x": C-2, "y": C+2, "color": "#c48a33"},
    {"x": C, "y": C+4, "color": "#c48a33"},
    {"x": C+2, "y": C+6, "color": "#78350f"},
    {"x": C-10, "y": C-2, "color": "#78350f"},
    {"x": C-8, "y": C+2, "color": "#c48a33"},
    {"x": C, "y": C-6, "color": "#c48a33"},
    {"x": C+4, "y": C-6, "color": "#fbbf24"},
    {"x": C-4, "y": C-6, "color": "#fbbf24"},
    {"x": C+4, "y": C+2, "color": "#fde68a"},
])

# 瓦解 - star flash
write_spec("talent-armor-t2a", [
    {"x": C, "y": 8, "color": "#fbbf24"},
    {"x": C, "y": 40, "color": "#c48a33"},
    {"x": 8, "y": C, "color": "#c48a33"},
    {"x": 40, "y": C, "color": "#fbbf24"},
    {"x": C-4, "y": 14, "color": "#fde68a"},
    {"x": C+4, "y": 14, "color": "#fde68a"},
    {"x": C-4, "y": 34, "color": "#78350f"},
    {"x": C+4, "y": 34, "color": "#78350f"},
    {"x": 14, "y": C-2, "color": "#fde68a"},
    {"x": 34, "y": C-2, "color": "#fde68a"},
    {"x": 14, "y": C+2, "color": "#78350f"},
    {"x": 34, "y": C+2, "color": "#78350f"},
])

# 碾碎 - stacked shields
write_spec("talent-armor-t2b", [
    {"x": C-6, "y": 16, "color": "#fbbf24"},
    {"x": C, "y": 16, "color": "#fbbf24"},
    {"x": C+6, "y": 16, "color": "#fbbf24"},
    {"x": C-8, "y": 20, "color": "#c48a33"},
    {"x": C-3, "y": 20, "color": "#c48a33"},
    {"x": C+3, "y": 20, "color": "#c48a33"},
    {"x": C+8, "y": 20, "color": "#c48a33"},
    {"x": C-4, "y": 24, "color": "#78350f"},
    {"x": C+4, "y": 24, "color": "#78350f"},
    {"x": C, "y": 24, "color": "#78350f"},
    {"x": C-1, "y": 28, "color": "#c48a33"},
    {"x": C+1, "y": 28, "color": "#c48a33"},
    {"x": C, "y": 31, "color": "#c48a33"},
])

# 碎颅 - flag + sword
write_spec("talent-armor-t3a", [
    {"x": C+2, "y": 12, "color": "#fbbf24"},
    {"x": C-2, "y": 14, "color": "#fbbf24"},
    {"x": C+4, "y": 15, "color": "#fbbf24"},
    {"x": C-4, "y": 18, "color": "#c48a33"},
    {"x": C+2, "y": 19, "color": "#c48a33"},
    {"x": C-2, "y": 22, "color": "#c48a33"},
    {"x": C, "y": 26, "color": "#78350f"},
    {"x": C, "y": 30, "color": "#78350f"},
    {"x": C, "y": 34, "color": "#c48a33"},
    {"x": C-1, "y": 37, "color": "#c48a33"},
    {"x": C+1, "y": 37, "color": "#c48a33"},
    {"x": C+6, "y": 16, "color": "#c48a33"},
    {"x": C+8, "y": 18, "color": "#78350f"},
])

# 摧坚 - crossed axes
write_spec("talent-armor-t3b", [
    {"x": C-8, "y": 16, "color": "#fbbf24"},
    {"x": C-6, "y": 18, "color": "#c48a33"},
    {"x": C-4, "y": 20, "color": "#c48a33"},
    {"x": C, "y": C, "color": "#fde68a"},
    {"x": C+2, "y": C+2, "color": "#c48a33"},
    {"x": C+4, "y": 20, "color": "#c48a33"},
    {"x": C+6, "y": 16, "color": "#fbbf24"},
    {"x": C+4, "y": 28, "color": "#78350f"},
    {"x": C+2, "y": 26, "color": "#78350f"},
    {"x": C-4, "y": 28, "color": "#78350f"},
    {"x": C-6, "y": 32, "color": "#c48a33"},
    {"x": C+6, "y": 32, "color": "#c48a33"},
])

# ===== 致命洞察 (crit) filler icons =====
print("\nCrit filler icons:")

# 锐眼 - dagger diagonal
write_spec("talent-crit-t1a", [
    {"x": C+2, "y": 10, "color": "#ef4444"},
    {"x": C+4, "y": 13, "color": "#ef4444"},
    {"x": C+2, "y": 16, "color": "#ca3e59"},
    {"x": C, "y": 19, "color": "#ca3e59"},
    {"x": C-2, "y": 22, "color": "#ca3e59"},
    {"x": C-4, "y": 25, "color": "#7f1d1d"},
    {"x": C-6, "y": 28, "color": "#7f1d1d"},
    {"x": C-8, "y": 31, "color": "#7f1d1d"},
    {"x": C+3, "y": 14, "color": "#f87171"},
    {"x": C+1, "y": 17, "color": "#f87171"},
    {"x": C-1, "y": 21, "color": "#f87171"},
])

# 残酷 - red eye
write_spec("talent-crit-t1b", [
    {"x": C-8, "y": C-4, "color": "#ef4444"},
    {"x": C-6, "y": C-6, "color": "#ca3e59"},
    {"x": C-2, "y": C-6, "color": "#ef4444"},
    {"x": C+2, "y": C-6, "color": "#ef4444"},
    {"x": C+6, "y": C-4, "color": "#ca3e59"},
    {"x": C-2, "y": C, "color": "#7f1d1d"},
    {"x": C+2, "y": C, "color": "#7f1d1d"},
    {"x": C, "y": C-3, "color": "#7f1d1d"},
    {"x": C-4, "y": C+4, "color": "#ca3e59"},
    {"x": C+4, "y": C+4, "color": "#ca3e59"},
    {"x": C, "y": C+6, "color": "#ef4444"},
])

# 深创 - sparks
write_spec("talent-crit-t2a", [
    {"x": C, "y": 10, "color": "#ef4444"},
    {"x": C-6, "y": 16, "color": "#ef4444"},
    {"x": C+6, "y": 16, "color": "#ef4444"},
    {"x": C-4, "y": C, "color": "#ca3e59"},
    {"x": C+4, "y": C, "color": "#ca3e59"},
    {"x": C, "y": C, "color": "#f87171"},
    {"x": C-8, "y": C+4, "color": "#7f1d1d"},
    {"x": C+8, "y": C+4, "color": "#7f1d1d"},
    {"x": C-2, "y": C-6, "color": "#f87171"},
    {"x": C+2, "y": C-6, "color": "#f87171"},
    {"x": C-3, "y": 32, "color": "#7f1d1d"},
    {"x": C+3, "y": 32, "color": "#7f1d1d"},
])

# 喋血 - serrated blade
write_spec("talent-crit-t2b", [
    {"x": C-4, "y": 14, "color": "#ef4444"},
    {"x": C-2, "y": 14, "color": "#ef4444"},
    {"x": C, "y": 14, "color": "#ca3e59"},
    {"x": C-5, "y": 18, "color": "#ca3e59"},
    {"x": C-1, "y": 18, "color": "#ca3e59"},
    {"x": C-3, "y": 22, "color": "#ca3e59"},
    {"x": C+1, "y": 22, "color": "#7f1d1d"},
    {"x": C-1, "y": 26, "color": "#7f1d1d"},
    {"x": C+3, "y": 26, "color": "#7f1d1d"},
    {"x": C, "y": 30, "color": "#ca3e59"},
    {"x": C-2, "y": 30, "color": "#ca3e59"},
    {"x": C, "y": 34, "color": "#7f1d1d"},
])

# 追魂 - skull mark
write_spec("talent-crit-t3a", [
    {"x": C-5, "y": 16, "color": "#ef4444"},
    {"x": C-2, "y": 16, "color": "#ef4444"},
    {"x": C+2, "y": 16, "color": "#ef4444"},
    {"x": C+5, "y": 16, "color": "#ef4444"},
    {"x": C-6, "y": 20, "color": "#ca3e59"},
    {"x": C+6, "y": 20, "color": "#ca3e59"},
    {"x": C-4, "y": 20, "color": "#7f1d1d"},
    {"x": C+4, "y": 20, "color": "#7f1d1d"},
    {"x": C-2, "y": 24, "color": "#ca3e59"},
    {"x": C+2, "y": 24, "color": "#ca3e59"},
    {"x": C-1, "y": 28, "color": "#7f1d1d"},
    {"x": C+1, "y": 28, "color": "#7f1d1d"},
    {"x": C, "y": 30, "color": "#ef4444"},
    {"x": C, "y": 33, "color": "#7f1d1d"},
])

# 暴虐 - crossed daggers
write_spec("talent-crit-t3b", [
    {"x": C-6, "y": 14, "color": "#ef4444"},
    {"x": C-4, "y": 16, "color": "#ca3e59"},
    {"x": C-2, "y": 18, "color": "#ca3e59"},
    {"x": C, "y": C, "color": "#f87171"},
    {"x": C+2, "y": 22, "color": "#ca3e59"},
    {"x": C+4, "y": 24, "color": "#7f1d1d"},
    {"x": C+6, "y": 26, "color": "#7f1d1d"},
    {"x": C+6, "y": 14, "color": "#ef4444"},
    {"x": C-6, "y": 26, "color": "#7f1d1d"},
    {"x": C-4, "y": 28, "color": "#7f1d1d"},
    {"x": C+4, "y": 16, "color": "#ca3e59"},
    {"x": C-2, "y": 32, "color": "#ca3e59"},
    {"x": C+2, "y": 32, "color": "#ca3e59"},
])

# ===== Tier completion reward icons (tree-colored variations) =====
print("\nTier reward icons:")

# T0: Star/awakening
for tree, color, hi in [("normal","#2bb873","#4ade80"), ("armor","#c48a33","#fbbf24"), ("crit","#ca3e59","#ef4444")]:
    write_spec(f"tier-reward-T0-{tree}", [
        {"x": C, "y": 10, "color": hi},
        {"x": C-6, "y": 16, "color": hi},
        {"x": C+6, "y": 16, "color": hi},
        {"x": C-8, "y": 22, "color": color},
        {"x": C-3, "y": 22, "color": color},
        {"x": C+3, "y": 22, "color": color},
        {"x": C+8, "y": 22, "color": color},
        {"x": C-10, "y": 26, "color": hi},
        {"x": C+10, "y": 26, "color": hi},
        {"x": C-6, "y": 30, "color": color},
        {"x": C, "y": 30, "color": color},
        {"x": C+6, "y": 30, "color": color},
    ])

# T1: Upgrade arrow
for tree, color, hi in [("normal","#2bb873","#4ade80"), ("armor","#c48a33","#fbbf24"), ("crit","#ca3e59","#ef4444")]:
    write_spec(f"tier-reward-T1-{tree}", [
        {"x": C, "y": 8, "color": hi},
        {"x": C-4, "y": 14, "color": hi},
        {"x": C+4, "y": 14, "color": hi},
        {"x": C-6, "y": 18, "color": color},
        {"x": C+6, "y": 18, "color": color},
        {"x": C-2, "y": 20, "color": color},
        {"x": C+2, "y": 20, "color": color},
        {"x": C-4, "y": 26, "color": color},
        {"x": C+4, "y": 26, "color": color},
        {"x": C, "y": 32, "color": color},
        {"x": C-6, "y": 34, "color": hi},
        {"x": C+6, "y": 34, "color": hi},
    ])

# T2: Flame emblem
for tree, color, hi in [("normal","#2bb873","#4ade80"), ("armor","#c48a33","#fbbf24"), ("crit","#ca3e59","#ef4444")]:
    write_spec(f"tier-reward-T2-{tree}", [
        {"x": C, "y": 8, "color": hi},
        {"x": C-3, "y": 12, "color": hi},
        {"x": C+3, "y": 12, "color": hi},
        {"x": C-6, "y": 16, "color": color},
        {"x": C, "y": 16, "color": color},
        {"x": C+6, "y": 16, "color": color},
        {"x": C-4, "y": 22, "color": color},
        {"x": C+4, "y": 22, "color": color},
        {"x": C-2, "y": 28, "color": hi},
        {"x": C+2, "y": 28, "color": hi},
        {"x": C, "y": 34, "color": color},
        {"x": C-3, "y": 38, "color": color},
        {"x": C+3, "y": 38, "color": color},
    ])

# T3: Crown
for tree, color, hi in [("normal","#2bb873","#4ade80"), ("armor","#c48a33","#fbbf24"), ("crit","#ca3e59","#ef4444")]:
    write_spec(f"tier-reward-T3-{tree}", [
        {"x": C-6, "y": 14, "color": hi},
        {"x": C-3, "y": 12, "color": hi},
        {"x": C, "y": 11, "color": hi},
        {"x": C+3, "y": 12, "color": hi},
        {"x": C+6, "y": 14, "color": hi},
        {"x": C-8, "y": 18, "color": color},
        {"x": C+8, "y": 18, "color": color},
        {"x": C-5, "y": 20, "color": color},
        {"x": C+5, "y": 20, "color": color},
        {"x": C-2, "y": 22, "color": color},
        {"x": C+2, "y": 22, "color": color},
        {"x": C, "y": 26, "color": color},
        {"x": C-1, "y": 30, "color": hi},
        {"x": C+1, "y": 30, "color": hi},
    ])

# T4: Aura/sun disk
for tree, color, hi in [("normal","#2bb873","#4ade80"), ("armor","#c48a33","#fbbf24"), ("crit","#ca3e59","#ef4444")]:
    write_spec(f"tier-reward-T4-{tree}", [
        {"x": C, "y": 10, "color": hi},
        {"x": C-6, "y": 14, "color": hi},
        {"x": C+6, "y": 14, "color": hi},
        {"x": C-10, "y": C, "color": color},
        {"x": C+10, "y": C, "color": color},
        {"x": C-8, "y": C+4, "color": color},
        {"x": C+8, "y": C+4, "color": color},
        {"x": C-4, "y": C+6, "color": hi},
        {"x": C+4, "y": C+6, "color": hi},
        {"x": C, "y": C+2, "color": color},
        {"x": C-12, "y": C-3, "color": color},
        {"x": C+12, "y": C-3, "color": color},
        {"x": C-2, "y": 34, "color": color},
        {"x": C+2, "y": 34, "color": color},
    ])

# ===== Emblems =====
print("\nEmblems:")

# normal emblem - leaf + whirlwind
write_spec("emblem-normal", [
    {"x": C, "y": 8, "color": "#4ade80"},
    {"x": C-8, "y": 16, "color": "#4ade80"},
    {"x": C+8, "y": 16, "color": "#4ade80"},
    {"x": C-12, "y": C, "color": "#2bb873"},
    {"x": C+12, "y": C, "color": "#2bb873"},
    {"x": C-10, "y": C+4, "color": "#2bb873"},
    {"x": C+10, "y": C+4, "color": "#2bb873"},
    {"x": C-6, "y": C+6, "color": "#166534"},
    {"x": C+6, "y": C+6, "color": "#166534"},
    {"x": C-2, "y": 28, "color": "#166534"},
    {"x": C+2, "y": 28, "color": "#166534"},
    {"x": C, "y": 32, "color": "#4ade80"},
    {"x": C-4, "y": C-2, "color": "#bbf7d0"},
    {"x": C+4, "y": C-2, "color": "#bbf7d0"},
])

# armor emblem - shield + crossed hammers
write_spec("emblem-armor", [
    {"x": C-8, "y": 12, "color": "#fbbf24"},
    {"x": C+8, "y": 12, "color": "#fbbf24"},
    {"x": C-6, "y": 16, "color": "#fbbf24"},
    {"x": C+6, "y": 16, "color": "#fbbf24"},
    {"x": C-3, "y": 16, "color": "#c48a33"},
    {"x": C+3, "y": 16, "color": "#c48a33"},
    {"x": C-10, "y": 20, "color": "#c48a33"},
    {"x": C+10, "y": 20, "color": "#c48a33"},
    {"x": C-6, "y": 24, "color": "#c48a33"},
    {"x": C+6, "y": 24, "color": "#c48a33"},
    {"x": C-2, "y": 28, "color": "#78350f"},
    {"x": C+2, "y": 28, "color": "#78350f"},
    {"x": C, "y": 32, "color": "#78350f"},
    {"x": C, "y": 20, "color": "#fde68a"},
])

# crit emblem - dagger + blood drop
write_spec("emblem-crit", [
    {"x": C, "y": 8, "color": "#ef4444"},
    {"x": C-2, "y": 12, "color": "#ef4444"},
    {"x": C+2, "y": 12, "color": "#ef4444"},
    {"x": C-4, "y": 16, "color": "#ca3e59"},
    {"x": C, "y": 16, "color": "#ca3e59"},
    {"x": C+4, "y": 16, "color": "#ca3e59"},
    {"x": C-6, "y": 22, "color": "#ca3e59"},
    {"x": C, "y": 22, "color": "#f87171"},
    {"x": C+6, "y": 22, "color": "#ca3e59"},
    {"x": C-2, "y": 28, "color": "#7f1d1d"},
    {"x": C+2, "y": 28, "color": "#7f1d1d"},
    {"x": C, "y": 34, "color": "#7f1d1d"},
    {"x": C-3, "y": 36, "color": "#ef4444"},
    {"x": C+3, "y": 36, "color": "#ef4444"},
])

# talent system logo - three converging triangles
write_spec("emblem-talent", [
    {"x": C, "y": 6, "color": "#4ade80"},
    {"x": C-10, "y": 22, "color": "#c48a33"},
    {"x": C+10, "y": 22, "color": "#ef4444"},
    {"x": C-5, "y": 14, "color": "#2bb873"},
    {"x": C+5, "y": 14, "color": "#ca3e59"},
    {"x": C-8, "y": 18, "color": "#fbbf24"},
    {"x": C+8, "y": 18, "color": "#7f1d1d"},
    {"x": C-3, "y": 18, "color": "#166534"},
    {"x": C+3, "y": 18, "color": "#f87171"},
    {"x": C, "y": C, "color": "#fde68a"},
    {"x": C-4, "y": 26, "color": "#78350f"},
    {"x": C+4, "y": 26, "color": "#7f1d1d"},
    {"x": C-2, "y": 30, "color": "#c48a33"},
    {"x": C+2, "y": 30, "color": "#ef4444"},
], max_pixels=20)

print(f"\nDone! {len(os.listdir(SPEC_DIR))} specs in {SPEC_DIR}")
