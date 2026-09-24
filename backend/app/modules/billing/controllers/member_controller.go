package controllers

import (
	"errors"
	"net/http"
	"strconv"

	httpcontract "github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
	billing "goravel/app/services/billing"
	"goravel/app/services/billing/providers"
)

type MemberController struct {
	orders   *billing.OrderService
	payments *billing.PaymentService
}

func NewMemberController() *MemberController {
	return &MemberController{orders: billing.NewOrderService(), payments: billing.NewPaymentService(billing.DefaultGatewayRegistry())}
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
	return ctx.Response().Status(http.StatusCreated).Json(httpcontract.Json{"data": order})
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
	return ctx.Response().Success().Json(httpcontract.Json{"data": orders, "meta": httpcontract.Json{"page": page, "per_page": perPage, "total": total}})
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
	return ctx.Response().Success().Json(httpcontract.Json{"data": order})
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
