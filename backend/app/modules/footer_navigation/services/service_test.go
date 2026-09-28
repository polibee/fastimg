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
