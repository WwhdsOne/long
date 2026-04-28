# Purpose

`AGENTS.md` 只放长期协作规则，不放一次性开发需求、临时任务单或实现方案。

一次性任务说明写进 `docs/`，不要继续堆进这里。

# 语言：中文

- 所有回复、解释、代码注释、commit message 均使用中文。
- 变量名、函数名等遵循项目约定，可保留英文。
- 代码注释使用中文，除非项目已有英文注释惯例。
- 错误信息和日志的解读用中文说明。

# 项目概述

Redis Vote Wall — 一个 Vue 3 + Go (Hertz) + Redis + SSE 实时按钮计数墙项目。玩家点击按钮投票，装备/英雄系统提供成长数值，世界 Boss 提供协作战斗与掉落。

# Repository Index

这里只维护长期稳定的"目录职责 + 少量关键入口"，作为仓库导航索引。

不维护全量文件枚举，不把它写成一次性任务清单。

发现新模块时，只有当它已经成为长期稳定入口，再补进这里。

## 目录结构总览

```
/Users/Learning/web/long/          ← 项目根目录 (所有命令默认在此执行)
├── AGENTS.md                      ← 长期协作规则
├── CLAUDE.md                      ← Claude Code 指引
├── Makefile                       ← 顶层开发、构建、校验入口
├── Dockerfile
├── README.md
├── package.json
│
├── backend/                       ← Go + Hertz 后端 (模块路径: long)
│   ├── go.mod
│   ├── go.sum
│   ├── config.yaml
│   ├── config.example.yaml
│   ├── cmd/
│   │   ├── server/main.go         ← 启动入口
│   │   └── printcfg/main.go       ← 配置打印工具
│   ├── internal/
│   │   ├── vote/                  ← 核心业务 (Store 是 Redis 操作统一入口)
│   │   ├── httpapi/               ← HTTP 路由层 (Hertz)
│   │   ├── events/                ← SSE 实时推送 (Hub + Dispatcher)
│   │   ├── config/                ← Consul 配置加载
│   │   ├── ratelimit/             ← 内存限流器
│   │   ├── nickname/              ← 敏感词校验
│   │   ├── admin/                 ← 管理后台 Session 鉴权
│   │   ├── playerauth/            ← 玩家 JWT 登录鉴权
│   │   └── oss/                   ← 阿里云 OSS 直传签名
│   └── public/                    ← 前端构建产物 (不提交到仓库)
│
├── frontend/                      ← Vue 3 + Vite 前端
│   ├── package.json
│   ├── vite.config.js
│   ├── index.html
│   ├── src/
│   │   ├── main.js                ← 应用入口
│   │   ├── App.vue                ← 根组件
│   │   ├── pages/                 ← 页面组件
│   │   │   ├── PublicPage.vue     ← 玩家前台主页面
│   │   │   ├── AdminPage.vue      ← 管理后台容器
│   │   │   ├── BattlePage.vue     ← Boss 战页面
│   │   │   ├── TalentsPage.vue    ← 天赋系统页面
│   │   │   ├── ArmoryPage.vue     ← 个人资料页
│   │   │   ├── MessagesPage.vue   ← 留言墙
│   │   │   └── publicPageState.js ← 前台共享状态 (Pinia store)
│   │   ├── components/admin/      ← 后台各 Tab 组件
│   │   ├── utils/                 ← 工具模块
│   │   └── assets/                ← 静态资源
│   └── public/                    ← 公共静态文件
│
├── deploy/                        ← 容器与 Nginx 部署配置
├── docs/                          ← 一次性任务、设计、计划、记录
├── scripts/                       ← 本地脚本
└── pixel-assets/                  ← 像素资源与规范
```

## 关键目录与命令执行位置

| 操作类型 | 执行目录 | 命令示例 |
|---------|---------|---------|
| Go 后端测试/编译 | `backend/` | `cd backend && go test ./...` |
| Go 后端运行 | 项目根目录 | `go -C backend run ./cmd/server` |
| 前端开发/构建 | 项目根目录 | `npm --prefix frontend run dev` |
| Make 命令 | 项目根目录 | `make dev`, `make test`, `make build` |
| 前端依赖安装 | 项目根目录 | `npm --prefix frontend install` |

**重要**: Go 命令 (`go test`, `go vet`, `go build`) 必须在 `backend/` 目录下执行，否则会报错。

## 顶层

- `Makefile`：顶层开发、构建、校验入口
- `deploy/`：容器与 Nginx 部署配置
- `scripts/`：本地脚本
- `docs/`：一次性任务、设计、计划、记录

## 后端 `backend/`

Go + Hertz 服务，模块路径 `long`。

- `cmd/server/main.go`：启动入口，连线 Redis、组装依赖、启动 HTTP
- `cmd/printcfg/main.go`：配置打印工具
- `internal/vote/`：核心业务（按钮、装备、英雄、Boss、天赋、留言、公告、资源）。`Store` 是 Redis 操作统一入口
- `internal/httpapi/`：HTTP 路由层（Hertz）。`router.go` 定义接口契约（Go interface），各 `*_routes.go` 注册路由
- `internal/events/`：SSE 实时推送。`Hub` 管理订阅者，`Dispatcher` 广播状态变更
- `internal/config/`：Consul 配置加载、校验与热更新监听
- `internal/ratelimit/`：内存限流器（IP + 昵称双维度）
- `internal/nickname/`：敏感词校验
- `internal/admin/`：管理后台 Session 鉴权
- `internal/playerauth/`：玩家 JWT 登录鉴权
- `internal/oss/`：阿里云 OSS 直传签名

## 前端 `frontend/`

Vue 3 + Vite，不引入路由库，通过 `window.location.pathname` 判断页面。

- `src/main.js`：应用入口
- `src/App.vue`：根组件，按路径区分 PublicPage / AdminPage
- `src/pages/PublicPage.vue`：玩家前台主页面（战斗、天赋、资料、留言）
- `src/pages/AdminPage.vue`：管理后台容器
- `src/pages/BattlePage.vue`：Boss 战页面
- `src/pages/TalentsPage.vue`：天赋系统页面
- `src/pages/ArmoryPage.vue`：个人资料页
- `src/pages/MessagesPage.vue`：留言墙
- `src/pages/publicPageState.js`：前台共享状态（Pinia store）
- `src/components/admin/`：后台各 Tab 组件
- `src/utils/`：工具模块（自动点击、Boss 状态、实时传输等）

# 关键架构点

后端无传统数据库，所有数据存在 Redis。配置通过 Consul KV 拉取 YAML，配置变更后服务主动 `os.Exit(0)`，由外部进程管理器拉起新进程。

前端不引入路由库，通过 `window.location.pathname` 判断页面（`/admin` → 管理后台，否则 → 玩家前台）。

Vite 开发服务器将 `/api` 代理到 `http://127.0.0.1:2333`，支持 WebSocket 代理。

前端构建产物输出到 `backend/public/`，该目录不提交到仓库。Docker 单镜像内置 nginx + Go 双进程，对外只暴露一个端口。

## 实时通信

- 公共状态通过 **SSE** (`/api/events`) 推送，连接建立时先下发完整快照
- 个人状态通过 **SSE** (`/api/events/me`) 推送，需要已登录
- 后端 `events.Hub` 管理订阅者，`events.Dispatcher` 将业务变更广播给订阅者
- WebSocket 通道 (`/api/ws`) 作为降级/备选

## Redis 数据结构

所有数据存储在 Redis 中，无传统数据库。详见 `README.md` 中 "Redis 数据结构" 章节。按钮通过 `vote:buttons:index`（Sorted Set）维护显式索引，后台保存接口写入索引；低频兜底扫描 `vote:button:*` 手工新增键并补入索引。

# 命令速查

开发前先安装前端依赖：

```bash
make deps
```

后端需要 Consul 环境变量：

```bash
export CONSUL_ADDR=http://127.0.0.1:8500
export CONSUL_CONFIG_KEY=vote-wall/dev
```

配置结构见 `README.md` 中 "Consul 配置" 章节。`backend/config.example.yaml` 和 `backend/config.yaml` 仅作参考，不直接参与运行时加载。测试也会走 Consul，`LoadTest()` 会拉取配置但不启动 watcher。

启动完整开发环境（同时启动 Go 后端和 Vite 前端）：

```bash
make dev
```

单独启动后端或前端：

```bash
make backend-run        # Go 后端，默认 127.0.0.1:2333
make frontend-dev       # Vite 前端，默认 localhost:5173
```

构建前端产物到 `backend/public/`：

```bash
make build
```

运行测试：

```bash
make test               # 运行 Go 后端测试
npm --prefix frontend run test   # 运行 Vitest 前端测试
```

CI 校验（测试 + vet + 构建）：

```bash
make check
```

运行单个后端测试包：

```bash
go -C backend test ./internal/vote/...   # 测试 vote 包
go -C backend test -run TestXxx ./...    # 运行指定测试函数
```

# 测试特性

后端测试使用 `miniredis`（内存 Redis mock），不需要真实 Redis。测试通过 `newTestStore(t)` 创建带 mock Redis 的 Store 实例。

前端测试是静态分析测试：读取源文件为字符串，检查关键代码模式是否存在。不渲染组件，不依赖 happy-dom/jsdom。

# Read Policy

先搜索，后读取。

先用 `rg --files` 或 `rg -n` 缩小范围，再按需用 `sed -n`、`rg -n`、定点打开文件。

一次只读当前决策需要的少量文件；读到足够做决定就停止。

禁止因为"不确定在哪"而连续展开多个目录、批量通读大量文件或做无目标扫仓库。

# Change Policy

改动前先说明准备查看哪些入口文件、为什么看它们。

只改与当前任务直接相关的文件，不顺手做无关重构、清理或样式统一。

如果发现相邻问题但不影响当前任务，只记录，不顺手扩散修改。

# Tool Policy

本仓库内禁止使用 MCP、connector、app 和其他远程集成能力。

默认只用本地 shell、git、仓库文件和本地非破坏性检查。

优先使用本地命令：`rg`、`sed`、`git`、项目测试命令。

# RTK Priority

所有可由 `rtk` 代理、且与本仓库工作流重叠的命令，优先使用 `rtk`，以压缩输出、节省 token。

这条规则仍然服从现有 Tool Policy：只使用本地 shell、git、仓库文件和本地非破坏性检查。

当 `rtk` 与普通 shell 命令功能重叠时，优先使用 `rtk` 版本。

只有在 `rtk` 没有对应能力，或属于写操作、复杂交互操作时，才退回普通命令或内置编辑能力。

如果仓库已经提供统一入口，优先写"`rtk` + 项目入口命令"。

优先级默认是：先 `Makefile`，再语言原生命令。

文件与目录探索优先：

- `rtk find`
- `rtk tree`
- `rtk ls`
- `rtk grep`
- `rtk read`

Git 检查优先：

- `rtk git status`
- `rtk git diff`
- `rtk git log`

项目入口优先：

- `rtk make test`
- `rtk make check`

Go / 后端检查与运行：

- `rtk go run ./cmd/server`：需要 Consul 环境变量，否则会立即报错退出
- `rtk go test ./...`：必须在 `backend/` 目录下执行，从项目根目录会报错
- `rtk go vet ./...`：同上，必须在 `backend/` 目录下

Frontend 检查与运行：

- `rtk npm --prefix frontend run dev`
- `rtk npm --prefix frontend run build`
- `rtk npm --prefix frontend run test`

已知不可用的 rtk 命令：

- `rtk vitest run`：vitest 解析器会失败，改用 `npm --prefix frontend run test`

容器相关只在确实处理镜像或容器问题时使用 `rtk docker ...`。

不要在这里扩展与当前仓库无关的 RTK 命令族。

不要写 `rtk gh`、`rtk curl`，也不要补入 Python、Rust、.NET、Ruby、AWS、`kubectl`、`psql` 等与当前仓库无关的能力表。

# Git Workflow

遇到功能改动、接口改动、权限改动、数据模型改动时，统一切到 `dev` 分支开发。

禁止直接在 `main` 上做这类开发改动。

如果 `dev` 不存在，先从 `main` 创建 `dev`；如果已存在，先切到 `dev` 并确认当前工作区状态。

完成后先在 `dev` 提交，再合并回 `main`；需要推送时，同时推送 `main` 和 `dev`。

最终说明中明确给出 `dev` 提交和合并后的 `main` 提交。

# Task Docs

一次性任务、设计说明、执行计划、阶段记录统一写入 `docs/`。

新任务文档应使用清晰文件名，优先采用 `YYYY-MM-DD-主题.md` 形式，便于检索和归档。

不要把临时背景、单次约束或阶段性 TODO 回写到 `AGENTS.md`。
# AGENTS.md

Behavioral guidelines to reduce common LLM coding mistakes. Merge with project-specific instructions as needed.

**Tradeoff:** These guidelines bias toward caution over speed. For trivial tasks, use judgment.

## 1. Think Before Coding

**Don't assume. Don't hide confusion. Surface tradeoffs.**

Before implementing:
- State your assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them - don't pick silently.
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, stop. Name what's confusing. Ask.

## 2. Simplicity First

**Minimum code that solves the problem. Nothing speculative.**

- No features beyond what was asked.
- No abstractions for single-use code.
- No "flexibility" or "configurability" that wasn't requested.
- No error handling for impossible scenarios.
- If you write 200 lines and it could be 50, rewrite it.

Ask yourself: "Would a senior engineer say this is overcomplicated?" If yes, simplify.

## 3. Surgical Changes

**Touch only what you must. Clean up only your own mess.**

When editing existing code:
- Don't "improve" adjacent code, comments, or formatting.
- Don't refactor things that aren't broken.
- Match existing style, even if you'd do it differently.
- If you notice unrelated dead code, mention it - don't delete it.

When your changes create orphans:
- Remove imports/variables/functions that YOUR changes made unused.
- Don't remove pre-existing dead code unless asked.

The test: Every changed line should trace directly to the user's request.

## 4. Goal-Driven Execution

**Define success criteria. Loop until verified.**

Transform tasks into verifiable goals:
- "Add validation" → "Write tests for invalid inputs, then make them pass"
- "Fix the bug" → "Write a test that reproduces it, then make it pass"
- "Refactor X" → "Ensure tests pass before and after"

For multi-step tasks, state a brief plan:
```
1. [Step] → verify: [check]
2. [Step] → verify: [check]
3. [Step] → verify: [check]
```

Strong success criteria let you loop independently. Weak criteria ("make it work") require constant clarification.
