package migrations

import (
	"encoding/json"

	"goravel/app/facades"
	"goravel/app/models"
)

// M20260925000006AddWatermarkEntitlementDefaults adds the new plan capability
// to existing JSON entitlement rows without changing subscription snapshots.
// Built-in paid plans opt in by default; an already explicit value is kept so
// an administrator's choice is never overwritten by a restart.
type M20260925000006AddWatermarkEntitlementDefaults struct{}

func (m *M20260925000006AddWatermarkEntitlementDefaults) Signature() string {
	return "20260925000006_add_watermark_entitlement_defaults"
}

func (m *M20260925000006AddWatermarkEntitlementDefaults) Up() error {
	if !facades.Schema().HasTable("plans") {
		return nil
	}
	var plans []models.Plan
	if err := facades.Orm().Query().Model(&models.Plan{}).Get(&plans); err != nil {
		return err
	}
	for _, plan := range plans {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal([]byte(plan.EntitlementsJSON), &fields); err != nil || fields == nil {
			continue
		}
		if _, exists := fields["watermark_enabled"]; exists {
			continue
		}
		enabled := plan.Code == "creator" || plan.Code == "pro"
		encoded, err := json.Marshal(enabled)
		if err != nil {
			return err
		}
		fields["watermark_enabled"] = encoded
		canonical, err := json.Marshal(fields)
		if err != nil {
			return err
		}
		if _, err := facades.Orm().Query().Model(&models.Plan{}).Where("id = ?", plan.ID).Update(map[string]any{
			"entitlements_json": string(canonical),
		}); err != nil {
			return err
		}
	}
	return nil
}

func (m *M20260925000006AddWatermarkEntitlementDefaults) Down() error {
	// The field is part of the persisted contract; removing it would make the
	// plan editor ambiguous and could silently disable a paid capability.
	return nil
}
