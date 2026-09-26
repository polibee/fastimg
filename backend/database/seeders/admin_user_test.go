package seeders

import "testing"

func TestFastImgAdminPermissionsIncludeBillingAndMediaManagement(t *testing.T) {
	permissions := fastImgAdminPermissions()
	seen := make(map[string]bool, len(permissions))
	for _, permission := range permissions {
		seen[permission.Name] = true
	}
	for _, required := range []string{
		"admin.orders.view",
		"admin.payment_transactions.view",
		"admin.payment_events.view",
		"admin.refunds.view",
		"admin.billing.fulfill",
		"admin.media.view",
		"admin.media.update",
		"admin.media.delete",
		"admin.api_tokens.delete",
		"admin.backups.manage",
		"admin.backups.download",
	} {
		if !seen[required] {
			t.Fatalf("missing admin permission %q", required)
		}
	}
}
