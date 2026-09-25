package resource

import (
	"errors"
	"testing"

	coreresource "goravel/app/core/resource"
)

func TestPlansManifestRejectsInvalidPlanWrites(t *testing.T) {
	manifest := Manifest()
	if manifest.WritePreparer == nil {
		t.Fatal("plans manifest must enforce domain write preparation")
	}
	_, err := coreresource.PrepareWrite(manifest, "create", map[string]any{
		"price_amount": float64(-1),
	})
	if !errors.Is(err, coreresource.ErrWriteValidation) {
		t.Fatalf("error = %v, want ErrWriteValidation", err)
	}
}
