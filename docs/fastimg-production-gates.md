# FastImg 生产门禁执行清单

本文件把“已经写代码”和“已经在目标生产环境验收”分开。没有外部凭证、目标域名、隔离恢复库或压测环境时，门禁必须保持 `blocked`，不能用本地 fake、配置存在或单元测试替代。

## 当前审计结论（2026-09-26）

| 门禁 | 当前状态 | 已有证据 | 仍缺什么 |
|---|---|---|---|
| 真实支付回调 | `blocked` | XCash、NOWPayments、PayPal Provider 离线签名/协议测试；统一回调幂等、金额和订单校验 | 各渠道真实沙盒或小额交易、真实公网 HTTPS 回调、重复回调/失败重试/对账记录 |
| TLS/反向代理 | `designed` | `deploy/nginx/fastimg.conf.example`；生产配置校验要求非回环 `APP_URL`、非 `disable` 的 DB SSL 和精确 CORS | 目标域名证书、代理可信头、证书续期、外网 HTTPS 冒烟与回滚 |
| PostgreSQL 备份 | `implemented` | `scripts/fastimg-db-backup.ps1` 使用自定义格式、无 owner/privilege，支持目录验证 | 持久化备份位置、加密/保留策略、异机或对象存储副本 |
| PostgreSQL 恢复 | `blocked` | 脚本拒绝恢复到源库并要求显式确认 | DBA 创建隔离目标库后，完成恢复、迁移版本、关键表计数和应用读写验收 |
| 依赖漏洞扫描 | `blocked` | 扫描脚本已 fail-closed | 当前环境没有 `govulncheck`；联网前端审计和扫描结果尚未执行 |
| 对象存储/CDN | `blocked` | Storage Provider 边界、防盗链、签名 URL 和 CDN 约束已有文档 | S3-compatible Provider、对象迁移/回源、CDN 缓存失效、带宽计量和恢复验证 |
| 压力测试 | `implemented` | `scripts/fastimg-load-smoke.ps1` 只访问公开 GET 状态接口，不产生上传/订单副作用 | 目标生产拓扑下的上传、媒体读取、回调、队列、429、P95/P99 和容量基线 |

本轮开发环境门禁刷新：`http://127.0.0.1:53085/api/v1/discovery/status` 低风险 GET 冒烟 `100/100` 成功，约 `283 req/s`；这只能证明本机公开状态接口可达，不能替代生产压测。依赖扫描按默认 fail-closed 执行，因缺少 `govulncheck` 且未在审批网络环境运行前端 `pnpm audit` 返回退出码 `2`，依赖门禁仍为 `blocked`。

## 1. 真实支付回调验收

每个渠道单独验收，不能把 XCash 成功推断为 NOWPayments 或 PayPal 成功。验收订单必须是专门的小额测试订单，完成后手动核对：

1. 创建订单和支付请求，保存订单号、Provider payment ID、请求时间和配置版本。
2. 在渠道侧完成真实沙盒或小额支付；回跳只验证页面返回，不能作为履约证据。
3. 确认公网 HTTPS Webhook 收到原始请求；保留脱敏后的请求 ID、事件 ID、HTTP 状态和服务端处理结果。
4. 验证错误签名、旧时间戳、重复事件、金额/币种不匹配、跨订单 Provider ID、超时重试都不会发放权益。
5. 验证合法事件只结算一次，订单、支付意图、支付交易、Webhook 事件和订阅用量能够关联查询。
6. 将证据写入发布记录，渠道状态才可以从 `implemented` 变成 `integrated`；完成失败恢复和对账演练后才是 `verified`。

当前仓库不自动向真实渠道发起支付、不生成伪造的生产回调，也不把 API Key 存入脚本或日志。

## 2. TLS 与反向代理

1. 使用 `deploy/nginx/fastimg.conf.example` 生成目标环境配置，替换域名、证书路径和后端端口。
2. 先完成 `nginx -t`、证书链检查、HTTP 到 HTTPS 跳转和外网 `curl -I`，再开启 HSTS。
3. `APP_URL` 使用最终 HTTPS 域名；`CORS_ALLOWED_ORIGINS` 只列出实际的前端 Origin，不使用 `*`。
4. 代理只转发 `Host`、`X-Real-IP`、`X-Forwarded-For`、`X-Forwarded-Proto`；应用不会把任意用户提交的 `X-Forwarded-*` 当成可信来源。
5. 生产前验证上传大小、超时、稳定图片链接和错误响应没有被代理缓存。

## 3. 备份与恢复

只读查看工具：

```powershell
pwsh -File .\scripts\fastimg-db-backup.ps1 -VerifyDump
```

脚本默认把备份写到系统临时目录，不会扩大仓库；生产环境必须改为加密且受访问控制的备份位置。恢复必须使用 DBA 创建的独立数据库，禁止使用 `fastimg_dev`：

```powershell
pwsh -File .\scripts\fastimg-db-backup.ps1 `
  -OutputPath D:\secure-backups\fastimg-20260926.dump `
  -RestoreDatabase fastimg_restore_20260926 `
  -ConfirmRestore -VerifyDump
```

恢复后执行迁移版本、用户/媒体/订单/用量/审计关键表计数和只读登录验收，再销毁临时恢复库。销毁目标库不是脚本默认行为，必须由 DBA 按保留策略处理。

## 4. 依赖扫描

```powershell
# 默认 fail-closed，不联网、不安装工具
pwsh -File .\scripts\fastimg-dependency-audit.ps1

# 在审批的联网构建环境执行前端生产依赖审计
pwsh -File .\scripts\fastimg-dependency-audit.ps1 -RunNetworkScan
```

Go 使用官方 `govulncheck ./...`；前端只扫描生产依赖。扫描工具缺失、联网审计未执行或发现 high/critical 均不能通过门禁。不得为了让扫描通过而随意升级依赖或改锁文件。

## 5. 对象存储与 CDN

当前 Local Provider 只适合开发和小规模验证。正式生产还需要：

- S3-compatible Provider 的 Put/Get/Delete/Exists/Copy/签名 URL 契约测试；
- 上传完成后数据库状态与对象存在性的一致性补偿；
- 原图、Variant、缩略图和删除回收站的生命周期策略；
- CDN 回源鉴权、签名/防盗链在缓存命中前生效，策略变化触发失效；
- 对象版本或备份、跨区域恢复、外链带宽入账和费用告警。

在这些证据完成前，不能把本地文件存储称为对象存储/CDN 生产方案。

## 6. 压力测试

先用无副作用公开状态接口验证服务是否可达：

```powershell
pwsh -File .\scripts\fastimg-load-smoke.ps1 -BaseUrl https://img.example.com -Requests 1000 -Concurrency 20
```

该脚本不是完整生产压测，只用于连通性和低风险基线。正式压测必须在隔离环境使用测试账户/媒体，覆盖上传、媒体读取、签名 URL、Webhook、队列和数据库连接池，并记录吞吐、P50/P95/P99、错误率、429、CPU、内存、数据库连接、Redis 队列和对象存储耗时。禁止对生产公开站点直接进行高并发写入测试。

## 发布判定

只有所有目标环境门禁达到 `verified`，并且回滚、备份恢复和告警联系人已记录，才能宣称正式生产发布。当前项目仍是“可继续开发/可做受控集成验收”，不是正式生产发布完成。
