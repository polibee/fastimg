# FastImg PayPal Provider

## 当前状态

PayPal 已接入统一 `PaymentGateway` 的离线 Provider，默认关闭。当前覆盖 OAuth client-credentials、Orders v2 `CAPTURE` 创建、审批链接、订单/捕获查询、capture refund 请求和 PayPal Webhook Verification API 验签；没有真实 PayPal Sandbox 凭据，不能视为 Sandbox 或生产验收完成。

## 配置

```dotenv
PAYPAL_ENABLED=false
PAYPAL_ENVIRONMENT=sandbox
PAYPAL_API_BASE_URL=https://api-m.sandbox.paypal.com
PAYPAL_CLIENT_ID=
PAYPAL_CLIENT_SECRET=
PAYPAL_WEBHOOK_ID=
PAYPAL_TIMEOUT=10s
PAYPAL_RETURN_URL=https://img.example.com/plans
PAYPAL_CANCEL_URL=https://img.example.com/plans
```

生产环境使用 PayPal Live base URL 和独立 Live Client ID/Secret/Webhook ID；sandbox 与 live 凭据不能混用。OAuth token 只在进程内缓存并提前过期，不写入数据库、日志或前端。

## 交易流程

1. 订单价格由 FastImg 服务端快照决定，Provider 创建 PayPal Orders v2 请求，使用 `intent=CAPTURE`、订单号作为 `invoice_id/reference_id` 和 `PayPal-Request-Id` 幂等键。
2. 返回 `approve` 链接给会员完成 PayPal 授权；同步 return 只用于回到前端，不直接把订单标记为已支付。
3. 回到服务端后查询订单/捕获状态；`COMPLETED` 需要金额、货币、订单和捕获 ID 通过统一 Webhook/账务校验才履约。
4. PayPal Webhook 通过 `/v1/notifications/verify-webhook-signature` 验证，而不是相信客户端或未验签事件；通用 WebhookService 再做事件幂等和订单金额校验。

## 状态与退款

`CREATED`/`APPROVED` 为需要用户动作，`COMPLETED` 为成功候选，`VOIDED`/`CANCELED` 为失败。捕获状态 `PENDING`、拒绝、逆转和风控复核不能自动发放套餐。

退款使用 PayPal capture refund endpoint，并复用平台 Refund 的幂等键；退款成功追加 debit 财务流水，不修改原支付事实。Provider 查询不到 capture ID 或渠道返回不确定状态时，保持 pending，不伪造成功。

## 离线验证与启用门禁

```powershell
$env:GOCACHE='D:\laragon\www\fastimg\go-vue-admin\.gocache-payment'
go test ./app/services/billing/providers/paypal ./app/services/billing ./app/modules/billing/controllers -count=1
```

启用前必须补齐真实 Sandbox 的 OAuth、订单创建、审批、捕获、退款、Webhook Verification、重复通知、超时后查询和金额/货币不匹配证据；`PAYPAL_ENABLED` 在门禁完成前保持 `false`。
