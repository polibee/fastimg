package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/database/orm"
	httpcontract "github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
	adminmiddleware "goravel/app/http/middleware"
	"goravel/app/models"
	auditservices "goravel/app/services/audit"
	billing "goravel/app/services/billing"
)

var (
	ErrPlanPriceDeleteRequiresArchive = errors.New("plan price must be archived before deletion")
	ErrPlanPriceDeleteProtected       = errors.New("plan price is referenced by an order")
)

// PlanPriceController manages immutable commercial price versions from the
// plan editor. There is deliberately no standalone plan_prices resource/menu:
// a price only makes sense in the context of its member plan.
type PlanPriceController struct{}

type createPlanPriceRequest struct {
	AmountMinor   int64  `json:"amount_minor"`
	Currency      string `json:"currency"`
	BillingPeriod string `json:"billing_period"`
	TrialDays     int    `json:"trial_days"`
}

func NewPlanPriceController() *PlanPriceController { return &PlanPriceController{} }

func (c *PlanPriceController) Index(ctx httpcontract.Context) httpcontract.Response {
	planID := positiveRouteID(ctx, "id")
	if planID == 0 || !planExists(planID) {
		return adminmiddleware.APIError(ctx, http.StatusNotFound, "PLAN_NOT_FOUND")
	}
	var prices []models.PlanPrice
	if err := facades.Orm().Query().Where("plan_id = ?", planID).
		OrderBy("billing_period", "asc").OrderBy("currency", "asc").OrderByDesc("effective_from").Get(&prices); err != nil {
		return adminmiddleware.APIError(ctx, http.StatusInternalServerError, "PLAN_PRICES_UNAVAILABLE")
	}
	return ctx.Response().Success().Json(httpcontract.Json{"data": prices})
}

func (c *PlanPriceController) Create(ctx httpcontract.Context) httpcontract.Response {
	planID := positiveRouteID(ctx, "id")
	if planID == 0 || !planExists(planID) {
		return adminmiddleware.APIError(ctx, http.StatusNotFound, "PLAN_NOT_FOUND")
	}
	var input createPlanPriceRequest
	if err := ctx.Request().Bind(&input); err != nil {
		return adminmiddleware.APIError(ctx, http.StatusUnprocessableEntity, "PLAN_PRICE_INVALID")
	}
	input.Currency = strings.ToUpper(strings.TrimSpace(input.Currency))
	input.BillingPeriod = strings.ToLower(strings.TrimSpace(input.BillingPeriod))
	if input.TrialDays < 0 || input.TrialDays > 365 {
		return adminmiddleware.APIError(ctx, http.StatusUnprocessableEntity, "PLAN_PRICE_INVALID")
	}
	price := models.PlanPrice{
		PlanID: uint(planID), Currency: input.Currency, AmountMinor: input.AmountMinor,
		BillingPeriod: input.BillingPeriod, TrialDays: input.TrialDays,
		Status: "active", EffectiveFrom: time.Now().UTC(),
		Version: fmt.Sprintf("v%d-%s-%s", time.Now().UTC().UnixNano(), strings.ToLower(input.Currency), input.BillingPeriod),
	}
	if err := billing.ValidatePlanPrice(price); err != nil {
		return adminmiddleware.APIError(ctx, http.StatusUnprocessableEntity, "PLAN_PRICE_INVALID")
	}
	if err := facades.Orm().Transaction(func(tx orm.Query) error {
		now := time.Now().UTC()
		// Only one active price is selectable for a plan/currency/period. The
		// old row is archived, never edited, so existing order snapshots remain.
		if _, err := tx.Table("plan_prices").Where("plan_id = ? AND currency = ? AND billing_period = ? AND status = ?", planID, price.Currency, price.BillingPeriod, "active").Update(map[string]any{"status": "archived", "effective_to": now}); err != nil {
			return err
		}
		return tx.Create(&price)
	}); err != nil {
		return adminmiddleware.APIError(ctx, http.StatusInternalServerError, "PLAN_PRICE_SAVE_FAILED")
	}
	recordPlanPriceAudit(ctx, "plans.price.create", price)
	return ctx.Response().Status(http.StatusCreated).Json(httpcontract.Json{"data": price})
}

func (c *PlanPriceController) Archive(ctx httpcontract.Context) httpcontract.Response {
	planID := positiveRouteID(ctx, "id")
	priceID := positiveRouteID(ctx, "price_id")
	if planID == 0 || priceID == 0 {
		return adminmiddleware.APIError(ctx, http.StatusNotFound, "PLAN_PRICE_NOT_FOUND")
	}
	var price models.PlanPrice
	if err := facades.Orm().Query().Where("id = ? AND plan_id = ?", priceID, planID).First(&price); err != nil {
		return adminmiddleware.APIError(ctx, http.StatusNotFound, "PLAN_PRICE_NOT_FOUND")
	}
	if price.Status != "active" {
		return ctx.Response().Success().Json(httpcontract.Json{"data": price})
	}
	now := time.Now().UTC()
	if _, err := facades.Orm().Query().Table("plan_prices").Where("id = ? AND plan_id = ? AND status = ?", priceID, planID, "active").Update(map[string]any{"status": "archived", "effective_to": now}); err != nil {
		return adminmiddleware.APIError(ctx, http.StatusInternalServerError, "PLAN_PRICE_SAVE_FAILED")
	}
	price.Status = "archived"
	price.EffectiveTo = &now
	recordPlanPriceAudit(ctx, "plans.price.archive", price)
	return ctx.Response().Success().Json(httpcontract.Json{"data": price})
}

func (c *PlanPriceController) Delete(ctx httpcontract.Context) httpcontract.Response {
	planID := positiveRouteID(ctx, "id")
	priceID := positiveRouteID(ctx, "price_id")
	if planID == 0 || priceID == 0 {
		return adminmiddleware.APIError(ctx, http.StatusNotFound, "PLAN_PRICE_NOT_FOUND")
	}
	var price models.PlanPrice
	if err := facades.Orm().Query().Where("id = ? AND plan_id = ?", priceID, planID).First(&price); err != nil {
		return adminmiddleware.APIError(ctx, http.StatusNotFound, "PLAN_PRICE_NOT_FOUND")
	}
	referenced, err := facades.Orm().Query().Table("order_items").Where(
		"product_type = ? AND product_id = ? AND product_version = ?", "plan", planID, price.Version,
	).Exists()
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusInternalServerError, "PLAN_PRICE_DELETE_FAILED")
	}
	if err := canDeletePlanPrice(price.Status, referenced); err != nil {
		switch {
		case errors.Is(err, ErrPlanPriceDeleteRequiresArchive):
			return adminmiddleware.APIError(ctx, http.StatusConflict, "PLAN_PRICE_DELETE_REQUIRES_ARCHIVE")
		case errors.Is(err, ErrPlanPriceDeleteProtected):
			return adminmiddleware.APIError(ctx, http.StatusConflict, "PLAN_PRICE_DELETE_PROTECTED")
		default:
			return adminmiddleware.APIError(ctx, http.StatusConflict, "PLAN_PRICE_DELETE_FAILED")
		}
	}
	if _, err := facades.Orm().Query().Table("plan_prices").Where("id = ? AND plan_id = ?", priceID, planID).Delete(); err != nil {
		return adminmiddleware.APIError(ctx, http.StatusInternalServerError, "PLAN_PRICE_DELETE_FAILED")
	}
	recordPlanPriceAudit(ctx, "plans.price.delete", price)
	return ctx.Response().Success().Json(httpcontract.Json{"data": price})
}

func canDeletePlanPrice(status string, referenced bool) error {
	if strings.ToLower(strings.TrimSpace(status)) == "active" {
		return ErrPlanPriceDeleteRequiresArchive
	}
	if referenced {
		return ErrPlanPriceDeleteProtected
	}
	return nil
}

func planExists(id int64) bool {
	var plan models.Plan
	return facades.Orm().Query().Where("id = ?", id).First(&plan) == nil
}

func positiveRouteID(ctx httpcontract.Context, name string) int64 {
	value := ctx.Request().RouteInt64(name)
	if value < 1 {
		return 0
	}
	return value
}

func recordPlanPriceAudit(ctx httpcontract.Context, action string, price models.PlanPrice) {
	identity, err := facades.Auth(ctx).ID()
	if err != nil || strings.TrimSpace(identity) == "" {
		return
	}
	var operatorID uint
	_, _ = fmt.Sscan(identity, &operatorID)
	if operatorID > 0 {
		_ = auditservices.NewAuditService().Record(operatorID, action, map[string]any{
			"plan_id": price.PlanID, "price_id": price.ID, "currency": price.Currency,
			"billing_period": price.BillingPeriod, "amount_minor": price.AmountMinor,
		})
	}
}
