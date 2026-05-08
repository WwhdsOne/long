# 战斗热链路 MessagePack 渐进迁移方案

> 状态：待实现
> 日期：2026-05-08
> 结论：本次只迁移战斗高频 Redis 热字段 `bossCurrent.parts` 与 `player:talent_state:<nickname>:<bossID>.state`。读取按 `MessagePack -> JSON` 双格式兼容，写入统一落 MessagePack，通过正常战斗写回让活跃战斗数据渐进升级。

## 一、目标与边界

本次方案只处理战斗期高频热链路，不改对外协议：

- `bossCurrent.parts`
- `player:talent_state:<nickname>:<bossID>.state`

明确不纳入本次范围：

- Boss 模板 `layout`
- Boss 历史 `parts`
- pending talent events 队列 JSON
- 其他低频 Redis JSON 字段

对外保持不变：

- HTTP 路由不变
- SSE payload 不变
- Redis key 名与 hash field 名不变

## 二、现状与问题

当前仓库里已有一部分 MessagePack 迁移已经落地：

- 玩家普通天赋状态 `player:talents:<nickname>.state` 已支持 MessagePack
- 读取策略已经采用“先 MessagePack，失败再 JSON”

但战斗热链路仍有两个高频对象继续走 JSON：

- 当前 Boss 的 `parts`
- 玩家战斗态 `TalentCombatState`

这两个字段的问题很直接：

- 战斗点击、流血 tick、魔法连锁等高频路径会频繁序列化/反序列化
- 当前 `store.go` / `talent.go` 中仍散落 `sonic.Marshal` / `sonic.Unmarshal`
- 如果继续追加战斗效果状态，热链路 JSON 开销会持续放大

因此本次只补齐战斗期 Redis 热字段，不扩大到模板、历史和事件队列。

## 三、迁移原则

迁移策略固定如下：

1. 读取时先按 MessagePack 解码
2. MessagePack 解码失败时回退 JSON 解码
3. 若旧 JSON 成功读出，则本次请求内先正常使用
4. 后续任何正常战斗写回都统一落 MessagePack
5. 不做双写 JSON
6. 不跑离线批量迁移脚本
7. 不新开 Redis key，不加版本字段

这意味着：

- 非活跃旧数据可以继续保持原样，直到下一次进入实际战斗写路径
- 活跃战斗数据会自然渐进升级
- 回滚时仍然只需要处理现有字段，不需要额外清理迁移痕迹

## 四、热对象编解码收口

本次在 `backend/internal/core` 内新增一组私有编解码助手，只处理热对象：

- `encodeBossParts`
- `decodeBossParts`
- `encodeTalentCombatState`
- `decodeTalentCombatState`

实现要求锁定为：

- 编码统一使用现有 `github.com/vmihailenco/msgpack/v5`
- 解码顺序固定为 `msgpack.Unmarshal -> sonic.Unmarshal`
- `TalentCombatState` 解码后统一补齐所有 nil map
- `BossPart` / `TalentCombatState` 不改对外 JSON tag
- 如需稳定 MessagePack 字段名，可补 `msgpack` tag，但不额外改结构语义

这里的重点不是抽象层次，而是避免后续又在热路径里新增零散 JSON 编解码。

## 五、Boss 部位热链路改造

### 5.1 读取入口

以下入口切到 `decodeBossParts`：

- `currentBossFromCmdable()`
- `normalizeBoss()`

行为要求：

- Redis 哈希字段仍然叫 `parts`
- 旧 JSON `parts` 仍可被正常读取
- 新 MessagePack `parts` 也能正常读取

### 5.2 写入入口

以下战斗写路径切到 `encodeBossParts`：

- `AttackBossPartAFK*`
- `applyBossPartDamage`
- `persistTalentTickState`

行为要求：

- 只改战斗期间 `bossCurrent.parts`
- 后台激活 Boss、保存历史、模板保存本次不改
- `/api/battle/state` 与 SSE 继续输出 JSON，前端无感知

### 5.3 渐进升级方式

如果 Redis 中当前仍是旧 JSON `parts`：

- 读路径照常得到内存 `[]BossPart`
- 下一次任意战斗写回都会把同一 field 升级成 MessagePack

也就是说，迁移动作由真实战斗自然触发，不单独加迁移脚本。

## 六、战斗天赋状态热链路改造

### 6.1 读取入口

`GetTalentCombatState()` 改为通过 `decodeTalentCombatState` 读取 `state`。

行为要求：

- 旧 JSON `state` 继续可读
- 新 MessagePack `state` 可读
- 所有 nil map 补齐行为与当前实现保持一致

### 6.2 写入入口

以下入口统一改为写 MessagePack：

- `SaveTalentCombatState()`
- `applyBossPartDamage` 内联写入
- `persistTalentTickState()`
- 其他走 `SaveTalentCombatState()` 的调用链

行为要求：

- Redis 哈希字段名仍然叫 `state`
- pending talent events 队列继续保持 JSON，不在本次范围

### 6.3 渐进升级方式

如果某玩家某次 Boss 战里的 `state` 还是旧 JSON：

- `GetTalentCombatState()` 仍可读出
- 一旦该玩家在该 Boss 战里再次发生保存，该字段自然改写为 MessagePack

## 七、代码收口要求

本次需要把散落在热路径里的以下逻辑收口到编解码助手：

- `sonic.Marshal(boss.Parts)`
- `sonic.Unmarshal([]byte(partsRaw), &parts)`
- `sonic.Marshal(combatState)`
- `sonic.Unmarshal([]byte(raw), &state)`

目标不是做大重构，而是只收掉本次迁移明确涉及的重复序列化逻辑，避免未来又新增新的 JSON 热点。

## 八、测试方案

必须补齐以下场景：

### 8.1 `bossCurrent.parts` 兼容读取

- 旧 JSON `parts` 能被 `currentBossForRoom()` 正常读出
- 新 MessagePack `parts` 也能正常读出

### 8.2 `bossCurrent.parts` 渐进写回

- 先写入旧 JSON `parts`
- 触发一次战斗写回
- 断言 Redis 中 `parts` 已升级为 MessagePack
- 再次读取结果一致

### 8.3 `TalentCombatState` 兼容读取

- 旧 JSON `state` 能被 `GetTalentCombatState()` 正常读取
- 新 MessagePack `state` 能正常读取
- nil map 会被补齐

### 8.4 `TalentCombatState` 渐进写回

- 先写旧 JSON `state`
- 触发一次保存路径
- 断言 Redis 中 `state` 已升级为 MessagePack
- 再次读取状态一致

### 8.5 回归验证

- 现有战斗点击测试继续通过
- 流血 tick、魔法回响、Silver Storm、Judgment Day、Collapse 相关测试继续通过
- `/api/battle/state` 与个人态协议不新增格式改动

建议优先扩展：

- `backend/internal/core/store_test.go`
- `backend/internal/core/magic_trigger_store_test.go`
- 如有必要，可新增小型 codec 单测文件

## 九、风险与回滚

主要风险：

- `parts` 双格式读取不严谨，导致当前 Boss 读取失败
- `TalentCombatState` nil map 补齐漏项，触发运行时行为回归
- 战斗 tick 或技能触发路径仍遗留 JSON 写法，导致迁移不完整

降险方式：

- 只动战斗热链路，不扩到模板、历史和事件队列
- 读取统一走双格式兼容
- 写入统一走单一 MessagePack
- 用现有战斗链路测试兜底回归

回滚口径：

- 对外协议未变
- Redis key / field 未变
- 如需回滚，只需回滚服务代码，不需要清洗迁移后字段名

## 十、实施顺序

建议按以下顺序执行：

1. 先新增文档并链接到现有 MessagePack 方案入口
2. 先补战斗热链路兼容/迁移测试，确认红灯
3. 再新增私有编解码助手并改读路径
4. 再改战斗写路径统一写 MessagePack
5. 最后跑相关后端测试与回归校验

## 十一、与既有文档关系

这篇文档是 [天赋溢出节点与 MessagePack 存储方案](./2026-05-07-天赋溢出节点与MessagePack存储方案.md) 的战斗热链路补充。

关系边界如下：

- 2026-05-07 文档处理“玩家普通天赋状态 `TalentState` 的 MessagePack 迁移”
- 本文处理“战斗态 `TalentCombatState` 与 Boss 当前 `parts` 的 MessagePack 渐进迁移”

两篇文档共同组成当前 Redis MessagePack 迁移的主线方案。
