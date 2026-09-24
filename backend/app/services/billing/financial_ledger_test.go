package billing

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFinancialEntryRequiresPositiveAmountAndStableSource(t *testing.T) {
	require.NoError(t, ValidateFinancialEntry(990, "USD", "payment", "credit", "payment_intent", "42", "ledger-42"))
	require.Error(t, ValidateFinancialEntry(0, "USD", "payment", "credit", "payment_intent", "42", "ledger-42"))
	require.Error(t, ValidateFinancialEntry(990, "usd", "payment", "credit", "payment_intent", "42", "ledger-42"))
	require.Error(t, ValidateFinancialEntry(990, "USD", "payment", "credit", "", "42", "ledger-42"))
	balances := ApplyFinancialDirection(0, "credit", 990)
	balances = ApplyFinancialDirection(balances, "debit", 300)
	require.Equal(t, int64(690), balances)
}
