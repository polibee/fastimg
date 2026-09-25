package discovery

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"goravel/app/facades"
	"goravel/app/models"
	planservices "goravel/app/services/plans"
	settingsservices "goravel/app/services/settings"
	storageservices "goravel/app/services/storage"
)

var (
	ErrDiscoveryDisabled        = errors.New("discovery is disabled")
	ErrDiscoverySubmissionsOff  = errors.New("discovery submissions are disabled")
	ErrDiscoveryMediaNotFound   = errors.New("public discovery media not found")
	ErrDiscoveryVariantNotFound = errors.New("public discovery variant not found")
	ErrDiscoveryAlreadyPending  = errors.New("discovery submission already pending")
)

const (
	VisibilityPrivate  = "private"
	VisibilityPublic   = "public"
	ModerationPending  = "pending"
	ModerationApproved = "approved"
)

type FeedItem struct {
	ID           uint   `json:"id"`
	OriginalName string `json:"original_name"`
	ContentType  string `json:"content_type"`
	Width        int64  `json:"width"`
	Height       int64  `json:"height"`
	SizeBytes    int64  `json:"size_bytes"`
	CreatedAt    any    `json:"created_at"`
	ThumbnailURL string `json:"thumbnail_url"`
	OriginalURL  string `json:"original_url"`
}

type FeedPage struct {
	Items   []FeedItem `json:"items"`
	Page    int        `json:"page"`
	PerPage int        `json:"per_page"`
	Total   int64      `json:"total"`
}

type Status struct {
	Enabled            bool `json:"enabled"`
	SubmissionsEnabled bool `json:"submissions_enabled"`
}

type PublicVariant struct {
	ObjectKey   string
	ContentType string
	OwnerUserID uint
}

type Repository interface {
	ListPublic(ctx context.Context, page, perPage int) ([]models.MediaAsset, int64, error)
	FindPublicVariant(ctx context.Context, mediaID uint, variant string) (PublicVariant, error)
	Submit(ctx context.Context, userID, mediaID uint) (models.MediaAsset, error)
}

type DatabaseRepository struct{}

func NewDatabaseRepository() *DatabaseRepository { return &DatabaseRepository{} }

func (r *DatabaseRepository) ListPublic(_ context.Context, page, perPage int) ([]models.MediaAsset, int64, error) {
	page, perPage = NormalizePage(page, perPage)
	query := facades.Orm().Query().Table("media_assets").Where(
		"status = ? AND deleted_at IS NULL AND visibility = ? AND moderation_status = ?",
		"ready", VisibilityPublic, ModerationApproved,
	)
	var items []models.MediaAsset
	var total int64
	if err := query.OrderByDesc("created_at").OrderByDesc("id").Paginate(page, perPage, &items, &total); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *DatabaseRepository) FindPublicVariant(_ context.Context, mediaID uint, variant string) (PublicVariant, error) {
	if !validVariant(variant) {
		return PublicVariant{}, ErrDiscoveryVariantNotFound
	}
	var asset models.MediaAsset
	if err := facades.Orm().Query().Table("media_assets").Where(
		"id = ? AND status = ? AND deleted_at IS NULL AND visibility = ? AND moderation_status = ?",
		mediaID, "ready", VisibilityPublic, ModerationApproved,
	).First(&asset); err != nil {
		return PublicVariant{}, ErrDiscoveryMediaNotFound
	}
	var mediaVariant models.MediaVariant
	if err := facades.Orm().Query().Where("media_asset_id = ? AND name = ? AND status = ?", mediaID, variant, "ready").First(&mediaVariant); err != nil {
		return PublicVariant{}, ErrDiscoveryVariantNotFound
	}
	var object models.StorageObject
	if err := facades.Orm().Query().Where("id = ? AND status = ?", mediaVariant.StorageObjectID, "ready").First(&object); err != nil || strings.TrimSpace(object.ObjectKey) == "" {
		return PublicVariant{}, ErrDiscoveryVariantNotFound
	}
	return PublicVariant{ObjectKey: object.ObjectKey, ContentType: object.ContentType, OwnerUserID: asset.UserID}, nil
}

func (r *DatabaseRepository) Submit(_ context.Context, userID, mediaID uint) (models.MediaAsset, error) {
	var asset models.MediaAsset
	if err := facades.Orm().Query().Where("id = ? AND user_id = ? AND status = ? AND deleted_at IS NULL", mediaID, userID, "ready").First(&asset); err != nil {
		return models.MediaAsset{}, ErrDiscoveryMediaNotFound
	}
	if asset.Visibility == VisibilityPublic && asset.ModerationStatus == ModerationPending {
		return asset, ErrDiscoveryAlreadyPending
	}
	now := time.Now().UTC()
	if _, err := facades.Orm().Query().Table("media_assets").Where("id = ? AND user_id = ?", mediaID, userID).Update(map[string]any{
		"visibility": VisibilityPublic, "moderation_status": ModerationPending, "discovery_submitted_at": now, "updated_at": now,
	}); err != nil {
		return models.MediaAsset{}, err
	}
	asset.Visibility = VisibilityPublic
	asset.ModerationStatus = ModerationPending
	return asset, nil
}

type Service struct {
	repository Repository
	storage    storageservices.StorageProvider
	settings   *settingsservices.SettingService
}

func NewService(repository Repository, storage storageservices.StorageProvider) *Service {
	return &Service{repository: repository, storage: storage, settings: settingsservices.NewSettingService()}
}

func NewDatabaseService(storage storageservices.StorageProvider) *Service {
	return NewService(NewDatabaseRepository(), storage)
}

func (s *Service) Status(_ context.Context) Status {
	return Status{Enabled: settingBool(s.settings.Resolve("discover.enabled", "true")), SubmissionsEnabled: settingBool(s.settings.Resolve("discover.submissions_enabled", "true"))}
}

func (s *Service) Feed(ctx context.Context, page, perPage int) (FeedPage, error) {
	status := s.Status(ctx)
	if !status.Enabled {
		return FeedPage{}, ErrDiscoveryDisabled
	}
	page, perPage = NormalizePage(page, perPage)
	assets, total, err := s.repository.ListPublic(ctx, page, perPage)
	if err != nil {
		return FeedPage{}, err
	}
	items := make([]FeedItem, 0, len(assets))
	for _, asset := range assets {
		items = append(items, FeedItem{
			ID: asset.ID, OriginalName: asset.OriginalName, ContentType: asset.ContentType,
			Width: asset.Width, Height: asset.Height, SizeBytes: asset.SizeBytes,
			CreatedAt: asset.CreatedAt, ThumbnailURL: contentURL(asset.ID, "thumbnail"), OriginalURL: contentURL(asset.ID, "original"),
		})
	}
	return FeedPage{Items: items, Page: page, PerPage: perPage, Total: total}, nil
}

func (s *Service) Content(ctx context.Context, mediaID uint, variant string) (string, []byte, error) {
	if !s.Status(ctx).Enabled {
		return "", nil, ErrDiscoveryDisabled
	}
	publicVariant, err := s.repository.FindPublicVariant(ctx, mediaID, variant)
	if err != nil {
		return "", nil, err
	}
	content, err := s.storage.Get(ctx, publicVariant.ObjectKey)
	if err != nil {
		return "", nil, fmt.Errorf("read public discovery media: %w", err)
	}
	if err := planservices.RecordBandwidthUsage(ctx, publicVariant.OwnerUserID, int64(len(content)), fmt.Sprintf("discovery:%d:%s:%d", mediaID, variant, time.Now().UTC().UnixNano()), time.Now()); err != nil {
		return "", nil, err
	}
	return publicVariant.ContentType, content, nil
}

func (s *Service) Submit(ctx context.Context, userID, mediaID uint) error {
	status := s.Status(ctx)
	if !status.Enabled {
		return ErrDiscoveryDisabled
	}
	if !status.SubmissionsEnabled {
		return ErrDiscoverySubmissionsOff
	}
	_, err := s.repository.Submit(ctx, userID, mediaID)
	return err
}

func NormalizePage(page, perPage int) (int, int) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 24
	}
	if perPage > 48 {
		perPage = 48
	}
	return page, perPage
}

func settingBool(value string) bool { return strings.EqualFold(strings.TrimSpace(value), "true") }

func validVariant(value string) bool {
	return value == "original" || value == "thumbnail" || value == "medium"
}

func contentURL(mediaID uint, variant string) string {
	base := strings.TrimRight(facades.Config().GetString("app.url", "http://127.0.0.1:53083"), "/")
	return fmt.Sprintf("%s/api/v1/discovery/media/%d/content?variant=%s", base, mediaID, variant)
}
