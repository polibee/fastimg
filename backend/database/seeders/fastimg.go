package seeders

import (
	"encoding/json"
	"time"

	"goravel/app/facades"
	"goravel/app/models"
	"goravel/app/services/quota"
)

type FastImg struct{}

func (s *FastImg) Signature() string {
	return "FastImg"
}

func (s *FastImg) Run() error {
	plans, err := buildFastImgPlans()
	if err != nil {
		return err
	}
	for _, plan := range plans {
		persisted, err := ensureFastImgPlan(plan)
		if err != nil {
			return err
		}
		if persisted.Code != "free" {
			if err := ensureFastImgPlanPrices(persisted); err != nil {
				return err
			}
		}
	}

	for _, permission := range []models.Permission{
		{Name: "admin.plans.view", DisplayName: "Plans.view"},
		{Name: "admin.plans.create", DisplayName: "Plans.create"},
		{Name: "admin.plans.update", DisplayName: "Plans.update"},
		{Name: "admin.plans.delete", DisplayName: "Plans.delete"},
		{Name: "admin.orders.view", DisplayName: "Orders.view"},
		{Name: "admin.payment_transactions.view", DisplayName: "Payment transactions.view"},
		{Name: "admin.payment_events.view", DisplayName: "Payment events.view"},
		{Name: "admin.refunds.view", DisplayName: "Refunds.view"},
		{Name: "admin.billing.fulfill", DisplayName: "Billing.fulfill"},
	} {
		exists, err := facades.Orm().Query().Model(&models.Permission{}).Where("name = ?", permission.Name).Exists()
		if err != nil {
			return err
		}
		if !exists {
			if err := facades.Orm().Query().Create(&permission); err != nil {
				return err
			}
		}
	}

	return nil
}

func ensureFastImgPlan(plan models.Plan) (models.Plan, error) {
	exists, err := facades.Orm().Query().Model(&models.Plan{}).Where("code = ?", plan.Code).Exists()
	if err != nil {
		return models.Plan{}, err
	}
	if !exists {
		if err := facades.Orm().Query().Create(&plan); err != nil {
			return models.Plan{}, err
		}
		return plan, nil
	}
	var persisted models.Plan
	if err := facades.Orm().Query().Model(&models.Plan{}).Where("code = ?", plan.Code).First(&persisted); err != nil {
		return models.Plan{}, err
	}
	return persisted, nil
}

func ensureFastImgPlanPrices(plan models.Plan) error {
	if !facades.Schema().HasTable("plan_prices") {
		return nil
	}
	for _, price := range []models.PlanPrice{
		{PlanID: plan.ID, Version: "v1-monthly-cny", Currency: "CNY", AmountMinor: plan.PriceAmount, BillingPeriod: "monthly", Status: "active"},
		{PlanID: plan.ID, Version: "v1-yearly-cny", Currency: "CNY", AmountMinor: plan.PriceAmount * 10, BillingPeriod: "yearly", Status: "active"},
	} {
		exists, err := facades.Orm().Query().Model(&models.PlanPrice{}).Where("plan_id = ? AND version = ?", price.PlanID, price.Version).Exists()
		if err != nil {
			return err
		}
		if exists {
			continue
		}
		price.EffectiveFrom = time.Now().UTC()
		if err := facades.Orm().Query().Create(&price); err != nil {
			return err
		}
	}
	return nil
}

func buildFastImgPlans() ([]models.Plan, error) {
	free, err := buildFastImgFreePlan()
	if err != nil {
		return nil, err
	}
	creator, err := buildFastImgPaidPlan("creator", "Creator", "For creators who need more storage and API capacity", 1999, 20, 20000, 1000, false, 20)
	if err != nil {
		return nil, err
	}
	pro, err := buildFastImgPaidPlan("pro", "Pro", "For sites and teams with higher delivery limits", 4999, 100, 100000, 5000, false, 100)
	if err != nil {
		return nil, err
	}
	return []models.Plan{free, creator, pro}, nil
}

func buildFastImgPaidPlan(code, name, description string, amountMinor, storageGB, maxFileMB, dailyUploads int64, adsEnabled bool, tokenLimit int64) (models.Plan, error) {
	entitlement := quota.DefaultFreeEntitlement()
	entitlement.StorageBytes = storageGB * 1_000_000_000
	entitlement.MaxFileBytes = maxFileMB * 1_000_000
	entitlement.DailyUploads = dailyUploads
	entitlement.MonthlyAPIUploads = dailyUploads * 20
	entitlement.MonthlyBandwidthBytes = storageGB * 20_000_000_000
	entitlement.TransformCount = dailyUploads * 10
	entitlement.APIRatePerMinute = dailyUploads
	entitlement.TokenLimit = tokenLimit
	entitlement.AdsEnabled = adsEnabled
	entitlement.WatermarkEnabled = true
	encoded, err := json.Marshal(entitlement)
	if err != nil {
		return models.Plan{}, err
	}
	return models.Plan{
		Code: code, Name: name, Description: description, PriceAmount: amountMinor,
		Currency: "CNY", BillingPeriod: "monthly", EntitlementsJSON: string(encoded),
		Status: "active", SortOrder: map[string]int{"creator": 20, "pro": 30}[code],
	}, nil
}

func buildFastImgFreePlan() (models.Plan, error) {
	encoded, err := json.Marshal(quota.DefaultFreeEntitlement())
	if err != nil {
		return models.Plan{}, err
	}

	return models.Plan{
		Code:             "free",
		Name:             "Free",
		Description:      "Free media hosting baseline",
		PriceAmount:      0,
		Currency:         "CNY",
		BillingPeriod:    "monthly",
		EntitlementsJSON: string(encoded),
		Status:           "active",
		SortOrder:        10,
	}, nil
}
