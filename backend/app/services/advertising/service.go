package advertising

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"goravel/app/facades"
	planservices "goravel/app/services/plans"
	"goravel/app/services/quota"
)

const (
	CreativeTypeText   = "text"
	CreativeTypeImage  = "image"
	CreativeTypeScript = "script"

	PlacementHeader = "header"
	PlacementFooter = "footer"
	PlacementLeft   = "left"
	PlacementRight  = "right"
)

var (
	ErrInvalidCreativeType = errors.New("invalid advertising creative type")
	ErrInvalidPlacement    = errors.New("invalid advertising placement")
	ErrCreativeRequired    = errors.New("advertising creative content is required")
	ErrUnsafeTargetURL     = errors.New("advertising target URL must use http or https")
)

type PublicAd struct {
	ID              uint   `json:"id"`
	Name            string `json:"name"`
	Placement       string `json:"placement"`
	CreativeType    string `json:"creative_type"`
	CreativeContent string `json:"creative_content"`
	TargetURL       string `json:"target_url,omitempty"`
}

type storedAd struct {
	ID              uint
	Name            string
	Placement       string
	CreativeURL     string
	CreativeType    string
	CreativeContent string
	TargetURL       string
	PlanCode        string
	StartsAt        string
	EndsAt          string
}

type legacyStoredAd struct {
	ID          uint
	Name        string
	Placement   string
	CreativeURL string
	TargetURL   string
}

func PrepareAdvertisingWrite(_ string, payload map[string]any) error {
	placement := strings.TrimSpace(stringValue(payload["placement"]))
	if !ValidPlacement(placement) {
		return fmt.Errorf("%w: %s", ErrInvalidPlacement, placement)
	}

	creativeType := strings.TrimSpace(stringValue(payload["creative_type"]))
	if !ValidCreativeType(creativeType) {
		return fmt.Errorf("%w: %s", ErrInvalidCreativeType, creativeType)
	}
	creativeContent := strings.TrimSpace(stringValue(payload["creative_content"]))
	if creativeContent == "" {
		return ErrCreativeRequired
	}
	targetURL := strings.TrimSpace(stringValue(payload["target_url"]))
	if targetURL != "" {
		parsed, err := url.Parse(targetURL)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			return ErrUnsafeTargetURL
		}
	}

	payload["placement"] = placement
	payload["creative_type"] = creativeType
	payload["creative_content"] = creativeContent
	if _, exists := payload["target_url"]; exists {
		payload["target_url"] = targetURL
	}
	return nil
}

func ValidPlacement(value string) bool {
	switch value {
	case PlacementHeader, PlacementFooter, PlacementLeft, PlacementRight:
		return true
	default:
		return false
	}
}

func ValidCreativeType(value string) bool {
	switch value {
	case CreativeTypeText, CreativeTypeImage, CreativeTypeScript:
		return true
	default:
		return false
	}
}

// AdAvailableForUser is the single delivery predicate for plan-targeted and
// scheduled ads. Empty plan code means the ad is global. The end boundary is
// exclusive so an ad never remains visible after its configured expiry.
func AdAvailableForUser(userPlanCode, adPlanCode string, startsAt, endsAt *time.Time, now time.Time) bool {
	if adPlanCode = strings.TrimSpace(adPlanCode); adPlanCode != "" && !strings.EqualFold(adPlanCode, strings.TrimSpace(userPlanCode)) {
		return false
	}
	if startsAt != nil && now.Before(startsAt.UTC()) {
		return false
	}
	if endsAt != nil && !now.Before(endsAt.UTC()) {
		return false
	}
	return true
}

func (s *Service) ListPublished(placement string) ([]PublicAd, error) {
	return s.listPublished(placement, "", time.Now().UTC())
}

func (s *Service) listPublished(placement, planCode string, now time.Time) ([]PublicAd, error) {
	if placement != "" && !ValidPlacement(placement) {
		return nil, ErrInvalidPlacement
	}

	query := facades.Orm().Query().Table("advertising").Where("status = ?", "active")
	if placement != "" {
		query = query.Where("placement = ?", placement)
	}
	query = query.OrderBy("id", "desc")

	if facades.Schema().HasColumn("advertising", "creative_content") {
		var rows []storedAd
		if err := query.Get(&rows); err != nil {
			return nil, err
		}
		return publicAds(filterAds(rows, planCode, now)), nil
	}

	var rows []legacyStoredAd
	if err := query.Get(&rows); err != nil {
		return nil, err
	}
	result := make([]PublicAd, 0, len(rows))
	for _, row := range rows {
		result = append(result, PublicAd{ID: row.ID, Name: row.Name, Placement: row.Placement, CreativeType: CreativeTypeImage, CreativeContent: row.CreativeURL, TargetURL: row.TargetURL})
	}
	return result, nil
}

// ShouldShowAds is the single runtime decision boundary for member ads.
// Administrators may still preview/manage ads, but member delivery must use
// the effective subscription entitlement.
func ShouldShowAds(enabled bool) bool { return enabled }

func (s *Service) ListPublishedForUser(userID uint, placement string) ([]PublicAd, error) {
	subscription, err := planservices.NewPlanService().SubscriptionForUser(userID)
	if err != nil {
		return nil, err
	}
	entitlement, err := quota.ParseEntitlementJSON(subscription.EntitlementSnapshotJSON)
	if err != nil {
		return nil, err
	}
	if !ShouldShowAds(entitlement.AdsEnabled) {
		return []PublicAd{}, nil
	}
	snapshot, err := planservices.ParseSubscriptionSnapshot(subscription.EntitlementSnapshotJSON)
	if err != nil {
		return nil, err
	}
	planCode := ""
	if snapshot.HasPlan {
		planCode = snapshot.Plan.Code
	}
	return s.listPublished(placement, planCode, time.Now().UTC())
}

type Service struct{}

func NewService() *Service { return &Service{} }

func publicAds(rows []storedAd) []PublicAd {
	result := make([]PublicAd, 0, len(rows))
	for _, row := range rows {
		creativeType := strings.TrimSpace(row.CreativeType)
		if !ValidCreativeType(creativeType) {
			creativeType = CreativeTypeImage
		}
		content := strings.TrimSpace(row.CreativeContent)
		if content == "" {
			content = row.CreativeURL
		}
		result = append(result, PublicAd{ID: row.ID, Name: row.Name, Placement: row.Placement, CreativeType: creativeType, CreativeContent: content, TargetURL: row.TargetURL})
	}
	return result
}

func filterAds(rows []storedAd, planCode string, now time.Time) []storedAd {
	filtered := make([]storedAd, 0, len(rows))
	for _, row := range rows {
		if AdAvailableForUser(planCode, row.PlanCode, parseScheduleTime(row.StartsAt), parseScheduleTime(row.EndsAt), now) {
			filtered = append(filtered, row)
		}
	}
	return filtered
}

func parseScheduleTime(value string) *time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	for _, layout := range []string{time.RFC3339Nano, "2006-01-02 15:04:05", "2006-01-02T15:04:05"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			parsed = parsed.UTC()
			return &parsed
		}
	}
	return nil
}

func stringValue(value any) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}
