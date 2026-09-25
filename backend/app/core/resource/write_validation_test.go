package resource

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestPrepareWriteUsesOptionalManifestPreparer(t *testing.T) {
	called := false
	manifest := Manifest{WritePreparer: func(operation string, payload map[string]any) error {
		called = true
		if operation != "update" || payload["status"] != "active" {
			t.Fatalf("validator received operation=%q payload=%v", operation, payload)
		}
		payload["status"] = "disabled"
		return nil
	}}
	input := map[string]any{"status": "active"}
	prepared, err := PrepareWrite(manifest, "update", input)
	if err != nil {
		t.Fatalf("valid write rejected: %v", err)
	}
	if !called {
		t.Fatal("manifest validator was not called")
	}
	if prepared["status"] != "disabled" || input["status"] != "active" {
		t.Fatalf("prepared payload = %v, original payload = %v", prepared, input)
	}
}

func TestPrepareWriteWrapsDomainRejection(t *testing.T) {
	manifest := Manifest{WritePreparer: func(string, map[string]any) error {
		return errors.New("invalid plan price")
	}}
	if _, err := PrepareWrite(manifest, "create", map[string]any{"price_amount": -1}); !errors.Is(err, ErrWriteValidation) {
		t.Fatalf("error = %v, want ErrWriteValidation", err)
	}
}

func TestPrepareWriteDoesNotRequireHookForExistingResources(t *testing.T) {
	if _, err := PrepareWrite(Manifest{}, "create", map[string]any{"name": "post"}); err != nil {
		t.Fatalf("resource without opt-in validator rejected: %v", err)
	}
}

func TestPrepareWriteCopiesNestedJSONValuesBeforeHookMutation(t *testing.T) {
	input := map[string]any{
		"metadata": map[string]any{"visibility": "private", "labels": []any{"original"}},
	}
	manifest := Manifest{WritePreparer: func(_ string, payload map[string]any) error {
		metadata := payload["metadata"].(map[string]any)
		metadata["visibility"] = "public"
		metadata["labels"].([]any)[0] = "normalized"
		return nil
	}}
	prepared, err := PrepareWrite(manifest, "update", input)
	if err != nil {
		t.Fatalf("prepare nested payload: %v", err)
	}
	if input["metadata"].(map[string]any)["visibility"] != "private" || input["metadata"].(map[string]any)["labels"].([]any)[0] != "original" {
		t.Fatalf("write hook mutated caller-owned input: %v", input)
	}
	if prepared["metadata"].(map[string]any)["visibility"] != "public" || prepared["metadata"].(map[string]any)["labels"].([]any)[0] != "normalized" {
		t.Fatalf("normalized payload = %v", prepared)
	}
}

func TestManifestJSONDoesNotExposeWritePreparer(t *testing.T) {
	manifest := Manifest{WritePreparer: func(string, map[string]any) error { return nil }}
	encoded, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("marshal resource manifest: %v", err)
	}
	if strings.Contains(string(encoded), "WriteValidator") {
		t.Fatalf("internal write hook leaked into manifest JSON: %s", encoded)
	}
}
