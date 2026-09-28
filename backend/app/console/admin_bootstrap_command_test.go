package console

import "testing"

func TestAdminBootstrapCommandMetadata(t *testing.T) {
	command := AdminBootstrapCommand{}
	if command.Signature() != "admin:bootstrap" {
		t.Fatalf("signature = %q", command.Signature())
	}
	if command.Description() == "" {
		t.Fatal("description must not be empty")
	}
}

func TestBootstrapFastImgPermissionsIncludeOperationalAdminFeatures(t *testing.T) {
	seen := make(map[string]bool)
	for _, permission := range bootstrapFastImgPermissions() {
		seen[permission.Name] = true
	}
	for _, name := range []string{"admin.api_tokens.view", "admin.api_tokens.delete", "admin.reports.view", "admin.reports.update", "admin.storage.view"} {
		if !seen[name] {
			t.Fatalf("bootstrap permissions missing %q", name)
		}
	}
}
