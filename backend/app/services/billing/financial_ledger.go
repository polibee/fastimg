package billing

import (
	"errors"
	"fmt"
	"strings"

	"goravel/app/facades"
	"goravel/app/models"
)

var ErrInvalidFinancialEntry = errors.New("invalid financial entry")

func ValidateFinancialEntry(amount int64, currency, kind, direction, sourceType, sourceID, idempotencyKey string) error {
	if amount <= 0 || len(currency) != 3 || currency != strings.ToUpper(currency) || kind == "" || (direction != "credit" && direction != "debit") || sourceType == "" || sourceID == "" || idempotencyKey == "" {
		return ErrInvalidFinancialEntry
	}
	return nil
}

func ApplyFinancialDirection(balance int64, direction string, amount int64) int64 {
	if direction == "debit" {
		return balance - amount
	}
	return balance + amount
}

type FinancialLedger struct{}

func NewFinancialLedger() *FinancialLedger { return &FinancialLedger{} }

func (l *FinancialLedger) Append(entry models.FinancialTransaction) error {
	if err := ValidateFinancialEntry(entry.AmountMinor, entry.Currency, entry.Type, entry.Direction, entry.SourceType, entry.SourceID, entry.IdempotencyKey); err != nil {
		return err
	}
	if exists, err := facades.Orm().Query().Model(&models.FinancialTransaction{}).Where("idempotency_key = ?", entry.IdempotencyKey).Exists(); err != nil {
		return err
	} else if exists {
		return nil
	}
	if err := facades.Orm().Query().Create(&entry); err != nil {
		return fmt.Errorf("append financial transaction: %w", err)
	}
	return nil
}
