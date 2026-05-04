# 房间数量与显示名解耦 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把房间真实 ID 固定为按数量自动生成的连续数字字符串，并把房间显示名改为 Redis 持久化、后台可编辑的独立数据；同时确保切房、房间可加入判定和 Boss 循环语义不因显示名改造而回归。

**Architecture:** 后端配置层把 `room.ids` 收敛为 `room.count`，核心房间逻辑继续围绕稳定 `roomId` 运转，但房间列表与后台接口新增 `displayName` 读写。前端房间选择器和后台管理页统一消费 `displayName`，未命名时回退 `房间 N`。旧用户若持有非法房间值，继续由后端读取时回落默认房间，不做精确迁移。

**Tech Stack:** Go, Hertz, Redis, Vue 3, Vite, Vitest

---

## 文件结构与责任

### 后端配置与核心

- Modify: `backend/internal/config/config.go`
  - 把房间配置从 `IDs` 改为 `Count`
  - 解析 YAML `room.count`
  - 负责默认值、校验和标准化
- Modify: `backend/internal/config/config_test.go`
  - 覆盖 `room.count`、`default_room` 回退、配置兼容口径
- Modify: `backend/internal/core/room.go`
  - 基于 `count` 自动生成真实房间 ID
  - 为房间列表补 `DisplayName`
  - 提供 Redis 读取显示名和默认回退逻辑
- Modify: `backend/internal/core/store_test.go`
  - 覆盖房间 ID 自动生成、非法旧房间回退、显示名回退与覆盖

### 后端接口与后台管理

- Modify: `backend/internal/httpapi/admin_routes.go`
  - 挂载房间管理接口
- Create: `backend/internal/httpapi/admin_room_routes.go`
  - `GET /api/admin/rooms`
  - `PUT /api/admin/rooms/:roomId`
- Modify: `backend/internal/httpapi/router_test.go`
  - 覆盖后台房间读取、修改显示名、非法 `roomId` 与清空显示名

### 前端前台

- Modify: `frontend/src/components/RoomSelector.vue`
  - 优先显示 `displayName`
  - 无值时回退 `房间 N`
- Modify: `frontend/src/pages/BattlePage.vue`
  - 当前房间展示优先显示 `displayName`
- Modify: `frontend/src/components/RoomSelector.compact.test.js`
  - 覆盖显示名优先与默认回退
- Modify: `frontend/src/pages/BattlePage.roomMode.test.js`
  - 覆盖战斗页当前房间标题使用显示名

### 前端后台

- Modify: `frontend/src/pages/admin/useAdminPage.js`
  - 增加房间管理数据读取
- Modify: `frontend/src/pages/admin/useAdminPageActions.js`
  - 增加保存房间显示名动作
- Modify: `frontend/src/pages/AdminPage.vue`
  - 增加“房间管理”标签页入口
- Create: `frontend/src/components/admin/AdminRoomTab.vue`
  - 房间显示名编辑 UI
- Create: `frontend/src/components/admin/AdminRoomTab.layout.test.js`
  - 覆盖后台房间管理页结构
- Create: `frontend/src/pages/AdminPage.roomTab.test.js`
  - 覆盖后台标签页接线

### 文档与配置样例

- Modify: `backend/config.yaml`
  - 把 `room.ids` 改成 `room.count`
- Modify: `backend/config.example.yaml`
  - 同步新口径
- Modify: `docs/architecture/2026-05-05-房间数量与显示名解耦方案.md`
  - 若实现阶段收敛了接口与回退口径，回写当前有效规格
- Modify: `docs/architecture/2026-05-03-房间分线方案.md`
  - 增补“真实房间 ID 改为 count 生成、显示名外置”的现行口径说明
- Modify: `docs/README.md`
  - 若文档标题或入口说明调整，保持索引同步

---

### Task 1: 收敛配置为 `room.count`

**Files:**
- Modify: `backend/internal/config/config.go`
- Modify: `backend/internal/config/config_test.go`
- Modify: `backend/config.yaml`
- Modify: `backend/config.example.yaml`

- [ ] **Step 1: 写失败测试，定义 `room.count` 新口径**

在 `backend/internal/config/config_test.go` 新增或改造测试，覆盖：

```go
func TestLoadParsesRoomCount(t *testing.T) {
    cfg := mustLoadConfigFromYAML(t, `
room:
  enabled: true
  count: 3
  default_room: "2"
  switch_cooldown_seconds: 300
`)

    if cfg.Room.Count != 3 {
        t.Fatalf("expected room count 3, got %d", cfg.Room.Count)
    }
    if cfg.Room.DefaultRoom != "2" {
        t.Fatalf("expected default room 2, got %q", cfg.Room.DefaultRoom)
    }
}
```

- [ ] **Step 2: 运行测试，确认按旧实现失败**

Run: `go -C backend test ./internal/config -run 'TestLoadParsesRoomCount|TestLoadDefaultsInvalidRoomCount'`

Expected: FAIL，提示 `count` 未解析或 `RoomConfig` 结构不匹配。

- [ ] **Step 3: 最小实现配置解析**

修改 `backend/internal/config/config.go`：

- `RoomConfig` 从：

```go
type RoomConfig struct {
    Enabled        bool
    IDs            []string
    DefaultRoom    string
    SwitchCooldown time.Duration
}
```

改为：

```go
type RoomConfig struct {
    Enabled        bool
    Count          int
    DefaultRoom    string
    SwitchCooldown time.Duration
}
```

- `fileConfig.Room` 从：

```go
Room struct {
    Enabled               bool     `yaml:"enabled"`
    IDs                   []string `yaml:"ids"`
    DefaultRoom           string   `yaml:"default_room"`
    SwitchCooldownSeconds int      `yaml:"switch_cooldown_seconds"`
} `yaml:"room"`
```

改为：

```go
Room struct {
    Enabled               bool   `yaml:"enabled"`
    Count                 int    `yaml:"count"`
    DefaultRoom           string `yaml:"default_room"`
    SwitchCooldownSeconds int    `yaml:"switch_cooldown_seconds"`
} `yaml:"room"`
```

- 配置装配时写入 `Count`

- [ ] **Step 4: 补默认值与校验**

在 `backend/internal/config/config.go` 补最小规则：

- `count <= 0` 时回退到 `1`
- `default_room` 留给核心房间层二次标准化

同步修改 `backend/config.yaml` 与 `backend/config.example.yaml`：

```yaml
room:
  enabled: true
  count: 3
  default_room: "1"
  switch_cooldown_seconds: 300
```

- [ ] **Step 5: 运行测试，确认通过**

Run: `go -C backend test ./internal/config`

Expected: PASS

- [ ] **Step 6: 提交**

```bash
git add backend/internal/config/config.go backend/internal/config/config_test.go backend/config.yaml backend/config.example.yaml
git commit -m "feat: use room count config"
```

---

### Task 2: 核心房间模型按数量生成稳定 ID

**Files:**
- Modify: `backend/internal/core/room.go`
- Modify: `backend/internal/core/store_test.go`

- [ ] **Step 1: 写失败测试，覆盖房间 ID 自动生成与非法旧用户回退**

在 `backend/internal/core/store_test.go` 增加测试：

```go
func TestConfiguredRoomIDsGeneratedFromCount(t *testing.T) {
    store := newTestStoreWithRoomConfig(t, RoomConfig{
        Enabled: true,
        Count:   3,
        DefaultRoom: "1",
    })

    got := store.configuredRoomIDs()
    want := []string{"1", "2", "3"}
    if !reflect.DeepEqual(got, want) {
        t.Fatalf("expected %v, got %v", want, got)
    }
}

func TestResolvePlayerRoomFallsBackWhenStoredRoomInvalid(t *testing.T) {
    roomID, err := store.ResolvePlayerRoom(ctx, "阿明")
    if err != nil {
        t.Fatal(err)
    }
    if roomID != "1" {
        t.Fatalf("expected fallback room 1, got %q", roomID)
    }
}

func TestListRoomsKeepsJoinableRuleAfterCountRefactor(t *testing.T) {
    rooms, err := store.ListRooms(ctx, "阿明")
    if err != nil {
        t.Fatal(err)
    }
    room := findRoomInfo(t, rooms.Rooms, "2")
    if room.Joinable {
        t.Fatalf("expected room 2 to stay not joinable without active boss or cycle")
    }
}
```

- [ ] **Step 2: 运行测试，确认失败**

Run: `go -C backend test ./internal/core -run 'TestConfiguredRoomIDsGeneratedFromCount|TestResolvePlayerRoomFallsBackWhenStoredRoomInvalid'`

Expected: FAIL，提示 `Count` 逻辑未接通、房间列表仍依赖 `IDs`，或 `joinable` 相关断言未满足。

- [ ] **Step 3: 最小实现房间 ID 生成**

修改 `backend/internal/core/room.go`：

- `RoomConfig` 改为：

```go
type RoomConfig struct {
    Enabled        bool
    Count          int
    DefaultRoom    string
    SwitchCooldown time.Duration
}
```

- `normalizeRoomConfig` 改为基于 `Count` 生成 `1..count`
- `configuredRoomIDs()` 继续返回稳定 `[]string`
- `defaultRoomID()` 若超界，回退 `"1"`

- [ ] **Step 4: 确认旧用户非法房间值继续回默认房间**

保留并验证 `normalizeRoomID()` 当前语义：

```go
if s.isKnownRoom(roomID) {
    return roomID
}
return s.defaultRoomID()
```

这样老脏值不迁移，读取时直接回默认房间。

- [ ] **Step 5: 运行测试，确认通过且 `joinable` 规则未回归**

Run: `go -C backend test ./internal/core`

Expected: PASS

- [ ] **Step 6: 提交**

```bash
git add backend/internal/core/room.go backend/internal/core/store_test.go
git commit -m "feat: generate stable room ids from count"
```

---

### Task 3: 为房间列表增加 `displayName`

**Files:**
- Modify: `backend/internal/core/room.go`
- Modify: `backend/internal/core/store_test.go`

- [ ] **Step 1: 写失败测试，定义显示名覆盖与默认回退**

在 `backend/internal/core/store_test.go` 增加测试：

```go
func TestListRoomsUsesDisplayNameFromRedis(t *testing.T) {
    mustSetRoomDisplayName(t, store, "2", "高压线")

    rooms, err := store.ListRooms(ctx, "阿明")
    if err != nil {
        t.Fatal(err)
    }

    room := findRoomInfo(t, rooms.Rooms, "2")
    if room.DisplayName != "高压线" {
        t.Fatalf("expected display name 高压线, got %q", room.DisplayName)
    }
}

func TestListRoomsFallsBackToDefaultDisplayName(t *testing.T) {
    room := findRoomInfo(t, rooms.Rooms, "1")
    if room.DisplayName != "房间 1" {
        t.Fatalf("expected fallback display name, got %q", room.DisplayName)
    }
}
```

- [ ] **Step 2: 运行测试，确认失败**

Run: `go -C backend test ./internal/core -run 'TestListRoomsUsesDisplayNameFromRedis|TestListRoomsFallsBackToDefaultDisplayName'`

Expected: FAIL，提示 `DisplayName` 字段不存在或未赋值。

- [ ] **Step 3: 最小实现显示名字段与 Redis 读取**

修改 `backend/internal/core/room.go`：

- `RoomInfo` 新增：

```go
DisplayName string `json:"displayName"`
```

- 增加最小 helper：

```go
func (s *Store) roomNamesKey() string {
    return s.redisPrefix + "room:names"
}

func defaultRoomDisplayName(roomID string) string {
    return "房间 " + strings.TrimSpace(roomID)
}
```

- 在 `roomInfos()` 中对每个房间读取显示名；无值时回退默认文案

- [ ] **Step 4: 提供最小读写能力给后续后台接口复用**

在 `backend/internal/core/room.go` 加只做当前需求的最小方法：

```go
func (s *Store) GetRoomDisplayName(ctx context.Context, roomID string) (string, error)
func (s *Store) SetRoomDisplayName(ctx context.Context, roomID string, displayName string) error
```

规则：

- `roomID` 非法返回 `ErrRoomNotFound`
- `displayName` 为空时删除 Redis field
- `displayName` 读取层统一回退到 `房间 N`，避免接口契约分叉

- [ ] **Step 5: 运行测试，确认通过**

Run: `go -C backend test ./internal/core`

Expected: PASS

- [ ] **Step 6: 提交**

```bash
git add backend/internal/core/room.go backend/internal/core/store_test.go
git commit -m "feat: add room display names"
```

---

### Task 4: 后台增加房间显示名管理接口

**Files:**
- Modify: `backend/internal/httpapi/admin_routes.go`
- Create: `backend/internal/httpapi/admin_room_routes.go`
- Modify: `backend/internal/httpapi/router_test.go`

- [ ] **Step 1: 写失败测试，定义后台读取与保存接口**

在 `backend/internal/httpapi/router_test.go` 增加测试：

```go
func TestAdminRoomsReturnsDisplayNames(t *testing.T) {}
func TestAdminRoomUpdateSavesDisplayName(t *testing.T) {}
func TestAdminRoomUpdateClearsDisplayNameWhenEmpty(t *testing.T) {}
func TestAdminRoomUpdateRejectsUnknownRoom(t *testing.T) {}
```

测试断言：

- 已登录管理员才能访问
- `GET /api/admin/rooms` 返回 `id` + `displayName`
- `PUT /api/admin/rooms/2` 可更新 `"高压线"`
- `displayName: ""` 会清空自定义名

- [ ] **Step 2: 运行测试，确认失败**

Run: `go -C backend test ./internal/httpapi -run 'TestAdminRooms|TestAdminRoomUpdate'`

Expected: FAIL，提示路由不存在。

- [ ] **Step 3: 实现最小后台路由**

新增 `backend/internal/httpapi/admin_room_routes.go`：

- `GET /api/admin/rooms`
- `PUT /api/admin/rooms/:roomId`

接口依赖最小 store 能力：

```go
type adminRoomStore interface {
    ListRooms(context.Context, string) (core.RoomList, error)
    SetRoomDisplayName(context.Context, string, string) error
}
```

实现口径：

- 后台房间列表可直接复用 `ListRooms(ctx, "")`
- 只暴露 `id`、`displayName`，可按需带 `cycleEnabled`、`currentBossName`

- [ ] **Step 4: 在后台总路由注册**

修改 `backend/internal/httpapi/admin_routes.go`：

```go
registerAdminRoomRoutes(router, options)
```

- [ ] **Step 5: 运行测试，确认通过**

Run: `go -C backend test ./internal/httpapi`

Expected: PASS

- [ ] **Step 6: 提交**

```bash
git add backend/internal/httpapi/admin_routes.go backend/internal/httpapi/admin_room_routes.go backend/internal/httpapi/router_test.go
git commit -m "feat: add admin room display name routes"
```

---

### Task 5: 前台房间展示切到 `displayName`

**Files:**
- Modify: `frontend/src/components/RoomSelector.vue`
- Modify: `frontend/src/pages/BattlePage.vue`
- Modify: `frontend/src/components/RoomSelector.compact.test.js`
- Modify: `frontend/src/pages/BattlePage.roomMode.test.js`

- [ ] **Step 1: 写失败测试，定义显示名优先规则**

在 `frontend/src/components/RoomSelector.compact.test.js` 增加断言源码包含：

```js
expect(source).toContain("room?.displayName")
expect(source).toContain("房间 ")
```

在 `frontend/src/pages/BattlePage.roomMode.test.js` 增加断言：

```js
expect(battleSource).toContain("currentRoom.value?.displayName")
```

- [ ] **Step 2: 运行测试，确认失败**

Run: `bun --cwd=frontend run vitest run src/components/RoomSelector.compact.test.js src/pages/BattlePage.roomMode.test.js`

Expected: FAIL，提示仍只使用 `room.id`。

- [ ] **Step 3: 最小实现前台显示名逻辑**

修改 `frontend/src/components/RoomSelector.vue`：

```js
function defaultRoomLabel(roomId) {
  return `房间 ${String(roomId || '1').trim() || '1'}`
}

function roomLabel(room) {
  const displayName = String(room?.displayName || '').trim()
  if (displayName) return displayName
  return defaultRoomLabel(room?.id)
}
```

修改 `frontend/src/pages/BattlePage.vue`：

- 当前房间标题优先 `currentRoom.value?.displayName`
- 无值时回退 `房间 ${currentRoomId}`

- [ ] **Step 4: 运行测试，确认通过**

Run: `bun --cwd=frontend run vitest run src/components/RoomSelector.compact.test.js src/pages/BattlePage.roomMode.test.js`

Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add frontend/src/components/RoomSelector.vue frontend/src/pages/BattlePage.vue frontend/src/components/RoomSelector.compact.test.js frontend/src/pages/BattlePage.roomMode.test.js
git commit -m "feat: show room display names in public pages"
```

---

### Task 6: 后台增加房间管理页与保存动作

**Files:**
- Modify: `frontend/src/pages/admin/useAdminPage.js`
- Modify: `frontend/src/pages/admin/useAdminPageActions.js`
- Modify: `frontend/src/pages/AdminPage.vue`
- Create: `frontend/src/components/admin/AdminRoomTab.vue`
- Create: `frontend/src/components/admin/AdminRoomTab.layout.test.js`
- Create: `frontend/src/pages/AdminPage.roomTab.test.js`

- [ ] **Step 1: 写失败测试，定义后台房间页结构**

新建 `frontend/src/components/admin/AdminRoomTab.layout.test.js`：

```js
import { expect, test } from 'vitest'
import { readFileSync } from 'node:fs'

test('AdminRoomTab 提供房间显示名编辑表单', () => {
  const source = readFileSync(new URL('./AdminRoomTab.vue', import.meta.url), 'utf8')
  expect(source).toContain('displayName')
  expect(source).toContain('保存房间名')
})
```

同时新增 `frontend/src/pages/AdminPage.roomTab.test.js`，覆盖：

```js
expect(pageSource).toContain("admin.activeTab === 'rooms'")
expect(pageSource).toContain('AdminRoomTab')
```

- [ ] **Step 2: 运行测试，确认失败**

Run: `bun --cwd=frontend run vitest run src/components/admin/AdminRoomTab.layout.test.js src/pages/AdminPage.roomTab.test.js`

Expected: FAIL，组件或标签页不存在。

- [ ] **Step 3: 实现最小后台数据与动作**

修改 `frontend/src/pages/admin/useAdminPage.js`：

- 增加 `roomManagementList = ref([])`
- 增加 `fetchAdminRoomsMeta()`

修改 `frontend/src/pages/admin/useAdminPageActions.js`：

- 增加 `saveRoomDisplayName(roomId, displayName)`

```js
const response = await fetch(`/api/admin/rooms/${encodeURIComponent(roomId)}`, {
  method: 'PUT',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ displayName }),
})
```

- [ ] **Step 4: 实现最小后台 UI**

新增 `frontend/src/components/admin/AdminRoomTab.vue`：

- 循环展示房间列表
- 输入框编辑 `displayName`
- 保存按钮提交

修改 `frontend/src/pages/AdminPage.vue`：

- 新增 `rooms` 标签
- 接入 `AdminRoomTab`

- [ ] **Step 5: 运行测试，确认通过**

Run: `bun --cwd=frontend run vitest run src/components/admin/AdminRoomTab.layout.test.js src/pages/AdminPage.roomTab.test.js`

Expected: PASS

- [ ] **Step 6: 运行相关前端回归**

Run: `bun --cwd=frontend run vitest run src/pages/AdminPage.roomTab.test.js src/pages/AdminPage.shopTab.test.js src/pages/AdminPage.taskTab.test.js src/pages/admin/useAdminPageActions.equipmentGenerate.test.js`

Expected: PASS

- [ ] **Step 7: 提交**

```bash
git add frontend/src/pages/admin/useAdminPage.js frontend/src/pages/admin/useAdminPageActions.js frontend/src/pages/AdminPage.vue frontend/src/components/admin/AdminRoomTab.vue frontend/src/components/admin/AdminRoomTab.layout.test.js frontend/src/pages/AdminPage.roomTab.test.js
git commit -m "feat: add admin room display name editor"
```

---

### Task 7: 文档回写与全量验证

**Files:**
- Modify: `docs/architecture/2026-05-05-房间数量与显示名解耦方案.md`
- Modify: `docs/architecture/2026-05-03-房间分线方案.md`
- Modify: `docs/README.md`

- [ ] **Step 1: 同步当前规格与旧房间方案文档**

在以下文档同步最终落地口径：

- 真实房间 ID 不再通过 `room.ids` 手工配置
- 当前现行口径以 `room.count` 自动生成 `1..N`
- 房间显示名外置到 Redis 与后台管理
- `GET /api/rooms` 的 `displayName` 未命名时由后端直接回退 `房间 N`

涉及文件：

- `docs/architecture/2026-05-05-房间数量与显示名解耦方案.md`
- `docs/architecture/2026-05-03-房间分线方案.md`
- `docs/README.md`

- [ ] **Step 2: 跑后端相关测试**

Run: `go -C backend test ./internal/config ./internal/core ./internal/httpapi`

Expected: PASS

- [ ] **Step 3: 跑前端相关测试**

Run: `bun --cwd=frontend run vitest run src/components/RoomSelector.compact.test.js src/pages/BattlePage.roomMode.test.js src/components/admin/AdminRoomTab.layout.test.js src/pages/AdminPage.roomTab.test.js`

Expected: PASS

- [ ] **Step 4: 跑手工全量校验入口**

Run: `make check`

Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add docs/architecture/2026-05-05-房间数量与显示名解耦方案.md docs/architecture/2026-05-03-房间分线方案.md docs/README.md
git commit -m "docs: sync room count and display name model"
```

---

## 实施备注

- 现有工作区已有未提交改动：`backend/config.example.yaml`
- 实施时不要覆盖无关更改，先读清当前差异再合并
- 如果后端路由测试使用 mock store，需要补最小 mock 方法，避免为测试过度扩大接口面
- 如果 `ListRooms(ctx, "")` 在后台读取上引入歧义，可单独加 `ListAdminRooms(ctx)`，但只有在复用成本明显更高时才做

## 完成标准

- 配置文件只写 `room.count`
- 系统自动生成真实房间 ID `1..N`
- 前台、战斗页、后台均显示 `displayName`
- 后台可编辑、清空房间显示名
- 非法旧用户房间值自动回默认房间
- 所有相关测试通过，`make check` 通过
