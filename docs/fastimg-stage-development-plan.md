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
5. 图片异步执行原图规范化；显示尺寸由接收站 CSS 或自身处理链控制，不为每张图片额外生成缩略图、中图、WebP 或 AVIF。
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

- 已实现：Local Provider、图片内容/扩展名/MIME/大小/像素/动画帧校验、重新编码清除源元数据、唯一原图对象写入、上传会话与幂等键、用户/套餐/配额检查、对象写入校验、私有媒体列表/预览/回收站/恢复，以及 `GET /api/v1/uploads/{id}` 状态查询和 `POST /api/v1/uploads/{id}/retry` 用户恢复入口。
- 恢复规则：processing 会话的 `updated_at` 已超过 10 分钟才可被恢复请求接管；接管会在用户行锁内刷新更新时间。逐个核验登记对象的 MIME、大小和 SHA-256，全部匹配才确认用量并标 ready；对象缺失/损坏则删除可清理对象、释放预占并标记 failed，要求用户重新上传。
- 已实现后台资源：plans、广告位、文件夹、相册进入后台资源导航；文件夹/相册启用 `user_id` own scope，广告位使用后台权限。Generated Resource Engine、RBAC 与审计复用框架，不复制实现。
- 已实现套餐快照：新订阅在现有快照列写入套餐条款、权益和 SHA-256 内容版本；配额读取兼容新 envelope 与旧纯权益 JSON，订阅 API 优先显示快照并校验版本。旧订阅缺失历史价格/名称，只能兼容显示当前套餐并返回 `snapshot_available: false`。该变更未新增或运行数据库迁移。
- 已实现用量摘要/上传入账切片：`GET /api/v1/me/usage` 按认证用户读取套餐快照额度；存储汇总为生命周期余额，API/上传/流量/处理等按 UTC 月周期汇总。上传完成在同一事务内通过 quota Service 记录 SHA-256 幂等存储流水、月度 `upload` 次数流水和 `transform` 流水，并兼容读取旧 `storage_bytes` 名称；一次成功的媒体处理作业（生成原图、缩略图和中图）计为一次 transform，处理中的上传会话预占一次，失败会话不计数。月度上传和处理流水按上传会话创建时所属 UTC 月记账，与并发预占统计周期一致，缺失创建时间的历史会话回退到完成时间。上传在图像解码前进行额度预检，拒绝已知超额请求；`BeginUpload` 在用户行锁事务中复核并原子预占实际 Variant 字节，避免并发请求仅凭预检越额。`GET /api/v1/me/usage/ledger` 仅按认证用户范围分页返回流水，不暴露幂等键。上传配额检查在用户行锁事务内为历史缺少活动订阅的用户补建 Free 快照；原因是 Goravel `Query.First` 对 0 行返回 nil，不能用来判定不存在，需使用 `Exists()`。每日上传限制已接入；月度 API 上传限制按 UTC 月统计本人 `ready` 与 `processing` 上传会话，失败会话不计数，用户行锁串行化并发预占；月度转换限制合计已完成 transform 流水和本人 `processing` 会话，同样串行化预占。额度耗尽分别返回 HTTP 429 / `MONTHLY_API_UPLOAD_LIMIT_REACHED` 或 `MONTHLY_TRANSFORM_LIMIT_REACHED`。本轮补入所有者专属永久删除 API：回收站图片先进入 `cleanup_pending`，Local/未来 Provider 删除所有唯一 Variant 对象成功后才在同一事务标记对象、Variant、媒体 tombstone 并写负向 storage ledger；删除失败可重试且不提前释放额度。带宽流水与限制已在后续 M3 切片完成；软删除不得提前释放。
- 多语言约束：`docs/i18n.md` 是各阶段强制开发与验收规则；媒体库页面和后台导航使用 vue-i18n 的 `media.*` Key，注册中英文 JSON 语言包。i18n 专项测试覆盖语言文件路径/嵌套 Key 及页面/导航 Key 使用。
- 当前前端形态：会员端使用 `/`、`/media`、`/plans` 和 `MemberShell`；管理端页面统一在 `/admin/**` 和 `AdminShell`。`/admin/media` 已提供跨用户媒体管理 API/RBAC/审计支撑；文件夹/相册仍是 own-scope，不能替代管理员媒体运营入口。
- 会员首页 `/` 提供上传入口、最近本人媒体及套餐/存储摘要；`/media` 调用认证后的 `/api/v1/uploads` 与 own-scope `/api/v1/media`，首页和媒体页共用上传队列、处理状态轮询、配额/大小错误与重试反馈。本人媒体页支持受保护缩略图预览、文件名搜索、服务端分页（48 条/页）、软删除、回收站恢复和带二次确认的清空回收站；搜索/回收站切换回到第一页，删除当前页最后一张后自动回退到有效页。会员上传结果和媒体详情现在返回带 `APP_URL` 域名的稳定公开 URL、Markdown、HTML、BBCode 以及 Variant 地址；`/tokens` 已提供 Personal API Token 的创建、一次性展示、删除和轮换界面，后端已加入 hash、Scope、会员归属删除和管理员独立删除权限。订阅及用量仍从公开套餐目录、当前认证用户订阅和 `/me/usage` 读取，流量按真实返回字节计量，不伪报 0，不下单、不升级、不更改订阅。分页、链接和 Token 文案通过 `member` namespace 同步中英文；会员前端不传 `user_id`，所有权由服务端认证身份决定。会员端已接入文件夹归档和相册批量关联基础操作；防盗链、详情编辑、完整批量管理、订单中心和 `/discover` 尚未完整实现；Token 删除迁移已应用、开发进程已重启，API 上传/链接/删除真实验收仍待独立执行。
- 本轮验证：`go test ./... -count=1` 全部 Go 包通过（含 Feature）；`pnpm run build` 通过；`node --test tests/fastimg-i18n.test.mjs` 4/4 通过。完整 admin Node 测试为 23/24，唯一失败仍是 `tests/resource-actions.test.ts`：测试要求过滤未知 action kind，但通用实现返回 `unknown`；该文件与对应实现不属于本轮改动。
- 会员媒体页本轮专项 RED/GREEN 与 i18n/member-plan 回归测试在 Windows Node fallback 中通过 15/15。构建尚未验证：WSL 初始化返回 `E_ACCESSDENIED`；Windows pnpm 尝试在线解析/重装 pnpm 后因无交互终端中止，未能进入 TypeScript/Vite 构建。现有 Vite 53081 仍可服务旧应用，但其 `/src/router/index.ts` 和 `MemberShell` 热服务内容不含新路由/导航键，故浏览器验收未通过；未杀停工作中的服务，等 WSL 可用后只重启本项目 dev server 再验收。没有提交真实图片上传、执行迁移或变更数据库/服务配置。
- `fastimg_dev` 的只读迁移状态确认 plans、media、advertising、folders、albums 迁移均为 Ran；本轮明确授权应用的仅为广告位、文件夹、相册及其管理权限。没有修改 Laragon PostgreSQL/Redis 配置。Windows 侧 5432/6379 可连接，但 WSL 内启动的后端命令仍尝试访问 `127.0.0.1:6379` 并出现连接拒绝警告；Redis 的 WSL 访问路径和队列运行验收仍未完成。
- 本地预览（2026-09-23 20:29 更新）：Vite `53081`（PID 896）继续代理到 Windows API `53082`；后端从更新后的源码重新编译并只重启仓库自有进程（旧 PID 11404 已核实为 `fastimg-dev-next.exe`，新 PID 28316 为 `fastimg-dev-next-updated.exe`）。当前 WSL NAT 网关为 `172.29.160.1`，仅通过进程环境绑定，未写入 `.env`。`wslnet url 53081` 验证 WSL/Windows 两侧 HTTP 200；Windows 直连后端网关和经 Vite 同源请求 `/api/v1/auth/me` 均返回预期未认证 401，证明 API 和代理路径可达。Windows PostgreSQL/Redis 端口 5432/6379 可连接；WSL Feature bootstrap 仍会打印其自身 `127.0.0.1:6379` 连接拒绝告警，但本轮全量 Go 测试通过。没有执行迁移或真实图片上传；浏览器真实上传仍未验收。日志和 PID 文件位于 `backend/storage/logs/fastimg-dev-20260923-202958.log`、`backend/storage/bin/fastimg-dev-20260923-202958.pid`。
- 本地预览纠正（2026-09-24）：对 `53081/admin/plans` 白屏排查后确认，WSL listener PID 759 的 cwd 正是本 checkout 的 `admin/`，不是其他项目；根因是 Vite 进程长时间运行，内存中 AdminShell 转换模块落后于磁盘源码（源码含“前往用户端”，服务响应不含）。仅重启已核实归属的 PID 759，保持端口 `53081`，现 PID 880；日志/PID 为 `/tmp/fastimg-vite-53081-20260924.log` 与 `/tmp/fastimg-vite-53081-20260924.pid`。`wslnet url 53081` 验证 WSL 和 Windows 均 HTTP 200，`/api/v1/auth/me` 未登录返回预期 401。Codex 浏览器已验证 `/admin/plans` 展示套餐管理表格及“前往用户端”链接，会员 `/plans` 展示套餐/用量及管理员可见的“管理后台”入口；无登录态的独立浏览器会正常显示登录表单。白屏已通过重启本项目 Vite 消除；未碰其他项目、数据库、迁移或 Laragon 配置。生产构建仍未验证（Windows 缺少 Rolldown 原生可选绑定；本轮使用 WSL Vite dev server 验收）。

- 会员媒体详情切片（2026-09-24）：新增认证后 `GET /api/v1/media/{id}`，Service/Repository 先按当前用户验证归属，再返回本人 ready 媒体元数据与已就绪的原图/缩略图/中图；越权与不存在统一返回 `MEDIA_NOT_FOUND`。会员端新增 `/media/:id` 只读详情页、从媒体卡片进入的入口和 zh-CN/en-US 文案；图片预览继续经认证内容端点读取，不生成公开或稳定链接。同步 OpenAPI 契约及 Service/API 路由/i18n 测试。验证：Go 媒体 Service 与 OpenAPI 测试通过，前端媒体/路由/i18n 专项 15/15 通过。生产构建未能启动：Windows pnpm 检查试图移除 modules 并要求交互确认，本轮未确认、未更改依赖或锁文件；Vue 类型检查/构建仍待可用工具环境验证。未执行迁移、数据库写入或重启开发服务。
- 会员文件夹/相册基础切片（2026-09-24）：新增认证后的 `/api/v1/me/folders` 与 `/api/v1/me/albums` 列表、创建、更新、删除接口，复用业务 Service/Repository；所有写入强制使用认证用户，父文件夹和相册封面媒体必须属于当前用户，名称限制为非空且不超过 120 个字符，相册访问策略限制为 `private`/`unlisted`/`public`。新增会员 `/folders`、`/albums` 页面及 MemberShell 导航，中英文文案与 OpenAPI 契约同步。当前只完成容器 CRUD，尚未新增媒体关联表，因此页面明确提示媒体整理关联将在后续切片接入；后台 `/admin/folders`、`/admin/albums` 仍是管理员资源，不与会员页面重合。验证：`go test ./... -count=1` 全部通过；前端会员集合/媒体/路由/i18n 专项 18/18 通过。未执行迁移、数据库写入或生产构建；WSL 当前返回 `E_ACCESSDENIED`，无法启动或用 `wslnet url` 做浏览器运行态验收。
- 会员文件夹/相册运行态修复（2026-09-24）：发现生成资源的旧版平铺重定向会抢占 `/folders`、`/albums`，导致会员页提示无匹配路由。已同步修复 `index.ts` 与 Vite 实际解析的 `index.js`：旧后台重定向排除这两个会员路径，管理员资源继续固定在 `/admin/folders`、`/admin/albums`。重启并确认仅本仓库的 Vite 会话后，浏览器已加载 `/folders` 和 `/albums`，导航、空状态、创建表单均正常，无新增控制台错误。当前 Vite 使用项目专属 tmux 会话 `fastimg-vite53081`、WSL PID 1552，后端为 Windows PID 20848、端口 53082；`wslnet url 53081` 验证 WSL/Windows 均 HTTP 200。`go test ./... -count=1` 与前端专项 18/18 仍通过；生产构建仍未验证，未执行迁移或数据库写入。
- 未完成：S3 Provider、真正异步处理/恢复队列、分片上传、完整批量整理、管理员跨用户媒体运营增强工作台、Token 迁移与生产客户端矩阵、WebP/AVIF、自动清理和后台失败任务中心。
- 重要限制：数据库最终确认失败会保留 `202 processing` 会话/对象；10 分钟后可由用户手动调用恢复接口，但自动队列扫描、失败任务后台、管理员重试与定期清理仍未实现。新增的永久删除由所有者同步触发，并支持同请求重试；它不等于 Task 7 的异步回收站/孤儿对象调度。租约阈值针对当前 10 MB 同步 Local 上传；扩展到远程或长时间 Provider 前必须改为可续期 lease/Job ownership。

## 六、M3：稳定链接、分享与防盗链

### 功能计划

1. 生成原图、缩略图、中图、WebP、AVIF、纯 URL、Markdown、HTML、BBCode。
2. 支持 private、link、public 三种可见性。
3. 支持密码分享、过期分享、撤销分享和访问记录。
4. 公开发现页自动展示 ready、public 且未被拒绝的媒体；普通上传不需要等待管理员审核即可通过稳定链接访问并进入发现页。
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
3. 当前实现固定为：`upload:write`、`media:read`、`media:delete`；不再提供可选 Scope 表单。
4. 订单、用量、文件夹、相册、分享、防盗链、批量/分片上传和 Webhook 均不是 Personal API Token 的开放面。
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

1. 发现页后台支持启用/停用；举报和后台治理不依赖会员投稿。
2. 图片、用户、域名支持举报。
3. 举报去重、阈值复核、人工审核和申诉。
4. 管理员可以隐藏、删除、恢复违规图片。
5. 管理员可以限制上传、暂停或封禁账户。
6. 所有处罚、恢复、额度调整和永久删除都写入审计。
7. 图片处理增加 SVG 清洗、恶意文件、超大像素、压缩炸弹和 SSRF 防护。
8. 公开内容先进入发现页，再通过举报、隐藏、拒绝和恢复完成事后治理。

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

- 上传创建、状态查询、重试和会员媒体详情统一返回唯一 `original`、`url`、`markdown`、`html`、`bbcode`。
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
- 新增会话认证的 `POST /api/v1/tokens`、`GET /api/v1/tokens`、`DELETE /api/v1/tokens/{id}`、`POST /api/v1/tokens/{id}/rotate`；新 Token 固定为 `upload:write`、`media:read`、`media:delete`。
- 会员资源认证中间件支持登录会话或 `Authorization: Bearer fst_...` Personal API Token；管理员 `/api/v1/admin/**` 仍只接受框架会话认证，Token 不能进入后台 RBAC。
- 会员 `/tokens` 页面支持创建、一次性复制、撤销、轮换和中英文 Scope 展示；不把完整 Token 写入 URL、LocalStorage 或日志。
- `go test ./... -count=1` 通过；前端全量 `.mjs` 合约测试通过 39/39，包含 i18n、路由、Token 页面、媒体、集合和计划边界。
- `fastimg_dev` 已执行 Token/媒体相关迁移并重启当前开发后端；代码级验收覆盖 Token 创建 Scope 固定、own-scope 路由和会话专属路由拒绝。生产环境还需要按用户/IP/Token 限流、审计脱敏和真实客户端矩阵验收。

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

## 三十三、2026-09-24：M7 PayPal Provider 离线适配

### 本轮完成

- 新增 `PAYPAL_*` 配置和统一 registry 注册；支持 sandbox/live base URL 配置，默认关闭，OAuth Client ID/Secret/Webhook ID 只在服务端使用。
- 实现 OAuth client-credentials token 缓存、Orders v2 `intent=CAPTURE` 创建、PayPal approval URL、订单/捕获查询、capture refund 请求和 `PayPal-Request-Id` 幂等头。
- Webhook 通过 PayPal Verification API 校验 transmission headers、Webhook ID 和原始事件；只将捕获完成事件作为成功候选，拒绝未验签或状态不确定事件。
- 新增 `docs/fastimg-payment-provider-paypal.md`，明确沙盒配置、状态、退款和真实启用门禁。

### 验证与边界

- PayPal OAuth、Orders v2 stub、查询捕获、Webhook Verification stub 测试通过：`go test ./app/services/billing/providers/paypal ./app/services/billing ./app/modules/billing/controllers -count=1`；Xcash/NOWPayments 回归仍通过。
- 本轮没有访问真实 PayPal 网络，没有配置真实 Client Secret，没有执行迁移或重启后端；不能将 PayPal 描述为 Sandbox/生产已验证。
- M7 代码适配已完成，下一阶段进入 M8 运维能力和 Fake 成功回调/本地迁移验收；真实渠道只有在各自沙盒证据完成后才能启用。

## 三十四、2026-09-24：Fake 支付成功与履约验收入口

- 开发环境新增认证会员接口 `POST /api/v1/orders/{id}/payments/fake/succeed`，只允许当前用户自己的订单，生产环境固定返回不可用；前端结算页在 Fake 支付意图创建后显示“确认测试支付”。
- 服务端通过 Fake Provider 的状态转换生成内部事件，再进入统一 `IngestVerified` 入口，复用 Webhook 事件去重、订单金额/币种校验、支付流水和履约任务，不增加未验签生产旁路。
- 会员看到订单由 `pending_payment` 进入 `paid/fulfilled` 的状态变化；管理员仍通过 `/admin/orders` 和履约权限处理异常任务。真实 Xcash/NOWPayments/PayPal 不受该开发入口影响。
- 验证：会员 UI 静态契约、zh-CN/en-US JSON、`node node_modules/vue-tsc/bin/vue-tsc.js -b --pretty false`、支付 Provider 和 billing Go 测试通过；尚未在 `fastimg_dev` 执行迁移和浏览器运行态验收。

## 三十五、2026-09-24：fastimg_dev 本地迁移与运行态准备

- 已先通过 `go run . artisan migrate:status --no-ansi` 检查 `fastimg_dev`，仅 `20260924000006_create_billing_tables` 为 Pending；随后执行 `go run . artisan migrate --no-ansi`，仅该 billing migration 成功应用。
- 已执行 `go run . artisan db:seed --no-ansi`；`db:table plan_prices`、`db:table orders` 和再次查询迁移状态确认支付表、价格目录和 migration batch `[7] Ran`。
- 后端已由本项目独立进程运行在 `http://127.0.0.1:53082`，`GET /api/v1/plans` 返回 Free、Creator、Pro 及月付/年付 CNY 价格；支付渠道仍保持默认关闭。
- 常规 Vite 运行被当前共享 `node_modules` 缺少 Windows `@tailwindcss/oxide-win32-x64-msvc` 原生绑定阻断；没有修改依赖或锁文件。开发阶段临时使用静态 `dist` 预览和 `/api` 代理运行在 `http://127.0.0.1:5181`，该预览不替代源码构建，待恢复同版本 Windows 依赖后再做源码热更新验收。

## 三十六、2026-09-24：M8 管理端运营闭环与运行态修复

### 本轮完成

- 广告位的 `starts_at` / `ends_at` 使用管理端 shadcn-vue 日期弹层、日历和时间输入组合，不再让运营人员直接填写 RFC3339；保存时仍由资源表单统一转换为 API 时间格式。
- 套餐的 `entitlements_json` 使用面向运营人员的权益表单，覆盖存储、单文件、上传、API、带宽、图片处理、Token 和广告开关；同时保留 JSON 作为后端持久化格式。结算内部使用 `plan_prices` 维护月/年周期、币种、金额、价格版本和生效区间；价格不再绑定单一网关。开发阶段不注册独立的 `/admin/plan-prices` 页面，也不创建 `admin.plan_prices.*` 权限；旧地址不提供兼容入口，价格目录由 Seeder/结算 Service 提供。
- 订单、支付流水、支付事件、退款记录在管理端注册为独立资源菜单；订单创建时保存价格快照，支付意图保存会员本次选择的 Provider。会员端仍只能查看自己的订单和支付状态。
- 新增系统设置页面 `/admin/settings`：SEO、网关开关、统计、站点验证/广告验证/自定义代码、站点基础设置；支付密钥不进入数据库表单，只从后端环境变量读取。新增统计页面 `/admin/statistics`，展示用户、媒体、相册、文件夹、订单和支付流水汇总。
- 后台补齐跨用户媒体库 `/admin/media`、相册和文件夹资源注册；上传写入的媒体记录可以在后台按用户查看。媒体访问记录 `/admin/media-access-logs` 修复为可读页面，并验证可显示访问条数、类型、签名和放行状态。
- 注册遗漏的系统设置迁移，并修复 PostgreSQL 保留字 `group` 导致设置列表 500 的排序问题。开发环境前端 53083 通过明确的开发 API 基址访问本项目后端 53085，避免误连其他项目；不修改共享 `.env` 或 Laragon 服务配置。
- 所有新增管理资源、设置页面和权益表单补齐中英文 locale；会员端和管理端的职责、权限、own-scope/all-scope 边界保持分离。

### 验证与边界

- `go test ./app/services/billing ./app/modules/billing/controllers ./app/modules/admin/registry ./database/seeders ./database/migrations -count=1` 通过。
- `node node_modules/vue-tsc/bin/vue-tsc.js -b --pretty false` 通过；前端专项测试 `9/9` 通过，包含资源表单、广告、支付管理、媒体访问记录和管理页面契约。
- `fastimg_dev` 已应用系统设置和计划价格迁移并重新执行 seed；运行态确认 `/api/v1/admin/orders`、`/api/v1/admin/payment-transactions`、`/api/v1/admin/media`、`/api/v1/admin/media-access-logs`、`/api/v1/admin/settings` 返回 200。
- Vite 生产构建仍受共享 `node_modules` 缺少 `@tailwindcss/oxide-win32-x64-msvc` 原生可选绑定阻断；没有改动依赖或锁文件。真实 PayPal、XCash、NOWPayments 仍默认关闭，未宣称外部沙盒或生产收款已验证。

## 三十七、2026-09-24：M8 Personal API 最小权限、支付渠道选择与 SEO

### 本轮完成

- Personal API Token 固定为 `upload:write`、`media:read`、`media:delete`，移除会员端可选 Scope 表单；Token 只允许上传、读取本人图片列表/信息/链接和删除本人图片。订单、支付、用量、文件夹、相册、分享、防盗链、Webhook、批量上传、重试和管理员 API 均要求会员会话或管理员 RBAC。
- 增加 `X-API-Key` 认证，并提供 `/api/upload`、`/api/images`、`/api/image/{id}` 兼容别名；别名复用现有上传/媒体 Service，不形成第二套 API 业务逻辑。
- `plan` 管理表单不再编辑旧的价格字段；计划只维护权益和产品信息，`plan_prices` 维护版本化金额/币种/周期。价格不再强制绑定单个网关，会员结算页读取当前已注册 Provider 并选择支付渠道。
- 支付设置补充 PayPal、XCash、NOWPayments 的环境、API 地址、回调/跳转地址等非敏感配置；密钥仍只从服务端环境变量读取。启用开关会参与 Provider 注册，多个渠道可以同时启用。
- 后端新增 `/sitemap.xml`、`/robots.txt`；会员首页、套餐页和发现页动态设置 SEO meta；前端 `build:ssg` 生成已注册公开首页、套餐页和发现页的静态 SEO 首屏，公开相册需等真实公开路由、访问策略和审核接口完成后加入，私有内容不会进入静态 HTML。

### 验证与边界

- 路由和 Scope 代码已完成，并已在本地运行态通过 Personal Token 请求允许接口与被拒绝接口，验收 `200/201` 与 `403 TOKEN_ENDPOINT_NOT_ALLOWED`。
- `plan_prices` 已移除 `gateway_code` 列；价格与支付渠道完全解耦，历史订单仍使用自己的 PaymentIntent Provider，不回写旧价格。
- SEO 设置包含站点地图启用开关和额外公开路径；`/sitemap.xml` 只接受公开路径，自动拒绝 `/admin`、`/api`、查询串和锚点，`robots.txt` 与开关同步。
- 当前真实渠道仍需各自服务端环境变量、Webhook 配置和沙盒验证；默认只有 Fake Provider 可用，不宣称真实收款已完成。

## 三十八、2026-09-25：M4 Personal API Token 运行态闭环验收

### 本轮完成

- 使用 `fastimg_dev` 的开发账号创建一次 Personal API Token；服务端忽略客户端提交的非法 Scope，固定授予 `upload:write`、`media:read`、`media:delete`。
- 使用 `Authorization: Bearer fst_...` 验证本人媒体列表和兼容列表接口；使用 `X-API-Key: fst_...` 验证单文件 multipart 上传、本人媒体详情、链接返回和软删除。
- 上传 `admin/src/assets/hero.png` 返回 `201`、`ready` 和 `original`/Markdown/HTML/BBCode 等链接；详情返回 200，删除后返回 `deleted`。
- 使用同一 Token 请求 `/api/v1/me/usage` 和 `/api/v1/orders` 均返回 `403 TOKEN_ENDPOINT_NOT_ALLOWED`，确认订单和用量没有被 Personal API Token 放开。
- 验收 Token 在完成测试后立即撤销，明文 Token 未保留在数据库、仓库或运行日志中。

### 验证与边界

- 运行态后端：`http://127.0.0.1:53085`；Personal API Token 允许面返回 200/201，受限面返回 403。
- 应用级 Token/IP 限流已实现并在本地运行态验证 429、`Retry-After` 和限流响应头；生产压测、客户端矩阵测试和真实 Provider 沙盒验收仍属于 M4/M7/M8 后续门禁，不能以本地闭环替代。

## 三十九、2026-09-25：M4 Personal API 统一错误契约与请求追踪

### 本轮完成

- 新增全局请求上下文中间件：统一处理安全的 `X-Request-ID`，缺失或非法时生成 `req_...`，并始终通过响应头返回最终请求 ID。
- Personal API 的 Token 认证、上传、媒体列表/详情/内容/删除/恢复/归档错误统一返回 `code`、`request_id`、`retryable`；限流响应额外保留 `retry_after_seconds` 和 `Retry-After`。
- `retryable` 由 HTTP 状态统一推导：429、500、502、503、504 为可退避重试，其余认证、权限、参数、资源和业务冲突错误为不可盲目重试。
- 保持会员会话和管理员后台的职责边界；本轮没有扩大 Personal API 的接口范围，也没有把管理员 API 暴露给 Token。

### 验证与边界

- 请求 ID 规范化、非法 ID 替换和可重试状态测试通过；媒体、上传、Token 认证及开发者限流定向 Go 测试通过。
- 已重建并重启当前项目后端 `53085`，真实 HTTP 验收确认未认证媒体请求会生成并回显 `X-Request-ID`，Token 调用被拒绝时会返回带同一请求 ID 的标准错误体；临时验收 Token 已撤销。
- 生产级日志关联、跨服务 trace、负载压测、客户端兼容矩阵以及真实支付 Provider 沙盒仍未完成；本轮不宣称生产观测体系已达标。

## 四十、2026-09-25：M2 会员媒体与相册关联基础操作

### 本轮完成

- 新增 `album_media` 多对多关系表，保留 `user_id`、唯一 `(album_id, media_asset_id)`、排序字段和查询索引；文件夹仍使用 `media_assets.folder_id` 单归档，不把两种整理方式混成一张表。
- 新增会员 own-scope 接口：
  - `GET /api/v1/me/albums/{id}/media` 查询当前用户相册中的媒体 ID；
  - `POST /api/v1/me/albums/{id}/media` 批量加入媒体，最多 100 个 ID，重复加入返回 `skipped_ids`，不会产生重复关系；
  - `DELETE /api/v1/me/albums/{id}/media/{media_id}` 移除单个相册关系。
- 服务端同时校验相册归属、媒体归属、媒体 `ready` 状态和批次大小；跨用户媒体、已删除媒体、不存在相册不会被关联。Personal API Token 仍不能调用相册和整理接口。
- 相册列表返回 `media_count`；会员图片库增加“加入相册”操作，相册页显示图片数量；中英文 locale、OpenAPI 和错误码同步更新。
- 修复相册创建/更新读取 map 时使用 `First` 导致“数据已写入但接口返回不存在”的运行态问题，统一改为可可靠读取的 `Get` 结果检查。

### 验证与边界

- `fastimg_dev` 已应用 `20260925000004_create_album_media_table`；全量 `go test ./... -count=1`、前端 `vue-tsc`、路由测试 7/7、i18n 测试 4/4 通过。
- 重启 `53085` 后真实验收通过：相册创建、图片上传、重复 ID 批量关联、列表读取和移除关系均成功；临时相册和临时媒体已清理。
- 当前已实现基础关联、单图加入、相册内容网格和批量移除；拖拽排序、批量跨相册移动、公开相册 SEO 和发现页仍按后续媒体整理与 M3/M8 计划推进。

## 四十一、2026-09-25：M2 相册内容网格与批量整理

### 本轮完成

- 相册内容页新增会员路由 `/albums/:id`，从相册列表进入后加载 own-scope 媒体详情和认证缩略图，支持跳转图片详情。
- 新增 `DELETE /api/v1/me/albums/{id}/media` 批量移除接口，复用最多 100 个媒体 ID 的校验；响应使用明确的 `removed_ids`/`skipped_ids` 字段，不存在的关系进入 `skipped_ids`，不会影响同批次其他关系。
- 相册页面支持全选、部分选择和批量移除；图片库继续提供单图加入相册入口，文件夹仍保持单归档关系。
- 同步中英文 locale、路由、OpenAPI、错误处理和阶段进度；没有开放给 Personal API Token，也没有改变管理员全站媒体权限模型。

### 验证与边界

- 后端集合服务、控制器、OpenAPI 定向测试与全量 Go 测试通过；前端 `vue-tsc`、路由 7/7、i18n 4/4 通过。真实 HTTP 验收覆盖相册创建、ready 图片上传、重复 ID 加入、列表读取、批量移除、移除后空列表和临时数据清理；批量移除响应已修正为 `removed_ids`/`skipped_ids`，避免客户端误判。
- 当前后端仍运行在 `53085`，前端预览仍运行在 `53083`；生产构建继续受 Windows Tailwind 原生可选依赖限制。
- 拖拽排序、批量跨相册移动、公开相册访问/SEO、发现页和 CDN 缓存策略仍未实现。

## 四十二、2026-09-25：M2 相册排序

### 本轮完成

- 新增会员 own-scope `PATCH /api/v1/me/albums/{id}/media/order`，要求提交相册当前全部媒体 ID 的唯一顺序；缺项、越权媒体、重复 ID 或超过 100 个 ID 会拒绝，避免只更新部分关系造成顺序漂移。
- 复用 `album_media.sort_order`，在事务内锁定相册关系并按请求顺序保存 `sort_order`；查询接口继续按排序字段返回媒体 ID。
- 相册内容页支持拖拽调整顺序，也提供键盘/按钮上移、下移；只有点击“保存排序”才提交，未保存状态和失败提示均有中英文文案。
- 更新 OpenAPI、路由和 Service 测试；排序接口不开放给 Personal API Token，仍严格使用会员会话 own-scope。

### 验证与边界

- 全量 `go test ./... -count=1`、前端 `vue-tsc`、路由 7/7、i18n 4/4 通过。
- 后端重建并重启在 `53085`；真实 HTTP 验收创建临时相册并上传两张 ready 图片，确认默认顺序为 `[51,50]`，提交 `[50,51]` 后响应和再次查询均为 `[50,51]`，随后完成临时数据清理。
- 批量跨相册移动、公开相册访问/SEO、发现页和 CDN 缓存策略仍未实现。

## 四十三、2026-09-25：M3 真实发现页与公开 API

### 本轮完成

- 新增公开发现状态接口 `GET /api/v1/discovery/status`、分页接口 `GET /api/v1/discovery/feed` 和公开内容接口 `GET /api/v1/discovery/media/{id}/content?variant=original`。发现页返回 `ready`、未删除、`visibility=public` 且未被管理员拒绝的媒体；普通上传完成后即可进入发现页，空数据也返回稳定 JSON 分页结构。
- 不再提供会员主动投稿接口。发现页采用“先公开、后治理”：用户上传完成后默认 `public + approved`，用户无需等待管理员；用户或游客可以对公开图片举报，举报进入后台举报队列。
- 新增迁移 `20260925000005_add_discovery_state_to_media`，普通上传默认写入 `public + approved`，上传完成后立即可通过稳定链接访问；用户可在会员媒体详情切换 `private` 或 `link`。`discovery_submitted_at` 仅保留历史审计字段，不再作为发现页查询条件。
- 会员端新增公开 `/discover` 页面：游客可访问，使用真实发现 API 加载图片瀑布流、分页加载、图片内容链接和登录后举报入口；导航、SEO 元数据和 SSG 公开页面清单已同步中英文。
- 管理员通过 `/admin/media` 的 `visibility`、`moderation_status` 和媒体状态管理公开内容；历史 `discovery_submitted_at` 不再作为用户功能或发现门槛，用户举报继续进入 `/admin/reports`，管理员的全站媒体视角与会员 own-scope 保持分离。

### 验证与边界

- 后端目标测试、迁移测试、OpenAPI 测试通过；迁移实际应用到 `fastimg_dev`。重建并重启后端 `http://127.0.0.1:53085` 后，真实请求确认 status `200`、feed 空结果 `200`、不存在媒体内容 `404 DISCOVERY_MEDIA_NOT_FOUND`；前端 `/discover` 入口返回 `200`。
- `vue-tsc`、路由 `8/8`、i18n `4/4`、会员导航专项测试通过。完整 Go 测试的所有包（含 `goravel/tests/feature`）输出均为 `ok`，命令最后仅因磁盘空间不足清理自定义 GOCACHE 返回非零码，不能把该命令表述为干净退出。
- 本轮完成的是公开发现页，不等同于真实 SSR；当前仍是 Vue SPA + SSG 首屏。后续再实现独立 Web SSR 入口、请求级数据加载和服务端部署适配；公开相册、自动审核、申诉和独立发现审核工作台仍未完成。

## 四十四、2026-09-25：M3 上传公开链接与套餐流量闭环

### 本轮完成

- 上传完成响应和会员媒体详情的 `links` 改为使用 `APP_URL` 生成带 APP_KEY 签名的绝对公开地址，不再返回只能在登录态使用的相对 `/api/v1/media/.../content` 地址。链接包含原图、缩略图、中图以及 URL、Markdown、HTML、BBCode 格式，可直接复制到论坛、博客和 Markdown 内容。
- 稳定公开地址使用 `/i/{media_id}?variant=...&signature=...`，不保存明文分享 Token、不暴露对象存储路径；删除媒体、热链策略拒绝和 APP_KEY 轮换都会使地址失效。原有有时效签名 URL 继续兼容。
- 实际返回内容的字节数接入 `bandwidth/download` 流水：稳定链接、限时签名链接、分享链接、发现页内容和会员自己的媒体内容都会在响应前按媒体所属用户的当前 UTC 月套餐限额计量；重复访问使用幂等来源键，超额返回 `429 BANDWIDTH_QUOTA_EXCEEDED`，计量不可用返回 `503 BANDWIDTH_METERING_UNAVAILABLE`。
- `/api/v1/me/usage` 返回 `bandwidth_metered: true`，会员套餐页按服务端 `usage.bandwidth` 和 `monthly_bandwidth_bytes` 展示真实用量，不再把未接入的流量显示成零。
- 保持原有套餐限制的服务端校验：存储空间、单文件大小、每日上传数、月上传数、月处理次数、Token 数量和流量均由实际业务服务读取订阅快照并拒绝超限请求；套餐权益不是仅用于展示。

### 验证与边界

- Go 定向测试、后端构建、前端 `vue-tsc` 和上传/套餐静态契约测试通过；完整 Go 测试仍有工作树中支付网关配置测试引用未实现 `MissingGatewaySettings`/`GatewayURLs` 的独立失败，需要在支付网关切片补齐。
- 运行态后端已重建并重启在 `http://127.0.0.1:53085`；真实登录上传 `hero.png` 返回 `ready` 和带域名的稳定链接，匿名请求该链接返回 `200 image/png`，随后 `/api/v1/me/usage` 的 bandwidth 用量实际增加。
- 本地 `.env` 已固定使用后端 `53085`，避免占用其他项目的 `53082`；生产环境应将 `APP_URL` 设置为反向代理后的真实站点域名。Windows Tailwind 原生可选依赖仍缺失，因此当前不能用 Vite 生产构建刷新 `admin/dist`，源代码和后端已完成，构建依赖恢复后需重新构建前端并刷新预览。

## 四十五、2026-09-25：M3 媒体卡片紧凑化与管理员套餐豁免

### 本轮完成

- 会员 `/media` 媒体库改为紧凑单图卡片：固定预览高度、移除卡片默认上下留白、文件夹/相册选择器并排且使用紧凑控件，保留详情、回收站和归档操作；一张上传图片仍只对应一张卡片。
- 管理员身份由 RBAC 核心管理权限判定，包含 `admin.users.view`、`admin.users.manage`、`admin.roles.manage`、`admin.permissions.manage`。管理员上传、媒体处理和访问不再被订阅快照中的 Free/付费会员额度拦截，但仍保留 10 MB 请求和图片处理器等系统级技术上限，并继续写入媒体、用量和审计记录。
- 后台用户编辑页 `/admin/users/{id}/edit` 新增“套餐与额度”选择器；管理员可以为用户（包括自己）指定当前生效套餐，服务端更新订阅权益快照。会员 `/plans` 对管理员显示“管理员额度”和“不限量”，同时保留当前套餐档位展示。

### 验证与边界

- Go 定向测试：RBAC、套餐、媒体和套餐控制器通过；后端使用最新源码重建并重启在 `http://127.0.0.1:53085`。
- 前端媒体、首页、套餐、i18n 测试通过；Vite 构建和 `vue-tsc -b --pretty false` 通过；浏览器实测 `/plans` 显示管理员额度和存储/流量“不限量”，`/admin/users/1/edit` 显示套餐选择器，`/media` 卡片实际高度由约 567 px 降至约 272 px（当前预览视口下）。
- 管理员“套餐档位”与“管理员额度豁免”是两个独立概念；不会自动替当前账户切换到付费计划，也不会绕过系统级上传安全限制。

## 四十六、2026-09-25：支付网关官方地址自动生成

### 本轮完成

- `/admin/settings` 的支付网关表单按官方渠道要求收敛为凭证输入：PayPal 输入环境、Client ID、Client Secret、Webhook ID；Xcash 输入 Appid、HMAC 密钥；NOWPayments 输入 API Key、IPN Secret。
- API 根地址、Webhook/IPN 地址、成功回跳和取消回跳不再让管理员手工填写，表单按站点 URL和渠道环境生成只读值；PayPal 的 Webhook 地址用于 Developer 后台创建 Webhook，Xcash 的通知地址用于项目管理，NOWPayments 的 IPN 地址用于 Dashboard 的 IPN 设置。
- 后端 Provider 注册以官方 API、回调和回跳地址为默认值，同时读取管理员保存的自定义 URL；PayPal 缺少 Webhook ID 时也不会注册为可用渠道。
- 自动地址统一读取系统设置 `site_url`；前端管理表单与后端 Provider 注册不再使用旧的 `site.url` 键，避免管理员配置的正式域名被当前开发端口覆盖。

### 官方依据与边界

- PayPal：OAuth 2.0 Client Credentials、Orders v2、Webhook Verification API；Xcash：官方 `https://pay.xca.sh` `/v1/invoice`、Appid/HMAC 和通知地址；NOWPayments：官方 `https://api.nowpayments.io/v1/invoice`、API Key、IPN Secret 和 `ipn_callback_url`。
- 成功/取消回跳只用于把用户带回订单详情页，不能改变订单状态；权益仍由验签 Webhook/IPN 驱动。开发环境显示的 `127.0.0.1` 地址只能用于本地观察，第三方支付回调必须使用公网 HTTPS 站点 URL。

## 四十七、2026-09-25：按套餐控制图片水印

### 本轮完成

- `Entitlement` 新增 `watermark_enabled`。Free 默认关闭；FastImg 初始 Creator/Pro 种子套餐开启，管理员可以在 `/admin/plans/{id}/edit` 的权益表单中单独切换每个计划，保存后新订阅和管理员手动分配的订阅快照立即生效。
- 上传服务端在配额校验后读取当前用户的订阅快照，而不是相信前端传入的开关；开启水印时在唯一原图编码前处理 PNG/JPEG，GIF 的每一帧也处理，不存在可绕过水印的预览 Variant。
- `/admin/settings` 的“其他设置”增加 `watermark.text`、`watermark.domain` 和 `watermark.fallback_image_url`：文字默认使用 `FastImg`，域名可选追加到水印，加载失败时会员端使用配置图片或内置 FastImg 兜底图；内置处理器支持 ASCII 字母、数字和常用网址符号，不引入新的字体或图片处理框架依赖。
- 修复套餐权益编辑器的广告和水印开关：使用显式 `model-value`/更新回调同步到 `entitlements_json`，保存后由服务端规范化并在新上传时生效。新增公开 `GET /api/v1/site/presentation`，只提供经过 URL 安全校验的加载失败兜底图地址；会员首页、上传结果、媒体库、相册、详情和发现页均在图片加载失败时显示兜底水印图。
- 会员 `/plans` 展示当前公开套餐是否在上传时添加水印；旧订阅快照缺少新字段时兼容为关闭，计划写入会规范化输出 `watermark_enabled`。

### 验证与边界

- 已先写入并观察水印处理器和权益解析的失败测试，再实现代码；水印测试确认原图、缩略图和中图字节均发生变化，旧快照兼容测试通过。
- 已执行迁移 `20260925000006_add_watermark_entitlement_defaults` 到用户授权的 `fastimg_dev`：只为旧计划补入缺失字段，Free 为关闭，内置 Creator/Pro 为开启，已有明确值不覆盖；不修改订阅历史快照。后端已重建并重启在 `53085`，前端构建、`vue-tsc`、i18n 和 `/admin/plans/1/edit` 浏览器验收通过。
- 水印只在上传时生成并固化到 Variant；已存在的媒体不会被套餐切换回溯重写。后续如需“按访问时动态水印”必须另行设计缓存、原图保护和重新生成策略。
- 本轮新增测试先验证计划开关绑定、域名水印差异、公开兜底图 URL 安全过滤，再实现代码；前端构建和类型检查通过。水印设置保存后不需要数据库迁移，动态系统设置由现有设置服务加密/持久化策略管理。
- 同时补充了订阅快照兼容：水印字段加入前生成的旧版本哈希仍可读取，避免管理员编辑计划后旧订阅被错误判定为不可用；本地浏览器验收确认 Creator 水印开关可关闭、保存、刷新后恢复开启，会员 `/plans` 订阅接口保持 200。

## 四十八、2026-09-25：计划开关可见性与系统设置完整保存

### 本轮修复

- 套餐权益表单的“显示广告”和“启用图片水印”不再只显示一个无文字状态的开关：开关使用显式 `model-value`/更新回调，带有可访问名称，并同步显示“已启用/未启用”；管理员可以在 `/admin/plans/{id}/edit` 修改后点击保存，服务端更新 `entitlements_json`。
- 修复 `/admin/settings` 保存中断：网关凭证字段使用 `value_type=secret`，设置服务现在按字符串规则校验并继续使用 `APP_KEY` 加密保存；可选整数设置为空时合法，消费者按自身安全默认值处理。这样统计保留天数、上传大小等未填写字段不会阻断后续水印设置提交。
- 设置保存改为完整字段链路验收：管理员保存后必须出现“设置已保存”，后端日志中全部设置 PUT 请求返回 200；前端仍不回显任何密钥明文。

### 验证与边界

- RED/GREEN：先验证 `secret` 和空可选整数会失败，再补充设置校验测试和实现；前端专项测试 `17/17`、`vue-tsc`、Vite 生产构建均通过。
- 本地后端已重建并重启在 `http://127.0.0.1:53085`，前端静态预览使用 `http://127.0.0.1:53084`；浏览器确认设置页出现“设置已保存”，计划 1 的广告显示“已启用”、水印显示“未启用”，并可实际切换。
- 53084/53085 仅为当前开发进程端口；生产部署必须通过正式域名/HTTPS 配置 `APP_URL` 或 `site_url`，支付回调地址不能使用本地回环地址。

## 四十九、2026-09-25：举报闭环与 C 端个人举报记录

### 本轮实现

- C 端发现页继续使用 `POST /api/v1/media/{id}/reports`，服务端只接受 `ready`、公开、审核通过且不属于当前用户的媒体，避免举报私有资源、未审核资源和自举报。
- 登录用户新增 `GET /api/v1/me/reports`，只返回自己的举报记录和处理结果；新增 `/reports` 页面与会员导航入口，支持中英文状态、处理说明和时间展示。
- 管理员 `/admin/reports` 保持资源引擎统一列表，举报状态和处理说明改为只读；通过 `report-resolve` 批量动作处理：`dismiss`（驳回举报）、`hide_media`（转私有并标记违规）、`restore_media`（恢复公开并标记通过）。动作在事务内同时更新举报和媒体，并写入 `moderation.report.resolve` 审计记录。
- 举报处理使用管理员权限和资源范围校验，不向 C 端开放全量举报、举报人信息或管理员 API；C 端只看自己的举报状态。

### 验证与边界

- 后端定向测试覆盖举报动作 payload、资源动作注册、控制器和管理员动作控制器；前端举报页面、导航、双语文案、账单/计划回归测试共 `19/19` 通过，`vue-tsc` 与 Vite 构建通过。
- 后端已重建并重启在 `http://127.0.0.1:53085`，前端静态预览在 `http://127.0.0.1:53084`；未登录访问 `/api/v1/me/reports` 返回 `401 AUTH_UNAUTHORIZED`，公开站点接口返回 `200`。
- 举报记录表继续作为审核证据队列，不与媒体表、审计表或支付流水重复；下一步可补充管理员筛选统计和通知，但不改变当前 C/Admin 边界。

## 五十、2026-09-25：支付渠道协议边界、动态注册与审计可读性

### 本轮完成

- 修复支付创建失败的两类根因：XCash 适配器按官方协议使用渠道支持的字段，统一层不再强制注入渠道未声明的 `notify_url`/`return_url`；同时将后端默认入站请求超时从 3 秒调整为 30 秒，外部 Provider 仍使用自己的 10 秒出站超时，避免正常的支付创建过程被框架提前截断。
- NOWPayments 注册条件改为“启用 + API Key + API 地址”。API Key 用于创建/查询支付；IPN Secret 只用于 IPN 验签和自动履约。因此管理员只填写 API Key 后，NOWPayments 会出现在 C 端渠道列表；补齐 IPN Secret 后才具备自动回调结算能力。
- 网关注册表支持热刷新：管理员保存设置、会员读取可用渠道、支付创建和 Webhook 处理前都会重新加载当前配置，不需要重启才能看到新增/修改的渠道。
- 统一支付接口收敛为创建、查询、验签三个基础能力；退款拆为可选 `RefundGateway`，渠道未实现退款时明确返回不支持，不再要求所有适配器实现空的退款接口。
- Provider 网络和协议错误统一脱敏为可识别的 502/503 安全错误，客户端不再只得到无法定位的通用 500，也不会泄露第三方响应原文、请求体或密钥。
- 审计模块增加业务分类、操作名称和结果字段；管理员列表默认隐藏历史 `http.*` 技术噪声，展示“创建支付请求/支付回调/媒体上传/设置更新”等面向小白用户的事件和失败摘要，技术 JSON 仅在详情中展开。成功的普通 GET/HEAD 和成功会话刷新不再新增审计噪声，失败请求和重要变更仍保留。

### 验证结果

- Go 定向测试通过：XCash、NOWPayments、PayPal、billing、audit 和 billing controllers；项目专用 Go 缓存位于 `backend/.runtime`，避免继续膨胀用户级缓存。
- 后端成功重建并运行在 `http://127.0.0.1:53085`；真实 C 端订单 6 使用 XCash 创建支付成功，返回 `https://pay.xca.sh/pay/...`，接口 HTTP 201，耗时约 2 秒。NOWPayments API Key-only 配置已在 C 端渠道列表可见。
- 前端审计页面构建和类型检查通过；旧技术日志仍保留在数据库供取证，但默认管理员查询不再展示它们。完整页面验收需在后端重启后重新加载 `/admin/audit-logs`。

### 仍需真实环境验收

- NOWPayments、XCash 和 PayPal 的真实回调、重复回调、错误签名、支付状态查询、退款能力和对账仍需使用官方沙盒或小额生产交易逐渠道验收。
- 第三方回调必须使用公网 HTTPS 域名，不能使用本地 `127.0.0.1`；当前开发端口只用于本机联调。
- 本轮 `wslnet url` 由于 WSL 服务返回 `E_ACCESSDENIED` 未完成，Windows 侧 PowerShell 已确认前端 53084 和后端 53085 可访问。

## 五十一、2026-09-25：会员支付直达、失败提示与订单取消

### 本轮完成

- 会员结算页点击“立即支付”成功后立即跳转 `checkout_url`，不再先渲染一个需要二次点击的“继续支付”链接；只有 Provider 没有返回地址时才保留页面提示。
- 支付失败会读取后端安全错误码：Provider 连接超时显示“支付渠道连接超时”，Provider 拒绝显示“支付渠道拒绝请求”，并明确说明订单没有扣款；不会把外部连接问题伪装成普通订单加载失败。
- 订单详情页新增取消待支付订单入口；结算页原有取消入口继续保留。后端只允许 `created`/`pending_payment` 状态取消，并将重复取消、已支付或已过期订单映射为 `ORDER_NOT_CANCELLABLE`/409，而不是 500。

### 验证与边界

- 前端 billing UI 回归测试先验证了原实现缺少直接跳转，再通过；后端取消错误码测试先失败后通过。前端类型检查和构建、支付/审计/订单相关 Go 定向测试需在本轮最终验收中再次执行。
- 当前开发环境对 `api.nowpayments.io` 直连在 5 秒左右超时，后端 Provider 10 秒超时后返回 `PAYMENT_PROVIDER_UNAVAILABLE`；XCash 连接可用。该限制属于当前网络出口/第三方可达性，不能通过前端重试掩盖，生产需配置可达的 HTTPS 网络出口后再验收 NOWPayments。

## 五十二、2026-09-25：媒体公开/私有边界、管理员预览与举报审计

### 本轮完成

- 新上传媒体默认 `visibility=public`、`moderation_status=approved`，不阻塞用户使用、外链分享或进入发现页；公开稳定链接和发现页都只允许 `public` 且未被拒绝的媒体。
- 会员媒体详情新增访问权限表单：`public` 可直接复制绝对链接，`link` 不进入发现页但持有链接即可访问，`private` 不生成稳定公开链接，只允许本人登录后的 own-scope 内容接口访问。显式临时签名 URL/分享链接仍是用户主动创建的例外，并受过期、撤销和违规状态控制。
- 新增 `PATCH /api/v1/media/{id}/visibility`，所有者只能修改自己的 ready 媒体；服务端不接受 `user_id`，并复用媒体服务的可见性校验。媒体列表/详情响应现在返回 `visibility` 和 `moderation_status`。
- 管理员媒体库继续使用 `/admin/media` 的 all-scope Resource，不再复用会员 `/api/v1/media`。管理员打开媒体详情时使用受 `admin.media.view` 保护的 `/api/v1/admin/media/{id}/content` 缩略图预览，因此可以确认其他用户上传的图片，但不会改变其公开/私有设置。
- 公开稳定投递 `/i/{id}` 现在检查媒体可见性和违规状态；违规媒体不能通过稳定链接或分享链接继续返回。举报创建的审计动作从笼统 `media.create` 改为 `moderation.report.create`，管理员处理继续记录 `moderation.report.resolve`。
- NOWPayments 配置校验与 Provider 注册边界同步修正：创建支付只要求 API Key，IPN Secret 只在回调验签使用；当前 503 若仍出现，依据实测属于 `api.nowpayments.io` 外部网络不可达，不再误判成缺少 IPN Secret。

### 验证与边界

- 先观察到失败的可见性、默认值、审计分类测试，再实现；后端媒体、链接、分享、支付配置和审计定向测试通过，前端媒体专项测试通过（旧 route-parts 测试同步修正）。
- 现有历史媒体不会被自动从私有批量改成公开，避免未经用户确认扩大曝光；新增迁移只会把未进入历史发现审核、已完成的历史 `pending` 媒体标为 `approved`，不改变其现有可见性。之后的新上传默认公开且通过，用户仍可逐张切换为私有或仅链接可见。
- 该轮尚未宣称生产 CDN/对象存储已完成；管理员预览走应用受控读取，后续接入 CDN 时必须保持同一可见性、审核和访问日志策略。

## 五十三、2026-09-26：套餐运行时权益与管理员媒体专用操作

### 本轮完成

- Personal API Token 创建和轮换现在读取当前订阅的 `token_limit`；只统计未过期的 active Token，`0` 表示不限量，超过非零上限返回 `409 TOKEN_LIMIT_REACHED`。会员 API 每分钟 Token 限额改为读取 `api_rate_per_minute`，同一来源 IP 仍保留 300 次保护；管理员不受会员 Token 维度限额限制。
- 会员广告接口不再只判断广告自身的 active 状态，而是先读取当前订阅 `ads_enabled`；关闭广告的套餐返回空广告列表，广告管理员配置仍保留在 `/admin/advertising`。
- `/admin/media` 的 `status`、`visibility`、`moderation_status` 改为只读，新增专用动作：隐藏、恢复、审核通过、审核拒绝和永久删除。动作通过 Media Admin Service 执行，永久删除复用存储清理/额度释放流程；成功和失败尝试均写入 `admin.media.<operation>` 审计记录。
- 公开计划目录和当前订阅计划响应移除 `price_amount`、`currency`、`billing_period` 三个旧计划级价格字段，价格展示和下单只使用 `prices[]` 的活动价格版本；订单内部快照继续由结算域维护。

### 验证与边界

- 先新增 Token 限额、广告权益、媒体专用动作和公开套餐契约的 RED 测试，再实现 GREEN；聚焦包测试和 `go test ./... -count=1` 全部通过。
- 本轮未宣称 Redis 图片队列、真实趋势统计、管理员相册媒体关系、公开相册 SSR/SSG、真实支付回调、TLS、备份恢复、对象存储/CDN 或压力测试门禁完成；这些仍按下一阶段顺序继续开发。

## 五十四、2026-09-26：Redis 异步队列、真实趋势统计与管理员相册媒体关系

### 本轮完成

- 新增真实趋势统计 Service 和管理员接口 `GET /api/v1/admin/statistics/trends`。统计按用户、媒体、相册、订单、支付流水的真实时间字段聚合，并从 `usage_ledgers` 的下载记录累计带宽字节；接口返回连续的日粒度时间桶，不再用前端静态占位数字。缺失的可选业务表按未启用指标处理，已存在表的查询失败会返回 `STATISTICS_UNAVAILABLE`，避免把数据库错误伪装成零用量。
- `/admin/statistics` 增加起止日期、趋势表和带宽字节展示，日期范围限制为 1 至 93 天；中英文文案通过 `statistics` namespace 提供。
- 新增管理员相册媒体关系页 `/admin/albums/:id/media` 和三项受 RBAC 保护的 API：查询、加入、移除。加入时服务端校验相册所属用户、媒体所属用户和 `ready` 状态，禁止管理员把其他用户媒体跨用户挂入相册；移除只解除 `album_media` 关系，不删除媒体；所有变更写入 `admin.albums.media.add/remove` 审计。
- 新增 Redis Job：`fastimg.billing.fulfill-order` 和 `fastimg.media.recover-upload`。支付履约和过期上传恢复均保留数据库状态为事实来源，失败按有限次数重试；`media:dispatch-recovery` 每 5 分钟扫描最多 100 个超时 `processing` 会话并投递恢复任务。
- 上传确认仍保持同步，确保 C 端成功后立即获得多格式链接；本轮只把可重试的恢复、履约放入 Redis，不为了“异步”破坏现有上传交互。队列配置复用 Goravel Redis Queue 和 Provider 注册，不新增内存队列或第二套调度器。
- OpenAPI、前端路由、管理员页面、双语 locale、队列契约测试和阶段记录已同步。

### 验证结果

- `go test ./... -count=1` 通过；前端 Node 合约测试 `69/69` 通过；`vue-tsc -b --pretty false` 通过；Vite 生产构建通过。
- Redis 本机 `127.0.0.1:6379` 返回 `PONG`；后端已重建并重启到 `http://127.0.0.1:53085`，统计路由和管理员相册关系路由已注册，匿名统计请求返回 `401`，管理员统计请求返回连续日桶。
- 管理员相册列表过滤掉已物理删除媒体，只展示 `ready` 媒体；本地运行态重建后需要再次刷新 `/admin/albums/:id/media` 验证已有数据。

### 边界与未完成项

- 本轮未执行数据库迁移，也未向业务库写入测试数据；相册关系使用现有 `album_media` 表，趋势统计使用现有业务表和用量流水。
- 当前已验证 Redis 可连接和应用启动，但尚未启动独立生产 Worker 做真实任务消费、失败重试、Redis 重启恢复和积压告警演练；不能据此宣称异步队列生产就绪。
- 公开相册真实访问路由、SEO/SSR/SSG、对象存储/CDN、真实支付回调、TLS、备份恢复和压力测试门禁仍未完成。

## 五十五、2026-09-26：上传自动公开与发现页事后治理

### 规则修正

- 普通上传完成后写入 `visibility=public`、`moderation_status=approved`，用户不需要等待管理员审核即可查看、复制和使用稳定图片链接；用户仍可自行切换为 `private` 或 `link`。
- 发现页不是主动投稿队列，而是公开媒体的展示面。只要媒体 `ready`、未删除、`visibility=public` 且未被拒绝，上传后自动进入发现页；举报、隐藏和拒绝是发布后的治理流程。
- 用户端不再显示投稿入口，也不再调用 `discovery-submit`；举报按钮直接创建举报并关联管理员举报处理。
- 管理员媒体操作保持事后运营：`hide` 只改为私有，不把图片标成违规；`reject` 才写入 `moderation_status=rejected` 并阻断公开投递；`approve`、`restore` 和永久删除继续走专用 Service，禁止通用 CRUD 直接写生命周期或审核字段。
- 新迁移 `20260926000001_default_uploaded_media_approved` 更新 PostgreSQL 默认值，并只把“已完成、未投稿发现页、历史 pending”的媒体改为 `approved`；不会把历史私有媒体改成公开，也不会覆盖已进入发现审核的记录。该迁移只注册，尚未执行。

### 验证边界

- 已补充默认状态测试、发现查询的事后治理条件和管理员隐藏/拒绝分离实现；`go test ./... -count=1`、前端 Node 测试 `69/69`、`vue-tsc -b --pretty false` 和 Vite 生产构建均通过。
- 后端已重建并重启到 `http://127.0.0.1:53085`；`GET /api/v1/discovery/status` 返回 `200` 且只返回 `enabled`，旧投稿路径返回 `410 DISCOVERY_SUBMISSIONS_DISABLED`，OpenAPI 已移除投稿操作。当前本地站点配置的发现总开关为关闭，因此 feed 按设计返回 `DISCOVERY_DISABLED`；开启总开关后将按公开、ready、非拒绝规则返回图片。
- 真实队列失败重试/Redis 重启、真实支付回调、TLS、备份恢复、对象存储/CDN 和压力测试仍按生产门禁逐项推进，不能用本地配置或单元测试替代真实验收。

## 五十六、2026-09-26：功能关联审计修复

### 本轮完成

- 修复上传配额边界：普通网页登录上传按日上传、文件大小、存储和转换权益计量；只有 Personal API Token 的单文件上传消耗 `monthly_api_uploads`。API Token 上传仍经过同一上传服务、幂等键、媒体处理和绝对链接生成流程。
- 修复管理员上传水印兜底：管理员继续使用当前设置/计划的水印开关；订阅快照缺失或损坏时不再返回 `SUBSCRIPTION_UNAVAILABLE` 阻断管理员上传，普通会员仍保留严格订阅错误。
- 修复广告运行时条件：会员广告接口同时应用 `ads_enabled`、`plan_code`、`starts_at` 和 `ends_at`，结束时间按排他边界处理；公共广告查询仍可用于管理员预览。
- 修复举报与发现页关联：举报对象条件改为 `ready + public + 非 rejected`，与发现页一致，支持上传后事后举报；不再要求先进入人工投稿审核队列。
- 修复管理员媒体前端资源契约：`status`、`visibility`、`moderation_status` 均为只读；状态变化只能调用隐藏、恢复、通过、拒绝、永久删除专用动作，后端继续复用存储清理、用量释放和审计链路。
- 收紧支付开发默认值：`payment.fake.enabled` 在 `production` 环境默认关闭，开发/测试环境仍可显式或按本地默认启用；真实渠道配置和回调仍必须单独验收。

### 验证结果

- 先执行 RED 测试，再实现 GREEN：上传来源、广告时段/套餐、举报可见性、管理员水印和 fake 网关默认策略专项测试通过。
- 后端 `go test ./...` 通过；前端 Node 合约测试 `69/69` 通过；`vue-tsc -b --pretty false` 通过；Vite 生产构建通过。
- 后端已使用新源码重建并重启到 `http://127.0.0.1:53085`；运行态确认 `/api/v1/discovery/status` 返回 `200`，公开计划响应只使用 `prices[]`，不再输出旧计划级价格字段。

### 仍需保留的生产门禁

- 真实 Redis Worker 消费、失败重试和 Redis 重启恢复尚未完成演练；现有实现已把履约和过期上传恢复投递到 Redis，但上传主流程仍同步以保证成功后立即返回链接。
- 真实支付回调、TLS/反向代理、隔离备份恢复、依赖漏洞扫描、对象存储/CDN、公开相册 SSR/SSG 和业务压力测试仍不能用本地测试替代。

## 六十、2026-09-27：公开相册访问与 SSG

### 本轮完成

- 新增访客相册 API `GET /api/v1/public/albums/{id}` 和图片内容 API `GET /api/v1/public/albums/{id}/media/{media_id}/content`；相册关系、公开性、删除状态和违规状态由后端逐次校验。
- 新增会员应用内的公开路由 `/a/:id`。它与登录后的 `/albums/:id`、管理端 `/admin/albums/:id/media` 完全分离，不会把访客带入后台。
- 公开相册页面支持瀑布流图片、错误兜底、动态 title/description/canonical/Open Graph，并同步中英文文案和 OpenAPI 契约。
- `build:ssg` 支持使用 `SSG_PUBLIC_ALBUM_IDS` 指定要预渲染的公开相册；脚本通过 `SSG_API_ORIGIN` 读取公开数据，未配置或相册不可公开时跳过，不会把私有相册写入静态产物。

### 验证边界

- 已补充公开相册 Service、路由、OpenAPI、前端页面和 SSG 合约测试；需要运行完整后端/前端测试并在本地数据库中用一个公开相册完成真实 HTTP 验收。
- 当前完成的是 SPA + 可选 SSG，不是请求级 SSR；公开相册不会自动全部写入静态站点，生产构建时必须显式提供 ID 清单。
- Redis Worker 消费/重试、真实支付回调、TLS、备份恢复、对象存储/CDN 和压力测试门禁仍未完成。

## 五十七、2026-09-26：公开入口、真实控制台与 SSG 运行态修复

### 本轮完成

- 管理员控制台不再只展示用户、角色和权限三个占位统计；新增媒体、相册、文件夹、订单和支付流水卡片，全部读取 `/api/v1/admin/overview` 的数据库计数，并保留后台资源快捷入口。
- 修复公开套餐卡片与当前价格契约不一致的问题：公开 `/api/v1/plans` 只返回 `prices[].amount_minor/currency/billing_period`，会员端不再读取已移除的旧计划级价格字段，游客价格不再出现 `NaN undefined`。
- 游客路由验收通过：`/`、`/plans`、`/discover` 直接打开；首页上传按钮显示登录入口但不会发起上传，套餐目录可读，个人媒体、订单、Token 等私有页面仍由路由守卫保护。
- 修复 `build:ssg` 产物：公开页面继续生成 description、Open Graph、canonical 和 JSON-LD，同时保留 `#app` 挂载点，Vue 脚本加载后可以继续接管真实页面；静态首屏不写入 Token、用户媒体或管理数据。

### 当前 SEO 结论

- 已实现：公开页面 SPA SEO 元数据、`/sitemap.xml`、`/robots.txt`、首页/套餐/发现页 SSG 首屏和公开 API 数据边界。
- 尚未实现：真正请求级 SSR。当前仍是 Vue SPA + SSG，独立 `web/` SSR 入口、服务端渲染运行时和部署适配仍是后续切片，不能把当前实现描述为 SSR。

### 验证结果

- 前端专项合约测试、`vue-tsc -b --pretty false` 和 `npm run build:ssg` 通过；SSG 生成 `/`、`/plans`、`/discover` 三个入口且产物保留 `id="app"`。
- 浏览器游客验收：三个公开路由均未跳转登录，套餐真实显示 `¥19.99/ 月`、`¥49.99/ 月`；管理员控制台显示当前数据库实数（本地验收时用户 1、媒体 27、相册 2、文件夹 0、订单 8、支付流水 0）。

## 五十八、2026-09-26：注册、Turnstile 与邮箱验证

### 本轮完成

- 退出登录由会员端和管理员端统一回到公开首页 `/`；即使退出接口暂时不可用，也会清理本地认证状态并导航到首页，不把用户困在登录页。
- 新增公开 `/register` 和 `/verify-email` 页面。首页、套餐、发现页继续允许游客访问；上传、媒体、订单、Token 等私有能力仍需登录。
- 新增认证公开接口：`GET /api/v1/auth/registration-policy`、`POST /api/v1/auth/register`、`GET /api/v1/auth/verify-email`、`POST /api/v1/auth/resend-verification`。注册成功默认创建 Free 订阅，不自动赋予管理员权限。
- 后台“注册与登录”设置支持：开放注册、全局 Cloudflare Turnstile、登录验证、注册验证、邮箱验证、邮箱域名白名单、验证链接有效期。Turnstile Site Key 可以返回给浏览器；Secret Key 使用 `APP_KEY` 加密保存，列表接口只返回已配置占位符。
- Turnstile 适配严格使用 Cloudflare 官方 Siteverify 接口。后端在登录、注册和重发验证邮件时校验 token，限制 token 最大长度；不记录 token、Secret Key 或原始验证链接。缺少 Secret Key 返回配置错误，验证失败返回可本地化的业务错误码。
- 邮箱验证 token 只保存 SHA-256 摘要，原始 token 只进入邮件链接；token 单次消费、默认 30 分钟过期，可在后台设置 5–1440 分钟。管理员创建的用户视为已验证，历史用户迁移时按创建时间补齐已验证状态，避免打开策略后锁死现有管理账号。
- 邮箱白名单只在后端读取，支持换行、逗号和分号分隔的精确域名；白名单内容不进入公开策略响应。注册验证邮件统一由 `app/services/email` 发送，支持通用 SMTP、阿里云 DirectMail `SingleSendMail` 和 Resend API，后台可启用/关闭并选择当前服务。SMTP、阿里云 AccessKey 和 Resend API Key 均按 `APP_KEY` 加密保存。

### 验证边界

- 已通过后端认证、策略纯函数、迁移和控制器编译测试，以及前端 TypeScript 检查和生产构建。
- 当前本地默认关闭 Turnstile 和邮箱验证，因此开发账号可继续登录，注册不要求外部服务；管理员在 `/admin/settings` 开启后才会强制执行。
- 未登录重发验证邮件已接入可配置的邮箱/IP冷却和 UTC 自然日上限，默认分别为 60 秒、10 秒、单邮箱 5 次、单 IP 20 次；缓存不可用时拒绝发送，避免绕过限流消耗邮件额度。后台字段位于“注册与登录”设置组，触发限制返回 `AUTH_VERIFICATION_RESEND_RATE_LIMITED` 和 `retry_after_seconds`。
- 真实 Cloudflare Widget、邮件服务投递（SMTP/阿里云/Resend）、域名白名单生产验收仍需要部署方填写正式密钥、正式域名和邮件服务配置；本地单元测试不能替代第三方服务验收。

### 2026-09-27 本地运行验收补充

- 按本地开发授权创建专用 PostgreSQL 角色 `fastimg_app` 和数据库 `fastimg_dev`，没有修改其他 Laragon 数据库；Redis 继续使用 Laragon 的 `127.0.0.1:6379`。
- `fastimg_dev` 已执行全部已注册迁移和 Seeder，`migrate:status` 显示全部 `[1] Ran` 或 `[2] Ran`；已预置 Free、Creator、Pro 三个启用套餐和开发管理员账号。
- 修复 `20260926000002_add_email_verified_at_to_users` 在 Goravel Schema Builder 与 ORM 跨连接时自持表锁的问题，Schema 变更和历史用户回填现在通过同一 PostgreSQL 命令完成，并加入迁移回归测试。
- 当前本地进程使用后端 `http://127.0.0.1:53085`、会员/管理前端 `http://127.0.0.1:53084`；前端首页、管理入口、发现状态、套餐、注册策略、`sitemap.xml` 和 `robots.txt` 实测返回 HTTP 200。该端口分配仅适用于本次开发进程，不代表生产配置。

## 五十九、2026-09-26：会员到期自动降级与通知设置

### 本轮完成

- 新增订阅到期生命周期：付费订单履约时根据月付/年付和试用天数写入 `subscriptions.ends_at`；运行时解析订阅发现已过期时，将旧订阅标记为 `expired`，并自动创建配置的 fallback 订阅，默认使用 `free`。
- 新增 `grace_period_ends_at`，用于记录到期后的数据保留/续费提示窗口。宽限期不继续提供付费上传额度；默认策略是保留媒体和稳定链接，禁止超出 Free 额度的新上传，允许查看、下载和删除。
- 新增 `subscription_notification_deliveries` 去重表，以 `subscription:{id}:expiry:{kind}` 保证提醒和到期邮件不重复发送；记录发送状态、尝试次数、下次重试时间和最后错误。
- 新增 `subscriptions:process-expiry` 命令并加入每日调度：按后台设置发送 7/3/1 天提醒、到期通知和站内通知；邮件未配置或关闭时延后重试，不阻塞套餐降级。
- `/admin/settings` 新增“会员到期”分组，可设置邮件通知开关、fallback 套餐编码、宽限期天数、提醒天数和超出免费额度的处理策略；中英文字段、说明和选择框已同步。

### 验证结果和边界

- TDD 先验证到期判定、提醒日解析、宽限期和周期计算的 RED/GREEN；`go test ./... -count=1` 通过，前端 `npm run build` 通过。
- 已实现服务端自动降级和开发环境可配置入口；真实 SMTP、阿里云 DirectMail、Resend 投递仍需要部署方填写正式凭证、发件域名并完成 SPF/DKIM/DMARC 配置。
- 已有历史付费订阅如果没有 `ends_at`，不会被猜测降级；需要通过真实订单履约或管理员明确设置期限后才进入到期流程。

## 六十一、2026-09-27：公开相册与对象存储运行时关联

### 本轮完成

- 公开相册新增访客路由 `/a/:id`、公开 API 和图片内容 API；相册、媒体归属、公开状态、完成状态、删除状态和违规状态由后端逐次校验。
- 公开发现页与公开相册内容读取统一通过对象存储 RuntimeRegistry 解析当前主存储，不再在控制器中固定使用本地磁盘；因此启用 R2、阿里云 OSS 或腾讯 COS 后，读取链路与上传链路使用同一存储配置。
- `build:ssg` 支持 `SSG_PUBLIC_ALBUM_IDS` 和 `SSG_API_ORIGIN`，只把 API 确认公开的相册生成 `/a/:id/` 静态 SEO 入口；未公开或查询失败的相册跳过。

### 验证结果和边界

- 公开相册和 RuntimeRegistry 先完成 RED 测试，再通过受影响 Go 包、Go 全量测试、前端契约测试、`vue-tsc -b` 和 SSG 构建。
- 当前本地后端 `53085` 已重启，Redis 队列启动日志显示正在消费 `default` 队列；未停止或修改 Laragon Redis。
- 本地数据库当前没有可用于正向验收的公开相册，因此已验证私有/不存在相册返回 `PUBLIC_ALBUM_NOT_FOUND`；正向图片访问需要管理员先将测试相册设置为公开。
- 真正 SSR、队列失败重试与 Redis 重启演练、真实支付回调、TLS、备份恢复、对象存储生产连通性、CDN 和压力测试仍属于生产门禁，不因本轮通过构建而宣称完成。

## 六十二、2026-09-27：Redis 失败任务中心

### 本轮完成

- 新增管理员失败任务 API：`GET /api/v1/admin/tasks` 和 `POST /api/v1/admin/tasks/{uuid}/retry`。列表支持分页，重新投递复用 Goravel Queue Failer，不复制队列实现。
- 新增 `/admin/tasks` 页面，面向非技术管理员显示任务、连接、队列和失败时间；可见权限为 `admin.tasks.view`，重新投递按钮额外需要 `admin.tasks.retry`。
- 失败任务响应只返回必要运维摘要，不返回原始任务 payload、异常堆栈或密钥；成功重新投递写入 `queue.task.retry` 审计日志。
- 新增中英文错误提示、OpenAPI 契约、路由导航和权限迁移 `20260927000003_add_task_permissions`。该迁移已按既有本地授权仅执行到 `fastimg_dev`。

### 验证边界

- RED/GREEN：失败任务权限/分页测试先失败后通过；管理员路由、API Client、页面和 OpenAPI 合约测试通过；`vue-tsc -b --pretty false` 通过。
- 当前后端仍使用 Redis 队列，实时上传仍同步完成以保证 C 端立即得到图片链接；失败任务页面是人工运维入口，不等于完成真实故障注入验收。
- 独立 Worker 消费、构造失败任务、重新投递、Redis 重启恢复、真实支付回调、TLS、备份恢复、对象存储/CDN 和压力测试仍是生产门禁。

## 六十三、2026-09-27：设置契约与对象存储统计

### 本轮完成

- 修复 `/admin/settings` 选择框保存失败：`select` 只是前端展示类型，提交系统设置时统一映射为后端支持的 `string`，因此邮件服务类型等选择项不再触发 `SETTINGS_INVALID`。敏感字段仍由现有设置服务使用 `APP_KEY` 加密保存，列表接口只返回占位符。
- 对象存储设置从通用站点设置表单中抽成独立分组，继续使用每个 Provider 自己的保存按钮；这样保存 R2、阿里云 OSS、腾讯 COS 时不会被其他站点字段阻断，也不会把云凭证混入普通设置提交。
- 新增管理员接口 `/api/v1/admin/storage/statistics` 和 `/admin/storage` 动态统计区。只对已启用的连接生成对应 Provider 统计页卡，展示对象数、已存容量、允许访问次数和估算带宽；统计来源明确标记为 `fastimg_database`，由本站 `storage_objects` 与允许访问记录计算。
- 不虚构云厂商账单/出口流量：R2、阿里云 OSS、腾讯 COS 的真实账单、请求计费和公网出口仍属于各厂商控制台或后续专用 Billing API 集成，不等同于本站媒体访问统计。
- 更新对象存储注册入口：腾讯云使用 `https://curl.qcloud.com/sTyZPtN7`，阿里云使用 `https://www.aliyun.com/minisite/goods?userCode=i7hvp048`。云连接仍可并行启用，启用后管理菜单和统计页按连接动态出现。

### 验证结果和边界

- RED/GREEN：前端保存契约、对象存储统计页面和设置布局测试通过；后端全量 `go test ./... -count=1` 通过；前端全量 Node 测试 `113/113`、`vue-tsc -b --pretty false` 和 `npm run build:ssg` 通过。
- 真实验收：未认证访问统计接口返回 `401 AUTH_UNAUTHORIZED`；管理员登录后统计接口返回当前启用连接及 `fastimg_database` 数据源；`email.provider` 以 `value_type=string` 成功写入；腾讯云和阿里云注册链接通过管理接口返回新地址。
- 本轮没有新增数据库迁移，也没有修改 Laragon PostgreSQL/Redis 配置；本地后端已安全替换并重启在 `53085`，本次 Go 临时缓存已清理。
- 仍未宣称云厂商生产凭证连通、厂商账单 API、对象存储/CDN 生产切换、真实支付回调、TLS、备份恢复和压力测试门禁完成。

## 六十四、2026-09-27：对象存储统计扩展

### 本轮完成

- `/api/v1/admin/storage/statistics` 扩展为本站侧运营统计：对象状态分布、原图/缩略图/中图变体分布、关联媒体数、孤儿对象数、可回收容量和 24 小时健康错误数。
- 每个已启用连接返回近 30 天每日桶，包含上传数量、上传容量、允许访问次数和估算访问流量；管理端页面展示最近 7 天明细，空日期也返回 0，避免趋势图表断线。
- 接入已有 `storage_health_checks` 表读取最近检查状态、延迟和错误次数；不把连接配置状态误当成云厂商健康状态。
- 管理端对象存储页面增加对象/变体分布、健康摘要、孤儿对象和趋势区域，中英文 locale 同步维护；OpenAPI 契约同步扩展。

### 验证边界

- RED/GREEN：新增前端统计字段契约先失败后通过；统计专项测试、双语 locale 检查、`vue-tsc -b --pretty false` 和 `npm run build:ssg` 通过。
- 后端新构建成功并重启在 `53085`；管理员真实接口返回当前启用连接的 31 个日趋势桶、健康状态和异常字段。当前开发库没有媒体对象，所以对象、媒体和孤儿对象数量为 0。
- 本轮 Go 定向测试启动阶段受到 Windows `Access is denied` 影响，但后端完整编译和真实 HTTP 统计接口验收通过；该测试环境问题不能等同于业务测试通过。
- 仍未接入 R2、阿里云 OSS、腾讯 COS 的官方账单、云端真实容量、CDN 命中率和公网出口流量；这些数据必须通过各 Provider 独立凭证/API 获取，不能用本站估算值替代。

## 六十五、2026-09-27：修复站点地图白屏

### 本轮完成

- 修复管理端 SEO 区域的 `sitemap.xml` 和 `robots.txt` 入口：开发环境不再使用前端当前端口或 `site_url` 直接拼接，而是使用 API 后端基地址打开机器可读文件。
- Vite 开发服务器和预览服务器增加 `/sitemap.xml`、`/robots.txt` 到后端的代理，直接打开前端端口的公开文件也不会再被 SPA fallback 成空白页面。
- 保留生产环境同源行为：未配置 `VITE_API_BASE_URL` 时使用当前站点来源；配置后使用显式后端/网关来源。

### 验证结果和边界

- 设置布局与开发代理测试 `10/10` 通过，`vue-tsc -b --pretty false` 通过，`npm run build` 通过。
- 后端 `http://127.0.0.1:53085/sitemap.xml` 实测返回 `200 application/xml`；白屏根因是前端端口错误承接 XML 请求，而不是站点地图 XML 生成失败。
- sitemap 内容仍受后台 `sitemap.enabled`、公开路径过滤和 `site_url` 规范域名设置控制；真实生产域名、TLS 和反向代理仍需部署时配置。

## 六十六、2026-09-27：会员上传页剪贴板粘贴上传

### 本轮完成

- 首页 `/` 和媒体库 `/media` 共用的 `MemberUploadPanel` 增加全局 `paste` 监听；登录用户按 `Ctrl+V` 粘贴截图或其他图片文件后，文件会进入现有上传队列，不新增接口、不绕过现有套餐和服务端校验。
- 只处理剪贴板中的图片文件；普通文字粘贴不调用 `preventDefault()`，因此不会破坏输入框和浏览器默认粘贴行为。游客不会读取剪贴板，上传仍由登录认证保护。
- 粘贴提示增加中英文文案；上传完成后的预览、状态轮询和四种带域名链接复制继续复用现有实现。

### 验证结果和边界

- RED/GREEN：`tests/member-upload-links.test.mjs` 的剪贴板契约先失败后通过，专项测试 `4/4` 通过。
- 支持范围与文件选择器一致：当前接受 JPG、PNG、GIF；浏览器无法提供图片文件时不拦截粘贴事件。真实浏览器剪贴板权限和不同浏览器的截图格式仍应在部署验收中抽测。
- 本轮没有新增 API、权限、数据库迁移或存储配置；仍需按生产门禁完成真实支付回调、TLS、备份恢复、对象存储/CDN 和压力测试。

## 六十七、2026-09-27：修复粘贴上传请求失败和会话错误提示

### 本轮完成

- 修复后端 CORS 预检：浏览器会员上传使用的 `Idempotency-Key` 现在与 `Content-Type`、`Authorization` 一起出现在 `Access-Control-Allow-Headers`，跨 53084 前端与 53085 后端的 multipart POST 可以真正发送。
- 会员上传 composable 对 `AUTH_UNAUTHORIZED` 单独显示“登录状态已失效，请重新登录后再上传”，不再把认证问题伪装成文件损坏或格式错误。
- 仅重启当前项目自己的 53085 后端进程；没有修改 Laragon 数据库、Redis 或生产 CORS 白名单规则。

### 验证结果和边界

- 实测预检 `OPTIONS http://127.0.0.1:53085/api/v1/uploads` 返回 `204`，并允许 `Idempotency-Key`。
- 使用浏览器当前登录账号按 `Ctrl+V` 粘贴测试 PNG，实际上传返回 `ready`，预览和 URL、Markdown、HTML、BBCode 四种链接均显示；测试图片已移入回收站清理。
- 前端相关契约、类型检查和生产构建继续通过；Go 定向测试启动仍受本机 Windows `Access is denied` 运行环境限制，但后端构建成功并已完成真实 HTTP 预检验收。

## 六十八、2026-09-27：修复公开链接端口并切换为单原图存储

### 本轮完成

- 修正开发环境 `backend/.env` 的 `APP_URL` 为 `http://127.0.0.1:53085`。53083 是历史前端端口且当前没有服务，导致上传结果中的 `<img>` 预览和公开 `/i/{id}` 链接无法打开；签名本身在 53085 已实测返回图片。
- 上传处理器不再生成或写入 `thumbnail`、`medium`，`ProcessedImage`、上传预留、媒体列表/详情、链接格式、公开发现、公开相册、会员内容和管理员预览均只接受 `original`。
- 上传结果保留用户需要的 `original`、`url`、`markdown`、`html`、`bbcode` 五类信息；接收网站自行用 CSS 或自己的图片处理服务控制显示尺寸。
- 为避免部署切换时丢失正在处理的旧任务，恢复逻辑临时兼容已存在的三 Variant 处理会话；新上传永远只建立一个原图对象。旧媒体的历史缩略图/中图不主动删除，永久删除仍会清理其全部旧对象，避免误删用户数据。
- 同步更新 OpenAPI、开发者 API、数据契约、产品设计和前端预览请求；发现页和公开相册保留旧的 `thumbnail_url` 字段名以降低客户端破坏性，但其值现在指向原图内容地址。

### 验证结果和边界

- 根因验收：`53083` 对用户提供的链接返回连接拒绝；同一签名 URL 改为 `53085` 返回 `200 image/png`。
- 新增/更新原图唯一存储、链接键集合和公开 Variant 契约测试；Go 测试启动仍受本机 Windows Go cache `Access is denied` 限制，不能把该环境限制误报为测试通过。
- 已完成后端重新编译和前端类型/构建前置检查；重启后需重新打开媒体详情或重新上传一次，才能拿到使用 53085 的新链接。生产环境必须把 `APP_URL` 设置成真实 HTTPS 域名，不应把 53085 带入生产。
