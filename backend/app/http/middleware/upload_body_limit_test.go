package middleware

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLimitUploadBodyRejectsMoreThanConfiguredBytes(t *testing.T) {
	body := limitUploadBody(io.NopCloser(strings.NewReader("12345")), 4)
	data, err := io.ReadAll(body)

	var tooLarge *http.MaxBytesError
	require.True(t, errors.As(err, &tooLarge))
	require.Len(t, data, 4)
}
