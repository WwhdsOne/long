# Pixel Asset Pipeline

AI 文本模型生成像素坐标 JSON → Python 脚本渲染 PNG → 自动校验。

## 工作流

```
Claude 生成 specs/*.json
         ↓
python scripts/render_pixels.py specs/  -o output/
         ↓
python scripts/validate_pixels.py specs/  output/
         ↓
output/*.png  →  frontend/public/effects/
```

## Spec JSON 格式

```json
{
    "name": "slash-green-0",
    "size": [48, 48],
    "constraints": {
        "maxPixels": 10,
        "allowAlpha": false
    },
    "pixels": [
        {"x": 23, "y": 24, "color": "#4ade80"},
        {"x": 25, "y": 26, "color": "#2bb873", "alpha": 255}
    ]
}
```

## 约束规则

| 字段 | 说明 |
|------|------|
| `x` | 从左到右，0 起始 |
| `y` | 从上到下，0 起始 |
| `color` | Hex 格式 #RRGGBB |
| `alpha` | 可选，默认 255，设为 128 实现透明度 |

## 命令

```bash
# 渲染所有 spec
python scripts/render_pixels.py specs/ -o output/

# 渲染单个 spec
python scripts/render_pixels.py specs/slash-green-0.json

# 仅校验不输出
python scripts/render_pixels.py specs/ --dry-run

# 校验已渲染的 PNG 是否符合 spec 约束
python scripts/validate_pixels.py specs/ output/
```

## 校验规则

- 像素数量 ≤ maxPixels
- 不允许半透明像素（除非 allowAlpha: true）
- 不允许使用 spec 未定义的颜色
- 不允许像素贴到画布边缘
- spec 定义的像素数 = 渲染出的像素数
