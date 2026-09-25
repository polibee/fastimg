package planservices

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"

	"goravel/app/services/quota"
)

var ErrInvalidPlan = errors.New("invalid plan")

// PreparePlanWrite validates plan-domain rules and canonicalizes persisted values.
// Generic required field and type checks remain owned by the Resource Engine.
func PreparePlanWrite(operation string, payload map[string]any) error {
	if operation != "create" && operation != "update" {
		return fmt.Errorf("%w: unsupported write operation", ErrInvalidPlan)
	}
	if value, exists := payload["price_amount"]; exists {
		amount, err := parsePlanPrice(value)
		if err != nil {
			return err
		}
		payload["price_amount"] = amount
	}
	if value, exists := payload["currency"]; exists {
		currency, ok := value.(string)
		if !ok || len(currency) != 3 || currency != upperASCII(currency) {
			return fmt.Errorf("%w: currency must be three uppercase letters", ErrInvalidPlan)
		}
		for _, char := range currency {
			if char < 'A' || char > 'Z' {
				return fmt.Errorf("%w: currency must be three uppercase letters", ErrInvalidPlan)
			}
		}
	}
	if value, exists := payload["billing_period"]; exists {
		period, ok := value.(string)
		if !ok || (period != "monthly" && period != "yearly") {
			return fmt.Errorf("%w: billing_period is unsupported", ErrInvalidPlan)
		}
	}
	if value, exists := payload["status"]; exists {
		status, ok := value.(string)
		if !ok || (status != "active" && status != "disabled") {
			return fmt.Errorf("%w: status is unsupported", ErrInvalidPlan)
		}
	}
	if value, exists := payload["entitlements_json"]; exists {
		encoded, ok := value.(string)
		if !ok {
			return fmt.Errorf("%w: entitlements_json must be a JSON string", ErrInvalidPlan)
		}
		entitlement, err := quota.ParseEntitlementJSON(encoded)
		if err != nil {
			return fmt.Errorf("%w: entitlements_json: %v", ErrInvalidPlan, err)
		}
		canonical, err := json.Marshal(entitlement)
		if err != nil {
			return fmt.Errorf("%w: encode entitlements: %v", ErrInvalidPlan, err)
		}
		payload["entitlements_json"] = string(canonical)
	}
	return nil
}

func parsePlanPrice(value any) (int64, error) {
	var amount int64
	switch number := value.(type) {
	case int:
		amount = int64(number)
	case int32:
		amount = int64(number)
	case int64:
		amount = number
	case uint:
		if uint64(number) > math.MaxInt64 {
			return 0, fmt.Errorf("%w: price_amount exceeds the supported range", ErrInvalidPlan)
		}
		amount = int64(number)
	case uint64:
		if number > math.MaxInt64 {
			return 0, fmt.Errorf("%w: price_amount exceeds the supported range", ErrInvalidPlan)
		}
		amount = int64(number)
	case float64:
		if math.IsNaN(number) || math.IsInf(number, 0) || math.Trunc(number) != number || number < 0 || number >= math.Exp2(63) {
			return 0, fmt.Errorf("%w: price_amount must be a non-negative integer", ErrInvalidPlan)
		}
		amount = int64(number)
	case json.Number:
		parsed, err := number.Int64()
		if err != nil {
			return 0, fmt.Errorf("%w: price_amount must be a non-negative integer", ErrInvalidPlan)
		}
		amount = parsed
	default:
		return 0, fmt.Errorf("%w: price_amount must be a non-negative integer", ErrInvalidPlan)
	}
	if amount < 0 {
		return 0, fmt.Errorf("%w: price_amount cannot be negative", ErrInvalidPlan)
	}
	return amount, nil
}

func upperASCII(value string) string {
	bytes := []byte(value)
	for index, char := range bytes {
		if char >= 'a' && char <= 'z' {
			bytes[index] = char - ('a' - 'A')
		}
	}
	return string(bytes)
}
