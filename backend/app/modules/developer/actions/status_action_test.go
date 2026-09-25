package actions

import "testing"

func TestParseStatusParamsOnlyAllowsTerminalAdminStates(t *testing.T) {
	for _, status := range []string{"disabled", "revoked"} {
		got, err := ParseStatusParams(map[string]any{"status": status})
		if err != nil || got != status {
			t.Fatalf("status %q parsed as %q with error %v", status, got, err)
		}
	}
	for _, payload := range []map[string]any{
		{"status": "active"},
		{"status": "disabled", "extra": true},
		{"status": 1},
	} {
		if _, err := ParseStatusParams(payload); err == nil {
			t.Fatalf("invalid payload accepted: %#v", payload)
		}
	}
}
