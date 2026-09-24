package billing

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/goravel/framework/contracts/database/orm"

	"goravel/app/facades"
	"goravel/app/models"
	planservices "goravel/app/services/plans"
)

var ErrFulfillmentNotReady = errors.New("fulfillment is not ready")

func CanFulfillPayment(status string) bool { return status == "succeeded" }

func NextFulfillmentStatus(orderStatus string, delivered bool) string {
	if orderStatus != "paid" && orderStatus != "fulfillment_pending" {
		return orderStatus
	}
	if delivered {
		return "fulfilled"
	}
	return "fulfillment_pending"
}

type FulfillmentService struct{}

func NewFulfillmentService() *FulfillmentService { return &FulfillmentService{} }

func (s *FulfillmentService) EnqueuePaidOrder(orderID uint) error {
	if orderID == 0 {
		return ErrFulfillmentNotReady
	}
	exists, err := facades.Orm().Query().Model(&models.FulfillmentTask{}).Where("order_id = ?", orderID).Exists()
	if err != nil || exists {
		return err
	}
	return facades.Orm().Query().Create(&models.FulfillmentTask{OrderID: orderID, Status: "pending", AvailableAt: timePtr(time.Now().UTC())})
}

func (s *FulfillmentService) Process(orderID uint) error {
	var task models.FulfillmentTask
	if err := facades.Orm().Query().Where("order_id = ?", orderID).First(&task); err != nil {
		return ErrFulfillmentNotReady
	}
	var order models.Order
	if err := facades.Orm().Query().Where("id = ?", orderID).First(&order); err != nil {
		return ErrOrderNotFound
	}
	if order.Status == "fulfilled" || task.Status == "completed" {
		return nil
	}
	if order.Status != "paid" && order.Status != "fulfillment_pending" {
		return ErrFulfillmentNotReady
	}
	err := facades.Orm().Transaction(func(tx orm.Query) error {
		var item models.OrderItem
		if err := tx.Where("order_id = ?", order.ID).First(&item); err != nil {
			return err
		}
		var snapshot PlanPriceSnapshot
		if err := json.Unmarshal([]byte(order.PriceSnapshot), &snapshot); err != nil {
			return err
		}
		entitlements, err := json.Marshal(snapshot.Entitlements)
		if err != nil {
			return err
		}
		plan := models.Plan{Code: snapshot.PlanCode, Name: snapshot.PlanName, Description: snapshot.PlanDescription, PriceAmount: snapshot.AmountMinor, Currency: snapshot.Currency, BillingPeriod: snapshot.BillingPeriod, EntitlementsJSON: string(entitlements), Status: "active"}
		plan.ID = snapshot.PlanID
		subscriptionSnapshot, err := planservices.BuildSubscriptionSnapshot(plan)
		if err != nil {
			return err
		}
		var subscription models.Subscription
		if exists, findErr := tx.Model(&models.Subscription{}).Where("user_id = ? AND status = ?", order.UserID, "active").Exists(); findErr != nil {
			return findErr
		} else if exists {
			if err := tx.Where("user_id = ? AND status = ?", order.UserID, "active").First(&subscription); err != nil {
				return err
			}
			subscription.PlanID, subscription.Status, subscription.StartsAt, subscription.EntitlementSnapshotJSON = snapshot.PlanID, "active", timePtr(time.Now().UTC()), subscriptionSnapshot
			if _, err := tx.Where("id = ?", subscription.ID).Update(map[string]any{"plan_id": subscription.PlanID, "status": subscription.Status, "starts_at": subscription.StartsAt, "entitlement_snapshot_json": subscription.EntitlementSnapshotJSON}); err != nil {
				return err
			}
		} else {
			subscription = models.Subscription{UserID: order.UserID, PlanID: snapshot.PlanID, Status: "active", StartsAt: timePtr(time.Now().UTC()), EntitlementSnapshotJSON: subscriptionSnapshot}
			if err := tx.Create(&subscription); err != nil {
				return err
			}
		}
		now := time.Now().UTC()
		if _, err := tx.Where("id = ?", order.ID).Update(map[string]any{"status": "fulfilled", "fulfilled_at": now}); err != nil {
			return err
		}
		_, err = tx.Where("id = ?", task.ID).Update(map[string]any{"status": "completed", "completed_at": now, "last_error": nil})
		return err
	})
	if err != nil {
		_, _ = facades.Orm().Query().Where("id = ?", task.ID).Update(map[string]any{"status": "pending", "attempts": task.Attempts + 1, "last_error": err.Error()})
		return fmt.Errorf("fulfillment failed: %w", err)
	}
	return nil
}

func timePtr(value time.Time) *time.Time { return &value }
