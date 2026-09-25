package planservices

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/database/orm"

	"goravel/app/facades"
	"goravel/app/models"
	"goravel/app/services/quota"
)

var (
	ErrPlanNotFound            = errors.New("plan not found")
	ErrSubscriptionExists      = errors.New("active subscription already exists")
	ErrSubscriptionUnavailable = errors.New("active subscription is unavailable")
)

type PlanService struct{}

func NewPlanService() *PlanService { return &PlanService{} }

func (s *PlanService) ActivePlans() ([]models.Plan, error) {
	var plans []models.Plan
	if err := facades.Orm().Query().Where("status = ?", "active").OrderBy("sort_order").OrderBy("id").Get(&plans); err != nil {
		return nil, err
	}
	return plans, nil
}

func (s *PlanService) EnsureFreeSubscription(userID uint) error {
	return EnsureFreeSubscriptionWithQuery(facades.Orm().Query(), userID)
}

func EnsureFreeSubscriptionWithQuery(query orm.Query, userID uint) error {
	exists, err := query.Model(&models.Subscription{}).Where("user_id = ? AND status = ?", userID, "active").Exists()
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	var plan models.Plan
	exists, err = query.Model(&models.Plan{}).Where("code = ?", "free").Exists()
	if err != nil {
		return err
	}
	if !exists {
		plan, err = buildDefaultFreePlan()
		if err != nil {
			return err
		}
		if err := query.Create(&plan); err != nil {
			return err
		}
	}
	if err := query.Where("code = ? AND status = ?", "free", "active").First(&plan); err != nil {
		return ErrPlanNotFound
	}

	entitlement := quota.DefaultFreeEntitlement()
	if strings.TrimSpace(plan.EntitlementsJSON) != "" {
		parsed, parseErr := quota.ParseEntitlementJSON(plan.EntitlementsJSON)
		if parseErr != nil {
			return fmt.Errorf("%w: free plan entitlements are invalid: %v", ErrInvalidPlan, parseErr)
		}
		entitlement = parsed
	}
	encodedEntitlements, encodeErr := json.Marshal(entitlement)
	if encodeErr != nil {
		return encodeErr
	}
	plan.EntitlementsJSON = string(encodedEntitlements)
	encoded, encodeErr := BuildSubscriptionSnapshot(plan)
	if encodeErr != nil {
		return encodeErr
	}
	return query.Create(&models.Subscription{
		UserID: userID, PlanID: plan.ID, Status: "active",
		EntitlementSnapshotJSON: encoded,
	})
}

// buildDefaultFreePlan is the runtime safety net for installations where the
// plans table exists but the optional seed command has not run yet. It only
// creates a plan when no free-plan row exists; an explicitly disabled Free plan
// remains disabled and is not silently reactivated.
func buildDefaultFreePlan() (models.Plan, error) {
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

func (s *PlanService) SubscriptionForUser(userID uint) (*models.Subscription, error) {
	return EnsureActiveSubscriptionWithQuery(facades.Orm().Query(), userID)
}

// AssignPlan is the administrative/manual assignment path used during
// development and operations. It updates the active subscription snapshot so
// all domain services observe the same plan terms immediately.
func (s *PlanService) AssignPlan(userID, planID uint) (*models.Subscription, error) {
	if userID == 0 || planID == 0 {
		return nil, ErrInvalidSubscriptionOwner
	}
	var assigned *models.Subscription
	err := facades.Orm().Transaction(func(tx orm.Query) error {
		var user models.User
		if err := tx.Where("id = ?", userID).First(&user); err != nil {
			return ErrInvalidSubscriptionOwner
		}
		var plan models.Plan
		if err := tx.Where("id = ? AND status = ?", planID, "active").First(&plan); err != nil {
			return ErrPlanNotFound
		}
		entitlement := quota.DefaultFreeEntitlement()
		if strings.TrimSpace(plan.EntitlementsJSON) != "" {
			parsed, err := quota.ParseEntitlementJSON(plan.EntitlementsJSON)
			if err != nil {
				return fmt.Errorf("%w: invalid plan entitlements: %v", ErrInvalidPlan, err)
			}
			entitlement = parsed
		}
		encodedEntitlements, err := json.Marshal(entitlement)
		if err != nil {
			return err
		}
		plan.EntitlementsJSON = string(encodedEntitlements)
		snapshot, err := BuildSubscriptionSnapshot(plan)
		if err != nil {
			return err
		}
		var subscription models.Subscription
		exists, err := tx.Model(&models.Subscription{}).Where("user_id = ? AND status = ?", userID, "active").Exists()
		if err != nil {
			return err
		}
		now := time.Now().UTC()
		if exists {
			if err := tx.Where("user_id = ? AND status = ?", userID, "active").First(&subscription); err != nil {
				return err
			}
			if _, err := tx.Where("id = ?", subscription.ID).Update(map[string]any{
				"plan_id": plan.ID, "starts_at": now, "entitlement_snapshot_json": snapshot,
			}); err != nil {
				return err
			}
			subscription.PlanID = plan.ID
			subscription.StartsAt = &now
			subscription.EntitlementSnapshotJSON = snapshot
		} else {
			subscription = models.Subscription{UserID: userID, PlanID: plan.ID, Status: "active", StartsAt: &now, EntitlementSnapshotJSON: snapshot}
			if err := tx.Create(&subscription); err != nil {
				return err
			}
		}
		assigned = &subscription
		return nil
	})
	if err != nil {
		return nil, err
	}
	return assigned, nil
}
