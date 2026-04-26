#!/usr/bin/env python3
"""生成恶搞装备图标 Pixel Art 提示词 —— 交互式"""

import sys

SLOTS = ["weapon", "helmet", "chest", "gloves", "legs", "accessory"]

SLOT_DESC = {
    "weapon": "a weapon",
    "helmet": "a helmet",
    "chest": "chest armor",
    "gloves": "gloves",
    "legs": "leg armor",
    "accessory": "an accessory",
}

TEMPLATE = (
    "Pixel art, game equipment icon, {slot_desc}, {rarity_desc}, "
    "16-bit sprite, pixelated edges, hard outlines, limited color palette, "
    "no gradients, no blur, crisp pixels visible, "
    "128x128, isolated on transparent background, centered, "
    "no text, no background scene."
)


def select_slot() -> str | None:
    """交互式选择部位，数字选单。返回部位 key 或 None（全部）。"""
    print("\n选择部位：")
    print("  0. 全部部位")
    for i, s in enumerate(SLOTS, 1):
        print(f"  {i}. {s}")
    print("  q. 退出")

    while True:
        choice = input("请输入数字 (0-6) 或 q: ").strip()
        if choice.lower() == "q":
            sys.exit(0)
        if choice == "0":
            return None
        if choice in ("1", "2", "3", "4", "5", "6"):
            return SLOTS[int(choice) - 1]
        print("输入无效，重新输入")


def main():
    print("=" * 50)
    print("  恶搞装备 Pixel Art Prompt 生成器")
    print("=" * 50)

    # 选择部位
    slot = select_slot()

    # 输入恶搞描述
    print("\n输入恶搞描述（比如 'Lululemon 黑色瑜伽裤'）：")
    desc = input("> ").strip()
    if not desc:
        print("描述不能为空")
        sys.exit(1)

    # 生成
    slots_to_gen = [slot] if slot else SLOTS
    for s in slots_to_gen:
        prompt = TEMPLATE.format(
            slot_desc=SLOT_DESC[s],
            rarity_desc=desc,
        )
        print(f"\n--- {s} ---")
        print(prompt)

    print("\n完成！")


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print("\n已退出")
        sys.exit(0)