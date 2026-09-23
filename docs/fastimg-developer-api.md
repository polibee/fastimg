# FastImg Developer API

## 1. API 目标

Personal API Token 用于个人脚本、PicGo、ShareX、CI、博客构建工具和自建应用。API 必须让用户完成完整闭环：

```text
生成 Token -> 上传图片 -> 等待处理 -> 获取各种链接 -> 删除自己的图片
```

API 与网页端共享用户、媒体、配额、审核、存储和用量服务，不维护第二套上传逻辑。

> 说明：API 不使用 `developer` URL 命名空间。上传、媒体和配额接口使用统一的用户资源路径；调用方由登录会话或 `Authorization: Bearer <Personal API Token>` 区分。请求不需要传递 `developer` 参数，也不需要提交 `developer: true`。

## 2. 基本信息

```text
Base URL: https://img.example.com/api/v1
认证方式: Authorization: Bearer <Personal API Token>
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

用于大文件、网络不稳定或需要断点续传的客户端。创建会话、上传分片和完成确认都必须携带登录会话或 Personal API Token，并执行相同的用户归属和配额校验。

普通上传和分片上传不能各自实现一套媒体写入逻辑：两者最终都必须进入同一个 `CompleteUpload` 用例，生成同样的 MediaAsset、Variant、用量流水和审核任务。

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
  "scopes": ["media:read", "usage:read"],
  "expires_at": "2027-09-23T00:00:00Z"
}
```

Token 固定拥有以下基础能力：

```text
upload:write
links:read
media:delete
```

基础能力只能操作当前用户自己的媒体。可选能力由用户勾选：

```text
media:read
usage:read
webhook:manage
```

创建响应中的 `token` 只返回一次：

```json
{
  "data": {
    "id": 12,
    "name": "my-blog",
    "token": "fst_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
    "prefix": "fst_xxxx",
    "scopes": ["upload:write", "links:read", "media:delete", "media:read", "usage:read"],
    "expires_at": "2027-09-23T00:00:00Z",
    "created_at": "2026-09-23T00:00:00Z"
  }
}
```

### 4.2 Token 列表、撤销和轮换

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

撤销是不可逆操作。轮换会创建新 Token 并立即撤销旧 Token；旧 Token 的请求不能因为存在旧的缓存而继续成功。

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

批量接口使用多个 `files[]` 字段。单次最大文件数、总大小和并发数由套餐权益配置，Free 默认建议最多 10 个文件、总大小不超过单次上传额度。

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
      {"client_name": "cover.png", "status": "processing", "upload_id": 1001},
      {"client_name": "avatar.jpg", "status": "ready", "upload_id": 1002, "links": {"url": "...", "markdown": "..."}},
      {"client_name": "bad.svg", "status": "failed", "error": {"code": "UPLOAD_TYPE_NOT_ALLOWED"}}
    ],
    "summary": {"total": 3, "accepted": 2, "failed": 1}
  }
}
```

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
      "thumbnail": "https://img.example.com/i/abc/thumbnail.webp",
      "medium": "https://img.example.com/i/abc/medium.webp",
      "webp": "https://img.example.com/i/abc/image.webp",
      "avif": "https://img.example.com/i/abc/image.avif",
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

`links` 是固定对象。Variant 暂不可用时对应值为 `null`，不能返回猜测出来的 URL。

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

查询自己上传的图片不需要额外配置 Scope；如果 Token 属于同一用户，则可以获取该图片的处理状态和链接。其他用户的媒体统一返回 `MEDIA_ACCESS_DENIED`。

### 5.6 删除自己的图片

```http
DELETE /api/v1/media/1001
Authorization: Bearer fst_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

删除进入回收站并立即使该媒体的分享链接失效。物理对象由异步清理任务删除。该能力属于个人 Token 基础能力，不要求额外配置 `media:delete`。

## 6. 查询自己的媒体

如果 Token 创建时选择了 `media:read`，可以查询当前用户的媒体列表：

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

没有 `media:read` 时，Token 仍然可以读取自己刚刚上传的结果和链接，但不能调用全量媒体列表。

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

## 9. 错误码

| HTTP | Code | 说明 |
| ---: | --- | --- |
| 401 | `TOKEN_INVALID` | Token 不存在、格式错误或已撤销 |
| 401 | `TOKEN_EXPIRED` | Token 已过期 |
| 403 | `USER_SUSPENDED` | 所属用户被禁用 |
| 403 | `MEDIA_ACCESS_DENIED` | 媒体不属于当前用户 |
| 413 | `UPLOAD_SIZE_EXCEEDED` | 超过单文件限制 |
| 415 | `UPLOAD_TYPE_NOT_ALLOWED` | 图片格式不支持 |
| 422 | `UPLOAD_INVALID_IMAGE` | 文件不是可处理图片 |
| 409 | `IDEMPOTENCY_CONFLICT` | 相同幂等键对应不同请求 |
| 429 | `RATE_LIMITED` | Token、用户或 IP 达到速率限制 |
| 409 | `BATCH_PARTIAL_FAILURE` | 批量上传部分成功、部分失败 |
| 429 | `QUOTA_EXCEEDED` | 存储、流量、上传或 API 配额不足 |
| 503 | `UPLOAD_QUEUE_UNAVAILABLE` | 上传后处理队列不可用 |

## 10. 安全和配额

- 默认 Token 只能管理所属用户自己的媒体。
- 每个 Token、用户和 IP 分别限流。
- Token 认证成功时只更新最后使用时间和统计，不记录原文。
- 上传失败不扣最终存储额度；预占额度必须释放或进入可追踪状态。
- 相同幂等键重复请求返回第一次结果，不重复创建媒体。
- Free 用户的上传、链接和删除能力可用，但仍受存储、流量、文件大小和速率限制。
- 服务器端生成链接，客户端不得根据媒体 ID、哈希或文件名拼接 URL。

## 11. 版本兼容

- 所有开发者 API 放在 `/api/v1` 下。
- 新增链接类型只能新增 `links` 字段，不能修改已有 `url`、`markdown`、`html` 和 `bbcode` 的含义。
- 客户端必须忽略未知字段和未知链接键。
- 弃用字段至少保留一个版本周期，并在 OpenAPI 和变更日志中标记。
- 错误码不能复用为其他业务语义。
- 批量上传和单文件上传必须共享相同的文件校验、配额、审核和用量规则。
