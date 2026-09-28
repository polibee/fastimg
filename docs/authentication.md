# 认证设计

## 目标

项目是前后端分离架构，采用：

~~~text
短时效 JWT Access Token
+
HttpOnly Secure Refresh Token
+
Refresh Token Rotation
~~~

Goravel v1.18 同时支持 JWT 和 Session 驱动；本项目选择 JWT 作为 Admin API 的访问凭证，避免前端和后端必须共享页面 Session。

## Token 规则

~~~text
Access Token
- 短时效
- 前端仅保存在内存
- 请求使用 Authorization: Bearer

Refresh Token
- HttpOnly
- Secure
- SameSite 按部署拓扑配置
- 服务器端可撤销

默认时效：Access Token 为 `JWT_TTL=60` 分钟；持久 Refresh Token 为 `JWT_REFRESH_TTL=20160` 分钟（默认 30 天），该值同时控制 PostgreSQL 会话记录和浏览器 Cookie。前端会在 Access Token 到期前约 2 分钟自动轮换。`SESSION_LIFETIME=120` 是 Goravel Session 驱动的空闲时长，不是本项目 JWT 登录态的主要过期时间。
- Redis 保存会话和轮换状态
~~~

禁止把长期 Access Token 或 Refresh Token 写入 localStorage。

## 流程

~~~text
POST /api/v1/auth/login
        ↓
验证账号和密码
        ↓
返回 Access Token
写入 Refresh Token Cookie
        ↓
访问受保护 API
        ↓
Access Token 过期
        ↓
POST /api/v1/auth/refresh
        ↓
验证并轮换 Refresh Token
        ↓
返回新的 Access Token
~~~

登出时撤销 Refresh Token，并清除 Cookie。密码使用 Goravel Hash 能力，不自行实现加密或哈希。

## 认证边界

- Goravel Auth 负责身份验证基础能力；
- Admin Core 负责 JWT Claims、Refresh Token 会话和撤销；
- RBAC 负责角色、权限和数据范围；
- Controller/Service 边界必须再次验证权限；
- 前端路由守卫只负责用户体验，不能替代后端授权。

## Redis

Redis 用于：

- PostgreSQL 权威 Refresh Token 会话；
- Redis 镜像和高速读取；
- Token 撤销；
- 轮换重放检测；
- 登录限流；
- 验证码和临时状态。

## 失败场景

必须处理：

- Access Token 过期；
- Refresh Token 过期；
- Refresh Token 重放；
- 用户被停用；
- 用户角色被撤销；
- Redis 不可用；
- 登录失败限流；
- 多设备登录；
- 主动退出所有设备。

## API

~~~text
POST /api/v1/auth/login
POST /api/v1/auth/refresh
POST /api/v1/auth/logout
POST /api/v1/auth/logout-all
GET  /api/v1/auth/me
GET  /api/v1/auth/registration-policy
POST /api/v1/auth/register
GET  /api/v1/auth/verify-email?token=...
POST /api/v1/auth/resend-verification
~~~

## 当前实现状态

第一条可运行垂直链路已完成：

- `backend/app/models/user.go` 用户模型；
- `users` 表迁移；
- Goravel JWT 登录、当前用户、刷新和登出接口；
- Goravel Hash 密码校验；
- 管理员不再使用固定演示凭据；部署 CLI 通过 `admin:bootstrap` 生成一次性随机密码，开发 Seeder 仅在显式提供 `FASTIMG_SEED_ADMIN_EMAIL` 和 `FASTIMG_SEED_ADMIN_PASSWORD` 时创建账号；
- 前端登录页、内存 Access Token 和路由守卫。
- `roles`、`permissions`、`role_user`、`permission_role` RBAC 表；
- 管理员用户、角色、权限列表接口：
  `GET /api/v1/admin/users`、`GET /api/v1/admin/roles`、`GET /api/v1/admin/permissions`；
- `AdminUser` Seeder 在显式提供开发环境管理员变量时创建 `super-admin` 及基础管理权限；生产部署使用 `admin:bootstrap`，不执行固定账号 Seeder。生产初始化命令会创建或启用指定邮箱的 `super-admin`，补齐 FastImg 域权限并将管理员账号标记为已验证；命令输出会明确显示 `role: super-admin`；
- 角色创建、详情、编辑、删除和权限分配接口；
- 用户角色查询和绑定接口；
- 基于 `admin.users.view`、`admin.roles.manage`、`admin.permissions.manage` 的后端细粒度授权；
- API 错误统一返回稳定 `code`，前端按模块语言包渲染中文或英文。
- 登录和当前用户接口返回权限标识，前端据此隐藏无权菜单和操作，并在路由层显示统一无权页面；后端权限中间件仍是最终授权边界。

当前实现已接入独立的随机 Refresh Token：登录写入 HttpOnly Cookie，PostgreSQL 保存权威会话，Redis 保存镜像和高速副本；刷新时一次性消费并轮换，Redis 不可用时自动查询 PostgreSQL，重放返回 401，登出和 `logout-all` 会撤销并清除 Cookie。登录失败按“邮箱 + IP”组合限流，Redis 正常时使用 Redis 计数，Redis 不可用时切换 PostgreSQL；认证事件和关键用户/角色管理操作会写入 PostgreSQL 审计日志，并在管理端支持筛选、分页和详情查看；管理操作 metadata 只保存目标 ID、状态和权限 ID 等安全字段。生产 HTTPS 下 Cookie 使用 `Secure`；部署 CLI 的临时 HTTP 地址会根据 `APP_URL` 暂时关闭 `Secure`，切换正式 HTTPS 并重启后自动恢复。系统角色 `super-admin` 受到保护，最后一个具备管理权限的管理员不能被移除。

## 测试

必须验证登录、刷新、轮换、撤销、过期、重放检测、停用用户、权限变更和跨来源 Cookie 配置。

## 注册与登录安全策略

注册邮箱验证统一使用 `app/services/email` 的 EmailService。当前支持通用 SMTP、阿里云 DirectMail `SingleSendMail` 和 Resend API 三种传输方式，由后台 `/admin/settings` 的“邮件服务”分组选择并启用。开发阶段可以关闭 `email.enabled`，不要求填写真实凭证；开启 `auth.registration.email_verification_enabled` 前必须先配置一个可用的邮件服务。

SMTP、阿里云 AccessKey 和 Resend API Key 按系统设置密钥策略使用 `APP_KEY` 加密保存，设置列表接口只返回 `__configured__` 占位符。阿里云使用官方 `2015-11-23` API 的 HMAC-SHA1 请求签名，Resend 使用官方 `/emails` Bearer API，注册控制器不直接依赖供应商协议。

游客可以访问首页、套餐、发现页和注册页；只有上传、个人媒体、订单、Token 等用户数据能力需要认证。会员端和管理员端退出后都返回 `/`，认证接口失败也不会阻止本地退出导航。

公开认证接口：

| 方法 | 路径 | 作用 |
| --- | --- | --- |
| `GET` | `/api/v1/auth/registration-policy` | 返回注册、Turnstile、邮箱验证开关和公开 Site Key；不返回 Secret Key 或白名单域名 |
| `POST` | `/api/v1/auth/register` | 创建普通会员并自动分配 Free 订阅 |
| `GET` | `/api/v1/auth/verify-email?token=...` | 消费一次性邮箱验证 token |
| `POST` | `/api/v1/auth/resend-verification` | 重新发送邮箱验证链接 |

后台设置键位于 `auth.*`：`auth.registration.enabled` 控制注册、`auth.turnstile.enabled` 控制 Turnstile 总开关，`auth.turnstile.site_key` / `auth.turnstile.secret_key` 配置 Cloudflare Widget（Secret Key 使用 APP_KEY 加密），`auth.login.turnstile_enabled` 和 `auth.registration.turnstile_enabled` 分别控制登录与注册验证，`auth.registration.email_verification_enabled` 控制新用户邮箱验证，`auth.registration.email_whitelist_enabled` 与 `auth.registration.email_whitelist_domains` 控制后端域名白名单，`auth.registration.verification_expiry_minutes` 控制 5–1440 分钟的链接有效期。验证邮件重发保护由 `auth.registration.verification_resend_protection_enabled` 控制；`auth.registration.verification_resend_email_cooldown_seconds`、`auth.registration.verification_resend_ip_cooldown_seconds` 分别控制邮箱/IP 冷却时间，`auth.registration.verification_resend_daily_email_limit`、`auth.registration.verification_resend_daily_ip_limit` 控制 UTC 自然日发送上限。

Turnstile 必须由后端调用官方 Siteverify 接口完成最终判断，浏览器 token 不能直接作为可信结果。邮箱验证 token 只保存摘要、一次消费并过期；生产启用邮箱验证前必须配置 SMTP 并完成真实收信验收。
