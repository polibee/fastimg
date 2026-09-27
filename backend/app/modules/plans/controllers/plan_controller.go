package controllers

import (
	"errors"
	"strconv"
	"time"

	"github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
	"goravel/app/models"
	billing "goravel/app/services/billing"
	planservices "goravel/app/services/plans"
	"goravel/app/services/quota"
	rbacservices "goravel/app/services/rbac"
)

type PlanController struct{}

func NewPlanController() *PlanController { return &PlanController{} }

func (p *PlanController) Index(ctx http.Context) http.Response {
	plans, err := planservices.NewPlanService().ActivePlans()
	if err != nil {
		return ctx.Response().Status(500).Json(http.Json{"code": "PLANS_UNAVAILABLE"})
	}
	data := make([]map[string]any, 0, len(plans))
	for _, plan := range plans {
		public, err := publicPlan(plan)
		if err != nil {
			return ctx.Response().Status(500).Json(http.Json{"code": "PLANS_UNAVAILABLE"})
		}
		prices, err := billing.NewPriceCatalog().ActivePricesForPlan(plan.ID)
		if err != nil {
			return ctx.Response().Status(500).Json(http.Json{"code": "PLANS_UNAVAILABLE"})
		}
		public["prices"] = publicPlanPrices(prices)
		data = append(data, public)
	}
	return ctx.Response().Success().Json(http.Json{"data": data})
}

func publicPlanPrices(prices []models.PlanPrice) []map[string]any {
	items := make([]map[string]any, 0, len(prices))
	for _, price := range prices {
		items = append(items, map[string]any{"id": price.ID, "version": price.Version, "currency": price.Currency, "amount_minor": price.AmountMinor, "billing_period": price.BillingPeriod, "trial_days": price.TrialDays, "status": price.Status})
	}
	return items
}

func (p *PlanController) Subscription(ctx http.Context) http.Response {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return ctx.Response().Status(401).Json(http.Json{"code": "AUTH_UNAUTHORIZED"})
	}
	subscription, err := planservices.NewPlanService().SubscriptionForUser(userID)
	if err != nil {
		return ctx.Response().Status(500).Json(http.Json{"code": "SUBSCRIPTION_UNAVAILABLE"})
	}
	snapshot, err := planservices.ParseSubscriptionSnapshot(subscription.EntitlementSnapshotJSON)
	if err != nil {
		return ctx.Response().Status(500).Json(http.Json{"code": "SUBSCRIPTION_UNAVAILABLE"})
	}
	var public map[string]any
	if snapshot.HasPlan {
		public = map[string]any{
			"id": snapshot.Plan.ID, "code": snapshot.Plan.Code, "name": snapshot.Plan.Name,
			"description":  snapshot.Plan.Description,
			"entitlements": snapshot.Entitlements,
		}
	} else {
		// Older subscriptions persisted entitlements only. Preserve their quota
		// behavior, but identify that plan display terms are not historically known.
		var plan models.Plan
		if err := facades.Orm().Query().Find(&plan, subscription.PlanID); err != nil {
			return ctx.Response().Status(500).Json(http.Json{"code": "PLAN_NOT_FOUND"})
		}
		public, err = publicPlan(plan)
		if err != nil {
			return ctx.Response().Status(500).Json(http.Json{"code": "PLANS_UNAVAILABLE"})
		}
	}
	return ctx.Response().Success().Json(http.Json{"data": map[string]any{
		"subscription":       subscription,
		"plan":               public,
		"plan_version":       snapshot.PlanVersion,
		"snapshot_available": snapshot.HasPlan,
	}})
}

func (p *PlanController) Usage(ctx http.Context) http.Response {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return ctx.Response().Status(401).Json(http.Json{"code": "AUTH_UNAUTHORIZED"})
	}
	var entries []models.UsageLedger
	if err := facades.Orm().Query().Where("user_id = ?", userID).Get(&entries); err != nil {
		return ctx.Response().Status(500).Json(http.Json{"code": "USAGE_UNAVAILABLE"})
	}
	usageEntries := make([]quota.UsageEntry, 0, len(entries))
	for _, entry := range entries {
		usageEntries = append(usageEntries, quota.UsageEntry{
			ResourceType: entry.ResourceType, Delta: entry.Delta, PeriodKey: entry.PeriodKey,
		})
	}
	subscription, err := planservices.NewPlanService().SubscriptionForUser(userID)
	if err != nil {
		return ctx.Response().Status(500).Json(http.Json{"code": "SUBSCRIPTION_UNAVAILABLE"})
	}
	entitlements, err := quota.ParseEntitlementJSON(subscription.EntitlementSnapshotJSON)
	if err != nil {
		return ctx.Response().Status(500).Json(http.Json{"code": "SUBSCRIPTION_UNAVAILABLE"})
	}
	administrator, err := rbacservices.NewRBACService().IsAdministrator(userID)
	if err != nil {
		return ctx.Response().Status(500).Json(http.Json{"code": "USAGE_UNAVAILABLE"})
	}
	if administrator {
		// Zero is the existing contract for unlimited member entitlements.
		entitlements = quota.Entitlement{}
	}
	periodKey := time.Now().UTC().Format("2006-01")
	usage := quota.AggregateUsage(usageEntries, periodKey)
	return ctx.Response().Success().Json(http.Json{"data": http.Json{
		"user_id": userID, "period_key": periodKey, "usage": usage, "limits": entitlements, "administrator": administrator,
		"bandwidth_metered": true,
	}})
}

type adminSubscriptionPayload struct {
	PlanID uint `json:"plan_id"`
}

func (p *PlanController) AdminSubscription(ctx http.Context) http.Response {
	userID := uint(ctx.Request().RouteInt64("id"))
	if userID == 0 {
		return ctx.Response().Status(404).Json(http.Json{"code": "RBAC_USER_NOT_FOUND"})
	}
	subscription, err := planservices.NewPlanService().SubscriptionForUser(userID)
	if err != nil {
		return ctx.Response().Status(500).Json(http.Json{"code": "SUBSCRIPTION_UNAVAILABLE"})
	}
	plans, err := planservices.NewPlanService().ActivePlans()
	if err != nil {
		return ctx.Response().Status(500).Json(http.Json{"code": "PLANS_UNAVAILABLE"})
	}
	return ctx.Response().Success().Json(http.Json{"data": adminSubscriptionData(subscription, plans)})
}

func (p *PlanController) UpdateAdminSubscription(ctx http.Context) http.Response {
	userID := uint(ctx.Request().RouteInt64("id"))
	if userID == 0 {
		return ctx.Response().Status(404).Json(http.Json{"code": "RBAC_USER_NOT_FOUND"})
	}
	var payload adminSubscriptionPayload
	if err := ctx.Request().Bind(&payload); err != nil || payload.PlanID == 0 {
		return ctx.Response().Status(422).Json(http.Json{"code": "VALIDATION_ERROR"})
	}
	subscription, err := planservices.NewPlanService().AssignPlan(userID, payload.PlanID)
	if errors.Is(err, planservices.ErrPlanNotFound) {
		return ctx.Response().Status(422).Json(http.Json{"code": "PLAN_NOT_FOUND"})
	}
	if errors.Is(err, planservices.ErrInvalidSubscriptionOwner) {
		return ctx.Response().Status(404).Json(http.Json{"code": "RBAC_USER_NOT_FOUND"})
	}
	if err != nil {
		return ctx.Response().Status(500).Json(http.Json{"code": "SUBSCRIPTION_UPDATE_FAILED"})
	}
	plans, err := planservices.NewPlanService().ActivePlans()
	if err != nil {
		return ctx.Response().Status(500).Json(http.Json{"code": "PLANS_UNAVAILABLE"})
	}
	return ctx.Response().Success().Json(http.Json{"data": adminSubscriptionData(subscription, plans)})
}

func adminSubscriptionData(subscription *models.Subscription, plans []models.Plan) map[string]any {
	items := make([]map[string]any, 0, len(plans))
	for _, plan := range plans {
		public, err := publicPlan(plan)
		if err == nil {
			items = append(items, public)
		}
	}
	return map[string]any{
		"subscription": map[string]any{"id": subscription.ID, "user_id": subscription.UserID, "plan_id": subscription.PlanID, "status": subscription.Status, "starts_at": subscription.StartsAt, "ends_at": subscription.EndsAt, "grace_period_ends_at": subscription.GracePeriodEndsAt},
		"plans":        items,
	}
}

func (p *PlanController) UsageLedger(ctx http.Context) http.Response {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return ctx.Response().Status(401).Json(http.Json{"code": "AUTH_UNAUTHORIZED"})
	}
	page, _ := strconv.Atoi(ctx.Request().Query("page", "1"))
	perPage, _ := strconv.Atoi(ctx.Request().Query("per_page", "20"))
	ledger, err := quota.ListUsageLedger(userID, page, perPage)
	if err != nil {
		return ctx.Response().Status(500).Json(http.Json{"code": "USAGE_UNAVAILABLE"})
	}
	return ctx.Response().Success().Json(http.Json{
		"data": ledger.Data,
		"meta": http.Json{
			"page": ledger.Page, "per_page": ledger.PerPage,
			"total": ledger.Total, "last_page": ledger.LastPage,
		},
	})
}

func publicPlan(plan models.Plan) (map[string]any, error) {
	result := map[string]any{
		"id": plan.ID, "code": plan.Code, "name": plan.Name, "description": plan.Description,
		"status": plan.Status, "sort_order": plan.SortOrder,
	}
	entitlements, err := quota.ParseEntitlementJSON(plan.EntitlementsJSON)
	if err != nil {
		return nil, err
	}
	result["entitlements"] = entitlements
	return result, nil
}

func authenticatedUserID(ctx http.Context) (uint, error) {
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
