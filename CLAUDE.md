# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 语言

- 所有回复、解释、代码注释、commit message 均使用中文
- 变量名、函数名等遵循项目约定，可保留英文

# RTK Priority

所有可由 `rtk` 代理、且与本仓库工作流重叠的命令，优先使用 `rtk`，以压缩输出、节省 token。

这条规则仍然服从现有 Tool Policy：只使用本地 shell、git、仓库文件和本地非破坏性检查。

当 `rtk` 与普通 shell 命令功能重叠时，优先使用 `rtk` 版本。

只有在 `rtk` 没有对应能力，或属于写操作、复杂交互操作时，才退回普通命令或内置编辑能力。

如果仓库已经提供统一入口，优先写“`rtk` + 项目入口命令”。

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

- `rtk go run ./cmd/server`
- `rtk go test ./...`
- `rtk go vet ./...`

Frontend 检查与运行：

- `rtk npm --prefix frontend run dev`
- `rtk npm --prefix frontend run build`
- `rtk npm --prefix frontend run test`
- `rtk vitest run`：仅在直接进入 `frontend/` 目录工作时使用

容器相关只在确实处理镜像或容器问题时使用 `rtk docker ...`。

## 项目概述

Redis Vote Wall — 一个 Vue 3 + Go (Hertz) + Redis + SSE 实时按钮计数墙项目。玩家点击按钮投票，装备/英雄系统提供成长数值，世界 Boss 提供协作战斗与掉落。

## 常用命令

开发前先安装前端依赖：

```bash
make deps
```

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

## 前置条件

后端需要 Consul 提供运行时配置，启动前必须设置环境变量：

```bash
export CONSUL_ADDR=http://127.0.0.1:8500
export CONSUL_CONFIG_KEY=vote-wall/dev
```

配置结构见 `README.md` 中 "Consul 配置" 章节。`backend/config.example.yaml` 和 `backend/config.yaml` 仅作参考，不直接参与运行时加载。测试也会走 Consul，`LoadTest()` 会拉取配置但不启动 watcher。

## 仓库目录

- `backend/`：Go + Hertz 服务、API、SSE、业务逻辑
- `frontend/`：Vue 3 + Vite 页面与前端测试
- `deploy/`：容器与 Nginx 部署配置
- `scripts/`：本地脚本
- `docs/`：一次性任务、设计、计划、记录
- `Makefile`：顶层开发、构建、校验入口
- `backend/cmd/server/main.go`：后端启动入口
- `frontend/src/main.js`：前端启动入口

## 架构概览

### 后端分层

```
cmd/server/main.go          # 启动入口：连线 Redis，组装依赖，启动 HTTP 服务
  ├── internal/config/       # Consul 配置加载、校验与热更新监听
  ├── internal/vote/         # 核心业务逻辑：按钮、装备、英雄、Boss、留言、公告
  │                           # Store 是 Redis 操作的统一入口
  ├── internal/events/       # SSE 实时推送：Hub 管理订阅者，Dispatcher 广播状态变更
  ├── internal/httpapi/      # HTTP 路由层（Hertz 框架）
  │                           # router.go 定义所有接口契约（Go interface），各 *_routes.go 注册路由
  ├── internal/ratelimit/    # 内存限流器（IP + 昵称双维度）
  ├── internal/nickname/     # 敏感词校验
  ├── internal/admin/        # 管理后台 Session 鉴权
  ├── internal/playerauth/   # 玩家 JWT 登录鉴权
  └── internal/oss/          # 阿里云 OSS 直传签名
```

### 前端分层

```
frontend/src/
  ├── main.js                # Vue 3 应用入口
  ├── App.vue                # 根组件，按路径区分 PublicPage / AdminPage
  ├── pages/
  │   ├── PublicPage.vue     # 玩家前台主页面
  │   ├── AdminPage.vue      # 管理后台容器
  │   ├── ProfilePage.vue    # 个人资料页
  │   ├── BattlePage.vue     # Boss 战页面
  │   ├── MessagesPage.vue   # 留言墙
  │   └── admin/             # 后台子模块
  ├── components/admin/      # 后台各 Tab 组件
  └── utils/                 # 前端工具模块（自动点击、Boss 状态、实时传输等）
```

前端不引入路由库，通过 `window.location.pathname` 判断当前页面（`/admin` → 管理后台，否则 → 玩家前台）。

### 实时通信

- 公共状态通过 **SSE** (`/api/events`) 推送，连接建立时先下发完整快照
- 个人状态通过 **SSE** (`/api/events/me`) 推送，需要已登录
- 后端 `events.Hub` 管理订阅者，`events.Dispatcher` 将业务变更广播给订阅者
- WebSocket 通道 (`/api/ws`) 作为降级/备选

### Redis 数据结构

所有数据存储在 Redis 中，无传统数据库。详见 `README.md` 中 "Redis 数据结构" 章节。按钮通过 `vote:buttons:index`（Sorted Set）维护显式索引，后台保存接口写入索引；低频兜底扫描 `vote:button:*` 手工新增键并补入索引。

### 关键设计约定

- **配置热更新**：Consul 配置变更后服务会主动 `os.Exit(0)`，由外部进程管理器（systemd/Docker）拉起新进程
- **前端代理**：Vite 开发服务器将 `/api` 代理到 `http://127.0.0.1:2333`，支持 WebSocket 代理
- **构建输出**：前端构建产物输出到 `backend/public/`，该目录不提交到仓库
- **Docker 单镜像**：镜像内置 nginx + Go 双进程，对外只暴露一个端口

## 文件读写策略

- 先用 `rg --files` 或 `rg -n` 缩小范围，再按需用 `sed -n`、`rg -n` 或定点打开文件
- 一次只读当前决策需要的少量文件，读到足够做决定就停止
- 禁止因为"不确定在哪"而连续展开多个目录、批量通读大量文件

## 变更策略

- 改动前先说明准备查看哪些入口文件、为什么看它们
- 只改与当前任务直接相关的文件，不顺手做无关重构、清理或样式统一
- 如果发现相邻问题但不影响当前任务，只记录，不顺手扩散修改

## 工具策略

- 本仓库内禁止使用 MCP、connector、app 和其他远程集成能力
- 默认只用本地 shell、git、仓库文件和本地非破坏性检查
- 所有可由 `rtk` 代理且与本仓库工作流重叠的命令，优先使用 `rtk`
- 优先级默认：先 `Makefile`，再语言原生命令
- 当 `rtk` 与普通 shell 命令功能重叠时，优先使用 `rtk` 版本；只有在 `rtk` 没有对应能力或属于写操作、复杂交互操作时，才退回普通命令

## Git 工作流

- 功能改动、接口改动、权限改动、数据模型改动时，统一切到 `dev` 分支开发
- 禁止直接在 `main` 上做这类开发改动
- 如果 `dev` 不存在，先从 `main` 创建 `dev`；如果已存在，先切到 `dev` 并确认当前工作区状态
- 完成后先在 `dev` 提交，再合并回 `main`；需要推送时，同时推送 `main` 和 `dev`

## 任务文档

- 一次性任务、设计说明、执行计划、阶段记录统一写入 `docs/`
- 文件名优先采用 `YYYY-MM-DD-主题.md` 形式，便于检索和归档
