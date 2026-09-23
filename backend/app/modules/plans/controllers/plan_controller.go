package controllers

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
	"goravel/app/models"
	planservices "goravel/app/services/plans"
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
		data = append(data, publicPlan(plan))
	}
	return ctx.Response().Success().Json(http.Json{"data": data})
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
	var plan models.Plan
	if err := facades.Orm().Query().Find(&plan, subscription.PlanID); err != nil {
		return ctx.Response().Status(500).Json(http.Json{"code": "PLAN_NOT_FOUND"})
	}
	return ctx.Response().Success().Json(http.Json{"data": map[string]any{
		"subscription": subscription,
		"plan":         publicPlan(plan),
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
	usage := map[string]int64{}
	for _, entry := range entries {
		usage[entry.ResourceType] += entry.Delta
	}
	return ctx.Response().Success().Json(http.Json{"data": http.Json{"user_id": userID, "usage": usage}})
}

func publicPlan(plan models.Plan) map[string]any {
	result := map[string]any{
		"id": plan.ID, "code": plan.Code, "name": plan.Name, "description": plan.Description,
		"price_amount": plan.PriceAmount, "currency": plan.Currency,
		"billing_period": plan.BillingPeriod, "status": plan.Status, "sort_order": plan.SortOrder,
	}
	if plan.EntitlementsJSON != "" {
		var entitlements map[string]any
		if json.Unmarshal([]byte(plan.EntitlementsJSON), &entitlements) == nil {
			result["entitlements"] = entitlements
		}
	}
	return result
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
