package billing

import "strings"

func ClassifyReconciliation(providerAmount int64, providerCurrency string, internalAmount int64, internalCurrency string, internalExists bool) string {
	if !internalExists {
		return "missing_internal"
	}
	if providerAmount != internalAmount {
		return "amount_mismatch"
	}
	if !strings.EqualFold(providerCurrency, internalCurrency) {
		return "currency_mismatch"
	}
	return "matched"
}
