package billing

import "testing"

func TestFakeGatewayDefaultDisabledInProduction(t *testing.T) {
	if fakeGatewayDefaultEnabled("production") {
		t.Fatal("fake gateway must be disabled by default in production")
	}
	if !fakeGatewayDefaultEnabled("development") {
		t.Fatal("fake gateway should remain available by default in development")
	}
}
