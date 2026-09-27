# FastImg 数据契约与状态机

本文档是产品设计和代码实现之间的约束。字段可以增加，但不能在没有迁移、接口兼容说明和测试的情况下改变既有语义。

## 1. ID、时间和金额

- 数据库主键使用项目现有整数 ID 约定；对外公开的媒体、分享、上传会话和 API Key 不直接暴露可枚举内部 ID。
- 对外 Token、对象 Key 和幂等键使用不可预测值。
- 数据库存储 UTC 时间；接口统一返回 ISO 8601 时间。
- 套餐价格使用最小货币单位整数，不使用浮点数。
- 文件大小使用字节整数；前端只负责格式化展示。
- 所有软删除实体使用 `deleted_at` 或项目现有等价约定。

## 2. MediaAsset 字段契约

| 字段 | 类型/约束 | 说明 |
| --- | --- | --- |
| id | 内部主键 | 不直接作为公开 URL |
| user_id | 必填外键 | 所有权 |
| folder_id | 可空外键 | 所属文件夹 |
| original_name | 字符串 | 原始名称，仅用于展示，不参与对象路径 |
| object_key | 字符串 | 原图 StorageObject 引用 |
| content_hash | 字符串 | 内容哈希，用于去重提示和完整性校验 |
| mime_type | 枚举/字符串 | 以服务端检测为准 |
| size_bytes | 非负整数 | 原始对象大小 |
| width/height | 非负整数 | 图片尺寸 |
| visibility | private/link/public | 默认 public；用户可在媒体详情切换为 private 或 link |
| moderation_status | pending/approved/rejected/manual_review | 普通上传默认为 approved；举报或后台治理可在发布后进入复核状态 |
| processing_status | pending/processing/ready/failed | 处理状态 |
| deleted_at | 可空时间 | 回收站时间 |
| created_at/updated_at | UTC 时间 | 审计和排序 |

媒体对外“可访问”不能只判断一个字段，必须同时满足：未删除、处理 ready、分享/公开策略允许，且不是被管理员拒绝。普通上传默认 `public + approved`，可立即通过稳定绝对链接访问并进入发现页；`link` 媒体不参与发现页但持有链接即可访问；`private` 媒体不生成稳定外部链接，只允许媒体所有者登录后访问，管理员预览必须经过 `admin.media.view`。发现页不需要单独投稿字段，采用发布后举报和治理。所有显式创建的临时签名 URL/分享链接仍受过期、撤销和违规状态控制；被拒绝媒体不得通过任何公开投递路径返回。

文件夹归档规则：`folder_id` 可为空；非空时必须引用同一 `user_id` 的文件夹。清空归档只更新媒体所属关系，不改变媒体处理状态、存储用量或链接权限。会员端只能修改自己的 `ready` 媒体，管理员跨用户调整必须走后台授权用例并记录审计。

## 3. MediaVariant 字段契约

Variant 类型首期只保留：

```text
original
```

每个 Variant 必须保存：`media_id`、`variant_type`、`object_key`、`size_bytes`、`width`、`height`、`mime_type`、`processing_status`、`error_code`、`created_at`。

同一 `media_id + variant_type` 只能有一个当前有效版本。重处理使用替换或版本化策略，不能产生无法清理的孤儿对象。

## 4. UsageLedger 契约

```text
user_id
resource_type: storage | bandwidth | upload | api | transform
delta: signed integer
source_type: upload | delete | restore | download | api_request | adjustment
source_id
idempotency_key
period_key
created_at
```

规则：

- 正数表示增加占用或消耗，负数表示释放或冲销。
- `upload`、`restore`、`download`、`api_request` 来源只能写正数；`delete` 来源只能写负数；人工 `adjustment` 可正可负但不能为零。
- 同一个 `idempotency_key` 只能成功写入一次。
- `idempotency_key` 由 `user_id + resource_type + source_type + source_id + period_key` 规范化后生成 SHA-256；相同 key 的重复事件且 delta 相同视为已完成，delta 不同则报冲突，不可静默覆盖。
- 数据库流水写入必须和对应业务状态转换处于同一事务；上传/媒体生命周期先锁定用户行，再写流水以串行化同一用户的余额变更。唯一索引作为最终防重约束。
- 存储使用 `resource_type=storage`、`period_key=lifetime`；软删除进入回收站不释放额度，只有对象物理删除成功后才写负数流水。历史 `storage_bytes` 仅作为读取兼容名。
- 管理员赠送额度必须记录操作者、原因和过期时间。
- 当前余额可以缓存，但最终账务依据是流水和周期聚合。

会员只能通过 `GET /api/v1/me/usage/ledger?page=1&per_page=20` 分页读取自己的用量流水；用户归属从认证身份取得，不接受请求传入 `user_id`。每页默认 20、最大 100，按流水 ID 倒序。响应只含 `data` 与 `meta`，不暴露幂等键。
- 删除失败不能提前写入释放空间的流水。
- 月度 `upload` 与 `transform` 以 `UploadSession.created_at` 所属 UTC 月作为预占和最终入账周期；成功完成、处理中预占和统计摘要必须使用同一周期。旧会话缺失创建时间时，完成入账回退到完成时间所属 UTC 月。

## 5. Personal API Token 与分享 Token

```text
原始值 -> 随机生成 -> 仅返回一次 -> 保存 hash -> 后续只比较 hash
```

- 数据库不保存完整 Personal API Token 或分享 Token。
- 日志、审计、错误响应和统计标签不得包含原始值。
- Key/Token 撤销通过状态和撤销时间实现，不能依赖删除记录后猜测失效。
- 轮换是创建新 Key 后撤销旧 Key，不能原地修改 hash 造成并发请求歧义。

FastImg 对外产品名称统一使用 `Personal API Token`，数据库表和内部类型可以使用 `api_tokens`；旧文档中的 `ApiKey` 只作为内部兼容命名，不得在用户界面同时出现两种术语。

### ApiToken 字段契约

| 字段 | 说明 |
| --- | --- |
| id | 内部 ID，不作为 Token |
| user_id | Token 所属用户 |
| name | 用户可识别名称 |
| token_hash | 完整 Token 的不可逆摘要 |
| prefix | 仅保存非敏感前缀，便于用户识别 |
| scopes | 允许的操作范围 |
| status | active/disabled/revoked |
| expires_at | 可空过期时间 |
| last_used_at | 最近成功认证时间 |
| last_used_ip | 最近成功来源 IP，按隐私策略保留 |
| usage_count | 可选聚合值，不能作为计费用量来源 |
| created_at/revoked_at | 生命周期时间 |

Token 原文只存在于创建响应和用户当前页面内存中；服务端认证只使用 `token_hash`。认证失败日志必须只记录 prefix、Token ID 和 request ID。

### Scope 规则

当前 Personal API Token 采用最小固定开放面，不在 C 端创建表单中让用户自行勾选 Scope，也不把链接读取拆成独立的 `links:read` 权限：

| Scope | 当前用途 | 数据范围 |
| --- | --- | --- |
| `upload:write` | 单文件上传 | 只能创建当前 Token 所属用户的媒体 |
| `media:read` | 查询本人媒体列表、详情、内容和上传完成后的各种链接 | 只能读取当前 Token 所属用户的媒体 |
| `media:delete` | 删除本人媒体到回收站 | 只能删除当前 Token 所属用户的媒体 |

当前不向 Personal API Token 开放 `links:read`、`usage:read`、`webhook:manage`、批量上传、上传重试、套餐/用量、文件夹、相册、分享链接、防盗链和任何管理员接口。链接属于本人媒体详情/上传结果的一部分，由 `media:read` 统一保护；删除能力由 `media:delete` 单独保护。上传完成后的 `links` 使用 `APP_URL` 生成带 APP_KEY 签名的绝对公开地址，只返回唯一的 `original`、`url`、`markdown`、`html`、`bbcode`；不暴露对象存储地址。接收站自行使用 CSS 或自身处理链控制显示尺寸，不由 FastImg 为每张图片额外保存缩略图和中图。

公开相册使用 `GET /api/v1/public/albums/{id}`，不接受会员 Token 或管理员身份作为绕过条件。只有相册 `visibility=public` 且媒体同时满足 `ready`、未删除、`visibility=public`、`moderation_status != rejected` 时才出现在响应；公开图片内容地址绑定相册 ID 与媒体 ID，服务端每次读取都会重新校验关系和公开状态。

公开媒体、分享链接、发现页和会员自己的媒体读取都会在返回内容前按实际响应字节写入 `bandwidth/download` 流水。当前 UTC 月累计值通过 `/api/v1/me/usage` 的 `usage.bandwidth` 返回，响应同时提供 `bandwidth_metered: true`；超出套餐非零 `monthly_bandwidth_bytes` 时拒绝本次响应，不允许把套餐流量限制当作展示字段。

会员网页登录会话与 Personal API Token 是两种不同的认证边界。网页登录会话可以访问会员端的订单、套餐用量、文件夹、相册、分享链接和防盗链页面；Personal API Token 只用于脚本、PicGo、ShareX、CI 等自动化上传和本人媒体闭环。

管理员后台使用独立的登录会话和 `admin.*` RBAC 权限。管理员不会通过 Personal API Token 调用 `/api/v1/admin/**`；后台的 `api_tokens` 资源只允许授权管理员查看 Token 的非敏感运营元数据，并按独立的更新/删除权限执行停用、撤销或删除，不能创建、读取明文或导出 Token。

## 6. 状态机

### UploadSession

```text
created
  -> uploading
  -> uploaded
  -> processing
  -> completed

created/uploading -> cancelled
created/uploading -> expired
uploaded/processing -> failed
```

非法迁移必须返回 `UPLOAD_INVALID_STATE`。过期会话不能重新完成，客户端必须创建新会话。

### MediaAsset

```text
pending -> processing -> ready
pending -> rejected
processing -> failed
failed -> processing
ready -> deleted
deleted -> ready
deleted -> cleanup_pending -> physically_deleted
ready -> expired
```

`deleted -> ready` 只允许回收站恢复；如果原对象已被清理，恢复必须失败并说明原因。永久删除先原子迁移到 `cleanup_pending`，阻止并发恢复；存储 Provider 删除所有对象成功后，数据库事务将对象/Variant 标为 deleted、媒体标为 `physically_deleted` 并写入负向存储流水。失败保持 `cleanup_pending`，允许以相同请求安全重试；Provider 的 `Delete` 必须幂等。共享存储对象在引用计数/共享删除契约实现前拒绝永久删除。

### Subscription

套餐及权益快照约定：

- `price_amount` 是非负最小货币单位整数；首期 `billing_period` 只允许 `monthly`、`yearly`，状态只允许 `active`、`disabled`。
- `entitlements_json` 使用完整 JSON 对象，字段固定为 `storage_bytes`、`max_file_bytes`、`daily_uploads`、`monthly_api_uploads`、`monthly_bandwidth_bytes`、`transform_count`、`api_rate_per_minute`、`token_limit`、`ads_enabled`、`watermark_enabled`；字节和次数均为整数。
- `transform_count` 按 UTC 月统计成功完成的媒体处理作业；单张媒体生成原图、缩略图和中图合计为一次。处理中上传会话预占一次，成功后以幂等流水确认，失败会话释放预占且不计成功用量。
- `storage_bytes` 必须大于零；其他数值上限不得小于零，零表示该项不设上限；`ads_enabled` 和 `watermark_enabled` 必须显式为布尔值。
- 新写入必须包含全部字段且不能包含未知字段，JSON/API 输出统一使用 snake_case。读取存量快照时兼容旧 Go 默认编码的 PascalCase 字段名；旧快照缺少 `watermark_enabled` 时按关闭处理，不能把旧快照改写为另一组权限语义。
- 已存在的版本化订阅快照如果是在 `watermark_enabled` 加入前生成，继续接受旧的内容哈希并按关闭处理；新计划写入和新订阅统一使用包含水印字段的新哈希，避免修改套餐后会员订阅页出现 `SUBSCRIPTION_UNAVAILABLE`。

站点水印设置使用动态系统设置键：`watermark.text` 为上传时写入原图、缩略图和中图的文字，`watermark.domain` 为可选追加域名，`watermark.fallback_image_url` 为会员端图片加载失败时的备用图片地址。后端公开 `GET /api/v1/site/presentation` 只返回经过 URL 安全校验的兜底图地址，不返回其他设置或凭证；空值使用内置 FastImg 兜底图。
- 系统设置写入的 `value_type` 支持 `string`、`secret`、`boolean`、`integer` 和 `json`。`secret` 只接受提交值用于加密更新，列表接口统一返回占位符；可选整数允许空字符串，具体业务读取时必须使用安全默认值。管理员设置表单不得因为未配置的支付密钥或可选统计参数阻断其他设置保存。
- 创建订阅时在现有 `entitlement_snapshot_json` 字段保存完整订阅快照：`schema_version`、`plan_version`、套餐 ID/编码/名称/描述/价格/货币/周期和完整权益。配额读取兼容从快照 envelope 读取权益；旧的纯权益 JSON 继续有效。
- `plan_version` 是对规范化套餐条款与权益计算的 `sha256:<hex>` 内容版本：相同条款得到相同版本，任一条款/权益变化得到新版本；读取时重新计算并校验，快照内容不能静默偏离版本。该版本标识内容，不承诺按时间单调递增；订阅周期/创建时间负责时间顺序。
- 新订阅和后续订单必须使用创建时快照，套餐后续编辑不得回写订阅或订单快照。既有纯权益快照没有保存套餐名称/价格等信息，不能推断历史条款；API 对这类记录回退显示当前套餐并明确 `snapshot_available: false`。不在未授权时对本地数据库做回填。

```text
pending -> active -> scheduled_downgrade -> expired
active -> cancelled
active -> refunded
```

套餐升级可以创建新的权益快照；套餐降级在当前周期结束时切换。已生效快照不能随意被后台编辑覆盖。

### Order / PaymentIntent / PaymentTransaction / PaymentEvent

详细接口与状态机见 `docs/fastimg-payment-gateway.md`。数据层必须保证：

- `public_order_no`、`gateway_code + event_id`、PaymentIntent 幂等键和渠道外部流水号唯一。
- 订单商品、价格、货币、套餐版本和权益快照在支付后不可覆盖。
- PaymentTransaction 只追加；退款、拒付、冲正和人工调整使用新的反向流水。
- 支付成功与订阅履约分开记录，履约失败不能伪造为未支付，也不能丢失补偿任务。
- 当前套餐、配额和广告权益必须来源于 Subscription/EntitlementSnapshot，不得直接读取支付渠道响应。

### Price / Discount / Invoice / Reconciliation

- `Price` 保存产品、套餐版本、周期、货币、最小货币单位金额和启停时间。
- `Discount` 保存优惠码、适用产品、有效期、总次数、单用户次数和叠加规则。
- `Invoice` 保存订单价格、折扣、税费、客户快照和账单状态；历史账单不可被当前价格覆盖。
- `ReconciliationRecord` 保存渠道、结算批次、外部流水、平台流水、匹配状态、差异金额、证据和处理人。
- 对账状态至少为 `matched`、`missing_internal`、`missing_gateway`、`amount_mismatch`、`currency_mismatch`、`duplicate`、`pending_review`。
- 加密支付必须额外保存资产、网络、合约地址、收款地址、报价金额、汇率版本、报价过期时间、交易哈希、区块高度、当前确认数、要求确认数、链状态和风控状态。
- 交易哈希、网络、收款地址和资产组合必须可去重；链重组不得覆盖历史链上观察记录。

### ModerationTask

```text
ordinary upload: approved; post-publication report/review: pending | manual_review -> approved
pending -> processing -> rejected
processing -> manual_review
manual_review -> approved | rejected
```

审核服务超时、额度不足或网络错误进入 `manual_review`/重试，不进入 approved。

## 7. API 字段兼容规则

- 新增响应字段默认可选，旧客户端必须仍能解析。
- 删除字段前必须经过弃用周期并更新 OpenAPI。
- 枚举新增值时，前端必须提供未知值降级展示。
- 错误码一旦发布不能复用给其他语义。
- 分页字段固定使用 `data` 和 `meta`，不要为单个模块创建不同格式。
- 二进制上传和 JSON 元数据分开处理，不能把大文件读入内存后再写数据库。
- 图片链接响应使用固定 `links` 对象；Variant 未生成时返回 `null`，不得返回猜测 URL。
- `201` 表示媒体已 ready 并返回链接，`202` 表示仍在 processing 并返回 status URL。
- 链接键新增时保持旧键不变；客户端必须忽略未知链接键。

## 8. 治理与处罚契约

### DiscoverySettings

```text
enabled: boolean
moderation_mode: manual | automatic | hybrid
default_sort: latest | random | popular
report_threshold: integer
disabled_message: string
updated_by: user_id
updated_at: UTC time
```

### Report

```text
id
media_id
reported_user_id
reporter_id nullable
reason: illegal | copyright | spam | harassment | privacy | other
description
evidence_url nullable
status: pending | reviewing | resolved | dismissed
resolution
dedupe_key
resolved_by
resolved_at
created_at
```

### ModerationAction

```text
id
target_type: media | user | report
target_id
action: hide_media | delete_media | restore_media | restrict_upload |
        restrict_discovery | suspend_user | ban_user | dismiss_report
reason
evidence
expires_at nullable
operator_id
created_at
```

治理动作必须追加记录，不能覆盖原处理历史。恢复图片或解封账户是新的动作记录，不是删除处罚记录。

### ModerationAppeal

```text
id
user_id
target_type: media | user | moderation_action
target_id
reason
description
evidence
status: pending | reviewing | upheld | overturned | adjusted
reviewer_id nullable
review_note nullable
resolved_at nullable
created_at
```

约束：

- 同一用户、同一目标、同一处罚只能存在一个未完成申诉。
- 举报去重使用 `reporter_id + media_id + reason + active_period`，重复提交返回原举报 ID。
- 举报阈值触发复核任务，不直接生成不可逆封禁动作。
- 处罚动作和申诉结果都必须追加审计记录。

## 9. 防盗链契约

### HotlinkPolicy

```text
mode: off | referer | signed | hybrid
allow_no_referer: boolean
hotlink_domains: normalized host[]
signed_url_ttl_seconds: 60..86400
updated_at: UTC time
```

策略评估顺序固定为：媒体状态、可见性/分享权限、签名 URL、Referer 白名单、用户/Token/IP/域名/流量限制、对象读取。CDN 命中时也必须执行等价策略判断，策略变更要触发缓存失效或版本隔离。

### SignedURLPayload

```text
media_id
url: public delivery URL, never object-storage origin URL
expires_at: required finite time
key_id: active signing key version
scope: view
```

当前实现使用应用层 HMAC-SHA256，签名绑定媒体 ID、Variant 和过期时间；公开投递路径是 `/i/{id}`，永不返回对象存储源地址。过期、篡改、格式错误或与当前策略不兼容时统一返回 `404 LINK_NOT_FOUND`，不得暴露具体校验步骤。`media_access_logs` 只记录投递决策和最小请求元数据，不记录签名、密码或完整 Token；记录写入失败不阻断图片响应。

管理员访问查询使用独立的 all-scope RBAC 接口 `GET /api/v1/admin/media-access-logs` 和页面 `/admin/media-access-logs`。允许按媒体、Variant、投递模式、结果和 Referer 主机筛选；接口只投影脱敏访问字段，不把访问日志表当作会员 own-scope 媒体接口。

## 10. 事务和一致性边界

### 同一数据库事务

- 创建业务记录和对应的用量预占记录
- 支付事件落库和订单状态迁移
- 审核结果和媒体公开状态变更
- 分享链接创建和策略记录

### 异步最终一致

- 数据库记录与对象存储对象
- 媒体记录与派生 Variant
- 访问事件与统计聚合
- Webhook 发送与第三方接收

异步边界必须有状态字段、重试任务、失败原因和人工恢复入口。
