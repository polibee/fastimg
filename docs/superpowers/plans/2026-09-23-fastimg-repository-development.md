# FastImg Repository Development Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在真实的 `go-vue-admin` 仓库中实现一个免费用户可完整使用、付费用户可扩容增强的 FastImg 媒体托管 MVP。

**Architecture:** 复用现有 Goravel、Resource Engine、RBAC、审计、Redis、队列、文件系统、OpenAPI 和 shadcn-vue。标准资源通过 `admin:make-resource` 生成；上传、媒体库、分享、支付、审核、统计使用 `backend/app/modules/<module>` 的业务边界和 `backend/app/services/<domain>` 的领域服务。

**Tech Stack:** Go/Goravel v1.18、GORM/ORM、PostgreSQL、Redis、Vue 3、TypeScript、shadcn-vue、Vite、OpenAPI、Redis queue；首期 Local + S3-compatible 存储 Provider，支付/审核/CDN 使用 fake 或可替换接口。

**Spec:** `docs/fastimg-product-design.md`、`docs/architecture.md`、`docs/module-layout.md`、`docs/resource-engine.md`

**Localization:** `docs/i18n.md`（所有功能共同遵守的中英文实现与验收约束）

**Requirements:** `docs/fastimg-requirements-matrix.md`

**Developer API:** `docs/fastimg-developer-api.md`

**Frontend Design:** `docs/fastimg-frontend-design.md`

**Moderation Policy:** `docs/fastimg-moderation-policy.md`

**Compliance Policy:** `docs/fastimg-compliance-policy.md`

**Storage Cost Policy:** `docs/fastimg-storage-cost-policy.md`

**Hotlink Protection:** `docs/fastimg-hotlink-protection.md`

**Payment Gateway:** `docs/fastimg-payment-gateway.md`

**Production Readiness:** `docs/fastimg-production-readiness.md`

**Stage Development Plan:** `docs/fastimg-stage-development-plan.md`

## Global Constraints

- 业务依赖方向只能是 `Business Module -> Admin Core -> Goravel`。
- 新业务资源放在 `backend/app/modules/<name>`，跨资源业务 Service 放在 `backend/app/services/<domain>`。
- 迁移统一放在 `backend/database/migrations`，必须审阅后手动执行。
- 普通后台资源使用 `admin:make-resource`；复杂流程才使用 Custom Page。
- 不重复实现用户、RBAC、通用 CRUD、审计、配置、队列、文件系统和基础 UI。
- 所有用户资源执行 `user_id` 归属校验；所有额度变化写入用量流水。
- Redis 与 PostgreSQL 使用项目现有配置，禁止静默内存降级。
- 所有用户可见文案在同一功能变更内同步交付 `zh-CN`/`en-US` JSON；前端仅通过 vue-i18n namespace key 读取，后端通过 Goravel Localization 本地化，禁止内联双语文案和重复建设 i18n。
- 每个前端功能必须纳入语言文件路径/嵌套 Key 一致性测试，并覆盖对应页面使用语言 Key；统一运行 `cd admin && node --test tests/fastimg-i18n.test.mjs`。
- 每阶段完成后运行后端测试、前端类型检查和构建，并形成独立本地提交。

## Review Focus

- `admin:make-resource` 生成的资源是否只写入生成专属 discovery 文件，未手改通用核心页面。
- `routes/web.go` 中的业务路由是否全部经过现有认证/权限中间件。
- 迁移是否只使用项目当前 PostgreSQL 约定，未引入未验证的数据库特性。
- 队列和文件系统失败时，媒体状态、配额预占和对象清理是否可恢复。
- 任何 ID、批量操作、导出和分享 URL 是否都进行资源归属检查。

## Dependency Graph

```text
Task 0 基线/映射
  -> Task 1 标准资源和迁移
  -> Task 2 套餐/配额/用量
  -> Task 3 存储/上传
  -> Task 4 媒体库/派生图
  -> Task 5 分享链接
  -> Task 6 Personal API Token
  -> Task 7 清理/失败恢复
  -> Task 8 审核/发现页
  -> Task 9 订单/广告/统计
  -> Task 10 发布验收
```

Task 2 是上传和支付的共同前置；Task 3 是媒体库的前置；Task 4 是分享、审核和统计的前置。Task 9 中订单可以先使用 fake Provider，不应阻塞 Free 闭环。

## Milestones

| 里程碑 | 包含任务 | 可验收结果 |
| --- | --- | --- |
| M0 基线 | 0 | 现有项目测试、目录映射和服务约束明确 |
| M1 账户可用 | 1-2 | Free 套餐、权益、配额和用量流水可以读取和扣减 |
| M2 上传可用 | 3-4 | 用户可以上传、处理、浏览、删除和恢复自己的媒体 |
| M3 分享可用 | 5-6 | 用户可以生成稳定链接和使用 Personal API Token 上传 |
| M4 可运营 | 7-8 | 清理、重试、审核、举报和发现页具备安全闭环 |
| M5 可商业化 | 9 | fake 订单、广告和基础统计可以独立验收 |
| M6 可发布 | 10 | 迁移、队列、Redis、API、前端构建和验收报告完整 |

每个里程碑必须有：代码、测试结果、迁移审阅结果、已知限制、回滚说明和本地提交。没有真实外部 Provider 的阶段只能标记为 fake/offline，不得标记为生产集成完成。

---

## Task 0: 建立 FastImg 模块映射和基线

**Files:**
- Create: `go-vue-admin/docs/fastimg-module-map.md`
- Modify: none in framework code

- [x] 记录 `backend/app/modules`、`backend/app/services`、`backend/database/migrations`、`backend/routes/web.go`、`backend/bootstrap/migrations.go` 和 `backend/bootstrap/schedule.go` 的扩展位置。
- [x] 记录 Resource Generator、Registry discovery、OpenAPI、前端生成 API 的现有入口。
- [x] 检查 PostgreSQL/Redis 端口和 `backend/.env`，不修改用户服务配置。
- [x] 在 `backend` 运行 `go test ./...`，在 `admin` 运行 `pnpm exec vue-tsc --noEmit` 和 `pnpm run build`，保存基线结果（环境/工具链失败项见模块映射文档，不据此宣称全绿）。
- [x] 检查 Windows Laragon 主机上的 `127.0.0.1:5432` 和 `127.0.0.1:6379`，只使用项目现有 PostgreSQL/Redis 配置。

## Task 1: 生成标准套餐、文件夹、相册和广告资源

**Files:**
- Reuse existing: `backend/app/modules/plans/`、`backend/app/services/plans/`、`backend/database/migrations/20260923000001_create_plans_subscriptions_usage_tables.go` and the generic admin Resource page.
- Generate via command: the not-yet-existing folder, album, and advertising resources, their migrations and frontend modules.
- Modify generic extension only: `backend/app/core/resource/registry.go` and `backend/app/core/admin/controllers/*`; keep plan business rules in `backend/app/services/plans/`.
- Modify generated discovery files only through `admin:make-resource`

- [x] 复用已存在的手写 plans Resource 和迁移；不对现有文件再次执行 `admin:make-resource plans`，因为生成器会拒绝覆盖非生成文件（见 SDD ruling）。
- [x] 使用 `--scope=own --owner-field=user_id` 为用户文件夹、相册资源声明数据范围；广告位使用管理员权限控制。
- [x] 审阅生成的 model/request/repository/service/controller/manifest/routes/permissions/menu/migration。
- [x] 增加套餐权益 JSON/结构化字段、价格周期和启停状态的领域校验；挂接现有 Resource Engine 写入入口，兼容旧权益快照字段名。
- [x] 为 Resource Engine 增加可选的通用写入准备钩子；套餐规则保留在 `services/plans`，其他资源默认不变。
- [x] 为套餐权益建立内容版本和用户订阅快照，避免编辑套餐影响新建订阅；旧纯权益快照兼容但无法还原历史价格/名称。
- [x] 运行 advertising/folders/albums 模块检查和生成资源测试；迁移审阅后仅按用户授权应用到 `fastimg_dev`，未触及其他数据库。

## Task 2: 建立配额、用量流水和订阅服务

**Files:**
- Create: `backend/app/services/quota/`
- Create: `backend/app/services/billing/`
- Create: `backend/app/modules/plans/` 中的订阅用例和用量 API
- Create: 对应迁移和测试

- [ ] 实现 Free/Creator/Pro 的可配置权益读取。
- [ ] 实现 `ReserveStorage`、`ConfirmStorage`、`ReleaseStorage`、`CheckUploadLimit`、`CheckBandwidth` 和 `RecordUsage`。
- [ ] 使用 PostgreSQL 事务或 Redis 原子操作避免并发上传超限。
- [ ] 用量流水使用幂等 source key，重复事件不重复计量。
- [ ] 实现升级即时生效、降级周期末生效和超额只限制新增操作。
- [ ] 覆盖免费用户达到容量、并发上传、删除恢复和套餐切换测试。

> 会员端页面当前使用扁平 URL：`/` 为上传首页，`/media` 为本人媒体库，`/plans` 为本人套餐/用量；后台分别位于 `/admin/**`。具体会员页面范围和验证记录见 `docs/superpowers/plans/2026-09-23-member-admin-route-separation.md`。不表示升级/降级、支付履约或跨用户媒体管理已实现。订阅生命周期仍受本任务订阅状态/存储契约约束，不得以覆盖当前行或直接改 `plan_id` 实现伪升级/降级。

## Task 3: 存储 Provider 和上传会话

**Files:**
- Create: `backend/app/modules/storage/`
- Create: `backend/app/modules/uploads/`
- Modify: `backend/config/filesystems.go`，只增加命名磁盘/业务配置，不修改框架文件系统实现
- Modify: `backend/routes/web.go`，注册上传接口
- Create: 上传 API 和 Provider 合约测试

- [x] 定义 `StorageProvider` 接口：`Put`、`CompleteMultipartUpload`、`GetMetadata`、`Delete`、`CreateSignedURL`、`Copy`、`Exists`。
- [x] 实现 Local Provider，使用框架文件系统，不依赖绝对路径。
- [ ] 实现 S3-compatible Provider；若项目依赖不具备，先保留接口和 fake Provider，不修改锁文件绕过问题。
- [x] 创建上传会话时检查用户状态、真实文件类型声明、文件大小和配额。
- [x] 完成确认时验证对象存在、大小、哈希、会话归属和幂等键。
- [ ] 对象成功而数据库失败时记录补偿任务，不标记媒体 ready。

> 状态查询 API 已补齐，但恢复 Job、超时租约与人工重试尚未实现；对象写入后数据库确认失败会保留 processing 记录，不能视为恢复闭环。

## Task 4: 媒体资产、派生图和媒体库

**Files:**
- Create: `backend/app/modules/media/`
- Create: `backend/app/services/media/`
- Create: `admin/src/modules/media/`
- Create: `admin/src/locales/zh-CN/media.json`、`admin/src/locales/en-US/media.json`
- Modify: `admin/src/i18n/index.ts` 注册媒体命名空间
- Modify: `backend/routes/web.go`，注册媒体专用 API
- Create: 媒体归属、状态和派生图测试

- [x] 创建 `MediaAsset`、`MediaVariant`、`StorageObject`、`UploadSession` 的模型和迁移（仅代码审阅，尚未执行迁移）。
- [ ] 完整实现媒体状态迁移；当前另新增 `cleanup_pending -> physically_deleted` 清理迁移，expired 清理和后台恢复流程仍未完成。
- [ ] 完成媒体列表、私有预览、回收站和恢复；所有者确认后永久删除 API 已实现，会员端只读详情现已实现，详情编辑、批量操作和后台清理仍未实现。
- [ ] 自定义媒体 API 与 Repository 双重执行用户归属检查；尚未通过 Resource Engine manifest 表达 own scope。
- [ ] 已提供 shadcn-vue 媒体库网格切片；列表/详情工作台和浏览器验收未完成。
- [x] 当前媒体库页面、导航、空/错误/恢复状态通过 vue-i18n namespace key 展示；中英文 JSON 同步并通过 Key 一致性测试。媒体详情/编辑/批量操作等仍按其他条目待完成。
- [x] 建立会员扁平入口 `/`、`/media`、`/plans`，管理页面统一收拢到 `/admin/**`；旧 `/app*` 会员地址保留明确重定向，详情见 `2026-09-23-member-admin-route-separation.md`。
- [x] 会员首页与媒体库共用上传队列和处理状态逻辑；首页显示最近本人媒体和套餐/存储摘要。媒体库继续提供本人媒体预览、搜索、软删除和恢复；请求不发送 `user_id`。
- [ ] 管理员跨用户媒体运营工作台仍未实现：不注册 `/admin/media`，也不通过 own-scope 文件夹/相册资源冒充全站管理。
- [ ] 会员端稳定链接复制、详情编辑/批量操作、媒体关联整理、Token、订单和发现页仍待后续阶段；只读媒体详情页已提供 `/media/:id`，文件夹/相册基础 CRUD 已提供 `/folders`、`/albums`。
- [x] 使用本地可测图片处理器生成缩略图、中图并清理源元数据；异步处理、WebP/AVIF 尚未实现。

## Task 5: 分享链接和稳定访问 URL

**Files:**
- Create: `backend/app/modules/shares/`
- Create: `backend/app/services/shares/`
- Create: `admin/src/modules/shares/`
- Modify: `backend/routes/web.go`，注册分享和公开访问路由
- Create: 分享 Token、密码、过期和防盗链测试

- [ ] 创建 `ShareLink`，只保存 token hash、状态、过期时间和密码摘要。
- [ ] 实现 URL、Markdown、HTML、BBCode 和 Variant 链接生成。
- [ ] 服务端校验私有性、密码、过期、防盗链和访问权限。
- [ ] 访问事件异步写入，不阻塞图片返回。
- [ ] 对象存储使用私有桶/私有目录，公开访问通过应用策略或签名 URL。

## Task 6: Personal API Token 和开发者 API

**Files:**
- Create: `backend/app/modules/developer/`
- Create: `backend/app/services/developer/`
- Create: `admin/src/modules/developer/`
- Modify: `backend/routes/web.go` 和 OpenAPI 契约
- Create: Personal API Token 和幂等上传测试

- [x] Personal API Token 只保存 hash，创建时只返回一次完整值（源代码已实现；迁移和运行验收待后续）。
- [x] 实现撤销、轮换、权限范围、过期和最后使用时间；会员端不能恢复已撤销 Token，后台可按权限停用/撤销。
- [ ] API 上传复用同一上传、媒体和配额服务，禁止复制网页上传逻辑。
- [x] 无 active 订阅时自动补齐 Free；仅在完全缺少 `free` 套餐记录时创建默认 Free，已停用的 Free 不自动恢复。
- [ ] 支持 IP、用户、Key 限流和 `Idempotency-Key`。
- [x] 更新 OpenAPI；`admin/src/generated/api.ts` 需在可访问运行后端后重新生成。

### Token API 交付细化

- [x] 将用户界面术语统一为 `Personal API Token`，内部表名使用 `api_tokens`。
- [x] 实现 Token 创建、列表、删除、轮换、过期和最近使用信息；后台 `/admin/api_tokens` 只读运营字段并支持批量停用/撤销，以及独立删除权限控制的删除动作。
- [x] 管理端不声明或返回 `token_hash`，通用资源列表统一执行 Manifest 字段投影；管理员只能看到用户 ID、前缀、Scope、状态和使用元数据。
- [x] 首期固定基础能力为 `upload:write`、`links:read`、`media:delete`；可选 Scope 为 `media:read`、`usage:read`、`webhook:manage`。
- [x] 实现 `Authorization: Bearer <token>` 认证，不接受 URL query 或表单字段传递 Token（迁移执行后验收）。
- [ ] 实现 `POST /api/v1/uploads` 的 multipart 上传、表单元数据和 `Idempotency-Key`，同时支持登录会话和 Personal API Token。
- [x] 实现 `POST /api/v1/uploads/batch`，按文件返回独立状态，限制最多 5 个文件、单文件 10 MB、批次总大小 50 MB；使用派生幂等键串行处理。真实 Token/数据库/运行进程验收仍待集成阶段。
- [ ] 实现 `/api/v1/uploads/sessions` 分片会话接口，用于大文件和断点续传。
- [ ] 让普通上传和分片上传共同调用 `CompleteUpload` 用例，不复制媒体写入、配额、审核和用量逻辑。
- [x] 实现 `GET /api/v1/media` 的 `media:read` 可选权限、统一分页和筛选（Token 路由源码已接入）。
- [ ] ready 时返回 `201` 和 `links.original/thumbnail/medium/webp/avif/url/markdown/html/bbcode`。
- [ ] processing 时返回 `202`、媒体 ID 和 status URL；失败时返回稳定错误码和重试状态。
- [x] Token 默认可以获取自己图片链接和删除自己图片；所有操作必须执行 Token 所属用户的资源归属校验（真实数据库验收待迁移）。
- [ ] 增加 curl、JavaScript、PicGo/ShareX 配置示例和完整 API 契约测试（OpenAPI Token 路径/Schema 已补）。
- [x] 将开发者 API 示例和响应字段写入 `docs/fastimg-developer-api.md`，实现与文档同一版本交付；运行/数据库限制已明确记录。
- [ ] 增加单文件、批量、部分失败、重复幂等键和分页查询的 API 契约测试。
- [x] 按前端设计实现 Token 一次性展示、复制、轮换和撤销确认；Token 上传状态、配额和真实 API 验收待下一切片。

### 广告位简化约束（2026-09-24）

- [x] 广告后台只保留 `header`、`footer`、`left`、`right` 四个固定位置，继续使用现有 advertising Generic Resource。
- [x] 新建/编辑位置选项和中英文文案已同步；本切片不增加 Campaign、竞价、复杂定向或点击统计。
- [ ] 会员端广告展示、套餐免广告和素材审核在广告投放切片中实现，不把后台 CRUD 误认为完整广告系统。

### 管理员广告内容与会员展示（2026-09-24）

- [x] 广告只由管理员配置；新增文本、图片、JavaScript 三种内容类型及 `creative_content`，会员端没有广告写入能力。
- [x] 新增认证会员只读 `/api/v1/ads`，MemberShell 按四个固定位置展示；脚本使用无同源权限的 `sandbox="allow-scripts"` iframe，文本不使用 `v-html`。
- [x] 生成并注册 `20260924000004_add_advertising_creative_fields`，兼容旧 `creative_url` 图片数据；迁移尚未执行。
- [ ] 套餐免广告、广告审核、展示/点击统计、复杂定向仍属于后续商业化切片。

## Task 7: 回收站、清理任务和失败恢复

**Files:**
- Modify: `backend/bootstrap/schedule.go`，注册清理调度
- Create: `backend/app/services/storage/cleanup.go`
- Create: 队列 Job、失败任务查询和管理端重试 Action
- Create: 清理和恢复测试

- [x] 删除进入软删除/回收站，不立即物理删除。
- [x] 所有者通过 `DELETE /api/v1/media/{id}/permanent` 显式确认同步永久删除：所有对象删除成功后才原子冲销存储流水；失败保持 `cleanup_pending` 且不释放额度。
- [ ] 异步清理对象、派生图和孤儿对象（定期调度/后台批处理仍未实现）。
- [ ] 处理超时上传会话、失败处理任务和对象删除失败重试（用户可对 `cleanup_pending` 媒体重复调用永久删除接口恢复失败清理；定期队列、失败任务查询与管理员重试仍未实现）。
- [ ] 每个任务具备幂等键、最大重试、退避、失败原因和审计事件。
- [ ] 任务状态异常时可人工重试，不能永久卡在 processing（上传可经 `POST /api/v1/uploads/{id}/retry` 恢复；管理员操作和其他任务类型未覆盖）。

## Task 8: 审核、举报和发现页

**Files:**
- Create: `backend/app/modules/moderation/`
- Create: `backend/app/services/moderation/`
- Create: `admin/src/modules/moderation/`
- Create: `admin/src/modules/discovery/`
- Modify: `backend/routes/web.go`

- [ ] 定义审核 Provider 接口，首期提供 offline fake 和 manual review。
- [ ] 只有公开、ready、approved 媒体才允许进入发现查询。
- [ ] 实现举报、隐藏、恢复、用户封禁和审核原因。
- [ ] Provider 不可用时保持 pending/manual_review，不自动公开。
- [ ] 瀑布流实现分页、去重、空状态、失败重试、举报和隐私字段过滤。
- [ ] 增加发现页设置：总开关、审核模式、默认排序和举报阈值；普通公开 ready 媒体自动进入发现页。
- [ ] 实现举报处理动作：驳回、隐藏、删除、恢复、限制上传、暂停和封禁。
- [ ] 为每次治理动作保存原因、证据、操作者、时间、处罚期限和审计记录。
- [ ] 验证停用发现页不影响私有媒体、用户自己的分享链接和后台举报处理。
- [ ] 实现同一用户/媒体/原因的举报去重和重复提交返回原举报 ID。
- [ ] 将举报阈值实现为复核触发器，不允许仅凭数量直接永久封禁账户。
- [ ] 增加处罚通知、用户申诉、管理员复核和申诉结果通知。
- [ ] 将发现页设置、举报查看、举报处理、账户处罚拆成独立权限并覆盖越权测试。
- [ ] 按治理策略覆盖 critical/high/medium/low 风险级别、举报去重、申诉和处理时限指标。
- [ ] 在上传、Token、发现页、举报和注销流程中链接服务条款、隐私政策、可接受使用政策和数据删除说明。
- [ ] 增加 EXIF 清理、数据保留、Token 泄露和跨用户访问事件的验收记录。
- [ ] 增加存储、流量、回收站、孤儿对象和异常流量指标，验证 Free 套餐不会绕过资源边界。
- [ ] 实现防盗链模式、域名白名单、无 Referer 策略和 Signed URL 的 Provider/领域测试。

## Task 9: 订单、广告和统计

**Files:**
- Create: `backend/app/modules/billing/`
- Create: `backend/app/modules/advertising/`
- Create: `backend/app/modules/analytics/`
- Create: `backend/app/services/billing/`、`advertising/`、`analytics/`
- Create: `admin/src/modules/billing/`、`advertising/`、`analytics/`
- Modify: `backend/routes/web.go`、`backend/bootstrap/schedule.go`

- [ ] 订单状态实现 `created -> pending_payment -> paid -> active` 和退款分支。
- [ ] 定义 `PaymentGateway` 接口和 fake/manual Provider；业务模块不得直接依赖具体支付渠道。
- [ ] 创建 Order、OrderItem、PaymentIntent、PaymentTransaction、PaymentEvent 和 Subscription 权益快照。
- [ ] 支付回调按订单号、流水号和事件 ID 幂等；金额、货币、签名和订单状态必须服务端校验。
- [ ] 将“支付成功”和“权益履约成功”拆开，履约失败进入可重试补偿队列。
- [ ] 覆盖升级、续费、取消、部分退款、全额退款、退款失败、拒付和人工调整流水。
- [ ] 实现渠道可用性和路由配置，按货币、地区、渠道健康状态过滤支付 Provider。
- [ ] 保存价格、优惠、货币、套餐版本和权益快照；首期只支持固定金额/百分比优惠。
- [ ] 提供订单收据/账单摘要，预留 Invoice Provider，不把收据标记为税务发票。
- [ ] 增加渠道流水对账模型和差异状态，差异只能进入人工复核或补偿流程。
- [ ] 验证支付中刷新、重复提交、支付成功但履约失败和渠道熔断场景。
- [ ] 预留 Crypto Provider：资产、网络、报价、收款地址、确认数和链上交易哈希与法币支付统一关联。
- [ ] 覆盖加密支付少付、多付、报价过期、错误网络、确认不足、链重组、重复交易和风控复核。
- [ ] 广告位、素材、周期、套餐免广告规则和展示/点击事件可配置。
- [ ] 异步记录媒体访问、上传、下载、删除、审核和订单事件。
- [ ] 提供日/月聚合，计费用量不能依赖前端统计。

## Task 10: OpenAPI、前端验收和发布

**Files:**
- Modify: `backend/app/openapi/` 相关契约注册
- Modify: `admin/src/generated/api.ts`，通过生成脚本更新
- Use and update: `docs/fastimg-release-runbook.md`
- Create: 集成测试、端到端验收清单和部署配置说明

- [ ] 验证 `/api/openapi.json` 包含新增接口和错误码。
- [ ] 运行后端 `go test ./...`。
- [ ] 运行前端 `pnpm exec vue-tsc --noEmit` 和 `pnpm run build`。
- [ ] 使用真实 PostgreSQL/Redis 配置验证迁移、队列、限流和任务失败恢复。
- [ ] 验证免费用户完整闭环、跨用户访问拒绝、并发配额、重复支付和审核故障。
- [ ] 每个阶段创建本地里程碑提交，验证完成前不推送 GitHub。
- [ ] 发布报告明确区分 `designed`、`implemented`、`integrated`、`verified`、`enabled` 和 `disabled`，不得把 fake/offline Provider 标记为生产完成。

## Vertical Slice Delivery

不要按“先把所有 Model 写完、再把所有页面写完”的方式交付。每个里程碑按可运行的垂直切片完成：

### Slice A：Free 上传闭环

- 套餐读取和存储配额
- UploadSession 创建/完成
- Local Storage Provider
- MediaAsset 和 Variant
- 用户媒体库
- 基础链接复制
- 删除、回收站和恢复

验收：新用户可以不付费上传图片、查看图片、复制链接、删除并恢复；超出配额时有明确错误。

### Slice B：开发者闭环

- Personal API Token
- API 上传
- 幂等键
- API 限流
- API 配额响应头
- OpenAPI/TypeScript Client

验收：使用 Personal API Token 完成上传、查询和删除，撤销 Token 后立即失效。

### Slice C：安全公开闭环

- 公开性和分享策略
- 审核 fake/manual
- 举报和隐藏
- 发现页
- 访问事件

验收：只有公开、ready、approved 媒体进入发现页；密码、过期和撤销链接在服务端生效。

### Slice D：运营闭环

- 管理端按用户查看媒体
- 任务重试
- 用量流水
- 运营仪表盘
- 广告位和 Free 展示规则

验收：管理员可以定位用户、媒体、任务和用量，但所有敏感操作有审计记录。

### Slice E：商业化闭环

- Creator/Pro 套餐
- fake/manual 订单
- 升级/降级
- 防盗链和自定义过期
- 高级统计

验收：升级即时生效，降级周期末生效，重复支付事件不会重复激活订阅。

## Definition of Done

一个任务只有同时满足以下条件才算完成：

- 代码位于正确的模块或 Service 目录，没有将业务逻辑放入 Admin Core。
- API、权限、数据范围、字段可读写策略和错误码已经定义。
- 正常路径、资源越权、配额边界、依赖失败和重复请求都有测试。
- 迁移已人工审阅，未使用未验证的数据库特性。
- 后端 `go test ./...`、前端 `pnpm exec vue-tsc --noEmit` 和必要的 `pnpm run build` 通过。
- 异步任务具有幂等键、重试、失败记录和人工恢复路径。
- 文档中更新了真实已完成能力和仍为 fake/offline 的 Provider。
- 形成独立本地提交；阶段完成前不推送远程仓库。

## Release Gates

发布 Free MVP 前必须通过：

1. 新用户自动获得 Free 套餐，且不用付费即可上传、管理、分享和使用基础 API。
2. 并发上传无法突破存储、单文件和每日上传限制。
3. 用户无法读取、修改、导出或删除其他用户媒体。
4. 对象存储、Redis、队列、图片处理和审核失败时不会错误标记成功。
5. 删除、恢复、永久清理和用量扣减可以通过流水和任务状态追踪。
6. 未审核、私有和密码媒体不会进入公共发现页。
7. 重复上传完成、支付回调和 Webhook 不会造成重复扣减或重复激活。
8. 管理员操作、封禁、额度调整和永久删除均有审计记录。
