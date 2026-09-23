# FastImg 阶段性开发计划与详细功能清单

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans or superpowers:subagent-driven-development to implement this plan task-by-task. 每个阶段完成后必须独立测试、审阅迁移、更新文档并形成阶段性提交。

**Goal:** 基于现有 `go-vue-admin` 能力，分阶段交付一个免费可用、付费可扩展、支付渠道可替换、具备生产运营边界的媒体托管平台 FastImg。

**Architecture:** 复用 Goravel、Resource Engine、RBAC、审计、Redis、队列、文件系统、OpenAPI 和 shadcn-vue。标准后台 CRUD 使用生成器；上传、媒体、支付、审核、统计和复杂工作台使用业务模块与 Service，不重复建设底层框架能力。

**Spec:** `fastimg-product-design.md`、`fastimg-requirements-matrix.md`、`fastimg-data-contracts.md`、`fastimg-production-readiness.md`

## 一、总体开发策略

### 1.1 交付原则

- 先完成 Free 用户完整闭环，再开发付费和真实支付。
- 每个阶段都必须产生可运行、可测试、可回滚的垂直切片。
- 业务模块只能通过 Service 使用配额、媒体、存储、支付、审核和统计能力。
- 不把真实支付、加密货币、CDN 或自定义域名作为 Free MVP 的前置条件。
- Fake/offline Provider 可以用于开发和验收，但不能标记为生产集成完成。
- 所有涉及媒体、订单、用量和支付的状态变更必须支持幂等、审计和失败恢复。

### 1.2 模块边界

```text
backend/app/modules/plans
backend/app/modules/uploads
backend/app/modules/media
backend/app/modules/shares
backend/app/modules/developer
backend/app/modules/moderation
backend/app/modules/billing
backend/app/modules/advertising
backend/app/modules/analytics

backend/app/services/quota
backend/app/services/storage
backend/app/services/media
backend/app/services/shares
backend/app/services/billing
backend/app/services/moderation
backend/app/services/analytics
```

标准套餐、广告位、订单、支付流水等列表优先使用 Resource Engine；上传工作台、媒体瀑布流、支付工作台、审核队列和统计图表使用 Custom Page。

## 二、阶段总览

| 阶段 | 名称 | 目标 | 生产状态 |
|---|---|---|---|
| M0 | 基线与架构准备 | 确认仓库、服务、模块和开发约束 | 不发布 |
| M1 | 账户、套餐与配额 | Free 套餐和用量边界可用 | 可进入开发环境 |
| M2 | 上传、存储与媒体库 | 用户可以上传和管理图片 | 可做内部测试 |
| M3 | 稳定链接、分享与防盗链 | 图片可安全访问和分享 | 可做灰度测试 |
| M4 | Personal API Token | 开发者可以通过 API 使用图床 | 可做灰度测试 |
| M5 | 审核、举报与运营后台 | 建立公开内容治理闭环 | 可做小范围测试 |
| M6 | Fake 商业化 | 订单、订阅、权益和广告可验收 | 不接真实收款 |
| M7 | 真实支付与加密支付 | 接入支付网关和 Crypto Provider | Provider 验证后启用 |
| M8 | 生产发布与规模化 | 备份、监控、CDN、域名和恢复演练 | 通过门禁后发布 |

## 三、M0：基线与架构准备

### 功能范围

- 确认 Goravel、Vue、Resource Engine、RBAC、审计、OpenAPI、队列和文件系统扩展点。
- 确认 PostgreSQL、Redis、对象存储配置和本地启动方式。
- 建立 FastImg 模块、路由、权限、迁移和前端页面命名规范。
- 创建错误码、业务 ID、幂等键和状态机约定。
- 建立 OpenAPI、TypeScript Client、测试、日志和审计基线。

### 主要产物

- 模块映射、环境变量清单、迁移规范和发布手册。
- `plans`、`permissions`、`audit`、`usage` 等基础资源的设计。
- 基线测试报告：后端测试、前端类型检查和构建。

### 验收

- PostgreSQL 和 Redis 连接可验证。
- 不修改框架核心文件即可注册业务模块。
- 新业务依赖方向符合 `Business Module -> Admin Core -> Goravel`。
- 迁移、队列、OpenAPI 和前端生成入口已确认。

## 四、M1：账户、套餐与配额

### 功能计划

1. 新用户自动获得 Free 套餐。
2. 管理员维护套餐名称、价格、周期、排序、启停和权益版本。
3. 套餐权益包含存储、单文件大小、每日上传、月 API、月流量、处理次数、速率、广告、防盗链、Webhook 和自定义域名。
4. 建立 Subscription 和 EntitlementSnapshot，套餐修改不影响历史订阅。
5. 建立 UsageLedger，记录上传、删除、恢复、下载、API、处理和管理员调整。
6. 实现存储预占、确认、释放和并发额度校验。
7. 套餐升级即时生效，降级周期末生效；降级不删除已有媒体。
8. 超额时只阻止新增资源或高成本操作，不影响用户查看已有资源的策略性访问。

### 后端模块

- `modules/plans`
- `services/quota`
- `services/billing` 中的订阅读取部分

### API

```text
GET /api/v1/plans
GET /api/v1/subscription
GET /api/v1/me/usage
GET /api/v1/me/usage/ledger
GET /api/v1/quota
```

### 验收

- Free 用户可查询权益和用量。
- 并发请求不能突破存储和上传限制。
- 删除失败不能提前释放配额。
- 用量重复事件不会重复计量。
- 管理员调整额度必须有原因和审计记录。

## 五、M2：上传、存储与媒体库

### 功能计划

1. 创建 UploadSession，支持单文件上传、批量上传和后续分片扩展。
2. 校验真实文件格式、扩展名、MIME、大小、像素、动画帧数和用户配额。
3. 建立 `StorageProvider`，首期实现 Local Provider 和 Fake Provider，预留 S3-compatible Provider。
4. 建立 MediaAsset、MediaVariant、StorageObject。
5. 图片异步生成 thumbnail、medium、WebP、AVIF。
6. 默认清理 GPS、设备信息等 EXIF。
7. 用户媒体库支持网格、列表、筛选、搜索、文件夹、相册、批量操作和详情页。
8. 删除进入回收站，支持恢复和永久删除。
9. 处理失败、对象存在但数据库失败、数据库存在但对象缺失都必须可重试或人工修复。

### API

```text
POST   /api/v1/uploads
POST   /api/v1/uploads/batch
GET    /api/v1/uploads/{id}
GET    /api/v1/media
GET    /api/v1/media/{id}
PATCH  /api/v1/media/{id}
DELETE /api/v1/media/{id}
POST   /api/v1/media/{id}/restore
```

### 验收

- 用户上传后得到稳定媒体 ID。
- `201` 表示 ready；`202` 表示 processing，并返回状态查询地址。
- 用户只能访问自己的媒体。
- Variant 未生成时链接返回 `null`，不能猜测 URL。
- 删除、恢复、处理失败和重试不会重复扣减或释放配额。

## 六、M3：稳定链接、分享与防盗链

### 功能计划

1. 生成原图、缩略图、中图、WebP、AVIF、纯 URL、Markdown、HTML、BBCode。
2. 支持 private、link、public 三种可见性。
3. 支持密码分享、过期分享、撤销分享和访问记录。
4. 公开发现页只允许 ready、approved 且公开的媒体。
5. 实现关闭、Referer 白名单、Signed URL 和混合防盗链模式。
6. 支持防盗链域名归一化、无 Referer 策略、CDN 缓存失效和签名过期。
7. 对象存储保持私有，稳定链接由应用或 CDN 策略控制。

### API

```text
POST   /api/v1/media/{id}/share-links
GET    /api/v1/share-links
DELETE /api/v1/share-links/{id}
GET    /s/{token}
GET/PUT /api/v1/media/{id}/hotlink-policy
GET    /api/v1/hotlink-domains
POST   /api/v1/hotlink-domains
DELETE /api/v1/hotlink-domains/{id}
POST   /api/v1/media/{id}/signed-url
```

### 验收

- 分享密码和过期时间由服务端校验。
- 被删除、隐藏、封禁或过期的媒体不能继续公开访问。
- 防盗链拒绝不泄露对象源地址。
- CDN 命中不能绕过媒体状态、分享和防盗链策略。

## 七、M4：Personal API Token 与开发者 API

### 功能计划

1. Token 只保存 hash，完整 Token 创建时只显示一次。
2. 支持创建、列表、撤销、禁用、轮换、过期和最后使用信息。
3. 基础能力固定为：`upload:write`、`links:read`、`media:delete`。
4. 可选能力为：`media:read`、`usage:read`、`webhook:manage`。
5. API 上传复用网页上传、媒体、配额、审核和用量 Service。
6. 支持单文件、批量、分片和断点续传。
7. 支持 `Authorization: Bearer`、`Idempotency-Key`、API 限流和配额响应头。
8. API 错误统一返回错误码、request ID 和 retryable 标识。
9. 增加 curl、JavaScript、Python、PicGo 和 ShareX 示例。

### 验收

- Token 可以上传、获取自己图片链接和删除自己图片。
- Token 不能操作其他用户媒体。
- Token 撤销后立即失效。
- 重复幂等键不会重复创建媒体或扣配额。
- 批量部分失败不会影响其他文件结果。

## 八、M5：审核、举报与运营后台

### 功能计划

1. 发现页后台支持启用/停用和单独关闭投稿。
2. 图片、用户、域名支持举报。
3. 举报去重、阈值复核、人工审核和申诉。
4. 管理员可以隐藏、删除、恢复违规图片。
5. 管理员可以限制投稿、限制上传、暂停或封禁账户。
6. 所有处罚、恢复、额度调整和永久删除都写入审计。
7. 图片处理增加 SVG 清洗、恶意文件、超大像素、压缩炸弹和 SSRF 防护。
8. 公开内容必须通过审核状态机后才能进入发现页。

### 管理端页面

- 会员管理
- 用户媒体库
- 举报队列
- 审核工作台
- 违规账户处理
- 申诉处理
- 发现页配置
- 任务失败与人工重试

### 验收

- 发现页停用不影响私有媒体。
- 自动阈值不会直接造成不可逆永久封禁。
- 审核服务失败时进入人工复核，不自动公开。
- 管理操作具备操作者、原因、证据和时间。

## 九、M6：Fake 商业化与订阅履约

### 功能计划

1. 建立 Order、OrderItem、PaymentIntent、PaymentTransaction、PaymentEvent。
2. 实现统一 `PaymentGateway` 接口和 `fake/manual Provider`。
3. 支持套餐价格、周期、货币、优惠和权益快照。
4. 订单状态：created、pending_payment、paid、fulfilled、canceled、refunded。
5. 支付成功和权益履约分开处理。
6. 支持升级、续费、降级、取消、部分退款、全额退款和履约补偿。
7. 支付回调按订单号、渠道流水号和事件 ID 幂等。
8. 支付成功后刷新配额、API 速率、广告免除、防盗链、自定义域名和 Webhook 权益。
9. 广告位、素材、周期、套餐定向、展示和点击统计可配置。
10. 管理端支持订单、支付流水、履约任务、退款和人工调整。

### 验收

- Free 服务不依赖支付网关。
- 重复支付回调不会重复激活订阅或增加权益。
- 支付成功但履约失败时可以人工重试。
- 退款和人工调整只追加流水，不覆盖原记录。
- fake 支付完成前端订单和后台订单闭环。

## 十、M7：真实法币支付与加密货币支付

### 法币支付

- 按 Provider 接入真实支付渠道，不修改订单和订阅领域模型。
- 支持渠道启停、货币、地区、金额范围、优先级、熔断和密钥版本。
- 验证真实回调签名、金额、货币、订单状态和重复事件。
- 支持退款查询、异步退款、支付失败、渠道超时和渠道不可用。
- 建立支付流水对账和差异复核。

### 加密货币支付

- 支持资产、网络、合约地址、精度、报价和过期时间。
- 收款地址、Memo/Tag、交易哈希、区块高度和确认数可追踪。
- 只有达到要求确认数且通过风险校验才激活权益。
- 处理少付、多付、错误网络、过期报价、重复交易、链重组和双花风险。
- 不保存用户私钥或助记词，首期使用托管或可审计 Provider。
- 高风险地址、制裁名单、混币风险和 Provider 风控失败进入人工复核。

### 生产门禁

真实 Provider 必须完成：真实商户配置、签名验证、回调测试、重复事件、金额差异、退款、对账和故障演练。加密 Provider 还必须完成确认数、链重组、风控和错误网络测试。

## 十一、M8：生产发布与规模化

### 功能计划

1. S3-compatible 对象存储、私有桶、CDN 和回源策略。
2. 自定义域名 DNS 验证、证书签发、续期、暂停和删除。
3. PostgreSQL、对象存储、配置和密钥备份。
4. Redis 丢失恢复、队列重建和失败任务补偿。
5. 上传、处理、外链、支付、举报、队列和存储告警。
6. API、对象存储、图片处理、CDN 和支付压测。
7. 代码回滚、兼容迁移、对象恢复和灾备演练。
8. 账户注销、数据导出、媒体永久删除和保留期执行。
9. 发布报告区分 `designed`、`implemented`、`integrated`、`verified` 和 `enabled`。

### 发布条件

- Free MVP 所有 P0 Gate 通过。
- 生产数据库和对象存储恢复演练通过。
- 真实 Provider 只在对应 Provider 验证完成后启用。
- 监控、告警、审计和人工恢复入口可用。
- 已知限制、回滚方式和负责人记录在发布报告中。

## 十二、每阶段统一开发流程

每个阶段按以下顺序执行：

1. 更新需求矩阵、数据契约和接口目录。
2. 创建或审阅迁移，确认索引、外键、唯一约束、软删除和回滚边界。
3. 先写 Service/Repository/API 契约测试，再实现业务逻辑。
4. 使用 Resource Generator 生成标准 CRUD，复杂流程使用 Custom Page。
5. 增加权限、数据范围、审计、错误码和幂等处理。
6. 增加异步任务的重试、失败记录、补偿和人工恢复入口。
7. 更新 OpenAPI 和 TypeScript Client。
8. 运行后端测试、前端类型检查、构建和真实 PostgreSQL/Redis 集成检查。
9. 更新发布手册、验收记录和已知限制。
10. 创建本地阶段性提交；验证前不推送远程仓库。

## 十三、统一 Definition of Done

- 功能已落在正确模块，没有把业务逻辑写进 Admin Core。
- API、权限、数据范围、错误码、状态机和数据契约已更新。
- 正常、越权、重复请求、额度边界、依赖失败和恢复路径都有测试。
- 迁移已人工审阅，PostgreSQL 上验证通过。
- 异步任务具备幂等键、重试、退避、失败原因和人工恢复入口。
- OpenAPI、前端 Client、用户文档和管理端页面保持一致。
- 后端测试、前端类型检查和构建通过。
- 文档明确标注真实 Provider、fake Provider 和未启用能力。
- 阶段验收结果、回滚方式和已知问题已经记录。

