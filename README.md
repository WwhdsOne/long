# Redis Vote Wall

一个 `frontend/` + `backend/` 分层的 `Vue + Go + Redis + SSE` 实时按钮计数墙项目。

## 功能

- 页面会展示所有符合规则的 Redis 按钮键。
- 访问者先输入昵称，之后点击都会记到这个昵称名下。
- 任意访问者点击按钮后，总数立即 `+1`。
- 所有在线用户都会通过 SSE 实时看到最新计数。
- 页面会实时展示个人累计点击和排行榜。
- 你后面只要往 Redis 新增一个新键，前端就会自动展示新按钮。
- 前端静态页和后端 API/SSE 统一由一个 Go 服务承载，并可打成单一 Docker 镜像。
- 后端会在内存里做爆发点击限流，超出人类能力的频率会被拉黑 10 分钟。
- 昵称会在后端统一做敏感词校验，当前接入整个 `konsheng/Sensitive-lexicon` 仓库里的文本词表。
- 后端运行和测试都会从 Consul 拉取 YAML 配置，本地不需要单独放配置文件。

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

个人统计和排行榜使用：

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

## Consul 配置

后端启动时只认这两个环境变量：

- `CONSUL_ADDR`
- `CONSUL_CONFIG_KEY`

它会从 Consul KV 里拉取一份 YAML 配置，并在配置变化后主动退出，让外部拉起新进程。

存进 Consul 的 YAML 内容可以长这样：

```yaml
port: 2333
redis:
  host: 127.0.0.1
  port: 6379
  username: ""
  password: ""
  db: 0
  tls_enabled: false
redis_prefix: "vote:button:"
button_poll_interval_ms: 3000
rate_limit:
  limit: 12
  window_ms: 2000
  blacklist_ms: 600000
critical_hit:
  chance_percent: 5
  count: 5
```

## 限流规则

- 默认按客户端 IP 统计点击频率
- `2` 秒内超过 `12` 次点击会被判定为异常爆发
- 命中后会进入 `10` 分钟黑名单
- 这些值都可以通过 Consul 里的 YAML 配置调整

可调整配置：

- `rate_limit.limit`
- `rate_limit.window_ms`
- `rate_limit.blacklist_ms`
- `critical_hit.chance_percent`
- `critical_hit.count`

## 昵称敏感词校验

- 昵称校验由后端统一执行，前端只展示后端返回的校验结果。
- 当前词表来自 `konsheng/Sensitive-lexicon`，仓库内置在 `backend/internal/nickname/lexicon/upstream/`。
- 当前实现会加载该目录下所有 vendored `.txt` 词表并自动去重，不再只限于政治类词表。
- 上游许可证文本随仓库一并保存在 `backend/internal/nickname/lexicon/LICENSE.konsheng.txt`。

## 本地启动

先安装项目和前端依赖：

```bash
npm install
npm --prefix frontend install
```

然后准备 Consul 环境变量：

```bash
export CONSUL_ADDR=http://127.0.0.1:8500
export CONSUL_CONFIG_KEY=vote-wall/dev
```

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
这个目录是构建输出，不提交到仓库。

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
  -e CONSUL_ADDR=http://your-consul:8500 \
  -e CONSUL_CONFIG_KEY=vote-wall/prod \
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
