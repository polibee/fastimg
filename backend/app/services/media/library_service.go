package media

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"goravel/app/models"
	planservices "goravel/app/services/plans"
	storageservices "goravel/app/services/storage"
)

var ErrInvalidVariantName = errors.New("invalid media variant")

type PermanentDeleteRepository interface {
	PreparePermanentDelete(ctx context.Context, userID, mediaID uint) (models.MediaAsset, []models.StorageObject, error)
	FinalizePermanentDelete(ctx context.Context, userID, mediaID uint) error
}

type MediaFolderRepository interface {
	AssignFolder(ctx context.Context, userID, mediaID uint, folderID *uint) (models.MediaAsset, error)
}

type MediaContent struct {
	Bytes       []byte
	ContentType string
}

type MediaLibraryService struct {
	repository MediaLibraryRepository
	storage    storageservices.StorageProvider
}

func NewMediaLibraryService(repository MediaLibraryRepository, storage storageservices.StorageProvider) *MediaLibraryService {
	return &MediaLibraryService{repository: repository, storage: storage}
}

func (s *MediaLibraryService) List(ctx context.Context, userID uint, trash bool, search string, page, perPage int) ([]MediaListItem, int64, error) {
	return s.repository.ListOwned(ctx, userID, trash, strings.TrimSpace(search), page, perPage)
}

func (s *MediaLibraryService) GetDetails(ctx context.Context, userID, mediaID uint) (MediaListItem, error) {
	asset, err := s.repository.FindOwned(ctx, userID, mediaID, false)
	if err != nil {
		return MediaListItem{}, err
	}
	item := MediaListItem{Asset: asset}
	for _, name := range []string{"original"} {
		variant, variantErr := s.repository.FindOwnedVariant(ctx, userID, mediaID, name)
		if errors.Is(variantErr, ErrVariantNotFound) {
			continue
		}
		if variantErr != nil {
			return MediaListItem{}, variantErr
		}
		if variant.Variant.Status == "ready" && variant.Object.Status == "ready" {
			item.Variants = append(item.Variants, variant)
		}
	}
	return item, nil
}

func (s *MediaLibraryService) GetContent(ctx context.Context, userID, mediaID uint, name string) (MediaContent, error) {
	switch name {
	case "original":
	default:
		return MediaContent{}, ErrInvalidVariantName
	}
	item, err := s.repository.FindOwnedVariant(ctx, userID, mediaID, name)
	if err != nil {
		return MediaContent{}, err
	}
	if item.Variant.Status != "ready" || item.Object.Status != "ready" || strings.TrimSpace(item.Object.ContentType) == "" {
		return MediaContent{}, ErrVariantNotFound
	}
	content, err := s.storage.Get(ctx, item.Object.ObjectKey)
	if err != nil {
		return MediaContent{}, err
	}
	if err := planservices.RecordBandwidthUsage(ctx, userID, int64(len(content)), fmt.Sprintf("media:%d:%s:%d", mediaID, name, time.Now().UTC().UnixNano()), time.Now()); err != nil {
		return MediaContent{}, err
	}
	return MediaContent{Bytes: content, ContentType: item.Object.ContentType}, nil
}

func (s *MediaLibraryService) UpdateVisibility(ctx context.Context, userID, mediaID uint, visibility string) (models.MediaAsset, error) {
	repository, ok := s.repository.(interface {
		UpdateVisibility(context.Context, uint, uint, string) (models.MediaAsset, error)
	})
	if !ok {
		return models.MediaAsset{}, errors.New("media visibility update is not supported")
	}
	return repository.UpdateVisibility(ctx, userID, mediaID, visibility)
}

func (s *MediaLibraryService) SoftDelete(ctx context.Context, userID, mediaID uint) (models.MediaAsset, error) {
	return s.repository.SoftDelete(ctx, userID, mediaID)
}

func (s *MediaLibraryService) Restore(ctx context.Context, userID, mediaID uint) (models.MediaAsset, error) {
	return s.repository.Restore(ctx, userID, mediaID)
}

func (s *MediaLibraryService) AssignFolder(ctx context.Context, userID, mediaID uint, folderID *uint) (models.MediaAsset, error) {
	repository, ok := s.repository.(MediaFolderRepository)
	if !ok {
		return models.MediaAsset{}, errors.New("media folder assignment is not supported")
	}
	return repository.AssignFolder(ctx, userID, mediaID, folderID)
}

// PermanentDelete removes all stored variants before the repository releases
// the user's storage usage. Provider deletes must be idempotent so a retry can
// safely resume after a partial failure.
func (s *MediaLibraryService) PermanentDelete(ctx context.Context, userID, mediaID uint) (models.MediaAsset, error) {
	repository, ok := s.repository.(PermanentDeleteRepository)
	if !ok {
		return models.MediaAsset{}, errors.New("permanent media deletion is not supported")
	}
	asset, objects, err := repository.PreparePermanentDelete(ctx, userID, mediaID)
	if err != nil {
		return models.MediaAsset{}, err
	}
	deletedKeys := make(map[string]struct{}, len(objects))
	for _, object := range objects {
		if object.Status == "deleted" {
			continue
		}
		if object.Status != "ready" || strings.TrimSpace(object.ObjectKey) == "" {
			return models.MediaAsset{}, fmt.Errorf("media %d contains a storage object that is not ready", mediaID)
		}
		if _, exists := deletedKeys[object.ObjectKey]; exists {
			continue
		}
		if err := s.storage.Delete(ctx, object.ObjectKey); err != nil {
			return models.MediaAsset{}, fmt.Errorf("delete media storage object: %w", err)
		}
		deletedKeys[object.ObjectKey] = struct{}{}
	}
	if err := repository.FinalizePermanentDelete(ctx, userID, mediaID); err != nil {
		return models.MediaAsset{}, fmt.Errorf("finalize permanent media deletion: %w", err)
	}
	asset.Status = "physically_deleted"
	return asset, nil
}

func PermanentDeleteConfirmed(value string) bool { return value == "permanently-delete" }

// EmptyTrash permanently deletes the current user's trash one asset at a time.
// If cleanup fails, already finalized assets stay deleted and remaining items
// remain visible for a later retry.
func (s *MediaLibraryService) EmptyTrash(ctx context.Context, userID uint) (int, error) {
	deleted := 0
	for {
		items, _, err := s.List(ctx, userID, true, "", 1, 100)
		if err != nil {
			return deleted, err
		}
		if len(items) == 0 {
			return deleted, nil
		}
		for _, item := range items {
			if _, err := s.PermanentDelete(ctx, userID, item.Asset.ID); err != nil {
				return deleted, err
			}
			deleted++
		}
	}
}

func EmptyTrashConfirmed(value string) bool { return value == "empty-trash" }

func canRestoreMediaStatus(status string) bool { return status == "deleted" }

func mediaTrashStatuses() []any { return []any{"deleted", "cleanup_pending"} }
