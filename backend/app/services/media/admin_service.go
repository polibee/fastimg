package media

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"goravel/app/facades"
	auditservices "goravel/app/services/audit"
	storageservices "goravel/app/services/storage"
)

var ErrInvalidAdminMediaOperation = errors.New("invalid administrator media operation")

func ValidateAdminMediaOperation(operation string) error {
	switch strings.TrimSpace(operation) {
	case "hide", "restore", "approve", "reject", "permanent_delete":
		return nil
	default:
		return ErrInvalidAdminMediaOperation
	}
}

type AdminService struct {
	repository *DatabaseRepository
	storage    storageservices.StorageProvider
}

func NewAdminService() *AdminService {
	return &AdminService{
		repository: NewDatabaseRepository(),
		storage:    storageservices.NewLocalProvider(facades.Storage().Disk("fastimg")),
	}
}

// Apply is the administrator-only media lifecycle boundary. Generic resource
// CRUD may list media, but every state-changing operation comes through here
// so ownership, storage cleanup, and audit recording stay together.
func (s *AdminService) Apply(ctx context.Context, operatorID, mediaID uint, operation string) (err error) {
	operation = strings.TrimSpace(operation)
	if err = ValidateAdminMediaOperation(operation); err != nil {
		return err
	}
	outcome := "failed"
	defer func() {
		metadata := map[string]any{
			"media_asset_id": mediaID,
			"operation":      operation,
			"outcome":        outcome,
		}
		if err != nil {
			metadata["error"] = err.Error()
		}
		_ = auditservices.NewAuditService().Record(operatorID, "admin.media."+operation, metadata)
	}()

	var asset struct {
		ID     uint
		UserID uint
		Status string
	}
	if queryErr := facades.Orm().Query().Table("media_assets").Where("id = ?", mediaID).First(&asset); queryErr != nil || asset.ID == 0 {
		return ErrMediaNotFound
	}

	switch operation {
	case "permanent_delete":
		if asset.Status != "deleted" && asset.Status != "cleanup_pending" && asset.Status != "physically_deleted" {
			if _, err = s.repository.SoftDelete(ctx, asset.UserID, mediaID); err != nil {
				return err
			}
		}
		_, err = NewMediaLibraryService(s.repository, s.storage).PermanentDelete(ctx, asset.UserID, mediaID)
		if err != nil {
			return err
		}
	default:
		if operation == "restore" {
			if asset.Status == "deleted" {
				if _, restoreErr := s.repository.Restore(ctx, asset.UserID, mediaID); restoreErr != nil {
					return restoreErr
				}
			}
		}
		updates := moderationStateUpdates(operation)
		updates["updated_at"] = time.Now().UTC()
		if _, err = facades.Orm().Query().Table("media_assets").Where("id = ?", mediaID).Update(updates); err != nil {
			return fmt.Errorf("update media moderation state: %w", err)
		}
	}
	outcome = "succeeded"
	return nil
}

func moderationStateUpdates(operation string) map[string]any {
	switch operation {
	case "hide":
		return map[string]any{"visibility": VisibilityPrivate}
	case "reject":
		return map[string]any{"visibility": VisibilityPrivate, "moderation_status": ModerationRejected}
	case "restore":
		return map[string]any{"visibility": VisibilityPublic, "moderation_status": ModerationApproved}
	case "approve":
		return map[string]any{"moderation_status": ModerationApproved}
	default:
		return map[string]any{}
	}
}
