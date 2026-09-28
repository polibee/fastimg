package services

import "testing"

func TestNavigationTargetTypes(t *testing.T) {
	for _, value := range []string{"page", "route", "external", "friends"} {
		if !validTargetType(value) {
			t.Fatalf("target type %q should be accepted", value)
		}
	}
	if validTargetType("javascript") {
		t.Fatal("unknown target type should be rejected")
	}
}

func TestNavigationTargetValidation(t *testing.T) {
	valid := map[string]string{
		"page":     "privacy",
		"route":    "/friends",
		"external": "https://example.com/docs",
		"friends":  "friends",
	}
	for targetType, targetValue := range valid {
		if err := validateTarget(targetType, targetValue); err != nil {
			t.Errorf("validateTarget(%q, %q) returned error: %v", targetType, targetValue, err)
		}
	}
	for _, target := range [][2]string{
		{"page", "/privacy"},
		{"route", "//evil.example"},
		{"external", "javascript:alert(1)"},
		{"friends", "https://example.com"},
	} {
		if err := validateTarget(target[0], target[1]); err == nil {
			t.Errorf("validateTarget(%q, %q) accepted an invalid target", target[0], target[1])
		}
	}
}
