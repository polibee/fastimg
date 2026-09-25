package controllers

import (
	"errors"
	"net/http"
	"strconv"

	httpcontract "github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
	"goravel/app/models"
	billing "goravel/app/services/billing"
	"goravel/app/services/billing/providers"
	"goravel/app/services/billing/providers/fake"
)

type MemberController struct {
	orders   *billing.OrderService
	payments *billing.PaymentService
	gateways *providers.Registry
	webhooks *billing.WebhookService
}

func NewMemberController() *MemberController {
	gateways := billing.DefaultGatewayRegistry()
	return &MemberController{orders: billing.NewOrderService(), payments: billing.NewPaymentService(gateways), gateways: gateways, webhooks: billing.NewWebhookService(gateways)}
}

func (c *MemberController) Gateways(ctx httpcontract.Context) httpcontract.Response {
	return ctx.Response().Success().Json(httpcontract.Json{"data": c.gateways.Codes()})
}

func (c *MemberController) CreateOrder(ctx httpcontract.Context) httpcontract.Response {
	userID, err := memberUserID(ctx)
	if err != nil {
		return ctx.Response().Status(http.StatusUnauthorized).Json(httpcontract.Json{"code": "AUTH_UNAUTHORIZED"})
	}
	var input billing.CreateOrderRequest
	if err := ctx.Request().Bind(&input); err != nil {
		return ctx.Response().Status(http.StatusUnprocessableEntity).Json(httpcontract.Json{"code": "ORDER_VALIDATION_FAILED"})
	}
	input.IdempotencyKey = ctx.Request().Header("Idempotency-Key")
	order, err := c.orders.CreatePlanOrder(ctx.Context(), userID, input)
	if err != nil {
		return billingError(ctx, err)
	}
	return ctx.Response().Status(http.StatusCreated).Json(httpcontract.Json{"data": publicOrder(order)})
}

func (c *MemberController) ListOrders(ctx httpcontract.Context) httpcontract.Response {
	userID, err := memberUserID(ctx)
	if err != nil {
		return ctx.Response().Status(http.StatusUnauthorized).Json(httpcontract.Json{"code": "AUTH_UNAUTHORIZED"})
	}
	page, _ := strconv.Atoi(ctx.Request().Query("page", "1"))
	perPage, _ := strconv.Atoi(ctx.Request().Query("per_page", "20"))
	orders, total, err := c.orders.ListOwnOrders(ctx.Context(), userID, page, perPage)
	if err != nil {
		return ctx.Response().Status(http.StatusInternalServerError).Json(httpcontract.Json{"code": "ORDERS_UNAVAILABLE"})
	}
	data := make([]map[string]any, 0, len(orders))
	for _, order := range orders {
		data = append(data, publicOrder(&order))
	}
	return ctx.Response().Success().Json(httpcontract.Json{"data": data, "meta": httpcontract.Json{"page": page, "per_page": perPage, "total": total}})
}

func (c *MemberController) ShowOrder(ctx httpcontract.Context) httpcontract.Response {
	userID, err := memberUserID(ctx)
	if err != nil {
		return ctx.Response().Status(http.StatusUnauthorized).Json(httpcontract.Json{"code": "AUTH_UNAUTHORIZED"})
	}
	order, err := c.orders.GetOwnOrder(ctx.Context(), userID, uint(ctx.Request().RouteInt64("id")))
	if err != nil {
		return ctx.Response().Status(http.StatusNotFound).Json(httpcontract.Json{"code": "ORDER_NOT_FOUND"})
	}
	return ctx.Response().Success().Json(httpcontract.Json{"data": publicOrder(order)})
}

func (c *MemberController) StartPayment(ctx httpcontract.Context) httpcontract.Response {
	userID, err := memberUserID(ctx)
	if err != nil {
		return ctx.Response().Status(http.StatusUnauthorized).Json(httpcontract.Json{"code": "AUTH_UNAUTHORIZED"})
	}
	var input billing.StartPaymentRequest
	if err := ctx.Request().Bind(&input); err != nil {
		return ctx.Response().Status(http.StatusUnprocessableEntity).Json(httpcontract.Json{"code": "PAYMENT_VALIDATION_FAILED"})
	}
	intent, err := c.payments.StartPayment(ctx.Context(), userID, uint(ctx.Request().RouteInt64("id")), input, ctx.Request().Header("Idempotency-Key"))
	if err != nil {
		return billingError(ctx, err)
	}
	return ctx.Response().Status(http.StatusCreated).Json(httpcontract.Json{"data": intent})
}

func (c *MemberController) CancelOrder(ctx httpcontract.Context) httpcontract.Response {
	userID, err := memberUserID(ctx)
	if err != nil {
		return ctx.Response().Status(http.StatusUnauthorized).Json(httpcontract.Json{"code": "AUTH_UNAUTHORIZED"})
	}
	if err := c.orders.CancelOwnOrder(ctx.Context(), userID, uint(ctx.Request().RouteInt64("id"))); err != nil {
		return billingError(ctx, err)
	}
	return ctx.Response().NoContent(http.StatusNoContent)
}

func (c *MemberController) CompleteFakePayment(ctx httpcontract.Context) httpcontract.Response {
	if facades.Config().GetString("app.env", "production") == "production" {
		return ctx.Response().Status(http.StatusNotFound).Json(httpcontract.Json{"code": "FAKE_GATEWAY_DISABLED"})
	}
	userID, err := memberUserID(ctx)
	if err != nil {
		return ctx.Response().Status(http.StatusUnauthorized).Json(httpcontract.Json{"code": "AUTH_UNAUTHORIZED"})
	}
	orderID := uint(ctx.Request().RouteInt64("id"))
	if _, err := c.orders.GetOwnOrder(ctx.Context(), userID, orderID); err != nil {
		return ctx.Response().Status(http.StatusNotFound).Json(httpcontract.Json{"code": "ORDER_NOT_FOUND"})
	}
	var intent models.PaymentIntent
	exists, err := facades.Orm().Query().Model(&models.PaymentIntent{}).Where("order_id = ? AND user_id = ? AND provider_code = ?", orderID, userID, "fake").Exists()
	if err != nil || !exists || facades.Orm().Query().Where("order_id = ? AND user_id = ? AND provider_code = ?", orderID, userID, "fake").OrderByDesc("id").First(&intent) != nil {
		return ctx.Response().Status(http.StatusConflict).Json(httpcontract.Json{"code": "PAYMENT_INTENT_NOT_FOUND"})
	}
	gateway, err := c.gateways.Get("fake")
	if err != nil {
		return ctx.Response().Status(http.StatusConflict).Json(httpcontract.Json{"code": "GATEWAY_UNAVAILABLE"})
	}
	fakeGateway, ok := gateway.(*fake.Provider)
	if !ok {
		return ctx.Response().Status(http.StatusConflict).Json(httpcontract.Json{"code": "GATEWAY_UNAVAILABLE"})
	}
	event, err := fakeGateway.Transition(intent.ProviderPaymentID, "succeeded")
	if err != nil {
		return ctx.Response().Status(http.StatusConflict).Json(httpcontract.Json{"code": "PAYMENT_TRANSITION_FAILED"})
	}
	event, duplicate, err := c.webhooks.IngestVerified(ctx.Context(), "fake", event, []byte(event.EventID))
	if err != nil {
		return ctx.Response().Status(http.StatusConflict).Json(httpcontract.Json{"code": "PAYMENT_CONFIRMATION_FAILED"})
	}
	return ctx.Response().Success().Json(httpcontract.Json{"data": httpcontract.Json{"event_id": event.EventID, "accepted": true, "duplicate": duplicate}})
}

func memberUserID(ctx httpcontract.Context) (uint, error) {
	identity, err := facades.Auth(ctx).ID()
	if err != nil {
		return 0, err
	}
	parsed, err := strconv.ParseUint(identity, 10, 32)
	if err != nil || parsed == 0 {
		return 0, errors.New("invalid authenticated user")
	}
	return uint(parsed), nil
}

func billingError(ctx httpcontract.Context, err error) httpcontract.Response {
	switch {
	case errors.Is(err, billing.ErrInvalidOrderRequest):
		return ctx.Response().Status(http.StatusUnprocessableEntity).Json(httpcontract.Json{"code": "ORDER_VALIDATION_FAILED"})
	case errors.Is(err, billing.ErrFreePlanNotPayable):
		return ctx.Response().Status(http.StatusConflict).Json(httpcontract.Json{"code": "FREE_PLAN_NO_PAYMENT"})
	case errors.Is(err, billing.ErrPlanPriceNotFound):
		return ctx.Response().Status(http.StatusConflict).Json(httpcontract.Json{"code": "PLAN_PRICE_UNAVAILABLE"})
	case errors.Is(err, billing.ErrOrderNotFound):
		return ctx.Response().Status(http.StatusNotFound).Json(httpcontract.Json{"code": "ORDER_NOT_FOUND"})
	case errors.Is(err, billing.ErrPaymentNotAllowed):
		return ctx.Response().Status(http.StatusConflict).Json(httpcontract.Json{"code": "PAYMENT_NOT_ALLOWED"})
	case errors.Is(err, providers.ErrGatewayUnavailable):
		return ctx.Response().Status(http.StatusConflict).Json(httpcontract.Json{"code": "GATEWAY_UNAVAILABLE"})
	default:
		return ctx.Response().Status(http.StatusInternalServerError).Json(httpcontract.Json{"code": "BILLING_OPERATION_FAILED"})
	}
}

func publicOrder(order *models.Order) map[string]any {
	return map[string]any{"id": order.ID, "public_order_no": order.PublicOrderNo, "user_id": order.UserID, "status": order.Status, "currency": order.Currency, "total_amount_minor": order.TotalAmountMinor, "expires_at": order.ExpiresAt, "paid_at": order.PaidAt, "fulfilled_at": order.FulfilledAt, "created_at": order.CreatedAt}
}
