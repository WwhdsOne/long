# 购买体力与登录接入 Cloudflare Turnstile 实施说明

## 本次范围

- 给 `POST /api/shop/stamina/full/purchase` 和 `POST /api/player/auth/login` 接入 Cloudflare Turnstile。
- 前端继续复用现有“确认购买体力”弹窗，不新增独立页面。
- 前端登录继续复用现有登录弹窗，不新增独立页面。
- 后端负责决定是否触发验证，首版按 `purchase_stamina_sample_rate` 做独立抽样。
  - 购买体力按抽样决定是否验证。
  - 登录在 Turnstile 开关开启时每次都要求验证。

## 后端改动

- 在 `backend/internal/config/config.go` 新增 `turnstile` 配置：
  - `enabled`
  - `site_key`
  - `secret_key`
  - `purchase_stamina_sample_rate`
  - `verify_timeout_ms`
- 购买体力路由支持可选 JSON 请求体：
- 登录路由也支持可选 JSON 请求体：

```json
{
  "turnstileToken": "optional-string"
}
```

- 新增本地 Turnstile 服务组件，负责：
  - 判断本次购买是否命中验证
  - 判断本次登录是否必须验证
  - 使用 `secret + response + remoteip` 调用 Cloudflare `siteverify`
  - 返回 `allow / require / invalid / unavailable`
- 后端错误码新增：
  - `CAPTCHA_REQUIRED`
  - `CAPTCHA_INVALID`
  - `CAPTCHA_VERIFY_UNAVAILABLE`
- 访问日志对 `turnstileToken` 做脱敏，避免验证码 token 落库。

## 前端交互

- `frontend/src/pages/publicPageState.js` 的购买体力和登录请求都改为可携带 `turnstileToken`。
- `frontend/src/pages/ShopPage.vue` 在原确认弹窗内按需渲染 Turnstile 容器。
- `frontend/src/pages/PublicPage.vue` 在原登录弹窗内按需渲染 Turnstile 容器。
- 首次点击确认后，如果服务端返回 `CAPTCHA_REQUIRED`：
  - 弹窗内显示“本次购买需要完成人机验证”
  - 动态加载 Cloudflare 官方脚本
  - 渲染 Turnstile 小组件
- 用户验证成功后，前端自动携带 token 重试同一购买请求。
- 登录弹窗命中验证时会显示“登录前需要先完成人机验证”，验证成功后自动带 token 重试登录。
- 关闭弹窗、购买成功、验证失败、验证过期时，都会清理当前 token；关闭弹窗和成功购买时还会清空验证 UI。
- 关闭登录弹窗、登录成功、验证失败、验证过期时，也会同步清理登录验证码状态。

## 配置建议

`backend/config.example.yaml` 已补充示例：

```yaml
turnstile:
  enabled: false
  site_key: ""
  secret_key: ""
  purchase_stamina_sample_rate: 0.5
  verify_timeout_ms: 3000
```

建议上线前先在 Consul 中补齐该段配置，再打开 `enabled`。
