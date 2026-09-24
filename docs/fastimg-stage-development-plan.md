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
- 所有新增或修改的用户可见文案必须在同一变更中同步实现 `zh-CN`、`en-US`；前端按 `docs/i18n.md` 使用 vue-i18n 与 `admin/src/locales/<locale>/<namespace>.json`，后端按 Goravel Localization 返回稳定错误码/Key/参数。
- 禁止页面硬编码中文/英文、`t(中文, 英文)` 内联双语辅助函数或另建语言状态/翻译引擎；每个功能阶段必须验证两种语言文件路径及 JSON Key 集合一致。

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
POST   /api/v1/uploads/{id}/retry
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

### 当前实现进度（2026-09-23）

当前是可供内部开发预览的多个纵向切片，不代表 M1/M2 验收完成，更不可作为生产就绪声明。

- 已实现：Local Provider、图片内容/扩展名/MIME/大小/像素/动画帧校验、重新编码清除源元数据、thumbnail/medium 生成、上传会话与幂等键、用户/套餐/配额检查、对象写入校验、私有媒体列表/预览/回收站/恢复，以及 `GET /api/v1/uploads/{id}` 状态查询和 `POST /api/v1/uploads/{id}/retry` 用户恢复入口。
- 恢复规则：processing 会话的 `updated_at` 已超过 10 分钟才可被恢复请求接管；接管会在用户行锁内刷新更新时间。逐个核验登记对象的 MIME、大小和 SHA-256，全部匹配才确认用量并标 ready；对象缺失/损坏则删除可清理对象、释放预占并标记 failed，要求用户重新上传。
- 已实现后台资源：plans、广告位、文件夹、相册进入后台资源导航；文件夹/相册启用 `user_id` own scope，广告位使用后台权限。Generated Resource Engine、RBAC 与审计复用框架，不复制实现。
- 已实现套餐快照：新订阅在现有快照列写入套餐条款、权益和 SHA-256 内容版本；配额读取兼容新 envelope 与旧纯权益 JSON，订阅 API 优先显示快照并校验版本。旧订阅缺失历史价格/名称，只能兼容显示当前套餐并返回 `snapshot_available: false`。该变更未新增或运行数据库迁移。
- 已实现用量摘要/上传入账切片：`GET /api/v1/me/usage` 按认证用户读取套餐快照额度；存储汇总为生命周期余额，API/上传/流量/处理等按 UTC 月周期汇总。上传完成在同一事务内通过 quota Service 记录 SHA-256 幂等存储流水、月度 `upload` 次数流水和 `transform` 流水，并兼容读取旧 `storage_bytes` 名称；一次成功的媒体处理作业（生成原图、缩略图和中图）计为一次 transform，处理中的上传会话预占一次，失败会话不计数。月度上传和处理流水按上传会话创建时所属 UTC 月记账，与并发预占统计周期一致，缺失创建时间的历史会话回退到完成时间。上传在图像解码前进行额度预检，拒绝已知超额请求；`BeginUpload` 在用户行锁事务中复核并原子预占实际 Variant 字节，避免并发请求仅凭预检越额。`GET /api/v1/me/usage/ledger` 仅按认证用户范围分页返回流水，不暴露幂等键。上传配额检查在用户行锁事务内为历史缺少活动订阅的用户补建 Free 快照；原因是 Goravel `Query.First` 对 0 行返回 nil，不能用来判定不存在，需使用 `Exists()`。每日上传限制已接入；月度 API 上传限制按 UTC 月统计本人 `ready` 与 `processing` 上传会话，失败会话不计数，用户行锁串行化并发预占；月度转换限制合计已完成 transform 流水和本人 `processing` 会话，同样串行化预占。额度耗尽分别返回 HTTP 429 / `MONTHLY_API_UPLOAD_LIMIT_REACHED` 或 `MONTHLY_TRANSFORM_LIMIT_REACHED`。本轮补入所有者专属永久删除 API：回收站图片先进入 `cleanup_pending`，Local/未来 Provider 删除所有唯一 Variant 对象成功后才在同一事务标记对象、Variant、媒体 tombstone 并写负向 storage ledger；删除失败可重试且不提前释放额度。带宽流水与限制、订阅状态迁移仍未完成；软删除不得提前释放。
- 多语言约束：`docs/i18n.md` 是各阶段强制开发与验收规则；媒体库页面和后台导航使用 vue-i18n 的 `media.*` Key，注册中英文 JSON 语言包。i18n 专项测试覆盖语言文件路径/嵌套 Key 及页面/导航 Key 使用。
- 当前前端形态：会员端使用 `/`、`/media`、`/plans` 和 `MemberShell`；管理端页面统一在 `/admin/**` 和 `AdminShell`。`/admin/media` 尚未提供跨用户媒体管理 API/RBAC/审计支撑，因此不注册；文件夹/相册为 own-scope，不作为全站管理资源。
- 会员首页 `/` 提供上传入口、最近本人媒体及套餐/存储摘要；`/media` 调用认证后的 `/api/v1/uploads` 与 own-scope `/api/v1/media`，首页和媒体页共用上传队列、处理状态轮询、配额/大小错误与重试反馈。本人媒体页支持受保护缩略图预览、文件名搜索、服务端分页（48 条/页）、软删除和回收站恢复；搜索/回收站切换回到第一页，删除当前页最后一张后自动回退到有效页。会员媒体详情现返回并可复制认证内容端点的 URL、Markdown、HTML、BBCode；这些不是公开稳定分享链接。`/tokens` 已提供 Personal API Token 的创建、一次性展示、撤销和轮换界面，后端已加入 hash、Scope 和会员认证源码。订阅及用量仍从公开套餐目录、当前认证用户订阅和 `/me/usage` 读取，流量未计量时不伪报 0，不下单、不升级、不更改订阅。分页、链接和 Token 文案通过 `member` namespace 同步中英文；会员前端不传 `user_id`，所有权由服务端认证身份决定。会员端防盗链、详情编辑、批量管理、相册关联、订单中心和 `/discover` 尚未完整实现；Token 迁移、运行进程重启以及 API 上传/链接/删除真实验收尚未执行。
- 本轮验证：`go test ./... -count=1` 全部 Go 包通过（含 Feature）；`pnpm run build` 通过；`node --test tests/fastimg-i18n.test.mjs` 4/4 通过。完整 admin Node 测试为 23/24，唯一失败仍是 `tests/resource-actions.test.ts`：测试要求过滤未知 action kind，但通用实现返回 `unknown`；该文件与对应实现不属于本轮改动。
- 会员媒体页本轮专项 RED/GREEN 与 i18n/member-plan 回归测试在 Windows Node fallback 中通过 15/15。构建尚未验证：WSL 初始化返回 `E_ACCESSDENIED`；Windows pnpm 尝试在线解析/重装 pnpm 后因无交互终端中止，未能进入 TypeScript/Vite 构建。现有 Vite 53081 仍可服务旧应用，但其 `/src/router/index.ts` 和 `MemberShell` 热服务内容不含新路由/导航键，故浏览器验收未通过；未杀停工作中的服务，等 WSL 可用后只重启本项目 dev server 再验收。没有提交真实图片上传、执行迁移或变更数据库/服务配置。
- `fastimg_dev` 的只读迁移状态确认 plans、media、advertising、folders、albums 迁移均为 Ran；本轮明确授权应用的仅为广告位、文件夹、相册及其管理权限。没有修改 Laragon PostgreSQL/Redis 配置。Windows 侧 5432/6379 可连接，但 WSL 内启动的后端命令仍尝试访问 `127.0.0.1:6379` 并出现连接拒绝警告；Redis 的 WSL 访问路径和队列运行验收仍未完成。
- 本地预览（2026-09-23 20:29 更新）：Vite `53081`（PID 896）继续代理到 Windows API `53082`；后端从更新后的源码重新编译并只重启仓库自有进程（旧 PID 11404 已核实为 `fastimg-dev-next.exe`，新 PID 28316 为 `fastimg-dev-next-updated.exe`）。当前 WSL NAT 网关为 `172.29.160.1`，仅通过进程环境绑定，未写入 `.env`。`wslnet url 53081` 验证 WSL/Windows 两侧 HTTP 200；Windows 直连后端网关和经 Vite 同源请求 `/api/v1/auth/me` 均返回预期未认证 401，证明 API 和代理路径可达。Windows PostgreSQL/Redis 端口 5432/6379 可连接；WSL Feature bootstrap 仍会打印其自身 `127.0.0.1:6379` 连接拒绝告警，但本轮全量 Go 测试通过。没有执行迁移或真实图片上传；浏览器真实上传仍未验收。日志和 PID 文件位于 `backend/storage/logs/fastimg-dev-20260923-202958.log`、`backend/storage/bin/fastimg-dev-20260923-202958.pid`。
- 本地预览纠正（2026-09-24）：对 `53081/admin/plans` 白屏排查后确认，WSL listener PID 759 的 cwd 正是本 checkout 的 `admin/`，不是其他项目；根因是 Vite 进程长时间运行，内存中 AdminShell 转换模块落后于磁盘源码（源码含“前往用户端”，服务响应不含）。仅重启已核实归属的 PID 759，保持端口 `53081`，现 PID 880；日志/PID 为 `/tmp/fastimg-vite-53081-20260924.log` 与 `/tmp/fastimg-vite-53081-20260924.pid`。`wslnet url 53081` 验证 WSL 和 Windows 均 HTTP 200，`/api/v1/auth/me` 未登录返回预期 401。Codex 浏览器已验证 `/admin/plans` 展示套餐管理表格及“前往用户端”链接，会员 `/plans` 展示套餐/用量及管理员可见的“管理后台”入口；无登录态的独立浏览器会正常显示登录表单。白屏已通过重启本项目 Vite 消除；未碰其他项目、数据库、迁移或 Laragon 配置。生产构建仍未验证（Windows 缺少 Rolldown 原生可选绑定；本轮使用 WSL Vite dev server 验收）。

- 会员媒体详情切片（2026-09-24）：新增认证后 `GET /api/v1/media/{id}`，Service/Repository 先按当前用户验证归属，再返回本人 ready 媒体元数据与已就绪的原图/缩略图/中图；越权与不存在统一返回 `MEDIA_NOT_FOUND`。会员端新增 `/media/:id` 只读详情页、从媒体卡片进入的入口和 zh-CN/en-US 文案；图片预览继续经认证内容端点读取，不生成公开或稳定链接。同步 OpenAPI 契约及 Service/API 路由/i18n 测试。验证：Go 媒体 Service 与 OpenAPI 测试通过，前端媒体/路由/i18n 专项 15/15 通过。生产构建未能启动：Windows pnpm 检查试图移除 modules 并要求交互确认，本轮未确认、未更改依赖或锁文件；Vue 类型检查/构建仍待可用工具环境验证。未执行迁移、数据库写入或重启开发服务。
- 会员文件夹/相册基础切片（2026-09-24）：新增认证后的 `/api/v1/me/folders` 与 `/api/v1/me/albums` 列表、创建、更新、删除接口，复用业务 Service/Repository；所有写入强制使用认证用户，父文件夹和相册封面媒体必须属于当前用户，名称限制为非空且不超过 120 个字符，相册访问策略限制为 `private`/`unlisted`/`public`。新增会员 `/folders`、`/albums` 页面及 MemberShell 导航，中英文文案与 OpenAPI 契约同步。当前只完成容器 CRUD，尚未新增媒体关联表，因此页面明确提示媒体整理关联将在后续切片接入；后台 `/admin/folders`、`/admin/albums` 仍是管理员资源，不与会员页面重合。验证：`go test ./... -count=1` 全部通过；前端会员集合/媒体/路由/i18n 专项 18/18 通过。未执行迁移、数据库写入或生产构建；WSL 当前返回 `E_ACCESSDENIED`，无法启动或用 `wslnet url` 做浏览器运行态验收。
- 会员文件夹/相册运行态修复（2026-09-24）：发现生成资源的旧版平铺重定向会抢占 `/folders`、`/albums`，导致会员页提示无匹配路由。已同步修复 `index.ts` 与 Vite 实际解析的 `index.js`：旧后台重定向排除这两个会员路径，管理员资源继续固定在 `/admin/folders`、`/admin/albums`。重启并确认仅本仓库的 Vite 会话后，浏览器已加载 `/folders` 和 `/albums`，导航、空状态、创建表单均正常，无新增控制台错误。当前 Vite 使用项目专属 tmux 会话 `fastimg-vite53081`、WSL PID 1552，后端为 Windows PID 20848、端口 53082；`wslnet url 53081` 验证 WSL/Windows 均 HTTP 200。`go test ./... -count=1` 与前端专项 18/18 仍通过；生产构建仍未验证，未执行迁移或数据库写入。
- 未完成：S3 Provider、真正异步处理/恢复队列、批量和分片上传、媒体与相册关联及批量整理、管理员跨用户媒体运营工作台、Token 迁移与真实 API 上传/链接/删除验收、WebP/AVIF、自动清理和后台失败任务中心。
- 重要限制：数据库最终确认失败会保留 `202 processing` 会话/对象；10 分钟后可由用户手动调用恢复接口，但自动队列扫描、失败任务后台、管理员重试与定期清理仍未实现。新增的永久删除由所有者同步触发，并支持同请求重试；它不等于 Task 7 的异步回收站/孤儿对象调度。租约阈值针对当前 10 MB 同步 Local 上传；扩展到远程或长时间 Provider 前必须改为可续期 lease/Job ownership。

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

## 十四、2026-09-24 开发进度：会员媒体文件夹归档

### 已实现

- `MediaAsset` 增加可空 `folder_id`，新增 `20260924000001_add_media_folder_id` 迁移并注册到 Goravel migration 列表。
- 新增会员接口 `PATCH /api/v1/media/{id}/folder`，只接受当前认证用户自己的 ready 媒体和自己的文件夹；传 `folder_id: null` 可移回未归档。
- 媒体列表和详情契约返回 `folder_id`；跨用户或不存在的文件夹返回 `FOLDER_NOT_FOUND`，不存在或非 ready 媒体返回 `MEDIA_NOT_FOUND`。
- 会员 `/media` 加载自己的文件夹，并为每张图片提供文件夹选择器；失败提示同步写入 zh-CN/en-US。
- OpenAPI、后端路由、媒体服务、前端契约测试和双语文案已同步。

### 验证与边界

- `go test ./app/services/media ./app/modules/media/controllers ./app/openapi ./database/migrations ./bootstrap` 通过。
- `node --test --test-isolation=none tests/member-media.test.mjs tests/fastimg-i18n.test.mjs tests/fastimg-routes.test.mjs tests/member-collections.test.mjs` 通过，20/20。
- `pnpm run build` 和真实 PostgreSQL 迁移尚未在本切片执行；迁移文件已审阅但没有写入 `fastimg_dev`，因此当前运行中的旧后端不会提供该接口。
- 相册媒体关联仍然延期，后续需要明确单媒体多相册关系、排序和批量操作契约后再建 pivot 表。

## 十五、2026-09-24 开发进度：认证链接格式

- 上传创建、状态查询、重试和会员媒体详情统一返回 `original`、`thumbnail`、`medium`、`url`、`markdown`、`html`、`bbcode`。
- 会员 `/media/:id` 提供 URL、Markdown、HTML、BBCode 复制控件，文案同步维护 zh-CN/en-US。
- 当前链接仍指向认证内容端点，未实现公开稳定分享、密码/过期分享、签名 URL 或防盗链；不得把本切片描述为 M3 完成。
- `go test ./... -count=1` 通过；前端会员/媒体/i18n/路由/集合专项测试 21/21 通过；本轮未执行迁移、数据库写入或后端重启。

## 十六、2026-09-24 开发进度：本人分享链接

- 新增 `share_links` 模型和迁移草案，保存 token hash、前缀、状态、媒体归属和可选过期时间；完整 token 只在创建响应中返回一次。
- 新增 `POST /api/v1/media/{id}/share-links`、`GET /api/v1/share-links`、`DELETE /api/v1/share-links/{id}` 和公开 `GET /s/{token}`。
- 创建和撤销均按认证用户校验媒体/分享归属；公开访问重新检查 active、过期、媒体 ready、Variant ready 和对象 ready 状态。
- 会员 `/media/:id` 支持创建分享并复制 URL，`/share-links` 支持查看摘要和撤销，双语文案同步完成。
- `go test ./... -count=1` 通过；前端会员/媒体/i18n/路由/集合专项测试 22/22 通过。
- 迁移尚未执行，当前运行后端未重启；密码分享、防盗链、签名 URL、访问事件和真实 PostgreSQL 验收仍未完成。

## 十七、2026-09-24 开发进度：Personal API Token 基础能力

- 新增 `api_tokens` 模型和迁移草案；只保存 SHA-256 token hash、非敏感前缀、Scope、状态、过期时间、最近使用时间/IP 和调用次数，完整值只在创建或轮换响应中返回一次。
- 新增会话认证的 `POST /api/v1/tokens`、`GET /api/v1/tokens`、`DELETE /api/v1/tokens/{id}`、`POST /api/v1/tokens/{id}/rotate`；基础 Scope 固定为 `upload:write`、`links:read`、`media:delete`，可选 `media:read`、`usage:read`、`webhook:manage`。
- 会员资源认证中间件支持登录会话或 `Authorization: Bearer fst_...` Personal API Token；管理员 `/api/v1/admin/**` 仍只接受框架会话认证，Token 不能进入后台 RBAC。
- 会员 `/tokens` 页面支持创建、一次性复制、撤销、轮换和中英文 Scope 展示；不把完整 Token 写入 URL、LocalStorage 或日志。
- `go test ./... -count=1` 通过；前端全量 `.mjs` 合约测试通过 39/39，包含 i18n、路由、Token 页面、媒体、集合和计划边界。
- 迁移尚未执行，当前运行后端未重启；真实数据库 Token 创建、Token 上传、链接读取、自己媒体删除和撤销即时失效仍待下一次集成验收。生产环境还需要按用户/IP/Token 限流、审计脱敏和批量/分片上传。

## 十八、2026-09-24 开发进度：管理端 Token 运营边界

- `api_tokens` 已接入后台 Generic Resource，管理 URL 为 `/admin/api_tokens`；会员端 `/tokens` 仍是用户自己的 Token 管理页面，两者不共用页面权限。
- 管理员列表按用户查看 Token 的用户 ID、名称、前缀、Scope、状态、过期时间、最近使用时间/IP、调用次数和撤销时间；Manifest 不声明 `token_hash`，后台不会读取或导出 Token 摘要。
- 管理端只提供批量停用/撤销动作，动作通过 Admin Action Handler 调用开发者领域 Service；不允许通用 CRUD 删除 Token，也不允许把状态直接恢复为 active。
- 通用资源列表现在统一执行字段权限投影，防止数据库中未声明的内部字段出现在列表响应中；中英文资源、字段、状态和导航分组文案已同步。
- 权限新增 `admin.api_tokens.view` 和 `admin.api_tokens.update`，默认管理员 Seeder 会补齐；迁移和 Seeder 尚未在 `fastimg_dev` 执行，运行中的后台需在后续集成阶段重启后验收。
- `go test ./... -count=1` 通过；前端 `.mjs` 合约测试 39/39 通过。生产构建/TypeScript 独立检查受当前 Windows 依赖镜像缺失类型包影响，未宣称通过。

## 十九、2026-09-24 开发进度：Free 套餐兜底

- 上传配额检查仍然优先读取用户已有 active 订阅；没有 active 订阅时会尝试补齐 Free。
- 如果 `plans` 表存在但没有任何 `free` 套餐记录，运行时会创建当前默认 Free 权益（只创建缺失记录）；如果管理员明确将已有 Free 套餐停用，则返回 `SUBSCRIPTION_UNAVAILABLE`，不会自动复活套餐配置。
- 该兜底解决“只执行迁移、未执行种子导致管理员和会员都无法上传”的开发环境常见问题；首次请求仍需要数据库表已迁移且数据库连接正常。
- 新增 Free 套餐默认权益单元测试；`go test ./... -count=1` 通过。未执行数据库写入、迁移或服务重启，因此运行中的进程尚未现场验收。

## 二十、2026-09-24 开发进度：开发者 API 批量上传

### 已实现

- 新增会员/Personal API Token 共用的 `POST /api/v1/uploads/batch`，继续复用现有 `UploadService`、图片校验、配额、存储和幂等逻辑，没有复制一套上传业务。
- 批量请求接受多个 `files[]` 字段，最多 5 个文件、单文件最多 10,000,000 字节、批次总大小最多 50,000,000 字节；解析阶段错误整体拒绝。
- 批次要求 `Idempotency-Key`，按文件顺序派生独立内部键，逐文件串行处理；业务阶段允许部分成功，返回 HTTP `207`，每项返回成功数据或稳定错误码及最终汇总。
- 批量请求体限制、路由、OpenAPI Schema、控制器解析测试和幂等键测试已同步。

### 验证与边界

- 定向批量上传解析/幂等键测试、OpenAPI 契约测试通过；`go test ./... -count=1` 全量通过。
- 尚未执行迁移、数据库写入、后端重启、真实 Token 上传或真实 PostgreSQL/Redis 集成验收；前端生产构建和 TypeScript 独立检查仍受当前 Windows 依赖镜像缺失类型包影响。
- 分片上传、按套餐动态批量上限、上传进度回调和批量取消仍未实现，不能把本切片描述为完整开发者上传平台。

## 二十一、2026-09-24 开发进度：广告位简化为四个固定位置

### 已实现

- 广告 Resource 的 `placement` 选项收敛为 `header`、`footer`、`left`、`right`，后台地址保持 `/admin/advertising`，继续复用 Generic Resource、RBAC、字段校验和审计边界。
- 中英文管理端选项文案同步为“页眉/页脚/左侧栏/右侧栏”和对应英文；旧的 `upload`、`dashboard`、`discovery` 不再出现在新建/编辑选项中。
- 当前切片只处理广告位元数据和启停管理，不新增 Campaign、竞价、复杂定向、分成或统计系统。

### 验证与边界

- 广告 Manifest 固定位置测试和前端双语契约测试已补齐；仅固定位置变更不需要迁移，内容类型切片另增 `20260924000004_add_advertising_creative_fields`。
- 会员端实际广告渲染已在下一切片接入；按套餐免广告判断、素材审核和点击统计仍是后续切片，不能把后台位置配置描述为完整广告投放系统。

## 二十二、2026-09-24 开发进度：管理员广告内容与会员端展示

### 已实现

- 管理员广告资源新增 `creative_type`（文本/图片/JavaScript）和 `creative_content`，写入经过广告 Service 校验；跳转地址只允许 HTTP/HTTPS，脚本内容要求为源码而不是嵌入式 `<script>` 标签。
- 新增会员只读 `GET /api/v1/ads?placement=header|footer|left|right`，只返回 `status=active` 的广告；没有会员广告写入接口，也不接受会员提交 `user_id`。
- `MemberShell` 接入页眉、页脚、左侧栏、右侧栏四个 `MemberAdSlot`。文本、图片和脚本分别以插值、图片、无同源权限的沙箱 iframe 渲染；不使用明显卡片，不使用 `v-html`。
- 共享 Generic Resource 表单补齐 `textarea` 和字段说明渲染；广告 `creative_content` 现在是清晰的多行等宽输入区，并通过中英文提示说明文本、图片 URL 与 JavaScript 源码的填写方式。
- 旧数据仍可读取：迁移前只有 `creative_url` 的行按图片广告兼容展示。管理端新建/编辑新的内容类型需要先应用广告字段迁移。

### 验证与边界

- `go test ./app/services/advertising ./app/modules/advertising/resource ./app/openapi -run 'TestPrepareAdvertisingWrite|TestManifestExposesAdminCreativeTypesAndContent|TestSpecDocumentsMemberAdvertisingContract' -count=1` 通过。
- `node --test --test-isolation=none tests/member-advertising.test.mjs` 通过 3/3；前端生产构建和运行态浏览器刷新尚未在本轮执行。
- 新增迁移 `20260924000004_add_advertising_creative_fields` 已生成并注册但未执行；没有重启后端、写入数据库或修改 Laragon PostgreSQL/Redis 配置。当前广告仍不包含套餐免广告判断、审核队列、展示/点击统计和复杂定向。

## 二十三、2026-09-24 开发进度：游客公开首页与套餐目录

### 已实现

- 移除 MemberShell 父路由的全局登录要求，`/` 和 `/plans` 可由游客访问；个人媒体、媒体详情、文件夹、相册、分享链接和 Personal API Token 子路由继续逐项要求登录。
- 首页保持“上传媒体”定位：游客看到上传说明和登录入口，登录后才启用文件选择、拖放和上传反馈；未修改共享上传 composable 的失败关闭逻辑。
- 套餐页先读取公开 `GET /api/v1/plans`，游客可查看免费/付费套餐及权益；订阅、用量和当前套餐只在存在会话时读取，游客显示登录后查看用量的入口。
- MemberShell 游客导航仅显示首页、套餐和登录入口；登录用户才看到个人媒体能力，管理员入口继续由权限判断控制。首页账户摘要和最近上传区均提供游客态文案。
- 同步补齐 zh-CN/en-US 文案、TypeScript 路由和运行时 JavaScript 路由镜像；没有新增迁移，也没有放宽后端上传认证。

### 验证与边界

- 先新增游客访问/上传 CTA/公开套餐契约测试，确认 RED；实现后 `node --test --test-isolation=none tests/fastimg-routes.test.mjs tests/member-home.test.mjs tests/member-plans.test.mjs` 通过 17/17。
- 前端完整 `.mjs` 契约套件通过 49/49；当前未执行生产构建，不能把开发依赖镜像状态描述为生产构建已验证。
- 未执行数据库迁移、数据库写入或修改 Laragon PostgreSQL/Redis；上传接口仍需在运行态通过未认证请求返回 401 后，再进行已登录真实上传验收。

## 二十四、2026-09-24 开发进度：开发预览 Tailwind 样式恢复

### 根因与处理

- 复现确认：样式异常时页面只注入 26 条 CSS 规则，包含 reset 和主题变量，但 `@layer utilities` 为空；浏览器页面结构正常、控制台无错误。
- 根因是隔离 Vite 配置为规避 Windows 缺失 Tailwind Oxide 原生包而移除了 `@tailwindcss/vite`，导致 `style.css` 中的 Tailwind v4 utility 无法生成。恢复插件后又确认依赖树缺少 `@tailwindcss/oxide-win32-x64-msvc`，原始配置会以 `MODULE_NOT_FOUND` 启动失败。
- 已恢复 `admin/vite.fastimg-isolated.config.mjs` 的 Tailwind 插件；开发依赖使用本机已有的同版本 Windows 原生包链接，不修改 `package.json` 或 `pnpm-lock.yaml`。该链接属于本地开发运行时补齐，依赖重装后需要重新确认。

### 验证与边界

- Vite 已在 `53083` 运行，后端在 `53084` 运行；前端页面 HTTP 200、公开套餐 API HTTP 200、未认证上传 API HTTP 401。
- 浏览器刷新已确认会员导航、卡片、表单、间距、颜色和边框样式恢复，控制台无 error/warning。
- 新增样式配置契约测试；先观察到 RED，修复后聚焦前端测试通过 25/25，完整 `.mjs` 测试通过 50/50。未执行生产构建，数据库和 Laragon PostgreSQL/Redis 未修改。

## 二十五、2026-09-24 开发进度：密码分享链接

### 已实现

- 分享链接创建接口新增可选 `password`，长度限制为 8–72 个字符；空值保持普通公开分享。
- 分享服务复用框架 Hash 能力，只向 `share_links.password_hash` 写入 Hash，不在模型序列化、列表接口或日志中暴露密码。
- 公开 `GET /s/{token}` 接口支持 `password` 查询参数；受保护链接缺少或密码错误统一返回 `401 SHARE_PASSWORD_REQUIRED`，不泄露对象源地址，也不区分“密码错误”和其他链接状态。
- 会员媒体详情页增加双语密码输入框，创建请求同时提交过期时间和密码；列表仍只显示 token 前缀和状态，完整 URL 只在创建成功时返回。
- OpenAPI、开发者 API 文档、中文/英文文案和服务层测试同步更新；不新增迁移，复用已存在的 `password_hash` 列。

### 验证与边界

- 先观察到服务层密码测试、OpenAPI 契约测试和会员端表单测试 RED；实现后 `go test ./app/services/shares ./app/openapi -count=1` 通过，会员媒体专项测试 `10/10` 通过。
- 前端 Node 测试在 Windows 默认隔离模式下受 `spawn EPERM` 影响，改用项目现有的 `--test-isolation=none` 模式完成验证；不是业务失败。
- 尚未执行数据库迁移、真实密码分享创建/访问或后端重启；隔离 Vite 生产构建和 `vue-tsc -b` 已通过。签名 URL、Referer 白名单/无 Referer 策略、访问记录和 CDN 缓存失效仍未实现。

## 二十六、2026-09-24 开发进度：会员导航紧凑化

### 已实现

- 会员端导航不再直接使用完整页面标题作为按钮文案，新增中英文 `member.nav.*` 短标签：首页/Home、套餐/Plans、图片/Images、分享/Shares、API Token/API 等。
- 每个导航链接继续使用完整页面标题作为 `aria-label` 和 `title`，保持键盘、屏幕阅读器和悬停提示的完整语义。
- 导航容器加入 `min-width: 0`、横向滚动和不换行布局；品牌文字在小屏折叠为图标，语言、登录和退出操作保留。
- 仅调整会员端呈现，不改变 `/`、`/media`、`/plans` 等根路由，不改变 `/admin/**` 管理导航和任何权限边界。

### 验证与边界

- 新增导航契约测试，先观察到短标签、可访问名称和窄屏布局断言 RED；实现后会员导航专项测试 `2/2` 通过。
- 53083 浏览器已确认登录态窄屏布局：第一行保留品牌、账号、语言和退出，第二行完整显示会员入口；本切片不涉及数据库、API、迁移或后端重启。

## 二十七、2026-09-24 开发进度：M3 签名 URL、防盗链与访问记录

### 已实现

- 新增 `links` Service，复用现有媒体与存储边界，使用 `APP_KEY` 生成应用层 HMAC-SHA256 签名 URL；签名绑定媒体 ID、Variant 和 Unix 过期时间，过期范围限制为 60 秒至 24 小时，公开地址统一为 `/i/{id}`，不返回对象存储地址。
- 新增会员接口：`POST /api/v1/media/{id}/signed-url`、`GET/PUT /api/v1/media/{id}/hotlink-policy`、`GET/POST /api/v1/hotlink-domains`、`DELETE /api/v1/hotlink-domains/{id}`。所有写入和读取均由认证用户推导媒体/域名归属，不接受客户端 `user_id`。
- 防盗链模式实现为 `off`、`referer`、`signed`、`hybrid`；Referer 主机归一化为小写 host（保留端口），支持无 Referer 开关。`/s/{token}` 分享访问也走同一策略判定，不能绕过 `signed` 模式。
- 新增 `media_hotlink_policies`、`hotlink_domains`、`media_access_logs` 迁移草案与模型。允许/拒绝、过期、签名错误、策略拒绝和存储失败都会尽力写入最小访问记录；记录失败不阻断媒体响应，不记录签名、密码或完整 Token。
- 会员媒体详情增加签名 URL 生成、过期时间、策略选择、无 Referer 开关和 Referer 域名维护入口；中英文文案通过现有 `member` namespace 同步维护。
- OpenAPI 和开发者 API 文档同步补齐路径、Schema、错误边界和 Local Provider 限制。

### 验证与边界

- 先观察到签名策略单元测试和 OpenAPI/会员页面契约测试 RED；实现后 `go test ./... -count=1` 全部 Go 包通过，前端全量 Node 合约测试通过 `54/54`，`vue-tsc -b` 通过。
- 迁移 `20260924000005_create_link_security_tables` 已生成并注册但未执行；当前运行后端未重启，未进行真实 PostgreSQL 策略写入、签名图片访问、Referer 拒绝或访问日志查询验收。
- 当前 Local Provider 的原生 `CreateSignedURL` 仍不支持，因此本切片使用应用层 HMAC；CDN 原生签名、缓存失效、访问记录后台报表、带宽计量和清理任务仍未完成，不能将本切片描述为生产级 CDN 防盗链。

## 二十八、2026-09-24 开发进度：M3 数据库迁移、真实访问验收与管理员访问记录

### 本轮完成

- 在用户已授权的本地 `fastimg_dev` PostgreSQL 库执行并确认 `20260924000001` 至 `20260924000005` 五项待迁移；未修改 Laragon PostgreSQL/Redis 配置，也未操作其他数据库。
- 新增 all-scope `GET /api/v1/admin/media-access-logs`，要求 `admin.media_access_logs.view`，支持 `media_id`、`variant`、`delivery_mode`、`result`、`referer_host` 和分页筛选；响应不包含签名、密码或完整 Token。管理员页面为 `/admin/media-access-logs`，复用 AdminShell、RBAC、API Client 与双语 locale。
- 修复空 `SUM()` 聚合扫描 `NULL` 导致 Free 管理员上传 500 的问题，新增 `sql.NullInt64` 回归测试；同时修复策略表显式写入、默认 `off` 读取和撤销域名幂等恢复。
- 只停止并重启已确认属于本仓库的后端进程；当前后端 `127.0.0.1:53084`，前端 `0.0.0.0:53083`，前端代理通过进程环境指向 `http://127.0.0.1:53084`，未停止其他项目进程。

### 真实验收结果

- Windows 侧 `curl.exe`：会员首页 `200`、公开套餐 `200`、未认证签名 URL `401`、OpenAPI 包含 `/admin/media-access-logs`；管理员访问记录接口返回 `{data, meta}`。
- 管理员使用本地 Free 套餐上传 `admin/src/assets/hero.png`，媒体 ID `26` 返回 `ready`。
- 应用层签名 URL 投递 `200`；`referer` 模式无 Referer/未知域名均 `404`，白名单域名 `200`；`signed` 模式有效签名 `200`。验收结束恢复媒体策略 `off` 并撤销测试域名；管理员查询显示 `16` 条测试记录（`allowed=10`、`hotlink_denied=6`）。
- WSL `wslnet ctx/url` 在本机返回 `Wsl/Service/*/E_ACCESSDENIED`，本轮以 Windows 侧 `curl.exe` 和 `netstat` 验证可达性；该限制不等同于业务接口失败。

### 自动化验证与剩余项

- 后端全量 `go test ./... -count=1`、前端全量 Node 合约测试 `56/56`、`vue-tsc -b` 均通过；新增管理员访问页 OpenAPI/路由/脱敏/i18n 测试先 RED 后 GREEN。
- `npm run build` 的 Windows shim 找不到 `vue-tsc`；直接 Node 入口类型检查通过，但 Vite 生产构建被工作树已有的 `@floating-ui/vue -> vue-demi` 缺失依赖阻断，未修改依赖清单或锁文件。
- M3 仍不是生产 CDN 防盗链：Local Provider 原生签名、CDN 缓存失效、带宽入账、访问聚合统计、访问日志清理和大规模日志分区尚未完成。

### 29. 2026-09-24：会员上传成功后的多格式链接展示

- 上传接口原有的 `links` 响应继续作为唯一数据来源；会员端 `useMemberUpload` 同时保留立即 `ready` 和轮询 `ready` 的 `url`、`markdown`、`html`、`bbcode`。
- `/` 和 `/media` 的共享上传队列在成功项下直接展示四种链接及独立复制按钮，复制成功有本地化反馈；游客仍只能看到登录提示，不能触发上传。
- 本切片没有新增 API、迁移、权限或链接拼接逻辑；这些链接当前是认证内容端点格式，公开分享/密码分享/签名 URL 仍由媒体详情与 M3 策略控制。
- RED/GREEN：新增 `tests/member-upload-links.test.mjs`，先验证 3 项失败再实现；前端全量 Node 合约测试 `59/59`，`node node_modules/vue-tsc/bin/vue-tsc.js -b` 通过。生产 Vite 构建仍受工作树已有 `vue-demi` 缺失依赖影响，未修改依赖文件。

## 三十、2026-09-24：M6 会员结算与订单中心

### 本轮完成

- 会员端保持与管理端分离：`/plans` 公开套餐目录，认证会员才可创建订单；`/checkout/:orderId`、`/orders`、`/orders/:id` 只读自己的订单并复用 `/api/v1/orders` own-scope API，导航新增短标签“订单”。
- 套餐页只使用服务端返回的 `plan_prices` 活动价格，前端不提交金额；Free 套餐保留免费可用分支，Creator/Pro 由 Seeder 预置月付/年付 CNY 版本化价格。支付成功前不切换订阅、不发放权益。
- 中英文 `member` locale 同步补齐结算、订单、支付中、空状态和导航文案；会员页面复用现有 shadcn-vue Card、Alert、Skeleton、Button，不重复建设框架组件。
- 保持管理员范围不变：后台订单、支付流水、Webhook 和退款查询在 `/admin/**`，由对应 RBAC 权限控制；会员页面不会展示全站订单或管理员财务数据。

### 验证与边界

- `node admin/tests/billing-ui.test.mjs` 和中英文 JSON 解析通过；后端 `go test ./database/seeders ./app/services/billing ./app/modules/plans/controllers -count=1` 通过。
- `pnpm exec vue-tsc -b` 在当前 Windows pnpm 环境仅输出锁文件检查后长时间无结果，已终止；未修改依赖或锁文件。后续用直接 Node 类型检查和 Vite 构建复核。
- 迁移尚未执行到 `fastimg_dev`，因此当前运行后端不会看到 `plan_prices` 价格，会员页会安全地显示无可购买的付费价格；Fake Provider 仍缺少面向会员的“模拟成功回调”入口，当前支付按钮创建 pending 支付意图，不宣称支付完成。
- 下一切片优先补 Fake/离线验收成功路径和管理端退款/履约操作，再按计划进入 Xcash、NOWPayments、PayPal Provider，真实渠道默认关闭。

## 三十一、2026-09-24：M7 Xcash Provider 离线适配

### 本轮完成

- 新增 Xcash Provider 配置 `backend/config/payment.go` 和 `XCASH_*` 环境变量；注册逻辑位于统一 billing registry，只有 `XCASH_ENABLED=true` 且必要密钥存在时注册，默认不影响 Free/Fake。
- 按 Xcash 官方协议实现 `XC-Appid`、`XC-Timestamp`、`XC-Nonce`、`XC-Signature` HMAC-SHA256；Webhook 校验 AppID、原始 body 签名和 5 分钟时间窗，拒绝缺失/过期/伪造事件。
- 实现 `/v1/invoice` hosted invoice 创建、公开状态查询和统一 Provider 状态映射。Webhook 没有法币金额时先查询账单，禁止把链上 `pay_amount` 直接当成订单金额。
- `completed/confirmed` 只有在金额、币种、Provider ID 通过通用 WebhookService 校验后才会进入支付成功；少付、多付、高风险、错误网络进入 `pending_review`/`failed`，不自动履约。Xcash 自动退款明确保持未实现，进入人工处理边界。
- 新增 `docs/fastimg-payment-provider-xcash.md`，记录配置、签名、状态、测试和真实渠道启用门禁。

### 验证与边界

- Xcash HTTP stub、签名向量、时间窗、禁用 Provider 无网络调用、账单创建/查询/Webhook 回填金额测试通过：`go test ./app/services/billing/providers/xcash ./app/services/billing ./app/modules/billing/controllers -count=1`。
- 本轮未请求真实 Xcash 网络，未写入支付密钥，未执行数据库迁移或重启后端；不能将 Xcash 描述为沙盒或生产收款已验证。
- 下一步按 M7 顺序实现 NOWPayments，再实现 PayPal；每个 Provider 保持独立配置、签名/回调测试和默认关闭状态。

## 三十二、2026-09-24：M7 NOWPayments Provider 离线适配

### 本轮完成

- 新增 `NOWPAYMENTS_*` 配置和统一 registry 注册；默认关闭，API Key/IPN Secret 只留在服务端配置边界。
- 实现 hosted invoice 创建、`GET /v1/payment/{id}` 查询、`x-api-key` 请求认证，以及 `x-nowpayments-sig` HMAC-SHA512 IPN 验签。IPN JSON 递归排序 Key 后再计算签名，避免依赖回调字段顺序。
- 状态映射只把 `finished` 作为成功候选；`waiting`、`confirming`、`confirmed`、`sending` 保持 pending，`partially_paid` 进入 pending review，失败/过期/退款不履约。
- 新增 `docs/fastimg-payment-provider-nowpayments.md`，明确 Payment API/Invoice API、密钥、网络/资产、少付、重复回调和沙盒门禁。

### 验证与边界

- NOWPayments HTTP stub、创建/查询、IPN canonical JSON 验签和 partial payment 测试通过：`go test ./app/services/billing/providers/nowpayments -count=1`；Xcash 与统一 billing 回归仍通过。
- 本轮没有访问真实 NOWPayments 网络，没有配置真实密钥，没有执行迁移或重启后端；不能将该 Provider 描述为沙盒/生产已验证。
- 下一步实现 PayPal Orders v2/capture/webhook 验签，再进入 Fake 成功回调、迁移和本地运行态支付链路验收。
