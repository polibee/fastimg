# FastImg 防盗链设计

## 1. 目标

防盗链用于降低未经授权的第三方嵌入和流量消耗，不能替代用户权限、分享链接过期、Token 鉴权或套餐流量限制。

设计目标：

- 用户可以控制图片允许出现在哪些站点。
- 私有媒体不能因为防盗链配置而变成公开媒体。
- 防盗链失败不能泄露对象存储真实地址。
- 图片访问主链路不依赖前端 JavaScript。
- 在 Referer 缺失、浏览器隐私策略和 CDN 缓存场景下有明确行为。

## 2. 防盗链模式

### Off

不检查来源，只执行媒体可见性、分享策略、Token/签名和套餐流量限制。Free 默认使用此模式或基础保护模式。

### Referer Allowlist

只允许配置的域名来源访问：

```text
https://example.com/*
https://www.example.com/*
https://blog.example.org/*
```

服务端只比较规范化后的 scheme/host/port，不使用任意字符串包含判断。`*.example.com` 是否包含根域名由明确规则决定，不能隐式扩大匹配范围。

### Signed URL

访问 URL 包含短期签名：

```text
/i/asset/medium.webp?expires=...&signature=...
```

签名内容至少包含：媒体/Variant 标识、过期时间、策略版本和必要的下载参数。签名 URL 适合私有分享、下载和高风险资源，不适合作为永久嵌入链接。

### Hybrid

先校验公开性和分享策略，再按以下顺序执行：

1. 合法的短期签名直接允许。
2. 配置的 Referer 域名允许访问。
3. 无 Referer 按站点策略处理。
4. 其他来源拒绝或返回配置的占位图。

Creator/Pro 默认推荐 Hybrid；Pro 可以为自定义域名启用强制 Signed URL。

## 3. 套餐策略

| 套餐 | 防盗链建议 |
| --- | --- |
| Free | Off 或基础 Referer 保护；限制白名单域名数量 |
| Creator | Referer Allowlist，支持多个站点和无 Referer 策略 |
| Pro | Hybrid、Signed URL、自定义域名、规则优先级和访问统计 |

防盗链不是强制用户升级的唯一方式。Free 用户仍可使用稳定公开 URL，但受流量、速率和异常访问限制。

## 4. Referer 规则

### 4.1 缺失 Referer

Referer 可能因浏览器隐私策略、HTTPS 到 HTTP 降级、App、命令行或安全策略而缺失。必须由用户选择：

- `allow_missing`：允许无 Referer 请求，适合兼容性优先。
- `deny_missing`：拒绝无 Referer，适合流量保护优先。
- `signed_only`：无 Referer 只能使用有效签名。

默认不应把无 Referer 自动当作恶意请求；用户可在 Creator/Pro 中配置。

### 4.2 域名匹配

- 只接受 `http`/`https` 的规范化来源。
- 默认忽略 Referer path，只比较 host；高级规则可以选择 path 前缀。
- IDN 域名统一转为规范形式后比较。
- 端口不同视为不同来源，除非用户明确配置。
- 不允许把 `example.com.attacker.com` 匹配为 `example.com`。
- 白名单变更必须记录审计事件。

## 5. 请求处理顺序

```text
请求图片
  -> 解析媒体/Variant
  -> 检查媒体存在、未删除、处理 ready
  -> 检查公开性/分享密码/分享过期
  -> 检查 Signed URL
  -> 检查 Referer Allowlist
  -> 检查用户、Token、IP、域名和流量规则
  -> 允许对象读取或返回拒绝响应
```

任何拒绝都不能返回对象存储真实地址、数据库 ID 详情或内部错误。

## 6. 失败响应

默认返回 `403` 和稳定错误码：

```json
{
  "error": {
    "code": "HOTLINK_BLOCKED",
    "message": "Image access is not allowed from this source",
    "request_id": "req_..."
  }
}
```

可配置响应模式：

- `403`：默认，适合 API 和程序调用。
- 占位图：适合公开网页，但占位图本身必须防止递归引用。
- 统一关闭页：只能使用站点固定地址，禁止把用户输入拼成跳转地址。

禁止默认 302 到任意外部 URL，避免形成开放重定向和额外流量放大。

## 7. 稳定链接与签名链接的区别

| 类型 | 是否稳定 | 适用场景 |
| --- | --- | --- |
| Public URL | 是 | 公开博客、论坛、文档 |
| Referer-protected URL | 是 | 允许固定站点嵌入 |
| Signed URL | 否，受过期时间限制 | 私有分享、临时下载、高风险资源 |
| API response URL | 根据媒体策略 | 返回给用户或客户端使用 |

稳定 URL 不能保证任何地方永久可用；媒体删除、分享撤销、账户处罚、套餐流量耗尽和站点关闭都可以使访问失效。

## 8. 用户端配置

防盗链页面提供：

- 当前模式
- 允许的域名列表
- 无 Referer 策略
- 失败响应模式
- 是否对原图和 Variant 使用相同规则
- 当前规则命中/拒绝统计
- 最近被拦截的来源摘要

新增域名必须经过格式校验和归一化。用户不能配置任意响应头、脚本、跳转 URL 或对象存储路径。

## 9. 管理端配置

管理员可以：

- 全局启用/停用防盗链默认策略。
- 为套餐设置域名数量、签名 TTL、无 Referer 默认策略和流量限制。
- 查看异常来源、媒体、用户、Token 和 Referer 统计。
- 对单用户、单媒体或单域名临时封禁。
- 在安全事件中暂停高风险来源或撤销相关 Token。

全局策略不能绕过用户私有媒体的访问控制，只能进一步收紧访问。

## 10. API

```text
GET    /api/v1/media/{id}/hotlink-policy
PUT    /api/v1/media/{id}/hotlink-policy
GET    /api/v1/hotlink-domains
POST   /api/v1/hotlink-domains
DELETE /api/v1/hotlink-domains/{id}
POST   /api/v1/media/{id}/signed-url
```

Token 可以管理自己媒体的防盗链配置，但不能修改其他用户媒体。高级能力受套餐权益限制；基础公开链接和删除自己媒体的能力不应因为没有高级防盗链权益而消失。

## 11. CDN 和缓存约束

- CDN 必须在缓存命中前执行签名、Referer 和策略校验，或使用安全的边缘鉴权。
- 私有/签名 URL 不能被公共缓存复用。
- 缓存 Key 必须区分媒体、Variant、策略版本和必要签名参数。
- 改变白名单、撤销分享或删除媒体后，必须有缓存失效或短 TTL 策略。
- 访问拒绝响应不能被公共 CDN 长时间缓存。

## 12. 统计和告警

记录最小必要信息：媒体 ID、用户 ID、Variant、命中/拒绝、规则类型、域名摘要、状态码和时间。

告警条件：

- 单媒体短时间大量拒绝。
- 单域名或 IP 突发流量。
- 无 Referer 请求异常增长。
- 签名校验失败异常增长。
- CDN 回源流量与用户用量不一致。

## 13. 测试要求

- 允许域名、子域名和端口匹配正确。
- `example.com.attacker.com` 不能匹配 `example.com`。
- 缺失 Referer 按配置处理。
- 有效签名可以在无 Referer 时访问。
- 过期或篡改签名返回 `SIGNED_URL_INVALID`。
- 私有媒体不会因 Referer 白名单变成公开。
- 删除、撤销分享和封禁账户后旧 URL 按策略失效。
- 拒绝响应不泄露对象存储地址。
- CDN 缓存不会绕过策略变化。
