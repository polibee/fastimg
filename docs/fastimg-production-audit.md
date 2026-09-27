# FastImg 生产审计与安全门禁

## 本轮审计范围

本轮覆盖后端启动配置、CORS、认证与会员 API 边界、媒体访问和防盗链、支付订单/支付意图/回调幂等、审计日志脱敏、文件存储路径和开发磁盘清理工具，并补充了备份、隔离恢复、依赖扫描、只读并发冒烟和 TLS 反向代理模板。未执行真实支付、生产发布、用户媒体删除或外部服务写入。

## 2026-09-26 门禁证据

- Laragon PostgreSQL 18 客户端工具可用，`fastimg_dev` 只读连接成功。
- `scripts/fastimg-db-backup.ps1 -VerifyDump` 已生成临时目录自定义格式备份，约 0.75 MB，`pg_restore --list` 读取成功；没有恢复到任何数据库。
- 当前 `fastimg_app` 数据库账号没有创建数据库权限，因此恢复演练仍需 DBA 先创建隔离目标库，不能把源库当作恢复目标。
- `scripts/fastimg-load-smoke.ps1` 对公开发现状态接口执行 10 请求、并发批次 2，10/10 返回 HTTP 200；这只是低风险连通性证据，不是生产压测。
- 当前环境未发现 `govulncheck`；Go 漏洞扫描和联网前端生产依赖审计均未执行，依赖门禁保持阻断。

## 已修复

### 生产配置

- PostgreSQL `sslmode` 不再硬编码，使用 `DB_SSLMODE`；生产环境禁止为空或 `disable`。
- 生产启动前检查 `APP_DEBUG=false`、`APP_URL` 为非回环绝对 URL、`APP_KEY` 和 `JWT_SECRET` 至少 32 个字符。
- CORS 使用 `CORS_ALLOWED_ORIGINS` 的精确 Origin 白名单；生产没有白名单时拒绝跨域，禁止 `*`。
- 框架 CORS 兜底配置也不再使用通配符。

### 支付与订单

- 同一用户的支付幂等键必须绑定同一订单、同一渠道；复用到其他订单或渠道返回 `PAYMENT_IDEMPOTENCY_CONFLICT`。
- 取消订单改为带状态条件的原子更新，支付回调和取消请求并发时不会把已付款订单改成已取消。
- 支付交易以 `provider_code + provider_transaction_id` 去重；重复成功回调只复用已结算事实，跨订单/支付意图复用会拒绝。
- 成功处理的回调事件写入 `processing_status=processed` 和 `processed_at`，便于后台对账。

### 审计与运维

- HTTP 审计使用递归脱敏，密码、访问令牌、刷新令牌、API Key、支付密钥和 Cookie 不进入审计详情。
- `scripts/fastimg-disk-audit.ps1` 已修复编码解析错误，默认只读；只有显式 `-Apply` 才会清理可重建缓存，并保护 `backend/storage/fastimg` 和日志目录。

### 邮箱验证重发保护

- 未登录的 `POST /api/v1/auth/resend-verification` 现在经过统一的验证邮件重发限流 Service；保护默认开启，缓存不可用时拒绝发送而不是绕过限流。
- 后台 `/admin/settings` 的“注册与登录”组可配置保护开关、单邮箱冷却秒数、单 IP 冷却秒数、单邮箱每日上限和单 IP 每日上限。
- 默认值为：同一邮箱 60 秒、同一 IP 10 秒、单邮箱每日 5 次、单 IP 每日 20 次；参数由后端再次做安全范围归一化，不能通过表单写入无限制或非正数。
- 触发限制返回 `AUTH_VERIFICATION_RESEND_RATE_LIMITED` 和 `retry_after_seconds`；未知邮箱仍返回通用成功响应，避免通过重发接口枚举账号。

### 2026-09-26 功能关联修复

- 上传来源现在显式区分会员网页登录和 Personal API Token：只有 Token 上传计入 `monthly_api_uploads`；两者仍共同复用媒体处理、存储、用量和链接服务。管理员保留上传和水印兜底能力，不因缺失会员订阅快照被错误拦截。
- 广告投放同时校验套餐 `ads_enabled`、广告 `plan_code`、开始时间和结束时间；过期或不匹配的广告不会进入会员接口响应。
- 举报创建与发现页采用同一事后治理规则：媒体上传即为 `public + approved`，只有明确 `rejected` 的媒体不再接受公开举报/发现投放；隐藏、恢复、通过、拒绝和永久删除继续走管理员专用动作并写审计。
- 管理端媒体前端元数据已与后端 Manifest 对齐，状态、可见性和审核状态不再进入通用编辑/删除；只能通过专用媒体动作执行生命周期变更。
- 假支付渠道默认仅在非 production 环境启用；生产环境必须显式配置真实渠道并完成回调验收，避免配置缺失时误把 fake 当成可上线支付。

## 生产部署必填

```dotenv
APP_ENV=production
APP_DEBUG=false
APP_URL=https://img.example.com
APP_KEY=<至少32字符的独立密钥>
JWT_SECRET=<至少32字符的独立密钥>
DB_CONNECTION=postgres
DB_SSLMODE=verify-full
CORS_ALLOWED_ORIGINS=https://img.example.com,https://admin.example.com
```

如果前端和 API 同域部署，`CORS_ALLOWED_ORIGINS` 可以只填写实际需要跨域的管理端/客户端 Origin；不要填写 `*`。

## 验收命令

```powershell
# 只读检查磁盘，默认不会删除任何目录
powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\fastimg-disk-audit.ps1

# 后端单元/集成测试
powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\fastimg-go.ps1 test ./... -count=1

# 生产门禁相关测试
powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\fastimg-go.ps1 test ./app/services/production ./app/http/corspolicy ./app/services/billing -count=1
```

## 尚未宣称完成的生产门禁

- 真实 PayPal、NOWPayments、XCash 生产凭证与真实回调尚未在本地验收，不把沙盒或配置存在性当作支付成功证明。
- 当前本地文件存储尚未完成对象存储迁移、CDN、备份恢复和跨节点一致性验收。
- 仍需在发布环境补充反向代理 TLS、可信代理头配置、隔离数据库恢复演练、依赖漏洞扫描、真实业务压测和告警接入。

详细命令和发布判定见 [fastimg-production-gates.md](./fastimg-production-gates.md)。
