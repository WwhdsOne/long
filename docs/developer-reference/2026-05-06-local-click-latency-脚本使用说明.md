# `local_click_latency` 脚本使用说明

## 作用

`backend/cmd/local_click_latency` 是一个本地压测小工具，只做一件事：

- 登录指定玩家账号
- 建立 1 条 `WebSocket /api/ws`
- 顺序发送二进制 `click`
- 逐个等待 `click_ack`
- 输出这批点击的确认延迟统计

它适合回答两个问题：

- 单连接顺序点击时，每次点击从发出到收到 `click_ack` 要多久
- 单连接顺序点击时，平均能打出多少次确认点击/秒

它**不是**并发压测工具，也不会模拟多个连接。

## 点击哪里

这个脚本不会自己选房间，也不会自己找 Boss。

当前后端点击链路是：

- `slug` 只表示“点 Boss 的哪个部位”
- 房间来自“这个玩家当前所在房间”

代码口径：

- [ClickButton](/Users/Learning/web/long/backend/internal/core/store.go:1044) 现在只接受 `boss-part:` 前缀
- [clickBossPart](/Users/Learning/web/long/backend/internal/core/store.go:1774) 会先调用 `ResolvePlayerRoom`
- [ResolvePlayerRoom](/Users/Learning/web/long/backend/internal/core/room.go:183) 读取的是该玩家当前房间

所以结论很直接：

- 传 `-slug boss-part:1-0` 的含义是“点击当前房间 Boss 的 `(1,0)` 部位”
- 不是“点击 1 号房间”
- 不是“自动寻找某个 Boss”

你刚才那条命令里的 `-slug feel` 在当前代码里是无效的，脚本现在会直接本地报错，不再把错误请求发到服务器。

## 房间怎么定

房间是按玩家昵称持久化的，不绑浏览器页面，也不绑这条 WebSocket 连接。

也就是说：

- 先用同一个账号在网页里切到目标房间，再跑脚本，可以
- 先通过接口把同一个账号切到目标房间，再跑脚本，也可以

如果不切，脚本就会打这个账号当前保存的房间。未设置时通常会落到 `hall`，而 `hall` 没有活跃 Boss 时会直接失败。

切房间接口是 [room_routes.go](/Users/Learning/web/long/backend/internal/httpapi/room_routes.go:34) 里的：

- `POST /api/rooms/join`

## `slug` 怎么写

格式必须是：

```text
boss-part:x-y
```

例如：

- `boss-part:0-0`
- `boss-part:1-0`
- `boss-part:2-3`

这里的 `x` / `y` 是当前 Boss 部位布局坐标，必须真正在当前房间 Boss 的 `parts` 里存在，否则服务端会返回“部位不存在”。

最稳的做法是：

- 先打开对应房间战斗页
- 确认要打的部位坐标
- 再把这个坐标写进 `-slug`

## 基本命令

```bash
go -C backend run ./cmd/local_click_latency \
  -base https://www.wclick.top \
  -nickname Wwhds \
  -password '123456' \
  -slug boss-part:0-0 \
  -count 500
```

常用参数：

- `-base`
  - 站点地址
- `-nickname`
  - 压测账号昵称
- `-password`
  - 压测账号密码
- `-slug`
  - 目标 Boss 部位，必须是 `boss-part:x-y`
- `-count`
  - 点击次数
- `-pause`
  - 每次点击之间的停顿，例如 `5ms`
- `-timeout`
  - 单次 HTTP/读写超时，默认 `10s`
- `-handshake-wait`
  - 建连后额外等待一小段时间再开始点
- `-insecure-origin`
  - 额外附带 `Origin: <base>`，用于某些网关校验

查看帮助：

```bash
go -C backend run ./cmd/local_click_latency -h
```

## 建议测试步骤

### 1. 先把测试账号切到目标房间

推荐直接在前台页面用这个账号切房间。

如果你要用接口，可以自己先登录后调用：

```bash
curl -c /tmp/long.cookies -b /tmp/long.cookies \
  -H 'Content-Type: application/json' \
  -d '{"nickname":"Wwhds","password":"123456"}' \
  https://www.wclick.top/api/player/auth/login

curl -c /tmp/long.cookies -b /tmp/long.cookies \
  -H 'Content-Type: application/json' \
  -d '{"roomId":"2"}' \
  https://www.wclick.top/api/rooms/join
```

脚本自己也会再登录一次，但房间是按昵称保存的，所以不会影响这个结论。

### 2. 确定目标部位坐标

例如你要测当前 Boss 的 `(0,0)` 部位，就用：

```bash
-slug boss-part:0-0
```

### 3. 跑脚本

```bash
go -C backend run ./cmd/local_click_latency \
  -base https://www.wclick.top \
  -nickname Wwhds \
  -password '123456' \
  -slug boss-part:0-0 \
  -count 500
```

## 输出含义

输出示例大概是：

```text
账号: Wwhds
按钮: boss-part:0-0
连接: 单个 WebSocket
样本数: 500
总耗时: 4.2s
平均吞吐: 118.73 次/秒
最小延迟: 4.1ms
平均延迟: 7.8ms
p50 延迟: 7.2ms
p95 延迟: 11.4ms
p99 延迟: 16.9ms
最大延迟: 24.0ms
```

这里最有用的是：

- `平均吞吐`
  - 单连接顺序点击时，平均每秒确认多少次点击
- `p50 / p95 / p99`
  - 点击确认延迟分位

## 常见报错

- `-slug 必须是 boss-part:x-y`
  - 说明还在用旧按钮 key，例如 `feel`
- `第 N 次点击返回文本错误`
  - 服务端返回了业务错误 JSON，常见原因是：
  - 当前房间没有活跃 Boss
  - 目标部位坐标不存在
  - 目标部位已经被打死
  - 压测账号没进对房间
- `建立 WebSocket 失败`
  - 常见是网关、证书、来源校验或站点不可达

## 边界

- 这是“单连接、顺序点击、逐个等确认”的延迟脚本
- 它测的是 `click -> click_ack` 这一段用户体感最接近的链路
- 它不代表多连接并发上限
- 它不会自动切房间，也不会自动发现部位坐标
