你当前的实现是通过 adaptor.HertzHandler 包装 net/http handler，这只是过渡方案，不符合要求。

现在必须进行重构：**禁止使用 adaptor，必须改为原生 Hertz handler 实现。**

------

## 🎯 目标

将现有基于 net/http 的 handler：

```go
func(w http.ResponseWriter, r *http.Request)
```

全部重写为 Hertz 原生 handler：

```go
func(ctx context.Context, c *app.RequestContext)
```

------

## ❌ 明确禁止

以下内容一律不允许出现：

- adaptor.HertzHandler
- http.ResponseWriter
- *http.Request
- 任何 net/http handler 包装

如果出现，说明没有按要求完成

------

## ✅ 必须做到

### 1. 使用 Hertz 原生 API

- 路由参数：`c.Param(...)`
- Query 参数：`c.Query(...)`
- JSON 解析：`c.Bind(...)` 或 `c.BindAndValidate(...)`
- JSON 返回：`c.JSON(status, data)`

------

### 2. Handler 结构必须保持轻量

禁止在 handler 中写复杂业务逻辑：

```go
// ❌ 错误
func handler(...) {
    // 大量业务逻辑
}
```

必须：

```go
// ✅ 正确
func handler(...) {
    // 解析参数
    // 调用 store/service
    // 返回结果
}
```

------

### 3. 保留现有业务逻辑

必须继续调用：

- options.Store.xxx(...)
- publishChange(...)
- 原有错误处理逻辑（errors.Is 判断）

只允许改“HTTP 层”，不允许改业务行为

------

### 4. 错误返回统一使用 c.JSON

例如：

```go
c.JSON(400, map[string]string{
    "error": "INVALID_REQUEST",
})
```

------

### 5. 示例改造（参考风格）

将：

```go
var body struct {
    Nickname string `json:"nickname"`
}
if err := sonic.NewDecoder(r.Body).Decode(&body); err != nil {
    writeJSON(w, 400, ...)
}
```

改为：

```go
var body struct {
    Nickname string `json:"nickname"`
}
if err := c.Bind(&body); err != nil {
    c.JSON(400, ...)
    return
}
```

------

## 🚀 执行步骤

1. 删除 adaptor 相关代码
2. 重写所有 handler 为 Hertz 风格
3. 保持原有接口路径不变
4. 保持所有返回结构一致

------

## ⚠️ 输出要求

- 必须是多文件结构（不要生成一个巨型文件）
- handler 按模块拆分（equipment / hero / admin 等）
- 不允许出现 500 行以上文件

------

请开始重构，并先输出改造 plan（文件拆分 + 路由结构），确认后再写代码。
