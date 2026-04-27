#!/usr/bin/env python3
"""Generate pixel art JSON specs for 30 main talent nodes."""
import json, os

SPEC_DIR = "/Users/Learning/web/long/pixel-assets/specs"
os.makedirs(SPEC_DIR, exist_ok=True)

def write_spec(name, pixels, size=48, max_pixels=14):
    spec = {
        "name": name,
        "size": [size, size],
        "constraints": {"maxPixels": max_pixels, "allowAlpha": False},
        "pixels": pixels
    }
    path = os.path.join(SPEC_DIR, f"{name}.json")
    with open(path, 'w') as f:
        json.dump(spec, f, indent=2, ensure_ascii=False)
    print(f"  {name}.json ({len(pixels)}px)")

C = 24  # center of 48x48

# Color palettes
GN = "#2bb873"  # normal green
GH = "#4ade80"  # normal highlight
GD = "#166534"  # normal dark
GW = "#bbf7d0"  # normal white highlight
AU = "#c48a33"  # armor gold
AH = "#fbbf24"  # armor highlight
AD = "#78350f"  # armor dark
AW = "#fde68a"  # armor white
CR = "#ca3e59"  # crit red
CH = "#ef4444"  # crit highlight
CD = "#7f1d1d"  # crit dark
CW = "#f87171"  # crit light

# ========================================================
# 均衡攻势 (normal) - 10 nodes
# ========================================================
print("=== 均衡攻势 ===")

# normal_core - 暴风连击: multiple slashes converging
write_spec("talent-normal_core", [
    {"x": C-8, "y": C-4, "color": GH}, {"x": C-6, "y": C-2, "color": GH},
    {"x": C-4, "y": C, "color": GW}, {"x": C-2, "y": C+2, "color": GN},
    {"x": C, "y": C, "color": GW},
    {"x": C+8, "y": C-4, "color": GH}, {"x": C+6, "y": C-2, "color": GH},
    {"x": C-8, "y": C+4, "color": GD}, {"x": C+8, "y": C+4, "color": GD},
    {"x": C-4, "y": C-8, "color": GN}, {"x": C+4, "y": C-8, "color": GN},
    {"x": C, "y": C-10, "color": GW},
])

# normal_atk_up - 攻击强化: upward sword with glow
write_spec("talent-normal_atk_up", [
    {"x": C, "y": 8, "color": GW}, {"x": C-1, "y": 11, "color": GH},
    {"x": C+1, "y": 11, "color": GH}, {"x": C-2, "y": 15, "color": GN},
    {"x": C+2, "y": 15, "color": GN}, {"x": C-3, "y": 19, "color": GN},
    {"x": C+3, "y": 19, "color": GN}, {"x": C-4, "y": 23, "color": GD},
    {"x": C+4, "y": 23, "color": GD}, {"x": C-5, "y": 26, "color": GD},
    {"x": C+5, "y": 26, "color": GD}, {"x": C, "y": 30, "color": GD},
    {"x": C-6, "y": 32, "color": GN}, {"x": C+6, "y": 32, "color": GN},
])

# normal_dmg_amp - 伤害增幅: explosion burst
write_spec("talent-normal_dmg_amp", [
    {"x": C, "y": 10, "color": GW}, {"x": C-4, "y": 14, "color": GH},
    {"x": C+4, "y": 14, "color": GH}, {"x": C-8, "y": C, "color": GN},
    {"x": C+8, "y": C, "color": GN}, {"x": C-6, "y": C+2, "color": GN},
    {"x": C+6, "y": C+2, "color": GN}, {"x": C, "y": C-2, "color": GW},
    {"x": C-10, "y": 18, "color": GD}, {"x": C+10, "y": 18, "color": GD},
    {"x": C-2, "y": 30, "color": GD}, {"x": C+2, "y": 30, "color": GD},
    {"x": C-6, "y": 28, "color": GN}, {"x": C+6, "y": 28, "color": GN},
])

# normal_soft_atk - 软组织特攻: target reticle on soft spot
write_spec("talent-normal_soft_atk", [
    {"x": C, "y": 12, "color": GW}, {"x": C-3, "y": 16, "color": GH},
    {"x": C+3, "y": 16, "color": GH}, {"x": C-6, "y": C, "color": GN},
    {"x": C+6, "y": C, "color": GN}, {"x": C-4, "y": C+6, "color": GN},
    {"x": C+4, "y": C+6, "color": GN}, {"x": C, "y": 32, "color": GD},
    {"x": C-2, "y": 36, "color": GD}, {"x": C+2, "y": 36, "color": GD},
    {"x": C-8, "y": 14, "color": GD}, {"x": C+8, "y": 14, "color": GD},
    {"x": C, "y": 30, "color": GW},
])

# normal_charge - 蓄力返还: ring returning inward
write_spec("talent-normal_charge", [
    {"x": C-8, "y": C-8, "color": GH}, {"x": C+8, "y": C-8, "color": GH},
    {"x": C+8, "y": C+8, "color": GN}, {"x": C-8, "y": C+8, "color": GN},
    {"x": C-4, "y": C-4, "color": GW}, {"x": C+4, "y": C-4, "color": GW},
    {"x": C+4, "y": C+4, "color": GH}, {"x": C-4, "y": C+4, "color": GH},
    {"x": C, "y": C-6, "color": GD}, {"x": C, "y": C+6, "color": GD},
    {"x": C-6, "y": C, "color": GD}, {"x": C+6, "y": C, "color": GD},
    {"x": C-10, "y": C-6, "color": GN}, {"x": C+10, "y": C+6, "color": GN},
])

# normal_chase_up - 追击强化: pursuit pounce arrow
write_spec("talent-normal_chase_up", [
    {"x": C+6, "y": 8, "color": GW}, {"x": C+4, "y": 12, "color": GH},
    {"x": C+2, "y": 16, "color": GH}, {"x": C, "y": C, "color": GW},
    {"x": C-2, "y": 24, "color": GN}, {"x": C-4, "y": 28, "color": GN},
    {"x": C-6, "y": 32, "color": GD},
    {"x": C+4, "y": 10, "color": GH}, {"x": C+2, "y": 14, "color": GH},
    {"x": C-2, "y": 22, "color": GN}, {"x": C-4, "y": 26, "color": GN},
    {"x": C-6, "y": 30, "color": GD}, {"x": C-8, "y": 34, "color": GD},
])

# normal_combo_ext - 连击扩展: chain extending
write_spec("talent-normal_combo_ext", [
    {"x": C-10, "y": C, "color": GN}, {"x": C-6, "y": C-3, "color": GH},
    {"x": C-2, "y": C+3, "color": GN}, {"x": C+2, "y": C-3, "color": GH},
    {"x": C+6, "y": C+3, "color": GN}, {"x": C+10, "y": C, "color": GD},
    {"x": C-10, "y": C-3, "color": GD}, {"x": C-6, "y": C, "color": GN},
    {"x": C-2, "y": C, "color": GW}, {"x": C+2, "y": C, "color": GW},
    {"x": C+6, "y": C, "color": GN}, {"x": C+10, "y": C-3, "color": GD},
    {"x": C-2, "y": C+6, "color": GD}, {"x": C+2, "y": C+6, "color": GD},
])

# normal_encircle - 围剿: encircling ring
write_spec("talent-normal_encircle", [
    {"x": C-6, "y": C-10, "color": GH}, {"x": C+6, "y": C-10, "color": GH},
    {"x": C-10, "y": C-6, "color": GN}, {"x": C+10, "y": C-6, "color": GN},
    {"x": C-10, "y": C+6, "color": GN}, {"x": C+10, "y": C+6, "color": GN},
    {"x": C-6, "y": C+10, "color": GD}, {"x": C+6, "y": C+10, "color": GD},
    {"x": C, "y": C-8, "color": GW}, {"x": C, "y": C+8, "color": GD},
    {"x": C-8, "y": C, "color": GN}, {"x": C+8, "y": C, "color": GN},
    {"x": C, "y": C, "color": GW},
])

# normal_low_hp - 残血收割: dripping blade
write_spec("talent-normal_low_hp", [
    {"x": C, "y": 8, "color": GW}, {"x": C-1, "y": 12, "color": GH},
    {"x": C+1, "y": 12, "color": GH}, {"x": C-2, "y": 16, "color": GN},
    {"x": C+2, "y": 16, "color": GN}, {"x": C-1, "y": 20, "color": GN},
    {"x": C+1, "y": 20, "color": GN}, {"x": C, "y": 24, "color": GD},
    {"x": C, "y": 28, "color": GD}, {"x": C, "y": 32, "color": GD},
    {"x": C-2, "y": 34, "color": GN}, {"x": C+2, "y": 34, "color": GN},
])

# normal_ultimate - 白银风暴: silver star burst
write_spec("talent-normal_ultimate", [
    {"x": C, "y": 6, "color": GW}, {"x": C-4, "y": 10, "color": GW},
    {"x": C+4, "y": 10, "color": GW}, {"x": C-8, "y": 16, "color": GH},
    {"x": C+8, "y": 16, "color": GH}, {"x": C-10, "y": C, "color": GN},
    {"x": C+10, "y": C, "color": GN}, {"x": C-8, "y": 32, "color": GN},
    {"x": C+8, "y": 32, "color": GN}, {"x": C-4, "y": 38, "color": GD},
    {"x": C+4, "y": 38, "color": GD}, {"x": C, "y": 42, "color": GD},
    {"x": C, "y": C, "color": GW},
])

# ========================================================
# 碎盾攻坚 (armor) - 10 nodes
# ========================================================
print("=== 碎盾攻坚 ===")

# armor_core - 灭绝穿甲: piercing arrow through wall
write_spec("talent-armor_core", [
    {"x": C-8, "y": C+6, "color": AU}, {"x": C-6, "y": C+4, "color": AU},
    {"x": C-4, "y": C+2, "color": AH}, {"x": C-2, "y": C, "color": AH},
    {"x": C, "y": C-2, "color": AW}, {"x": C+2, "y": C-4, "color": AH},
    {"x": C+4, "y": C-6, "color": AU}, {"x": C+8, "y": C-4, "color": AD},
    {"x": C+6, "y": C-2, "color": AU},
    {"x": C-10, "y": C+2, "color": AD}, {"x": C-8, "y": C, "color": AD},
    {"x": C+2, "y": C+4, "color": AD}, {"x": C+4, "y": C+6, "color": AD},
])

# armor_pen_up - 穿甲强化: reinforced arrow
write_spec("talent-armor_pen_up", [
    {"x": C-4, "y": 14, "color": AH}, {"x": C-2, "y": 18, "color": AH},
    {"x": C, "y": 22, "color": AW}, {"x": C+2, "y": 26, "color": AU},
    {"x": C+4, "y": 30, "color": AD}, {"x": C+6, "y": 34, "color": AD},
    {"x": C-6, "y": 16, "color": AH}, {"x": C-8, "y": 20, "color": AU},
    {"x": C+2, "y": 12, "color": AH}, {"x": C, "y": 16, "color": AU},
    {"x": C-2, "y": 28, "color": AU}, {"x": C-4, "y": 32, "color": AD},
])

# armor_boss_hunter - 首领猎杀: crosshair scope
write_spec("talent-armor_boss_hunter", [
    {"x": C, "y": 8, "color": AH}, {"x": C, "y": 40, "color": AU},
    {"x": 8, "y": C, "color": AU}, {"x": 40, "y": C, "color": AH},
    {"x": C-4, "y": C-4, "color": AW}, {"x": C+4, "y": C-4, "color": AW},
    {"x": C-4, "y": C+4, "color": AU}, {"x": C+4, "y": C+4, "color": AU},
    {"x": C-6, "y": C, "color": AD}, {"x": C+6, "y": C, "color": AD},
    {"x": C, "y": C-6, "color": AW}, {"x": C, "y": C+6, "color": AD},
    {"x": C-8, "y": 12, "color": AU}, {"x": C+8, "y": 12, "color": AU},
])

# armor_heavy_scale - 以强制强: balance scales
write_spec("talent-armor_heavy_scale", [
    {"x": C, "y": 8, "color": AH}, {"x": C-8, "y": 20, "color": AU},
    {"x": C+8, "y": 20, "color": AU}, {"x": C-10, "y": 26, "color": AD},
    {"x": C+10, "y": 26, "color": AD}, {"x": C-6, "y": 28, "color": AH},
    {"x": C+6, "y": 28, "color": AH}, {"x": C-4, "y": 32, "color": AD},
    {"x": C+4, "y": 32, "color": AD}, {"x": C, "y": 14, "color": AW},
    {"x": C-2, "y": 22, "color": AU}, {"x": C+2, "y": 22, "color": AU},
    {"x": C, "y": 36, "color": AD},
])

# armor_heavy_atk - 重甲特攻: heavy hammer strike
write_spec("talent-armor_heavy_atk", [
    {"x": C-6, "y": 12, "color": AH}, {"x": C-4, "y": 12, "color": AH},
    {"x": C-2, "y": 12, "color": AH}, {"x": C, "y": 12, "color": AH},
    {"x": C+2, "y": 12, "color": AH}, {"x": C+4, "y": 12, "color": AH},
    {"x": C+6, "y": 12, "color": AH}, {"x": C-1, "y": 16, "color": AU},
    {"x": C+1, "y": 16, "color": AU}, {"x": C, "y": 22, "color": AU},
    {"x": C, "y": 28, "color": AD}, {"x": C, "y": 34, "color": AD},
    {"x": C-2, "y": 38, "color": AU}, {"x": C+2, "y": 38, "color": AU},
])

# armor_collapse_ext - 崩塌延长: crack extending
write_spec("talent-armor_collapse_ext", [
    {"x": C, "y": 8, "color": AH}, {"x": C-1, "y": 13, "color": AU},
    {"x": C+2, "y": 17, "color": AU}, {"x": C-3, "y": 20, "color": AH},
    {"x": C+1, "y": 23, "color": AD}, {"x": C-2, "y": 27, "color": AD},
    {"x": C+3, "y": 30, "color": AD}, {"x": C, "y": 34, "color": AD},
    {"x": C-4, "y": 16, "color": AH}, {"x": C+2, "y": 14, "color": AW},
    {"x": C-2, "y": 36, "color": AU}, {"x": C+1, "y": 38, "color": AU},
    {"x": C-4, "y": 10, "color": AW},
])

# armor_auto_strike - 自动打击: gear/cog
write_spec("talent-armor_auto_strike", [
    {"x": C, "y": 6, "color": AH}, {"x": C-6, "y": 14, "color": AU},
    {"x": C+6, "y": 14, "color": AU}, {"x": C-8, "y": C, "color": AH},
    {"x": C+8, "y": C, "color": AH}, {"x": C-6, "y": 34, "color": AU},
    {"x": C+6, "y": 34, "color": AU}, {"x": C, "y": 42, "color": AD},
    {"x": C-3, "y": C-6, "color": AW}, {"x": C+3, "y": C+6, "color": AD},
    {"x": C-8, "y": 18, "color": AD}, {"x": C+8, "y": 30, "color": AD},
    {"x": C-4, "y": C, "color": AU}, {"x": C+4, "y": C, "color": AU},
])

# armor_ruin - 废墟打击: collapsed structure
write_spec("talent-armor_ruin", [
    {"x": C-4, "y": 12, "color": AH}, {"x": C+2, "y": 12, "color": AH},
    {"x": C-6, "y": 18, "color": AU}, {"x": C+4, "y": 18, "color": AU},
    {"x": C-8, "y": 24, "color": AD}, {"x": C+6, "y": 24, "color": AD},
    {"x": C-5, "y": 28, "color": AD}, {"x": C, "y": 28, "color": AD},
    {"x": C+3, "y": 30, "color": AU}, {"x": C-3, "y": 34, "color": AU},
    {"x": C+2, "y": 36, "color": AD}, {"x": C-8, "y": 16, "color": AH},
    {"x": C+6, "y": 16, "color": AW},
])

# armor_pen_convert - 破甲转化: transform/conversion arrow
write_spec("talent-armor_pen_convert", [
    {"x": C-8, "y": C, "color": AU}, {"x": C-6, "y": C-2, "color": AU},
    {"x": C-4, "y": C-4, "color": AH}, {"x": C-2, "y": C-6, "color": AH},
    {"x": C+2, "y": C+6, "color": AH}, {"x": C+4, "y": C+4, "color": AH},
    {"x": C+6, "y": C+2, "color": AU}, {"x": C+8, "y": C, "color": AU},
    {"x": C-10, "y": C+2, "color": AD}, {"x": C+10, "y": C+2, "color": AD},
    {"x": C, "y": C, "color": AW},
    {"x": C-3, "y": C+6, "color": AD}, {"x": C+3, "y": C-6, "color": AD},
    {"x": C-5, "y": C+8, "color": AU}, {"x": C+5, "y": C-8, "color": AU},
], max_pixels=15)

# armor_ultimate - 审判日: gavel/judgment hammer
write_spec("talent-armor_ultimate", [
    {"x": C-6, "y": 8, "color": AH}, {"x": C-3, "y": 8, "color": AH},
    {"x": C+3, "y": 8, "color": AH}, {"x": C+6, "y": 8, "color": AH},
    {"x": C, "y": 12, "color": AW}, {"x": C, "y": 16, "color": AU},
    {"x": C, "y": 22, "color": AU}, {"x": C, "y": 28, "color": AD},
    {"x": C, "y": 34, "color": AD}, {"x": C-3, "y": 38, "color": AU},
    {"x": C+3, "y": 38, "color": AU},
    {"x": C-8, "y": 14, "color": AH}, {"x": C+8, "y": 14, "color": AH},
    {"x": C-2, "y": 20, "color": AW}, {"x": C+2, "y": 20, "color": AW},
], max_pixels=15)

# ========================================================
# 致命洞察 (crit) - 10 nodes
# ========================================================
print("=== 致命洞察 ===")

# crit_core - 溢杀: overflowing blood vessel
write_spec("talent-crit_core", [
    {"x": C, "y": 10, "color": CH}, {"x": C-3, "y": 14, "color": CH},
    {"x": C+3, "y": 14, "color": CH}, {"x": C-5, "y": 18, "color": CR},
    {"x": C+5, "y": 18, "color": CR}, {"x": C-4, "y": 22, "color": CR},
    {"x": C+4, "y": 22, "color": CR}, {"x": C, "y": 22, "color": CW},
    {"x": C-2, "y": 28, "color": CD}, {"x": C+2, "y": 28, "color": CD},
    {"x": C-4, "y": 32, "color": CR}, {"x": C+4, "y": 32, "color": CR},
    {"x": C, "y": 36, "color": CD},
])

# crit_omen_resonate - 死兆共鸣: resonating concentric rings
write_spec("talent-crit_omen_resonate", [
    {"x": C-6, "y": C-6, "color": CH}, {"x": C+6, "y": C-6, "color": CH},
    {"x": C-6, "y": C+6, "color": CR}, {"x": C+6, "y": C+6, "color": CR},
    {"x": C-10, "y": C-8, "color": CD}, {"x": C+10, "y": C-8, "color": CD},
    {"x": C-10, "y": C+8, "color": CR}, {"x": C+10, "y": C+8, "color": CR},
    {"x": C-3, "y": C-3, "color": CW}, {"x": C+3, "y": C-3, "color": CW},
    {"x": C-3, "y": C+3, "color": CH}, {"x": C+3, "y": C+3, "color": CH},
    {"x": C, "y": C-8, "color": CW}, {"x": C, "y": C+8, "color": CD},
])

# crit_cruel - 残忍: sharp fang
write_spec("talent-crit_cruel", [
    {"x": C-2, "y": 8, "color": CH}, {"x": C, "y": 10, "color": CH},
    {"x": C-4, "y": 14, "color": CR}, {"x": C-1, "y": 16, "color": CR},
    {"x": C+2, "y": 18, "color": CR}, {"x": C-3, "y": 22, "color": CR},
    {"x": C+3, "y": 24, "color": CD}, {"x": C, "y": 26, "color": CD},
    {"x": C+1, "y": 30, "color": CD}, {"x": C-2, "y": 34, "color": CR},
    {"x": C+2, "y": 36, "color": CR}, {"x": C, "y": 40, "color": CD},
])

# crit_skinner - 剥皮: curved skinning knife
write_spec("talent-crit_skinner", [
    {"x": C+4, "y": 10, "color": CH}, {"x": C+2, "y": 12, "color": CH},
    {"x": C, "y": 16, "color": CR}, {"x": C-2, "y": 20, "color": CR},
    {"x": C-4, "y": 24, "color": CR}, {"x": C-6, "y": 28, "color": CD},
    {"x": C-8, "y": 32, "color": CD}, {"x": C+3, "y": 14, "color": CW},
    {"x": C+1, "y": 20, "color": CW}, {"x": C-3, "y": 28, "color": CD},
    {"x": C-1, "y": 34, "color": CD}, {"x": C+1, "y": 36, "color": CR},
])

# crit_bleed - 致命出血: blood droplets falling
write_spec("talent-crit_bleed", [
    {"x": C, "y": 8, "color": CH}, {"x": C-2, "y": 12, "color": CR},
    {"x": C+2, "y": 14, "color": CR}, {"x": C-1, "y": 19, "color": CD},
    {"x": C+1, "y": 21, "color": CH}, {"x": C-3, "y": 26, "color": CD},
    {"x": C+3, "y": 28, "color": CD}, {"x": C, "y": 33, "color": CR},
    {"x": C-2, "y": 37, "color": CD}, {"x": C+2, "y": 39, "color": CD},
    {"x": C-4, "y": 17, "color": CH}, {"x": C+3, "y": 23, "color": CW},
])

# crit_omen_kill - 斩杀预兆: execution threshold mark
write_spec("talent-crit_omen_kill", [
    {"x": C-8, "y": 16, "color": CH}, {"x": C+8, "y": 16, "color": CH},
    {"x": C-4, "y": 18, "color": CR}, {"x": C+4, "y": 18, "color": CR},
    {"x": C, "y": 20, "color": CW}, {"x": C-6, "y": 24, "color": CR},
    {"x": C+6, "y": 24, "color": CR}, {"x": C-2, "y": 28, "color": CD},
    {"x": C+2, "y": 28, "color": CD}, {"x": C, "y": 32, "color": CD},
    {"x": C-3, "y": 36, "color": CR}, {"x": C+3, "y": 36, "color": CR},
    {"x": C, "y": 14, "color": CW},
])

# crit_omen_reap - 死兆收割: reaping scythe hook
write_spec("talent-crit_omen_reap", [
    {"x": C+6, "y": 8, "color": CH}, {"x": C+4, "y": 12, "color": CR},
    {"x": C+2, "y": 16, "color": CR}, {"x": C, "y": 20, "color": CW},
    {"x": C-2, "y": 24, "color": CR}, {"x": C-4, "y": 28, "color": CD},
    {"x": C-6, "y": 32, "color": CD}, {"x": C, "y": 14, "color": CH},
    {"x": C-2, "y": 18, "color": CR}, {"x": C-4, "y": 22, "color": CR},
    {"x": C+6, "y": 10, "color": CW}, {"x": C-1, "y": 30, "color": CD},
    {"x": C-3, "y": 34, "color": CR},
])

# crit_death_ecstasy - 死亡狂喜: skull with aura
write_spec("talent-crit_death_ecstasy", [
    {"x": C-4, "y": 12, "color": CH}, {"x": C+4, "y": 12, "color": CH},
    {"x": C-6, "y": 16, "color": CR}, {"x": C+6, "y": 16, "color": CR},
    {"x": C-4, "y": 18, "color": CD}, {"x": C+4, "y": 18, "color": CD},
    {"x": C-2, "y": 22, "color": CR}, {"x": C+2, "y": 22, "color": CR},
    {"x": C, "y": 24, "color": CW}, {"x": C-1, "y": 28, "color": CD},
    {"x": C+1, "y": 28, "color": CD}, {"x": C, "y": 32, "color": CD},
    {"x": C-8, "y": 18, "color": CW}, {"x": C+8, "y": 18, "color": CW},
])

# crit_final_cut - 终末血斩: large downward blade
write_spec("talent-crit_final_cut", [
    {"x": C, "y": 6, "color": CH}, {"x": C-2, "y": 10, "color": CH},
    {"x": C+2, "y": 10, "color": CH}, {"x": C-3, "y": 15, "color": CR},
    {"x": C+3, "y": 15, "color": CR}, {"x": C-2, "y": 20, "color": CR},
    {"x": C+2, "y": 20, "color": CR}, {"x": C-1, "y": 25, "color": CD},
    {"x": C+1, "y": 25, "color": CD}, {"x": C, "y": 30, "color": CD},
    {"x": C, "y": 34, "color": CD}, {"x": C-1, "y": 38, "color": CR},
    {"x": C+1, "y": 38, "color": CR},
])

# crit_ultimate - 末日审判: doom skull crossed marks
write_spec("talent-crit_ultimate", [
    {"x": C, "y": 8, "color": CH}, {"x": C-5, "y": 14, "color": CR},
    {"x": C+5, "y": 14, "color": CR}, {"x": C-6, "y": 18, "color": CD},
    {"x": C+6, "y": 18, "color": CD}, {"x": C-3, "y": 20, "color": CR},
    {"x": C+3, "y": 20, "color": CR}, {"x": C, "y": 22, "color": CW},
    {"x": C-2, "y": 26, "color": CD}, {"x": C+2, "y": 26, "color": CD},
    {"x": C-4, "y": 30, "color": CR}, {"x": C+4, "y": 30, "color": CR},
    {"x": C-1, "y": 34, "color": CD}, {"x": C+1, "y": 34, "color": CD},
])

print(f"\nDone! Total specs: {len(os.listdir(SPEC_DIR))}")
