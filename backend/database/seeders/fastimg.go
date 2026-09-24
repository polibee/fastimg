package seeders

import (
	"encoding/json"

	"goravel/app/facades"
	"goravel/app/models"
	"goravel/app/services/quota"
)

type FastImg struct{}

func (s *FastImg) Signature() string {
	return "FastImg"
}

func (s *FastImg) Run() error {
	plan, err := buildFastImgFreePlan()
	if err != nil {
		return err
	}

	exists, err := facades.Orm().Query().Model(&models.Plan{}).Where("code = ?", plan.Code).Exists()
	if err != nil {
		return err
	}
	if !exists {
		if err := facades.Orm().Query().Create(&plan); err != nil {
			return err
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
