# 运营报表脚本设计

## 概述

基于 MongoDB 中的游戏数据，生成运营报表。CLI 工具放 `cmd/report/`，查询+格式化逻辑放 `internal/report/`。

## 目标

- 运营/策划可通过 `go run ./cmd/report <子命令>` 一键生成 Markdown 报表
- 覆盖三个维度：玩家活跃、Boss 战况、经济系统
- 支持日/周/月及自定义时间区间

## 文件结构

```
cmd/report/main.go              # CLI 入口、参数解析、调度
internal/report/
├── player_activity.go           # 活跃度查询
├── boss_stats.go                # Boss 战况查询
├── economy.go                   # 经济系统查询
├── types.go                     # 报表数据结构
├── query.go                     # Mongo 聚合辅助（时间分桶等）
└── formatter.go                 # 终端表格 + .md 输出
```

## CLI 接口

```bash
go -C backend run ./cmd/report daily          # 昨日日报
go -C backend run ./cmd/report weekly         # 上周周报
go -C backend run ./cmd/report monthly        # 上月月报
go -C backend run ./cmd/report custom --from 2026-04-01 --to 2026-04-30

--out <dir>     # .md 输出目录，默认 ./reports/
--no-md         # 只打印终端，不写 .md 文件
```

时间窗口推算：
- `daily` → 昨天 00:00 ~ 今天 00:00
- `weekly` → 上周一 00:00 ~ 本周一 00:00
- `monthly` → 上月 1 日 00:00 ~ 本月 1 日 00:00

所有子命令输出同一结构完整报表，仅时间窗口不同。

## 报表内容

### 一、玩家活跃

| 指标 | 数据来源 | 说明 |
|------|---------|------|
| 独立访问用户数 | `access_logs` (去重 client_ip) | 按日聚合 |
| 活跃玩家数 | `domain_events` (去重 nickname) | 有操作的真实玩家 |
| 新增玩家数 | `domain_events` (首次出现的 nickname) | 区间内首次出现 |
| 总请求量 | `access_logs` count | 区间内 API 调用总数 |
| P95 延迟 | `access_logs` latency_ms 聚合 | 服务端延迟 |
| Top 10 活跃玩家 | `domain_events` group by nickname | 按操作次数 |

### 二、Boss 战况

| 指标 | 数据来源 | 说明 |
|------|---------|------|
| Boss 生成次数 | `boss_history` count | 区间内 started_at |
| Boss 击杀次数 | `boss_history` count (status=defeated) | 被击杀的 Boss 数 |
| 击杀率 | 击杀/生成 | 百分比 |
| 总伤害量 | `boss_history.damage` 汇总 | 全量玩家伤害 |
| 平均存活时间 | defeated_at - started_at 平均 | 秒 |
| Top 10 伤害榜 | `boss_history.damage` 汇总 | 按玩家 nickname |
| Top 5 掉落装备 | `boss_history.loot` group by itemName | 出现次数 |

### 三、经济系统

| 指标 | 数据来源 | 说明 |
|------|---------|------|
| 商店总销售额 | `shop_purchase_logs` sum(price_gold) | 金币消耗 |
| 商店购买次数 | `shop_purchase_logs` count | 总成交 |
| 热销 Top 10 | `shop_purchase_logs` group by item_id | 按购买次数 |
| 任务完成数 | `task_claim_logs` count | 区间内领取数 |
| 任务奖励总金币 | `task_claim_logs.rewards.gold` sum | 系统产出 |
| 任务奖励总石头 | `task_claim_logs.rewards.stones` sum | 系统产出 |
| 任务奖励总天赋点 | `task_claim_logs.rewards.talent_points` sum | 系统产出 |
| 任务参与人数 (去重) | `task_claim_logs` distinct nickname | 做过任务的人数 |

## 输出格式

每一部分先输出终端表格，同时写入 `.md` 文件。Markdown 使用标题分层：

```markdown
# 运营报表 — 日报 2026-05-02

## 一、玩家活跃
| 指标 | 数值 |
|------|------|
| 独立访问用户 | 1,234 |
...

## 二、Boss 战况
...

## 三、经济系统
...
```

## 技术要点

- 使用 Mongo aggregation pipeline 做时间分桶和分组统计
- 时区统一使用 Asia/Shanghai (UTC+8)
- 时间戳均为 Unix 秒
- 遵循现有 cmd 工具模式：`main()` → `run()` → dispatch
- 不引入 `internal/report/` 外的依赖，Mongo 连接通过参数传入

## 非目标

- 不做实时数据（Redis 热数据）
- 不提供 Web 界面或 API
- 不接入外部 BI/图表
