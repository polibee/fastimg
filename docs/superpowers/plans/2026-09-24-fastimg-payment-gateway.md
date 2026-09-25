# FastImg Payment Gateway Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在保留 Free 免费闭环的前提下，为 FastImg 建立可离线验收、可审计、可幂等履约的统一支付网关，实现套餐定价、订单流水、Fake Provider，并按门禁接入 Xcash、NOWPayments 和 PayPal。

**Architecture:** 采用 `价格目录 -> Order/OrderItem -> PaymentIntent/PaymentAttempt -> Provider Adapter -> PaymentTransaction/WebhookEvent -> FulfillmentTask -> Subscription/EntitlementSnapshot` 的单向业务链。Provider 只负责外部渠道协议，Billing Service 负责订单和财务事实，Plans/Quota Service 负责权益履约；管理员和会员端分别使用 `/admin/**` 与根路径页面，均通过现有认证、RBAC、Resource Engine、审计、i18n 和 shadcn-vue 能力。

**Tech Stack:** Goravel v1.18、Go、GORM ORM、PostgreSQL、Redis/队列、Vue 3、Vue Router、Pinia、shadcn-vue、vue-i18n、现有 Resource Engine、OpenAPI/TypeScript API Client。

**Spec:** `D:\laragon\www\fastimg\docs\fastimg-payment-gateway-design.md`, `D:\laragon\www\fastimg\docs\fastimg-payment-data-model.md`, `D:\laragon\www\fastimg\docs\fastimg-payment-provider-xcash.md`, `D:\laragon\www\fastimg\docs\fastimg-payment-provider-nowpayments.md`, `D:\laragon\www\fastimg\docs\fastimg-payment-provider-paypal.md`, `D:\laragon\www\fastimg\docs\fastimg-payment-reconciliation.md`, `docs/fastimg-payment-gateway.md`

## Global Constraints

- Free 用户无需支付即可上传、管理自己的媒体、获取自己的链接和删除自己的媒体；支付不可成为免费核心闭环的硬依赖。
- 所有金额使用最小货币单位整数 `amount_minor`；加密资产金额使用带精度的十进制字符串，禁止 `float`。
- 套餐价格、折扣、税费、计费周期和权益必须在订单/订单明细中保存快照；价格发布后不能覆盖历史订单。
- Provider 不得计算套餐价格、激活订阅、修改配额或决定广告权益；这些动作只能由 Billing/Plans/Quota Service 完成。
- 浏览器回跳只展示状态，不授予权益；只有已验签、已去重、金额和货币匹配的渠道事实才能进入账务履约。
- PaymentTransaction、FinancialTransaction、WebhookEvent 和审计事实只追加；退款、拒付和人工调整使用反向流水，不编辑原流水。
- 每个订单创建、支付尝试、捕获、退款、Webhook、履约任务和补偿动作必须幂等；重复事件不得重复发放订阅或额度。
- Provider 默认 `enabled=false`；Fake Provider 先用于本地和自动化验收，真实 Provider 未完成沙盒签名/回调/失败验收时不得宣称可生产收款。
- Xcash、NOWPayments、PayPal 密钥只来自服务端环境/受控密钥配置，不能进入前端、普通日志、审计详情或数据库明文。
- 后台页面全部位于 `/admin/**`；会员端只能看到自己的订单、支付状态、订阅和账单；管理员跨用户能力必须有显式 RBAC 权限、数据范围和审计。
- 所有用户可见文本同步维护 `admin/src/locales/zh-CN/<feature>.json` 与 `admin/src/locales/en-US/<feature>.json`；API 错误码使用稳定 snake_case/code 契约。
- 复用现有 Goravel Auth/RBAC/Resource/审计/迁移/队列、Vue Router/Pinia/i18n/shadcn-vue/API Client；不要新建重复的通用核心。
- 只在用户明确授权的本地开发库执行迁移；本计划不包含生产数据库、真实扣款、真实退款、部署或发布。
- 后端、Vite/preview 和其他项目使用不同端口；使用进程级启动参数和 `wsl-cn-net` 端口流程，不能修改共享 Laragon 配置或其他 checkout。

## Review Focus

- 同一个内部幂等键、同一个 Provider 事件或同一个 PayPal capture 重试只能产生一个订单支付事实和一次履约；由 Task 2、Task 3 的重复提交/重复回调测试覆盖。
- 渠道状态为 pending、confirming、failed、expired、underpaid、overpaid、wrong_network 或 compliance_review 时不能激活订阅；由 Task 3、Task 7、Task 8 的状态映射测试覆盖。
- 订单价格已锁定后管理员修改套餐或价格不能改变订单金额、权益和账单；由 Task 1 的快照回归测试覆盖。
- Webhook 签名错误、时间窗口错误、payload 过大或事件金额/币种不匹配必须拒绝且不写入成功流水；由 Task 3、Task 7-9 的验签测试覆盖。
- 管理员只能查看/处理有权限的全站订单与流水，会员只能访问自己的订单；由 Task 5、Task 6 的 RBAC/own-scope 集成测试覆盖。

---

### Task 1: Price Catalog and Billing Persistence

**Files:**
- Create: `backend/database/migrations/20260924000006_create_billing_tables.go`
- Create: `backend/app/models/billing.go`
- Create: `backend/app/services/billing/price_catalog.go`
- Create: `backend/app/services/billing/price_catalog_test.go`
- Modify: `backend/app/models/plan.go`
- Modify: `backend/app/services/plans/plan_service.go`
- Modify: `backend/app/services/plans/plan_validation.go`
- Modify: `backend/app/modules/plans/controllers/plan_controller.go`
- Modify: `docs/fastimg-payment-gateway.md`
- Test: `backend/database/billing_migration_test.go`

**Interfaces:**
- Consumes: existing `models.Plan`, `planservices.ActivePlans`, `quota.Entitlement` parsing and `facades.Orm()`.
- Produces: `models.PlanPrice`, `models.Order`, `models.OrderItem`, `models.PaymentIntent`, `models.PaymentAttempt`, `models.PaymentTransaction`, `models.PaymentWebhookEvent`, `models.Refund`, `models.FinancialTransaction`, `models.FulfillmentTask`; `billing.PriceCatalog.GetActivePrice(planID, currency, period)` and `billing.PriceCatalog.SnapshotPlanPrice(planID, priceID)`.

- [ ] **Step 1: Write failing model and catalog tests**

  Add tests proving active price lookup rejects disabled/expired prices, returns `amount_minor` and a plan/entitlement snapshot, and that Free has no payable price requirement. Add tests proving an order snapshot remains unchanged when the current plan row is edited.

- [ ] **Step 2: Run focused tests to verify the RED state**

  Run from `backend`:

  ```powershell
  go test ./app/services/billing ./database -run "Price|Snapshot|Billing" -count=1
  ```

  Expected: FAIL because the billing models, catalog and migration do not exist.

- [ ] **Step 3: Add the migration and model layer**

  Create PostgreSQL tables for `plan_prices`, `orders`, `order_items`, `payment_intents`, `payment_attempts`, `payment_transactions`, `payment_webhook_events`, `refunds`, `financial_transactions` and `fulfillment_tasks`. Use integer minor units, JSONB snapshots, unique constraints for `public_order_no`, `(user_id, idempotency_key)`, `(provider_code, event_id)` and provider transaction identifiers. Add indexes on user/status/provider/time fields. Down migration must drop only these new tables in dependency order.

  Add Go models with explicit JSON tags, timestamps and status fields. Keep provider payload references/hash fields separate from public response fields; do not store credentials or raw secrets.

- [ ] **Step 4: Implement price versioning and snapshot catalog**

  Implement `PriceCatalog.GetActivePrice` to select only active, effective prices and validate `monthly|yearly`, ISO currency and non-negative integer amount. Implement `SnapshotPlanPrice` to copy plan terms, price terms and parsed entitlement JSON into a value object used by order creation. Update public plan output to expose active prices as an optional `prices` array without removing existing fields.

- [ ] **Step 5: Run focused tests and migration checks**

  Run:

  ```powershell
  go test ./app/services/billing ./app/services/plans ./database -run "Price|Snapshot|Billing" -count=1
  go test ./app/models ./app/services/quota -count=1
  ```

  Expected: PASS; migration test must verify table/constraint names and no existing plan/subscription data is rewritten.

- [ ] **Step 6: Update the design ledger**

  Record in `docs/fastimg-payment-gateway.md` that Task 1 persistence is implemented only when the tests pass; document any intentionally unsupported tax/invoice/provider fields as not yet integrated.

- [ ] **Step 7: Commit the independent slice**

  ```powershell
  git add backend/database/migrations/20260924000006_create_billing_tables.go backend/app/models/billing.go backend/app/services/billing/price_catalog.go backend/app/services/billing/price_catalog_test.go backend/database/billing_migration_test.go backend/app/models/plan.go backend/app/services/plans/plan_service.go backend/app/services/plans/plan_validation.go backend/app/modules/plans/controllers/plan_controller.go docs/fastimg-payment-gateway.md
  git commit -m "feat: add versioned billing price catalog and persistence"
  ```

### Task 2: Order Creation, Own-Scope APIs and Fake Payment Provider

**Files:**
- Create: `backend/app/services/billing/order_service.go`
- Create: `backend/app/services/billing/order_service_test.go`
- Create: `backend/app/services/billing/payment_service.go`
- Create: `backend/app/services/billing/payment_service_test.go`
- Create: `backend/app/services/billing/providers/provider.go`
- Create: `backend/app/services/billing/providers/fake/fake.go`
- Create: `backend/app/services/billing/providers/fake/fake_test.go`
- Create: `backend/app/modules/billing/controllers/member_controller.go`
- Create: `backend/app/modules/billing/controllers/fake_controller.go`
- Create: `backend/app/modules/billing/routes.go`
- Modify: `backend/routes/web.go`
- Modify: `backend/app/http/middleware/member_scope.go`
- Test: `backend/app/modules/billing/controllers/member_controller_test.go`

**Interfaces:**
- Consumes: Task 1 `PriceCatalog`, billing models and existing authenticated user ID helper pattern.
- Produces: `OrderService.CreatePlanOrder(ctx, userID, CreateOrderRequest)`, `OrderService.ListOwnOrders`, `OrderService.GetOwnOrder`; `PaymentService.StartPayment`; `providers.PaymentGateway` with `CreatePayment`, `QueryPayment`, `VerifyWebhook`, `CreateRefund`, `QueryRefund`; member routes `POST /api/v1/orders`, `GET /api/v1/orders`, `GET /api/v1/orders/{id}`, `POST /api/v1/orders/{id}/payments`, `POST /api/v1/orders/{id}/cancel`.

- [ ] **Step 1: Write failing order/API tests**

  Cover: missing plan price, Free order rejection, invalid currency/period, duplicate `(user_id, idempotency_key)` returning the original order, immutable snapshot after plan edit, cross-user order lookup returning not found/forbidden without leaking existence, and cancellation of only unpaid non-expired orders.

- [ ] **Step 2: Run the order tests in RED**

  ```powershell
  go test ./app/services/billing ./app/modules/billing -run "Order|Own|Cancel" -count=1
  ```

  Expected: FAIL because the service/controller packages are new.

- [ ] **Step 3: Implement order service transactionally**

  In one database transaction, resolve the active `PlanPrice`, copy plan/price/entitlement/customer snapshots, generate a non-guessable `public_order_no`, create one `Order` and one `OrderItem`, and return the existing order for a duplicate idempotency key. Never accept `user_id`, price, amount or entitlements from the request body.

- [ ] **Step 4: Implement the Provider interface and Fake Provider**

  Define request/response value types with `gateway_code`, external payment ID, status, checkout URL, occurrence time and provider metadata. Fake Provider must be deterministic, default to `pending`, expose a test-only transition helper, and never contact the network. It must return the same payment for the same idempotency key and support a fake webhook event for later fulfillment tests.

- [ ] **Step 5: Implement member controllers and routes**

  Add own-scope JSON APIs. `POST /orders/{id}/payments` accepts only an enabled gateway code and creates a PaymentIntent/PaymentAttempt, then calls the registered Provider. Member responses expose order status, amount, currency, checkout URL and safe payment status; they never expose provider secrets or raw payloads. Free plan remains directly available through existing subscription provisioning.

- [ ] **Step 6: Run service and HTTP tests**

  ```powershell
  go test ./app/services/billing ./app/modules/billing ./app/http/middleware -count=1
  ```

  Expected: PASS, including duplicate order/payment attempts and cross-user denial.

- [ ] **Step 7: Commit the Fake checkout slice**

  ```powershell
  git add backend/app/services/billing backend/app/modules/billing backend/routes/web.go backend/app/http/middleware/member_scope.go
  git commit -m "feat: add own-scope orders and fake payment checkout"
  ```

### Task 3: Webhook Ingestion, Payment Transactions and Fulfillment

**Files:**
- Create: `backend/app/services/billing/webhook_service.go`
- Create: `backend/app/services/billing/webhook_service_test.go`
- Create: `backend/app/services/billing/fulfillment_service.go`
- Create: `backend/app/services/billing/fulfillment_service_test.go`
- Create: `backend/app/modules/billing/controllers/webhook_controller.go`
- Modify: `backend/app/services/billing/payment_service.go`
- Modify: `backend/app/services/billing/providers/provider.go`
- Modify: `backend/routes/web.go`
- Modify: `backend/app/services/plans/plan_service.go`
- Modify: `backend/app/services/quota/ledger.go`
- Test: `backend/app/services/plans/subscription_recovery_test.go`

**Interfaces:**
- Consumes: Task 2 Provider event types, PaymentIntent and Order records, existing subscription snapshot and quota ledger services.
- Produces: `WebhookService.Ingest`, `FulfillmentService.EnqueuePaidOrder`, `FulfillmentService.Process`; webhook routes under `/api/v1/payment-gateways/{gateway}/webhook` and verified event-to-fulfillment behavior.

- [ ] **Step 1: Write failing webhook and fulfillment tests**

  Cover invalid signature, duplicate event ID, amount/currency mismatch, successful event creating exactly one credit transaction, pending event doing nothing, paid order with failed fulfillment remaining paid and creating a retry task, and repeated task execution not creating a second Subscription or quota ledger entry.

- [ ] **Step 2: Run the tests in RED**

  ```powershell
  go test ./app/services/billing ./app/services/plans ./app/services/quota -run "Webhook|Fulfillment|Paid|Duplicate" -count=1
  ```

  Expected: FAIL because ingestion and fulfillment are not implemented.

- [ ] **Step 3: Implement verified event ingestion**

  Read and size-limit the raw request body, call the Provider verifier, hash and persist the event, and return an idempotent accepted response for an existing `(gateway_code, event_id)`. In a transaction lock the PaymentIntent and Order, validate amount/currency/external ID, append a PaymentTransaction, and transition only allowed states. Do not perform email, image processing or Provider calls inside the transaction.

- [ ] **Step 4: Implement paid-order fulfillment**

  After commit, create an idempotent `FulfillmentTask`. The processor re-reads and locks the order, creates/updates the subscription using the order entitlement snapshot, writes quota/advertising changes through existing services, and marks the task fulfilled. Errors retain `Order=paid`, set fulfillment pending/retry metadata, and never turn a verified payment back into unpaid.

- [ ] **Step 5: Add Fake webhook route and provider registration**

  Register Fake Provider in an explicit development registry, add a local-only test endpoint or service helper guarded by environment configuration, and keep external webhook routes unauthenticated but signature-protected. Do not expose a production fake transition endpoint.

- [ ] **Step 6: Run focused and full backend tests**

  ```powershell
  go test ./app/services/billing ./app/services/plans ./app/services/quota ./app/modules/billing -count=1
  go test ./app/... -count=1
  ```

  Expected: billing tests PASS; the full command must either PASS or report only the pre-existing Redis/environment blocker with the exact package and log.

- [ ] **Step 7: Commit the accounting/fulfillment slice**

  ```powershell
  git add backend/app/services/billing backend/app/services/plans backend/app/services/quota backend/app/modules/billing backend/routes/web.go
  git commit -m "feat: add idempotent payment events and subscription fulfillment"
  ```

### Task 4: Refunds, Financial Ledger and Reconciliation Primitives

**Files:**
- Create: `backend/app/services/billing/refund_service.go`
- Create: `backend/app/services/billing/refund_service_test.go`
- Create: `backend/app/services/billing/financial_ledger.go`
- Create: `backend/app/services/billing/financial_ledger_test.go`
- Create: `backend/app/services/billing/reconciliation_service.go`
- Create: `backend/app/services/billing/reconciliation_service_test.go`
- Create: `backend/app/modules/billing/controllers/admin_finance_controller.go`
- Modify: `backend/app/services/billing/providers/provider.go`
- Modify: `backend/routes/web.go`
- Modify: `docs/fastimg-payment-reconciliation.md`

**Interfaces:**
- Consumes: Task 3 verified transactions and registered Provider refund/query methods.
- Produces: `RefundService.Request`, `RefundService.ApplyVerifiedResult`, `FinancialLedger.Append`, `ReconciliationService.Match`; admin APIs for refunds, manual adjustments, reconciliation records and fulfillment retry.

- [ ] **Step 1: Write failing refund/ledger tests**

  Cover full and partial refunds, cumulative refund upper bound, provider pending refund, provider failure, duplicate refund event, chargeback as a separate debit/negative fact, and manual adjustment requiring reason/operator/permission.

- [ ] **Step 2: Run RED tests**

  ```powershell
  go test ./app/services/billing -run "Refund|Ledger|Reconciliation" -count=1
  ```

  Expected: FAIL because the services do not exist.

- [ ] **Step 3: Implement append-only financial ledger**

  Add typed ledger append operations for payment credit, refund debit, chargeback debit, gateway fee and manual adjustment. Enforce non-negative amount inputs, currency consistency, source idempotency and a balance query derived from immutable entries. Do not update a previous financial fact.

- [ ] **Step 4: Implement refund orchestration**

  Validate order/payment ownership and refundable balance, create a Refund and negative transaction, call Provider through a unique idempotency key, and apply only verified completed/pending/failed responses. Member refund requests must be own-scope; admin approval/override requires separate permissions and audit.

- [ ] **Step 5: Implement reconciliation matching**

  Normalize provider records to internal keys, match by transaction ID/order/invoice then amount-currency-time window, and persist `matched`, `missing_internal`, `missing_provider`, `amount_mismatch`, `currency_mismatch`, `duplicate`, `refund_mismatch` or `compliance_review`. Matching never activates benefits automatically.

- [ ] **Step 6: Add protected admin APIs and tests**

  Add `GET /api/v1/admin/orders`, `GET /api/v1/admin/payment-transactions`, `GET /api/v1/admin/payment-webhook-events`, `GET /api/v1/admin/refunds`, `POST /api/v1/admin/orders/{id}/retry-fulfillment`, `POST /api/v1/admin/orders/{id}/manual-adjustment` and reconciliation endpoints behind dedicated permissions. Test unauthorized, own member, authorized admin and audit outcomes.

- [ ] **Step 7: Run tests and update reconciliation docs**

  ```powershell
  go test ./app/services/billing ./app/modules/billing -count=1
  ```

  Update `docs/fastimg-payment-reconciliation.md` with implemented fields and explicitly list provider settlement import as pending if it is not yet wired.

- [ ] **Step 8: Commit the finance slice**

  ```powershell
  git add backend/app/services/billing backend/app/modules/billing backend/routes/web.go docs/fastimg-payment-reconciliation.md
  git commit -m "feat: add refunds financial ledger and reconciliation primitives"
  ```

### Task 5: Admin Resources, Permissions, Navigation and i18n

**Files:**
- Create: `backend/app/modules/billing/resource/order_manifest.go`
- Create: `backend/app/modules/billing/resource/payment_transaction_manifest.go`
- Create: `backend/app/modules/billing/resource/refund_manifest.go`
- Create: `backend/app/modules/billing/resource/payment_event_manifest.go`
- Create: `backend/app/modules/billing/resource/permissions.go`
- Create: `backend/app/modules/billing/resource/manifest_test.go`
- Modify: `backend/app/modules/admin/registry/registry.go`
- Modify: `backend/app/modules/admin/registry/generated_resources.go`
- Create: `admin/src/modules/billing/admin.ts`
- Create: `admin/src/modules/billing/api.ts`
- Create: `admin/src/modules/billing/pages/AdminOrdersPage.vue`
- Create: `admin/src/modules/billing/pages/AdminPaymentTransactionsPage.vue`
- Create: `admin/src/modules/billing/pages/AdminReconciliationPage.vue`
- Modify: `admin/src/router/index.ts`
- Modify: `admin/src/router/index.js`
- Create: `admin/src/locales/zh-CN/billing.json`
- Create: `admin/src/locales/en-US/billing.json`
- Modify: `admin/src/i18n/index.ts`
- Modify: `admin/src/i18n/index.js`
- Modify: `docs/fastimg-frontend-design.md`

**Interfaces:**
- Consumes: Task 4 admin endpoints, existing Resource Registry and AdminShell permission guard.
- Produces: `/admin/orders`, `/admin/payment-transactions`, `/admin/payment-events`, `/admin/refunds`, `/admin/reconciliation`; permissions `admin.orders.view`, `admin.payment_transactions.view`, `admin.payment_events.view`, `admin.refunds.manage`, `admin.reconciliation.view`, `admin.billing.fulfill`.

- [ ] **Step 1: Write registry and route tests**

  Add Go tests asserting every billing resource has an `/admin/...` route, explicit permission, safe readable fields and no raw secret/payload field. Add frontend route tests or static assertions proving no billing page is registered under the member root.

- [ ] **Step 2: Run RED tests**

  ```powershell
  go test ./app/modules/billing/resource ./app/modules/admin/registry -run "Manifest|Billing" -count=1
  ```

  Expected: FAIL until resource manifests and registration are added.

- [ ] **Step 3: Implement resource manifests and permissions**

  Use Resource Engine for list/detail/filter/export-safe read models. Provider IDs, status, amounts, currency, dates, user ID and order number are readable; secrets, signatures, full payloads, API keys and private customer data are masked or excluded. Destructive/financial actions are explicit custom actions routed through billing services, never generic raw field updates.

- [ ] **Step 4: Implement admin pages and navigation**

  Add compact tables with filters for order number/user/status/provider/date, transaction direction/type, refund status and reconciliation difference. Use existing layout/components; do not create a second admin shell. Add links from order detail to attempts, events, transactions, refund and fulfillment state.

- [ ] **Step 5: Add Chinese/English locale namespaces**

  Add complete matching key sets for navigation, status labels, empty/error states, amounts, retry/refund confirmations and security notices. Replace literal visible strings in new Vue pages with `t('billing...')`; unknown enum values must fall back to a safe label.

- [ ] **Step 6: Run frontend/backend checks**

  ```powershell
  go test ./app/modules/billing/resource ./app/modules/admin/registry -count=1
  pnpm exec vue-tsc -b
  pnpm run build
  ```

  Expected: backend and existing project build checks pass; if the known Windows `vue-tsc`/dependency blocker remains, record exact output without changing dependency versions.

- [ ] **Step 7: Commit the admin slice**

  ```powershell
  git add backend/app/modules/billing backend/app/modules/admin/registry admin/src/modules/billing admin/src/router/index.ts admin/src/router/index.js admin/src/locales/zh-CN/billing.json admin/src/locales/en-US/billing.json admin/src/i18n docs/fastimg-frontend-design.md
  git commit -m "feat: add billing admin resources and localized pages"
  ```

### Task 6: Member Plans Checkout, Orders and Payment Status UI

**Files:**
- Create: `admin/src/modules/billing/api.ts`
- Create: `admin/src/modules/billing/types.ts`
- Create: `admin/src/modules/member/pages/MemberCheckoutPage.vue`
- Create: `admin/src/modules/member/pages/MemberOrdersPage.vue`
- Create: `admin/src/modules/member/pages/MemberOrderDetailPage.vue`
- Modify: `admin/src/modules/member/pages/MemberPlansPage.vue`
- Modify: `admin/src/router/index.ts`
- Modify: `admin/src/router/index.js`
- Modify: `admin/src/locales/zh-CN/member.json`
- Modify: `admin/src/locales/en-US/member.json`
- Modify: `admin/src/modules/member/components/MemberNavigation.vue`
- Test: `admin/src/modules/billing/billing-ui.test.mjs`
- Modify: `docs/fastimg-frontend-design.md`

**Interfaces:**
- Consumes: Task 2 member order/payment APIs and Task 5 billing locale/i18n conventions.
- Produces: member routes `/checkout/:orderId`, `/orders`, `/orders/:id`; plan cards with Free/paid distinction; checkout return/status polling without client-side fulfillment.

- [ ] **Step 1: Write UI contract tests**

  Test that guest users can browse `/plans` but checkout/upload actions require login, Free has no payment button, paid checkout sends only plan price ID/period/currency choice allowed by server, checkout shows provider URL/status, and order detail never renders provider secrets or raw webhook data.

- [ ] **Step 2: Run the UI tests in RED**

  ```powershell
  node admin/src/modules/billing/billing-ui.test.mjs
  ```

  Expected: FAIL until the billing pages/API helpers exist.

- [ ] **Step 3: Implement member billing API helpers**

  Add typed functions for listing plans/prices, creating an order with `Idempotency-Key`, listing/getting own orders, starting payment, canceling unpaid orders and requesting a refund. Normalize API errors through the existing API client; never put tokens or payment credentials in localStorage or URLs.

- [ ] **Step 4: Implement checkout and order pages**

  Add a concise checkout summary with plan/period/amount/currency snapshot, allowed gateway selection, terms acknowledgement, and external checkout button. After return, poll the own order endpoint with bounded backoff and display `payment pending`, `paid/fulfillment pending`, `fulfilled`, `failed`, `expired` or `canceled`. A successful browser redirect alone must never activate the plan.

- [ ] **Step 5: Integrate plan cards and member navigation**

  Keep Free as a clear usable option and show paid CTAs only when active prices are returned. Add Orders to MemberShell navigation without exposing admin resources. Update current plan/usage after fulfillment refreshes subscription data.

- [ ] **Step 6: Run frontend checks and locale parity**

  ```powershell
  node admin/src/modules/billing/billing-ui.test.mjs
  pnpm exec vue-tsc -b
  ```

  Expected: PASS for focused UI tests and type checking, or preserve exact pre-existing build blocker evidence.

- [ ] **Step 7: Commit the member checkout slice**

  ```powershell
  git add admin/src/modules/billing admin/src/modules/member admin/src/router/index.ts admin/src/router/index.js admin/src/locales/zh-CN/member.json admin/src/locales/en-US/member.json docs/fastimg-frontend-design.md
  git commit -m "feat: add member checkout and order status pages"
  ```

### Task 7: Xcash Hosted Invoice Adapter

**Files:**
- Create: `backend/app/services/billing/providers/xcash/client.go`
- Create: `backend/app/services/billing/providers/xcash/signature.go`
- Create: `backend/app/services/billing/providers/xcash/provider.go`
- Create: `backend/app/services/billing/providers/xcash/provider_test.go`
- Create: `backend/app/modules/billing/controllers/xcash_webhook_controller.go`
- Modify: `backend/app/services/billing/providers/registry.go`
- Modify: `backend/routes/web.go`
- Create: `backend/config/payment.go`
- Modify: `docs/fastimg-payment-provider-xcash.md`

**Interfaces:**
- Consumes: Task 2-3 Provider interface and webhook ingestion.
- Produces: `xcash.Provider`; HMAC-SHA256 request signing with `XC-Appid`, `XC-Timestamp`, `XC-Nonce`, `XC-Signature`; hosted invoice creation/query/refund mapping; `/api/v1/payment-gateways/xcash/webhook`.

- [ ] **Step 1: Write signature and mapping tests**

  Pin exact signature bytes (`nonce + timestamp + raw JSON body`), lowercase hex HMAC-SHA256, constant-time comparison, timestamp window, missing-header rejection, and mapping of completed/confirmed versus pending/expired/underpaid/overpaid/wrong-network/risk states.

- [ ] **Step 2: Run Xcash tests in RED**

  ```powershell
  go test ./app/services/billing/providers/xcash -count=1
  ```

  Expected: FAIL until the adapter exists.

- [ ] **Step 3: Implement an HTTP client with bounded timeout**

  Read `XCASH_API_BASE_URL`, `XCASH_APP_ID`, `XCASH_HMAC_KEY`, timeout and allowed methods from config. Serialize once, sign the exact bytes sent, use a request context timeout, redact headers/body in errors, and persist only provider IDs/metadata hashes required by the domain.

- [ ] **Step 4: Implement hosted invoice creation and status mapping**

  Map internal order number to `out_no`; generate methods from server-side enabled asset/network config; store `sys_no`, `pay_url`, expiry, selected chain/asset/address/amount and payload hash. Only verified completed/confirmed states produce `PaymentIntent.succeeded`.

- [ ] **Step 5: Wire webhook and offline gate**

  Register the adapter only when `XCASH_ENABLED=true`; otherwise return a stable `GATEWAY_UNAVAILABLE` response. Add the webhook controller to pass raw body and headers to `WebhookService`, with no member authentication requirement but mandatory Provider verification.

- [ ] **Step 6: Run tests and config lint**

  ```powershell
  go test ./app/services/billing/providers/xcash ./app/services/billing ./app/modules/billing -count=1
  ```

  Expected: PASS with no network call when disabled; sandbox calls are not claimed unless credentials and an external sandbox run are explicitly supplied.

- [ ] **Step 7: Commit the Xcash adapter**

  ```powershell
  git add backend/app/services/billing/providers/xcash backend/app/services/billing/providers/registry.go backend/app/modules/billing/controllers/xcash_webhook_controller.go backend/config/payment.go backend/routes/web.go docs/fastimg-payment-provider-xcash.md
  git commit -m "feat: add gated Xcash hosted invoice provider"
  ```

### Task 8: NOWPayments Hosted Invoice Adapter

**Files:**
- Create: `backend/app/services/billing/providers/nowpayments/client.go`
- Create: `backend/app/services/billing/providers/nowpayments/signature.go`
- Create: `backend/app/services/billing/providers/nowpayments/provider.go`
- Create: `backend/app/services/billing/providers/nowpayments/provider_test.go`
- Create: `backend/app/modules/billing/controllers/nowpayments_webhook_controller.go`
- Modify: `backend/app/services/billing/providers/registry.go`
- Modify: `backend/routes/web.go`
- Modify: `backend/config/payment.go`
- Modify: `docs/fastimg-payment-provider-nowpayments.md`

**Interfaces:**
- Consumes: Task 2-3 Provider interface and webhook ingestion.
- Produces: `nowpayments.Provider`; hosted invoice creation, status query, HMAC-SHA512 `x-nowpayments-sig` verification, currency/network whitelist validation, and `/api/v1/payment-gateways/nowpayments/webhook`.

- [ ] **Step 1: Write signature/status tests**

  Cover canonical JSON signing with IPN Secret, constant-time comparison, invalid/missing signature, waiting/confirming/confirmed/finished/failed/expired mapping, and rejection of unsupported pay currency/network.

- [ ] **Step 2: Run RED tests**

  ```powershell
  go test ./app/services/billing/providers/nowpayments -count=1
  ```

  Expected: FAIL until the adapter exists.

- [ ] **Step 3: Implement API client and invoice mapping**

  Read `NOWPAYMENTS_API_BASE_URL`, API key, IPN secret, timeout, fixed-rate/fee flags and server-side allowed currencies. Create an invoice using the locked internal amount/currency and generated callback/return URLs; never accept arbitrary `pay_currency` from the browser.

- [ ] **Step 4: Implement IPN ingestion adapter**

  Preserve the raw body for signature validation, calculate payload hash, map `order_id`/invoice/payment IDs, validate amount/currency/asset/network and return normalized events. Do not treat any non-failed status as successful; only confirmed/finished states can fulfill.

- [ ] **Step 5: Run tests and disabled-gateway checks**

  ```powershell
  go test ./app/services/billing/providers/nowpayments ./app/services/billing ./app/modules/billing -count=1
  ```

  Expected: PASS, and disabled configuration never performs an outbound request.

- [ ] **Step 6: Commit the NOWPayments adapter**

  ```powershell
  git add backend/app/services/billing/providers/nowpayments backend/app/services/billing/providers/registry.go backend/app/modules/billing/controllers/nowpayments_webhook_controller.go backend/config/payment.go backend/routes/web.go docs/fastimg-payment-provider-nowpayments.md
  git commit -m "feat: add gated NOWPayments invoice provider"
  ```

### Task 9: PayPal Orders/Capture Adapter

**Files:**
- Create: `backend/app/services/billing/providers/paypal/client.go`
- Create: `backend/app/services/billing/providers/paypal/provider.go`
- Create: `backend/app/services/billing/providers/paypal/webhook.go`
- Create: `backend/app/services/billing/providers/paypal/provider_test.go`
- Create: `backend/app/modules/billing/controllers/paypal_webhook_controller.go`
- Modify: `backend/app/services/billing/providers/registry.go`
- Modify: `backend/routes/web.go`
- Modify: `backend/config/payment.go`
- Modify: `docs/fastimg-payment-provider-paypal.md`

**Interfaces:**
- Consumes: Task 2-4 Provider interface, refund service and webhook ingestion.
- Produces: `paypal.Provider`; OAuth token caching without logging secrets, Orders v2 create/capture/query/refund, `PayPal-Request-Id` idempotency and `/api/v1/payment-gateways/paypal/webhook`.

- [ ] **Step 1: Write PayPal HTTP and webhook tests**

  Cover sandbox/live base URL selection, token failure, create order with `intent=CAPTURE`, exact amount/currency/reference mapping, capture retry after timeout querying first, request-id reuse, webhook event mapping for completed/pending/denied/reversed and invalid webhook verification.

- [ ] **Step 2: Run RED tests**

  ```powershell
  go test ./app/services/billing/providers/paypal -count=1
  ```

  Expected: FAIL until the adapter exists.

- [ ] **Step 3: Implement OAuth and Orders v2 client**

  Read `PAYPAL_ENVIRONMENT`, client ID/secret, webhook ID and timeout. Cache access tokens only in process memory with expiry margin. Create an Order from the immutable local snapshot; use `PayPal-Request-Id` for create/capture/refund idempotency and never accept client amount.

- [ ] **Step 4: Implement capture/query/refund status mapping**

  On capture timeout, query the PayPal order/capture before retrying. Map `COMPLETED` to a verified success candidate, `PENDING` to pending, and denied/reversed to failed/reversed; pass all successful candidates through Task 3 validation before fulfillment.

- [ ] **Step 5: Implement webhook verification and route**

  Verify PayPal webhook authenticity using configured webhook ID and PayPal verification endpoint/SDK boundary, persist event ID for dedupe, and map only supported event types. Keep the route unauthenticated for Provider delivery but reject missing/invalid verification.

- [ ] **Step 6: Run tests and disabled-gateway checks**

  ```powershell
  go test ./app/services/billing/providers/paypal ./app/services/billing ./app/modules/billing -count=1
  ```

  Expected: PASS without network when disabled; no claim of live PayPal acceptance without a real sandbox account and evidence.

- [ ] **Step 7: Commit the PayPal adapter**

  ```powershell
  git add backend/app/services/billing/providers/paypal backend/app/services/billing/providers/registry.go backend/app/modules/billing/controllers/paypal_webhook_controller.go backend/config/payment.go backend/routes/web.go docs/fastimg-payment-provider-paypal.md
  git commit -m "feat: add gated PayPal orders and capture provider"
  ```

### Task 10: OpenAPI, Configuration, Observability and Release Gates

**Files:**
- Modify: `backend/app/openapi/spec.go`
- Modify: `backend/config/payment.go`
- Create: `backend/app/services/billing/health.go`
- Create: `backend/app/services/billing/health_test.go`
- Create: `backend/app/console/billing/reconcile_command.go`
- Create: `backend/app/console/billing/retry_fulfillment_command.go`
- Modify: `backend/app/http/middleware/http_audit.go`
- Modify: `docs/fastimg-payment-gateway.md`
- Modify: `docs/fastimg-production-readiness.md`
- Modify: `docs/fastimg-stage-development-plan.md`
- Modify: `.env.example`
- Modify: `backend/.env.example`
- Modify: `.superpowers/sdd/2026-09-23-fastimg-repository-development/progress.md`

**Interfaces:**
- Consumes: Tasks 1-9 public API, provider registry, finance/reconciliation and existing audit/console infrastructure.
- Produces: documented OpenAPI contracts, provider health/availability view, safe operational commands, migration/runbook evidence and a truthful M6/M7/M8 status ledger.

- [ ] **Step 1: Write health/config tests**

  Cover disabled providers, missing required secrets, invalid base URL, unsupported currency/network, stale webhook backlog, fulfillment retry count, and redaction of credentials in health/audit output.

- [ ] **Step 2: Run RED tests**

  ```powershell
  go test ./app/services/billing ./app/console/billing -run "Health|Config|Reconcile|Fulfillment" -count=1
  ```

  Expected: FAIL until operational helpers are implemented.

- [ ] **Step 3: Implement config validation and provider health**

  Add startup-safe configuration validation that reports disabled/misconfigured providers without preventing Free uploads. Health output includes provider code, enabled flag, environment, last successful probe metadata and circuit state, never client secrets.

- [ ] **Step 4: Add reconciliation and fulfillment commands**

  Implement idempotent commands that process only explicitly selected provider/date/order scopes, support dry-run, log counts/IDs without payload secrets, and use existing queue/transaction boundaries. Commands must not silently refund, activate or alter production data.

- [ ] **Step 5: Complete OpenAPI and documentation**

  Add schemas for prices, orders, payment intents, checkout responses, transactions, webhooks, refunds and error codes. Update stage plan with `implemented`, `integrated`, `sandbox_verified`, `production_blocked` per provider. Update release runbook with migration, rollback, webhook replay, secret rotation and local port checks.

- [ ] **Step 6: Run the release verification matrix**

  ```powershell
  gofmt -w backend/app backend/database/migrations
  go test ./app/... ./database/... -count=1
  pnpm exec vue-tsc -b
  pnpm run build
  ```

  Also run the provider package tests with network disabled and a local HTTP stub. Run the authorized local migration only against `fastimg_dev`, then restart the backend and verify Free upload plus Fake checkout/fulfillment. Do not call a real payment gateway unless the user separately provides sandbox credentials and explicitly authorizes the external test.

- [ ] **Step 7: Update progress and create the final milestone commit**

  Record exact commands/results, migration target, running ports, known build blockers and provider gates in the SDD progress ledger. Commit only after verification:

  ```powershell
  git add backend admin docs .env.example backend/.env.example .superpowers/sdd/2026-09-23-fastimg-repository-development/progress.md
  git commit -m "docs: complete payment gateway rollout gates and operations"
  ```

## Phase Gates and Delivery Order

1. **M6 Fake commercialization:** Tasks 1-6. Deliverable is a free/paid plan catalog, own-scope checkout, Fake Provider, idempotent order/payment/fulfillment, refunds/ledger primitives, admin finance pages and member order pages. It must work without any external provider credentials.
2. **M7 provider integration:** Tasks 7-9, one provider at a time. Each provider remains disabled until its unit tests, signature tests, local HTTP stub tests, webhook replay tests and sandbox evidence pass. A provider failure must not break Free uploads or other providers.
3. **M8 operations/reconciliation:** Task 10. Deliverable is documented observability, reconciliation, replay/repair commands, secret redaction, migration/runbook evidence and a provider-specific production decision. No production-ready claim is allowed from unit tests alone.

## Self-Review

- **Spec coverage:** Pricing/versioning is Task 1; order/payment/attempt/transaction/webhook/fulfillment is Tasks 2-3; refunds/financial ledger/reconciliation is Task 4; admin pages/RBAC/i18n is Task 5; member plans/checkout/orders is Task 6; Xcash/NOWPayments/PayPal are Tasks 7-9; operations/OpenAPI/gates are Task 10. Free behavior and provider isolation are global constraints and tested in Tasks 1-3 and 10.
- **Placeholder scan:** No task depends on TBD, an unspecified validation step, or a future unnamed function. Unsupported features such as tax invoices, settlement-file import and automated dispute appeals are explicitly out of scope rather than hidden work.
- **Type consistency:** All Providers implement the same `PaymentGateway` interface; `WebhookService` consumes normalized Provider events; `FulfillmentService` consumes verified internal payment facts; member/admin APIs remain separate from service contracts.
- **Review focus coverage:** Idempotency is Tasks 2-4/7-9; non-success state handling is Tasks 3/7-9; snapshots are Task 1; signatures and mismatch rejection are Tasks 3/7-9; RBAC/own-scope is Tasks 2/4/5/6.
