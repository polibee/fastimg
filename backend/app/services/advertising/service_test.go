package advertising

import "testing"

func TestPrepareAdvertisingWriteValidatesAdminCreative(t *testing.T) {
	valid := map[string]any{
		"placement":        "header",
		"creative_type":    "script",
		"creative_content": " console.log('ad') ",
		"target_url":       "https://example.com/campaign",
	}
	if err := PrepareAdvertisingWrite("create", valid); err != nil {
		t.Fatalf("valid admin creative rejected: %v", err)
	}
	if valid["creative_content"] != "console.log('ad')" {
		t.Fatalf("creative content was not normalized: %#v", valid["creative_content"])
	}

	for _, creativeType := range []string{"text", "image", "script"} {
		payload := map[string]any{
			"placement":        "footer",
			"creative_type":    creativeType,
			"creative_content": "value",
		}
		if err := PrepareAdvertisingWrite("create", payload); err != nil {
			t.Fatalf("creative type %q rejected: %v", creativeType, err)
		}
	}
}

func TestPrepareAdvertisingWriteRejectsUnsafeOrIncompleteCreative(t *testing.T) {
	cases := []struct {
		name    string
		payload map[string]any
	}{
		{name: "invalid type", payload: map[string]any{"placement": "header", "creative_type": "html", "creative_content": "x"}},
		{name: "empty content", payload: map[string]any{"placement": "header", "creative_type": "text", "creative_content": "  "}},
		{name: "unsafe target", payload: map[string]any{"placement": "header", "creative_type": "text", "creative_content": "x", "target_url": "javascript:alert(1)"}},
		{name: "invalid placement", payload: map[string]any{"placement": "popup", "creative_type": "text", "creative_content": "x"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := PrepareAdvertisingWrite("create", tc.payload); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
