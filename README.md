# Redis Vote Wall

一个 `frontend/` + `backend/` 分层的 `Vue + Go(Hertz) + Redis + SSE` 实时按钮计数墙项目。

## 功能

- 页面会展示所有符合规则的 Redis 按钮键。
- 访问者先输入昵称，之后点击都会记到这个昵称名下。
- 任意访问者点击按钮后，总数会按当前装备与暴击结算出的实际增量上涨。
- 按钮支持标签分组；前台可按标签和关键字搜索，后台可配置哪些按钮参与星光轮换。
- 玩家可以穿戴装备，装备会让平时点击直接获得额外次数增量。
- 玩家可以对同装备类型做 `3 合 1` 升星；升星不会生成新装备，而是强化该账号下这个装备类型的属性。
- 玩家可以通过 Boss 掉落招募小小英雄；每次只允许 1 位英雄出战，英雄属性和被动会同时影响平时点击与 Boss 战。
- 玩家可以把多余装备和重复英雄分解成 `原石`；穿戴中的最后 1 件装备、出战中的最后 1 位英雄不会被分解掉。
- 玩家可以消耗 `20` 原石对已有装备做强化，按 `52% / 22% / 12% / 13% / 1%` 提升点击、暴击、暴击率、随机两项或大奖全加；连续 `30` 次未出大奖后第 `31` 次保底。
- 玩家可以消耗 `25` 原石对已有英雄做觉醒，按 `36% / 22% / 15% / 14% / 12% / 1%` 提升点击、暴击、暴击率、被动、随机两项或大奖全加；同样采用 `30` 次未中后第 `31` 次保底。
- 前台新增外观商店，一期上架 `流星彩带 / 纸片庆典 / 印章敲击 / 流萤追光` 四套外观的轨迹与点击特效拆件，全部按单件 `30` 原石售卖。
- 外观只影响当前玩家自己的前台视觉表现，不改变按钮可点击性、Boss 伤害、掉率或任何数值结算。
- 所有在线用户都会通过 SSE 实时收到公共态更新；已登录用户还会收到仅属于自己的个人状态更新。
- 页面会实时展示个人累计点击和排行榜。
- 支持单个全服世界 Boss：Boss 活动期间，同一次点击会同时记票并造成 Boss 伤害。
- 星光卡片每 5 分钟轮换一次；每轮随机挑 6 个可参与按钮排到前面，点击按双倍增量结算。
- Boss 击杀后只有伤害达到该 Boss 最大生命值 `1%` 的玩家才有掉落资格；装备和英雄掉落各自独立抽取。
- 玩家 HUD 的“最近掉落”会展示同一次 Boss 结算拿到的全部奖励，不会再被后一条奖励覆盖。
- 前台会展示当前 Boss 的装备/英雄掉落池，以及后端计算好的实际掉落概率。
- 前台支持更新公告提醒、公告历史和全站公共留言墙。
- 提供 `/admin` 管理后台，可登录后配置 Boss 循环池、模板掉落池、装备和前台按钮；玩家、按钮、装备、历史 Boss 都改为分页拉取，避免后台首屏聚合全量列表。
- 后台支持为按钮图片申请阿里云 OSS 直传凭证，前端可直传 OSS 后回填公共图片 URL。
- 左侧玩家 HUD 支持手动开启挂机，开启后会跟随最近一次手动点击的按钮持续自动点击；关闭页面后自动停止。
- 你后面只要往 Redis 新增一个新键，前端就会自动展示新按钮。
- 前端静态页和后端 API/SSE 统一由一个 Go 服务承载，并可打成单一 Docker 镜像。
- 后端会在内存里做爆发点击限流，默认同时按客户端 IP 和昵称统计点击频率，超出人类能力的频率会被拉黑 10 分钟。
- 昵称会在后端统一做敏感词校验，当前接入整个 `konsheng/Sensitive-lexicon` 仓库里的文本词表。
- 后端运行和测试都会从 Consul 拉取 YAML 配置，本地不需要单独放配置文件。

## 目录结构

- `Makefile`: 项目顶层命令入口，统一开发、构建、测试流程
- `frontend/`: Vue 页面、样式和 Vite 配置
- `backend/`: Hertz Go 服务、Redis 读写、SSE、限流和测试

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
- `tags`: JSON 数组字符串，按钮标签
- `starlight_eligible`: `1` 表示可以参与星光轮换
- `image_path`: 可选，任意可访问图片地址，推荐填 OSS/CDN 公共 URL
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

后端会维护按钮显式索引；如果你直接往 Redis 手工新增 `vote:button:*`，服务会在低频兜底扫描后补进索引并出现在页面上。

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

世界 Boss、装备与掉落使用：

```text
vote:boss:current
vote:boss:pool:index
vote:boss:pool:<templateId>
vote:boss:pool:<templateId>:loot
vote:boss:pool:<templateId>:hero-loot
vote:boss:cycle
vote:boss:<bossId>:damage
vote:boss:<bossId>:loot
vote:boss:<bossId>:hero-loot
vote:buttons:index
vote:buttons:starlight
vote:equipment:index
vote:heroes:index
vote:players:index
vote:equip:def:<itemId>
vote:hero:def:<heroId>
vote:user-inventory:<nickname>
vote:user-hero-inventory:<nickname>
vote:user-loadout:<nickname>
vote:user-active-hero:<nickname>
vote:user-last-reward:<nickname>
vote:user-last-forge-result:<nickname>
vote:user-gems:<nickname>
vote:user-equip-upgrade:<nickname>:<itemId>
vote:user-hero-upgrade:<nickname>:<heroId>
vote:user-cosmetics:<nickname>
vote:user-cosmetic-loadout:<nickname>
vote:announcements
vote:announcement:<id>
vote:messages
vote:message:<id>
```

- `vote:boss:current` 是 `Hash`
  - `id`
  - `template_id`
  - `name`
  - `status`
  - `max_hp`
  - `current_hp`
  - `started_at`
  - `defeated_at`
- `vote:boss:pool:index` 是 `Set`
  - member = Boss 模板 `templateId`
- `vote:boss:pool:<templateId>` 是 `Hash`
  - `name`
  - `max_hp`
- `vote:boss:pool:<templateId>:loot` 是 `Sorted Set`
  - member = 装备 `itemId`
  - score = 掉落权重
- `vote:boss:pool:<templateId>:hero-loot` 是 `Sorted Set`
  - member = 英雄 `heroId`
  - score = 掉落权重
- `vote:boss:cycle` 是 `Hash`
  - `enabled`
- `vote:boss:<bossId>:damage` 是 `Sorted Set`
  - member = 昵称
  - score = 对该 Boss 的累计伤害
- `vote:boss:<bossId>:loot` 是 `Sorted Set`
  - member = 装备 `itemId`
  - score = 掉落权重
- `vote:boss:<bossId>:hero-loot` 是 `Sorted Set`
  - member = 英雄 `heroId`
  - score = 掉落权重
- `vote:buttons:index` 是 `Sorted Set`
  - member = 按钮 `slug`
  - score = 按钮排序值 `sort`
- `vote:buttons:starlight` 是 `Set`
  - member = 可参与星光轮换的按钮 `slug`
- `vote:equipment:index` 是 `Set`
  - member = 装备 `itemId`
- `vote:heroes:index` 是 `Set`
  - member = 英雄 `heroId`
- `vote:players:index` 是 `Sorted Set`
  - member = 昵称
  - score = 最近活跃时间戳
- `vote:equip:def:<itemId>` 是 `Hash`
  - `name`
  - `slot` (`weapon` / `armor` / `accessory`)
  - `bonus_clicks`
  - `bonus_critical_chance_percent`
  - `bonus_critical_count`
- `vote:hero:def:<heroId>` 是 `Hash`
  - `name`
  - `image_path`
  - `image_alt`
  - `bonus_clicks`
  - `bonus_critical_chance_percent`
  - `bonus_critical_count`
  - `trait_type`
  - `trait_value`
- `vote:user-inventory:<nickname>` 是 `Hash`
  - field = `itemId`
  - value = 库存数量
- `vote:user-hero-inventory:<nickname>` 是 `Hash`
  - field = `heroId`
  - value = 库存数量
- `vote:user-loadout:<nickname>` 是 `Hash`
  - `weapon`
  - `armor`
  - `accessory`
- `vote:user-active-hero:<nickname>` 是 `String`
  - value = 当前出战英雄 `heroId`
- `vote:user-last-reward:<nickname>` 是 `Hash`
  - `boss_id`
  - `boss_name`
  - `item_id`
  - `item_name`
  - `granted_at`
  - `recent_rewards`：最近一次 Boss 结算奖励数组的 JSON 字符串，兼容同时掉装备和英雄
- `vote:user-last-forge-result:<nickname>` 是 `String`
  - value = 最近一次分解/强化/觉醒结果的 JSON 字符串
- `vote:user-gems:<nickname>` 是 `String`
  - value = 当前原石数量
- `vote:user-equip-upgrade:<nickname>:<itemId>` 是 `Hash`
  - `star_level`
  - `bonus_clicks`
  - `bonus_critical_chance_percent`
  - `bonus_critical_count`
  - `reforge_pity_counter`
- `vote:user-hero-upgrade:<nickname>:<heroId>` 是 `Hash`
  - `awaken_level`
  - `bonus_clicks`
  - `bonus_critical_chance_percent`
  - `bonus_critical_count`
  - `trait_value`
  - `pity_counter`
- `vote:user-cosmetics:<nickname>` 是 `Set`
  - member = 已拥有外观 `cosmeticId`
- `vote:user-cosmetic-loadout:<nickname>` 是 `Hash`
  - `trail`
  - `impact`
- `vote:announcements` 是 `Sorted Set`
  - member = 公告 `id`
  - score = 公告顺序 ID（越大越新）
- `vote:announcement:<id>` 是 `Hash`
  - `title`
  - `content`
  - `published_at`
  - `active`
- `vote:messages` 是 `Sorted Set`
  - member = 留言 `id`
  - score = 留言顺序 ID（越大越新）
- `vote:message:<id>` 是 `Hash`
  - `nickname`
  - `content`
  - `created_at`

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
  limit: 42
  window_ms: 2000
  blacklist_ms: 600000
critical_hit:
  chance_percent: 5
  count: 5
admin:
  username: "admin"
  password: "change-me"
  session_secret: "change-this-too"
player_auth:
  jwt_secret: "change-player-jwt-secret"
  jwt_ttl_seconds: 604800
oss:
  access_key_id: "your-ak"
  access_key_secret: "your-secret"
  bucket: "your-bucket"
  region: "cn-beijing"
  public_base_url: "https://cdn.example.com"
  upload_dir_prefix: "buttons"
  expire_seconds: 300
```

`button_poll_interval_ms` 现在表示“按钮索引低频兜底同步间隔”，不是公共状态轮询广播间隔。推荐生产环境调大到 `60000` 左右；只有你直接手工写 Redis 新按钮、没有走后台保存接口时，它才会影响这个按钮多久能被前台看到。

## 前台接口补充

- `POST /api/player/auth/login`
  - 请求体：`{ "nickname": "阿明", "password": "secret" }`
  - 首次使用该昵称时会直接为它设置密码；之后必须用同一昵称和密码登录
- `POST /api/player/auth/logout`
- `GET /api/player/auth/session`
  - 返回当前玩家登录态和昵称

- `POST /api/equipment/{itemId}/salvage`
  - 请求体：`{ "nickname": "阿明", "quantity": 2 }`
- `POST /api/equipment/{itemId}/reforge`
  - 请求体：`{ "nickname": "阿明" }`
- `POST /api/heroes/{heroId}/salvage`
  - 请求体：`{ "nickname": "阿明", "quantity": 1 }`
- `POST /api/heroes/{heroId}/awaken`
  - 请求体：`{ "nickname": "阿明" }`
- `GET /api/shop`
  - 支持可选查询参数 `nickname`
- `POST /api/shop/cosmetics/{cosmeticId}/purchase`
  - 请求体：`{ "nickname": "阿明" }`
- `POST /api/shop/cosmetics/equip`
  - 请求体：`{ "nickname": "阿明", "trailId": "trail-ribbon", "impactId": "impact-firefly" }`

说明：

- 前台写接口现在要求玩家先登录，后端优先使用 JWT 会话中的昵称，不再信任客户端随便传来的 `nickname`
- 现有请求体里的 `nickname` 字段保留是为了兼容前端结构，真实身份以后端登录态为准

## 后台接口补充

- `GET /api/admin/state`
  - 现在返回后台首屏摘要，不再携带按钮和装备全量列表
- `GET /api/admin/buttons?page=1&pageSize=20`
  - 返回按钮分页：`items / page / pageSize / total / totalPages`
- `GET /api/admin/equipment?page=1&pageSize=20`
  - 返回装备分页：`items / page / pageSize / total / totalPages`
- `GET /api/admin/boss/history?page=1&pageSize=20`
  - 返回历史 Boss 分页：`items / page / pageSize / total / totalPages`
- `POST /api/admin/players/{nickname}/password/reset`
  - 请求体：`{ "password": "new-secret" }`
  - 用于后台手动重置某个玩家昵称的密码，并让旧玩家会话失效

## 限流规则

- 默认同时按客户端 IP 和昵称统计点击频率
- `2` 秒内超过 `42` 次点击会被判定为异常爆发
- 命中后会进入 `10` 分钟黑名单
- 这些值都可以通过 Consul 里的 YAML 配置调整

可调整配置：

- `rate_limit.limit`
- `rate_limit.window_ms`
- `rate_limit.blacklist_ms`
- `critical_hit.chance_percent`
- `critical_hit.count`
- `admin.username`
- `admin.password`
- `admin.session_secret`
- `player_auth.jwt_secret`
- `player_auth.jwt_ttl_seconds`
- `oss.access_key_id`
- `oss.access_key_secret`
- `oss.bucket`
- `oss.region`
- `oss.public_base_url`
- `oss.upload_dir_prefix`
- `oss.expire_seconds`

## 昵称敏感词校验

- 昵称校验由后端统一执行，前端只展示后端返回的校验结果。
- 当前词表来自 `konsheng/Sensitive-lexicon`，仓库内置在 `backend/internal/nickname/lexicon/upstream/`。
- 当前实现会加载该目录下所有 vendored `.txt` 词表并自动去重，不再只限于政治类词表。
- 上游许可证文本随仓库一并保存在 `backend/internal/nickname/lexicon/LICENSE.konsheng.txt`。
- 公共留言内容复用同一套词表与匹配逻辑，命中敏感词会被后端拒绝。

## 本地启动

先安装前端依赖：

```bash
make deps
```

然后准备 Consul 环境变量：

```bash
export CONSUL_ADDR=http://127.0.0.1:8500
export CONSUL_CONFIG_KEY=vote-wall/dev
```

启动开发环境：

```bash
make dev
```

默认地址：

- 前端开发页：`http://localhost:5173`
- Go 后端接口：`http://localhost:2333`
- Go `pprof` 调试入口：`http://localhost:2333/debug/pprof/`
- 管理后台：`http://localhost:5173/admin`（构建后由 Go 服务统一托管）

## 构建

运行：

```bash
make build
```

前端产物会输出到 `backend/public/`，由后端统一静态托管。
这个目录是构建输出，不提交到仓库。

## 测试

运行 Go 后端测试：

```bash
make test
```

运行前端 Vitest：

```bash
npm --prefix frontend run test
```

如果你只想单独启动后端，也可以直接跑：

```bash
make backend-run
```

其他常用目标：

```bash
make frontend-dev
make frontend-build
make backend-test
make backend-vet
make check
```

## Docker 构建与运行

构建镜像：

```bash
docker buildx build --platform linux/amd64 -t long . --load
```

镜像会在构建阶段自动编译前端静态资源和 Go 服务，不依赖宿主机预先生成 `long` 二进制或 `backend/public` 目录。

运行容器：

```bash
docker run -d \
  --name long \
  --network docker-compose_app-net \
  -p 2333:2333 \
  -e CONSUL_ADDR=http://your-consul:8500 \
  -e CONSUL_CONFIG_KEY=vote-wall/prod \
  long
```

这个单镜像现在内置了 `nginx + go` 双进程：

- 宿主机访问入口：`http://localhost:2333`
- 容器内 `nginx`：监听 `2333`
- 容器内 Go：仅监听 `127.0.0.1:18080`

外部不需要再额外部署一个 Nginx 给它做反代；镜像内置配置位于 `deploy/nginx.container.conf`，参考版见 `deploy/nginx.vote-wall.conf.example`。
