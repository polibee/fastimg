# FastImg NOWPayments Provider

## 当前状态

NOWPayments 已按统一 `PaymentGateway` 接口完成 hosted invoice 创建、支付状态查询、IPN HMAC-SHA512 验签和状态映射。API Key 与 IPN Secret 按官方职责分离：只配置 API Key 时可以创建/查询支付并出现在会员结算页；IPN Secret 是自动验签和自动履约的前置条件。未完成真实沙盒/小额支付、回调和对账验收前，不宣称生产支付已经验证。

## 配置

```dotenv
NOWPAYMENTS_ENABLED=false
NOWPAYMENTS_API_BASE_URL=https://api.nowpayments.io
NOWPAYMENTS_API_KEY=
NOWPAYMENTS_IPN_SECRET=
NOWPAYMENTS_TIMEOUT=10s
NOWPAYMENTS_CALLBACK_URL=https://img.example.com/api/v1/payment-gateways/nowpayments/webhook
NOWPAYMENTS_SUCCESS_URL=https://img.example.com/orders/{order_id}?payment=success
NOWPAYMENTS_CANCEL_URL=https://img.example.com/orders/{order_id}?payment=cancelled
```

API Key 只由服务端通过 `x-api-key` 请求头发送；IPN Secret 只用于校验 `x-nowpayments-sig`，两者不能写入前端、URL、普通日志或订单响应。

管理员在“系统设置 > 支付网关 > NOWPayments”填写 API Key 和可选的 IPN Secret；官方 API 根地址默认是 `https://api.nowpayments.io`，IPN Callback、成功/取消回跳由系统根据 `site_url` 生成，敏感值会加密保存。`{order_id}` 在创建发票时替换为真实订单号，成功/取消回跳只用于展示订单状态，权益以验签 IPN 为准。生产启用自动履约前必须配置公网 HTTPS 回调地址和 IPN Secret。

保存设置后服务端会热刷新渠道注册表，不需要重启后端。渠道是否显示由“启用 + API Key + 官方 API 地址”决定；IPN Secret 缺失时只降级为“可发起支付、等待人工/查询确认”，不能把渠道错误地标记为不可用。

## 交易流程

1. FastImg 使用服务端订单快照调用 `POST /v1/invoice`，提交价格金额、法币、订单号、描述、IPN 回调和成功/取消地址。
2. 保存返回的 `payment_id` 和 `invoice_url`，会员只看到收银台链接和脱敏状态。
3. 必要时通过 `GET /v1/payment/{payment_id}` 查询最新状态；查询结果仍要经过 FastImg 金额、币种、订单和风控校验。
4. NOWPayments 向回调地址发送 IPN；先按 Provider 规则递归排序 JSON Key 后计算 HMAC-SHA512，再交给统一 WebhookService 去重和履约。

## 状态规则

| NOWPayments | FastImg | 自动履约 |
|---|---|---|
| `waiting` / `confirming` / `confirmed` / `sending` | `pending` | 否 |
| `finished` | `succeeded` 候选 | 通过统一金额/币种/签名校验后才可 |
| `partially_paid` | `pending_review` | 否 |
| `failed` / `expired` / `refunded` | `failed` | 否 |

不能使用“不是 failed 就算成功”的规则。少付、错链、重复支付和回调重放需要通过事件去重、对账和人工复核处理。退款仍保持人工边界，后续另行接入 Provider 支持的退款接口。

## 离线验证

```powershell
pwsh -File scripts/fastimg-go.ps1 test ./app/services/billing/providers/nowpayments ./app/services/billing ./app/modules/billing/controllers -count=1
```

上线前必须使用官方当前 API 文档和沙盒/小额环境验证支付币种、网络、最小金额、过期、少付、IPN 重复投递、错误签名、状态查询和对账。`NOWPAYMENTS_ENABLED` 在门禁完成前保持 `false`。API Key 仅用于服务端 `x-api-key` 请求头；IPN Secret 仅用于 `x-nowpayments-sig` 验签，二者都不能进入前端、公开价格响应或普通审计日志。
