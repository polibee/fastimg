package media

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/database/orm"

	"goravel/app/facades"
	"goravel/app/models"
	planservices "goravel/app/services/plans"
	"goravel/app/services/quota"
	rbacservices "goravel/app/services/rbac"
)

var (
	ErrAccountUnavailable           = errors.New("account is not available for uploads")
	ErrSubscriptionUnavailable      = errors.New("active subscription is unavailable")
	ErrIdempotencyConflict          = errors.New("idempotency key was already used for different content")
	ErrUploadInProgress             = errors.New("upload with this idempotency key is still processing")
	ErrDailyUploadLimit             = errors.New("daily upload limit reached")
	ErrMonthlyAPIUploadLimit        = errors.New("monthly API upload limit reached")
	ErrMonthlyTransformLimit        = errors.New("monthly image transform limit reached")
	ErrMediaNotFound                = errors.New("media not found")
	ErrFolderNotFound               = errors.New("folder not found")
	ErrMediaSharedStorage           = errors.New("media shares a storage object with another asset")
	ErrVariantNotFound              = errors.New("media variant not found")
	ErrInvalidVisibility            = errors.New("invalid media visibility")
	ErrStorageConnectionUnavailable = errors.New("storage connection is unavailable")
	ErrInvalidMediaExpiry           = errors.New("invalid media expiry")
)

type DatabaseRepository struct{}

func NewDatabaseRepository() *DatabaseRepository { return &DatabaseRepository{} }

// CheckUploadAllowance rejects known quota violations before image decoding.
// BeginUpload repeats the same checks while holding the user's row lock and
// persists the actual object-byte reservation after variants are prepared.
func (r *DatabaseRepository) CheckUploadAllowance(_ context.Context, input UploadMetadata) error {
	return facades.Orm().Transaction(func(tx orm.Query) error {
		var user models.User
		if err := tx.Where("id = ?", input.UserID).LockForUpdate().First(&user); err != nil {
			return err
		}
		if user.Status != "active" {
			return ErrAccountUnavailable
		}
		exists, err := tx.Model(&models.UploadSession{}).
			Where("user_id = ? AND idempotency_key = ?", input.UserID, input.IdempotencyKey).Exists()
		if err != nil {
			return err
		}
		currentSessionID := uint(0)
		if exists {
			var session models.UploadSession
			if err := tx.Where("user_id = ? AND idempotency_key = ?", input.UserID, input.IdempotencyKey).First(&session); err != nil {
				return err
			}
			currentSessionID = session.ID
			if session.SHA256 != input.SHA256 || session.SizeBytes != input.SizeBytes {
				return ErrIdempotencyConflict
			}
			if session.Status == "ready" || session.Status == "processing" {
				return nil
			}
		}
		return checkUploadQuota(tx, input, input.SizeBytes, currentSessionID)
	})
}

func (r *DatabaseRepository) BeginUpload(_ context.Context, input UploadMetadata, objects []PreparedObject) (reservation UploadReservation, err error) {
	err = facades.Orm().Transaction(func(tx orm.Query) error {
		storageConnectionID := input.StorageConnectionID
		if storageConnectionID == 0 {
			var local models.StorageConnection
			if err := tx.Where("provider_code = ? AND name = ?", models.StorageProviderLocal, "Local").First(&local); err != nil || local.ID == 0 {
				return ErrStorageConnectionUnavailable
			}
			storageConnectionID = local.ID
		}
		var user models.User
		if err := tx.Where("id = ?", input.UserID).LockForUpdate().First(&user); err != nil {
			return err
		}
		if user.Status != "active" {
			return ErrAccountUnavailable
		}

		exists, err := tx.Model(&models.UploadSession{}).
			Where("user_id = ? AND idempotency_key = ?", input.UserID, input.IdempotencyKey).Exists()
		if err != nil {
			return err
		}
		if exists {
			var session models.UploadSession
			if err := tx.Where("user_id = ? AND idempotency_key = ?", input.UserID, input.IdempotencyKey).First(&session); err != nil {
				return err
			}
			if session.SHA256 != input.SHA256 || session.SizeBytes != input.SizeBytes {
				return ErrIdempotencyConflict
			}
			reservation, err = loadReservation(tx, session)
			if err != nil {
				return err
			}
			switch session.Status {
			case "ready":
				reservation.Status = "ready"
				return nil
			case "processing":
				reservation.Status = "processing"
				reservation.InProgress = true
				return nil
			case "failed":
				if err := checkUploadQuota(tx, input, totalObjectBytes(objects), session.ID); err != nil {
					return err
				}
				if _, err := tx.Model(&models.UploadSession{}).Where("id = ?", session.ID).Update(map[string]any{
					"status": "processing", "reserved_bytes": totalObjectBytes(objects), "error_code": "",
				}); err != nil {
					return err
				}
				if _, err := tx.Model(&models.MediaAsset{}).Where("id = ? AND user_id = ?", session.MediaAssetID, input.UserID).Update("status", "processing"); err != nil {
					return err
				}
				for _, objectKey := range reservation.ObjectKeys {
					if _, err := tx.Model(&models.StorageObject{}).Where("object_key = ?", objectKey).Update("status", "pending"); err != nil {
						return err
					}
				}
				if _, err := tx.Model(&models.MediaVariant{}).Where("media_asset_id = ?", session.MediaAssetID).Update("status", "pending"); err != nil {
					return err
				}
				reservation.Status = "processing"
				return nil
			default:
				return ErrUploadInProgress
			}
		}

		if err := checkUploadQuota(tx, input, totalObjectBytes(objects), 0); err != nil {
			return err
		}
		asset := models.MediaAsset{
			UserID: input.UserID, OriginalName: input.OriginalName,
			ContentType: input.ContentType, Format: input.Format,
			SizeBytes: input.SizeBytes, SHA256: input.SHA256,
			Width: input.Width, Height: input.Height, Status: "processing",
			Visibility: DefaultVisibility, ModerationStatus: DefaultModerationStatus,
		}
		if err := tx.Create(&asset); err != nil {
			return err
		}
		session := models.UploadSession{
			UserID: input.UserID, MediaAssetID: asset.ID, IdempotencyKey: input.IdempotencyKey,
			SHA256: input.SHA256, SizeBytes: input.SizeBytes,
			ReservedBytes: totalObjectBytes(objects), Status: "processing",
		}
		if err := tx.Create(&session); err != nil {
			return err
		}
		objectKeys := make(map[string]string, len(objects))
		for _, prepared := range objects {
			object := models.StorageObject{
				StorageConnectionID: storageConnectionID, ObjectKey: prepared.Key,
				ContentType: prepared.ContentType, SizeBytes: prepared.SizeBytes,
				SHA256: prepared.SHA256, Status: "pending",
			}
			if err := tx.Create(&object); err != nil {
				return err
			}
			variant := models.MediaVariant{
				MediaAssetID: asset.ID, StorageObjectID: object.ID,
				Name: prepared.Name, Width: prepared.Width, Height: prepared.Height, Status: "pending",
			}
			if err := tx.Create(&variant); err != nil {
				return err
			}
			objectKeys[prepared.Name] = prepared.Key
		}
		reservation = UploadReservation{
			SessionID: session.ID, MediaID: asset.ID, Status: "processing", ObjectKeys: objectKeys,
		}
		return nil
	})
	return reservation, err
}

func (r *DatabaseRepository) CompleteUpload(_ context.Context, sessionID uint) error {
	return facades.Orm().Transaction(func(tx orm.Query) error {
		var session models.UploadSession
		if err := tx.Where("id = ?", sessionID).First(&session); err != nil {
			return err
		}
		var user models.User
		if err := tx.Where("id = ?", session.UserID).LockForUpdate().First(&user); err != nil {
			return err
		}
		if session.Status == "ready" {
			return nil
		}
		if session.Status != "processing" {
			return ErrUploadInProgress
		}
		var variants []models.MediaVariant
		if err := tx.Where("media_asset_id = ?", session.MediaAssetID).Get(&variants); err != nil {
			return err
		}
		for _, variant := range variants {
			if _, err := tx.Model(&models.StorageObject{}).Where("id = ?", variant.StorageObjectID).Update("status", "ready"); err != nil {
				return err
			}
		}
		if _, err := tx.Model(&models.MediaVariant{}).Where("media_asset_id = ?", session.MediaAssetID).Update("status", "ready"); err != nil {
			return err
		}
		if _, err := tx.Model(&models.MediaAsset{}).Where("id = ?", session.MediaAssetID).Update("status", "ready"); err != nil {
			return err
		}
		if _, err := quota.RecordUsageWithQuery(tx, quota.UsageRecord{
			UserID: session.UserID, ResourceType: "storage", Delta: session.ReservedBytes,
			SourceType: "upload", SourceID: fmt.Sprint(session.ID), PeriodKey: "lifetime",
		}); err != nil {
			return err
		}
		completedAt := time.Now().UTC()
		reservedAt := time.Time{}
		if session.CreatedAt != nil && !session.CreatedAt.IsNil() {
			reservedAt = session.CreatedAt.StdTime()
		}
		if _, err := quota.RecordUsageWithQuery(tx, quota.UsageRecord{
			UserID: session.UserID, ResourceType: "upload", Delta: 1,
			SourceType: "upload", SourceID: fmt.Sprint(session.ID), PeriodKey: uploadUsagePeriodKey(reservedAt, completedAt),
		}); err != nil {
			return err
		}
		if _, err := quota.RecordUsageWithQuery(tx, quota.UsageRecord{
			UserID: session.UserID, ResourceType: "transform", Delta: 1,
			SourceType: "upload", SourceID: fmt.Sprint(session.ID), PeriodKey: uploadUsagePeriodKey(reservedAt, completedAt),
		}); err != nil {
			return err
		}
		if _, err := tx.Model(&models.UploadSession{}).Where("id = ?", session.ID).Update(map[string]any{
			"status": "ready", "reserved_bytes": 0, "error_code": "",
		}); err != nil {
			return err
		}
		return nil
	})
}

func (r *DatabaseRepository) FailUpload(_ context.Context, sessionID uint, errorCode string) error {
	return facades.Orm().Transaction(func(tx orm.Query) error {
		var session models.UploadSession
		if err := tx.Where("id = ?", sessionID).First(&session); err != nil {
			return err
		}
		if session.Status != "processing" {
			return nil
		}
		if _, err := tx.Model(&models.UploadSession{}).Where("id = ?", session.ID).Update(map[string]any{
			"status": "failed", "reserved_bytes": 0, "error_code": errorCode,
		}); err != nil {
			return err
		}
		if _, err := tx.Model(&models.MediaAsset{}).Where("id = ?", session.MediaAssetID).Update("status", "failed"); err != nil {
			return err
		}
		var variants []models.MediaVariant
		if err := tx.Where("media_asset_id = ?", session.MediaAssetID).Get(&variants); err != nil {
			return err
		}
		for _, variant := range variants {
			if _, err := tx.Model(&models.StorageObject{}).Where("id = ?", variant.StorageObjectID).Update("status", "failed"); err != nil {
				return err
			}
		}
		_, err := tx.Model(&models.MediaVariant{}).Where("media_asset_id = ?", session.MediaAssetID).Update("status", "failed")
		return err
	})
}

func (r *DatabaseRepository) GetUploadStatus(_ context.Context, userID, sessionID uint) (UploadOutcome, error) {
	query := facades.Orm().Query().Model(&models.UploadSession{}).Where("id = ? AND user_id = ?", sessionID, userID)
	exists, err := query.Exists()
	if err != nil {
		return UploadOutcome{}, err
	}
	if !exists {
		return UploadOutcome{}, ErrMediaNotFound
	}
	var session models.UploadSession
	if err := query.First(&session); err != nil {
		return UploadOutcome{}, err
	}
	var asset models.MediaAsset
	if err := facades.Orm().Query().Where("id = ? AND user_id = ?", session.MediaAssetID, userID).First(&asset); err != nil {
		return UploadOutcome{}, err
	}
	return UploadOutcome{
		SessionID: session.ID, MediaID: asset.ID, Status: session.Status,
		Name: asset.OriginalName, ContentType: asset.ContentType,
		SizeBytes: asset.SizeBytes, Width: asset.Width, Height: asset.Height,
		ErrorCode: session.ErrorCode,
	}, nil
}

func (r *DatabaseRepository) ClaimStaleUpload(_ context.Context, userID, sessionID uint, now time.Time) (recovery UploadRecovery, err error) {
	err = facades.Orm().Transaction(func(tx orm.Query) error {
		var user models.User
		if err := tx.Where("id = ?", userID).LockForUpdate().First(&user); err != nil {
			return err
		}
		if user.Status != "active" {
			return ErrAccountUnavailable
		}
		exists, err := tx.Model(&models.UploadSession{}).Where("id = ? AND user_id = ?", sessionID, userID).Exists()
		if err != nil {
			return err
		}
		if !exists {
			return ErrMediaNotFound
		}
		var session models.UploadSession
		if err := tx.Where("id = ? AND user_id = ?", sessionID, userID).First(&session); err != nil {
			return err
		}
		var asset models.MediaAsset
		if err := tx.Where("id = ? AND user_id = ?", session.MediaAssetID, userID).First(&asset); err != nil {
			return err
		}
		recovery.Outcome = UploadOutcome{
			SessionID: session.ID, MediaID: asset.ID, Status: session.Status,
			Name: asset.OriginalName, ContentType: asset.ContentType,
			SizeBytes: asset.SizeBytes, Width: asset.Width, Height: asset.Height,
			ErrorCode: session.ErrorCode,
		}
		switch session.Status {
		case "ready", "failed":
			return nil
		case "processing":
			var updatedAt time.Time
			if session.UpdatedAt != nil && !session.UpdatedAt.IsNil() {
				updatedAt = session.UpdatedAt.StdTime()
			}
			if !uploadRecoveryLeaseExpired(session.Status, updatedAt, now) {
				return ErrUploadInProgress
			}
			if _, err := tx.Model(&models.UploadSession{}).Where("id = ? AND user_id = ? AND status = ?", sessionID, userID, "processing").
				Update("updated_at", now); err != nil {
				return err
			}
			recovery.Reservation, err = loadReservation(tx, session)
			if err != nil {
				return err
			}
			var variants []models.MediaVariant
			if err := tx.Where("media_asset_id = ?", session.MediaAssetID).Get(&variants); err != nil {
				return err
			}
			recovery.Objects = make([]PreparedObject, 0, len(variants))
			for _, variant := range variants {
				var object models.StorageObject
				if err := tx.Find(&object, variant.StorageObjectID); err != nil {
					return err
				}
				recovery.Objects = append(recovery.Objects, PreparedObject{
					Name: variant.Name, Key: object.ObjectKey, ContentType: object.ContentType,
					SHA256: object.SHA256, SizeBytes: object.SizeBytes,
					Width: variant.Width, Height: variant.Height,
				})
			}
			recovery.Claimed = true
			return nil
		default:
			return ErrUploadInProgress
		}
	})
	return recovery, err
}

func checkUploadQuota(tx orm.Query, input UploadMetadata, requestedStorageBytes int64, currentSessionID uint) error {
	administrator, err := rbacservices.IsAdministratorWithQuery(tx, input.UserID)
	if err != nil {
		return err
	}
	if administrator {
		// Administrators still produce normal media and usage records, but their
		// own account is not blocked by a member plan's quota snapshot.
		return nil
	}
	subscription, err := planservices.EnsureActiveSubscriptionWithQuery(tx, input.UserID)
	if err != nil {
		return ErrSubscriptionUnavailable
	}
	entitlement, err := quota.ParseEntitlementJSON(subscription.EntitlementSnapshotJSON)
	if err != nil {
		return ErrSubscriptionUnavailable
	}
	now := time.Now().UTC()
	if entitlement.DailyUploads > 0 {
		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		query := tx.Model(&models.UploadSession{}).
			Where("user_id = ? AND created_at >= ? AND (status = ? OR status = ?)", input.UserID, startOfDay, "processing", "ready")
		if currentSessionID > 0 {
			query = query.Where("id <> ?", currentSessionID)
		}
		count, err := query.Count()
		if err != nil {
			return err
		}
		if err := quota.CheckUploadLimit(count, entitlement.DailyUploads); errors.Is(err, quota.ErrQuotaExceeded) {
			return ErrDailyUploadLimit
		} else if err != nil {
			return err
		}
	}
	if CountsTowardMonthlyAPIUploads(input.Channel) {
		if err := quota.CheckMonthlyAPIUploadsWithQuery(tx, input.UserID, now, entitlement.MonthlyAPIUploads); errors.Is(err, quota.ErrQuotaExceeded) {
			return ErrMonthlyAPIUploadLimit
		} else if err != nil {
			return err
		}
	}
	if err := quota.CheckMonthlyTransformsWithQuery(tx, input.UserID, now, entitlement.TransformCount); errors.Is(err, quota.ErrQuotaExceeded) {
		return ErrMonthlyTransformLimit
	} else if err != nil {
		return err
	}
	if _, err := quota.ReserveStorage(quota.UsageState{}, entitlement, input.SizeBytes); err != nil {
		return err
	}

	var used sql.NullInt64
	if err := tx.Model(&models.UsageLedger{}).Where("user_id = ? AND (resource_type = ? OR resource_type = ?)", input.UserID, "storage", "storage_bytes").Sum("delta", &used); err != nil {
		return err
	}
	usedValue := nullableSumInt64(used)
	if usedValue < 0 {
		usedValue = 0
	}
	var reserved sql.NullInt64
	pendingQuery := tx.Model(&models.UploadSession{}).Where("user_id = ? AND status = ?", input.UserID, "processing")
	if currentSessionID > 0 {
		pendingQuery = pendingQuery.Where("id <> ?", currentSessionID)
	}
	if err := pendingQuery.Sum("reserved_bytes", &reserved); err != nil {
		return err
	}
	reservedValue := nullableSumInt64(reserved)
	storageEntitlement := entitlement
	storageEntitlement.MaxFileBytes = 0
	if requestedStorageBytes <= 0 {
		requestedStorageBytes = input.SizeBytes
	}
	_, err = quota.ReserveStorage(quota.UsageState{StorageBytes: usedValue, ReservedBytes: reservedValue}, storageEntitlement, requestedStorageBytes)
	return err
}

func nullableSumInt64(value sql.NullInt64) int64 {
	if !value.Valid {
		return 0
	}
	return value.Int64
}

func totalObjectBytes(objects []PreparedObject) int64 {
	var total int64
	for _, object := range objects {
		total += object.SizeBytes
	}
	return total
}

func uploadUsagePeriodKey(reservedAt, completedAt time.Time) string {
	if reservedAt.IsZero() {
		reservedAt = completedAt
	}
	return reservedAt.UTC().Format("2006-01")
}

func loadReservation(tx orm.Query, session models.UploadSession) (UploadReservation, error) {
	var variants []models.MediaVariant
	if err := tx.Where("media_asset_id = ?", session.MediaAssetID).Get(&variants); err != nil {
		return UploadReservation{}, err
	}
	objectKeys := make(map[string]string, len(variants))
	for _, variant := range variants {
		var object models.StorageObject
		if err := tx.Find(&object, variant.StorageObjectID); err != nil {
			return UploadReservation{}, err
		}
		objectKeys[variant.Name] = object.ObjectKey
	}
	return UploadReservation{
		SessionID: session.ID, MediaID: session.MediaAssetID,
		Status: session.Status, ObjectKeys: objectKeys,
	}, nil
}

type MediaListItem struct {
	Asset    models.MediaAsset
	Variants []MediaVariantObject
}

type MediaVariantObject struct {
	Variant     models.MediaVariant
	Object      models.StorageObject
	OwnerUserID uint
}

type MediaLibraryRepository interface {
	ListOwned(ctx context.Context, userID uint, trash bool, search string, page, perPage int) ([]MediaListItem, int64, error)
	FindOwned(ctx context.Context, userID, mediaID uint, includeDeleted bool) (models.MediaAsset, error)
	FindOwnedVariant(ctx context.Context, userID, mediaID uint, name string) (MediaVariantObject, error)
	UpdateVisibility(ctx context.Context, userID, mediaID uint, visibility string) (models.MediaAsset, error)
	SoftDelete(ctx context.Context, userID, mediaID uint) (models.MediaAsset, error)
	Restore(ctx context.Context, userID, mediaID uint) (models.MediaAsset, error)
}

func (r *DatabaseRepository) UpdateExpiry(_ context.Context, userID, mediaID uint, expiresAt *time.Time) (asset models.MediaAsset, err error) {
	if expiresAt != nil && !expiresAt.After(time.Now().UTC()) {
		return models.MediaAsset{}, ErrInvalidMediaExpiry
	}
	if err = facades.Orm().Query().Where("id = ? AND user_id = ? AND status = ?", mediaID, userID, "ready").First(&asset); err != nil {
		return models.MediaAsset{}, ErrMediaNotFound
	}
	if _, err = facades.Orm().Query().Model(&models.MediaAsset{}).Where("id = ? AND user_id = ? AND status = ?", mediaID, userID, "ready").Update(map[string]any{"expires_at": expiresAt, "expiry_notified_at": nil, "updated_at": time.Now().UTC()}); err != nil {
		return models.MediaAsset{}, err
	}
	asset.ExpiresAt = expiresAt
	asset.ExpiryNotifiedAt = nil
	return asset, nil
}

func (r *DatabaseRepository) UpdateVisibility(_ context.Context, userID, mediaID uint, visibility string) (asset models.MediaAsset, err error) {
	if !ValidVisibility(visibility) {
		return models.MediaAsset{}, ErrInvalidVisibility
	}
	if err = facades.Orm().Query().Where("id = ? AND user_id = ? AND status = ?", mediaID, userID, "ready").First(&asset); err != nil {
		return models.MediaAsset{}, ErrMediaNotFound
	}
	if _, err = facades.Orm().Query().Model(&models.MediaAsset{}).Where("id = ? AND user_id = ? AND status = ?", mediaID, userID, "ready").Update(map[string]any{"visibility": visibility, "updated_at": time.Now().UTC()}); err != nil {
		return models.MediaAsset{}, err
	}
	asset.Visibility = visibility
	return asset, nil
}

func (r *DatabaseRepository) AssignFolder(_ context.Context, userID, mediaID uint, folderID *uint) (asset models.MediaAsset, err error) {
	err = facades.Orm().Transaction(func(tx orm.Query) error {
		if err := tx.Where("id = ? AND user_id = ? AND status = ?", mediaID, userID, "ready").LockForUpdate().First(&asset); err != nil {
			return ErrMediaNotFound
		}
		if folderID != nil {
			exists, err := tx.Table("folders").Where("id = ? AND user_id = ?", *folderID, userID).Exists()
			if err != nil {
				return err
			}
			if !exists {
				return ErrFolderNotFound
			}
		}
		var folderValue any
		if folderID != nil {
			folderValue = *folderID
		}
		result, err := tx.Model(&models.MediaAsset{}).Where("id = ? AND user_id = ? AND status = ?", mediaID, userID, "ready").Update(map[string]any{"folder_id": folderValue})
		if err != nil {
			return err
		}
		if result.RowsAffected == 0 {
			return ErrMediaNotFound
		}
		asset.FolderID = folderID
		return nil
	})
	return asset, err
}

func (r *DatabaseRepository) ListOwned(_ context.Context, userID uint, trash bool, search string, page, perPage int) ([]MediaListItem, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	query := facades.Orm().Query().Where("user_id = ?", userID)
	if trash {
		query = query.WhereIn("status", mediaTrashStatuses())
	} else {
		query = query.Where("status = ?", "ready")
	}
	if search = strings.TrimSpace(search); search != "" {
		query = query.Where("LOWER(original_name) LIKE LOWER(?)", "%"+search+"%")
	}
	var assets []models.MediaAsset
	var total int64
	if err := query.OrderBy("id", "desc").Paginate(page, perPage, &assets, &total); err != nil {
		return nil, 0, err
	}
	items := make([]MediaListItem, 0, len(assets))
	assetIDs := make([]any, 0, len(assets))
	for _, asset := range assets {
		assetIDs = append(assetIDs, asset.ID)
		items = append(items, MediaListItem{Asset: asset})
	}
	if len(assetIDs) == 0 {
		return items, total, nil
	}
	var variants []models.MediaVariant
	if err := facades.Orm().Query().WhereIn("media_asset_id", assetIDs).Where("status = ? AND name = ?", "ready", "original").Get(&variants); err != nil {
		return nil, 0, err
	}
	objectIDs := make([]any, 0, len(variants))
	for _, variant := range variants {
		objectIDs = append(objectIDs, variant.StorageObjectID)
	}
	objectsByID := make(map[uint]models.StorageObject, len(objectIDs))
	if len(objectIDs) > 0 {
		var objects []models.StorageObject
		if err := facades.Orm().Query().WhereIn("id", objectIDs).Where("status = ?", "ready").Get(&objects); err != nil {
			return nil, 0, err
		}
		for _, object := range objects {
			objectsByID[object.ID] = object
		}
	}
	itemsByID := make(map[uint]*MediaListItem, len(items))
	for index := range items {
		itemsByID[items[index].Asset.ID] = &items[index]
	}
	for _, variant := range variants {
		item := itemsByID[variant.MediaAssetID]
		object, ok := objectsByID[variant.StorageObjectID]
		if item == nil || !ok {
			continue
		}
		item.Variants = append(item.Variants, MediaVariantObject{Variant: variant, Object: object})
	}
	return items, total, nil
}

func (r *DatabaseRepository) FindOwned(_ context.Context, userID, mediaID uint, includeDeleted bool) (models.MediaAsset, error) {
	query := facades.Orm().Query().Where("id = ? AND user_id = ?", mediaID, userID)
	if !includeDeleted {
		query = query.Where("status = ?", "ready")
	}
	var asset models.MediaAsset
	if err := query.First(&asset); err != nil {
		return models.MediaAsset{}, ErrMediaNotFound
	}
	return asset, nil
}

func (r *DatabaseRepository) FindOwnedVariant(ctx context.Context, userID, mediaID uint, name string) (MediaVariantObject, error) {
	if _, err := r.FindOwned(ctx, userID, mediaID, false); err != nil {
		return MediaVariantObject{}, err
	}
	var variant models.MediaVariant
	if err := facades.Orm().Query().Where("media_asset_id = ? AND name = ? AND status = ?", mediaID, name, "ready").First(&variant); err != nil {
		return MediaVariantObject{}, ErrVariantNotFound
	}
	var object models.StorageObject
	if err := facades.Orm().Query().Find(&object, variant.StorageObjectID); err != nil {
		return MediaVariantObject{}, err
	}
	return MediaVariantObject{Variant: variant, Object: object, OwnerUserID: userID}, nil
}

func (r *DatabaseRepository) SoftDelete(ctx context.Context, userID, mediaID uint) (models.MediaAsset, error) {
	asset, err := r.FindOwned(ctx, userID, mediaID, true)
	if err != nil {
		return models.MediaAsset{}, err
	}
	if asset.Status == "deleted" {
		return asset, nil
	}
	if asset.Status != "ready" {
		return models.MediaAsset{}, ErrMediaNotFound
	}
	now := time.Now().UTC()
	if _, err := facades.Orm().Query().Model(&models.MediaAsset{}).Where("id = ? AND user_id = ?", mediaID, userID).
		Update(map[string]any{"status": "deleted", "deleted_at": now}); err != nil {
		return models.MediaAsset{}, err
	}
	asset.Status, asset.DeletedAt = "deleted", &now
	return asset, nil
}

func (r *DatabaseRepository) Restore(_ context.Context, userID, mediaID uint) (asset models.MediaAsset, err error) {
	err = facades.Orm().Transaction(func(tx orm.Query) error {
		var user models.User
		if err := tx.Where("id = ?", userID).LockForUpdate().First(&user); err != nil {
			return ErrMediaNotFound
		}
		if err := tx.Where("id = ? AND user_id = ? AND status = ?", mediaID, userID, "deleted").LockForUpdate().First(&asset); err != nil {
			return ErrMediaNotFound
		}
		if !canRestoreMediaStatus(asset.Status) {
			return ErrMediaNotFound
		}
		result, err := tx.Model(&models.MediaAsset{}).Where("id = ? AND user_id = ? AND status = ?", mediaID, userID, "deleted").
			Update(map[string]any{"status": "ready", "deleted_at": nil})
		if err != nil {
			return err
		}
		if result.RowsAffected == 0 {
			return ErrMediaNotFound
		}
		asset.Status, asset.DeletedAt = "ready", nil
		return nil
	})
	return asset, err
}

func (r *DatabaseRepository) PreparePermanentDelete(_ context.Context, userID, mediaID uint) (asset models.MediaAsset, objects []models.StorageObject, err error) {
	err = facades.Orm().Transaction(func(tx orm.Query) error {
		asset, objects, err = PreparePermanentDeleteWithQuery(tx, userID, mediaID)
		return err
	})
	return asset, objects, err
}

func (r *DatabaseRepository) FinalizePermanentDelete(_ context.Context, userID, mediaID uint) error {
	return facades.Orm().Transaction(func(tx orm.Query) error {
		return FinalizePermanentDeleteWithQuery(tx, userID, mediaID)
	})
}

func PreparePermanentDeleteWithQuery(tx orm.Query, userID, mediaID uint) (asset models.MediaAsset, objects []models.StorageObject, err error) {
	var user models.User
	if err := tx.Where("id = ?", userID).LockForUpdate().First(&user); err != nil {
		return models.MediaAsset{}, nil, err
	}
	if err := tx.Where("id = ? AND user_id = ?", mediaID, userID).LockForUpdate().First(&asset); err != nil {
		return models.MediaAsset{}, nil, ErrMediaNotFound
	}
	if asset.Status == "physically_deleted" {
		return asset, nil, nil
	}
	if asset.Status != "deleted" && asset.Status != "cleanup_pending" {
		return models.MediaAsset{}, nil, ErrMediaNotFound
	}
	var variants []models.MediaVariant
	if err := tx.Where("media_asset_id = ?", mediaID).Get(&variants); err != nil {
		return models.MediaAsset{}, nil, err
	}
	objectIDs := make([]any, 0, len(variants))
	seen := make(map[uint]struct{}, len(variants))
	for _, variant := range variants {
		if _, exists := seen[variant.StorageObjectID]; exists {
			continue
		}
		seen[variant.StorageObjectID] = struct{}{}
		objectIDs = append(objectIDs, variant.StorageObjectID)
	}
	if len(objectIDs) == 0 {
		return models.MediaAsset{}, nil, ErrMediaNotFound
	}
	if err := tx.WhereIn("id", objectIDs).Get(&objects); err != nil {
		return models.MediaAsset{}, nil, err
	}
	if len(objects) != len(objectIDs) {
		return models.MediaAsset{}, nil, errors.New("media storage reference is incomplete")
	}
	for _, object := range objects {
		shared, err := tx.Model(&models.MediaVariant{}).
			Where("storage_object_id = ? AND media_asset_id <> ?", object.ID, mediaID).Exists()
		if err != nil {
			return models.MediaAsset{}, nil, err
		}
		if shared {
			return models.MediaAsset{}, nil, ErrMediaSharedStorage
		}
		if object.Status != "ready" && object.Status != "deleted" {
			return models.MediaAsset{}, nil, errors.New("media storage object is not ready for permanent deletion")
		}
	}
	if asset.Status == "deleted" {
		if _, err := tx.Model(&models.MediaAsset{}).Where("id = ? AND user_id = ? AND status = ?", mediaID, userID, "deleted").
			Update("status", "cleanup_pending"); err != nil {
			return models.MediaAsset{}, nil, err
		}
		asset.Status = "cleanup_pending"
	}
	return asset, objects, nil
}

func FinalizePermanentDeleteWithQuery(tx orm.Query, userID, mediaID uint) error {
	var user models.User
	if err := tx.Where("id = ?", userID).LockForUpdate().First(&user); err != nil {
		return err
	}
	var asset models.MediaAsset
	if err := tx.Where("id = ? AND user_id = ?", mediaID, userID).LockForUpdate().First(&asset); err != nil {
		return ErrMediaNotFound
	}
	if asset.Status == "physically_deleted" {
		return nil
	}
	if asset.Status != "cleanup_pending" {
		return ErrMediaNotFound
	}
	var variants []models.MediaVariant
	if err := tx.Where("media_asset_id = ?", mediaID).Get(&variants); err != nil {
		return err
	}
	if len(variants) == 0 {
		return ErrMediaNotFound
	}
	seen := make(map[uint]struct{}, len(variants))
	var releasedBytes int64
	for _, variant := range variants {
		if _, exists := seen[variant.StorageObjectID]; exists {
			continue
		}
		seen[variant.StorageObjectID] = struct{}{}
		var object models.StorageObject
		if err := tx.Where("id = ?", variant.StorageObjectID).First(&object); err != nil {
			return errors.New("media storage reference is incomplete")
		}
		switch object.Status {
		case "ready":
			if object.SizeBytes < 0 || releasedBytes > math.MaxInt64-object.SizeBytes {
				return errors.New("media storage size is invalid")
			}
			releasedBytes += object.SizeBytes
		case "deleted":
		default:
			return errors.New("media storage object was not deleted")
		}
		if _, err := tx.Model(&models.StorageObject{}).Where("id = ?", object.ID).Update("status", "deleted"); err != nil {
			return err
		}
	}
	if _, err := tx.Model(&models.MediaVariant{}).Where("media_asset_id = ?", mediaID).Update("status", "deleted"); err != nil {
		return err
	}
	if releasedBytes > 0 {
		if _, err := quota.RecordUsageWithQuery(tx, quota.UsageRecord{
			UserID: userID, ResourceType: "storage", Delta: -releasedBytes,
			SourceType: "delete", SourceID: fmt.Sprintf("media:%d", mediaID), PeriodKey: "lifetime",
		}); err != nil {
			return err
		}
	}
	_, err := tx.Model(&models.MediaAsset{}).Where("id = ? AND user_id = ? AND status = ?", mediaID, userID, "cleanup_pending").
		Update("status", "physically_deleted")
	return err
}
