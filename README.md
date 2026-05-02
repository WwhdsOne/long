# Redis Vote Wall

一个基于 `Vue 3 + Vite + Go(Hertz) + Redis + SSE` 的实时按钮计数墙项目。当前版本已经扩展出装备、英雄、Boss、天赋、任务、留言墙、后台管理和 Mongo 冷数据链路。

## 当前架构

- 前端位于 `frontend/`，不引入路由库，通过 `window.location.pathname` 切分页面。
- 后端位于 `backend/`，模块路径为 `long`。
- 公共态通过 `SSE /api/events` 推送，个人态通过 `SSE /api/events/me` 推送，WebSocket 仅作备选。
- Redis 负责热数据；Mongo 已启用，用于冷数据、日志、任务定义与归档等持久化内容。
- 前端构建产物输出到 `backend/public/`，由容器内的 nginx + Go 服务统一承载。

## 目录入口

- `AGENTS.md` / `CLAUDE.md`
  - 长期协作规则。两份文件内容保持完全一致。
- `Makefile`
  - 本地开发、构建、测试、hook 安装入口。
- `frontend/`
  - Vue 页面、Vite 配置、前端测试。
- `backend/`
  - Hertz 服务、业务逻辑、Redis/Mongo 访问、后端测试。
- `deploy/`
  - 容器入口脚本和 nginx 配置。
- `docs/`
  - 一次性设计、计划、总结、开发参考和历史归档。总索引见 [docs/README.md](./docs/README.md)。

## 本地开发

先安装前端依赖：

```bash
make deps
```

后端运行与测试依赖 Consul 配置，需要设置：

```bash
export CONSUL_ADDR=http://127.0.0.1:8500
export CONSUL_CONFIG_KEY=vote-wall/dev
```

常用命令：

```bash
make dev
make backend-run
make frontend-dev
make build
make test
npm --prefix frontend run test
make check
```

说明：

- `make dev`：同时启动 Go 后端和 Vite 前端。
- `make build`：只构建前端产物到 `backend/public/`。
- `make test`：运行后端测试。
- `make check`：本地手动全量校验，包含后端测试、`go vet`、前端测试和前端构建。
- Go 命令必须在 `backend/` 下执行，或使用 `go -C backend ...`。

## 提交前校验

仓库使用 `lefthook` 管理本地 `pre-commit`。

安装 hook：

```bash
make hooks-install
```

当前 `pre-commit` 会执行：

- `go -C backend fix ./...`
- `go -C backend test ./...`
- `go -C backend vet ./...`
- `npm --prefix frontend run test`

`make check` 仍保留为手动全量校验入口，但不再作为 GitHub Actions 部署步骤的一部分。

## 部署概览

当前 GitHub Actions 部署链路是：

1. 安装前端依赖。
2. 构建前端产物到 `backend/public/`。
3. 交叉编译后端 Linux `amd64` 发布二进制：

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go -C backend build -trimpath -buildvcs=false -ldflags "-w -s -buildid=" -o long ./cmd/server
```

4. 生成 `release.tar.gz`，其中包含：
   - `Dockerfile`
   - `.dockerignore`
   - `docker-compose.yml`
   - `backend/long`
   - `backend/public/`
   - `deploy/entrypoint.sh`
   - `deploy/nginx.container.conf`
5. 上传到服务器后，在服务器端执行 `docker compose build` 和 `docker compose up -d`。

运行镜像以服务器本地已有的 `long-basic:latest` 为基础镜像；部署前 workflow 会先检查该镜像是否存在。

## Redis 与配置

- 热数据仍在 Redis。
- 运行时配置通过 Consul KV 拉取 YAML。
- `backend/config.example.yaml` 和 `backend/config.yaml` 仅作参考，不直接参与线上运行时加载。
- 详细数据结构和配置字段，见本文档后续章节与 [docs/README.md](./docs/README.md) 的相关专题入口。

## 文档入口

如果你要看当前有效文档，优先从 [docs/README.md](./docs/README.md) 进入。

推荐入口：

- Mongo / 任务系统：
  - [docs/architecture/2026-05-01-任务系统与Mongo整合方案.md](./docs/architecture/2026-05-01-任务系统与Mongo整合方案.md)
  - [docs/architecture/2026-05-01-冷数据迁移与Mongo主存方案.md](./docs/architecture/2026-05-01-冷数据迁移与Mongo主存方案.md)
- 天赋与伤害链路：
  - [docs/reports/2026-04-28-天赋成本调整总结.md](./docs/reports/2026-04-28-天赋成本调整总结.md)
  - [docs/developer-reference/2026-04-26-天赋系统开发参考.md](./docs/developer-reference/2026-04-26-天赋系统开发参考.md)
  - [docs/developer-reference/2026-04-30-V2伤害计算链路总览.md](./docs/developer-reference/2026-04-30-V2伤害计算链路总览.md)

## Redis 数据结构

### 按钮

每个按钮使用一个 Redis `Hash`，键名格式：

```text
vote:button:<slug>
```

字段约定：

- `label`: 按钮显示文本
- `count`: 当前总数
- `sort`: 排序值，越小越靠前
- `enabled`: `1` 为展示，`0` 为隐藏
- `tags`: JSON 数组字符串，按钮标签
- `image_path`: 可选，任意可访问图片地址，推荐填 OSS/CDN 公共 URL
- `image_alt`: 可选，图片说明文本

按钮显式索引：

```text
vote:buttons:index
```

后端后台保存接口会维护该索引；如果你直接手工写 `vote:button:*`，服务会通过低频兜底同步把它补进索引。

### 用户与排行榜

```text
vote:user:<nickname>
vote:leaderboard
```

- `vote:user:<nickname>` 是 `Hash`
  - `nickname`
  - `click_count`
  - `updated_at`
- `vote:leaderboard` 是 `Sorted Set`
  - member = 昵称
  - score = 个人累计点击数

### 其他热数据

项目还会使用以下键族：

```text
vote:boss:*
vote:equipment:*
vote:heroes:*
vote:players:index
vote:user-inventory:*
vote:user-loadout:*
vote:resource:*
vote:announcements*
vote:messages*
```

具体字段说明与演进背景，优先查：

- [docs/architecture/2026-05-01-任务系统与Mongo整合方案.md](./docs/architecture/2026-05-01-任务系统与Mongo整合方案.md)
- [docs/architecture/2026-05-01-冷数据迁移与Mongo主存方案.md](./docs/architecture/2026-05-01-冷数据迁移与Mongo主存方案.md)

## Consul 配置

后端启动时识别两个核心环境变量：

- `CONSUL_ADDR`
- `CONSUL_CONFIG_KEY`

配置从 Consul KV 拉取 YAML；配置变更后，服务会主动退出，由外部进程管理器拉起新进程。
