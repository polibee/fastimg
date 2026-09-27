package media

import "testing"

func TestManifestUsesDomainActionsForMediaStateChanges(t *testing.T) {
	manifest := Manifest()
	if manifest.Fields[8].Writable || manifest.Fields[9].Writable || manifest.Fields[10].Writable {
		t.Fatal("media lifecycle fields must not be writable through generic CRUD")
	}
	if len(manifest.Actions) < 3 {
		t.Fatalf("expected dedicated media actions, got %d", len(manifest.Actions))
	}
}
