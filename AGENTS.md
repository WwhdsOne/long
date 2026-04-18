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

### Python

```bash
rtk pytest
rtk ruff check / format
rtk mypy .
rtk pip install / list
```

### Rust

```bash
rtk cargo build / test / clippy / fmt
```

### Go

```bash
rtk go build / test / vet
rtk golangci-lint run
```

### .NET / Ruby

```bash
rtk dotnet build / test
rtk rspec / rake / rubocop
```

### Infrastructure

```bash
rtk aws <service> <command>
rtk docker ps / logs / compose
rtk kubectl get / describe / logs
rtk psql <query>
```

### Meta Commands

```bash
rtk gain
rtk gain --history
rtk discover
```

## 项目结构

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

- 开发：`npm run dev`
- 前端构建：`npm run build`
- 后端测试：`npm test`
- 后端单独运行：`go -C backend run ./cmd/server`
- 前端单独构建：`npm --prefix frontend run build`
- 后端单独测试：`go -C backend test ./...`

## 修改约定

- 前端改动尽量保持现有“投票墙”视觉风格，不要回退成后台面板感
- 公共页和 `/admin` 后台是两套页面体验；改公共页时优先维护活动页氛围，改后台时再考虑信息密度和管理效率
- 后端逻辑优先放在 `backend/internal/` 下对应职责目录
- 按钮在平时必须保持可点击；不要把“无活动 Boss”做成点击错误
- 装备效果在平时点击和 Boss 战斗期间都要生效；如果调整点击结算，要同时检查普通计票和 Boss 伤害两条链路
- 第一版只支持单个当前世界 Boss；如果改这个前提，要同步检查 Redis 键设计、后台接口和前端展示
- 后台能力当前覆盖 Boss、装备、掉落池、按钮和玩家概览；新增后台模块时优先沿用现有 `/api/admin/*` 约定
- 如果改了接口、配置或部署方式，要同步更新 `README.md`
