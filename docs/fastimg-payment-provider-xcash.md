# FastImg Xcash Provider

## 目标与启用边界

FastImg 通过统一 `PaymentGateway` 接口接入 Xcash 账单收款，不把 Xcash 字段直接泄漏到套餐、订单或会员页面。Free 套餐和 Fake Provider 不依赖 Xcash；Provider 只有在 `XCASH_ENABLED=true` 且 AppID、HMAC 密钥完整时才注册，官方 API 地址默认使用官方地址。

当前实现是可离线测试的 Provider 适配，已覆盖签名、账单创建、公开状态查询、Webhook 验签和状态映射；2026-09-25 已使用本地管理端配置的 Xcash 凭证完成真实账单创建验证并拿到 `pay_url`，但尚未完成真实付款、Webhook 到账和权益履约验收，因此不宣称已经完成 Xcash 生产收款验收。

## 配置

```dotenv
XCASH_ENABLED=false
XCASH_API_BASE_URL=https://pay.xca.sh
XCASH_APP_ID=
XCASH_HMAC_KEY=
XCASH_TIMEOUT=10s
XCASH_CALLBACK_URL=https://img.example.com/api/v1/payment-gateways/xcash/webhook
XCASH_RETURN_URL=https://img.example.com/orders/{order_id}?payment=success
```

管理端会预填 `https://pay.xca.sh`，也允许改为兼容网关地址；开发环境的本地 HTTP stub 只用于离线 Provider 测试。真实 Xcash 通知地址必须使用公网 HTTPS，并按 Xcash 项目配置 IP 白名单和通知地址。

管理员在“系统设置 > 支付网关 > XCash”填写 Appid 和 HMAC key；官方 API 根地址默认是 `https://pay.xca.sh`，Webhook 地址需要在 Xcash 项目后台配置并与本地设置保持一致，成功回跳地址由系统根据 `site_url` 预填，均可修改，敏感值会加密保存。FastImg 按 Xcash 官方接口使用 `redirect_url`，不会向 `/v1/invoice` 发送不受官方接口支持的 `notify_url` 或 `return_url`；`{order_id}` 会在创建账单时替换为 FastImg 订单号，回跳地址只负责回到订单页，不能代替 Webhook 验签。

## 签名协议

请求和 Webhook 使用以下请求头：

```text
XC-Appid: <appid>
XC-Timestamp: <unix seconds>
XC-Nonce: <unique nonce>
XC-Signature: <lowercase hex>
```

签名原文严格为：

```text
XC-Nonce + XC-Timestamp + raw request body
```

算法为 HMAC-SHA256，签名必须与实际发送的原始 JSON 字节一致。FastImg 接收 Webhook 时同时校验 AppID、签名和 5 分钟时间窗；重复事件由通用 `PaymentWebhookEvent` 以 Provider + Event ID 去重。

## 账单流程

```text
FastImg order
  -> POST /v1/invoice
  -> Xcash sys_no + pay_url
  -> member opens pay_url
  -> Xcash invoice webhook
  -> verify signature and query invoice if needed
  -> PaymentTransaction credit
  -> order paid
  -> fulfillment task
```

创建请求只发送服务端订单快照计算出的金额，当前使用最小货币单位转换为两位小数；前端不能提交金额。Xcash 的 `sys_no` 保存为 `ProviderPaymentID`，`pay_url` 只返回给已认证的订单页面，不写入日志。

Xcash 账单 Webhook 示例通常只包含 `sys_no`、链、币、交易哈希和 `confirmed`，不一定包含原始法币金额。因此适配器在缺少 `amount` 或 `currency` 时调用公开账单查询接口，用查询到的金额进入 FastImg 的订单金额/币种一致性校验，不能用链上 `pay_amount` 直接当成 CNY/USD 金额。

## 状态和风控

| Xcash 状态 | FastImg Provider 状态 | 是否可履约 |
|---|---|---|
| `waiting` | `pending` | 否 |
| `completed` / `confirmed` | `succeeded` | 需通过统一金额、币种和订单校验 |
| `expired` / `wrong_network` | `failed` | 否 |
| `underpaid` / `overpaid` | `pending_review` | 否 |
| `risk_level=high/critical` | `pending_review` | 否 |

只有统一 `WebhookService` 将事件确认并追加支付流水后，订单才会进入 `paid`，履约服务再更新订阅权益。同步跳转到 `return_url` 不能改变订单状态。

## 退款

当前 Xcash Provider 不自动调用退款接口，退款必须进入管理员人工处理/外部钱包流程，并保留退款原因、证据和追加式财务流水。不能把链上转账或买家同步返回误认为退款成功。

## 测试与上线门禁

```powershell
pwsh -File scripts/fastimg-go.ps1 test ./app/services/billing/providers/xcash ./app/services/billing ./app/modules/billing/controllers -count=1
```

测试使用本地 HTTP stub，不访问真实 Xcash。启用前必须补齐真实沙盒/小额生产前验证、Webhook 重复投递、过期报价、少付、多付、风险复核、错误网络和失败重试证据，并保持 `XCASH_ENABLED=false` 直到这些门禁通过。
