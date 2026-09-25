package planservices

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"goravel/app/models"
	"goravel/app/services/quota"
)

const subscriptionSnapshotSchemaVersion = 1

// PlanSnapshot contains the commercial terms and entitlements captured when a
// subscription is created. It is stored with the subscription rather than
// resolved from the mutable current plan on each read.
type PlanSnapshot struct {
	ID            uint   `json:"id"`
	Code          string `json:"code"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	PriceAmount   int64  `json:"price_amount"`
	Currency      string `json:"currency"`
	BillingPeriod string `json:"billing_period"`
}

type SubscriptionSnapshot struct {
	SchemaVersion int               `json:"schema_version"`
	PlanVersion   string            `json:"plan_version"`
	Plan          PlanSnapshot      `json:"plan"`
	Entitlements  quota.Entitlement `json:"entitlements"`
	HasPlan       bool              `json:"-"`
}

// BuildSubscriptionSnapshot creates a content-addressed plan version. Equal
// terms produce the same version; any price, display, period, or entitlement
// change produces a different version without a separate mutable version row.
func BuildSubscriptionSnapshot(plan models.Plan) (string, error) {
	entitlements, err := quota.ParseEntitlementJSON(plan.EntitlementsJSON)
	if err != nil {
		return "", fmt.Errorf("parse plan entitlements: %w", err)
	}
	planSnapshot := PlanSnapshot{
		ID: plan.ID, Code: plan.Code, Name: plan.Name, Description: plan.Description,
		PriceAmount: plan.PriceAmount, Currency: plan.Currency, BillingPeriod: plan.BillingPeriod,
	}
	version, err := planVersion(planSnapshot, entitlements)
	if err != nil {
		return "", fmt.Errorf("encode plan version: %w", err)
	}
	snapshot := SubscriptionSnapshot{
		SchemaVersion: subscriptionSnapshotSchemaVersion,
		PlanVersion:   version,
		Plan:          planSnapshot,
		Entitlements:  entitlements,
	}
	encoded, err := json.Marshal(snapshot)
	if err != nil {
		return "", fmt.Errorf("encode subscription snapshot: %w", err)
	}
	return string(encoded), nil
}

// ParseSubscriptionSnapshot reads the current envelope and legacy entitlement-
// only snapshots. Legacy rows remain usable for quota checks but cannot claim
// to contain historical plan display terms that were never persisted.
func ParseSubscriptionSnapshot(payload string) (SubscriptionSnapshot, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(payload), &fields); err == nil && fields != nil {
		if _, versioned := fields["schema_version"]; versioned {
			var snapshot SubscriptionSnapshot
			if err := json.Unmarshal([]byte(payload), &snapshot); err != nil {
				return SubscriptionSnapshot{}, fmt.Errorf("decode subscription snapshot: %w", err)
			}
			if snapshot.SchemaVersion != subscriptionSnapshotSchemaVersion || snapshot.PlanVersion == "" {
				return SubscriptionSnapshot{}, fmt.Errorf("unsupported or incomplete subscription snapshot")
			}
			if _, err := quota.ParseEntitlementJSON(string(fields["entitlements"])); err != nil {
				return SubscriptionSnapshot{}, fmt.Errorf("invalid snapshot entitlements: %w", err)
			}
			version, err := planVersion(snapshot.Plan, snapshot.Entitlements)
			if err != nil {
				return SubscriptionSnapshot{}, fmt.Errorf("invalid snapshot plan: %w", err)
			}
			if snapshot.PlanVersion != version {
				// The watermark entitlement was added after versioned snapshots
				// shipped. Accept the original content hash when the snapshot has
				// no watermark field; historical subscriptions remain immutable.
				legacyVersion, legacyErr := legacyPlanVersion(snapshot.Plan, fields["entitlements"])
				if legacyErr != nil || snapshot.PlanVersion != legacyVersion {
					return SubscriptionSnapshot{}, fmt.Errorf("subscription snapshot version does not match its terms")
				}
			}
			snapshot.HasPlan = true
			return snapshot, nil
		}
	}
	entitlements, err := quota.ParseEntitlementJSON(payload)
	if err != nil {
		return SubscriptionSnapshot{}, err
	}
	return SubscriptionSnapshot{Entitlements: entitlements}, nil
}

func planVersion(plan PlanSnapshot, entitlements quota.Entitlement) (string, error) {
	versionInput, err := json.Marshal(struct {
		Plan         PlanSnapshot      `json:"plan"`
		Entitlements quota.Entitlement `json:"entitlements"`
	}{Plan: plan, Entitlements: entitlements})
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(versionInput)
	return "sha256:" + hex.EncodeToString(digest[:]), nil
}

func legacyPlanVersion(plan PlanSnapshot, payload json.RawMessage) (string, error) {
	var entitlements struct {
		StorageBytes          int64 `json:"storage_bytes"`
		MaxFileBytes          int64 `json:"max_file_bytes"`
		DailyUploads          int64 `json:"daily_uploads"`
		MonthlyAPIUploads     int64 `json:"monthly_api_uploads"`
		MonthlyBandwidthBytes int64 `json:"monthly_bandwidth_bytes"`
		TransformCount        int64 `json:"transform_count"`
		APIRatePerMinute      int64 `json:"api_rate_per_minute"`
		TokenLimit            int64 `json:"token_limit"`
		AdsEnabled            bool  `json:"ads_enabled"`
	}
	if err := json.Unmarshal(payload, &entitlements); err != nil {
		return "", err
	}
	versionInput, err := json.Marshal(struct {
		Plan         PlanSnapshot `json:"plan"`
		Entitlements any          `json:"entitlements"`
	}{Plan: plan, Entitlements: entitlements})
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(versionInput)
	return "sha256:" + hex.EncodeToString(digest[:]), nil
}
