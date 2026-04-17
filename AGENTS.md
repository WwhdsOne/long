# AGENTS

## 项目结构

- `frontend/` 是 Vue 前端，开发入口是 `frontend/src/App.vue`
- `backend/` 是 Go 后端，入口是 `backend/cmd/server/main.go`
- `backend/public/` 是前端构建产物目录，由后端静态托管，但不提交到仓库

## 配置约定

- 运行时和测试都通过 `CONSUL_ADDR` 与 `CONSUL_CONFIG_KEY` 从 Consul 读取配置
- 不要重新引入本地 YAML 或 `.env` 作为后端配置来源
- 如果改了配置结构，要同步更新 `backend/internal/config/config.go` 里的 YAML 解析结构和 `README.md` 里的 Consul 配置示例

## 常用命令

- 开发：`npm run dev`
- 前端构建：`npm run build`
- 后端测试：`npm test`
- 后端单独运行：`go -C backend run ./cmd/server`

## 修改约定

- 前端改动尽量保持现有“投票墙”视觉风格，不要回退成后台面板感
- 后端逻辑优先放在 `backend/internal/` 下对应职责目录
- 如果改了接口、配置或部署方式，要同步更新 `README.md`
