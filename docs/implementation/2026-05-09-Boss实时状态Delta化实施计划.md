# Boss 实时状态 Delta 化实施计划

## 目标

把当前战斗高频下行从“重复发送完整 Boss”改成：

- 首包全量 `snapshot`
- 高频 `boss_delta`
- 版本跳号自动 `sync_request`

本次只处理 `Boss` 高频实时协议，不继续改 `room_state`。

## 实施边界

本次改动涉及：

- `backend/internal/realtimepb/realtime.proto`
- `backend/internal/httpapi/realtime_proto.go`
- `backend/internal/httpapi/realtime_socket.go`
- `backend/internal/events/*`
- `frontend/src/utils/realtimeProto.js`
- `frontend/src/utils/realtimeTransport.js`
- `frontend/src/pages/publicPageState.js`

不涉及：

- Redis 存储格式
- HTTP 低频接口
- `room_state` 协议拆分
- 新旧协议兼容期

## 关键原则

- 先写失败测试，再写实现
- `click_ack` 即时手感不动
- 同批发布，不做双协议并行
- 先保证正确，再收包体

## 分阶段任务

### 第一阶段：协议骨架与测试基线

1. 后端补失败测试：
   - `snapshot` 能携带 `boss static + boss runtime + version`
   - `public_state` 事件会被编码成新的 `boss_delta`
   - `room_state` 仍保持现状
2. 前端补失败测试：
   - `boss_delta` 二进制帧可解码
   - 版本连续时能 merge
   - 版本跳号时能触发 `sync_request`
3. 确认失败原因来自缺少新协议，而不是测试本身错误

### 第二阶段：后端首包与 Delta 下行

1. 扩展 `realtime.proto`
   - 新增 `BossStatic`
   - 新增 `BossRuntime`
   - 新增 `BossPartRuntimeDelta`
   - 新增 `BossDelta`
2. 调整 `snapshot`
   - 不再只靠完整 `core.Snapshot.Boss`
   - 明确携带 `bossId/version/static/runtime`
3. 调整 `public_state -> boss_delta`
   - 只发 `bossId/version/currentHp/status/变化部位`
4. 保留 `click_ack`
   - 继续携带即时伤害与个人态
   - 不让它承担完整 Boss 广播

### 第三阶段：版本号与重同步

1. 服务端为公共 Boss 动态态维护单调递增版本号
2. 普通点击、挂机、流血 tick、Boss 切换都推进版本号
3. 复用现有 `sync_request`
   - 版本跳号
   - `bossId` 不一致
   - 本地缺少静态结构
   都走完整 `snapshot`

### 第四阶段：前端本地 merge

1. 前端状态拆成：
   - `bossStaticById`
   - `bossRuntime`
   - `bossVersion`
2. 首包建立基线
3. `boss_delta` 连续时局部 merge
4. 旧包丢弃
5. 跳号自动 `sync_request`

### 第五阶段：回归验证

必须验证：

- WebSocket 首包正常
- 普通点击手感不退化
- Boss 部位掉血与击碎正常
- Boss 切换正常
- 跳号后自动自愈
- 现有 `room_state`、`click_ack`、`online_count` 不回归

## 测试顺序

### 后端

- `go -C backend test ./internal/httpapi -run Realtime`
- `go -C backend test ./internal/events`
- `go -C backend test ./...`

### 前端

- `bun --cwd=frontend run test realtimeTransport`
- `bun --cwd=frontend run test PublicPage.clickResponse`
- `bun --cwd=frontend run test BattlePage.talentFx`
- `bun --cwd=frontend run test`

## 风险点

### 1. 版本号推进点漏掉

后果：

- 前端误判跳号
- 频繁全量重同步

控制：

- 先补覆盖点击、挂机、流血 tick、Boss 切换的测试

### 2. Delta 生成不完整

后果：

- 页面局部状态卡住
- 某些部位血量不同步

控制：

- 用现有 `partStateDeltas` 逻辑收敛生成路径
- 让 delta 测试直接断言具体部位字段

### 3. 前端 merge 写错

后果：

- 老包覆盖新包
- 新 Boss 混入旧 Boss 部位

控制：

- 明确 `bossId/version` 优先级
- 测试覆盖连续、旧包、跳号三类场景

## 完成标准

只有同时满足以下条件，才算本次实施完成：

- 高频 `Boss` 下行不再重复发送完整静态结构
- 前端能基于首包 + delta 正常战斗
- 跳号后能自动重同步
- 后端全量测试通过
- 前端相关测试通过
