package auditservices

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRequestBodyForAuditDoesNotParseMultipartPayload(t *testing.T) {
	read := false
	got := RequestBodyForAudit("multipart/form-data; boundary=upload-boundary", func() map[string]any {
		read = true
		return map[string]any{"file": "large upload"}
	})

	require.False(t, read)
	require.Equal(t, map[string]any{"omitted": "multipart"}, got)
}

func TestRequestBodyForAuditStillCapturesNonMultipartFields(t *testing.T) {
	want := map[string]any{"name": "sample"}
	got := RequestBodyForAudit("application/json", func() map[string]any { return want })

	require.Equal(t, want, got)
}

func TestRequestBodyForAuditSkipsMalformedMultipartContentType(t *testing.T) {
	read := false
	got := RequestBodyForAudit("multipart/form-data; boundary", func() map[string]any {
		read = true
		return nil
	})

	require.False(t, read)
	require.Equal(t, map[string]any{"omitted": "multipart"}, got)
}
