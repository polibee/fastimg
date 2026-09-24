# FastImg 支付网关与账务设计

## 1. 目标与边界

FastImg 的业务模块不直接调用支付宝、微信、Stripe、PayPal 或其他支付渠道。所有支付渠道都通过统一 Payment Gateway 接口接入，订单、支付流水、订阅和权益由平台自己的账务域负责。

支付网关只负责：创建支付意图、返回收银台信息、查询渠道状态、验签通知、申请退款和查询退款结果。它不负责决定用户买了什么、何时生效、如何扣配额或如何展示广告。

首期必须支持 `fake/manual` 网关用于离线验收；支付网关预留法币 Provider 和 Crypto Provider。接入真实渠道前，必须完成真实商户配置、签名验证、回调读写、重复通知和失败场景测试。

## 当前实现状态

截至 2026-09-24，M6 Task 1-6 已完成，M7 Xcash Task 7、NOWPayments Task 8 和 PayPal Task 9 已完成离线适配：已增加版本化 `plan_prices` 价格目录模型、订单/支付/退款/财务/履约表模型和本地迁移定义；价格快照会复制套餐条款与权益，避免后台修改当前套餐后改变历史订单。Seeder 保留 Free 免费套餐，并预置 Creator/Pro 付费层级及月付/年付 CNY 价格（仅在 `plan_prices` 表存在时写入）。会员端已增加 own-scope 订单创建、订单列表、订单详情、支付启动和取消 API，以及 `/plans`、`/checkout/:orderId`、`/orders` 的独立页面和导航；开发环境 Fake Provider 支持幂等创建、测试态状态转换和会员“确认测试支付”成功路径，该入口在生产环境返回不可用；Xcash Provider 已完成 HMAC、账单创建、状态查询、Webhook 验签和风险状态映射；NOWPayments Provider 已完成 hosted invoice、状态查询、IPN HMAC-SHA512 验签和部分支付复核映射；PayPal Provider 已完成 OAuth、Orders v2 CAPTURE、审批链接、捕获查询/退款和 Webhook Verification API 验签，三者默认关闭。Webhook 入口已具备事件去重、金额/币种校验和支付成功后创建履约任务的边界，履约服务会用订单快照更新订阅；退款上限、追加式财务流水、对账差异分类以及管理员订单/流水/Webhook/退款查询接口已建立。`fastimg_dev` 已执行 billing migration 并完成 seed，后端已在本地独立端口运行；后台退款/履约操作页面和三家渠道的真实沙盒证据仍待后续任务。

当前迁移使用文本字段保存受服务层校验的快照 JSON，以兼容项目现有 Goravel Schema 抽象；暂不宣称已完成 PostgreSQL JSONB 索引优化。前端价格只展示服务端返回的活动价格，不提交金额；Free 套餐不依赖支付网关，游客可以访问套餐目录，只有认证会员可以创建订单。

## 2. 核心领域关系

```text
User
  -> Order（购买什么、金额、快照）
      -> OrderItem（套餐/附加商品）
      -> PaymentIntent（本次支付意图）
          -> PaymentTransaction（渠道流水和状态变化）
              -> PaymentEvent（原始回调去重和审计）
      -> Subscription（权益生效周期）
          -> EntitlementSnapshot（购买时权益快照）
              -> Quota / Advertising / API Rate Limit
```

关键原则：订单是业务事实，渠道订单号只是外部引用；PaymentIntent 允许一次订单多次支付尝试；PaymentTransaction 只追加；Subscription 只有在账务域确认成功后才能激活；配额、广告免除、API 速率和高级能力都从权益快照读取。

## 3. 支付网关接口

领域接口放在 `backend/app/services/billing`，渠道适配器放在独立 Provider 包中：

```go
type PaymentGateway interface {
    CreatePayment(ctx context.Context, req CreatePaymentRequest) (PaymentSession, error)
    QueryPayment(ctx context.Context, req QueryPaymentRequest) (GatewayPayment, error)
    VerifyNotification(ctx context.Context, req NotificationRequest) (GatewayEvent, error)
    CreateRefund(ctx context.Context, req RefundRequest) (GatewayRefund, error)
    QueryRefund(ctx context.Context, req QueryRefundRequest) (GatewayRefund, error)
}
```

`CreatePaymentRequest` 必须包含平台订单号、PaymentIntent ID、金额、货币、商品摘要、回调地址、返回地址和幂等键。渠道 Provider 不得重新计算套餐价格。

Provider 统一返回：

```text
gateway_code
gateway_payment_id
gateway_transaction_id nullable
status: pending | succeeded | failed | canceled | requires_action
checkout_url nullable
client_secret nullable
occurred_at
```

敏感原始回调只能保存在受控存储或脱敏摘要中，不能进入普通日志和前端响应。

### 3.1 加密货币 Provider

加密货币支付作为网关渠道接入，不直接写入套餐、配额或媒体业务模块。Provider 负责创建收款请求、生成或获取收款地址、监听链上交易、计算确认状态、处理链回滚，并把结果转换为统一的 `PaymentTransaction`。

```text
CryptoPaymentRequest:
  asset: BTC | ETH | USDT | ...
  network: bitcoin | ethereum | tron | ...
  fiat_amount
  quote_currency
  quote_amount
  exchange_rate
  quote_expires_at
  deposit_address
  payment_uri nullable
  required_confirmations
  expires_at
```

设计约束：

- 资产、网络、合约地址和精度必须由服务端配置，不能信任前端传入。
- 同名资产在不同网络必须视为不同支付方式，例如同为 USDT 也必须区分网络。
- 订单先锁定法币价格、报价金额、汇率版本、资产、网络、精度和过期时间；支付金额不能按回调时汇率重新解释。
- 首期优先使用受监管或可审计的托管支付 Provider；自建私钥、热钱包、自动归集不属于 MVP。
- 平台不保存用户私钥，不要求用户提交助记词、私钥或交易所密码。
- 收款地址、Memo/Tag、合约地址和网络必须在用户确认页与支付回调中双重校验。

### 3.2 加密支付状态

```text
quote_created -> awaiting_payment -> detected
detected -> confirming -> confirmed -> settled
detected/confirming -> underpaid | overpaid | expired
confirming -> reorged -> confirming | failed
```

只有 `confirmed` 才能进入平台 `PaymentIntent.succeeded`。`detected` 只代表发现链上交易，不代表收款完成；确认数必须按资产和网络配置。链重组、双花风险、节点异常和 Provider 状态不一致必须进入人工复核或重试，不能自动发放权益。

### 3.3 金额差异

- 少付：保持订单未完成，展示还需支付金额；不得按少付金额直接激活套餐。
- 多付：默认不自动增加套餐权益，记录差额并按退款或人工补偿策略处理。
- 过期报价收到付款：进入 `pending_review`，不能按当前价格自动套用新订单。
- 资产或网络错误：即使链上有交易，也不能直接认定为目标订单已支付。
- 同一链上交易不得同时匹配多个订单；匹配键至少包含网络、交易哈希、输出/地址、资产和金额规则。

### 3.4 加密支付合规

加密支付必须受站点地区、资产、网络和风控配置控制。管理员可以停用某个资产、网络或地区。高风险地址、制裁名单命中、异常来源、混币服务风险或 Provider 风控失败时，订单进入 `compliance_review`，不自动发放权益。是否需要 KYC、交易限额、退款方式和数据保留期限由部署地区及合规要求决定，不能在业务代码中硬编码为全球通用规则。

## 4. 状态机

### Order

```text
created -> pending_payment -> paid -> fulfilled
created/pending_payment -> canceled
pending_payment -> expired
paid/fulfilled -> partially_refunded -> refunded
paid/fulfilled -> disputed
```

`paid` 表示平台账务确认已收款；`fulfilled` 表示订阅或权益已发放。两者分开，避免支付成功但发放失败时丢失补偿入口。

### PaymentIntent

```text
created -> pending -> requires_action -> succeeded
created/pending/requires_action -> failed
created/pending -> canceled | expired
```

一个订单可存在多个 PaymentIntent，但同一时刻最多一个未过期的主动支付意图。

### PaymentTransaction

```text
authorized -> captured -> settled
captured -> refund_pending -> refunded | refund_failed
captured -> disputed -> resolved
```

交易流水只追加，不允许编辑原记录修正金额或状态；金额调整使用 refund、chargeback 或 adjustment 流水。

## 5. 数据模型

### Order / OrderItem

```text
Order: id, public_order_no, user_id, currency, subtotal_amount,
discount_amount, total_amount, status, idempotency_key,
billing_snapshot, customer_snapshot, expires_at, paid_at, fulfilled_at

OrderItem: order_id, product_type: plan | addon, product_id,
product_version, quantity, unit_amount, discount_amount, total_amount,
entitlement_snapshot
```

价格、货币、套餐版本、周期、权益摘要和税费规则必须保存快照，不能因为后台修改当前套餐而改变历史订单含义。

### PaymentIntent / PaymentTransaction / PaymentEvent

```text
PaymentIntent: id, order_id, user_id, gateway_code, amount, currency,
status, gateway_payment_id, checkout_url, attempt_no, idempotency_key,
expires_at, succeeded_at, failed_at

PaymentTransaction: order_id, payment_intent_id, gateway_code,
type: payment | refund | chargeback | adjustment,
direction: credit | debit, amount, currency, status,
gateway_transaction_id, gateway_event_id, occurred_at

PaymentEvent: gateway_code, event_id, event_type, signature_valid,
payload_hash, payload_reference, processing_status, retry_count,
processed_at, created_at

CryptoPayment: payment_intent_id, asset, network, contract_address,
deposit_address, memo_tag, quote_amount, quote_currency, exchange_rate,
quote_expires_at, tx_hash, block_number, confirmations,
required_confirmations, chain_status, risk_status
```

唯一约束至少包括 `gateway_code + event_id`、`public_order_no`、PaymentIntent 幂等键和渠道外部流水号。

## 6. 支付成功后的编排

1. 验证签名、时间窗口、事件 ID、金额和货币。
2. 写入 PaymentEvent，重复事件直接返回已处理结果。
3. 事务内锁定订单和 PaymentIntent，追加 PaymentTransaction。
4. 将 PaymentIntent 标记 succeeded、Order 标记 paid，提交后投递 `order.paid`。
5. 订阅服务创建或切换 Subscription，保存权益快照。
6. 配额服务刷新存储、流量、API 速率和高级能力；广告服务刷新免广告状态。
7. 发送通知、Webhook 和统计事件；失败时重试，不回滚已确认的收款事实。
8. 若权益发放失败，订单保持 paid、履约状态为 pending，并进入补偿队列。

回调不能直接执行长时间图片处理、发邮件或调用多个外部 Provider。

## 7. 退款、拒付与取消

- 未支付订单取消只改变订单状态，不产生支付流水。
- 退款必须通过平台 Refund 用例；部分退款累计金额不得超过已收金额。
- 退款成功追加 debit 流水；异步退款在 `refund_pending` 期间不能显示成功。
- 拒付使用独立 `chargeback` 流程，不能伪装成普通退款。
- 退款或拒付不自动删除用户媒体；媒体保留和账户处罚分别依据合规及风控策略处理。

## 8. API

```text
GET  /api/v1/plans
POST /api/v1/orders
GET  /api/v1/orders
GET  /api/v1/orders/{id}
POST /api/v1/orders/{id}/payments
POST /api/v1/orders/{id}/cancel
POST /api/v1/orders/{id}/refund-requests
GET  /api/v1/subscription

GET  /api/v1/admin/orders
GET  /api/v1/admin/payment-transactions
POST /api/v1/admin/orders/{id}/retry-fulfillment
POST /api/v1/admin/orders/{id}/manual-adjustment
POST /api/v1/payment-gateways/{gateway}/notifications
```

回调只返回渠道所需的确认结果，不返回内部订单、用户、订阅和错误详情。

## 9. 渠道路由、优惠与对账

### 9.1 支付渠道路由

渠道选择由服务端根据站点配置、货币、地区、客户端能力和渠道健康状态决定，前端只能提交允许的 `gateway_code`，不能指定任意 Provider。

```text
请求订单
  -> 读取可用渠道
  -> 过滤货币/地区/套餐限制
  -> 排除维护中或连续失败渠道
  -> 生成 PaymentIntent
  -> 调用统一 PaymentGateway
```

渠道配置至少包含：启停状态、支付类型 `fiat | crypto`、支持货币/资产/网络、支持地区、最低/最高金额、优先级、失败阈值、熔断时间、确认策略和密钥版本。渠道失败只影响新支付，不影响已有订阅。

### 9.2 优惠和价格

- 套餐价格、周期、货币和权益必须由服务端价格目录决定。
- 优惠码必须保存适用套餐、适用周期、有效期、总次数、单用户次数和是否可叠加。
- 创建订单时锁定价格版本和优惠计算结果；支付时不得重新按当前价格计算。
- 折扣金额不能超过商品金额；订单取消、支付失败和退款必须明确折扣是否恢复使用次数。
- 首期只实现固定金额或百分比优惠，不实现复杂营销规则和多级分销。

### 9.3 发票与账单

首期至少提供订单收据/账单详情，包含订单号、用户、商品、原价、折扣、应付金额、货币、支付渠道、支付时间和退款状态。真实税务发票、税率和开票 Provider 独立于支付网关，通过 Invoice Provider 接口扩展。

```text
Invoice: order_id, invoice_no, status, currency,
subtotal_amount, discount_amount, tax_amount, total_amount,
customer_snapshot, issued_at, voided_at
```

发票和订单使用快照，不能因套餐、用户资料或税率后来变化而改写历史账单。

### 9.4 对账与差异

每日或按渠道结算周期导入渠道结算文件/API 摘要，与 `PaymentTransaction` 按渠道、外部流水号、金额、货币和日期匹配：

```text
matched | missing_internal | missing_gateway |
amount_mismatch | currency_mismatch | duplicate | pending_review
```

对账差异只能进入人工复核或补偿流程，不能自动修改订单金额。对账记录、操作人、证据和处理结果必须审计。

## 10. 功能关联

| 功能 | 关联事实 | 触发时机 |
|---|---|---|
| 套餐 | OrderItem + EntitlementSnapshot | 订单履约成功 |
| 存储/流量配额 | Subscription + UsageLedger | 激活、续费、降级、退款 |
| API 速率和 Token 数量 | EntitlementSnapshot | 权益刷新 |
| 广告免除 | Subscription entitlement | 读取广告前判断 |
| 域名/防盗链/Webhook | 套餐权益和资源状态 | 升级可用，降级按策略限制新增 |
| 加密支付 | CryptoPayment + 风控/确认状态 | 仅 confirmed 且通过风控后履约 |
| 统计收入 | PaymentTransaction | 流水确认后异步聚合 |
| 通知和审计 | Order/Subscription 事件 | 支付、履约、退款、人工调整 |

禁止用当前套餐名称解释历史订单，也禁止用订单 `paid` 直接绕过 Subscription/EntitlementSnapshot。

## 11. 前后台状态展示

用户端必须展示：订单号、商品快照、应付金额、支付渠道、支付中/成功/失败/已退款状态、权益生效时间、失败原因的安全摘要和重试入口。支付中不能重复创建无限订单；支付成功但权益尚未生效时显示“正在发放权益”。

管理端必须提供订单、支付意图、支付流水、回调事件、履约任务、退款和对账差异的关联查询。人工补偿、退款和调整操作必须二次确认并填写原因。

## 12. 安全与验收

- 金额使用最小货币单位整数；前端价格、套餐权益、用户 ID 和支付成功状态全部不可信。
- 订单创建、支付创建、回调、履约、退款和人工调整都必须幂等。
- 渠道密钥按 Provider 和环境隔离并支持轮换。
- 加密支付不接触用户私钥；Provider 凭据、Webhook 密钥、合约配置和归集权限必须分离并可轮换。
- 支付网关不可用时停止新支付，不修改已有订阅和额度；Free 服务不依赖支付网关。
- fake/manual Provider 必须完成创建订单、发起支付、回调、履约和查询。
- 重复回调不重复激活订阅、增加额度或发送重复业务事件。
- 支付成功但履约失败时，订单可查询并可人工重试履约。
- 部分退款、全额退款、退款失败和拒付均有独立流水。
- 加密支付覆盖报价过期、未支付、少付、多付、错误网络、确认数不足、链重组、重复交易和风控复核。
- 真实 Provider 未完成签名、回调和失败测试前，只能标记为 fake/offline。
