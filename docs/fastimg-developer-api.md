# FastImg Developer API

## 1. API 目标

Personal API Token 用于个人脚本、PicGo、ShareX、CI、博客构建工具和自建应用。API 必须让用户完成完整闭环：

```text
生成 Token -> 上传图片 -> 等待处理 -> 获取各种链接 -> 删除自己的图片
```

API 与网页端共享用户、媒体、配额、审核、存储和用量服务，不维护第二套上传逻辑。

> 说明：API 不使用 `developer` URL 命名空间。Token 客户端只使用本文定义的上传和本人媒体接口；调用方由登录会话、`Authorization: Bearer <Personal API Token>` 或 `X-API-Key: <Personal API Token>` 区分。请求不需要传递 `developer` 参数，也不需要提交 `developer: true`。

## 2. 基本信息

```text
Base URL: https://img.example.com/api/v1
认证方式: Authorization: Bearer <Personal API Token> 或 X-API-Key: <Personal API Token>
上传格式: multipart/form-data
字符集: UTF-8
```

Token 不允许通过 URL query、文件名、表单字段或 Cookie 传递。

## 3. 上传模式

### 3.1 普通上传

```text
POST /api/v1/uploads
POST /api/v1/uploads/batch
```

用于普通脚本、PicGo、ShareX 和小型图片上传。请求完成后服务端创建媒体并返回 `201` 或 `202`。

### 3.2 分片上传

```text
POST   /api/v1/uploads/sessions
POST   /api/v1/uploads/sessions/{id}/parts
POST   /api/v1/uploads/sessions/{id}/complete
GET    /api/v1/uploads/sessions/{id}
DELETE /api/v1/uploads/sessions/{id}
```

用于大文件、网络不稳定或需要断点续传的客户端。创建会话、上传分片和完成确认只接受会员网页登录会话；Personal API Token 只开放本文最小的单文件上传和本人媒体闭环，不开放分片、批量或重试接口。

普通上传和分片上传不能各自实现一套媒体写入逻辑：两者最终都必须进入同一个 `CompleteUpload` 用例，生成同样的 MediaAsset、Variant、用量流水和媒体状态。普通上传默认自动通过并进入发现页；发布后的举报、隐藏和拒绝由管理员治理。

## 4. Token 管理

### 4.1 生成 Token

```http
POST /api/v1/tokens
Authorization: Bearer <web-session-token>
Content-Type: application/json
```

```json
{
  "name": "my-blog",
    "scopes": [],
  "expires_at": "2027-09-23T00:00:00Z"
}
```

`expires_at` 可选：会员端可以选择“永久有效”（提交 `null`）或“自定义到期时间”（提交未来的 ISO 8601 时间）。服务端拒绝过去或等于当前时间的自定义时间；永久 Token 在数据库中保持 `NULL`，认证时不设置过期时间。

Token 固定拥有以下最小能力：

```text
upload:write
media:read
media:delete
```

创建响应中的 `token` 只返回一次：

```json
{
  "data": {
    "id": 12,
    "name": "my-blog",
    "token": "fst_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
    "prefix": "fst_xxxx",
    "scopes": ["upload:write", "media:read", "media:delete"],
    "expires_at": "2027-09-23T00:00:00Z",
    "created_at": "2026-09-23T00:00:00Z"
  }
}
```

### 4.2 Token 列表、删除和轮换

```http
GET    /api/v1/tokens
DELETE /api/v1/tokens/{id}
POST   /api/v1/tokens/{id}/rotate
```

列表只能返回：

```text
id, name, prefix, scopes, status, expires_at,
created_at, last_used_at, last_used_ip, usage_summary
```

删除是不可逆操作。轮换会创建新 Token 并立即撤销旧 Token；旧 Token 的请求不能因为存在旧的缓存而继续成功。`DELETE /api/v1/tokens/{id}` 只允许删除当前用户自己的 Token，删除后立即失效并从 Token 列表移除。

当前代码已提供上述 Token 管理 API 的 Service、Controller、OpenAPI 和会员端 `/tokens` 页面。Token 仅保存 SHA-256 hash，固定 Scope 为 `upload:write`、`media:read`、`media:delete`；会员资源认证已支持 `Authorization: Bearer fst_...` 和 `X-API-Key: fst_...`，并由 own-scope 服务端逻辑推导用户归属。

### 4.3 管理端 Token 运营

管理员通过 `/admin/api_tokens` 查看全站 Token 的非秘密元数据。该资源允许具备 `admin.api_tokens.update` 的管理员批量停用/撤销，允许具备独立 `admin.api_tokens.delete` 的管理员删除 Token；不提供创建、明文读取、`token_hash` 读取或完整 Token 导出。管理员 Personal API Token 不能进入 `/api/v1/admin/**`，后台仍使用框架会话和 RBAC。

## 5. 上传图片

### 5.1 单文件上传

```http
POST /api/v1/uploads
Authorization: Bearer fst_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
Idempotency-Key: blog-build-20260923-0001
Content-Type: multipart/form-data
```

字段：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `file` | binary | 是 | 图片文件 |
| `title` | string | 否 | 图片标题 |
| `folder_id` | integer | 否 | 当前用户自己的文件夹 |
| `album_id` | integer | 否 | 当前用户自己的相册 |
| `visibility` | enum | 否 | `private`、`link`、`public`，默认 `private` |
| `expires_at` | datetime | 否 | 分享过期时间，受套餐限制 |
| `tags` | string[] | 否 | 标签 |

服务端必须重新检测 MIME、文件头、大小和像素，不能信任客户端字段。

### 5.2 批量上传

```http
POST /api/v1/uploads/batch
Authorization: Bearer fst_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
Idempotency-Key: blog-build-batch-001
Content-Type: multipart/form-data
```

批量接口使用多个 `files[]` 字段。当前实现限制为最多 5 个文件、单文件最多 10,000,000 字节、批次总大小最多 50,000,000 字节；请求体还包含 multipart 边界和头部开销。当前按文件顺序串行处理，避免并发配额竞态。批量接口只接受会员网页登录会话，Personal API Token 不能调用；后续套餐权益化时只能收紧这些上限，不能绕过服务端再次校验。

```bash
curl -X POST "https://img.example.com/api/v1/uploads/batch" \
  -H "Authorization: Bearer $FASTIMG_TOKEN" \
  -H "Idempotency-Key: batch-20260923-001" \
  -F "files[]=@./cover.png" \
  -F "files[]=@./avatar.jpg" \
  -F "visibility=private"
```

批量接口返回每个文件独立的结果，不因单个文件失败而隐藏其他结果：

```json
{
  "data": {
    "items": [
      {"client_name": "cover.png", "status": "processing", "id": 1001, "upload_session_id": 2001},
      {"client_name": "avatar.jpg", "status": "ready", "id": 1002, "upload_session_id": 2002, "links": {"url": "...", "markdown": "..."}},
      {"client_name": "bad.svg", "status": "failed", "error": {"code": "UPLOAD_TYPE_NOT_ALLOWED"}}
    ],
    "summary": {"total": 3, "accepted": 2, "failed": 1}
  }
}
```

批量请求要求 `Idempotency-Key` 长度为 8 至 160 个字符。服务端会基于该批次键和文件顺序生成每个文件的内部幂等键；使用相同键、相同顺序重试会复用对应文件的上传会话，不会重复创建媒体。解析阶段失败（空批次、超过数量/大小、空文件或 multipart 无效）会整体拒绝；进入业务处理后允许部分成功，并统一返回 HTTP `207 Multi-Status`。每个失败项返回稳定的 `error.code`，不会把内部错误文本暴露给客户端。

### 5.3 ready 响应

当原图和请求的基础 Variant 已处理完成，返回 `201 Created`：

```json
{
  "data": {
    "id": 1001,
    "status": "ready",
    "filename": "cover.png",
    "mime_type": "image/png",
    "size_bytes": 183920,
    "width": 1600,
    "height": 900,
    "links": {
      "original": "https://img.example.com/i/abc/original.png",
      "url": "https://img.example.com/i/abc/image.png",
      "markdown": "![cover](https://img.example.com/i/abc/image.png)",
      "html": "<img src=\"https://img.example.com/i/abc/image.png\" alt=\"cover\">",
      "bbcode": "[img]https://img.example.com/i/abc/image.png[/img]"
    },
    "usage": {
      "storage_used": 183920,
      "storage_limit": 1073741824
    }
  },
  "request_id": "req_abc123"
}
```

`links` 是固定对象，只包含唯一存储的 `original` 与四种复制格式，不能返回猜测出来的 Variant URL。

ready 响应中的 `links` 会返回带 `APP_URL` 域名的绝对公开链接，包含 `original`、`url`、`markdown`、`html` 和 `bbcode`；公开地址使用 APP_KEY 签名，不暴露存储对象路径。客户端可以直接把这些值复制到论坛、博客和 Markdown 内容中，不需要再拼接域名或额外创建分享链接。展示尺寸由接收站的 CSS 或自身处理链决定。

公开链接每次成功返回的实际字节数都会记入媒体所属用户当前 UTC 月的 `bandwidth` 用量。套餐 `monthly_bandwidth_bytes` 为 `0` 表示不限量；达到非零上限后返回 `429 BANDWIDTH_QUOTA_EXCEEDED`，无法可靠计量时返回 `503 BANDWIDTH_METERING_UNAVAILABLE`，不会静默放行未计量流量。

### 5.4 processing 响应

如果审核、压缩或 Variant 生成尚未完成，返回 `202 Accepted`：

```json
{
  "data": {
    "id": 1001,
    "status": "processing",
    "status_url": "/api/v1/uploads/1001",
    "links": {}
  },
  "request_id": "req_abc123"
}
```

客户端应使用退避轮询，建议间隔为 1、2、4、8 秒，最大间隔由服务端文档规定。不能高频轮询造成新的限流。

### 5.5 查询上传结果

```http
GET /api/v1/uploads/1001
Authorization: Bearer fst_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

当前实现中，`GET /api/v1/uploads/{id}` 仍由会员网页登录会话保护；Personal API Token 调用会返回 `403 TOKEN_ENDPOINT_NOT_ALLOWED`。因此单文件上传返回 `202` 时，Token 客户端暂时不能通过该接口轮询处理状态，这是当前 API 闭环的已知缺口。下一步应将该只读状态接口改为 `media:read`，继续使用服务端按 Token 所属用户的 own-scope 校验；上传重试仍只开放给网页登录会话。

### 5.6 删除自己的图片

```http
DELETE /api/v1/media/1001
Authorization: Bearer fst_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

删除进入回收站并立即使该媒体的分享链接失效。物理对象由异步清理任务删除。该能力属于个人 Token 基础能力，不要求额外配置 `media:delete`。

回收站媒体可由所有者主动永久删除。该动作需要显式确认短语，后端先将状态锁定为 `cleanup_pending`，逐个幂等删除原图和派生图对象；全部成功后才在同一数据库事务中写入负向存储流水并将媒体标记为 `physically_deleted`。对象删除中断时保持 `cleanup_pending`，重复提交相同请求可安全重试；不得因为失败提前释放空间。

```http
DELETE /api/v1/media/{id}/permanent
Content-Type: application/json

{"confirm":"permanently-delete"}
```

仅媒体所有者可操作。媒体不在回收站或不属于当前用户时返回 `MEDIA_NOT_FOUND`；缺少精确确认短语返回 `MEDIA_DELETE_CONFIRMATION_REQUIRED`；存储对象仍被其他媒体引用时返回 `MEDIA_SHARED_STORAGE`。失败后仍处于 `cleanup_pending` 的项目会继续显示在回收站中，以便重试。该永久删除接口当前只接受会员网页登录会话，Personal API Token 不开放永久删除，避免脚本误删后无法恢复。

会员网页登录会话可以一次清空自己的回收站。该接口不开放给 Personal API Token，必须提交精确确认短语；服务端逐项执行幂等物理删除，返回本次完成数量，失败项目保留在回收站中：

```http
DELETE /api/v1/media/trash
Content-Type: application/json

{"confirm":"empty-trash"}
```

成功响应为 `{"data":{"deleted_count":2}}`；缺少确认短语返回 `MEDIA_TRASH_CONFIRMATION_REQUIRED`。

## 6. 查询自己的媒体

Personal API Token 固定具备 `media:read`，可以查询当前用户的媒体列表：

```http
GET /api/v1/media?page=1&per_page=20&status=ready
Authorization: Bearer fst_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

支持的查询参数：

```text
page, per_page, status, folder_id, album_id,
visibility, mime_type, created_from, created_to, search
```

分页响应必须使用统一格式：

```json
{
  "data": [{"id": 1001, "status": "ready", "links": {"url": "..."}}],
  "meta": {"page": 1, "per_page": 20, "total": 1, "last_page": 1}
}
```

当前新创建的 Token 不存在缺少 `media:read` 的情况；链接和媒体内容统一由 `media:read` 保护。

### 6.1 调整本人媒体文件夹

当前会员会话接口支持在图片上传完成后调整归档位置：

```http
PATCH /api/v1/media/1001/folder
Authorization: Bearer <session-token>
Content-Type: application/json

{"folder_id": 12}
```

`folder_id` 必须属于当前认证用户；传 `null` 表示移回未归档。跨用户或不存在的文件夹返回 `FOLDER_NOT_FOUND`，不属于当前用户或不是 `ready` 状态的媒体返回 `MEDIA_NOT_FOUND`。该接口暂不接受 `user_id`，也不允许通过管理员资源接口绕过会员端的所有权校验。当前只接受会员网页登录会话，Personal API Token 不开放文件夹归档。

### 6.2 创建和撤销分享链接

```http
POST /api/v1/media/1001/share-links
Authorization: Bearer <session-token>
Content-Type: application/json

{"expires_at":"2026-10-24T23:59:59Z","password":"至少八个字符"}
```

创建成功只返回一次完整 `/s/{token}` 地址；列表接口只返回 token 前缀，服务端不保存可恢复的明文 Token。`password` 可选，长度为 8–72 个字符，服务端只保存框架 Hash；设置密码的分享链接访问时使用 `GET /s/{token}?variant=original&password=...`，缺少或错误密码返回 `401 SHARE_PASSWORD_REQUIRED`。`GET /s/{token}` 会重新校验分享状态、过期时间、媒体状态、原图和当前防盗链策略，再读取私有对象。删除媒体、撤销链接或防盗链拒绝后不泄露对象存储地址。

### 6.3 签名 URL、防盗链与访问记录

```http
POST /api/v1/media/1001/signed-url
Authorization: Bearer <session-token>
Content-Type: application/json

{"variant":"original","expires_in":600}
```

`expires_in` 必须为 60–86400 秒；服务端使用 `APP_KEY` 以 HMAC-SHA256 绑定媒体 ID、Variant 和 Unix 过期时间，响应只返回应用地址 `/i/{id}?variant=...&expires=...&signature=...`，不返回 Local/S3 对象地址。APP_KEY 未配置时返回 `503 LINK_SIGNING_UNAVAILABLE`。签名 URL 仅在生成时返回，过期、篡改、媒体删除或 Variant 不可用统一返回 `404 LINK_NOT_FOUND`。

媒体所有者可以配置：

```http
GET /api/v1/media/1001/hotlink-policy
PUT /api/v1/media/1001/hotlink-policy
GET /api/v1/hotlink-domains
POST /api/v1/hotlink-domains
DELETE /api/v1/hotlink-domains/{id}
```

策略模式为 `off`、`referer`、`signed`、`hybrid`。Referer 域名会去除协议、路径和末尾点后转为小写保存，可选允许缺失 Referer；域名白名单按当前用户生效。`signed` 要求有效签名 URL，`referer` 要求请求 Referer 主机在白名单，`hybrid` 满足任一条件即可。公开 `/s/{token}` 分享链接同样执行策略判定，不能绕过 `signed` 模式。

每次公开投递会尽力写入 `media_access_logs`，记录媒体、Variant、投递模式、允许/拒绝结果、规范化 Referer 主机、User-Agent 和 UTC 时间；不保存签名、密码或完整 Token。访问记录写入失败不会让已允许的图片响应变成 5xx。当前 Local Provider 不支持原生对象签名 URL，因此 M3 使用应用层 HMAC；CDN 原生签名、缓存失效和下载带宽入账仍属于后续生产切片。

管理员可通过以下只读接口查询全站脱敏访问记录：

```http
GET /api/v1/admin/media-access-logs?page=1&per_page=20&media_id=1001&result=denied&delivery_mode=referer
Authorization: Bearer <admin-session-token>
```

该接口要求 `admin.media_access_logs.view`，支持 `media_id`、`variant`、`delivery_mode`、`result`、`referer_host` 筛选并返回 `{data, meta}` 分页结构；响应不包含 `signature`、分享/Personal Token、密码或对象存储源地址。管理页面为 `/admin/media-access-logs`，会员端不会看到全站访问记录。

## 7. 客户端示例

### curl

```bash
curl -X POST "https://img.example.com/api/v1/uploads" \
  -H "Authorization: Bearer $FASTIMG_TOKEN" \
  -H "Idempotency-Key: $(uuidgen)" \
  -F "file=@./image.png" \
  -F "visibility=link"
```

脚本应优先读取响应中的 `links.markdown` 或 `links.url`，不要自行拼接图片 URL。

### JavaScript

```js
const form = new FormData()
form.append('file', file)
form.append('visibility', 'link')

const response = await fetch('https://img.example.com/api/v1/uploads', {
  method: 'POST',
  headers: {
    Authorization: `Bearer ${import.meta.env.VITE_FASTIMG_TOKEN}`,
    'Idempotency-Key': crypto.randomUUID(),
  },
  body: form,
})

const result = await response.json()
const imageUrl = result.data.links?.url
```

生产前端不能把个人 Token 打包进公开浏览器代码；浏览器上传应使用用户登录会话或后端代理。

### Python

```python
import os
import requests

with open("image.png", "rb") as image:
    response = requests.post(
        "https://img.example.com/api/v1/uploads",
        headers={
            "Authorization": f"Bearer {os.environ['FASTIMG_TOKEN']}",
            "Idempotency-Key": "script-image-001",
        },
        files={"file": ("image.png", image, "image/png")},
        data={"visibility": "link"},
        timeout=60,
    )

response.raise_for_status()
print(response.json()["data"]["links"]["markdown"])
```

## 8. PicGo/ShareX 适配

### PicGo

提供官方配置示例：

```text
customUrl: https://img.example.com/api/v1/uploads
authHeader: Authorization: Bearer <token>
method: POST
fileField: file
responseUrlPath: data.links.url
responseMarkdownPath: data.links.markdown
```

实际 PicGo 插件配置应以插件支持的字段为准；服务端必须保证响应路径稳定。

### ShareX

提供可导入的 Custom Uploader 配置，包含：

- Request URL
- `Authorization` Header
- multipart 字段 `file`
- URL 路径 `data.links.url`
- 删除 URL `DELETE /api/v1/media/{id}`（如客户端支持）

Token 不写入仓库、截图配置公开链接或 CI 日志。

## 11. 当前 Personal API Token 最小开放面（以本节为准）

Personal API Token 不等同于会员登录会话。创建 Token 时不再提供可选 Scope，服务端固定授予以下三个 Scope：

```text
upload:write   POST /api/v1/uploads
media:read     GET  /api/v1/media、GET /api/v1/media/{id}、GET /api/v1/media/{id}/content
media:delete   DELETE /api/v1/media/{id}
```

兼容 PicGo、ShareX 和简单脚本的短路径如下，权限完全相同：

```text
POST   /api/upload
GET    /api/images
GET    /api/image/{image_id}
DELETE /api/image/{image_id}
```

认证支持 `Authorization: Bearer fst_...` 和 `X-API-Key: fst_...`。两种方式都只从 Token 所属用户推导资源归属，客户端不能提交 `user_id` 改变数据范围。

Personal API Token 请求按固定窗口计数：单个 Token 的每分钟上限读取当前套餐 `api_rate_per_minute`，同一来源 IP 每分钟最多 300 次。套餐值为 `0` 表示不限制 Token 维度，但仍保留 IP 保护；管理员 Token 不受会员套餐限额限制。网页登录会话不使用这组 Token 限额。允许请求会返回当前实际生效的上限：

```text
X-RateLimit-Limit: 30
X-RateLimit-Remaining: 29
X-RateLimit-Reset: 1790296519
```

超过 Token 限额时返回 `429 Too Many Requests` 和 `Retry-After` 秒数：

```json
{
  "code": "TOKEN_RATE_LIMITED",
  "retry_after_seconds": 58
}
```

限流计数只保存 Token/IP 的 SHA-256 派生键、请求数和窗口时间，不保存明文 Token；计数异常时返回 `503 TOKEN_RATE_LIMIT_STORE_UNAVAILABLE`，客户端应按退避策略重试。

### 11.1 统一错误响应与请求追踪

Personal API 的 JSON 错误响应统一包含以下字段：

```json
{
  "code": "TOKEN_ENDPOINT_NOT_ALLOWED",
  "request_id": "req_abc123",
  "retryable": false
}
```

服务端始终通过响应头返回 `X-Request-ID`。客户端可以提交安全的 `X-Request-ID`（最多 64 个 ASCII 字符，仅允许字母、数字、`.`、`_`、`:`、`-`）；不符合规则、过长或缺失时，服务端生成 `req_...` 标识。`request_id` 与响应头值保持一致，便于用户向站点管理员提供可检索的故障线索。

`retryable` 只表示是否值得按退避策略重试：429、500、502、503、504 为 `true`，认证、权限、参数、资源不存在和业务冲突类错误为 `false`。客户端必须优先按 `code` 处理业务，不应依赖错误文本；未知字段必须忽略。

Token 明确不能调用：批量上传、上传重试、套餐/用量、订单/支付、文件夹、相册、分享链接、防盗链配置、Webhook 和任何 `/api/v1/admin/**` 接口。上述功能必须使用会员网页登录会话；管理员后台必须使用管理员会话和 RBAC，管理员 Personal API Token 不能升级为后台权限。

当前认证矩阵如下，C 端 Token 不通过“隐藏菜单”限制，而是在后端路由中强制限制：

| 接口 | 会话 | Personal API Token | 权限/数据范围 |
| --- | --- | --- | --- |
| `POST /api/v1/uploads`、`POST /api/upload` | 允许 | 允许 | `upload:write`，只归属当前用户 |
| `GET /api/v1/media`、`GET /api/images` | 允许 | 允许 | `media:read`，只返回当前用户 |
| `GET /api/v1/media/{id}`、`GET /api/image/{id}` | 允许 | 允许 | `media:read`，只能读取自己的详情和链接 |
| `GET /api/v1/media/{id}/content` | 允许 | 允许 | `media:read`，只能读取自己的图片内容 |
| `GET /api/v1/uploads/{id}` | 允许 | 当前拒绝 | 已知缺口；下一步改为 `media:read` 的只读状态查询 |
| `DELETE /api/v1/media/{id}`、`DELETE /api/image/{id}` | 允许 | 允许 | `media:delete`，只能删除自己的图片到回收站 |
| 订单、用量、文件夹、相册、分享、防盗链、批量/重试 | 仅会话 | 拒绝 | 返回 `403 TOKEN_ENDPOINT_NOT_ALLOWED` |
| `/api/v1/admin/**` | 管理员会话 + RBAC | 拒绝 | 管理员全站能力由后台权限和数据范围决定 |

旧文档中出现的 `links:read`、`usage:read`、`webhook:manage` 和 Token 恢复能力不再作为新 Token 的开放选项；Token 删除通过会员端自己的 `DELETE /api/v1/tokens/{id}` 和后台独立删除权限提供，保留旧记录只为兼容历史数据，路由以当前最小开放面为准。

管理员 API 不是 Personal API Token 的用途：`/api/v1/admin/**` 仅接受管理员登录会话，并由 `admin.*` RBAC 权限、数据范围和审计规则控制。管理端的 `api_tokens` 资源只能查看 Token 的运营元数据，并由独立权限控制停用/撤销和删除，不能创建、读取明文或导出 Token；这样不会把管理员权限下放给任何 C 端 Token。

## 12. 套餐、价格和支付渠道边界

`plans` 与 `plan_prices` 是不同的业务数据表，但开发阶段不提供独立的 `plan_prices` 后台资源：

- `plans` 只维护产品身份、说明、排序、状态和 `entitlements_json` 权益额度。
- `plan_prices` 由 Seeder/结算 Service 维护可售卖的不可变价格版本：计划、版本、金额、币种、月/年周期、试用天数和生效区间。
- 公开套餐目录、当前订阅响应和管理端计划展示均不再输出 `plans.price_amount/currency/billing_period` 这组三个旧计划级价格字段；付费展示和下单只以活动 `plan_prices` 为准。`plan_prices` 不包含 `gateway_code`，价格表不再和支付渠道耦合。
- 价格不再绑定唯一支付网关。会员创建订单后，在结算页从当前已启用且已配置的渠道中选择 PayPal、XCash、NOWPayments 或开发测试网关；订单保存价格快照，支付意图保存用户本次选择的 Provider。

支付渠道的启用条件是“管理员开关 + 服务端必需密钥/API 地址配置 + 重启后注册成功”同时满足。仅勾选开关不会把未配置的渠道展示给会员。后台设置页按渠道填写 Appid/API Key、HMAC/IPN Secret、PayPal Client ID/Secret、Webhook/回调和成功/取消回跳地址；敏感值使用 `APP_KEY` 保护的 AES-GCM 密文保存，读取接口只返回“已配置”占位符，留空表示保持原值。环境变量 `PAYPAL_*`、`XCASH_*`、`NOWPAYMENTS_*` 仍作为部署级回退配置。会员结算页只展示实际注册成功的渠道，因此可以同时启用多个渠道并由会员自行选择。

## 13. SEO、站点地图和 SSG

- 后端公开提供 `/sitemap.xml` 和 `/robots.txt`，默认收录 `/`、`/plans`、`/discover`，禁止爬取 `/admin/` 和 `/api/`；生产环境通过 `APP_URL` 设置规范域名。
- 开发环境前端与后端使用独立端口时，管理端打开按钮和 Vite 代理都必须把这两个文件转发到后端；不能让前端 SPA fallback 接管 XML/文本响应。当前本地验证端口为前端 `53084`、后端 `53085`。
- 管理员设置的 `sitemap.enabled` 控制公开站点地图，`sitemap.extra_paths` 接受换行、逗号或分号分隔的公开路径；服务端会过滤 `/admin`、`/api`、查询串和锚点。
- 会员首页和套餐页在 SPA 运行时同步设置 title、description、Open Graph 和 canonical。
- 前端提供 `npm run build:ssg`：先构建 Vite，再根据公开路由清单生成 `/`、`/plans`、`/discover` 和通过 `SSG_PUBLIC_ALBUM_IDS` 指定的 `/a/:id/` 静态 HTML、canonical、Open Graph 和 JSON-LD。登录、上传、个人媒体、订单等私有页面不做静态暴露；相册只有在 API 返回 `visibility=public` 时才会生成。
- SSG 页面只是公开内容的 SEO 首屏，保留 Vue `#app` 挂载点后继续使用真实 API；不会把 Token、用户媒体或管理数据写入静态 HTML。当前没有请求级 SSR，不能以 SSG 产物替代 SSR 的动态首屏能力。

## 14. 公开发现 API

发现页是游客可读的公开能力，不接受 Personal API Token，也不返回未审核或私有媒体。

| 方法 | 路径 | 认证 | 说明 |
| --- | --- | --- | --- |
| `GET` | `/api/v1/discovery/status` | 无 | 返回发现页 `enabled` 状态 |
| `GET` | `/api/v1/discovery/feed?page=1&per_page=24` | 无 | 分页返回 `ready`、公开且未被拒绝的媒体；`per_page` 限制为 1–48 |
| `GET` | `/api/v1/discovery/media/{id}/content?variant=original` | 无 | 返回已审核媒体原图；Variant 仅允许 `original` |

`feed` 返回 `{data: [...], meta: {page, per_page, total}}`；每项包含媒体 ID、文件名、尺寸、类型、时间和 `thumbnail_url`/`original_url`。公开列表的服务端过滤条件固定为 `status=ready`、未软删除、`visibility=public`、`moderation_status != rejected`。普通新上传默认 `public + approved`，可立即进入发现页并通过稳定链接使用；隐藏或拒绝后从发现页移除，恢复后重新出现。管理员在 `/admin/media` 执行隐藏、恢复、审核通过、审核拒绝和永久删除，不能通过通用 CRUD 直接改写状态；举报进入 `/admin/reports`。

当前 `/discover` 是会员应用中的真实公开 SPA 页面，并已纳入 `build:ssg` 的公开路由清单；这不是 SSR。后续独立 `web/` SSR 入口必须继续复用以上公开 API 和审核边界，不得把管理员资源或用户私有媒体注入首屏。

## 15. 会员端展示兜底

```http
GET /api/v1/site/presentation
```

该接口无需登录，只返回会员端图片加载失败所需的非敏感配置：

```json
{
  "data": {
    "watermark_fallback_image_url": "https://img.example.com/public/image-unavailable.svg"
  }
}
```

后台设置 `watermark.fallback_image_url` 只接受公网 `http/https` 地址或站内绝对路径；非法值、空值和 `javascript:`/`data:` 等协议不会下发，前端使用内置 FastImg 兜底水印图。该接口不返回支付凭证、站点自定义代码或其他管理员设置。

## 15. 错误码

| HTTP | Code | 说明 |
| ---: | --- | --- |
| 401 | `TOKEN_INVALID` | Token 不存在、格式错误或已撤销 |
| 401 | `TOKEN_EXPIRED` | Token 已过期 |
| 403 | `USER_SUSPENDED` | 所属用户被禁用 |
| 403 | `MEDIA_ACCESS_DENIED` | 媒体不属于当前用户 |
| 413 | `UPLOAD_SIZE_EXCEEDED` | 超过单文件限制 |
| 413 | `UPLOAD_BATCH_TOO_LARGE` | 超过批次总大小限制 |
| 400 | `UPLOAD_BATCH_TOO_MANY_FILES` | 超过批次文件数量限制 |
| 415 | `UPLOAD_TYPE_NOT_ALLOWED` | 图片格式不支持 |
| 422 | `UPLOAD_INVALID_IMAGE` | 文件不是可处理图片 |
| 422 | `MEDIA_DELETE_CONFIRMATION_REQUIRED` | 永久删除缺少精确确认短语 |
| 409 | `IDEMPOTENCY_CONFLICT` | 相同幂等键对应不同请求 |
| 409 | `MEDIA_SHARED_STORAGE` | 存储对象仍被其他媒体引用，不能安全删除 |
| 429 | `RATE_LIMITED` | Token、用户或 IP 达到速率限制 |
| 429 | `MONTHLY_API_UPLOAD_LIMIT_REACHED` | 月度 API/上传次数达到套餐上限 |
| 429 | `MONTHLY_TRANSFORM_LIMIT_REACHED` | 月度图片处理作业达到套餐上限 |
| 409 | `BATCH_PARTIAL_FAILURE` | 批量上传部分成功、部分失败 |
| 429 | `QUOTA_EXCEEDED` | 存储、流量、上传或 API 配额不足 |
| 503 | `UPLOAD_QUEUE_UNAVAILABLE` | 上传后处理队列不可用 |
| 404 | `LINK_NOT_FOUND` | 签名 URL 过期、篡改、媒体不可用或防盗链拒绝；不区分具体原因 |
| 422 | `SIGNED_URL_EXPIRY_INVALID` | 签名 URL 有效期不在 60–86400 秒范围内 |
| 422 | `HOTLINK_POLICY_INVALID` | 防盗链模式不受支持 |
| 422 | `HOTLINK_DOMAIN_INVALID` | Referer 域名格式无效 |
| 409 | `HOTLINK_DOMAIN_EXISTS` | 当前用户已配置该域名 |
| 503 | `LINK_SIGNING_UNAVAILABLE` | 服务端未配置 APP_KEY，暂时不能生成签名 URL |

## 16. 安全和配额

- 默认 Token 只能管理所属用户自己的媒体。
- 每个 Token、用户和 IP 分别限流。
- Token 认证成功时只更新最后使用时间和统计，不记录原文。
- 上传失败不扣最终存储额度；预占额度必须释放或进入可追踪状态。
- 已知超出套餐的上传会在图像解码前拒绝；服务器仍会在数据库事务中重新检查配额并预占实际原图/Variant 存储，处理并发竞态。
- 相同幂等键重复请求返回第一次结果，不重复创建媒体。
- Free 用户的上传、链接和删除能力可用，但仍受存储、流量、文件大小和速率限制。
- 服务器端生成链接，客户端不得根据媒体 ID、哈希或文件名拼接 URL。

## 17. 版本兼容

- 所有开发者 API 放在 `/api/v1` 下。
- 新增链接类型只能新增 `links` 字段，不能修改已有 `url`、`markdown`、`html` 和 `bbcode` 的含义。
- 客户端必须忽略未知字段和未知链接键。
- 弃用字段至少保留一个版本周期，并在 OpenAPI 和变更日志中标记。
- 错误码不能复用为其他业务语义。
- 批量上传和单文件上传必须共享相同的文件校验、配额、审核和用量规则。
