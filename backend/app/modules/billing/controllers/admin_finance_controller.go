package controllers

import (
	"net/http"
	"strconv"
	"time"

	httpcontract "github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
	"goravel/app/models"
	billing "goravel/app/services/billing"
)

type AdminFinanceController struct{}

func NewAdminFinanceController() *AdminFinanceController { return &AdminFinanceController{} }

func (c *AdminFinanceController) Orders(ctx httpcontract.Context) httpcontract.Response {
	var rows []models.Order
	var total int64
	page, perPage := pageParams(ctx)
	if err := facades.Orm().Query().OrderByDesc("id").Paginate(page, perPage, &rows, &total); err != nil {
		return financeFailure(ctx)
	}
	return paginatedFinance(ctx, rows, page, perPage, total)
}

func (c *AdminFinanceController) Transactions(ctx httpcontract.Context) httpcontract.Response {
	var rows []models.PaymentTransaction
	var total int64
	page, perPage := pageParams(ctx)
	if err := facades.Orm().Query().OrderByDesc("id").Paginate(page, perPage, &rows, &total); err != nil {
		return financeFailure(ctx)
	}
	return paginatedFinance(ctx, rows, page, perPage, total)
}

func (c *AdminFinanceController) WebhookEvents(ctx httpcontract.Context) httpcontract.Response {
	var rows []models.PaymentWebhookEvent
	var total int64
	page, perPage := pageParams(ctx)
	if err := facades.Orm().Query().OrderByDesc("id").Paginate(page, perPage, &rows, &total); err != nil {
		return financeFailure(ctx)
	}
	return paginatedFinance(ctx, rows, page, perPage, total)
}

func (c *AdminFinanceController) Refunds(ctx httpcontract.Context) httpcontract.Response {
	var rows []models.Refund
	var total int64
	page, perPage := pageParams(ctx)
	if err := facades.Orm().Query().OrderByDesc("id").Paginate(page, perPage, &rows, &total); err != nil {
		return financeFailure(ctx)
	}
	return paginatedFinance(ctx, rows, page, perPage, total)
}

func (c *AdminFinanceController) RetryFulfillment(ctx httpcontract.Context) httpcontract.Response {
	orderID := ctx.Request().RouteInt64("id")
	if orderID <= 0 {
		return ctx.Response().Status(http.StatusNotFound).Json(httpcontract.Json{"code": "ORDER_NOT_FOUND"})
	}
	service := billing.NewFulfillmentService()
	if err := service.EnqueuePaidOrder(uint(orderID)); err != nil {
		// Existing task is still retryable; Process will return the real error.
		_ = err
	}
	if err := service.Process(uint(orderID)); err != nil {
		return ctx.Response().Status(http.StatusConflict).Json(httpcontract.Json{"code": "FULFILLMENT_RETRY_FAILED"})
	}
	return ctx.Response().Success().Json(httpcontract.Json{"data": httpcontract.Json{"order_id": orderID, "status": "fulfilled"}})
}

func pageParams(ctx httpcontract.Context) (int, int) {
	page, _ := strconv.Atoi(ctx.Request().Query("page", "1"))
	perPage, _ := strconv.Atoi(ctx.Request().Query("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	return page, perPage
}

func paginatedFinance(ctx httpcontract.Context, data any, page, perPage int, total int64) httpcontract.Response {
	return ctx.Response().Success().Json(httpcontract.Json{"data": data, "meta": httpcontract.Json{"page": page, "per_page": perPage, "total": total}})
}

func financeFailure(ctx httpcontract.Context) httpcontract.Response {
	return ctx.Response().Status(http.StatusInternalServerError).Json(httpcontract.Json{"code": "FINANCE_UNAVAILABLE", "at": time.Now().UTC()})
}
