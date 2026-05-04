# 用户态载荷瘦身优化总结

日期：2026-05-04

> 状态：已完成，代码已落在 `dev` 分支，未推送。
>
> 本文总结本次围绕点击链路和 `user_state` 个人态推送做的载荷瘦身优化，重点说明“哪些消息不再带大字段”和“为什么这样做仍能保证数据正确”。

## 一、背景

这次优化前，个人实时态 `user_state` 会复用同一套载荷结构。

其中几个字段体积明显偏大：

- `loadout`
- `talentCombatState`
- `combatStats`
- `recentRewards`

实际排查后发现：

1. 高频点击链路真正需要的是“本次点击立即反馈”。
2. `loadout` 和 `combatStats` 这类配置型数据不会在每次点击后变化。
3. 但在部分 `BroadcastUserAll` 场景里，点击后仍可能补发一份带 `loadout/combatStats` 的 `user_state`。

这会带来两个问题：

- 高频点击下重复传输不变的大字段，浪费带宽。
- 前端实时合并时，需要反复处理与本次点击无关的配置型数据。

## 二、优化目标

本次优化只做一件事：

> 点击时不再反复推送 `loadout/combatStats`，只有这些数据真实发生变化时，后端才推送一次最新值。

同时必须守住一个前提：

> 数据正确性优先，不能为了减载把前端状态做成“猜出来”的。

## 三、最终策略

本次落地后的策略分成三类消息。

### 1. 首次连接 / 重连首包

首次进入 SSE 连接，后端仍然发送带完整资料字段的 `user_state`。

保留字段包括：

- `loadout`
- `combatStats`
- 资源字段
- 个人 Boss 字段
- `recentRewards`
- `talentCombatState`

这样前端在任何一次新建连接时，都能拿到一份可信全量用户态，作为后续增量更新的基线。

### 2. 点击后的实时反馈

点击链路继续走现有小包即时反馈：

- `click_ack`
- 公共态 `public_state`
- 必要时的瘦身 `user_state`

点击后如果触发 `user_state` 补发，本次优化后默认不再带：

- `loadout`
- `combatStats`

点击链路只保留与战斗反馈直接相关的字段，例如：

- 金币 / 强化石 / 天赋点
- 个人 Boss 伤害
- `recentRewards`
- `talentCombatState`
- `talentEvents`

也就是说，点击只更新“这次战斗刚变动的数据”，不再重复发送装备栏和最终面板。

### 3. 配置型变更后的用户态广播

当玩家资料型数据真的发生变化时，后端仍然会推送带配置字段的 `user_state`。

本次明确保留完整资料广播的场景是：

- `equipment_changed`
- `equipment_meta_changed`

这两类变化会继续携带：

- `loadout`
- `combatStats`

这样前端不需要自行推导装备变化后的最终面板，而是直接接收后端给出的最新真值。

## 四、为什么这样做仍然安全

这次优化刻意没有做“任意字段脏更新 diff 合并”，原因是那种方案最容易引入状态漂移。

本次采用的是更保守的做法：

1. 初次连接始终有全量可信基线。
2. 高频点击只裁掉确定不会在点击时变化的配置型字段。
3. 装备类变更时，`loadout` 和 `combatStats` 仍然成组下发，不拆开推送。

这样可以避免两个常见问题：

- 只更新了装备栏，没同步最终面板。
- 前端靠旧 `loadout` 猜新 `combatStats`，结果与后端真实计算不一致。

因此，本次优化不是“让前端自己推导”，而是“减少不必要的重复真值广播”。

## 五、具体实现

### 1. `user_state` 载荷改为按场景选带资料字段

后端把 `realtimeUserStatePayload` 中的：

- `loadout`
- `combatStats`

改成了按场景决定是否填充。

对应实现位于：

- `backend/internal/events/hub.go`

### 2. 分发器按变更类型决定是否带资料字段

事件分发层新增了按 `StateChangeType` 判断是否携带资料字段的逻辑：

- `equipment_changed`
- `equipment_meta_changed`

返回 `true`，其余场景默认不带 `loadout/combatStats`。

对应实现位于：

- `backend/internal/events/dispatcher.go`

### 3. SSE 首包仍保留完整用户态

SSE `NewHandler` 在首次建立连接时，仍然构造完整 `user_state` 作为首包，保证前端有可信初始值。

这意味着：

- 优化只影响后续实时补发
- 不影响首次加载和重连恢复

## 六、验证结果

本次改动补了针对性测试，覆盖两条核心行为：

1. Boss 点击触发 `BroadcastUserAll` 时，补发的 `user_state` 不再包含：
   - `loadout`
   - `combatStats`
2. 装备变更触发 `equipment_changed` 时，补发的 `user_state` 仍然包含：
   - `loadout`
   - `combatStats`

本次提交前已执行：

```bash
go test -C backend ./internal/events ./internal/httpapi
```

结果：

- `ok  	long/internal/events`
- `ok  	long/internal/httpapi`

## 七、直接收益

这次优化带来的收益主要有三点：

1. 高频点击时，不再重复发送大体积且不变的资料字段。
2. 前端点击态合并更聚焦，只处理本次战斗相关数据。
3. 保留“首包全量 + 变更真值广播”的安全边界，不依赖前端猜测最终面板。

## 八、当前边界

本次优化暂时只收紧了：

- 点击后的 `user_state`
- 事件分发层的资料字段下发策略

还没有继续做的事情包括：

- 更细粒度的通用字段 diff
- `talentCombatState` 进一步裁剪
- `recentRewards` 单独拆频道
- 字段名短编码

这些方向虽然还能继续减载，但也会显著提高前后端状态同步复杂度。当前版本优先保证正确性，不继续扩大变更面。

## 九、结论

这次优化的核心不是“把所有消息都做成增量 diff”，而是先把最不该在点击时重复发送的两块资料字段拿掉：

- `loadout`
- `combatStats`

最终效果是：

- 点击消息更轻
- 配置变更仍然准确
- 首次连接仍然可信

在“节省载荷”和“保证数据总是对的”之间，这次实现选择了更稳的一侧。
