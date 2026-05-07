# 天赋溢出节点与 MessagePack 存储方案

> 状态：已审核，待实现
> 日期：2026-05-07
> 结论：新增独立“天赋点溢出消耗节点”`overflow_sink`；玩家天赋状态 Redis 存储切换为统一 `TalentState` 结构，主字段改为 `state`，采用 MessagePack 编码，并通过“读时写回迁移”让活跃玩家逐步无感完成迁移。

## 一、目标与边界

本次方案同时处理两件事：

- 新增一个不挂三系主树层级的独立溢出节点，消耗高额天赋点换随机微量永久强化
- 统一玩家天赋状态 Redis 存储模型，为后续继续扩展天赋附加状态留出口

本次结论已经锁定：

- 溢出入口是独立节点，不挂三系主树层级
- 单次消耗固定为 `1000` 天赋点
- 单次收益为随机一种属性 `+0.1%`
- 随机池固定为：
  - `软组织`
  - `弱点`
  - `重甲`
  - `暴击伤害`
  - `攻击力`
  - `全伤害`
- 洗点时，普通天赋和溢出节点一起清空并返还
- Redis 迁移采用读时先读新字段、失败降级旧字段、成功后立即写回新 MessagePack

本次不做：

- 不新增独立天赋接口
- 不做离线批量迁移脚本
- 不在首期删除旧 `talents` 字段
- 不改三系主树节点结构与 tier 解锁规则

## 二、现状问题

当前玩家天赋 Redis 存储本质上只覆盖“普通天赋等级”：

- key 维持 `player:talents:<nickname>`，代码中仍通过 `talentKey(nickname)` 取值
- 当前字段为 `talents`
- 当前值为 JSON

这个结构的问题已经很明确：

- 只能表达 `map[string]int` 形式的普通天赋等级
- 无法自然挂载“累计溢出投入”“随机命中分布”这类附加状态
- 后续如果继续扩字段，要么堆多个 field，要么继续追加脆弱兼容分支

因此这次直接把玩家天赋状态统一收口为一个结构体。

## 三、Redis 存储方案

### 3.1 键与字段

玩家天赋 key 保持不变：

```text
player:talents:<nickname>
```

实际实现仍通过：

```go
talentKey(nickname)
```

存储从当前：

- field：`talents`
- value：JSON

切换为：

- field：`state`
- value：MessagePack 编码后的统一结构体

### 3.2 统一状态结构

推荐结构如下：

```go
type TalentState struct {
    Talents         map[string]int `msgpack:"talents" json:"talents"`
    OverflowLevel   int64          `msgpack:"overflowLevel" json:"overflowLevel"`
    OverflowBonuses map[string]int `msgpack:"overflowBonuses" json:"overflowBonuses"`
}
```

字段语义：

- `Talents`
  - 继续承载普通天赋等级
- `OverflowLevel`
  - 表示总共投入过多少次溢出强化
  - 用于洗点返还总额与前端累计投入展示
- `OverflowBonuses`
  - 表示六种随机属性分别命中了多少次
  - 不存浮点，只存计数

固定属性 key：

- `soft_damage`
- `weak_damage`
- `heavy_damage`
- `crit_damage`
- `attack_power`
- `all_damage`

实际生效值统一按：

```text
count * 0.001
```

推导。

这个表示法比逐次记录 roll 历史更合适：

- 更紧凑
- 更容易序列化
- 足够支持展示
- 足够支持战斗结算
- 足够支持洗点返还

## 四、迁移策略

### 4.1 迁移原则

迁移采用“读时写回”，不跑批量脚本。

目标：

- 老号首次进入天赋页或触发天赋读路径时不报错
- 活跃玩家自动逐步迁移
- 首期保留旧字段，回滚兼容更稳

### 4.2 迁移顺序

迁移顺序固定如下：

1. 先读 `state`
2. 如果 `state` 存在：
   - 先按 MessagePack 解码
   - 失败后按 JSON 同结构体降级解码
3. 如果 `state` 不存在：
   - 回退读旧字段 `talents`
   - 先按旧 JSON `map[string]int` 解码
   - 再兼容更老的 `[]string`，并转成 `Lv1 map`
4. 只要旧格式成功读出，就立即构造新 `TalentState`
5. 在当前请求里同步回写 `state` 的 MessagePack
6. 默认保留旧 `talents` 字段，不在首期删除
7. 此后所有升级、洗点、新增溢出节点操作统一只写新 `state`

### 4.3 旧格式回写规则

旧格式读出后，统一回写为：

```go
TalentState{
    Talents:         oldTalents,
    OverflowLevel:   0,
    OverflowBonuses: map[string]int{},
}
```

这里有两个明确要求：

- `Talents` 直接继承旧值
- 溢出相关状态默认置空，不推测历史，不补虚构数据

### 4.4 为什么不做批量迁移

当前不选批量迁移脚本，原因很直接：

- 这是热数据，玩家活跃路径天然会触发读取
- 首次迁移逻辑简单，可直接挂在正常请求上
- 不需要额外上线窗口、脚本监控和失败补偿
- 保留旧字段后，线上回滚风险更低

因此首期直接采用请求内同步迁移。

## 五、溢出节点方案

### 5.1 节点定位

新增逻辑节点：

```text
overflow_sink
```

它的定位是：

- 不出现在 `GetTreeTalents` 的三系半圆主树里
- 前端在天赋页单独展示为一个独立卡片
- 不参与 tier 解锁
- 每次升级固定只升 1 级

### 5.2 升级规则

每次点击溢出强化时执行：

1. 检查 `talent_points >= 1000`
2. 从 6 个属性池中等概率随机 1 个
3. `OverflowLevel += 1`
4. `OverflowBonuses[randomStat] += 1`
5. `talent_points -= 1000`

点数不足时，沿用：

```text
ErrTalentPointsInsufficient
```

### 5.3 随机池与映射

随机池固定等权重：

- `soft_damage` -> 软组织增伤 `+0.1%`
- `weak_damage` -> 弱点增伤 `+0.1%`
- `heavy_damage` -> 重甲增伤 `+0.1%`
- `crit_damage` -> 暴击伤害 `+0.1%`
- `attack_power` -> 攻击力 `+0.1%`
- `all_damage` -> 全伤害 `+0.1%`

这里不做：

- 权重差异
- 保底
- 连续失败补偿
- 更复杂的随机池扩展协议

原因是首版目标很单纯：处理超大额天赋点溢出，给长期玩家一个稳定消耗口。

## 六、洗点与战斗属性接入

### 6.1 洗点规则

洗点规则固定如下：

- 普通天赋返还：沿用现有累计成本
- 溢出返还：`OverflowLevel * 1000`

重置后统一写回：

- `Talents = {}`
- `OverflowLevel = 0`
- `OverflowBonuses = {}`

清空后继续走现有：

```text
invalidatePlayerCombatCaches
```

确保溢出强化的战斗加成立即失效。

### 6.2 战斗属性接入

后端属性接入关系固定如下：

- `all_damage`
  - `AllDamageAmplify += 0.001 * count`
- `attack_power`
  - `AttackPowerPercent += 0.001 * count`
- `crit_damage`
  - `CritDamagePercentBonus += 0.001 * count`
- `soft_damage`
  - `PartTypeDamageSoftBonus += 0.001 * count`
- `weak_damage`
  - `PartTypeDamageWeakBonus += 0.001 * count`
- `heavy_damage`
  - `PartTypeDamageHeavyBonus += 0.001 * count`

若当前 `TalentModifiers` 尚无三项部位聚合字段，则补齐：

- `PartTypeDamageSoftBonus`
- `PartTypeDamageWeakBonus`
- `PartTypeDamageHeavyBonus`

最终统一接入：

```text
ApplyTalentEffectsToCombatStats
```

### 6.3 为什么按计数推导，不直接存浮点

不直接存浮点，原因是：

- 浮点序列化更容易引入格式差异
- 前后端展示都只关心 `0.1% * 次数`
- 洗点返还只需要看总投入次数，不需要历史 roll 顺序

所以这次统一只存整数计数。

## 七、接口与前端展示

### 7.1 接口策略

沿用现有接口，不新开 endpoint：

- `GET /api/talents/state`
- `POST /api/talents/upgrade`
- `POST /api/talents/reset`

`GET /api/talents/state` 新增返回字段：

- `overflowLevel`
- `overflowBonuses`
- `overflowUpgradeCost: 1000`

`POST /api/talents/upgrade` 需要支持：

- `talentId = "overflow_sink"`

对这个节点：

- `targetLevel` 忽略普通等级语义
- 请求语义固定为“升 1 级”

### 7.2 前端展示策略

前端天赋页保持主树不变：

- 不把溢出节点塞进半圆主树

在 `TalentsPage.vue` 增加一个独立卡片，展示：

- 当前可用天赋点
- 溢出等级
- 累计消耗
- 已获得属性汇总
- “消耗 1000 点随机强化”按钮

洗点提示文案同步更新为：

- 普通天赋与溢出强化都会被清空并返还

### 7.3 前端展示建议

建议前端把汇总值同时展示“命中次数 + 百分比换算”：

- 例如 `攻击力：12 次（+1.2%）`

这样有两个好处：

- 玩家能看懂累计结果
- 前端不需要理解更复杂的状态模型

## 八、实施顺序

建议按下面顺序落地：

1. 定义 `TalentState` 与 MessagePack 编解码辅助逻辑
2. 改造天赋状态读路径，完成 `state -> 旧 talents` 的读时迁移
3. 改造普通升级、溢出升级、洗点写路径，统一只写新 `state`
4. 新增溢出节点后端逻辑与战斗属性接入
5. 扩展 `GET /api/talents/state` 返回结构
6. 前端天赋页新增独立溢出卡片与洗点提示文案
7. 补齐后端和前端测试

## 九、测试方案

### 9.1 Redis 兼容测试

必须覆盖：

- 读新 `state` MessagePack 成功
- `state` 为 JSON 时可降级成功
- 无 `state` 时可读旧 `talents` JSON map
- 无 `state` 时可读更老的 `[]string`
- 旧格式读取成功后会回写新 `state`

### 9.2 溢出节点测试

必须覆盖：

- `overflow_sink` 升级扣 `1000` 点
- 随机结果只落在 6 个允许属性内
- `OverflowLevel` 与 `OverflowBonuses` 正确累计
- 点数不足时报 `ErrTalentPointsInsufficient`

### 9.3 洗点测试

必须覆盖：

- 普通天赋 + 溢出节点混合时返还总额正确
- 洗点后溢出状态归零
- 洗点后战斗属性恢复

### 9.4 前端测试

必须覆盖：

- 天赋页显示溢出节点卡片
- 状态接口新增字段被正确消费
- 洗点后卡片归零
- 老号迁移后页面不报错

## 十、风险与回滚

### 10.1 首期风险点

主要风险有三类：

- `state` 编解码实现不严谨，导致老号读取失败
- 溢出加成写入成功但战斗态未正确接入
- 洗点返还漏算溢出投入

### 10.2 降险措施

本方案已内建这些缓冲：

- `state` 先尝试 MessagePack，失败再按 JSON 同结构体解码
- 无 `state` 时仍保留旧 `talents` 兼容读取
- 首期保留旧 `talents` 字段，不立刻删除
- 所有新写路径统一只收口到 `state`，避免双写分叉

### 10.3 回滚口径

如果首期上线后需要快速回滚：

- 旧版本仍可继续读取旧 `talents` 字段
- 新增的 `state` 字段可以先忽略
- 因为首期不删旧字段，所以回滚不会依赖额外修复脚本

## 十一、不选方案说明

当前不选以下方案：

- 溢出节点挂入三系主树
  - 会把 tier、层级和 UI 结构一起拉复杂
- 单独为溢出节点新开 API
  - 没必要，现有升级/重置接口足够表达
- 直接把随机结果存成浮点百分比
  - 表达冗余，且不如整数计数稳定
- 跑离线全量迁移
  - 成本更高，收益不足
- 首期删除旧 `talents`
  - 回滚弹性太差

因此当前方案是：独立节点 + 统一状态结构 + 读时写回迁移。
