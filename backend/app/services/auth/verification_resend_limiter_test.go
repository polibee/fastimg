package authservices

import (
	"testing"
	"time"
)

func TestVerificationResendPolicyNormalizesUnsafeSettings(t *testing.T) {
	policy := NormalizeVerificationResendPolicy(VerificationResendPolicy{
		Enabled:              true,
		EmailCooldownSeconds: 0,
		IPCooldownSeconds:    -1,
		DailyEmailLimit:      9999,
		DailyIPLimit:         0,
	})

	if policy.EmailCooldownSeconds != defaultVerificationResendEmailCooldownSeconds {
		t.Fatalf("email cooldown = %d, want default %d", policy.EmailCooldownSeconds, defaultVerificationResendEmailCooldownSeconds)
	}
	if policy.IPCooldownSeconds != defaultVerificationResendIPCooldownSeconds {
		t.Fatalf("ip cooldown = %d, want default %d", policy.IPCooldownSeconds, defaultVerificationResendIPCooldownSeconds)
	}
	if policy.DailyEmailLimit != maxVerificationResendDailyEmailLimit {
		t.Fatalf("daily email limit = %d, want max %d", policy.DailyEmailLimit, maxVerificationResendDailyEmailLimit)
	}
	if policy.DailyIPLimit != defaultVerificationResendDailyIPLimit {
		t.Fatalf("daily ip limit = %d, want default %d", policy.DailyIPLimit, defaultVerificationResendDailyIPLimit)
	}
}

func TestVerificationResendDecisionEnforcesEmailAndIPWindows(t *testing.T) {
	now := time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC)
	policy := VerificationResendPolicy{
		Enabled:              true,
		EmailCooldownSeconds: 60,
		IPCooldownSeconds:    10,
		DailyEmailLimit:      5,
		DailyIPLimit:         20,
	}

	decision := verificationResendDecision(policy, verificationResendState{}, now)
	if !decision.Allowed {
		t.Fatal("first resend should be allowed")
	}

	decision = verificationResendDecision(policy, verificationResendState{
		EmailCooldownUntil: now.Add(30 * time.Second),
	}, now)
	if decision.Allowed || decision.Reason != verificationResendLimitEmailCooldown {
		t.Fatalf("email cooldown decision = %+v", decision)
	}

	decision = verificationResendDecision(policy, verificationResendState{
		IPCooldownUntil: now.Add(5 * time.Second),
	}, now)
	if decision.Allowed || decision.Reason != verificationResendLimitIPCooldown {
		t.Fatalf("ip cooldown decision = %+v", decision)
	}

	decision = verificationResendDecision(policy, verificationResendState{DailyEmailCount: 5}, now)
	if decision.Allowed || decision.Reason != verificationResendLimitDailyEmail {
		t.Fatalf("daily email decision = %+v", decision)
	}

	decision = verificationResendDecision(policy, verificationResendState{DailyIPCount: 20}, now)
	if decision.Allowed || decision.Reason != verificationResendLimitDailyIP {
		t.Fatalf("daily ip decision = %+v", decision)
	}
}
