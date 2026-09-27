package auditservices

import (
	"strings"
	"testing"
)

func TestShouldAuditHTTP(t *testing.T) {
	for _, path := range []string{"/api/v1/auth/login", "/api/v1/admin/users"} {
		if !ShouldAuditHTTP("POST", path) {
			t.Fatalf("expected API path to be audited: %s", path)
		}
	}
	for _, path := range []string{"/api/v1/admin/audit-logs", "/api/docs", "/"} {
		if ShouldAuditHTTP("GET", path) {
			t.Fatalf("did not expect path to be audited: %s", path)
		}
	}
	if ShouldAuditHTTP("OPTIONS", "/api/v1/admin/users") {
		t.Fatal("OPTIONS must not be audited")
	}
}

func TestShouldRecordHTTPSkipsSuccessfulReadNoiseButKeepsFailures(t *testing.T) {
	if ShouldRecordHTTP("GET", "/api/v1/ads?placement=header", 200) {
		t.Fatal("successful ad reads should not create audit noise")
	}
	if ShouldRecordHTTP("GET", "/api/v1/media/1/content", 200) {
		t.Fatal("successful media delivery should use media access logs instead")
	}
	if !ShouldRecordHTTP("GET", "/api/v1/ads?placement=header", 500) {
		t.Fatal("failed reads must remain auditable")
	}
	if !ShouldRecordHTTP("POST", "/api/v1/orders/1/payments", 201) {
		t.Fatal("payment mutations must remain auditable")
	}
	if ShouldRecordHTTP("POST", "/api/v1/auth/refresh", 200) {
		t.Fatal("successful refresh rotations should not flood the audit trail")
	}
}

func TestClassifyHTTPActionUsesFeatureNames(t *testing.T) {
	tests := []struct {
		method string
		path   string
		want   string
	}{
		{"POST", "/api/v1/orders/1/payments", "billing.payment.create"},
		{"POST", "/api/v1/uploads", "media.upload"},
		{"PUT", "/api/v1/admin/settings/payment.nowpayments.api_key", "settings.update"},
		{"POST", "/api/v1/payment-gateways/xcash/webhook", "billing.webhook.receive"},
		{"POST", "/api/v1/media/22/reports", "moderation.report.create"},
	}
	for _, test := range tests {
		if got := ClassifyHTTPAction(test.method, test.path); got != test.want {
			t.Fatalf("ClassifyHTTPAction(%q, %q) = %q, want %q", test.method, test.path, got, test.want)
		}
	}
}

func TestBuildHTTPAuditMetadataContainsRequestAndResponseSummary(t *testing.T) {
	metadata := BuildHTTPAuditMetadata(HTTPAuditInput{
		Method:       "POST",
		Path:         "/api/v1/admin/users",
		Query:        map[string]string{"page": "1"},
		RequestBody:  map[string]any{"email": "admin@example.com", "password": "secret"},
		Status:       201,
		ContentType:  "application/json",
		ResponseBody: []byte(`{"data":{"id":1,"access_token":"secret-token"}}`),
	})

	request, ok := metadata["request"].(map[string]any)
	if !ok || request["method"] != "POST" || request["path"] != "/api/v1/admin/users" {
		t.Fatalf("unexpected request summary: %#v", metadata["request"])
	}
	response, ok := metadata["response"].(map[string]any)
	if !ok || response["status"] != 201 {
		t.Fatalf("unexpected response summary: %#v", metadata["response"])
	}
	encoded, _, err := MarshalBounded(metadata)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "secret-token") || strings.Contains(string(encoded), "secret") {
		t.Fatalf("HTTP audit metadata leaked sensitive data: %s", encoded)
	}
}

func TestBuildHTTPAuditMetadataOmitsNonJSONResponseBody(t *testing.T) {
	metadata := BuildHTTPAuditMetadata(HTTPAuditInput{
		Method: "GET", Path: "/api/v1/admin/users/export", Status: 200,
		ContentType: "text/csv", ResponseBody: []byte("email,password\na@example.com,secret"),
	})
	response := metadata["response"].(map[string]any)
	if _, exists := response["body"]; exists {
		t.Fatalf("non-JSON response body should not be retained: %#v", response)
	}
	if response["body_bytes"] != 35 {
		t.Fatalf("unexpected non-JSON body size: %#v", response["body_bytes"])
	}
}
