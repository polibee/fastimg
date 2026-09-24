package billing

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReconciliationClassifiesProviderDifferences(t *testing.T) {
	require.Equal(t, "matched", ClassifyReconciliation(990, "USD", 990, "USD", true))
	require.Equal(t, "amount_mismatch", ClassifyReconciliation(989, "USD", 990, "USD", true))
	require.Equal(t, "currency_mismatch", ClassifyReconciliation(990, "EUR", 990, "USD", true))
	require.Equal(t, "missing_internal", ClassifyReconciliation(990, "USD", 990, "USD", false))
}
