package content

import (
	"reflect"
	"testing"
)

func TestSanitizeDocumentKeepsAllowedTiptapContent(t *testing.T) {
	linkMark := map[string]any{
		"type": "link",
		"attrs": map[string]any{
			"href":   "https://example.com",
			"target": "_blank",
			"rel":    "nofollow",
		},
	}
	input := map[string]any{
		"type": "doc",
		"content": []any{
			map[string]any{
				"type":    "heading",
				"attrs":   map[string]any{"level": float64(2)},
				"content": []any{map[string]any{"type": "text", "text": "About FastImg"}},
			},
			map[string]any{
				"type": "paragraph",
				"content": []any{map[string]any{
					"type":  "text",
					"text":  "Read more",
					"marks": []any{linkMark},
				}},
			},
		},
	}

	got, err := SanitizeDocument(input)
	if err != nil {
		t.Fatalf("sanitize allowed document: %v", err)
	}
	if !reflect.DeepEqual(got, input) {
		t.Fatalf("sanitized document = %#v, want %#v", got, input)
	}
}

func TestSanitizeDocumentRejectsDangerousNodesAndAttributes(t *testing.T) {
	cases := []map[string]any{
		{"type": "doc", "content": []any{map[string]any{"type": "script", "content": []any{map[string]any{"type": "text", "text": "alert(1)"}}}}},
		{"type": "doc", "content": []any{map[string]any{"type": "paragraph", "attrs": map[string]any{"onClick": "alert(1)"}}}},
		{"type": "doc", "content": []any{map[string]any{"type": "paragraph", "content": []any{map[string]any{"type": "text", "text": "x", "marks": []any{map[string]any{"type": "link", "attrs": map[string]any{"href": "javascript:alert(1)"}}}}}}}},
		{"type": "doc", "content": []any{map[string]any{"type": "image", "attrs": map[string]any{"src": "data:text/html,alert(1)"}}}},
	}

	for index, input := range cases {
		if _, err := SanitizeDocument(input); err == nil {
			t.Errorf("case %d was accepted", index)
		}
	}
}

func TestValidateURLRejectsPrivateTargetsAndDangerousProtocols(t *testing.T) {
	for _, raw := range []string{"javascript:alert(1)", "data:text/html,alert(1)", "http://127.0.0.1/admin", "http://localhost/admin", "http://10.0.0.1/", "http://[::1]/"} {
		if err := ValidateURL(raw, false); err == nil {
			t.Errorf("ValidateURL(%q) accepted a forbidden URL", raw)
		}
	}
	if err := ValidateURL("mailto:owner@example.com", true); err != nil {
		t.Fatalf("mailto URL should be accepted when enabled: %v", err)
	}
}
