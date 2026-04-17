# AGENTS

## 项目结构

- `frontend/` 是 Vue 前端，开发入口是 `frontend/src/App.vue`
- `backend/` 是 Go 后端，入口是 `backend/cmd/server/main.go`
- `backend/public/` 是前端构建产物，由后端静态托管

## 配置约定

- 运行时只读取 `backend/config.yaml`
- 测试只读取 `backend/config.test.yaml`
- 不要重新引入 `.env` 作为后端配置来源
- 如果要给用户示例配置，更新 `backend/config.example.yaml`

## 常用命令

- 开发：`npm run dev`
- 前端构建：`npm run build`
- 后端测试：`npm test`
- 后端单独运行：`go -C backend run ./cmd/server`

## 修改约定

- 前端改动尽量保持现有“投票墙”视觉风格，不要回退成后台面板感
- 后端逻辑优先放在 `backend/internal/` 下对应职责目录
- 如果改了接口、配置或部署方式，要同步更新 `README.md`
