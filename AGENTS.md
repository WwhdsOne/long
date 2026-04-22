# Purpose

`AGENTS.md` 只放长期协作规则，不放一次性开发需求、临时任务单或实现方案。

一次性任务说明写进 `docs/`，不要继续堆进这里。

# 语言：中文

- 所有回复、解释、代码注释、commit message 均使用中文。
- 变量名、函数名等遵循项目约定，可保留英文。
- 代码注释使用中文，除非项目已有英文注释惯例。
- 错误信息和日志的解读用中文说明。

# Repository Index

这里只维护长期稳定的“目录职责 + 少量关键入口”，作为仓库导航索引。

不维护全量文件枚举，不把它写成一次性任务清单。

发现新模块时，只有当它已经成为长期稳定入口，再补进这里。

- `backend/`：Go + Hertz 服务、API、SSE、业务逻辑
- `frontend/`：Vue 3 + Vite 页面与前端测试
- `deploy/`：容器与 Nginx 部署配置
- `scripts/`：本地脚本
- `docs/`：一次性任务、设计、计划、记录
- `Makefile`：顶层开发、构建、校验入口
- `backend/cmd/server/main.go`：后端启动入口
- `frontend/src/main.js`：前端启动入口

# Read Policy

先搜索，后读取。

先用 `rg --files` 或 `rg -n` 缩小范围，再按需用 `sed -n`、`rg -n`、定点打开文件。

一次只读当前决策需要的少量文件；读到足够做决定就停止。

禁止因为“不确定在哪”而连续展开多个目录、批量通读大量文件或做无目标扫仓库。

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

不要在这里扩展与当前仓库无关的 RTK 命令族。

不要写 `rtk gh`、`rtk curl`，也不要补入 Python、Rust、.NET、Ruby、AWS、`kubectl`、`psql` 等与当前仓库无关的能力表。

# Git Workflow

遇到功能改动、接口改动、权限改动、数据模型改动时，必须先从 `main` 拉出新分支再开始开发。

禁止直接在 `main` 上做这类开发改动。

完成后先在功能分支提交，再合并回 `main`；需要推送时，同时推送 `main` 和对应功能分支。

最终说明中明确给出本次功能分支名、功能分支提交和合并后的 `main` 提交。

# Task Docs

一次性任务、设计说明、执行计划、阶段记录统一写入 `docs/`。

新任务文档应使用清晰文件名，优先采用 `YYYY-MM-DD-主题.md` 形式，便于检索和归档。

不要把临时背景、单次约束或阶段性 TODO 回写到 `AGENTS.md`。
