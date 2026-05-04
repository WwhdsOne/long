# 项目协作规则

本文件只放长期稳定的协作规则、仓库导航和工作流约束。

- 不放一次性任务单、临时 TODO、阶段性实现方案。
- 一次性设计、计划、记录统一写入 `docs/`。
- `AGENTS.md` 与 `CLAUDE.md` 必须保持正文完全一致。

## 语言

- 所有回复、解释、代码注释、commit message 均使用中文。
- 变量名、函数名、类型名遵循项目既有约定，可保留英文。
- 错误信息和日志的解读使用中文说明。

## 项目概述

Redis Vote Wall 是一个 `Vue 3 + Vite + Go(Hertz) + Redis + SSE` 的实时按钮计数墙项目，已扩展出装备、英雄、Boss、天赋、任务、留言墙、后台管理和 Mongo 冷数据链路。

## 仓库导航

### 顶层

- `AGENTS.md` / `CLAUDE.md`
  - 长期协作规则。两份文件正文必须一致。
- `README.md`
  - 当前有效的项目入口文档。
- `Makefile`
  - 本地开发、构建、测试、hook 安装入口。
- `Dockerfile`
  - 当前容器镜像入口，基于 `long-basic:latest` 构建发布镜像。
- `deploy/`
  - 容器入口脚本和 nginx 配置。
- `docs/`
  - 一次性任务文档、设计、计划、总结、归档。

### 后端 `backend/`

- 模块路径：`long`
- 启动入口：`cmd/server/main.go`
- 配置打印工具：`cmd/printcfg/main.go`
- 核心业务：`internal/vote/`
- HTTP 路由层：`internal/httpapi/`
- 实时推送：`internal/events/`
- Consul 配置：`internal/config/`
- 管理后台鉴权：`internal/admin/`
- 玩家登录鉴权：`internal/playerauth/`
- OSS 签名：`internal/oss/`
- 冷数据与日志存储：`internal/mongostore/`
- 前端构建产物目录：`public/`

### 前端 `frontend/`

- 应用入口：`src/main.js`
- 根组件：`src/App.vue`
- 页面目录：`src/pages/`
- 后台组件：`src/components/admin/`
- 工具模块：`src/utils/`
- 静态资源：`src/assets/`、`public/`

## 当前架构事实

- 前端不引入路由库，通过 `window.location.pathname` 区分页面。
- 公共状态通过 `SSE /api/events` 推送，个人状态通过 `SSE /api/events/me` 推送。
- Redis 负责热数据；Mongo 已启用，用于冷数据、任务、日志和归档。
- 运行时配置通过 Consul KV 拉取 YAML；配置变更后服务会主动退出，由外部拉起新进程。
- 前端构建产物输出到 `backend/public/`，由容器内 nginx + Go 服务统一承载。
- GitHub Actions 当前只负责构建 release 包；服务器端使用 `docker compose build` / `docker compose up -d` 部署。

## 命令与执行位置

### 常用命令

```bash
make deps
make dev
make backend-run
make frontend-dev
make build
make test
make check
bun --cwd=frontend run test
make hooks-install
```

### 约束

- Go 命令必须在 `backend/` 目录执行，或使用 `go -C backend ...`。
- 前端命令默认在仓库根目录通过 `bun --cwd=frontend ...` 执行。
- `make check` 是本地手动全量校验入口，包含：
  - 后端测试
  - `go vet`
  - 前端测试
  - 前端构建

## 提交前校验

仓库使用 `lefthook` 管理本地 `pre-commit`。

- 安装命令：`make hooks-install`
- 当前 `pre-commit` 会执行：
  - 第一阶段前端更新：`bun --cwd=frontend install`
  - 第二阶段后端更新：`go -C backend mod tidy`
  - `go -C backend fix ./...`
  - `go -C backend test ./...`
  - `go -C backend vet ./...`
  - `bun --cwd=frontend run test`

## 部署约束

- GitHub Actions 产出 `release.tar.gz`，其中包含：
  - `Dockerfile`
  - `.dockerignore`
  - `docker-compose.yml`
  - `backend/long`
  - `backend/public/`
  - `deploy/entrypoint.sh`
  - `deploy/nginx.container.conf`
- 后端发布二进制当前使用：

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go -C backend build -trimpath -buildvcs=false -ldflags "-w -s -buildid=" -o long ./cmd/server
```

- 服务器端基础镜像要求为本地已存在的 `long-basic:latest`。

## 测试特性

- 后端测试使用 `miniredis`，不依赖真实 Redis。
- 前端测试以静态分析测试为主，读取源码字符串检查关键模式，不依赖浏览器渲染环境。

## Read Policy

- 先搜索，后读取。
- 先用 `rg --files`、`rg -n` 缩小范围，再按需用 `sed -n`、定点读取。
- 一次只读当前决策需要的少量文件，读到足够做决定就停止。
- 不要因为“不确定在哪”而连续展开多个目录或无目标扫仓库。

## Change Policy

- 改动前先说明准备查看哪些入口文件，以及为什么看它们。
- 只改与当前任务直接相关的文件，不顺手做无关重构、清理或风格统一。
- 如果发现相邻问题但不影响当前任务，只记录，不扩散修改。

## Tool Policy

- 本仓库内禁止使用 MCP、connector、app 和其他远程集成能力。
- 默认只用本地 shell、git、仓库文件和本地非破坏性检查。
- 优先使用本地命令：`rg`、`sed`、`git`、项目测试命令。

## RTK Priority

- 所有可由 `rtk` 代理、且与本仓库工作流重叠的命令，优先使用 `rtk`，以压缩输出、节省 token。
- 这条规则仍然服从 Tool Policy。
- 当 `rtk` 与普通 shell 命令重叠时，优先使用 `rtk` 版本。
- 若 `rtk` 没有对应能力，或属于写操作、复杂交互操作，再退回普通命令或内置编辑能力。
- 优先级默认是：先 `Makefile`，再语言原生命令。

优先探索命令：

- `rtk find`
- `rtk tree`
- `rtk ls`
- `rtk grep`
- `rtk read`

优先 Git 检查：

- `rtk git status`
- `rtk git diff`
- `rtk git log`

项目入口优先：

- `rtk make test`
- `rtk make check`
- `rtk make build`

## Git Workflow

- 遇到功能改动、接口改动、权限改动、数据模型改动时，统一切到 `dev` 分支开发。
- 禁止直接在 `main` 上做这类开发改动。
- 如果 `dev` 不存在，先从 `main` 创建；如果已存在，先切到 `dev` 并确认工作区状态。
- 完成后先在 `dev` 提交，再合并回 `main`；需要推送时，同时推送 `main` 和 `dev`。
- 最终说明中明确给出 `dev` 提交和合并后的 `main` 提交。

## docs 约束

- 一次性任务、设计说明、执行计划、阶段记录统一写入 `docs/`。
- 新文档优先放入 `docs/` 的现有分类目录，不要把大量新文件继续堆在 `docs/` 顶层。
- `docs/archive/` 下的文档不作为当前实现依据。
- `docs/README.md` 是 `docs/` 的总索引；新增正式文档时应同步更新入口。

## 工程原则

### Think Before Coding

- 明确假设，不要静默猜测。
- 有多种解释时，先把分歧讲清楚，不要悄悄选一种。
- 如果存在更简单的实现路径，要明确指出。

### Simplicity First

- 用满足需求的最小代码解决问题。
- 不为未被请求的扩展性、可配置性或抽象付出复杂度。
- 不为不可能发生的场景补复杂错误处理。

### Surgical Changes

- 每一处改动都应直接服务于当前任务。
- 只清理自己引入的冗余，不顺手清扫历史遗留问题。
- 若你的改动导致 import、变量、函数变成未使用状态，应一并清理。

### Goal-Driven Execution

- 把任务转换成可验证目标。
- 能写测试就先用测试或已有校验命令证明行为。
- 完成前至少说明你做了哪些验证；没有验证时明确说明原因。
