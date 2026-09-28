# FastImg 图床产品设计（基于 Go Vue Admin）

## 1. 产品目标

FastImg 是面向开发者、站长和内容创作者的免费媒体托管平台：免费用户可以完成上传、管理、分享和基础 API 使用；付费用户购买更大的存储、流量、更高的 API 配额和高级控制能力。

核心价值不是“把上传按钮做出来”，而是提供稳定、可追踪、可控制的媒体资产生命周期：

```text
上传 -> 校验 -> 存储 -> 处理 -> 自动可用并进入发现页 -> 分享/访问 -> 举报治理 -> 统计 -> 删除/恢复
```

## 2. 免费商业模型

### Free

| 权益 | 初始建议值 | 说明 |
| --- | ---: | --- |
| 存储 | 1 GB | 回收站内容继续占用 |
| 单文件 | 10 MB | 按真实文件大小校验 |
| 每日上传 | 100 张 | 按用户时区或站点时区统一配置 |
| 月 API 上传 | 500 次 | API 和客户端上传共用配额 |
| 月外链流量 | 5 GB | 服务端事件计量 |
| Personal API Token | 1 个 | 支持撤销和重新创建 |
| 图片处理 | 原图规范化 | 水印由套餐权益决定；接收站通过 CSS 或自身图片处理控制展示尺寸 |
| 链接格式 | URL、Markdown、HTML、BBCode | 免费可用 |
| 广告 | 可展示 | 通过套餐权益关闭 |
| 图片水印 | 默认关闭 | 按套餐权益控制，开启时写入唯一存储的原图 |

### Creator

适合站长和内容创作者：提高存储、流量、单文件和 API 配额；关闭广告；开启原图、批量下载、基础访问统计、长周期回收站和自定义过期时间。

### Pro

适合高频开发者和站点：增加自定义域名、防盗链、多个 Personal API Token、Webhook、高级统计、优先处理和更高速率。

套餐、价格、权益值、排序和启停均由后台管理；业务代码不能硬编码这些数值。`watermark_enabled` 是套餐权益而不是前端展示字段，上传服务端从当前订阅快照读取它；管理员可在套餐编辑页开启或关闭，水印文字、域名和加载失败兜底图分别由站点设置 `watermark.text`、`watermark.domain`、`watermark.fallback_image_url` 配置。

## 3. 复用 Go Vue Admin 的能力

以下能力不在 FastImg 重复建设：

- JWT、Refresh Token、登录限流、退出和会话
- 用户、角色、权限、数据范围和字段权限
- Resource Registry、Manifest、通用列表/表单/详情、搜索、筛选、分页、导出和批量 Action
- 系统设置、通知、审计日志、敏感字段脱敏
- Goravel ORM、迁移、Redis、队列、文件系统、验证、邮件和日志
- OpenAPI 文档和 TypeScript API Client 生成
- Vue Router、Pinia、i18n、shadcn-vue 基础组件和 Admin Shell

FastImg 只注册业务资源、业务权限、业务路由和业务服务。复杂流程使用自己的 Custom Page，但仍复用框架权限、API 错误、审计和 UI 组件契约。

### 公开内容、页脚导航与友情链接

站点运营内容与管理员工作台严格分离：

| 面向用户 | 面向管理员 |
| --- | --- |
| `/page/privacy`、`/page/terms`、`/page/about` | `/admin/content-pages`：草稿、发布、归档、SEO 字段 |
| `/friends`：已审核友链卡片与游客申请表 | `/admin/friend-links`：查看申请、通过、拒绝、隐藏 |
| MemberShell 页脚：按公开接口显示启用的分类、菜单和子菜单 | `/admin/footer-navigation`：管理分类、菜单、父子关系、排序和启用状态 |

内容编辑只接受 Tiptap JSON。服务端会再次限制节点、Mark、属性和 URL 协议，拒绝 `script`、`iframe`、事件属性、`javascript:`/`data:` URL、回环和内网目标。前端使用结构化 Vue 节点渲染，不接受原始 HTML，因此管理员误粘贴脚本不会直接进入页面。

友情链接默认进入 `pending`，公开接口只返回 `approved`；重复站点、危险 URL、IP/URL 冷却和每日上限在服务端处理。审核通过、拒绝和隐藏均是显式动作并写入审计日志，不能通过通用 CRUD 直接修改审核状态。

默认页面和默认页脚菜单由幂等种子创建，不覆盖管理员已编辑内容。SEO/SSG 清单包含首页、套餐、发现、友链和三个默认内容页；构建时生成可被搜索引擎直接读取的 HTML，运行时公开 API 负责内容更新后的回退展示。

## 4. 业务模块与真实代码目录

### 后端

```text
backend/app/modules/
├── media/              # MediaAsset、Folder、Album、MediaVariant 资源和专用页面契约
├── uploads/            # UploadSession、分片完成确认和上传 API
├── storage/            # 存储节点、Provider 注册和对象生命周期
├── plans/              # Plan、PlanEntitlement、Subscription 资源
├── billing/            # Order、PaymentEvent、退款/订阅状态
├── shares/             # ShareLink、公开访问和防盗链
├── moderation/         # ModerationTask、Report、封禁和审核 API
├── analytics/          # MediaEvent、聚合查询和仪表盘 API
├── advertising/        # AdSlot、AdCreative、Campaign
└── developer/          # Personal API Token、Webhook 和客户端配置

backend/app/services/
├── media/
├── uploads/
├── storage/
├── quota/
├── billing/
├── moderation/
└── analytics/
```

`backend/app/modules` 存放 Controller、Resource、Manifest、Routes、权限和模块级模型；跨多个资源复用的业务规则放在 `backend/app/services/<domain>`，不能放入 `backend/app/core`。

### 前端

```text
admin/src/modules/
├── media/              # 媒体库、媒体详情、文件夹、相册
├── uploads/            # 上传页、上传队列、进度和失败重试
├── plans/              # 套餐和用量
├── billing/            # 订单和支付状态
├── shares/             # 分享链接
├── moderation/         # 管理端审核中心
├── analytics/          # 用户和管理端统计
├── advertising/        # 管理端广告位
└── developer/          # Personal API Token、Webhook、API 文档入口
```

通用列表优先通过 `admin:make-resource` 生成；上传器、媒体瀑布流、媒体详情、用量图表、支付和审核工作台使用 Custom Page。

这里的 `admin/src` 是当前仓库唯一 Vue 应用的物理目录，不代表所有页面都属于管理员。MVP 可以继续共用同一个 Vite 应用和框架基础设施；会员页使用 `/`、`/media`、`/plans` + `MemberShell`，发现页走 `/discover`，后台使用 `/admin/**`、现有 `AdminShell` 和后台导航。用户媒体页面只对当前用户生效；管理员跨用户查询通过显式管理权限、数据范围和审计实现。

### 当前实现状态（2026-09-23）

会员端已建立扁平 URL 路由和 `MemberShell`：`/` 是上传优先首页，展示最近本人媒体及服务端返回的套餐/存储摘要；`/media` 支持本人图片上传、查看/预览、搜索、软删除和回收站恢复；`/media/:id` 已支持认证链接的 URL、Markdown、HTML、BBCode 复制、可过期/密码分享、签名 URL 生成和基础防盗链策略维护；`/share-links` 支持查看本人分享摘要和撤销；`/tokens` 支持 Personal API Token 创建、一次性复制、撤销和轮换；`/plans` 展示本人套餐与用量。首页和媒体页共用上传队列/状态轮询。管理页面位于 `/admin/**`；`/admin/media` 未注册，own-scope 文件夹/相册不作为全站管理入口；`/admin/media-access-logs` 是管理员专用的脱敏访问记录查询页。旧 `/app` 路径保留静态兼容重定向。`user_id` 不由会员前端提交，归属由认证上下文决定。Token/API 的源代码契约和会员管理页面已落地；M3 防盗链迁移、后端重启、Free 管理员上传和真实签名/Referer/访问记录验收已完成。CDN 原生签名、访问统计聚合报表、带宽计量、订单与公开发现仍在后续计划中；此处状态不代表阶段或生产验收完成。

## 5. 功能设计

### 5.1 用户端

#### 上传

- 拖拽、文件选择、剪贴板粘贴、多文件上传
- 创建上传会话，显示文件大小、类型、剩余额度和预计状态
- 小文件直传，大文件分片上传
- 普通 API 使用 `/api/v1/uploads`；大文件和断点续传使用 `/api/v1/uploads/sessions`
- 上传取消、重试、过期恢复
- 上传完成后异步生成缩略图和中图
- 图片真实类型校验、尺寸限制、像素限制和 EXIF 清理

#### 媒体库

- 网格/列表模式
- 文件夹、相册、标签、搜索和筛选
- 按格式、尺寸、大小、状态、公开性和时间筛选
- 批量移动、分享、下载、删除和恢复
- 媒体详情显示原图、派生图、哈希、尺寸、状态和访问统计

#### 分享

- 原图、缩略图和派生图分别生成链接
- URL、Markdown、HTML、BBCode 一键复制
- 公开、私有、仅链接、密码和过期访问
- 自定义过期时间属于 Creator/Pro 权益
- 私有对象默认不直接暴露对象存储 URL

#### 用量和套餐

- 展示存储、流量、每日上传、月 API、处理次数
- 显示用量明细和超额原因
- 升级立即生效，降级周期末生效
- 降级后不删除已有媒体，只限制新增超额资源

#### API Token

用户可以在个人中心生成 Personal API Token，用于 PicGo、ShareX、脚本、CI 和自建应用。

管理端对应 `/admin/api_tokens` 仅用于全站运营查看和状态处置：按用户查看非敏感 Token 元数据，按 `admin.api_tokens.update` 停用或撤销，按独立的 `admin.api_tokens.delete` 删除；不显示明文、不读取摘要。会员端 `/tokens` 只管理当前用户，两套页面和权限边界保持独立。

API 同时支持单图和批量上传。批量上传按文件返回 `ready`、`processing` 或 `failed`，单个文件失败不能掩盖其他文件的成功结果。

- 创建时设置名称、权限范围和可选过期时间
- 完整 Token 只展示一次，关闭页面后不能再次查看
- 支持启用、禁用、撤销、轮换和最后使用时间
- 默认 Token 自动拥有上传、获取自己图片链接和删除自己图片的基础能力
- 读取媒体列表、读取用量、管理 Webhook 等属于可选扩展权限
- API Token 不得写入日志、审计详情、浏览器 LocalStorage 或数据库明文
- 用户可以看到 Token 最近使用时间、来源 IP、调用次数和剩余配额

Token 使用标准 HTTP Header：

```http
Authorization: Bearer fst_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

上传成功后 API 必须直接返回原图、缩略图、中图、WebP、AVIF、Markdown、HTML、BBCode 和纯 URL 等链接；图片仍在异步处理时返回任务状态和查询地址。

### 5.2 管理端

#### 会员管理

- 用户列表、详情、状态、套餐、用量和最后活动时间
- 手动调整套餐、赠送额度、封禁和解封
- 查看指定用户媒体，但必须执行管理员权限和审计

#### 套餐与权益

- 套餐基本信息、价格、周期、权益、排序、启停
- 存储、单文件、每日上传、月 API、月流量、处理次数和 API 速率
- 免费广告、原图、批量下载、自定义域名、防盗链、Webhook 等开关
- 权益版本化，已有订阅保存权益快照，避免套餐修改影响历史订单

#### 媒体库与审核

- 按用户、状态、格式、大小、哈希、时间查询
- 待审核、举报、人工复核、隐藏、恢复和封禁
- 批量删除必须二次确认并产生审计记录
- 普通上传不等待审核即可使用，并自动进入发现页；被隐藏、拒绝或设为私有后不再展示
- 后台可以全局开启/停用发现页
- 后台可以配置审核模式、默认排序和举报阈值
- 处理举报时可以隐藏图片、删除图片、恢复图片、限制用户或封禁账户

#### 统计

- 用户数、活跃用户、上传数、存储增长、流量、失败任务、审核量
- 套餐转化、收入、退款、免费用户升级率
- 热门媒体、来源域名、设备和时间趋势

#### 广告

- MVP 只提供四个固定广告位：页眉 `header`、页脚 `footer`、左侧栏 `left`、右侧栏 `right`。
- 广告由管理员在 `/admin/advertising` 配置，会员端只读取启用中的广告；普通会员不能创建、修改或提交广告。
- 广告内容类型为文本、图片地址或 JavaScript 源码；脚本只能在会员端的隔离 iframe 中运行，不直接注入 MemberShell。
- 后台管理名称、内容类型、内容、跳转地址、启停、时间窗口和可选套餐编码；Free 默认可展示，Creator/Pro 是否免广告仍由套餐权益决定。
- 当前不做广告活动、复杂定向、竞价、分成和点击统计；后续需要时再增加独立 Campaign/Analytics 领域。
- 广告素材需要管理员审核后启用；不接受任意 HTML，JavaScript 配置只接受源码并通过沙箱加载，跳转地址仅允许 HTTP/HTTPS。

#### 发现页治理

- 后台显示发现页状态、公开媒体量、举报量和违规用户数
- 停用发现页后，公共发现接口关闭，但私有媒体和已有私有分享链接不受影响
- 发现页没有单独的投稿开关；举报和后台审核入口始终可以保留
- 隐藏或拒绝已公开媒体后立即从发现页移除，恢复后重新展示
- 支持人工审核、自动审核和混合审核模式
- 支持按图片、用户、IP、时间和举报原因筛选
- 支持批量隐藏图片、批量删除图片、批量限制用户和批量封禁账户
- 同一用户对同一媒体的相同原因在处理完成前合并为一条举报
- 举报数量达到阈值只触发人工复核，不能仅凭数量自动永久封禁账户
- 高置信度违规可以自动隐藏，但必须可恢复并进入人工复核
- 被处理用户可以查看处罚原因、期限和申诉入口
- 违规处理结果通过站内通知反馈，不向举报人泄露被处理用户的隐私信息

## 6. 核心数据设计

| 对象 | 必要字段/职责 |
| --- | --- |
| MediaAsset | user_id、状态、可见性、原始大小、哈希、MIME、宽高、对象引用 |
| MediaVariant | media_id、类型、大小、宽高、对象引用、处理状态 |
| Folder | user_id、parent_id、名称、软删除 |
| Album | user_id、名称、封面、可见性、密码摘要 |
| UploadSession | user_id、幂等键、文件元数据、分片状态、过期时间 |
| StorageObject | provider、bucket、object_key、etag、大小、删除状态 |
| Plan | 名称、周期、价格、状态、排序 |
| PlanEntitlement | plan_id、权益键、权益值、版本 |
| Subscription | user_id、plan_id、状态、周期、权益快照 |
| UsageLedger | user_id、资源类型、增量、来源、幂等键、时间 |
| ShareLink | media_id、token_hash、密码摘要、过期时间、状态 |
| ApiToken | user_id、token_hash、名称、权限、状态、最后使用时间 |
| ModerationTask | media_id、Provider、状态、结果、原因、重试次数 |
| ModerationAction | 目标类型、目标 ID、动作、原因、证据、操作者、处罚期限 |
| Report | media_id、举报人、原因、描述、状态、处理结果、处理人 |
| ModerationAppeal | user_id、目标、说明、证据、状态、复核人、复核结果 |
| MediaEvent | user_id、media_id、事件类型、字节数、来源、时间 |

所有用户拥有的数据必须包含明确的 `user_id` 或可追溯归属关系；媒体列表、详情、导出、写入和批量操作都必须执行归属判断。

## 7. 状态和关键规则

```text
UploadSession: created -> uploading -> uploaded -> processing -> completed
MediaAsset: pending -> processing -> ready -> deleted -> expired
Moderation: ordinary upload approved; post-publication report/review pending -> approved | rejected | manual_review
Order: created -> pending_payment -> paid -> active -> refunded
```

- 状态只能通过领域服务迁移，不能由 Controller 直接修改状态字段。
- 删除先进入回收站，永久删除通过异步任务执行。
- 存储配额采用预占 + 确认 + 释放模型，避免并发上传超限。
- 计费用量来自服务端事件或用量流水，不能接受前端上报。
- 订单支付回调和上传完成接口必须幂等。
- 外部存储、支付、审核不可用时不得默认放行或标记成功。

### 发现页状态

```text
disabled -> enabled
enabled -> disabled
```

- `enabled`：展示公开且未被拒绝的 ready 内容，举报进入后台治理。
- `disabled`：公共发现接口关闭；私有媒体、用户自己的分享链接和后台审核不受影响。

### 账户处罚等级

```text
normal -> upload_limited -> discovery_banned -> suspended -> banned
```

- `upload_limited`：限制上传或 API 上传，但允许查看、导出和删除自己的媒体。
- `discovery_banned`：禁止公开展示，已有违规媒体进入审核处理。
- `suspended`：暂时禁止登录后的写操作，保留申诉和导出窗口。
- `banned`：禁止登录和所有新写操作，下载策略由站点合规设置决定。

每次处罚记录级别、原因、证据、操作者、开始时间、结束时间和是否允许申诉。处罚不能通过删除用户记录实现。

### 申诉流程

```text
处罚生效 -> 用户提交申诉 -> 人工复核 -> 维持处罚 | 撤销处罚 | 调整处罚
```

申诉保存原处罚、用户说明、补充证据、复核人、复核时间和最终决定。已经永久删除的违规对象不因申诉自动恢复，恢复必须经过管理员确认和对象存在性检查。

## 8. MVP 验收范围

第一版只要求 Free 闭环：用户注册、上传、媒体库、文件夹、分享、Personal API Token、用量、回收站、基础审核、管理员按用户查看媒体、基础统计和后台套餐配置。

不进入第一版：社交关注评论、AI 生图、以图搜图、团队空间、复杂广告分成、视频转码和真实支付 Provider。真实支付和 CDN 必须在有连接、回调/读写和失败测试后才能声明完成。

## 9. 关键用户流程

### 9.1 首次上传

```text
注册/登录
  -> 读取当前套餐和用量
  -> 选择图片
  -> 创建 UploadSession
  -> 浏览器上传对象
  -> Complete 校验对象、哈希和配额
  -> 创建 MediaAsset(processing, moderation=approved)
  -> 处理并完成用量入账
  -> ready 后立即展示链接
  -> ready 后立即进入公开发现页
  -> 用户或游客可以举报公开图片
  -> 管理员事后隐藏、拒绝、恢复或永久删除
```

用户在图片进入 `ready` 前可以看到进度和失败原因；进入 `ready` 后即可获得稳定链接。处理失败必须提供重试，不允许静默丢失。

### 9.2 分享私有图片

```text
媒体详情
  -> 选择“创建分享链接”
  -> 选择公开/仅链接/密码/过期时间
  -> 校验套餐权益
  -> 保存 token_hash 和访问策略
  -> 返回一次性 Token 和代码片段
```

分享链接是访问策略，不是媒体本身的公开状态。撤销分享链接不应删除媒体；删除媒体后所有关联分享链接必须失效。

### 9.3 超额处理

```text
上传/分享/访问
  -> 配额服务判断具体资源
  -> 允许：预占/扣减并继续
  -> 拒绝：返回稳定错误码、当前值、上限和解决建议
```

解决建议只能是释放空间、等待周期重置或升级套餐；不能由前端绕过配额检查。

## 10. 权限与权益矩阵

权限和套餐是两个独立维度：权限控制“是否可以访问功能”，权益控制“额度和高级能力”。

| 功能 | 游客 | Free | Creator | Pro | 管理员 |
| --- | --- | --- | --- | --- | --- |
| 上传 | 否/可选临时 | 是 | 是 | 是 | 有后台权限即可，跳过会员额度 |
| 管理自己的媒体 | 否 | 是 | 是 | 是 | 可按权限查看，跳过会员额度 |
| 原图下载 | 公开策略 | 基础 | 是 | 是 | 是 |
| 自定义过期 | 否 | 默认值 | 是 | 是 | 配置 |
| 防盗链 | 否 | 否 | 基础 | 是 | 配置 |
| Personal API Token | 否 | 1 个 | 多个 | 多个 | 管理 |
| Webhook | 否 | 否 | 否 | 是 | 管理 |
| 公共发现页 | 浏览 | 可选发布 | 可选发布 | 可选发布 | 审核 |
| 会员管理 | 否 | 否 | 否 | 否 | 需要权限 |

管理员角色仍使用框架 RBAC。套餐不能赋予管理员权限，也不能替代后端权限判断；管理员的套餐档位仍可由后台用户编辑页手动指定，用于展示和后续业务策略，但不会把管理员按 Free/付费套餐额度拦截。管理员仍受系统级技术安全上限（例如单请求体积和图片处理器限制）约束。

## 11. 数据关系与删除策略

```text
User
 ├── Subscription -> Plan -> PlanEntitlement
 ├── Folder -> MediaAsset -> MediaVariant -> StorageObject
 ├── Album -> AlbumMedia -> MediaAsset
 ├── MediaAsset -> ShareLink
 ├── MediaAsset -> ModerationTask / Report
 ├── User -> ApiToken / Webhook
 └── User -> UsageLedger / MediaEvent / Order
```

删除策略：

1. 用户删除媒体：MediaAsset 进入回收站，分享链接立即失效。
2. 回收站恢复：恢复 MediaAsset 和必要的派生图引用，重新校验对象是否存在。
3. 永久删除：异步删除原图和派生图对象，删除成功后释放存储额度。
4. 对象删除失败：保留清理任务和失败原因，不释放额度，不标记清理完成。
5. 用户注销：按站点保留策略处理媒体、订单、审计和用量记录；计费与审计记录不能因业务软删除而丢失。

## 12. API 契约规范

### 12.1 通用响应

成功响应使用统一数据结构；列表响应必须包含分页元数据：

```json
{
  "data": {},
  "meta": {"page": 1, "per_page": 20, "total": 0, "last_page": 1},
  "request_id": "..."
}
```

错误响应：

```json
{
  "error": {
    "code": "QUOTA_STORAGE_EXCEEDED",
    "message": "Storage quota exceeded",
    "details": {"used": 1073741824, "limit": 1073741824},
    "request_id": "..."
  }
}
```

### 12.2 API 规则

- 所有写操作必须验证认证、权限、资源归属和套餐权益。
- 所有批量接口限制最大 ID 数量，并返回 succeeded/failures/skips。
- 所有异步任务创建接口返回任务 ID 和当前状态，不等待处理完成。
- 上传完成、支付回调、Webhook 接收必须支持幂等键。
- Personal API Token、分享 Token、密码和支付敏感字段不能出现在日志、审计详情或错误响应中。
- 资源列表使用白名单字段排序和筛选，禁止将客户端字段直接拼进 SQL。

### 12.7 发现页和治理 API

用户端：

```text
GET  /api/v1/discovery/status
GET  /api/v1/discovery/feed
POST /api/v1/media/{id}/reports
POST /api/v1/me/moderation-appeals
GET  /api/v1/me/moderation-actions
```

后台端：

```text
GET  /api/v1/admin/discovery/settings
PUT  /api/v1/admin/discovery/settings
GET  /api/v1/admin/moderation/reports
POST /api/v1/admin/moderation/reports/{id}/resolve
POST /api/v1/admin/moderation/media/{id}/hide
POST /api/v1/admin/moderation/media/{id}/restore
POST /api/v1/admin/moderation/users/{id}/restrict
POST /api/v1/admin/moderation/users/{id}/suspend
POST /api/v1/admin/moderation/users/{id}/ban
POST /api/v1/admin/moderation/appeals/{id}/resolve
```

治理接口必须复用现有 RBAC 和审计服务。`resolve` 请求必须包含处理结果、原因和可选处罚动作；前端不能提交任意账户状态值。

建议权限拆分为：

```text
 admin.discovery.manage       发现页启停和展示设置
admin.moderation.view        查看举报和审核队列
admin.moderation.resolve     驳回/通过/隐藏/恢复媒体
admin.moderation.enforce     限制、暂停和封禁账户
admin.moderation.appeal      处理用户申诉
```

查看权限不自动包含处罚权限；处罚权限不自动包含发现页全局设置权限。

### 12.3 Personal API Token API

```text
POST   /api/v1/tokens
GET    /api/v1/tokens
DELETE /api/v1/tokens/{id}
POST   /api/v1/tokens/{id}/rotate
GET    /api/v1/quota
```

创建请求：

```json
{
  "name": "PicGo home computer",
  "scopes": ["upload:write", "links:read"],
  "expires_at": "2027-09-23T00:00:00Z"
}
```

创建响应只返回一次完整 Token：

```json
{
  "data": {
    "id": 12,
    "name": "PicGo home computer",
    "token": "fst_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
    "scopes": ["upload:write", "links:read"],
    "expires_at": "2027-09-23T00:00:00Z"
  }
}
```

后续列表接口只返回：`id`、`name`、`scopes`、`status`、`created_at`、`expires_at`、`last_used_at`、`last_used_ip` 和用量摘要，不能返回完整 Token 或可逆密文。

### 12.4 Token 权限范围

```text
upload:write    上传图片和创建 UploadSession
links:read      获取当前上传图片的各种链接
media:read      查询当前用户媒体详情和列表
media:delete    删除当前用户媒体
usage:read      查询当前 Token/用户配额
webhook:manage  管理 Webhook
```

Token 校验顺序固定为：Token 有效性 -> 状态/过期 -> 用户状态 -> 资源归属 -> 基础能力/扩展 Scope -> 套餐配额 -> 业务操作。Token 权限不能绕过框架用户状态、封禁状态、资源归属和套餐限制。

个人 Token 的基础能力不可关闭：

```text
upload:write      上传图片
links:read        获取自己上传图片的各种链接
media:delete      删除自己上传的图片
```

这些能力只对 Token 所属用户自己的媒体生效。即使 Token 包含 `media:delete`，也不能删除其他用户的图片。

可选扩展能力：

```text
media:read        查询自己媒体列表和详情
usage:read        查询自己的用量
webhook:manage    管理自己的 Webhook
```

### 12.5 Token 上传 API

```text
POST /api/v1/uploads
Authorization: Bearer <token>
Content-Type: multipart/form-data
Idempotency-Key: <client-generated-key>
```

表单字段：

| 字段 | 必填 | 说明 |
| --- | --- | --- |
| `file` | 是 | 图片文件 |
| `folder_id` | 否 | 用户自己的文件夹 |
| `album_id` | 否 | 用户自己的相册 |
| `visibility` | 否 | `private`、`link`、`public`，新上传默认 `public`；用户可在媒体详情改为 `private` 或 `link` |
| `expires_at` | 否 | 分享过期时间，必须满足套餐权益 |
| `title` | 否 | 图片标题 |
| `tags` | 否 | 标签数组或逗号分隔值 |

示例：

```bash
curl -X POST "https://img.example.com/api/v1/uploads" \
  -H "Authorization: Bearer fst_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" \
  -H "Idempotency-Key: upload-20270923-001" \
  -F "file=@./cover.png" \
  -F "visibility=link"
```

图片已完成处理时返回 `201 Created`：

```json
{
  "data": {
    "id": 1001,
    "status": "ready",
    "size_bytes": 183920,
    "mime_type": "image/png",
    "links": {
      "original": "https://img.example.com/i/abc/original.png",
      "url": "https://img.example.com/i/abc/image.png",
      "markdown": "![cover](https://img.example.com/i/abc/image.png)",
      "html": "<img src=\"https://img.example.com/i/abc/image.png\" alt=\"cover\">",
      "bbcode": "[img]https://img.example.com/i/abc/image.png[/img]"
    }
  },
  "request_id": "req_..."
}
```

处理尚未完成时返回 `202 Accepted`：

```json
{
  "data": {
    "id": 1001,
    "status": "processing",
    "status_url": "/api/v1/uploads/1001",
    "links": {}
  }
}
```

客户端必须轮询 `status_url`，不能在 `processing` 状态下猜测链接。处理失败返回 `status=failed`、稳定错误码和可重试提示。

### 12.6 Token 链接查询 API

```text
GET /api/v1/uploads/{id}
Authorization: Bearer <token>
```

只要媒体属于 Token 所属用户且 Token 有效，就可以返回该媒体的完整 `links`。获取自己上传图片的链接属于个人 Token 的基础能力，不要求用户额外配置 Scope。

如果请求的媒体不属于 Token 所属用户，统一返回 `MEDIA_ACCESS_DENIED`，不能通过增加 Scope 绕过归属校验。查询其他媒体列表则需要额外的 `media:read`，删除自己媒体使用内置的 `media:delete` 基础能力。

链接返回固定键集合（只保存和投递原图）：

```text
original
url
markdown
html
bbcode
```

不存在的 Variant 返回 `null`，不能返回失效 URL。不同套餐可限制原图、AVIF、自定义尺寸和批量链接，但基础 `url`、Markdown、HTML 和 BBCode 应保持 Free 可用。

## 13. 非功能要求

### 第一版目标

- 普通媒体列表首屏 API 目标 P95 小于 500 ms（不含图片对象下载）。
- 创建上传会话 API 目标 P95 小于 300 ms。
- 图片访问主链路不等待统计写入、缩略图处理或外部审核。
- 队列任务失败后可查询、可重试，不允许无状态丢失。
- 所有跨用户访问拒绝、配额拒绝和管理员破坏性操作可审计。

### 可观测性

- 每个请求有 request ID。
- 上传会话、媒体、任务、订单和分享链接使用业务 ID 串联日志。
- 记录队列等待时间、处理耗时、失败原因和重试次数。
- 统计存储使用量、对象清理积压、审核积压和异常流量。

## 14. 版本范围和优先级

| 优先级 | 内容 | 版本 |
| --- | --- | --- |
| P0 | Free 上传、媒体库、稳定链接、回收站、配额、Personal API Token、基础审核 | MVP |
| P1 | Creator/Pro、订单 fake、批量下载、防盗链、访问统计、广告免除 | V1 |
| P2 | 真实支付、自定义域名、S3/CDN、多 Provider、Webhook | V1.5 |
| P3 | 团队空间、企业套餐、AI 标签、社交功能、视频 | 后续 |

任何 P2/P3 功能都不能阻塞 P0 的免费闭环；如果存储成本或审核成本上升，优先调整额度和流量，而不是删除免费核心能力。

## 15. 会员到期、自动降级和通知

会员订阅到期后不删除用户数据，也不撤销用户账号。系统将原订阅标记为 `expired`，保留订单、支付流水和权益快照，并创建或激活配置的 fallback 套餐，默认是 `free`。

- 付费订单履约时根据价格快照的月付/年付周期和试用天数写入 `ends_at`；没有结束时间的管理员手动分配视为人工维护的长期订阅。
- 运行时每次解析订阅都会检查 `ends_at`，定时任务只负责批量补偿、提前提醒和发送邮件，不能作为唯一的过期判断机制。
- 到期后新的上传、API 速率、Token 数量、带宽、图片处理、广告和水印等能力按 fallback 套餐执行。
- 已有媒体、相册、文件夹、Token 和历史链接保留。超出 Free 存储时默认禁止继续上传，但允许查看、下载和删除；续费或清理空间后恢复上传。
- Personal API Token 不因套餐到期自动删除；请求仍经过当前套餐权益检查。管理员 RBAC 与会员套餐是两个独立维度。
- 到期前默认在 7、3、1 天发送事务性邮件，并写入站内通知；到期当天发送自动降级通知。邮件事件使用唯一键和重试状态，邮件服务关闭时不阻塞降级。
- `/admin/settings` 的“会员到期”设置管理邮件开关、fallback 套餐编码、宽限期、提醒天数和超额策略。宽限期用于续费提示与数据保留，不延长付费上传额度。
