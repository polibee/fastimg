package resource

import "testing"

func TestManifestIsAdminOnlyAndDoesNotExposeTokenDigest(t *testing.T) {
	manifest := Manifest()
	if manifest.Name != "api_tokens" || manifest.Route != "/admin/api_tokens" || manifest.Table != "api_tokens" {
		t.Fatalf("unexpected API token manifest: %+v", manifest)
	}
	for _, field := range manifest.Fields {
		if field.Name == "token_hash" {
			t.Fatal("token digest must not be declared in the admin manifest")
		}
	}
	if len(manifest.Actions) != 3 || manifest.Actions[1].Kind != "api-token-status" || !manifest.Actions[1].Batch {
		t.Fatalf("expected status batch action and delete action, got %+v", manifest.Actions)
	}
	if manifest.Actions[2].Name != "delete" || manifest.Actions[2].Permission != "admin.api_tokens.delete" {
		t.Fatalf("expected admin delete action, got %+v", manifest.Actions[2])
	}
}
