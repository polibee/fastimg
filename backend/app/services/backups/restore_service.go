package backups

import (
	"archive/tar"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/goravel/framework/support/path"
	"github.com/klauspost/compress/zstd"

	"goravel/app/facades"
	"goravel/app/models"
	auditservices "goravel/app/services/audit"
	settingsservices "goravel/app/services/settings"
	storageservices "goravel/app/services/storage"
)

type RestoreRequest struct {
	JobID        uint   `json:"id"`
	Mode         string `json:"mode"`
	Confirmation string `json:"confirmation"`
}

type RestoreService struct {
	backups *BackupService
	tool    DatabaseTool
}

func NewRestoreService() *RestoreService {
	return &RestoreService{backups: NewBackupService(), tool: NewCommandDatabaseTool()}
}

func NewRestoreServiceWithDependencies(backups *BackupService, tool DatabaseTool) *RestoreService {
	if backups == nil {
		backups = NewBackupService()
	}
	if tool == nil {
		tool = NewCommandDatabaseTool()
	}
	return &RestoreService{backups: backups, tool: tool}
}

func (s *RestoreService) Restore(ctx context.Context, request RestoreRequest) (models.BackupJob, error) {
	if err := validateRestoreRequest(request); err != nil {
		return models.BackupJob{}, err
	}
	job, err := s.backups.Find(request.JobID)
	if err != nil {
		return models.BackupJob{}, err
	}
	if job.Kind != "restore" || job.Status != BackupStatusValidated || job.StoragePath == "" {
		return models.BackupJob{}, errors.New("backup restore job is not validated")
	}
	if err := ensureNewServerTarget(); err != nil {
		return models.BackupJob{}, err
	}
	if _, err := facades.Orm().Query().Table("backup_jobs").Where("id = ?", job.ID).Update(map[string]any{"status": BackupStatusRestoring, "updated_at": time.Now().UTC()}); err != nil {
		return models.BackupJob{}, err
	}
	if err := s.restoreArchive(ctx, job); err != nil {
		now := time.Now().UTC()
		_, _ = facades.Orm().Query().Table("backup_jobs").Where("id = ?", job.ID).Update(map[string]any{"status": BackupStatusRestoreFailed, "error_code": restoreErrorCode(err), "error_message": "restore failed", "completed_at": now, "updated_at": now})
		_ = auditservices.NewAuditService().Record(job.CreatedBy, "backup.restore.failed", map[string]any{"backup_id": job.ID, "error_code": restoreErrorCode(err)})
		return models.BackupJob{}, err
	}
	now := time.Now().UTC()
	_, err = facades.Orm().Query().Table("backup_jobs").Where("id = ?", job.ID).Update(map[string]any{"status": BackupStatusRestored, "completed_at": now, "updated_at": now, "error_code": "", "error_message": ""})
	if err != nil {
		return models.BackupJob{}, err
	}
	job.Status, job.CompletedAt = BackupStatusRestored, &now
	_ = auditservices.NewAuditService().Record(job.CreatedBy, "backup.restore.success", map[string]any{"backup_id": job.ID})
	return job, nil
}

func ensureNewServerTarget() error {
	if facades.Schema().HasTable("media_assets") {
		count, err := facades.Orm().Query().Table("media_assets").Count()
		if err != nil {
			return err
		}
		if count > 0 {
			return ErrRestoreTargetNotEmpty
		}
	}
	if facades.Schema().HasTable("users") {
		count, err := facades.Orm().Query().Table("users").Count()
		if err != nil {
			return err
		}
		// A freshly installed target may already contain the generated admin.
		if count > 1 {
			return ErrRestoreTargetNotEmpty
		}
	}
	return nil
}

func (s *RestoreService) restoreArchive(ctx context.Context, job models.BackupJob) error {
	archive, err := os.Open(job.StoragePath)
	if err != nil {
		return err
	}
	defer archive.Close()
	info, err := archive.Stat()
	if err != nil {
		return err
	}
	entries, err := readArchiveEntries(ctx, archive, info.Size(), DefaultArchiveLimits())
	if err != nil {
		return err
	}
	tempDir, err := os.MkdirTemp(backupRoot(), ".restore-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)
	dumpPath := filepath.Join(tempDir, "database.dump")
	if err := os.WriteFile(dumpPath, entries["database.dump"], 0o600); err != nil {
		return err
	}
	if err := s.tool.Restore(ctx, dumpPath); err != nil {
		return err
	}
	if err := restoreSettings(entries["settings.json"]); err != nil {
		return err
	}
	return restoreStorageEntries(ctx, entries)
}

func restoreSettings(body []byte) error {
	var settings SettingsExport
	if err := json.Unmarshal(body, &settings); err != nil {
		return err
	}
	service := settingsservices.NewSettingService()
	for key, value := range settings.Values {
		if isBackupSecretKey(key) {
			continue
		}
		if _, err := service.Upsert(key, value, "string", "restored", "Restored from FastImg backup"); err != nil {
			return err
		}
	}
	return nil
}

func restoreStorageEntries(ctx context.Context, entries map[string][]byte) error {
	root := filepath.Clean(path.Storage("fastimg"))
	disk := facades.Storage().Disk("fastimg")
	registry := storageservices.NewRuntimeRegistry(disk)
	for name, body := range entries {
		if !strings.HasPrefix(name, "storage/media/") && !strings.HasPrefix(name, "storage/variants/") {
			continue
		}
		key := strings.TrimPrefix(name, "storage/media/")
		if key == name {
			key = strings.TrimPrefix(name, "storage/variants/")
		}
		if key == "" || strings.Contains(key, "\\") {
			return ErrArchiveUnsafeEntry
		}
		// A cloud-backed source must be restored through the configured target
		// provider. Backup archives deliberately exclude provider credentials, so
		// this fails closed until the administrator configures the target
		// connection instead of silently placing cloud media on local disk.
		if facades.Schema().HasTable("storage_objects") {
			var object models.StorageObject
			if err := facades.Orm().Query().Where("object_key = ?", key).First(&object); err == nil && object.StorageConnectionID > 0 {
				var connection models.StorageConnection
				if err := facades.Orm().Query().Find(&connection, object.StorageConnectionID); err != nil {
					return err
				}
				if connection.ProviderCode != models.StorageProviderLocal {
					provider, err := registry.ProviderFor(connection)
					if err != nil {
						return err
					}
					if err := provider.Put(ctx, key, object.ContentType, body); err != nil {
						return err
					}
					continue
				}
			}
		}
		destination := filepath.Join(root, filepath.FromSlash(key))
		if !isWithinRoot(root, destination) {
			return ErrArchiveUnsafeEntry
		}
		if err := os.MkdirAll(filepath.Dir(destination), 0o700); err != nil {
			return err
		}
		if err := os.WriteFile(destination, body, 0o600); err != nil {
			return err
		}
	}
	return nil
}

func isWithinRoot(root, target string) bool {
	rel, err := filepath.Rel(root, target)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) && !filepath.IsAbs(rel)
}

func restoreErrorCode(err error) string {
	switch {
	case errors.Is(err, ErrRestoreTargetNotEmpty):
		return "BACKUP_RESTORE_TARGET_NOT_EMPTY"
	case errors.Is(err, ErrDatabaseToolUnavailable):
		return "BACKUP_DATABASE_RESTORE_FAILED"
	case errors.Is(err, ErrArchiveChecksum):
		return "BACKUP_ARCHIVE_INVALID"
	default:
		return "BACKUP_RESTORE_FAILED"
	}
}

func readArchiveEntries(ctx context.Context, source io.ReaderAt, size int64, limits ArchiveLimits) (map[string][]byte, error) {
	if _, err := ValidateArchive(ctx, source, size, limits); err != nil {
		return nil, err
	}
	decoder, err := zstd.NewReader(io.NewSectionReader(source, 0, size))
	if err != nil {
		return nil, err
	}
	defer decoder.Close()
	reader := tar.NewReader(decoder)
	entries := make(map[string][]byte)
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		body, err := io.ReadAll(io.LimitReader(reader, limits.MaxFileBytes+1))
		if err != nil || int64(len(body)) != header.Size {
			return nil, ErrArchiveUnsafeEntry
		}
		entries[header.Name] = body
	}
	return entries, nil
}
