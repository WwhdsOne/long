# Boss 实时状态 Delta 化与首包重同步方案

> 状态：待评审
> 日期：2026-05-09
> 结论：本次只改 `Boss` 高频实时下行，采用“首包全量 + 高频 delta + 真异常才重同步”的单协议切换方案；`room_state` 保持现状，不纳入本轮重构。

## 一、目标与边界

本次方案只处理战斗主链路里最重、最频繁的 `Boss` 实时状态广播。

目标：

- 保持用户无感，不改变点击手感和战斗表现
- 显著缩小高频下行包体
- 避免在高频包里重复发送 `Boss` 名、部位名、房间展示名、emoji 等长文本
- 明确首包、增量包、异常重同步三类协议职责

明确不纳入本次范围：

- `room_state` 进一步拆分或 delta 化
- 登录、商店、任务、背包等低频接口协议
- Redis 存储格式调整
- 前后端双协议兼容期

发布前提：

- 前后端同批发布
- 不要求旧前端与新后端混跑
- 不要求新前端兼容旧后端

## 二、当前问题

当前高频下行的主要问题不是连接模型，而是 `Boss` 公共态仍然过于“全量”。

### 2.1 当前包体里存在大量重复文本

当前 `Boss` 对象会重复携带：

- `boss.id`
- `boss.name`
- `boss.status`
- `boss.parts[*].x / y / type`
- `boss.parts[*].label / 展示名`
- 其他静态结构字段

其中：

- 英文字符串通常 1 字节/字符
- 汉字通常约 3 字节/字符
- emoji 通常约 4 字节/字符

`protobuf` 能减少字段编码开销，但不会把长文本“自动压到极小”。如果高频包继续反复带整份 `Boss` 结构，那么只是把“大 JSON”换成了“大 protobuf”。

### 2.2 高频事件实际只改了少量字段

一次普通点击、挂机攻击、流血 tick、魔法结算，真正变化的通常只有：

- `boss.currentHp`
- 少数几个部位的 `currentHp / alive`
- 当前玩家相关计数
- 少量战斗特效事件

未变化的字段占了绝大多数：

- `Boss` 名称
- 部位名称
- 坐标
- 部位类型
- 静态布局顺序

### 2.3 只靠“广播频率收敛”还不够

前面已经收窄了 `room_state` 的触发范围，但 `Boss` 高频包本身仍偏重。

因此这次优化的重点不再是：

- 再次延长节流窗口
- 再次减少广播次数

而是：

- 保留当前广播节奏
- 只发送真正变化的 `Boss` 动态字段

## 三、方案结论

本次正式采用：

- 首包全量 `snapshot`
- 高频 `boss_delta`
- 版本号校验
- 真异常才 `sync_request`

不采用：

- 高频继续发送完整 `Boss`
- 高频包携带部位名 / 房间名 / emoji
- 增量丢包后靠“静默容错”
- 新旧协议并行兼容

## 四、状态拆分

### 4.1 低频静态状态

这部分只在首包、切房、切 Boss、显式同步时发送一次：

- `boss.id`
- `boss.name`
- `boss.status` 的初始值
- `boss.maxHp`
- `boss.parts[*].index`
- `boss.parts[*].x`
- `boss.parts[*].y`
- `boss.parts[*].type`
- `boss.parts[*].label`
- 其他布局和展示所需静态字段

原则：

- 前端一旦拿到同一个 `boss.id` 的静态结构，就本地缓存
- 后续高频包不再重复发送这些字段

### 4.2 高频动态状态

这部分进入 `boss_delta`：

- `bossId`
- `version`
- `currentHp`
- `status`（仅在变化时）
- 变动部位列表
  - `index`
  - `currentHp`
  - `alive`
- `bossLeaderboardCount` 这类必要计数

原则：

- 只发变化字段
- 只发变化部位
- 不发未变动部位

### 4.3 保持独立的即时玩家反馈

当前 `click_ack` 已经承担：

- 点击确认
- 伤害数字
- 当前玩家资源变化
- 天赋触发事件

这条链路继续保留，不并入 `boss_delta`。

也就是说：

- `click_ack` 负责“即时手感”
- `boss_delta` 负责“公共战斗态收敛”

## 五、协议设计

### 5.1 首包 `snapshot`

首次连接、手动重同步、切房、切 Boss 时，下发完整战斗快照。

首包至少包含：

- `roomId`
- `bossId`
- `version`
- `boss static`
- `boss runtime`

其中：

- `boss static` 用于本地建模
- `boss runtime` 用于初始化当前动态值

前端在本地缓存：

- `bossStaticById[bossId]`
- `bossRuntime`
- `bossVersion`

### 5.2 高频包 `boss_delta`

高频下行只包含最小增量。

建议结构：

```json
{
  "type": "boss_delta",
  "bossId": "Compiler_Principles-30253",
  "version": 10241,
  "currentHp": 184230,
  "status": "active",
  "parts": [
    {"index": 3, "currentHp": 1200, "alive": true},
    {"index": 7, "currentHp": 0, "alive": false}
  ]
}
```

说明：

- `index` 取代部位名，避免重复长文本
- `currentHp` 使用绝对值而不是差值，降低前端累计误差风险
- `status` 仅在变化时发送

### 5.3 手动重同步 `sync_request`

当前已有 `sync_request` 入口，可直接复用，但要收窄触发条件。

触发时机：

- 前端本地缺失当前 `bossId` 的静态结构
- 前端发现 `bossId` 不匹配，且无法在本地直接切换
- 收到的 delta 缺关键字段，无法安全 merge
- 确认真实丢包或乱序，而不是服务端假跳号
- 页面恢复前台后主动请求一次（可保留为保守自愈策略）

服务端返回：

- 新的完整 `snapshot`

## 六、版本号策略

### 6.1 版本号职责

每个 `Boss` 动态状态维护单调递增 `version`，且它是“状态版本”，不是“发包版本”。

要求：

- 同一 `bossId` 下严格递增
- 只有会改变公共 `Boss` 运行态的事件才推进版本号
- 新 `Boss` 生成后版本号从新基线开始
- 同一轮广播里的所有订阅者看到的同一只 Boss，必须拿到同一个版本号
- `snapshot` 首包、SSE 首包、重复向多个客户端发同一状态，都不能单独推进版本号

### 6.2 版本推进边界

会推进版本号的事件：

- Boss 总血量变化
- Boss 任一部位 `currentHp / alive` 变化
- Boss `status` 变化
- Boss 切换

不会推进版本号的事件：

- `totalVotes` 变化但 Boss 运行态没变
- 在线人数变化
- 低频榜单变化
- 首次连接发 `snapshot`
- SSE 首包
- 同一轮 flush 中向多个客户端重复发送同一 Boss 状态

### 6.3 前端应用规则

前端收到 `boss_delta` 后按以下规则处理：

1. `bossId` 不一致
   - 若本地已有对应静态结构，则直接切换到该 Boss 基线
   - 否则触发 `sync_request`
2. `delta.version == local.version + 1`
   - 正常应用
3. `delta.version <= local.version`
   - 当作旧包或重复包丢弃
4. `delta.version > local.version + 1`
   - 不立刻认定异常，先确认是否存在服务端批次跳号
   - 只有确认无法安全继续时才 `sync_request`

这样可以保证：

- 正常路径下完全无感
- 异常路径下自动自愈
- 不会因为服务端错误推进版本而把 `snapshot` 变成高频常态

## 七、前端行为设计

### 7.1 本地状态结构

前端把 `Boss` 状态拆成两层：

- 静态缓存：`bossStaticById`
- 动态态：`bossRuntime`

组合渲染时：

- 通过 `bossId` 把两者拼成当前页面使用的 `boss`

### 7.2 收到 `boss_delta` 时的处理

处理顺序：

1. 校验 `bossId`
2. 校验 `version`
3. 更新 `bossRuntime.currentHp`
4. 逐个 merge `parts[*]`
5. 推进本地 `version`
6. 触发已有动画与视图刷新

这样前端页面不需要改成全新渲染模型，只是：

- 从“整对象替换”
- 改成“局部 merge”

### 7.3 用户无感的保证条件

只要满足下面三点，用户不会感知协议变化：

- `click_ack` 即时反馈保持不变
- `boss_delta` 在同样时机到达
- 版本跳号时能自动补全

## 八、服务端行为设计

### 8.1 首包职责

`sendSnapshot` 仍然负责返回完整初始战斗态，但要明确区分：

- 静态结构
- 动态状态

### 8.2 高频广播职责

普通点击、挂机伤害、流血 tick、魔法结算等高频路径，不再广播完整 `Boss`，只广播：

- `bossId`
- `version`
- 动态变化字段

### 8.3 必须发送全量的场景

以下场景禁止只发 delta：

- 首次连接
- `sync_request`
- 切房
- 新 Boss 替换旧 Boss
- Boss 击杀并生成下一只
- 后台修改 Boss 结构

## 九、为什么不做双协议兼容

这次明确不做新旧协议并行，原因很直接：

- 仓库当前默认前后端同批发布
- 增加兼容层会放大实现复杂度和测试成本
- 本次目标是减重，不应该顺手引入一整层历史协议分支

因此方案固定为：

- 同批发布
- 一次切换
- 协议不保留兼容期

## 十、实施顺序

建议分 4 步落地：

### 第一步：定义 schema

- 明确 `snapshot` 中的 `boss static` / `boss runtime`
- 新增 `boss_delta`
- 明确 `parts[*].index` 编码规则

### 第二步：服务端生成 delta

- 高频战斗路径生成变动部位列表
- 为公共 `Boss` 动态态维护版本号
- 下行改发 `boss_delta`

### 第三步：前端改为本地 merge

- 缓存 `bossStaticById`
- 本地维护 `bossRuntime + version`
- 收到 `boss_delta` 时按版本合并

### 第四步：补异常同步闭环

- 跳号自动 `sync_request`
- `bossId` 不匹配自动重拉
- 首包与切房路径保证完整覆盖

## 十一、测试要求

必须覆盖以下场景：

### 11.1 首包初始化

- 首次连接收到完整 `snapshot`
- 前端能建立静态缓存和动态态基线

### 11.2 普通点击 delta

- 普通点击只下发变化部位
- 不重复下发 `Boss` 名、部位名、静态布局

### 11.3 多部位同时变化

- 一次事件里多个部位掉血时，前端能正确 merge

### 11.4 Boss 击杀切换

- 旧 `bossId` 结束
- 新 `bossId` 到达完整快照
- 前端不会把旧 delta 合并到新 Boss

### 11.5 跳号重同步

- 人工制造 `version` 跳号
- 前端自动发 `sync_request`
- 收到全量后恢复正常

### 11.6 旧包乱序

- 收到较旧版本 delta
- 前端正确丢弃

## 十二、风险与取舍

主要风险：

- 服务端 delta 生成不完整，导致前端局部状态漏更新
- 前端 merge 逻辑写错，导致局部状态脏读
- 版本号推进点漏掉，导致跳号或伪连续

本次的主动取舍：

- 不追求一步到位把 `room_state` 也改成 delta
- 不在本轮动 Redis 热存储格式
- 不在本轮继续扩协议兼容层

这样可以把复杂度集中在：

- `Boss` 高频实时下行

而不是把问题扩散到整条实时系统。

## 十三、评审重点

本次 review 请重点看 5 件事：

- `boss static / boss runtime / boss_delta` 的边界是否清楚
- `index` 替代部位名是否满足前端渲染需要
- 版本号与 `sync_request` 兜底是否足够稳
- 哪些场景必须全量、哪些场景必须 delta，是否有漏项
- 是否还需要为排行榜或其他公共计数单独拆包
