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
	ErrDiscoveryMediaNotFound   = errors.New("public discovery media not found")
	ErrDiscoveryVariantNotFound = errors.New("public discovery variant not found")
)

const (
	VisibilityPrivate  = "private"
	VisibilityLink     = "link"
	VisibilityPublic   = "public"
	ModerationPending  = "pending"
	ModerationApproved = "approved"
	ModerationRejected = "rejected"
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
	Enabled bool `json:"enabled"`
}

type PublicVariant struct {
	ObjectKey   string
	ContentType string
	OwnerUserID uint
}

type Repository interface {
	ListPublic(ctx context.Context, page, perPage int) ([]models.MediaAsset, int64, error)
	FindPublicVariant(ctx context.Context, mediaID uint, variant string) (PublicVariant, error)
}

type SortedRepository interface {
	ListPublicSorted(ctx context.Context, page, perPage int, sort string) ([]models.MediaAsset, int64, error)
}

type DatabaseRepository struct{}

func NewDatabaseRepository() *DatabaseRepository { return &DatabaseRepository{} }

func (r *DatabaseRepository) ListPublic(_ context.Context, page, perPage int) ([]models.MediaAsset, int64, error) {
	page, perPage = NormalizePage(page, perPage)
	query := facades.Orm().Query().Table("media_assets").Where(
		"status = ? AND deleted_at IS NULL AND visibility = ? AND moderation_status <> ?",
		"ready", VisibilityPublic, ModerationRejected,
	)
	var items []models.MediaAsset
	var total int64
	if err := query.OrderByDesc("created_at").OrderByDesc("id").Paginate(page, perPage, &items, &total); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *DatabaseRepository) ListPublicSorted(_ context.Context, page, perPage int, sort string) ([]models.MediaAsset, int64, error) {
	page, perPage = NormalizePage(page, perPage)
	if sort != "hot" && sort != "trending" {
		return r.ListPublic(context.Background(), page, perPage)
	}
	where := "ma.status = ? AND ma.deleted_at IS NULL AND ma.visibility = ? AND ma.moderation_status <> ?"
	args := []any{"ready", VisibilityPublic, ModerationRejected}
	total, err := facades.Orm().Query().Table("media_assets").Where("status = ? AND deleted_at IS NULL AND visibility = ? AND moderation_status <> ?", args...).Count()
	if err != nil {
		return nil, 0, err
	}
	order := "(SELECT COUNT(*) FROM media_access_logs mal WHERE mal.media_asset_id = ma.id AND mal.result = 'allowed') DESC, ma.id DESC"
	if sort == "trending" {
		order = "(SELECT COUNT(*) FROM media_access_logs mal WHERE mal.media_asset_id = ma.id AND mal.result = 'allowed' AND mal.accessed_at >= CURRENT_TIMESTAMP - INTERVAL '7 days') DESC, ma.created_at DESC, ma.id DESC"
	}
	statement := fmt.Sprintf("SELECT ma.* FROM media_assets ma WHERE %s ORDER BY %s LIMIT ? OFFSET ?", where, order)
	args = append(args, perPage, (page-1)*perPage)
	var assets []models.MediaAsset
	if err := facades.Orm().Query().Raw(statement, args...).Scan(&assets); err != nil {
		return nil, 0, err
	}
	return assets, total, nil
}

func (r *DatabaseRepository) FindPublicVariant(_ context.Context, mediaID uint, variant string) (PublicVariant, error) {
	if !validVariant(variant) {
		return PublicVariant{}, ErrDiscoveryVariantNotFound
	}
	var asset models.MediaAsset
	if err := facades.Orm().Query().Table("media_assets").Where(
		"id = ? AND status = ? AND deleted_at IS NULL AND visibility = ? AND moderation_status <> ?",
		mediaID, "ready", VisibilityPublic, ModerationRejected,
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

type Service struct {
	repository     Repository
	storage        storageservices.StorageProvider
	storageResolve func(context.Context) (storageservices.StorageProvider, error)
	settings       *settingsservices.SettingService
}

func NewService(repository Repository, storage storageservices.StorageProvider) *Service {
	return &Service{repository: repository, storage: storage, settings: settingsservices.NewSettingService()}
}

// NewServiceWithStorageResolver keeps storage selection at the runtime
// boundary. Public delivery must use the currently configured primary
// provider, not a provider captured when routes were registered.
func NewServiceWithStorageResolver(repository Repository, resolver func(context.Context) (storageservices.StorageProvider, error)) *Service {
	return &Service{repository: repository, storageResolve: resolver, settings: settingsservices.NewSettingService()}
}

func NewDatabaseService(storage storageservices.StorageProvider) *Service {
	return NewService(NewDatabaseRepository(), storage)
}

func NewDatabaseServiceWithRuntimeStorage() *Service {
	disk := facades.Storage().Disk("fastimg")
	return NewServiceWithStorageResolver(NewDatabaseRepository(), func(_ context.Context) (storageservices.StorageProvider, error) {
		provider, _, err := storageservices.NewRuntimeRegistry(disk).Primary()
		return provider, err
	})
}

func (s *Service) Status(_ context.Context) Status {
	return Status{Enabled: settingBool(s.settings.Resolve("discover.enabled", "true"))}
}

func (s *Service) Feed(ctx context.Context, page, perPage int) (FeedPage, error) {
	return s.FeedSorted(ctx, page, perPage, "latest")
}

func (s *Service) FeedSorted(ctx context.Context, page, perPage int, sort string) (FeedPage, error) {
	status := s.Status(ctx)
	if !status.Enabled {
		return FeedPage{}, ErrDiscoveryDisabled
	}
	page, perPage = NormalizePage(page, perPage)
	if sort != "hot" && sort != "trending" {
		sort = "latest"
	}
	var assets []models.MediaAsset
	var total int64
	var err error
	if repository, ok := s.repository.(SortedRepository); ok {
		assets, total, err = repository.ListPublicSorted(ctx, page, perPage, sort)
	} else {
		assets, total, err = s.repository.ListPublic(ctx, page, perPage)
	}
	if err != nil {
		return FeedPage{}, err
	}
	items := make([]FeedItem, 0, len(assets))
	for _, asset := range assets {
		items = append(items, FeedItem{
			ID: asset.ID, OriginalName: asset.OriginalName, ContentType: asset.ContentType,
			Width: asset.Width, Height: asset.Height, SizeBytes: asset.SizeBytes,
			CreatedAt: asset.CreatedAt, ThumbnailURL: contentURL(asset.ID, "original"), OriginalURL: contentURL(asset.ID, "original"),
		})
	}
	return FeedPage{Items: items, Page: page, PerPage: perPage, Total: total}, nil
}

func (s *Service) Content(ctx context.Context, mediaID uint, variant string) (string, []byte, error) {
	return s.content(ctx, mediaID, variant, true)
}

// ContentPublic serves media that belongs to an explicitly public album. It
// keeps the same visibility, moderation, storage and bandwidth checks as the
// discovery delivery path, but does not make an album depend on the optional
// discovery feed switch.
func (s *Service) ContentPublic(ctx context.Context, mediaID uint, variant string) (string, []byte, error) {
	return s.content(ctx, mediaID, variant, false)
}

func (s *Service) content(ctx context.Context, mediaID uint, variant string, requireDiscovery bool) (string, []byte, error) {
	if requireDiscovery && !s.Status(ctx).Enabled {
		return "", nil, ErrDiscoveryDisabled
	}
	publicVariant, err := s.repository.FindPublicVariant(ctx, mediaID, variant)
	if err != nil {
		return "", nil, err
	}
	storage, err := s.resolveStorage(ctx)
	if err != nil {
		return "", nil, fmt.Errorf("resolve public discovery storage: %w", err)
	}
	content, err := storage.Get(ctx, publicVariant.ObjectKey)
	if err != nil {
		return "", nil, fmt.Errorf("read public discovery media: %w", err)
	}
	if err := planservices.RecordBandwidthUsage(ctx, publicVariant.OwnerUserID, int64(len(content)), fmt.Sprintf("discovery:%d:%s:%d", mediaID, variant, time.Now().UTC().UnixNano()), time.Now()); err != nil {
		return "", nil, err
	}
	return publicVariant.ContentType, content, nil
}

func (s *Service) resolveStorage(ctx context.Context) (storageservices.StorageProvider, error) {
	if s.storageResolve != nil {
		return s.storageResolve(ctx)
	}
	if s.storage == nil {
		return nil, errors.New("discovery storage is not configured")
	}
	return s.storage, nil
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

// IsDiscoverable is the shared post-moderation rule for a public media asset.
// Uploads are available immediately; only an explicit rejection removes them
// from discovery. Private and link-only media are never listed here.
func IsDiscoverable(visibility, moderationStatus string) bool {
	return visibility == VisibilityPublic && moderationStatus != ModerationRejected
}

func settingBool(value string) bool { return strings.EqualFold(strings.TrimSpace(value), "true") }

func validVariant(value string) bool {
	return value == "original"
}

func contentURL(mediaID uint, variant string) string {
	base := strings.TrimRight(facades.Config().GetString("app.url", "http://127.0.0.1:53085"), "/")
	return fmt.Sprintf("%s/api/v1/discovery/media/%d/content?variant=%s", base, mediaID, variant)
}
