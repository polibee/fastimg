# FastImg 与 Go Vue Admin 模块映射

这份文件把图床业务映射到当前仓库已经存在的目录和扩展入口。它是开发前的边界基线，不是新的框架层。

## 1. 已存在、直接复用的入口

| 能力 | 真实路径 | FastImg 使用方式 |
| --- | --- | --- |
| 业务模块 | `backend/app/modules/<name>` | 放资源、模型、Controller、Manifest、权限、菜单和模块路由边界 |
| 业务 Service | `backend/app/services/<domain>` | 放跨资源的配额、媒体、存储、审核和统计规则 |
| 通用资源 Registry | `backend/app/modules/admin/registry` | 通过生成专属 discovery 注册标准资源 |
| 资源生成器 | `backend/app/console` | 用 `admin:make-resource` 生成套餐、文件夹、相册等 CRUD 骨架 |
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
- 当前队列默认连接 Redis；图片处理、审核、清理和统计必须通过队列，不能在 HTTP 请求内同步完成。
- 当前项目的资源 Registry 更适合后台管理资源；用户媒体库和上传器不能强行套用通用 CRUD 页面。

## 6. 第一阶段允许修改的基础文件

只允许为注册和契约扩展修改以下基础入口：

- `backend/routes/web.go`
- `backend/bootstrap/migrations.go`
- `backend/bootstrap/schedule.go`
- `backend/config/filesystems.go`
- `backend/app/openapi/*`
- 生成器维护的 discovery 文件
- `admin/src/generated/api.ts` 的生成结果

禁止把图床领域规则写入 `backend/app/core`、通用 Resource 页面、RBAC Service 或通用文件系统实现。
