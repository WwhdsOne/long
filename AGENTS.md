# AGENTS

## 语言约定

- 所有回复、解释、代码注释、commit message 均使用中文。
- 变量名、函数名等遵循项目约定，可保留英文。
- 代码注释使用中文，除非项目已有英文注释惯例。
- 错误信息和日志的解读用中文说明。

## RTK 工具约定

- 所有命令前缀使用 `rtk`，优先使用 RTK 替代内置同类工具以节省 token。
- 功能重叠时，RTK 优先于内置工具。
- 仅当 RTK 无对应命令时，才使用内置工具或其他命令，例如写文件、补丁编辑、复杂多步操作。

### 优先级规则

| Claude Code 内置工具 | RTK 替代命令 | 说明 |
| --- | --- | --- |
| `Glob`（文件搜索） | `rtk find` / `rtk tree` / `rtk ls` | 压缩目录输出 |
| `Grep`（内容搜索） | `rtk grep` | 按文件分组、截断、去空白 |
| `Read`（读文件） | `rtk read` | 智能过滤，省去无用行 |
| `Bash` + `git` | `rtk git` | 紧凑 git 输出 |
| `Bash` + `gh` | `rtk gh` | 紧凑 GitHub CLI 输出 |
| `Bash` + `curl` | `rtk curl` | 自动检测 JSON，schema-only 模式 |
| `Bash` + `diff` | `rtk diff` | 仅显示变更行 |

### Node.js / Frontend

```bash
rtk pnpm install / add / run build
rtk npm run <script>
rtk npx tsc / eslint / prisma
rtk vitest run
rtk next build
rtk lint
rtk prettier --check .
rtk playwright test
rtk tsc --noEmit
```

### Go

```bash
rtk go build / test / vet
rtk golangci-lint run
```

## 项目结构

- `Makefile` 是项目顶层命令入口，统一提供开发、构建、测试命令
- `frontend/` 是 Vue 前端
- `frontend/src/App.vue` 只负责根据路径切换公共页和后台页
- `frontend/src/pages/PublicPage.vue` 是投票墙主页面，同时承载世界 Boss、背包、装备栏和排行榜展示
- `frontend/src/pages/AdminPage.vue` 是 `/admin` 管理后台页面
- `frontend/src/style.css` 是前台与后台共享样式入口
- `backend/` 是 Go 后端，入口是 `backend/cmd/server/main.go`
- `backend/internal/httpapi/router.go` 负责注册公开接口、装备接口和后台接口
- `backend/internal/vote/store.go` 是投票、Boss、装备、掉落的核心状态与业务实现
- `backend/internal/vote/admin.go` 放后台聚合读写逻辑
- `backend/internal/admin/` 放后台鉴权相关逻辑
- `backend/public/` 是前端构建产物目录，由后端静态托管，但不提交到仓库

## 配置约定

- 运行时和测试都通过 `CONSUL_ADDR` 与 `CONSUL_CONFIG_KEY` 从 Consul 读取配置
- 不要重新引入本地 YAML 或 `.env` 作为后端配置来源
- 如果改了配置结构，要同步更新 `backend/internal/config/config.go` 里的 YAML 解析结构和 `README.md` 里的 Consul 配置示例
- 当前 Consul 配置除了 Redis、HTTP、昵称、限流等已有字段外，还必须包含：
  - `admin.username`
  - `admin.password`
  - `admin.session_secret`
- 不要把后台账号配置硬编码回路由或前端

## 常用命令

- 安装依赖：`make deps`
- 开发：`make dev`
- 前端构建：`make build`
- 后端测试：`make test`
- 后端单独运行：`make backend-run`
- 前端单独运行：`make frontend-dev`
- 前端单独构建：`make frontend-build`
- 后端单独测试：`make backend-test`
- Go 静态检查：`make backend-vet`
- CI 校验：`make check`

## 修改约定

- 前端改动尽量保持现有“投票墙”视觉风格，不要回退成后台面板感
- 公共页和 `/admin` 后台是两套页面体验；改公共页时优先维护活动页氛围，改后台时再考虑信息密度和管理效率
- 后端逻辑优先放在 `backend/internal/` 下对应职责目录
- 较大的功能、接口、规则或管理后台改动，要同步在 `docs/` 下新增或更新对应说明文档，文件名优先使用 `日期-主题.md`
- 按钮在平时必须保持可点击；不要把“无活动 Boss”做成点击错误
- 装备效果在平时点击和 Boss 战斗期间都要生效；如果调整点击结算，要同时检查普通计票和 Boss 伤害两条链路
- 第一版只支持单个当前世界 Boss；如果改这个前提，要同步检查 Redis 键设计、后台接口和前端展示
- 后台能力当前覆盖 Boss、装备、掉落池、按钮和玩家概览；新增后台模块时优先沿用现有 `/api/admin/*` 约定
- 如果改了接口、配置或部署方式，要同步更新 `README.md`


## 🧱 文件拆分与模块化强约束（非常重要）

在生成任何代码时，必须遵循以下规则，禁止输出“单文件巨型实现”。

### 1. 文件大小限制（强制）
- 单个文件 **不允许超过 500 行**
- 超过 500 行视为严重违规

如果逻辑复杂，必须主动拆分，而不是继续往一个文件里追加代码

---

### 2. 必须按职责拆分（后端）

禁止将所有逻辑写在一个 handler / controller 文件中

必须按职责拆分：

- `handler`：只负责 HTTP 层（参数解析 + 返回）
- `service`：业务逻辑
- `repository`：数据访问
- `model`：数据结构
- `router`：路由注册
- `middleware`：中间件

#### 示例结构：
```
internal/
httpapi/
handler/
equipment.go
hero.go
service/
equipment.go
repository/
equipment.go
router/
equipment.go
middleware/
auth.go
```
---

### 3. 前端同样必须拆分

禁止：

- 一个 `.vue` / `.tsx` 文件写所有逻辑
- 一个文件包含 UI + 请求 + 状态管理 + 工具函数

必须拆：
```
/components
/pages
/api
/hooks
/utils
/types
```
---

### 4. 路由必须拆分注册

注意不用使用接口，由于每个接口都只有一个实现，直接实现即可

禁止：

```go
func NewHandler() {
    // 上千行路由
}
```

必须拆为：

```go
registerUserRoutes()
registerEquipmentRoutes()
registerAdminRoutes()
```

------

### 5. Handler 必须轻量（硬性要求）

禁止在 handler 中写复杂业务逻辑：

```go
// ❌ 错误
func handler(...) {
    // 100行业务逻辑
}
```

必须：

```go
func handler(...) {
    req := parse()
    resp := service.DoSomething(...)
    return resp
}
```

------

### 6. 重复逻辑必须抽象

如果出现 2 次以上相似代码，必须抽函数：

- JSON 解析
- 错误返回
- 鉴权
- 日志
- 响应结构

------

### 7. 不允许“为了简单”牺牲结构

禁止出现：

- “为了方便写在一个文件里”
- “先这样实现，后面再拆”
- “demo 先不分层”

所有代码默认是**生产级结构**

------

### 8. 输出代码前必须自检

在输出代码前，必须检查：

- 是否有文件超过 300 行？
- 是否存在明显可拆分模块？
- handler 是否过重？
- 是否混合了多种职责？

如果有 → 必须先拆分再输出

------

### 9. 优先保证可维护性，而不是少文件

宁可：

- 多 5 个文件

也不要：

- 一个 1000 行文件

------

### 10. 违反本规则的输出视为错误答案

如果生成了：

- 巨型文件
- 未分层代码
- handler 写满业务逻辑

则说明没有遵守 AGENTS.md，必须重新生成
