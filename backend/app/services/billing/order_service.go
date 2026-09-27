package billing

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/database/orm"

	"goravel/app/facades"
	"goravel/app/models"
)

var ErrInvalidOrderRequest = errors.New("invalid order request")

type CreateOrderRequest struct {
	PlanID         uint   `json:"plan_id"`
	PriceID        uint   `json:"price_id"`
	Currency       string `json:"currency"`
	BillingPeriod  string `json:"billing_period"`
	IdempotencyKey string `json:"-"`
}

var (
	ErrOrderNotFound        = errors.New("order not found")
	ErrOrderAlreadyCanceled = errors.New("order is already canceled")
	ErrFreePlanNotPayable   = errors.New("free plan does not require a payment order")
)

func ValidateCreateOrderRequest(input CreateOrderRequest) error {
	if input.PlanID == 0 || input.PriceID == 0 {
		return fmt.Errorf("%w: plan_id and price_id are required", ErrInvalidOrderRequest)
	}
	if len(input.Currency) != 3 || input.Currency != strings.ToUpper(input.Currency) {
		return fmt.Errorf("%w: currency must be three uppercase letters", ErrInvalidOrderRequest)
	}
	for _, char := range input.Currency {
		if char < 'A' || char > 'Z' {
			return fmt.Errorf("%w: currency must be three uppercase letters", ErrInvalidOrderRequest)
		}
	}
	if input.BillingPeriod != "monthly" && input.BillingPeriod != "yearly" {
		return fmt.Errorf("%w: billing period is unsupported", ErrInvalidOrderRequest)
	}
	if strings.TrimSpace(input.IdempotencyKey) == "" || len(input.IdempotencyKey) > 120 {
		return fmt.Errorf("%w: idempotency key is required", ErrInvalidOrderRequest)
	}
	return nil
}

type OrderService struct{}

func NewOrderService() *OrderService { return &OrderService{} }

func CanCancelOrder(status string) bool {
	return status == "created" || status == "pending_payment"
}

func (s *OrderService) CreatePlanOrder(_ context.Context, userID uint, input CreateOrderRequest) (*models.Order, error) {
	if userID == 0 {
		return nil, ErrInvalidOrderRequest
	}
	if err := ValidateCreateOrderRequest(input); err != nil {
		return nil, err
	}
	var order models.Order
	err := facades.Orm().Transaction(func(tx orm.Query) error {
		var existing models.Order
		exists, err := tx.Model(&models.Order{}).Where("user_id = ? AND idempotency_key = ?", userID, strings.TrimSpace(input.IdempotencyKey)).Exists()
		if err != nil {
			return err
		}
		if exists {
			if err := tx.Where("user_id = ? AND idempotency_key = ?", userID, strings.TrimSpace(input.IdempotencyKey)).First(&existing); err != nil {
				return err
			}
			order = existing
			return nil
		}

		var plan models.Plan
		if err := tx.Where("id = ? AND status = ?", input.PlanID, "active").First(&plan); err != nil {
			return ErrOrderNotFound
		}
		if plan.Code == "free" {
			return ErrFreePlanNotPayable
		}
		var price models.PlanPrice
		if err := tx.Where("id = ? AND plan_id = ? AND currency = ? AND billing_period = ? AND status = ?", input.PriceID, input.PlanID, input.Currency, input.BillingPeriod, "active").First(&price); err != nil {
			return ErrPlanPriceNotFound
		}
		snapshot, err := BuildPlanPriceSnapshot(plan, price)
		if err != nil {
			return err
		}
		priceSnapshot, err := json.Marshal(snapshot)
		if err != nil {
			return err
		}
		customerSnapshot, err := json.Marshal(map[string]any{"user_id": userID})
		if err != nil {
			return err
		}
		expiresAt := time.Now().UTC().Add(30 * time.Minute)
		order = models.Order{
			PublicOrderNo: generatePublicOrderNo(), UserID: userID, Status: "pending_payment", Currency: snapshot.Currency,
			SubtotalAmountMinor: snapshot.AmountMinor, TotalAmountMinor: snapshot.AmountMinor,
			PriceSnapshot: string(priceSnapshot), CustomerSnapshot: string(customerSnapshot),
			IdempotencyKey: strings.TrimSpace(input.IdempotencyKey), ExpiresAt: &expiresAt,
		}
		if err := tx.Create(&order); err != nil {
			return err
		}
		entitlementSnapshot, err := json.Marshal(snapshot.Entitlements)
		if err != nil {
			return err
		}
		item := models.OrderItem{OrderID: order.ID, ProductType: "plan", ProductID: plan.ID, ProductVersion: price.Version, Quantity: 1, UnitAmountMinor: snapshot.AmountMinor, TotalAmountMinor: snapshot.AmountMinor, EntitlementSnapshot: string(entitlementSnapshot)}
		return tx.Create(&item)
	})
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (s *OrderService) ListOwnOrders(_ context.Context, userID uint, page, perPage int) ([]models.Order, int64, error) {
	if userID == 0 {
		return nil, 0, ErrInvalidOrderRequest
	}
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	var orders []models.Order
	var total int64
	query := facades.Orm().Query().Where("user_id = ?", userID)
	if err := query.OrderByDesc("id").Paginate(page, perPage, &orders, &total); err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}

func (s *OrderService) GetOwnOrder(_ context.Context, userID, orderID uint) (*models.Order, error) {
	if userID == 0 || orderID == 0 {
		return nil, ErrOrderNotFound
	}
	var order models.Order
	if err := facades.Orm().Query().Where("id = ? AND user_id = ?", orderID, userID).First(&order); err != nil {
		return nil, ErrOrderNotFound
	}
	return &order, nil
}

func (s *OrderService) CancelOwnOrder(_ context.Context, userID, orderID uint) error {
	order, err := s.GetOwnOrder(context.Background(), userID, orderID)
	if err != nil {
		return err
	}
	if !CanCancelOrder(order.Status) {
		return ErrOrderAlreadyCanceled
	}
	now := time.Now().UTC()
	result, err := facades.Orm().Query().Where("id = ? AND user_id = ? AND (status = ? OR status = ?)", orderID, userID, "created", "pending_payment").Update(map[string]any{"status": "canceled", "canceled_at": now})
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return ErrOrderAlreadyCanceled
	}
	return nil
}

func generatePublicOrderNo() string {
	b := make([]byte, 5)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("FST-%d", time.Now().UTC().UnixNano())
	}
	return "FST-" + time.Now().UTC().Format("20060102-150405") + "-" + hex.EncodeToString(b)
}
