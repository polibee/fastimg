package planservices

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/goravel/framework/contracts/database/orm"

	"goravel/app/facades"
	"goravel/app/models"
	"goravel/app/services/quota"
)

var (
	ErrPlanNotFound       = errors.New("plan not found")
	ErrSubscriptionExists = errors.New("active subscription already exists")
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
	var existing models.Subscription
	if err := query.Where("user_id = ? AND status = ?", userID, "active").First(&existing); err == nil {
		return nil
	}

	var plan models.Plan
	if err := query.Where("code = ? AND status = ?", "free", "active").First(&plan); err != nil {
		return ErrPlanNotFound
	}

	snapshot := plan.EntitlementsJSON
	if strings.TrimSpace(snapshot) == "" {
		encoded, encodeErr := json.Marshal(quota.DefaultFreeEntitlement())
		if encodeErr != nil {
			return encodeErr
		}
		snapshot = string(encoded)
	}
	return query.Create(&models.Subscription{
		UserID: userID, PlanID: plan.ID, Status: "active",
		EntitlementSnapshotJSON: snapshot,
	})
}

func (s *PlanService) SubscriptionForUser(userID uint) (*models.Subscription, error) {
	var subscription models.Subscription
	if err := facades.Orm().Query().Where("user_id = ?", userID).Where("status = ?", "active").First(&subscription); err != nil {
		if provisionErr := s.EnsureFreeSubscription(userID); provisionErr != nil {
			return nil, err
		}
		if retryErr := facades.Orm().Query().Where("user_id = ?", userID).Where("status = ?", "active").First(&subscription); retryErr != nil {
			return nil, retryErr
		}
	}
	return &subscription, nil
}
