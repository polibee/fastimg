package links

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/goravel/framework/support/carbon"

	"goravel/app/facades"
	"goravel/app/models"
	mediaservices "goravel/app/services/media"
	planservices "goravel/app/services/plans"
	storageservices "goravel/app/services/storage"
)

var (
	ErrMediaNotFound          = errors.New("media not found")
	ErrVariantNotFound        = errors.New("media variant not found")
	ErrInvalidVariant         = errors.New("invalid media variant")
	ErrInvalidSignedURLExpiry = errors.New("invalid signed url expiry")
	ErrLinkSigningUnavailable = errors.New("link signing is unavailable")
	ErrLinkNotFound           = errors.New("link not found")
	ErrHotlinkDenied          = errors.New("hotlink denied")
	ErrHotlinkPolicyInvalid   = errors.New("invalid hotlink policy")
	ErrDomainNotFound         = errors.New("hotlink domain not found")
	ErrDomainConflict         = errors.New("hotlink domain already exists")
)

const (
	DefaultSignedURLLifetime = 10 * time.Minute
	MinSignedURLLifetime     = 1 * time.Minute
	MaxSignedURLLifetime     = 24 * time.Hour
)

var supportedVariants = map[string]struct{}{
	"original":  {},
	"thumbnail": {},
	"medium":    {},
}

type SignedURLInput struct {
	UserID    uint
	MediaID   uint
	Variant   string
	ExpiresIn time.Duration
}

type SignedURLView struct {
	URL       string    `json:"url"`
	Variant   string    `json:"variant"`
	ExpiresAt time.Time `json:"expires_at"`
}

type HotlinkPolicyView struct {
	MediaID        uint   `json:"media_id"`
	Mode           string `json:"mode"`
	AllowNoReferer bool   `json:"allow_no_referer"`
}

type HotlinkDomainView struct {
	ID        uint             `json:"id"`
	Host      string           `json:"host"`
	Status    string           `json:"status"`
	CreatedAt *carbon.DateTime `json:"created_at"`
}

type PublicMedia struct {
	Content     []byte
	ContentType string
}

type StableURLView struct {
	Variant string
	URL     string
}

type Service struct {
	storage storageservices.StorageProvider
	secret  string
	now     func() time.Time
}

func NewService(storage storageservices.StorageProvider, secret string) *Service {
	return &Service{storage: storage, secret: secret, now: time.Now}
}

func ValidateVariant(variant string) error {
	if _, ok := supportedVariants[variant]; !ok {
		return ErrInvalidVariant
	}
	return nil
}

func ValidateSignedURLLifetime(expiresIn time.Duration) error {
	if expiresIn < MinSignedURLLifetime || expiresIn > MaxSignedURLLifetime {
		return ErrInvalidSignedURLExpiry
	}
	return nil
}

func (s *Service) CreateSignedURL(ctx context.Context, input SignedURLInput) (SignedURLView, error) {
	if input.UserID == 0 || input.MediaID == 0 {
		return SignedURLView{}, ErrMediaNotFound
	}
	if err := ValidateVariant(input.Variant); err != nil {
		return SignedURLView{}, err
	}
	if input.ExpiresIn == 0 {
		input.ExpiresIn = DefaultSignedURLLifetime
	}
	if err := ValidateSignedURLLifetime(input.ExpiresIn); err != nil {
		return SignedURLView{}, err
	}
	if strings.TrimSpace(s.secret) == "" {
		return SignedURLView{}, ErrLinkSigningUnavailable
	}
	var asset models.MediaAsset
	if err := facades.Orm().Query().Where("id = ? AND user_id = ? AND status = ?", input.MediaID, input.UserID, "ready").First(&asset); err != nil {
		return SignedURLView{}, ErrMediaNotFound
	}
	var variant models.MediaVariant
	if err := facades.Orm().Query().Where("media_asset_id = ? AND name = ? AND status = ?", asset.ID, input.Variant, "ready").First(&variant); err != nil {
		return SignedURLView{}, ErrVariantNotFound
	}
	var object models.StorageObject
	if err := facades.Orm().Query().Where("id = ? AND status = ?", variant.StorageObjectID, "ready").First(&object); err != nil {
		return SignedURLView{}, ErrVariantNotFound
	}
	_ = object
	expiresAt := s.now().UTC().Add(input.ExpiresIn)
	signature := signMediaURL(input.MediaID, input.Variant, expiresAt, s.secret)
	query := url.Values{}
	query.Set("variant", input.Variant)
	query.Set("expires", strconv.FormatInt(expiresAt.Unix(), 10))
	query.Set("signature", signature)
	return SignedURLView{URL: fmt.Sprintf("/i/%d?%s", input.MediaID, query.Encode()), Variant: input.Variant, ExpiresAt: expiresAt}, nil
}

// CreateStableURLs creates permanent, absolute public URLs for the ready
// variants of a media asset. The URL contains no storage key and is safe to
// expose as a forum/embed link; APP_KEY is the revocation boundary.
func (s *Service) CreateStableURLs(_ context.Context, userID, mediaID uint) (map[string]string, error) {
	if userID == 0 || mediaID == 0 {
		return nil, ErrMediaNotFound
	}
	if strings.TrimSpace(s.secret) == "" {
		return nil, ErrLinkSigningUnavailable
	}
	var asset models.MediaAsset
	if err := facades.Orm().Query().Where("id = ? AND user_id = ? AND status = ?", mediaID, userID, "ready").First(&asset); err != nil {
		return nil, ErrMediaNotFound
	}
	if !mediaservices.AllowsPublicDelivery(asset.Visibility, asset.ModerationStatus) {
		return nil, ErrMediaNotFound
	}
	urls := make(map[string]string, len(supportedVariants))
	for variant := range supportedVariants {
		var mediaVariant models.MediaVariant
		if err := facades.Orm().Query().Where("media_asset_id = ? AND name = ? AND status = ?", asset.ID, variant, "ready").First(&mediaVariant); err != nil {
			continue
		}
		var object models.StorageObject
		if err := facades.Orm().Query().Where("id = ? AND status = ?", mediaVariant.StorageObjectID, "ready").First(&object); err != nil || strings.TrimSpace(object.ObjectKey) == "" {
			continue
		}
		query := url.Values{}
		query.Set("variant", variant)
		query.Set("signature", signStableMediaURL(mediaID, variant, s.secret))
		urls[variant] = absolutePublicURL(fmt.Sprintf("/i/%d?%s", mediaID, query.Encode()))
	}
	if urls["original"] == "" {
		return nil, ErrVariantNotFound
	}
	return urls, nil
}

func (s *Service) PublicSignedContent(ctx context.Context, mediaID uint, variant, expires, signature, referer, userAgent string) (PublicMedia, error) {
	if mediaID == 0 || ValidateVariant(variant) != nil {
		return PublicMedia{}, ErrLinkNotFound
	}
	stable := strings.TrimSpace(expires) == "" && verifyStableMediaURL(mediaID, variant, signature, s.secret)
	expiresUnix, err := strconv.ParseInt(expires, 10, 64)
	if expires != "" && err != nil {
		return PublicMedia{}, ErrLinkNotFound
	}
	expiresAt := time.Unix(expiresUnix, 0).UTC()
	if !stable && !verifyMediaURL(mediaID, variant, expiresAt, signature, s.secret, s.now().UTC()) {
		s.recordAccess(mediaID, 0, variant, HotlinkModeSigned, "invalid_signature", referer, userAgent)
		return PublicMedia{}, ErrLinkNotFound
	}
	policy, domains, err := s.loadPolicy(mediaID)
	if err != nil {
		return PublicMedia{}, ErrLinkNotFound
	}
	if !allowsHotlink(policy.Mode, referer, policy.AllowNoReferer, domains, true) {
		s.recordAccess(mediaID, 0, variant, policy.Mode, "hotlink_denied", referer, userAgent)
		return PublicMedia{}, ErrLinkNotFound
	}
	var asset models.MediaAsset
	if err := facades.Orm().Query().Where("id = ? AND status = ?", mediaID, "ready").First(&asset); err != nil {
		s.recordAccess(mediaID, 0, variant, policy.Mode, "media_not_found", referer, userAgent)
		return PublicMedia{}, ErrLinkNotFound
	}
	if !mediaservices.AllowsSignedDelivery(asset.Visibility, asset.ModerationStatus, stable, expires != "") {
		s.recordAccess(mediaID, 0, variant, policy.Mode, "visibility_denied", referer, userAgent)
		return PublicMedia{}, ErrLinkNotFound
	}
	var mediaVariant models.MediaVariant
	if err := facades.Orm().Query().Where("media_asset_id = ? AND name = ? AND status = ?", asset.ID, variant, "ready").First(&mediaVariant); err != nil {
		s.recordAccess(mediaID, 0, variant, policy.Mode, "media_not_found", referer, userAgent)
		return PublicMedia{}, ErrLinkNotFound
	}
	var object models.StorageObject
	if err := facades.Orm().Query().Where("id = ? AND status = ?", mediaVariant.StorageObjectID, "ready").First(&object); err != nil {
		s.recordAccess(mediaID, 0, variant, policy.Mode, "media_not_found", referer, userAgent)
		return PublicMedia{}, ErrLinkNotFound
	}
	content, err := s.storage.Get(ctx, object.ObjectKey)
	if err != nil {
		s.recordAccess(mediaID, 0, variant, policy.Mode, "storage_unavailable", referer, userAgent)
		return PublicMedia{}, ErrLinkNotFound
	}
	accessID := s.recordAccess(mediaID, 0, variant, policy.Mode, "allowed", referer, userAgent)
	sourceID := fmt.Sprintf("access:%d", accessID)
	if accessID == 0 {
		sourceID = fmt.Sprintf("delivery:%d:%s:%d", mediaID, variant, s.now().UTC().UnixNano())
	}
	if err := planservices.RecordBandwidthUsage(ctx, asset.UserID, int64(len(content)), sourceID, s.now()); err != nil {
		s.recordAccess(mediaID, 0, variant, policy.Mode, "bandwidth_denied", referer, userAgent)
		return PublicMedia{}, err
	}
	return PublicMedia{Content: content, ContentType: object.ContentType}, nil
}

func absolutePublicURL(path string) string {
	base := strings.TrimRight(facades.Config().GetString("app.url", "http://127.0.0.1:53083"), "/")
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	return base + "/" + strings.TrimLeft(path, "/")
}

func (s *Service) Allow(ctx context.Context, mediaID uint, referer string, signatureValid bool, deliveryMode string) error {
	if mediaID == 0 {
		return ErrHotlinkDenied
	}
	policy, domains, err := s.loadPolicy(mediaID)
	if err != nil {
		return ErrHotlinkDenied
	}
	if deliveryMode == "" {
		deliveryMode = policy.Mode
	}
	if !allowsHotlink(policy.Mode, referer, policy.AllowNoReferer, domains, signatureValid) {
		s.recordAccess(mediaID, 0, "original", policy.Mode, "hotlink_denied", referer, "")
		return ErrHotlinkDenied
	}
	s.recordAccess(mediaID, 0, "original", policy.Mode, "allowed", referer, "")
	_ = deliveryMode
	_ = ctx
	return nil
}

func (s *Service) GetPolicy(_ context.Context, userID, mediaID uint) (HotlinkPolicyView, error) {
	if !ownedReadyMedia(userID, mediaID) {
		return HotlinkPolicyView{}, ErrMediaNotFound
	}
	var policy models.HotlinkPolicy
	if err := facades.Orm().Query().Table("media_hotlink_policies").Where("media_asset_id = ? AND user_id = ?", mediaID, userID).First(&policy); err != nil {
		return HotlinkPolicyView{MediaID: mediaID, Mode: HotlinkModeOff}, nil
	}
	if policy.ID == 0 || !validHotlinkMode(policy.Mode) {
		return HotlinkPolicyView{MediaID: mediaID, Mode: HotlinkModeOff}, nil
	}
	return HotlinkPolicyView{MediaID: mediaID, Mode: policy.Mode, AllowNoReferer: policy.AllowNoReferer}, nil
}

func (s *Service) UpdatePolicy(_ context.Context, userID, mediaID uint, mode string, allowNoReferer bool) (HotlinkPolicyView, error) {
	if !ownedReadyMedia(userID, mediaID) {
		return HotlinkPolicyView{}, ErrMediaNotFound
	}
	if !validHotlinkMode(mode) {
		return HotlinkPolicyView{}, ErrHotlinkPolicyInvalid
	}
	result, err := facades.Orm().Query().Table("media_hotlink_policies").Where("media_asset_id = ? AND user_id = ?", mediaID, userID).Update(map[string]any{"mode": mode, "allow_no_referer": allowNoReferer, "updated_at": time.Now().UTC()})
	if err != nil {
		return HotlinkPolicyView{}, err
	}
	if result.RowsAffected == 0 {
		if err := facades.Orm().Query().Table("media_hotlink_policies").Create(&map[string]any{
			"user_id": userID, "media_asset_id": mediaID, "mode": mode, "allow_no_referer": allowNoReferer,
		}); err != nil {
			return HotlinkPolicyView{}, err
		}
	}
	return HotlinkPolicyView{MediaID: mediaID, Mode: mode, AllowNoReferer: allowNoReferer}, nil
}

func (s *Service) ListDomains(_ context.Context, userID uint) ([]HotlinkDomainView, error) {
	var domains []models.HotlinkDomain
	if err := facades.Orm().Query().Table("hotlink_domains").Where("user_id = ? AND status = ?", userID, "active").OrderBy("id", "asc").Get(&domains); err != nil {
		return nil, err
	}
	views := make([]HotlinkDomainView, 0, len(domains))
	for _, domain := range domains {
		views = append(views, HotlinkDomainView{ID: domain.ID, Host: domain.Host, Status: domain.Status, CreatedAt: domain.CreatedAt})
	}
	return views, nil
}

func (s *Service) AddDomain(_ context.Context, userID uint, raw string) (HotlinkDomainView, error) {
	host, err := normalizeHotlinkDomain(raw)
	if err != nil {
		return HotlinkDomainView{}, err
	}
	var existing models.HotlinkDomain
	if err := facades.Orm().Query().Table("hotlink_domains").Where("user_id = ? AND host = ?", userID, host).First(&existing); err == nil && existing.ID > 0 {
		if existing.Status == "active" {
			return HotlinkDomainView{}, ErrDomainConflict
		}
		if _, err := facades.Orm().Query().Table("hotlink_domains").Where("id = ? AND user_id = ?", existing.ID, userID).Update(map[string]any{"status": "active", "updated_at": time.Now().UTC()}); err != nil {
			return HotlinkDomainView{}, err
		}
		return HotlinkDomainView{ID: existing.ID, Host: existing.Host, Status: "active", CreatedAt: existing.CreatedAt}, nil
	}
	domain := models.HotlinkDomain{UserID: userID, Host: host, Status: "active"}
	if err := facades.Orm().Query().Create(&domain); err != nil {
		return HotlinkDomainView{}, err
	}
	return HotlinkDomainView{ID: domain.ID, Host: domain.Host, Status: domain.Status, CreatedAt: domain.CreatedAt}, nil
}

func (s *Service) DeleteDomain(_ context.Context, userID, id uint) error {
	result, err := facades.Orm().Query().Table("hotlink_domains").Where("id = ? AND user_id = ? AND status = ?", id, userID, "active").Update(map[string]any{"status": "revoked", "updated_at": time.Now().UTC()})
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return ErrDomainNotFound
	}
	return nil
}

func (s *Service) loadPolicy(mediaID uint) (models.HotlinkPolicy, []string, error) {
	policy := models.HotlinkPolicy{MediaAssetID: mediaID, Mode: HotlinkModeOff}
	if err := facades.Orm().Query().Table("media_hotlink_policies").Where("media_asset_id = ?", mediaID).First(&policy); err != nil {
		// The default is deliberately fail-open only for the policy row: the
		// signed URL and media readiness checks still remain mandatory.
		policy.Mode = HotlinkModeOff
	}
	if !validHotlinkMode(policy.Mode) {
		policy.Mode = HotlinkModeOff
	}
	var rows []models.HotlinkDomain
	if err := facades.Orm().Query().Table("hotlink_domains").Where("user_id = ? AND status = ?", policy.UserID, "active").Get(&rows); err != nil {
		return policy, nil, err
	}
	domains := make([]string, 0, len(rows))
	for _, row := range rows {
		domains = append(domains, row.Host)
	}
	return policy, domains, nil
}

func (s *Service) recordAccess(mediaID, shareID uint, variant, mode, result, referer, userAgent string) uint {
	if mediaID == 0 {
		return 0
	}
	var shareLinkID *uint
	if shareID > 0 {
		shareLinkID = &shareID
	}
	refererHost := normalizeRefererHost(referer)
	log := models.MediaAccessLog{MediaAssetID: mediaID, ShareLinkID: shareLinkID, Variant: variant, DeliveryMode: mode, Result: result, RefererHost: refererHost, UserAgent: strings.TrimSpace(userAgent), AccessedAt: s.now().UTC()}
	// Observability must never turn a successful media response into a 5xx.
	if err := facades.Orm().Query().Create(&log); err != nil {
		return 0
	}
	return log.ID
}

func ownedReadyMedia(userID, mediaID uint) bool {
	exists, err := facades.Orm().Query().Table("media_assets").Where("id = ? AND user_id = ? AND status = ?", mediaID, userID, "ready").Exists()
	return err == nil && exists
}

func validHotlinkMode(mode string) bool {
	return mode == HotlinkModeOff || mode == HotlinkModeReferer || mode == HotlinkModeSigned || mode == HotlinkModeHybrid
}
