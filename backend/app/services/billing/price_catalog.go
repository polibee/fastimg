package billing

import (
	"errors"
	"fmt"
	"strings"

	"goravel/app/facades"
	"goravel/app/models"
	"goravel/app/services/quota"
)

var (
	ErrInvalidPlanPrice  = errors.New("invalid plan price")
	ErrPlanPriceNotFound = errors.New("active plan price not found")
)

type PlanPriceSnapshot struct {
	PlanID          uint
	PlanCode        string
	PlanName        string
	PlanDescription string
	PriceID         uint
	PriceVersion    string
	Currency        string
	AmountMinor     int64
	BillingPeriod   string
	TrialDays       int
	Entitlements    quota.Entitlement
}

func ValidatePlanPrice(price models.PlanPrice) error {
	if price.AmountMinor < 0 {
		return fmt.Errorf("%w: amount_minor cannot be negative", ErrInvalidPlanPrice)
	}
	if len(price.Currency) != 3 || price.Currency != strings.ToUpper(price.Currency) {
		return fmt.Errorf("%w: currency must be three uppercase letters", ErrInvalidPlanPrice)
	}
	for _, char := range price.Currency {
		if char < 'A' || char > 'Z' {
			return fmt.Errorf("%w: currency must be three uppercase letters", ErrInvalidPlanPrice)
		}
	}
	if price.BillingPeriod != "monthly" && price.BillingPeriod != "yearly" {
		return fmt.Errorf("%w: billing period is unsupported", ErrInvalidPlanPrice)
	}
	if price.Status != "draft" && price.Status != "active" && price.Status != "archived" {
		return fmt.Errorf("%w: status is unsupported", ErrInvalidPlanPrice)
	}
	if price.Version == "" || price.PlanID == 0 {
		return fmt.Errorf("%w: plan and version are required", ErrInvalidPlanPrice)
	}
	return nil
}

func BuildPlanPriceSnapshot(plan models.Plan, price models.PlanPrice) (PlanPriceSnapshot, error) {
	if err := ValidatePlanPrice(price); err != nil {
		return PlanPriceSnapshot{}, err
	}
	if plan.ID == 0 || price.PlanID != plan.ID {
		return PlanPriceSnapshot{}, fmt.Errorf("%w: price does not belong to plan", ErrInvalidPlanPrice)
	}
	entitlements, err := quota.ParseEntitlementJSON(plan.EntitlementsJSON)
	if err != nil {
		return PlanPriceSnapshot{}, fmt.Errorf("%w: plan entitlements: %v", ErrInvalidPlanPrice, err)
	}
	return PlanPriceSnapshot{
		PlanID: plan.ID, PlanCode: plan.Code, PlanName: plan.Name, PlanDescription: plan.Description,
		PriceID: price.ID, PriceVersion: price.Version, Currency: price.Currency,
		AmountMinor: price.AmountMinor, BillingPeriod: price.BillingPeriod, TrialDays: price.TrialDays,
		Entitlements: entitlements,
	}, nil
}

type PriceCatalog struct{}

func NewPriceCatalog() *PriceCatalog { return &PriceCatalog{} }

func (c *PriceCatalog) GetActivePrice(planID uint, currency, period string) (*models.PlanPrice, error) {
	if planID == 0 || len(currency) != 3 || period == "" {
		return nil, ErrInvalidPlanPrice
	}
	var price models.PlanPrice
	query := facades.Orm().Query().Where("plan_id = ? AND currency = ? AND billing_period = ? AND status = ?", planID, currency, period, "active")
	if err := query.OrderByDesc("effective_from").OrderByDesc("id").First(&price); err != nil {
		return nil, ErrPlanPriceNotFound
	}
	if err := ValidatePlanPrice(price); err != nil {
		return nil, err
	}
	return &price, nil
}
