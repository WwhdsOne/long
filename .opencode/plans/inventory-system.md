# 背包系统开发者参考

**文档版本**：2026-04-26
**适用范围**：装备背包、穿戴、强化、分解、战斗属性计算

---

## 一、核心概念

### 1.1 装备定义 vs 装备实例

| 概念 | Redis Key | 说明 |
|---|---|---|
| **装备定义**（EquipmentDefinition） | `{prefix}equipment:{itemId}` | 装备模板，定义名称、槽位、稀有度、基础属性、掉落权重等 |
| **装备实例**（ItemInstance） | `{prefix}equipment:instance:{instanceId}` | 玩家实际拥有的装备，记录强化等级、消耗强化石、绑定/锁定状态 |

一个装备定义可以被多个玩家拥有，每个玩家拥有独立的装备实例。

### 1.2 玩家资源

| 资源 | Redis Key | 用途 |
|---|---|---|
| 金币（Gold） | `{prefix}user:{nickname}:resources` → `gold` | 强化消耗 |
| 强化石（Stones） | `{prefix}user:{nickname}:resources` → `stones` | 强化消耗、分解返还 |
| 天赋点（TalentPoints） | `{prefix}user:{nickname}:resources` → `talent_points` | 天赋系统 |

---

## 二、数据结构

### 2.1 InventoryItem（背包物品视图）

```go
type InventoryItem struct {
    ItemID               string  `json:"itemId"`               // 装备定义 ID
    InstanceID           string  `json:"instanceId,omitempty"` // 实例 ID（唯一标识）
    Name                 string  `json:"name"`                 // 显示名称
    Slot                 string  `json:"slot"`                 // 槽位：weapon/helmet/chest/gloves/legs/accessory
    Rarity               string  `json:"rarity"`               // 稀有度：普通/稀有/史诗/至臻
    ImagePath            string  `json:"imagePath,omitempty"`  // 图片路径
    ImageAlt             string  `json:"imageAlt,omitempty"`   // 图片描述
    Quantity             int64   `json:"quantity"`             // 数量（当前恒为 1）
    Equipped             bool    `json:"equipped"`             // 是否已穿戴
    EnhanceLevel         int     `json:"enhanceLevel,omitempty"` // 强化等级
    Bound                bool    `json:"bound,omitempty"`      // 是否绑定
    Locked               bool    `json:"locked,omitempty"`     // 是否锁定（锁定不可分解）
    AttackPower          int64   `json:"attackPower,omitempty"`          // 攻击力
    ArmorPenPercent      float64 `json:"armorPenPercent,omitempty"`      // 护甲穿透百分比
    CritRate             float64 `json:"critRate,omitempty"`             // 暴击率
    CritDamageMultiplier float64 `json:"critDamageMultiplier,omitempty"` // 暴击伤害倍率
    PartTypeDamageSoft   float64 `json:"partTypeDamageSoft,omitempty"`   // 软部位伤害加成
    PartTypeDamageHeavy  float64 `json:"partTypeDamageHeavy,omitempty"`  // 重部位伤害加成
    PartTypeDamageWeak   float64 `json:"partTypeDamageWeak,omitempty"`   // 弱点伤害加成
}
```

### 2.2 ItemInstance（Redis 存储结构）

```go
type ItemInstance struct {
    InstanceID   string `json:"instanceId"`   // 实例 ID
    ItemID       string `json:"itemId"`       // 装备定义 ID
    EnhanceLevel int    `json:"enhanceLevel"` // 强化等级
    SpentStones  int64  `json:"spentStones"`  // 累计消耗强化石（分解时返还 60%）
    Bound        bool   `json:"bound"`        // 是否绑定
    Locked       bool   `json:"locked"`       // 是否锁定
    CreatedAt    int64  `json:"createdAt"`    // 创建时间戳
}
```

### 2.3 Loadout（穿戴槽位）

```go
type Loadout struct {
    Weapon    *InventoryItem `json:"weapon,omitempty"`    // 武器
    Helmet    *InventoryItem `json:"helmet,omitempty"`    // 头盔
    Chest     *InventoryItem `json:"chest,omitempty"`     // 胸甲
    Gloves    *InventoryItem `json:"gloves,omitempty"`    // 手套
    Legs      *InventoryItem `json:"legs,omitempty"`      // 腿甲
    Accessory *InventoryItem `json:"accessory,omitempty"` // 饰品
}
```

每个槽位同时只能穿戴一件装备。穿戴新装备会自动替换旧装备。

---

## 三、Redis 存储结构

### 3.1 Key 布局

```
{prefix}equipment:{itemId}              # 装备定义 Hash
{prefix}equipment:instance:{instanceId} # 装备实例 Hash
{prefix}user:{nickname}:instances       # 玩家拥有的实例 ID 集合 (Set)
{prefix}user:{nickname}:loadout         # 玩家穿戴槽位 Hash (slot → instanceId)
{prefix}user:{nickname}:resources       # 玩家资源 Hash (gold/stones/talent_points)
{prefix}user:{nickname}:last_reward     # 最近一次掉落记录 Hash
{prefix}boss:loot:{bossId}             # Boss 掉落池 (有序集合或 Hash)
```

### 3.2 装备定义示例

```
HGETALL {prefix}equipment:sword_001
  name: "铁剑"
  slot: "weapon"
  rarity: "普通"
  weight: "100"
  drop_rate_percent: "10.0"
  attack_power: "10"
  armor_pen_percent: "0"
  crit_rate: "0"
  crit_damage_multiplier: "0"
  image_path: "/images/sword_001.png"
  image_alt: "一把普通的铁剑"
```

### 3.3 装备实例示例

```
HGETALL {prefix}equipment:instance:inst_abc123
  item_id: "sword_001"
  enhance_level: "3"
  spent_stones: "15"
  bound: "0"
  locked: "0"
  created_at: "1714123456"
```

---

## 四、装备生命周期

### 4.1 装备获得

**触发条件**：Boss 被击杀时，对每个伤害达标的参与者执行掉落判定。

**代码路径**：`store.go:1642-1748` → `finalizeBossKill`

**掉落流程**：
1. 加载 Boss 掉落池（`loadBossLoot`）
2. 对每个掉落条目执行 `rollLootDrops`（按 `DropRatePercent` 概率判定）
3. 掉落成功则：
   - 生成新的 `instanceID`
   - 写入 Redis 装备实例（`equipmentInstanceKey`）
   - 将 `instanceID` 加入玩家实例集合（`playerInstancesKey`）
   - 记录到 `lastRewardKey`
4. 返回 `ClickResult.RecentRewards`

**挂机模式**：装备掉落概率与手动点击相同，但金币/强化石奖励减半。

### 4.2 装备穿戴

**代码路径**：`store.go:670-702` → `EquipItem`

**流程**：
1. 验证玩家拥有该实例（`getOwnedInstance`）
2. 获取装备定义，确定槽位（`definition.Slot`）
3. Redis 事务：
   - `HSet loadoutKey slot instanceId`（写入穿戴槽位）
   - `ZAdd playerIndexKey`（更新玩家活跃时间）
4. 返回更新后的完整 `State`

**同槽位替换**：直接覆盖，旧装备自动变为"未穿戴"状态（仍在背包中）。

### 4.3 装备卸下

**代码路径**：`store.go:705-737` → `UnequipItem`

**流程**：
1. 验证玩家拥有该实例
2. 获取装备定义，确定槽位
3. Redis 事务：
   - `HDel loadoutKey slot`（清空穿戴槽位）
4. 返回更新后的完整 `State`

### 4.4 装备强化

**代码路径**：`store.go:740-794` → `EnhanceItem`

**流程**：
1. 验证玩家拥有该实例
2. 检查强化等级上限（`maxEnhanceLevel`，按稀有度不同）
3. 计算消耗：
   - `goldCost = enhanceGoldCost(currentLevel)`
   - `stoneCost = enhanceStoneCost(currentLevel)`
4. 检查玩家资源是否充足
5. Redis 事务：
   - `HIncrBy resources gold -goldCost`
   - `HIncrBy resources stones -stoneCost`
   - `HIncrBy instance spent_stones stoneCost`（记录累计消耗）
   - `HIncrBy instance enhance_level 1`
6. 返回更新后的完整 `State`

**强化等级上限**（按稀有度）：
- 普通：较低
- 稀有：中等
- 史诗：较高
- 至臻：最高

### 4.5 装备分解

**代码路径**：`store.go:842-903` → `SalvageItem`

**分解奖励**：
- 基础奖励：按稀有度获得金币和强化石（`salvageBaseReward`）
- 强化石返还：已消耗强化石的 60%（向下取整）

**流程**：
1. 验证玩家拥有该实例
2. 检查是否锁定（锁定不可分解）
3. 计算奖励
4. Redis 事务：
   - `SRem playerInstancesKey instanceId`（从背包移除）
   - `Del equipmentInstanceKey`（删除实例）
   - `HIncrBy resources gold +goldReward`
   - `HIncrBy resources stones +stoneGain`
   - 如果该装备正在穿戴，`HDel loadoutKey slot`
5. 返回 `SalvageResult`

### 4.6 一键分解

**代码路径**：`store.go:906-968` → `BulkSalvageUnequipped`

**规则**：
- 分解所有"未穿戴、未锁定、非至臻"的装备
- 排除已穿戴、已锁定、至臻品质的装备
- 返回分解统计和资源变化

### 4.7 装备锁定/解锁

**代码路径**：`store.go:797-839` → `LockItem` / `UnlockItem`

锁定的装备无法分解，但可以穿戴/卸下/强化。

---

## 五、战斗属性计算

### 5.1 计算流程

```
baseCombatStats()                    # 基础属性
    ↓
loadoutBonuses(loadout)              # 穿戴装备加成
    ↓
ComputeTalentModifiers(nickname)     # 天赋加成
    ↓
deriveCombatStats(stats)             # 派生属性
```

### 5.2 基础属性

```go
func (s *Store) baseCombatStats() CombatStats {
    return CombatStats{
        CriticalChancePercent: 5%,    // 基础暴击率
        CriticalCount:         1,     // 基础暴击次数
        AttackPower:           5,     // 基础攻击力
        ArmorPenPercent:       0,     // 基础护甲穿透
        CritDamageMultiplier:  1.5,   // 基础暴击伤害倍率
        AllDamageAmplify:      0,     // 全伤害增幅
        LowHpMultiplier:       1,     // 低血量伤害倍率
    }
}
```

### 5.3 穿戴装备加成

```go
func loadoutBonuses(loadout Loadout) (attackPower, armorPen, critRate, critDmgMult) {
    // 遍历所有穿戴槽位，累加各属性
    items := [Weapon, Helmet, Chest, Gloves, Legs, Accessory]
    for _, item := range items {
        attackPower += item.AttackPower
        armorPen += item.ArmorPenPercent
        critRate += item.CritRate
        critDmgMult += item.CritDamageMultiplier
    }
}
```

### 5.4 天赋加成

天赋系统通过 `ComputeTalentModifiers` 计算，影响：
- 攻击力百分比增幅
- 护甲穿透额外值
- 全伤害增幅
- 暴击伤害百分比加成
- 部位伤害加成（软/重/弱点）
- 低血量伤害倍率

### 5.5 派生属性

```go
func deriveCombatStats(stats CombatStats) CombatStats {
    stats.EffectiveIncrement = max(1, stats.AttackPower)
    stats.NormalDamage = stats.EffectiveIncrement

    // 暴击伤害取两种计算方式的较大值
    countBasedCriticalDamage = max(NormalDamage + CriticalCount - 1, NormalDamage)
    multiplierBasedCriticalDamage = NormalDamage * CritDamageMultiplier
    stats.CriticalDamage = max(countBasedCriticalDamage, multiplierBasedCriticalDamage)
}
```

---

## 六、装备与 Boss 战斗

### 6.1 伤害计算

```
CalcBossPartDamage(stats, partType, partArmor, alivePartCount, bossCurrentHP, bossMaxHP)
```

**计算公式**：
1. 基础攻击力 = `max(1, AttackPower)`
2. 部位类型系数 = `partType.DamageCoefficient()`
3. 护甲减伤 = `partArmor * (1 - ArmorPenPercent)`
4. 暴击判定 = `random() < CriticalChancePercent`
5. 最终伤害 = `(基础攻击力 * 部位系数 - 护甲减伤) * 暴击倍率 * 全伤害增幅 * 部位伤害加成 * 低血量加成`

### 6.2 部位类型

| 类型 | 说明 | 伤害系数 |
|---|---|---|
| 软部位（Soft） | 普通部位 | 1.0 |
| 重部位（Heavy） | 高护甲部位 | 系数较低 |
| 弱点（Weak） | 低护甲部位 | 系数较高 |

---

## 七、代码入口

| 操作 | 文件:行号 | 函数 |
|---|---|---|
| 装备穿戴 | `store.go:670` | `EquipItem` |
| 装备卸下 | `store.go:705` | `UnequipItem` |
| 装备强化 | `store.go:740` | `EnhanceItem` |
| 装备锁定 | `store.go:797` | `LockItem` |
| 装备解锁 | `store.go:802` | `UnlockItem` |
| 装备分解 | `store.go:842` | `SalvageItem` |
| 一键分解 | `store.go:906` | `BulkSalvageUnequipped` |
| Boss 击杀掉落 | `store.go:1642` | `finalizeBossKill` |
| 掉落判定 | `store.go:1771` | `rollLootDrops` |
| 战斗属性计算 | `store.go:1821` | `combatStatsForNickname` |
| Boss 伤害计算 | `store.go:1910` | `CalcBossPartDamage` |
| 挂机攻击 | `store.go:1176` | `AttackBossPartAFK` |

---

## 八、常见问题

### Q: 背包有容量限制吗？

A: 没有硬性上限。玩家可以拥有无限数量的装备实例。

### Q: 装备可以交易吗？

A: 不支持。装备实例绑定到玩家昵称，无法转移。

### Q: 分解装备会返还全部强化石吗？

A: 只返还 60%（向下取整）。例如消耗了 10 强化石，分解返还 6 强化石。

### Q: 挂机掉落的装备与手动点击有区别吗？

A: 没有区别。挂机模式下装备掉落概率相同，但金币/强化石奖励减半。

### Q: 同槽位穿戴新装备会怎样？

A: 直接覆盖，旧装备变为"未穿戴"状态（仍在背包中，不会丢失）。
