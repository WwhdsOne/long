# Redis Vote Wall

一个 `frontend/` + `backend/` 分层的 `Vue + Go + Redis + SSE` 实时按钮计数墙项目。

## 功能

- 页面会展示所有符合规则的 Redis 按钮键。
- 任意访问者点击按钮后，总数立即 `+1`。
- 所有在线用户都会通过 SSE 实时看到最新计数。
- 你后面只要往 Redis 新增一个新键，前端就会自动展示新按钮。
- 前端静态页和后端 API/SSE 统一由一个 Go 服务承载，并可打成单一 Docker 镜像。
- 后端会在内存里做爆发点击限流，超出人类能力的频率会被拉黑 10 分钟。

## 目录结构

- `frontend/`: Vue 页面、样式和 Vite 配置
- `backend/`: Go 服务、Redis 读写、SSE、限流和测试
- `scripts/`: Docker 重建和镜像导出脚本

## Redis 数据结构

每个按钮使用一个 Redis `Hash`，键名格式：

```text
vote:button:<slug>
```

字段约定：

- `label`: 按钮显示文本
- `count`: 当前总数
- `sort`: 排序值，越小越靠前
- `enabled`: `1` 为展示，`0` 为隐藏
- `image_path`: 可选，本地静态图或可访问图片地址
- `image_alt`: 可选，图片说明文本

示例：

```bash
redis-cli HSET vote:button:feel label "有感觉吗" count 0 sort 10 enabled 1
redis-cli HSET vote:button:understand label "有没有懂的" count 0 sort 20 enabled 1
redis-cli HSET vote:button:wechat-pity label "微信[可怜]表情" count 0 sort 30 enabled 1 image_path "/images/emojipedia-wechat-whimper.png" image_alt "微信可怜表情"
```

新增按钮示例：

```bash
redis-cli HSET vote:button:new-one label "新按钮" count 0 sort 40 enabled 1
```

后端会周期性扫描 `vote:button:*`，所以新增后几秒内就会自动出现在页面上。

## 限流规则

- 默认按客户端 IP 统计点击频率
- `2` 秒内超过 `12` 次点击会被判定为异常爆发
- 命中后会进入 `10` 分钟黑名单
- 这些值都可以通过 `backend/config.yaml` 调整

可调整配置：

- `rate_limit.limit`
- `rate_limit.window_ms`
- `rate_limit.blacklist_ms`

## 本地启动

先安装项目和前端依赖：

```bash
npm install
npm --prefix frontend install
```

先准备后端配置文件：

```bash
cp backend/config.example.yaml backend/config.yaml
```

`backend/config.yaml` 会被本地运行直接读取，Docker 构建时也会一起打进镜像。

启动开发环境：

```bash
npm run dev
```

默认地址：

- 前端开发页：`http://localhost:5173`
- Go 后端接口：`http://localhost:2333`

## 构建

运行：

```bash
npm run build
```

前端产物会输出到 `backend/public/`，由后端统一静态托管。

## 测试

运行 Go 后端测试：

```bash
npm test
```

如果你只想单独启动后端，也可以直接跑：

```bash
go -C backend run ./cmd/server
```

## Docker 构建与运行

构建镜像：

```bash
docker buildx build --platform linux/amd64 -t long . --load
```

运行容器：

```bash
docker run -d \
  --name long \
  --network host \
  long
```

一键删除旧容器和旧镜像并重新启动名为 `long` 的镜像：

```bash
bash ./scripts/rebuild-run.sh
```

这个脚本会：

- 删除旧的 `long` 容器
- 删除旧的 `long` 镜像
- 重新构建 `long`
- 使用 `--network host` 启动

## 导出并上传镜像

如果你想在本地构建好，再把镜像包传到服务器：

```bash
bash ./scripts/build-save-upload.sh user@your-server:/path/
```

这个脚本会：

- 删除本地旧的 `long` 镜像
- 删除旧的 `long.tar.gz`
- 使用 `buildx` 按 `linux/amd64` 重新构建 `long`
- 导出为 `long.tar.gz`
- 自动 `scp` 上传到你指定的服务器路径

也支持自定义镜像名和压缩包名：

```bash
bash ./scripts/build-save-upload.sh user@your-server:/path/ my-image my-image.tar.gz
```

## 可选 Nginx 反代

如果你有现成的 Nginx，可以把流量反代到这个 Go 容器。示例见 `deploy/nginx.vote-wall.conf.example`。
