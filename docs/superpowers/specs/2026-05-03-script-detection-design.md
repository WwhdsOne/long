# 脚本自动点击检测方案

## 背景

当前限流器 `ratelimit.Limiter` 只在短窗口（如 5s/80次）内检测突发超限。脚本可以将速率卡在阈值以下，长期持续运行而不被拦截。人类在 5 秒内可以达到 16 cps，但无法在 10 分钟或更长时间内保持。

## 目标

- 短窗口保持宽松（5s/80次），允许人类爆发
- 长窗口收紧（10分钟/600次、1小时/2400次），抓住持续高速的脚本
- 递增封禁：首次 10 分钟，每次再犯翻倍
- 后台提供解封接口，处理误封

## 方案

采用多窗口滑动限流（方案 A），在现有 `ratelimit.Limiter` 上扩展，不依赖 Redis/Mongo。

### 配置

```yaml
rate_limit:
  limit: 80
  window_ms: 5000              # 短窗口 5s
  blacklist_ms: 600000         # 首次封禁 10 分钟
  blacklist_multiplier: 2.0    # 再犯封禁时长翻倍
  offense_decay_ms: 86400000   # 24h 内无再犯自动重置计数
  medium:
    limit: 600
    window_ms: 600000          # 10 分钟
  long:
    limit: 2400
    window_ms: 3600000         # 1 小时
```

### 数据结构

```go
type Config struct {
    Limit               int
    Window              time.Duration
    BlacklistDuration   time.Duration
    BlacklistMultiplier float64
    OffenseDecay        time.Duration
    Medium              SubWindowConfig
    Long                SubWindowConfig
    Now                 func() time.Time
}

type SubWindowConfig struct {
    Limit  int
    Window time.Duration
}

type clientState struct {
    hits         []time.Time
    blockedUntil time.Time
    offenseCount int
    lastOffense  time.Time
}
```

### 判定逻辑

每次点击：

1. 清理当前时间之前超出最长窗口（1h）的旧时间戳
2. 在三个窗口内分别计数：
   - 短：`hits` 中 `now - 5s` 内的数量
   - 中：`hits` 中 `now - 10min` 内的数量
   - 长：`hits` 中 `now - 1h` 内的数量
3. 任一超限 → 封禁
4. 封禁时长：`blacklist_ms × multiplier^(offenseCount - 1)` 毫秒
5. 检查 `lastOffense`，如距上次违规超过 `offense_decay_ms`，重置 `offenseCount`
6. 每次违规 `offenseCount++`，记录 `lastOffense`

### 管理后台解封接口

```
POST /api/admin/players/unban
Authorization: <admin session cookie>
Body: {"nickname": "xxx"}
```

直接清掉该客户端在内存限流器中的 `blockedUntil` 和 `offenseCount`。

### 改动文件

| 文件 | 改动 |
|------|------|
| `internal/config/config.go` | 新增 `MediumWindowConfig`、`LongWindowConfig`，扩展 `RateLimitConfig` |
| `internal/ratelimit/limiter.go` | 多窗口判定、递增封禁、解封方法 |
| `internal/ratelimit/limiter_test.go` | 补充多窗口和递增封禁测试 |
| `internal/httpapi/admin_resource_routes.go` | 新增 `POST /api/admin/players/unban` |
| `internal/httpapi/router.go` | 在 `ButtonStore` 接口暴露限流器的 `Unblock` 能力 |
| `cmd/server/main.go` | 更新 `ratelimit.Config` 初始化 |

### 不涉及的改动

- Redis Lua 脚本不变
- MongoDB 不新增集合
- 前端不新增页面（解封在后台管理页面使用现有管理框架）

## 后续扩展

- 方案 B（间隔方差检测）可独立叠加，在 `hits` 中取最近 100 个时间戳计算标准差即可
- 如需持久化封禁记录，可在封禁触发时写一条 `domain_event` 到 Mongo
