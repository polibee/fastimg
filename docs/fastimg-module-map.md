# FastImg 与 Go Vue Admin 模块映射

这份文件把图床业务映射到当前仓库已经存在的目录和扩展入口。它是开发前的边界基线，不是新的框架层。

## 1. 已存在、直接复用的入口

| 能力 | 真实路径 | FastImg 使用方式 |
| --- | --- | --- |
| 业务模块 | `backend/app/modules/<name>` | 放资源、模型、Controller、Manifest、权限、菜单和模块路由边界 |
| 业务 Service | `backend/app/services/<domain>` | 放跨资源的配额、媒体、存储、审核和统计规则 |
| 通用资源 Registry | `backend/app/modules/admin/registry` | 通过生成专属 discovery 注册标准资源 |
| 资源生成器 | `backend/app/console` | 用 `admin:make-resource` 生成尚不存在的标准 CRUD 骨架；不要覆盖已手写的 plans 模块 |
| 资源写入扩展点 | `backend/app/core/resource/registry.go`、`backend/app/core/admin/controllers` | Manifest 可选 `WritePreparer` 在通用 create/update/bulk-update 落库前校验并规范化；领域规则仍放在对应 `services/<domain>` |
| 迁移 | `backend/database/migrations` | 新表统一写入此目录，人工审阅后执行 |
| Web 路由 | `backend/routes/web.go` | 注册认证后的自定义上传、分享、统计和开发者 API |
| 队列配置 | `backend/config/queue.go` | 复用 Redis queue，不实现内存降级 |
| 文件系统配置 | `backend/config/filesystems.go` | 注册 Local/兼容 Provider 所需的磁盘配置 |
| 定时任务 | `backend/bootstrap/schedule.go` | 注册回收站、孤儿对象、统计聚合和失败恢复任务 |
| OpenAPI | `backend/app/openapi` | 注册新增 API 契约和错误码 |
| 前端通用资源 | `admin/src/core/resource/pages` | 标准列表、表单、详情和 Action 页面 |
| 前端业务模块 | `admin/src/modules/<name>` | 上传器、媒体库、瀑布流、用量、审核和统计自定义页面 |
| 前端生成客户端 | `admin/src/generated/api.ts` | 通过 `pnpm run generate:api` 更新，不手工复制接口类型 |
| 审计与权限 | `backend/app/services/audit`、`backend/app/services/rbac` | 复用资源权限、数据范围、字段权限和审计 |

## 2. 资源模式选择

### Generic Resource

优先使用生成器：

- `plans`
- `plan_entitlements`
- `folders`
- `albums`
- `share_links`
- `api_tokens`
- `ad_slots`
- `ad_creatives`
- `reports`
- `storage_nodes`

这些资源主要是列表、表单、详情、筛选、导出和标准 Action。

### Custom Resource / Custom Page

必须使用业务页面：

- 上传工作台和分片进度
- 媒体网格/瀑布流
- 媒体详情和多 Variant 预览
- 用户用量与套餐对比
- 审核工作台
- 订单支付流程
- 统计图表
- 公共发现页

Custom Page 仍然必须使用框架的认证、权限、API 错误、通知、审计和 shadcn-vue 组件。

## 3. 路由注册规则

当前 Web 路由集中在 `backend/routes/web.go`。FastImg 自定义路由按以下边界注册：

```text
/api/v1/share-links/*   RequireAuthentication
/api/v1/tokens/*        RequireAuthentication (Token management)
/api/v1/uploads/*       Login session or Personal API Token
/api/v1/media/*         Login session or Personal API Token
/api/v1/quota            Login session or Personal API Token
/api/v1/billing/*       RequireAuthentication
/api/v1/analytics/*     RequireAuthentication
/api/v1/admin/*         existing ResourcePermission or business permission
/s/{token}               public handler + server-side share policy
```

不在 `backend/routes/web.go` 中直接查询数据库。每个自定义 Handler 只调用对应 Application Service。

## 4. 启动和注册边界

- `backend/bootstrap/app.go` 已统一注册 Seeders、Migrations、Schedule、Commands、Routing、Middleware、Providers 和 Config。
- FastImg 不新增第二个应用入口。
- 新迁移加入现有 migration discovery/注册机制。
- 新定时任务加入现有 `Schedule` 注册。
- 新资源由 `admin:make-resource` 更新生成 discovery；不得手工维护第二套 Registry。
- 自定义业务路由可以在 `routes/web.go` 接入，但 Controller/Service 必须位于对应业务模块。

## 5. 当前基础设施限制

- 默认数据库连接是 PostgreSQL；当前文档不把 MySQL 兼容性当作已验证能力。
- Redis 已用于 Cache、Queue、Refresh Token 和限流；图床不得静默回退内存。
- 当前文件系统默认提供 Local/Public 磁盘；S3-compatible 是新增 Provider 工作，不是现成完成能力。
- 当前队列默认连接 Redis；需要重试、恢复、履约或清理的长任务通过队列执行。为保持“上传成功立即返回链接”，当前上传确认仍同步完成；未来耗时图片变体处理必须沿用同一 Job 边界，不能把不可重试的重处理塞进 HTTP 请求。
- 当前项目的资源 Registry 更适合后台管理资源；用户媒体库和上传器不能强行套用通用 CRUD 页面。

### 5.1 当前 FastImg 管理资源

- `/admin/media`：跨用户媒体库，只读上传者、文件元数据和访问状态，管理员按 `admin.media.*` 权限执行查看；隐藏、恢复、审核通过、审核拒绝和永久删除必须走专用媒体动作，不能通过通用 CRUD 直接改写状态。每个动作写入 `admin.media.*` 审计记录。
- `/admin/folders`、`/admin/albums`：跨用户文件夹、相册管理；会员端 `/folders`、`/albums` 只操作当前登录用户自己的集合。
- `/admin/plans`：后台只管理计划权益和产品信息；结算内部使用 `plan_prices` 价格版本表，价格版本不绑定唯一支付网关，订单保存价格快照后不可被后续改价影响。开发阶段不注册独立价格版本后台页面。
- `/admin/orders`、`/admin/payment-transactions`、`/admin/payment-events`、`/admin/refunds`：订单、支付流水、网关事件和退款分别查询，履约与支付状态不混用。
- `/admin/settings`、`/admin/storage`、`/admin/statistics`、`/admin/media-access-logs`：站点配置、对象存储连接与本站侧用量、聚合指标和访问记录后台页面。对象存储设置按 Provider 独立保存；启用的连接才在 `/admin/storage` 生成统计卡片，统计来自 `storage_objects` 与允许访问记录，不冒充云厂商账单数据。设置密钥使用 `APP_KEY` 加密持久化，接口只返回占位符。
- `/admin/albums/:id/media`：管理员专用相册内容页。只允许把该相册所属用户自己的 `ready` 媒体加入相册；移除关系不删除媒体。所有变更复用集合 Service，并写入 `admin.albums.media.*` 审计。
- `/api/v1/admin/statistics/trends`：按日期返回用户、媒体、相册、订单、支付流水和下载带宽的真实时间桶；缺表按未启用指标处理，查询错误必须返回明确的统计不可用错误，不能静默伪造零值。

### 5.3 Redis 异步任务边界

- `fastimg.billing.fulfill-order`：支付回调完成数据库状态更新后投递履约任务；订单履约任务可重试 3 次，数据库中的履约任务记录是事实来源。
- `fastimg.media.recover-upload`：定时命令扫描超过租约阈值的 `processing` 上传会话并投递恢复任务；任务可重试 3 次，恢复失败不能把媒体伪造成 `ready`。
- `media:dispatch-recovery` 每 5 分钟运行一次，最多投递 100 个候选会话；Redis 不可用时保留数据库状态，下一轮继续扫描，不启用内存降级。
- 本阶段没有把实时上传确认改成后台队列，因为 C 端需要立即显示原图、缩略图、Markdown、HTML 和 BBCode 链接。队列运行验收仍需真实消费进程、失败重试和 Redis 重启恢复演练。
- `/admin/tasks` 是失败任务运维入口，不是通用 CRUD。列表使用框架 Failer，重新投递调用 `FailedJob.Retry()`；访问和重试分别受 `admin.tasks.view`、`admin.tasks.retry` 保护，重试成功写入 `queue.task.retry` 审计。
- 失败任务接口默认不返回原始 payload、异常堆栈和连接凭证，避免把任务内部数据扩散到前端；真实 Redis 消费、故障注入、失败重试和 Redis 重启恢复仍必须在发布前单独演练。

### 5.2 开发者 API 与公开 SEO

- Personal API Token 的最小权限固定为 `upload:write`、`media:read`、`media:delete`；Token 只允许当前用户的上传、图片列表/详情/链接和删除，不能调用会员结算或管理员 API。
- `/api/v1/*` 是版本化 API；`/api/upload`、`/api/images`、`/api/image/{id}` 是同一最小能力的客户端兼容别名，不复制业务逻辑。
- `plan_prices` 是结算使用的价格版本，不绑定唯一网关；结算页从已注册 Provider 列表中选择渠道，Provider 密钥只来自服务端设置的加密字段。开发阶段不暴露独立后台路由或 `admin.plan_prices.*` 权限，旧地址也不提供兼容入口，避免和“会员计划”形成重复菜单。
- `/sitemap.xml`、`/robots.txt` 和前端 `build:ssg` 为公开会员首页、套餐、发现页以及通过 `SSG_PUBLIC_ALBUM_IDS` 选择的公开相册提供 SEO 首屏；服务端只允许 `visibility=public` 的相册和公开媒体进入响应；私有媒体、Token、订单和后台路径不进入站点地图。

## 6. 第一阶段允许修改的基础文件

只允许为注册和契约扩展修改以下基础入口：

- `backend/routes/web.go`
- `backend/bootstrap/migrations.go`
- `backend/bootstrap/schedule.go`
- `backend/config/filesystems.go`
- `backend/app/openapi/*`
- `backend/app/core/resource/registry.go` 与 `backend/app/core/admin/controllers/*`：只允许添加通用 opt-in 扩展和调用边界，不放 FastImg 套餐/媒体业务规则
- 生成器维护的 discovery 文件
- `admin/src/generated/api.ts` 的生成结果

禁止把图床领域规则写入 `backend/app/core`、通用 Resource 页面、RBAC Service 或通用文件系统实现；可在通用 Resource Engine 增加中立的 opt-in 扩展接口，具体规则由业务 Service 提供。

## 7. 本地预览与端口隔离

- 前端 API 默认使用同源路径；Vite 开发服务器和预览服务器将 `/api`、`/sitemap.xml`、`/robots.txt` 代理到 `FASTIMG_BACKEND_URL`，当前 FastImg 本地默认后端为 `http://127.0.0.1:53085`。需要独立端口时，在启动 Vite 前通过进程环境设置该变量。
- 每个工作区启动前分别检查前端与后端候选端口，并显式传入两个不同的空闲端口；Vite 使用 `--strictPort`，端口被占用时应停止并重新选择，不能自动递增后误连其他项目。
- 后端使用进程级 `APP_HOST=0.0.0.0` 和 `APP_PORT=<backend-port>`；不要为方便预览覆盖或提交 `.env`，PostgreSQL/Redis 仍按用户配置管理。
- 开发服务需后台持有、日志落盘并记录 PID。WSL 环境向用户提供地址前，必须运行 `wslnet url <frontend-port>` 验证 Windows 侧访问；后端需同样验证其独立端口。
- `VITE_API_BASE_URL` 仅用于明确需要跨源 API 的部署；生产构建与部署必须显式验证 API 来源和 CORS/凭证策略。

## 8. 本地基线快照（2026-09-23）

以下只记录本次开发环境的观察结果，不是对其他部署环境的要求；检查期间没有启动服务、修改 `.env`、执行迁移或更改 Laragon 配置。

- `backend/.env` 当前缺失，仓库提供 `.env.example`；不能据此连接本机数据库或 Redis。
- Windows Laragon 主机的 PostgreSQL `5432`、Redis `6379` 均无监听；未发现可用于本次验收的 PostgreSQL/Redis 日志，因此未运行数据库集成或迁移。
- `go test ./...` 的非 Feature 包通过；`goravel/tests/feature` 在启动时因 `database.redis.default.host` 未配置导致 Schedule 初始化 panic，故全量 Go 测试未通过。
- 计划指定的 `pnpm exec vue-tsc --noEmit` 被 TypeScript 6.0.2 的 `TS5101` 阻断（`baseUrl` 弃用诊断）；项目现有 `pnpm run build` 中的 `vue-tsc -b` 与 Vite 构建通过。未为消除基线诊断擅自改动 TypeScript 配置。
- admin Node 测试基线为 19/20；唯一失败为 `resource-actions.test.ts` 中 `kind: unknown` 的既有预期与实现不一致，代码和测试相对 HEAD 均未修改。i18n 专项测试为 3/3 通过。

## 9. 当前开发态（2026-09-23）

本节是基线之后的开发快照；第 8 节保留为开发开始时的历史记录。

- `backend/.env` 已在本机配置并被 Git 忽略；已核对非敏感字段 `DB_CONNECTION=postgres`、`DB_DATABASE=fastimg_dev`。不要读取、打印或提交密钥字段。
- Windows 侧 PostgreSQL `5432`、Redis `6379` 可连接。WSL 内的 CLI 启动会把 Redis `127.0.0.1:6379` 当作 WSL 自身地址并收到拒绝；不得据此修改 Laragon 服务或代理配置。WSL Redis 接入需要使用项目认可的 WSL 网络配置另行处理。
- `fastimg_dev` 只读 `migrate:status` 显示 plans、media、advertising、folders、albums 迁移均为 Ran；本轮用户明确授权的新迁移只覆盖 advertising、folders、albums 和对应管理权限。
- 本地开发预览采用不同端口：Vite 前端 `53084`，Go API 后端 `53085`；浏览器请求经 Vite `/api` 代理，机器可读的 sitemap/robots 文件也经 Vite 转发到后端。WSL 网络可用时，向用户提供地址前运行 `wslnet url 53084`；若 WSL 返回 `E_ACCESSDENIED`，必须改用 Windows 侧 HTTP 验证并明确记录限制。
- 管理员登录已在浏览器会话验证；管理资源位于 `/admin/**`。会员端由独立 `MemberShell` 承载，`/`、`/media`、`/media/:id`、`/folders`、`/albums`、`/plans` 均为平铺会员 URL；会员上传、个人中心、Token、套餐用量和发现瀑布流按阶段继续交付。
- 最新验证：`go test ./... -count=1` 全通过；`pnpm run build` 通过；`node --test tests/fastimg-i18n.test.mjs` 4/4 通过；完整 admin Node 测试 23/24，唯一失败是通用 `resource-actions` 对未知 action kind 的既有行为/测试不一致。

## 9.1 当前运行态补充（2026-09-27）

- 当前标准前端配置是 `admin/vite.config.ts`，不是历史验证记录中的 `vite.fastimg-isolated.config.mjs`；历史端口记录保留在本文件和阶段台账中，不应作为当前启动命令使用。
- 管理端 SEO 按钮通过 API 后端基址打开 `/sitemap.xml`、`/robots.txt`，避免开发端口把 XML/文本请求回退为 Vue SPA 页面；生产环境仍应由反向代理把这两个公开路径转发到后端。

## 10. 前端应用边界与 SSR 演进

当前 `admin/` 是一个 Vue/Vite 工程，历史上同时承载 C 端和管理员页面。运行时已经通过 `/` 与 `/admin/**`、`MemberShell` 与 `AdminShell`、会员权限与管理员权限完成隔离，但代码入口容易让人误以为 C 端属于后台。

本阶段先做逻辑边界抽取，不直接搬迁目录：

```text
admin/src/
├─ apps/
│  ├─ member/routes.ts   # C 端和公开路由入口
│  └─ admin/routes.ts     # 管理员路由入口
├─ core/                  # 共享布局基础、Resource Engine、基础页面
├─ components/            # 共享 UI
├─ lib/、stores/、i18n/   # 共享基础设施
└─ modules/               # 领域页面和 API 模块
```

边界抽取的要求：

- `apps/member` 不得导入 `AdminShell`、管理员导航、管理员 Resource 页面或后台权限页面。
- `apps/admin` 不得把管理员页面挂到平铺会员 URL；后台页面必须位于 `/admin/**` 并继续执行管理员权限守卫。
- API、认证基础设施、i18n、shadcn-vue 组件和设计令牌可共享；业务页面和应用 Shell 不共享。
- 当前阶段不改变 URL、不复制 API 客户端、不增加第二套后端入口。

该边界是独立 `web/` 和 SSR/SSG 的前置条件，但不会自动产生 SSR：

- 当前 `build:ssg` 服务于已注册的公开首页、套餐、发现页和可选公开相册静态预渲染；公开相册必须通过真实 API 校验后才生成 `/a/:id/`，不能把相册 ID 当作公开授权。
- 真正 SSR 还需要独立 Web 入口、服务端渲染运行时、请求级数据加载、head/SEO 处理和部署适配。
- `admin/` 继续作为 SPA，不参与公开页面服务端渲染。
- 私有媒体、Token、订单和 `/admin/**` 不得进入公开预渲染、站点地图或公共缓存。

当公开页面需要独立缓存、CDN、SSR 或独立发布时，再将 `apps/member` 抽取为同层 `web/`，并把稳定共享能力迁移到 `packages/`；禁止直接复制整个 `admin/src`。
