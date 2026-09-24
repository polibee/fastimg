package billing

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRefundableAmountNeverExceedsCapturedAmount(t *testing.T) {
	require.Equal(t, int64(700), RefundableAmount(1000, 300))
	require.Equal(t, int64(0), RefundableAmount(1000, 1000))
	require.Equal(t, int64(0), RefundableAmount(1000, 1200))
	require.Error(t, ValidateRefundAmount(701, 700))
	require.NoError(t, ValidateRefundAmount(700, 700))
}
