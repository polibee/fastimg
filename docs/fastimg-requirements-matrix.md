# FastImg 需求矩阵与验收目录

本文档把产品设计转换为可追踪的需求编号。需求编号应出现在 Issue、提交说明、测试名称和验收报告中。

## 1. 业务需求

| 编号 | 需求 | 优先级 | 验收结果 |
| --- | --- | --- | --- |
| FR-001 | 新注册用户自动获得 Free 套餐 | P0 | 查询订阅返回 Free 及正确权益 |
| FR-002 | 用户可以上传允许格式的图片 | P0 | 上传完成后得到媒体记录和稳定资源标识 |
| FR-003 | 上传必须检查大小、格式、像素和用户配额 | P0 | 任一限制超出时拒绝并返回稳定错误码 |
| FR-004 | 用户只能管理自己的媒体 | P0 | 详情、列表、导出、批量和删除均拒绝越权 |
| FR-005 | 图片异步生成缩略图和中图 | P0 | 处理成功产生 Variant，失败可重试 |
| FR-006 | 用户可以复制 URL、Markdown、HTML 和 BBCode | P0 | 复制内容指向正确媒体 Variant |
| FR-007 | 用户可以删除和恢复媒体 | P0 | 删除进入回收站，恢复后重新可见 |
| FR-008 | 用户可以创建 Personal API Token | P0 | 完整 Token 只显示一次，撤销后立即失效 |
| FR-009 | Token API 上传复用网页上传和配额规则 | P0 | 相同文件通过 API 和网页得到一致业务结果 |
| FR-025 | Token 上传响应返回各种图片链接 | P0 | ready 时返回原图、缩略图、中图、WebP、AVIF、URL、Markdown、HTML、BBCode |
| FR-026 | Token 具备管理自己媒体的基础能力 | P0 | 默认可以获取自己图片链接和删除自己图片，不能操作他人媒体 |
| FR-027 | Token 支持过期、禁用、轮换和最近使用信息 | P0 | 过期/禁用 Token 立即拒绝 |
| FR-028 | Token 上传支持幂等键 | P0 | 客户端重试不创建重复媒体、不重复扣配额 |
| FR-029 | Token 支持批量上传 | P1 | 每个文件独立返回 ready/processing/failed 结果 |
| FR-030 | Token 可查询自己的媒体列表 | P1 | 需要 `media:read`，返回统一分页和媒体链接 |
| FR-031 | 开发者 API 采用版本化兼容规则 | P0 | `/api/v1`、链接键和错误码保持向后兼容 |
| FR-032 | 公开 ready 图片自动进入发现页并支持事后治理 | P0 | 普通上传不等待审核；隐藏、拒绝或设为私有后移除，恢复后重新展示 |
| FR-033 | 用户可以提交图片举报 | P0 | 举报进入后台队列并记录原因 |
| FR-034 | 管理员可以处理违规图片 | P0 | 支持隐藏、删除、恢复和处理记录 |
| FR-035 | 管理员可以处理违规账户 | P0 | 支持限制上传、暂停和封禁 |
| FR-036 | 违规处理必须可审计和可追溯 | P0 | 记录原因、证据、操作者、时间和处罚期限 |
| FR-037 | 发现页停用不影响私有媒体 | P0 | 私有访问和用户自己的分享策略保持有效 |
| FR-038 | 相同用户对同一媒体的重复举报可去重 | P0 | 返回原举报记录，不重复创建任务 |
| FR-039 | 举报阈值触发复核而非直接永久封禁 | P0 | 自动动作可恢复，永久处罚必须人工确认 |
| FR-040 | 被处罚用户可以提交申诉 | P1 | 申诉有状态、复核人、结果和通知 |
| FR-041 | 治理权限分级 | P0 | 查看、处理、处罚和发现页设置可独立授权 |
| FR-042 | 防盗链模式 | P1 | 支持关闭、Referer 白名单、签名 URL、混合模式 |
| FR-043 | 防盗链域名 | P1 | 用户可维护归一化域名白名单，管理员可查看和处置异常域名 |
| FR-044 | 签名 URL | P1 | 支持资源级签名链接、有限过期时间和签名校验 |
| FR-045 | 防盗链失败隔离 | P1 | 拒绝时不得泄露原图内容、对象存储地址或可绕过的重定向 |
| FR-046 | CDN 一致性 | P1 | CDN 缓存、回源和媒体处理链路不得绕过防盗链策略 |
| FR-047 | 支付网关抽象 | P1 | 业务模块不直接依赖具体支付渠道，渠道可替换 |
| FR-048 | 订单和支付流水 | P1 | 订单、支付意图、支付流水和回调事件可追踪且不可覆盖 |
| FR-049 | 支付回调幂等 | P1 | 重复通知不重复确认收款、激活订阅或增加权益 |
| FR-050 | 订阅权益履约 | P1 | 支付成功、订阅激活、配额刷新和广告规则变更可追踪 |
| FR-051 | 退款和拒付 | P2 | 支持部分退款、全额退款、退款失败和拒付流水 |
| FR-052 | 支付失败补偿 | P1 | 收款成功但权益发放失败时可查询并人工重试 |
| FR-053 | 支付渠道路由 | P1 | 按货币、地区、渠道状态和站点配置选择可用支付渠道 |
| FR-054 | 价格与优惠快照 | P1 | 订单锁定价格、货币、套餐版本和优惠计算结果 |
| FR-055 | 账单与收据 | P1 | 用户可查看订单账单摘要，历史账单不可被当前配置改写 |
| FR-056 | 支付对账 | P2 | 渠道流水与平台流水可匹配并处理差异 |
| FR-057 | 支付状态展示 | P1 | 支付中、履约中、成功、失败和退款状态在用户端可追踪 |
| FR-058 | 加密货币支付渠道 | P2 | Payment Gateway 支持资产、网络、报价和链上支付状态 |
| FR-059 | 链上确认与重组 | P2 | 按网络确认数发放权益，链重组和双花风险不自动放行 |
| FR-060 | 加密支付差额 | P2 | 少付、多付、过期报价、错误网络和重复交易进入明确处理流程 |
| FR-061 | 加密支付风控 | P2 | 资产、网络、地区和风险状态可配置，风险订单进入人工复核 |
| FR-010 | 用户可以查看存储、流量和上传用量 | P0 | 用量来自服务端流水，非前端计算 |
| FR-011 | 管理员可以按用户查看媒体 | P0 | 管理员权限和审计日志均生效 |
| FR-012 | 公开图片进入审核流程 | P0 | 未审核图片不会进入发现查询 |
| FR-013 | 用户可以举报公开媒体 | P0 | 举报生成任务并能在后台处理 |
| FR-014 | 管理员可以配置套餐权益 | P0 | 修改新订阅权益不影响历史订阅快照 |
| FR-015 | Free 用户可以看到基础广告 | P1 | 套餐权益决定广告是否展示 |
| FR-016 | Creator/Pro 支持更大存储和流量 | P1 | 升级后立即应用新额度 |
| FR-017 | 套餐降级不删除已有媒体 | P1 | 超出新额度时只限制新增资源 |
| FR-018 | 用户可以创建密码或过期分享链接 | P1 | 服务端强制校验密码和过期时间 |
| FR-019 | Pro 支持自定义域名和防盗链 | P2 | 域名和 Referer 规则在服务端生效 |
| FR-020 | Pro 支持 Webhook | P2 | 事件投递可重试并记录失败 |
| FR-021 | 用户可以查看媒体访问统计 | P1 | 统计异步记录且不阻塞访问 |
| FR-022 | 管理员可以查看运营、存储和收入统计 | P1 | 按时间范围和权限返回聚合数据 |
| FR-023 | 支付回调可幂等地激活订阅 | P2 | 重复回调不重复激活或延长周期 |
| FR-024 | 系统可以清理过期和孤儿对象 | P1 | 清理失败可重试，不错误释放额度 |

## 2. 非功能需求

| 编号 | 需求 | 目标 |
| --- | --- | --- |
| NFR-001 | 权限隔离 | 所有资源查询和写操作服务端执行归属校验 |
| NFR-002 | 可恢复性 | 上传、处理、清理、Webhook 失败可查询和重试 |
| NFR-003 | 幂等性 | 上传完成、支付回调、Webhook 和用量事件支持幂等 |
| NFR-004 | 可观测性 | 请求、媒体、任务、订单和对象可通过业务 ID 串联 |
| NFR-005 | 隐私 | 默认清理 GPS/EXIF，私有对象不直接公开 |
| NFR-006 | 性能 | 创建上传会话 P95 < 300ms；媒体列表 P95 < 500ms |
| NFR-007 | 安全 | Key、Token、密码和支付敏感数据不出现在日志 |
| NFR-008 | 成本控制 | 大型处理、流量和批量下载受套餐和限流约束 |
| NFR-009 | 数据一致性 | 额度、对象状态、媒体状态和清理状态可追踪 |
| NFR-010 | 兼容性 | 第一版以现有 PostgreSQL/Redis 配置作为验证基线 |
| NFR-011 | 财务一致性 | 金额使用最小货币单位整数，账务变化追加流水，不覆盖原记录 |
| NFR-012 | 渠道隔离 | 支付渠道异常不得破坏 Free 服务、已有订阅和媒体访问 |

## 3. API 目录

### 3.1 用户端 API

| 方法 | 路径 | 用途 | 认证 |
| --- | --- | --- | --- |
| GET | `/api/v1/me/usage` | 当前用量和权益 | 登录 |
| GET | `/api/v1/me/usage/ledger` | 用量流水 | 登录 |
| GET | `/api/v1/media` | 媒体列表 | 登录 |
| GET | `/api/v1/media/{id}` | 媒体详情 | 登录/分享策略 |
| PATCH | `/api/v1/media/{id}` | 更新标题、标签、可见性 | 登录 |
| DELETE | `/api/v1/media/{id}` | 移入回收站 | 登录 |
| POST | `/api/v1/media/{id}/restore` | 恢复媒体 | 登录 |
| POST | `/api/v1/media/{id}/share-links` | 创建分享链接 | 登录 |
| GET | `/api/v1/share-links` | 分享链接列表 | 登录 |
| DELETE | `/api/v1/share-links/{id}` | 撤销分享链接 | 登录 |
| GET | `/s/{token}` | 公开/密码分享访问 | 分享策略 |

### 3.2 开发者 API

| 方法 | 路径 | 用途 |
| --- | --- | --- |
| POST | `/api/v1/tokens` | 登录用户生成 Personal API Token |
| GET | `/api/v1/tokens` | Token 列表和最近使用信息 |
| DELETE | `/api/v1/tokens/{id}` | 撤销 Token |
| POST | `/api/v1/tokens/{id}/rotate` | 轮换 Token |
| POST | `/api/v1/uploads` | 单文件上传；支持登录会话或 Personal API Token |
| POST | `/api/v1/uploads/batch` | 批量上传；支持登录会话或 Personal API Token |
| GET | `/api/v1/uploads/{id}` | 查询处理状态和各种链接 |
| POST | `/api/v1/uploads/sessions` | 创建分片上传会话 |
| POST | `/api/v1/uploads/sessions/{id}/parts` | 上传分片 |
| POST | `/api/v1/uploads/sessions/{id}/complete` | 完成分片上传 |
| GET | `/api/v1/uploads/sessions/{id}` | 查询分片上传状态 |
| DELETE | `/api/v1/uploads/sessions/{id}` | 取消分片上传 |
| GET | `/api/v1/media` | 当前用户媒体列表，需要 `media:read` |
| DELETE | `/api/v1/media/{id}` | 删除当前用户媒体，属于 Token 基础能力 |
| GET | `/api/v1/quota` | 当前用户配额 |
| POST | `/api/v1/webhooks/test` | 测试 Webhook |
| GET/PUT | `/api/v1/media/{id}/hotlink-policy` | 查询/更新媒体防盗链策略 |
| GET | `/api/v1/hotlink-domains` | 查询当前用户防盗链域名 |
| POST | `/api/v1/hotlink-domains` | 添加防盗链域名 |
| DELETE | `/api/v1/hotlink-domains/{id}` | 删除防盗链域名 |
| POST | `/api/v1/media/{id}/signed-url` | 生成媒体签名 URL |

开发者 API 必须返回 `X-Quota-Used`、`X-Quota-Limit` 和 `X-Quota-Remaining`，超额时使用 HTTP 429 或稳定业务错误码，不返回数据库或对象存储原始错误。

### 3.3 管理端 API

标准套餐、相册、文件夹、广告位等 CRUD 通过 Resource Engine；以下复杂操作使用显式业务端点：

| 方法 | 路径 | 用途 |
| --- | --- | --- |
| GET | `/api/v1/admin/overview` | 运营概览 |
| GET | `/api/v1/admin/media/{id}/owner` | 媒体归属和用户摘要 |
| POST | `/api/v1/admin/media/actions/bulk-hide` | 批量隐藏媒体 |
| POST | `/api/v1/admin/media/actions/retry-processing` | 重试处理任务 |
| POST | `/api/v1/admin/moderation/{id}/approve` | 审核通过 |
| POST | `/api/v1/admin/moderation/{id}/reject` | 审核拒绝 |
| POST | `/api/v1/admin/users/{id}/grant-quota` | 赠送额度 |
| GET | `/api/v1/admin/analytics/summary` | 管理统计 |

发现页和举报治理：

| 方法 | 路径 | 用途 |
| --- | --- | --- |
| GET | `/api/v1/discovery/status` | 查询发现页状态 |
| GET | `/api/v1/discovery/feed` | 查询公开、ready 且未被拒绝的媒体 |
| POST | `/api/v1/media/{id}/reports` | 举报媒体 |
| POST | `/api/v1/me/moderation-appeals` | 提交申诉 |
| GET | `/api/v1/me/moderation-actions` | 查看自己的处罚记录 |
| GET/PUT | `/api/v1/admin/discovery/settings` | 查看/修改发现页开关和审核模式 |
| GET | `/api/v1/admin/moderation/reports` | 举报队列 |
| POST | `/api/v1/admin/moderation/reports/{id}/resolve` | 处理举报 |
| POST | `/api/v1/admin/moderation/media/{id}/hide` | 隐藏违规图片 |
| POST | `/api/v1/admin/moderation/media/{id}/restore` | 恢复图片 |
| POST | `/api/v1/admin/moderation/users/{id}/restrict` | 限制上传 |
| POST | `/api/v1/admin/moderation/users/{id}/suspend` | 暂停账户 |
| POST | `/api/v1/admin/moderation/users/{id}/ban` | 封禁账户 |
| POST | `/api/v1/admin/moderation/appeals/{id}/resolve` | 处理申诉 |

## 4. 资源与页面清单

| 资源/页面 | 模式 | P0 | 说明 |
| --- | --- | --- | --- |
| Plans | Generic Resource | 是 | 套餐基本信息 |
| Plan Entitlements | Generic Resource/嵌入表单 | 是 | 权益键值和版本 |
| Folders | Generic Resource | 是 | 用户数据范围 |
| Albums | Generic Resource + 自定义封面 | 是 | 用户数据范围 |
| Media Library | Custom Page | 是 | 网格、筛选、批量操作 |
| Upload Workbench | Custom Page | 是 | 进度、重试、分片 |
| Share Links | Generic Resource | 是 | Token 不显示 |
| Personal API Tokens | Custom Form + Generic List | 是 | Token 只显示一次 |
| Moderation Queue | Custom Page | 是 | 审核和举报工作台 |
| Usage Dashboard | Custom Page | 是 | 用量和升级入口 |
| Orders | Generic Resource + 状态页 | P1 | 首期 fake/manual |
| Ad Slots/Creatives | Generic Resource | P1 | 素材审核和周期 |
| Analytics | Custom Page | P1 | 聚合图表 |
| Discovery | Custom Page | P1 | 公开审核媒体 |

## 5. 风险与控制

| 风险 | 影响 | 控制措施 | 发布门槛 |
| --- | --- | --- | --- |
| 免费用户滥用存储 | 成本上升 | 文件/流量/速率限制、哈希去重、举报 | P0 必须有配额和限流 |
| 盗链导致流量失控 | 成本上升 | 签名 URL、防盗链、流量告警 | P1 前必须有流量统计 |
| 违法或侵权内容 | 合规风险 | 审核、举报、封禁、审计 | 未审核不公开 |
| 对象与数据库不一致 | 数据丢失/计量错误 | Outbox/清理任务、孤儿扫描、重试 | P0 必须可恢复 |
| 支付重复回调 | 财务错误 | 事件 ID 幂等和订单状态机 | 接真实支付前必须通过测试 |
| EXIF 泄露位置 | 隐私风险 | 默认清理、原图策略和告知 | P0 必须测试 GPS 清理 |
| 队列积压 | 上传不可用 | 队列监控、重试上限、人工重试 | P0 必须可观察 |
| 框架资源误用 | 研发返工 | Generic/Custom 明确分界 | Task 0 完成模块映射 |

## 6. 需求变更规则

- P0 需求变更必须说明对免费闭环、数据模型和迁移的影响。
- 新增套餐权益必须先定义权益键、数据类型、默认值、计费周期和失败行为。
- 新增外部 Provider 必须先定义接口、fake 实现、失败语义和集成验证方式。
- 新增公开内容能力必须同时补充审核、举报、隐私和限流要求。
- 任何会改变媒体、订单或用量状态的变更必须补充状态迁移和幂等测试。
- Token 新增 Scope 必须定义是否属于基础能力、用户可否创建、对应错误码和审计字段；不能改变个人 Token 对自己媒体的基本管理能力。
- 新增图片链接格式必须说明 Variant 来源、是否计入处理配额、是否允许 Free 使用和兼容旧客户端的字段名。
