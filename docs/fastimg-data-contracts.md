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
| visibility | private/link/public | 默认 private |
| moderation_status | pending/approved/rejected/manual_review | 审核状态 |
| processing_status | pending/processing/ready/failed | 处理状态 |
| deleted_at | 可空时间 | 回收站时间 |
| created_at/updated_at | UTC 时间 | 审计和排序 |

媒体对外“可访问”不能只判断一个字段，必须同时满足：未删除、处理 ready、分享/公开策略允许、审核状态允许。

## 3. MediaVariant 字段契约

Variant 类型首期固定为：

```text
original
thumbnail
medium
webp
avif
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
- 同一个 `idempotency_key` 只能成功写入一次。
- 管理员赠送额度必须记录操作者、原因和过期时间。
- 当前余额可以缓存，但最终账务依据是流水和周期聚合。
- 删除失败不能提前写入释放空间的流水。

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

### Scope 兼容规则

首期固定基础能力为 `upload:write`、`links:read`、`media:delete`，三者默认授予且不能被用户关闭，只能作用于 Token 所属用户自己的媒体。可选 Scope 为 `media:read`、`usage:read`、`webhook:manage`。新增可选 Scope 必须默认不授予已有 Token，并在 Token 详情中明确展示。

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
ready -> expired
```

`deleted -> ready` 只允许回收站恢复；如果原对象已被清理，恢复必须失败并说明原因。

### Subscription

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
pending -> processing -> approved
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
submissions_enabled: boolean
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
mode: off | referer_allowlist | signed_url | hybrid
allow_missing_referer: boolean
domain_ids: string[]
signed_url_ttl_seconds: bounded integer
placeholder_mode: deny | placeholder
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

签名至少绑定媒体 ID、规范化投递路径、过期时间和 scope。过期、撤销、格式错误或与当前策略不兼容时统一返回 `HOTLINK_BLOCKED`，不得暴露具体校验步骤。

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
