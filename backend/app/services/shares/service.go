package shares

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
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
	ErrShareNotFound         = errors.New("share link not found")
	ErrShareUnavailable      = errors.New("share link unavailable")
	ErrInvalidShareExpiry    = errors.New("invalid share expiry")
	ErrInvalidShareVariant   = errors.New("invalid share variant")
	ErrInvalidSharePassword  = errors.New("invalid share password")
	ErrSharePasswordRequired = errors.New("share password required")
)

const (
	minSharePasswordLength = 8
	maxSharePasswordLength = 72
)

type passwordHasher interface {
	Make(string) (string, error)
	Check(string, string) bool
}

type frameworkPasswordHasher struct{}

func (frameworkPasswordHasher) Make(password string) (string, error) {
	return facades.Hash().Make(password)
}

func (frameworkPasswordHasher) Check(password, hashed string) bool {
	return facades.Hash().Check(password, hashed)
}

type CreateInput struct {
	UserID    uint
	MediaID   uint
	ExpiresAt *time.Time
	Password  string
}

type ShareView struct {
	ID          uint             `json:"id"`
	MediaID     uint             `json:"media_id"`
	Token       string           `json:"token,omitempty"`
	URL         string           `json:"url"`
	TokenPrefix string           `json:"token_prefix"`
	Status      string           `json:"status"`
	ExpiresAt   *time.Time       `json:"expires_at"`
	CreatedAt   *carbon.DateTime `json:"created_at"`
}

type PublicMedia struct {
	Content     []byte
	ContentType string
}

type DeliveryPolicy interface {
	Allow(ctx context.Context, mediaID uint, referer string, signatureValid bool, deliveryMode string) error
}

type Repository interface {
	Create(ctx context.Context, input CreateInput) (models.ShareLink, string, error)
	ListOwned(ctx context.Context, userID uint) ([]models.ShareLink, error)
	Revoke(ctx context.Context, userID, id uint) error
	Resolve(ctx context.Context, token, variant, password string) (mediaservices.MediaVariantObject, error)
}

type DatabaseRepository struct {
	hasher passwordHasher
}

func NewDatabaseRepository() *DatabaseRepository {
	return &DatabaseRepository{hasher: frameworkPasswordHasher{}}
}

func (r *DatabaseRepository) Create(_ context.Context, input CreateInput) (models.ShareLink, string, error) {
	if input.UserID == 0 || input.MediaID == 0 {
		return models.ShareLink{}, "", ErrShareNotFound
	}
	if input.ExpiresAt != nil {
		if err := validateExpiry(*input.ExpiresAt); err != nil {
			return models.ShareLink{}, "", err
		}
	}
	passwordHash, err := makeSharePasswordHash(r.hasher, input.Password)
	if err != nil {
		return models.ShareLink{}, "", err
	}
	exists, err := facades.Orm().Query().Table("media_assets").Where("id = ? AND user_id = ? AND status = ?", input.MediaID, input.UserID, "ready").Exists()
	if err != nil {
		return models.ShareLink{}, "", err
	}
	if !exists {
		return models.ShareLink{}, "", ErrShareNotFound
	}
	raw, err := randomToken()
	if err != nil {
		return models.ShareLink{}, "", err
	}
	link := models.ShareLink{
		UserID: input.UserID, MediaAssetID: input.MediaID,
		TokenHash: tokenHash(raw), TokenPrefix: raw[:8], PasswordHash: passwordHash, Status: "active", ExpiresAt: input.ExpiresAt,
	}
	if err := facades.Orm().Query().Create(&link); err != nil {
		return models.ShareLink{}, "", err
	}
	return link, raw, nil
}

func validateExpiry(expiresAt time.Time) error {
	now := time.Now().UTC()
	if expiresAt.IsZero() {
		return nil
	}
	if !expiresAt.After(now) || expiresAt.After(now.Add(365*24*time.Hour)) {
		return ErrInvalidShareExpiry
	}
	return nil
}

func (r *DatabaseRepository) ListOwned(_ context.Context, userID uint) ([]models.ShareLink, error) {
	var links []models.ShareLink
	if err := facades.Orm().Query().Where("user_id = ?", userID).OrderBy("id", "desc").Get(&links); err != nil {
		return nil, err
	}
	return links, nil
}

func (r *DatabaseRepository) Revoke(_ context.Context, userID, id uint) error {
	now := time.Now().UTC()
	result, err := facades.Orm().Query().Table("share_links").Where("id = ? AND user_id = ? AND status = ?", id, userID, "active").Update(map[string]any{"status": "revoked", "revoked_at": now, "updated_at": now})
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return ErrShareNotFound
	}
	return nil
}

func (r *DatabaseRepository) Resolve(_ context.Context, token, variant, password string) (mediaservices.MediaVariantObject, error) {
	if len(strings.TrimSpace(token)) < 32 {
		return mediaservices.MediaVariantObject{}, ErrShareNotFound
	}
	var link models.ShareLink
	if err := facades.Orm().Query().Where("token_hash = ?", tokenHash(token)).First(&link); err != nil {
		return mediaservices.MediaVariantObject{}, ErrShareNotFound
	}
	if link.Status != "active" || (link.ExpiresAt != nil && !link.ExpiresAt.After(time.Now().UTC())) {
		return mediaservices.MediaVariantObject{}, ErrShareUnavailable
	}
	if err := verifySharePassword(r.hasher, password, link.PasswordHash); err != nil {
		return mediaservices.MediaVariantObject{}, err
	}
	if variant != "original" {
		return mediaservices.MediaVariantObject{}, ErrInvalidShareVariant
	}
	var asset models.MediaAsset
	if err := facades.Orm().Query().Where("id = ? AND user_id = ? AND status = ?", link.MediaAssetID, link.UserID, "ready").First(&asset); err != nil {
		return mediaservices.MediaVariantObject{}, ErrShareUnavailable
	}
	if asset.ModerationStatus == mediaservices.ModerationRejected {
		return mediaservices.MediaVariantObject{}, ErrShareUnavailable
	}
	var mediaVariant models.MediaVariant
	if err := facades.Orm().Query().Where("media_asset_id = ? AND name = ? AND status = ?", asset.ID, variant, "ready").First(&mediaVariant); err != nil {
		return mediaservices.MediaVariantObject{}, ErrShareUnavailable
	}
	var object models.StorageObject
	if err := facades.Orm().Query().Where("id = ? AND status = ?", mediaVariant.StorageObjectID, "ready").First(&object); err != nil {
		return mediaservices.MediaVariantObject{}, ErrShareUnavailable
	}
	return mediaservices.MediaVariantObject{Variant: mediaVariant, Object: object, OwnerUserID: asset.UserID}, nil
}

type Service struct {
	repository Repository
	storage    storageservices.StorageProvider
	policy     DeliveryPolicy
}

func NewService(repository Repository, storage storageservices.StorageProvider) *Service {
	return NewServiceWithPolicy(repository, storage, nil)
}

func NewServiceWithPolicy(repository Repository, storage storageservices.StorageProvider, policy DeliveryPolicy) *Service {
	return &Service{repository: repository, storage: storage, policy: policy}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (ShareView, error) {
	link, raw, err := s.repository.Create(ctx, input)
	if err != nil {
		return ShareView{}, err
	}
	return shareView(link, raw), nil
}

func (s *Service) List(ctx context.Context, userID uint) ([]ShareView, error) {
	links, err := s.repository.ListOwned(ctx, userID)
	if err != nil {
		return nil, err
	}
	views := make([]ShareView, 0, len(links))
	for _, link := range links {
		views = append(views, shareView(link, ""))
	}
	return views, nil
}

func (s *Service) Revoke(ctx context.Context, userID, id uint) error {
	return s.repository.Revoke(ctx, userID, id)
}

func (s *Service) PublicContent(ctx context.Context, token, variant, password, referer string) (PublicMedia, error) {
	item, err := s.repository.Resolve(ctx, token, variant, password)
	if err != nil {
		return PublicMedia{}, err
	}
	if s.policy != nil {
		if err := s.policy.Allow(ctx, item.Variant.MediaAssetID, referer, false, "share"); err != nil {
			return PublicMedia{}, ErrShareUnavailable
		}
	}
	content, err := s.storage.Get(ctx, item.Object.ObjectKey)
	if err != nil {
		return PublicMedia{}, ErrShareUnavailable
	}
	if err := planservices.RecordBandwidthUsage(ctx, item.OwnerUserID, int64(len(content)), fmt.Sprintf("share:%s:%s:%d", token, variant, time.Now().UTC().UnixNano()), time.Now()); err != nil {
		return PublicMedia{}, errors.Join(ErrShareUnavailable, err)
	}
	return PublicMedia{Content: content, ContentType: item.Object.ContentType}, nil
}

func makeSharePasswordHash(hasher passwordHasher, password string) (string, error) {
	password = strings.TrimSpace(password)
	if password == "" {
		return "", nil
	}
	if len([]rune(password)) < minSharePasswordLength || len([]rune(password)) > maxSharePasswordLength {
		return "", ErrInvalidSharePassword
	}
	return hasher.Make(password)
}

func verifySharePassword(hasher passwordHasher, password, passwordHash string) error {
	if passwordHash == "" {
		return nil
	}
	password = strings.TrimSpace(password)
	if password == "" || !hasher.Check(password, passwordHash) {
		return ErrSharePasswordRequired
	}
	return nil
}

func shareView(link models.ShareLink, raw string) ShareView {
	url := ""
	if raw != "" {
		base := strings.TrimRight(facades.Config().GetString("app.url", "http://127.0.0.1:53085"), "/")
		url = base + "/s/" + raw
	}
	return ShareView{ID: link.ID, MediaID: link.MediaAssetID, Token: raw, URL: url, TokenPrefix: link.TokenPrefix, Status: link.Status, ExpiresAt: link.ExpiresAt, CreatedAt: link.CreatedAt}
}

func randomToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func tokenHash(token string) string {
	digest := sha256.Sum256([]byte(token))
	return hex.EncodeToString(digest[:])
}
