# FastImg

FastImg 是面向开发者、站长和内容创作者的会员制媒体托管平台，提供稳定图片链接、会员媒体库、图片处理、Personal API Token、套餐权益、支付渠道、内容治理、防盗链和管理员运营后台。

项目基于 Go Vue Admin 底层框架开发，技术栈为 Goravel 1.18、Go、PostgreSQL、Redis、Vue 3、TypeScript、Vite 和 shadcn-vue。认证、RBAC、Resource、迁移、队列、审计、多语言和 API Client 等底层能力直接复用，不重复实现通用框架能力。

[![RackNerd VPS](https://img.shields.io/badge/RackNerd-VPS-2563eb?style=for-the-badge)](https://my.racknerd.com/aff.php?aff=7572)
[![Vast.ai GPU Cloud](https://img.shields.io/badge/Vast.ai-GPU%20Cloud-7c3aed?style=for-the-badge)](https://cloud.vast.ai/?ref_id=91181)

## 文档

- [English README](./README.md)
- [部署文档](./docs/fastimg-deployment.md)
- [生产门禁](./docs/fastimg-production-gates.md)
- [产品设计](./docs/fastimg-product-design.md)
- [阶段性开发计划](./docs/fastimg-stage-development-plan.md)
- [开发者 API](./docs/fastimg-developer-api.md)
- [支付网关设计](./docs/fastimg-payment-gateway.md)
- [前端设计](./docs/fastimg-frontend-design.md)
- [底层平台架构和路线图](./docs/README.md)

## 产品能力

### C 端会员功能

- 首页、套餐和发现页允许游客访问；上传和个人操作必须登录。
- 提供免费服务。付费套餐可以配置存储、单文件大小、上传次数、API、流量、处理次数、Token 数量、广告和水印权益。
- 图片上传完成后直接返回原图、Markdown、HTML、BBCode 和直链等绝对地址。
- 个人媒体库支持公开、未列出和私有图片，支持文件夹、相册、回收站、恢复、永久删除和访问控制。
- Personal API Token 固定只授予最小权限：`upload:write`、`media:read`、`media:delete`；过期时间支持自定义或永久。
- API Token 只能上传、获取本人图片列表/详情/链接和删除本人图片，不能调用管理员 API。
- 多个支付渠道可以同时启用；每笔订单选择一个渠道，默认网关只作为推荐，不会限制其他已配置渠道。

### 管理员后台

- 管理端统一使用 `/admin/**`，C 端使用 `/`、`/media`、`/plans`、`/discover`、`/folders` 和 `/albums`，两套导航和职责分离。
- 根据 RBAC 管理会员、套餐、广告位、媒体、相册、文件夹、订单、支付事件、退款、举报、访问记录、统计、设置、Token 和审计日志。
- 管理员媒体操作使用专用动作：隐藏、恢复、审核通过、审核拒绝和永久删除；禁止通过通用媒体 CRUD 直接修改状态，并且每个动作写入审计日志。
- 采用事后治理：用户上传图片默认可用，举报和管理员审核可以后续隐藏或拒绝，违规内容不会继续进入发现页。
- 支付渠道独立配置，凭证加密保存；支持启停、回调、订单快照、支付流水、Webhook 事件、履约和退款。
- 设置模块包含 SEO、站点地图、robots、邮箱验证、自定义代码、水印、上传限制和支付渠道配置。

### 开发和运维基础

- JWT 登录、刷新、退出，PostgreSQL 权威 Refresh Token 存储，Redis 用于加速、限流和队列。
- Resource Manifest 驱动列表、创建、编辑、详情、权限、菜单、搜索、筛选、关系和批量操作。
- 请求/响应审计和敏感字段脱敏，支持保留期清理。
- Redis 异步恢复和履约任务具有失败重试边界，PostgreSQL 是业务事实来源。
- 公开首页、套餐和发现页支持 SSG 首屏生成；当前没有启用完整 SSR。
- 本地存储适合开发和受控验收，生产仍需完成对象存储/CDN 和恢复验证。

## 本地运行

项目依赖由 Laragon 管理。启动 PostgreSQL 和 Redis 后执行；后端不得静默切换到内存实现。

~~~powershell
cd backend
go run .
~~~

另开终端启动管理面板：

~~~powershell
cd admin
pnpm install
pnpm dev
~~~

开发端口示例：

- Go API：http://127.0.0.1:3000
- Go/Vue C 端和管理面板：http://127.0.0.1:5180
- 管理端路由：http://127.0.0.1:5180/admin

同时运行多个项目时，每个项目都必须分配独立的前端和后端端口，Vite 使用 `strictPort`；向用户提供地址前验证真实端口，禁止端口被占用后自动递增并误连到其他项目。

数据库迁移必须先审阅生成文件，再按项目流程手动执行；生成器不会静默执行迁移。

## 资源生成

~~~powershell
cd backend
go run . admin:make-resource announcements --fields="title:text:required,status:select:required:draft=Draft|published=Published"
~~~

资源生成是面向业务开发的主流程：生成后由后台面板自动发现并展示。生成器不会覆盖人工文件；复杂业务应放入对应的 backend/app/modules/<module> 和 admin/src/modules/<module>，不要强行套用通用 Resource。

## 验证

~~~powershell
cd backend
go test ./...

cd ..\admin
pnpm exec vue-tsc --noEmit
pnpm run build
~~~

生成公开 SEO 静态页面：

~~~powershell
cd admin
pnpm run build:ssg
~~~

## 部署

Docker 部署和 Linux 源码部署脚本位于 [`deploy/`](./deploy/)，使用方式见[部署文档](./docs/fastimg-deployment.md)。两种方式只部署 FastImg 应用，PostgreSQL 和 Redis 必须使用外部服务，不会启动数据库容器。

~~~bash
cp deploy/docker/fastimg.env.example deploy/docker/fastimg.env
# 填写生产 APP_KEY、JWT_SECRET、APP_URL、CORS、PostgreSQL 和 Redis 配置。
deploy/docker/deploy.sh --check
deploy/docker/deploy.sh
~~~

部署脚本不会把真实支付回调、TLS、隔离备份恢复、依赖扫描、对象存储/CDN 或生产压力测试伪装成已通过；正式上线前必须逐项完成[生产门禁](./docs/fastimg-production-gates.md)。

## 目录约定

- 后端基础能力：backend/app/core
- 后端业务模块：backend/app/modules/<module>
- 后端 Service：backend/app/services/<domain>
- 后端命令：backend/app/console/<domain>
- 前端共享组件：admin/src/components
- 前端基础设施：admin/src/core
- 前端业务页面和组件：admin/src/modules/<module>

## 项目状态

FastImg 当前适合继续开发和受控集成验收，暂不宣称正式生产就绪。正式发布前必须在目标环境完成真实支付回调、TLS/反向代理、备份恢复、依赖漏洞扫描、对象存储/CDN、压力测试、监控和回滚验收。

生产部署前还需复核密钥、数据库、Redis、反向代理、日志保留、备份、监控、迁移、支付凭证和管理员初始化流程。本 README 是项目入口，不替代目标环境验收。
