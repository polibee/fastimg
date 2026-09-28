package media

import (
	"bytes"
	"testing"
)

func TestAnalyzeSecurityRejectsMismatchedMagic(t *testing.T) {
	result := AnalyzeSecurity("photo.png", "image/png", []byte("not an image"), 20_000_000)
	if result.Status != SecurityStatusHighRisk || result.Reason != SecurityReasonMagicMismatch {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestAnalyzeSecurityAllowsDecodedImage(t *testing.T) {
	png := []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0, 0, 0, 0, 0x49, 0x48, 0x44, 0x52,
		0, 0, 0, 1, 0, 0, 0, 1, 8, 6, 0, 0, 0,
		0x1f, 0x15, 0xc4, 0x89,
	}
	// The signature/header check is intentionally independent of full decode;
	// upload processing remains responsible for complete image validation.
	result := AnalyzeSecurity("photo.png", "image/png", bytes.Clone(png), 20_000_000)
	if result.Status != SecurityStatusClear {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestValidateVariantRequest(t *testing.T) {
	if _, err := ValidateVariantRequest("webp", 800, 70); err != ErrVariantFormatUnavailable {
		t.Fatalf("expected unavailable webp format, got %v", err)
	}
	request, err := ValidateVariantRequest("png", 1200, 80)
	if err != nil || request.Width != 1200 || request.Quality != 80 {
		t.Fatalf("unexpected request: %+v, %v", request, err)
	}
}
