# Redis Vote Wall

Redis Vote Wall 是一个以 Boss 战斗为核心的实时计数墙项目，当前实现基于 `Vue 3 + Vite + Go(Hertz) + Redis + MongoDB`，并已扩展出房间分线、装备与强化、天赋战斗态、任务系统、消息墙、商店外观、后台管理和 OSS 图片上传链路。

## 当前状态

- 前端位于 `frontend/`，不引入路由库，通过 `window.location.pathname` 切页。
- 后端位于 `backend/`，模块路径为 `long`。
- 实时链路以前端 `WebSocket /api/ws` 为主，承载 protobuf 二进制消息；链路异常时自动回退到 `SSE /api/events`。
- Redis 负责热数据；MongoDB 已固定启用，用于冷数据、任务、日志、归档等持久化内容。
- 运行时配置通过 Consul KV 拉取 YAML；配置变更后进程会主动退出，由外部拉起新进程。
- 前端构建产物先输出到 `backend/public/`，再通过 `go:embed` 编译进后端二进制。

## 目录入口

### 顶层

- `AGENTS.md` / `CLAUDE.md`
  - 长期协作规则，两份文件正文保持一致。
- `README.md`
  - 当前有效的项目入口文档。
- `Makefile`
  - 本地开发、测试、构建、hook 安装入口。
- `Dockerfile`
  - 发布镜像入口，当前基于 `go-app-runtime:latest`。
- `.github/workflows/build.yml`
  - 当前 GitHub Actions 构建与部署流程。
- `docs/`
  - 一次性方案、实施记录、开发参考、阶段总结、归档。
- `scripts/`
  - 本地运维与数据辅助脚本。
- `pixel-assets/`
  - 像素资源与视觉素材。

### 后端 `backend/`

- `cmd/server/main.go`
  - 服务启动入口。
- `cmd/printcfg/main.go`
  - 配置打印工具。
- `cmd/backfillbosskills/main.go`
  - Boss 击杀统计回填工具。
- `cmd/checkbosskills/main.go`
  - Boss 击杀统计检查工具。
- `cmd/local_click_latency/main.go`
  - 本地点击时延测试工具。
- `internal/httpapi/`
  - HTTP 路由、WebSocket、静态资源与接口入口。
- `internal/events/`
  - SSE 订阅中心与事件广播。
- `internal/core/`
  - 主要玩法、状态与存储逻辑。
- `internal/config/`
  - Consul 配置加载。
- `internal/mongostore/`
  - MongoDB 冷数据与日志存储。
- `internal/oss/`
  - OSS 上传与签名能力。
- `public/`
  - 前端构建产物目录，由 Vite 输出并嵌入二进制。

### 前端 `frontend/`

- `src/main.js`
  - 前端入口。
- `src/App.vue`
  - 根组件。
- `src/pages/`
  - 主要页面：战斗、资料页、任务、消息、商店、后台、天赋等。
- `src/components/admin/`
  - 后台管理组件。
- `src/utils/`
  - 实时传输、状态合并、格式化等工具。
- `src/proto/`
  - 前端实时协议生成代码。

## 本地开发

### 环境要求

- Go `1.26.2`
- Bun `1.3.13` 或兼容版本
- 可访问的 Redis、MongoDB、Consul

运行时配置由 Consul 提供，常见本地环境变量：

```bash
export CONSUL_ADDR=http://127.0.0.1:8500
export CONSUL_CONFIG_KEY=vote-wall/dev
```

配置字段参考：

- `backend/config.example.yaml`
- `backend/cmd/printcfg/main.go`

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

补充命令：

```bash
make backend-backfill-boss-kills
make backend-check-boss-kills
go -C backend run ./cmd/printcfg
```

### 命令约束

- Go 命令必须在 `backend/` 目录执行，或使用 `go -C backend ...`。
- 前端命令默认在仓库根目录通过 `bun --cwd=frontend ...` 执行。
- `make dev` 会同时启动 Go 后端和 Vite 前端。
- `make build` 只构建前端产物到 `backend/public/`。
- `make test` 只运行后端测试。
- `make check` 是本地手动全量校验入口，包含：
  - 后端测试
  - `go vet`
  - 前端测试
  - 前端构建

## 提交前校验

仓库使用 `lefthook` 管理本地 `pre-commit`。

安装：

```bash
make hooks-install
```

当前 `pre-commit` 会执行：

- `bun --cwd=frontend install`
- `go -C backend mod tidy`
- `go -C backend fix ./...`
- `go -C backend test ./...`
- `go -C backend vet ./...`
- `bun --cwd=frontend run test`

## 构建与部署

### 前端产物

Vite 当前输出目录是 `frontend/vite.config.js` 中的 `../backend/public`，后端通过 `backend/embed_public.go` 将其嵌入编译产物，因此发布镜像只需要后端二进制即可承载静态资源。

### 后端发布二进制

当前发布命令：

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go -C backend build -trimpath -buildvcs=false -ldflags "-w -s -buildid=" -o long ./cmd/server
```

### GitHub Actions

当前 workflow 位于 `.github/workflows/build.yml`，在 `main` 分支 push 后执行：

1. 安装前端依赖
2. 构建前端产物
3. 构建 Linux `amd64` 后端二进制
4. 生成 `release.tar.gz`
5. 上传服务器并执行部署

当前 release 包包含：

- `Dockerfile`
- `.dockerignore`
- `docker-compose.yml`
- `backend/long`

### 服务器部署约束

- 基础镜像要求服务器本地已存在 `go-app-runtime:latest`。
- workflow 当前在服务器端执行：

```bash
docker-compose down
docker-compose up -d --build --force-recreate
docker image prune -f
```

## 测试特性

- 后端测试使用 `miniredis`，不依赖真实 Redis。
- 前端测试主要使用 `vitest` 做源码模式校验和工具逻辑校验，不依赖完整浏览器渲染环境。

## 文档入口

当前有效的一次性设计、方案与实施记录统一放在 `docs/`，总索引见 [docs/README.md](./docs/README.md)。

推荐入口：

- Mongo / 任务系统主线：
  - [docs/architecture/2026-05-01-任务系统与Mongo整合方案.md](./docs/architecture/2026-05-01-任务系统与Mongo整合方案.md)
  - [docs/architecture/2026-05-01-冷数据迁移与Mongo主存方案.md](./docs/architecture/2026-05-01-冷数据迁移与Mongo主存方案.md)
  - [docs/architecture/2026-05-01-日志与MongoDB演进方案.md](./docs/architecture/2026-05-01-日志与MongoDB演进方案.md)
- 实时链路与降载优化：
  - [docs/architecture/2026-05-04-WebSocket-Protobuf-点击链路优化方案.md](./docs/architecture/2026-05-04-WebSocket-Protobuf-点击链路优化方案.md)
  - [docs/architecture/2026-05-05-实时公共态广播节流与载荷收敛方案.md](./docs/architecture/2026-05-05-实时公共态广播节流与载荷收敛方案.md)
  - [docs/architecture/2026-05-06-Redis-降载一期用户态与装备链路优化方案.md](./docs/architecture/2026-05-06-Redis-降载一期用户态与装备链路优化方案.md)
- 房间分线与战斗表现：
  - [docs/architecture/2026-05-03-房间分线方案.md](./docs/architecture/2026-05-03-房间分线方案.md)
  - [docs/implementation/2026-05-05-Boss房间部位显示与掉血同步Bug排查记录.md](./docs/implementation/2026-05-05-Boss房间部位显示与掉血同步Bug排查记录.md)
  - [docs/developer-reference/2026-05-06-local-click-latency-脚本使用说明.md](./docs/developer-reference/2026-05-06-local-click-latency-脚本使用说明.md)
