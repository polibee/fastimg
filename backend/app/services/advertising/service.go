package advertising

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"goravel/app/facades"
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
	if creativeType == CreativeTypeScript && strings.Contains(strings.ToLower(creativeContent), "</script") {
		return fmt.Errorf("%w: script content must be JavaScript source without a script tag", ErrCreativeRequired)
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

func (s *Service) ListPublished(placement string) ([]PublicAd, error) {
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
		return publicAds(rows), nil
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

func stringValue(value any) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}
