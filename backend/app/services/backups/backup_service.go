package backups

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/goravel/framework/support/path"

	"goravel/app/facades"
	"goravel/app/models"
	auditservices "goravel/app/services/audit"
	storageservices "goravel/app/services/storage"
)

const (
	BackupStatusQueued        = "queued"
	BackupStatusRunning       = "running"
	BackupStatusReady         = "ready"
	BackupStatusFailed        = "failed"
	BackupStatusValidated     = "validated"
	BackupStatusRestoring     = "restoring"
	BackupStatusRestored      = "restored"
	BackupStatusRestoreFailed = "restore_failed"
)

var ErrBackupActive = errors.New("a backup job is already running")

type BackupService struct {
	tool DatabaseTool
	mu   sync.Mutex
}

func NewBackupService() *BackupService { return &BackupService{tool: NewCommandDatabaseTool()} }

func NewBackupServiceWithTool(tool DatabaseTool) *BackupService {
	if tool == nil {
		tool = NewCommandDatabaseTool()
	}
	return &BackupService{tool: tool}
}

func (s *BackupService) Create(ctx context.Context, operatorID uint) (models.BackupJob, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var existing models.BackupJob
	if err := facades.Orm().Query().WhereIn("status", []any{BackupStatusQueued, BackupStatusRunning, BackupStatusRestoring}).First(&existing); err == nil && existing.ID > 0 {
		return models.BackupJob{}, ErrBackupActive
	}
	job := models.BackupJob{Kind: "site", Status: BackupStatusQueued, CreatedBy: operatorID}
	if err := facades.Orm().Query().Create(&job); err != nil {
		return models.BackupJob{}, err
	}
	_ = auditservices.NewAuditService().Record(operatorID, "backup.create", map[string]any{"backup_id": job.ID})
	// The HTTP request context ends as soon as the response is written. A
	// backup is a persisted job, so it must use an independent bounded worker
	// context instead of being cancelled with the request.
	go func() {
		workerCtx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
		defer cancel()
		s.run(workerCtx, job.ID)
	}()
	return job, nil
}

func (s *BackupService) run(ctx context.Context, id uint) {
	var job models.BackupJob
	if err := facades.Orm().Query().Find(&job, id); err != nil {
		return
	}
	now := time.Now().UTC()
	_, _ = facades.Orm().Query().Table("backup_jobs").Where("id = ?", id).Update(map[string]any{"status": BackupStatusRunning, "started_at": now, "updated_at": now})
	root := backupRoot()
	if err := os.MkdirAll(root, 0o700); err != nil {
		s.fail(id, "BACKUP_STORAGE_UNAVAILABLE", err)
		return
	}
	tempDir, err := os.MkdirTemp(root, ".job-")
	if err != nil {
		s.fail(id, "BACKUP_STORAGE_UNAVAILABLE", err)
		return
	}
	defer os.RemoveAll(tempDir)
	dumpPath := filepath.Join(tempDir, "database.dump")
	if err := s.tool.Dump(ctx, dumpPath); err != nil {
		s.fail(id, "BACKUP_DATABASE_DUMP_FAILED", err)
		return
	}
	dump, err := os.ReadFile(dumpPath)
	if err != nil {
		s.fail(id, "BACKUP_DATABASE_DUMP_FAILED", err)
		return
	}
	settings, err := exportDatabaseSettings()
	if err != nil {
		s.fail(id, "BACKUP_SETTINGS_EXPORT_FAILED", err)
		return
	}
	settingsBody, err := json.Marshal(settings)
	if err != nil {
		s.fail(id, "BACKUP_SETTINGS_EXPORT_FAILED", err)
		return
	}
	files := []ArchiveFile{{Name: "database.dump", Body: dump}, {Name: "settings.json", Body: settingsBody}}
	mediaFiles, err := snapshotMedia(ctx)
	if err != nil {
		s.fail(id, "BACKUP_MEDIA_EXPORT_FAILED", err)
		return
	}
	files = append(files, mediaFiles...)
	name := fmt.Sprintf("fastimg-backup-%s-%d.tar.zst", time.Now().UTC().Format("20060102-150405"), id)
	destination := filepath.Join(root, name)
	manifest, err := WriteArchive(ctx, destination, uint64(id), files, settings.ExcludedSecretKeys)
	if err != nil {
		s.fail(id, "BACKUP_ARCHIVE_WRITE_FAILED", err)
		return
	}
	info, err := os.Stat(destination)
	if err != nil {
		s.fail(id, "BACKUP_ARCHIVE_WRITE_FAILED", err)
		return
	}
	manifestBody, _ := json.Marshal(manifest)
	completed := time.Now().UTC()
	_, err = facades.Orm().Query().Table("backup_jobs").Where("id = ?", id).Update(map[string]any{
		"status": BackupStatusReady, "file_name": name, "storage_path": destination, "manifest_json": string(manifestBody),
		"size_bytes": info.Size(), "completed_at": completed, "updated_at": completed,
	})
	if err == nil {
		_ = auditservices.NewAuditService().Record(job.CreatedBy, "backup.ready", map[string]any{"backup_id": id, "size_bytes": info.Size()})
	}
}

func (s *BackupService) fail(id uint, code string, err error) {
	message := "backup job failed"
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "password") {
		message = err.Error()
	}
	now := time.Now().UTC()
	_, _ = facades.Orm().Query().Table("backup_jobs").Where("id = ?", id).Update(map[string]any{"status": BackupStatusFailed, "error_code": code, "error_message": message, "completed_at": now, "updated_at": now})
}

func backupRoot() string {
	if value := strings.TrimSpace(os.Getenv("FASTIMG_BACKUP_STORAGE_PATH")); value != "" {
		return filepath.Clean(value)
	}
	return path.Storage("backups")
}

func exportDatabaseSettings() (SettingsExport, error) {
	if !facades.Schema().HasTable("system_settings") {
		return SettingsExport{Values: map[string]string{}}, nil
	}
	var settings []models.SystemSetting
	if err := facades.Orm().Query().OrderBy("key").Get(&settings); err != nil {
		return SettingsExport{}, err
	}
	values := make([]SettingValue, 0, len(settings))
	for _, setting := range settings {
		values = append(values, SettingValue{Key: setting.Key, Value: setting.Value})
	}
	return ExportSettings(values), nil
}

func snapshotMedia(ctx context.Context) ([]ArchiveFile, error) {
	if !facades.Schema().HasTable("storage_objects") {
		return nil, nil
	}
	var objects []models.StorageObject
	if err := facades.Orm().Query().Where("status = ?", "ready").OrderBy("id").Get(&objects); err != nil {
		return nil, err
	}
	var connections []models.StorageConnection
	if facades.Schema().HasTable("storage_connections") {
		if err := facades.Orm().Query().Get(&connections); err != nil {
			return nil, err
		}
	}
	byID := make(map[uint]models.StorageConnection, len(connections))
	for _, connection := range connections {
		byID[connection.ID] = connection
	}
	disk := facades.Storage().Disk("fastimg")
	registry := storageservices.NewRuntimeRegistry(disk)
	providers := make(map[uint]storageservices.StorageProvider)
	files := make([]ArchiveFile, 0, len(objects))
	for _, object := range objects {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		provider := providers[object.StorageConnectionID]
		if provider == nil {
			connection := byID[object.StorageConnectionID]
			if connection.ID == 0 {
				return nil, errors.New("storage connection metadata missing")
			}
			var err error
			provider, err = registry.ProviderFor(connection)
			if err != nil {
				return nil, err
			}
			providers[object.StorageConnectionID] = provider
		}
		body, err := provider.Get(ctx, object.ObjectKey)
		if err != nil {
			return nil, err
		}
		files = append(files, ArchiveFile{Name: "storage/media/" + object.ObjectKey, Body: body})
	}
	return files, nil
}

func (s *BackupService) List(page, perPage int) ([]models.BackupJob, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	var jobs []models.BackupJob
	var total int64
	if err := facades.Orm().Query().OrderByDesc("id").Paginate(page, perPage, &jobs, &total); err != nil {
		return nil, 0, err
	}
	return jobs, total, nil
}

func (s *BackupService) Find(id uint) (models.BackupJob, error) {
	var job models.BackupJob
	if err := facades.Orm().Query().Find(&job, id); err != nil {
		return models.BackupJob{}, err
	}
	return job, nil
}

func (s *BackupService) Delete(operatorID, id uint) error {
	job, err := s.Find(id)
	if err != nil {
		return err
	}
	if job.StoragePath != "" {
		_ = os.Remove(job.StoragePath)
	}
	if _, err := facades.Orm().Query().Table("backup_jobs").Where("id = ?", id).Delete(); err != nil {
		return err
	}
	_ = auditservices.NewAuditService().Record(operatorID, "backup.delete", map[string]any{"backup_id": id})
	return nil
}

// ValidateFile validates a private uploaded archive and persists it as a
// restore job. The controller never exposes this path publicly.
func (s *BackupService) ValidateFile(ctx context.Context, operatorID uint, filename string, body []byte) (models.BackupJob, ValidationPreview, error) {
	if len(body) == 0 {
		return models.BackupJob{}, ValidationPreview{}, ErrArchiveTooLarge
	}
	root := backupRoot()
	if err := os.MkdirAll(root, 0o700); err != nil {
		return models.BackupJob{}, ValidationPreview{}, err
	}
	job := models.BackupJob{Kind: "restore", Status: BackupStatusValidated, FileName: filepath.Base(filename), CreatedBy: operatorID}
	if err := facades.Orm().Query().Create(&job); err != nil {
		return models.BackupJob{}, ValidationPreview{}, err
	}
	pathName := filepath.Join(root, fmt.Sprintf("restore-%d-upload.tar.zst", job.ID))
	if err := os.WriteFile(pathName, body, 0o600); err != nil {
		return models.BackupJob{}, ValidationPreview{}, err
	}
	file, err := os.Open(pathName)
	if err != nil {
		return models.BackupJob{}, ValidationPreview{}, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return models.BackupJob{}, ValidationPreview{}, err
	}
	preview, err := ValidateArchive(ctx, file, info.Size(), DefaultArchiveLimits())
	if err != nil {
		_ = os.Remove(pathName)
		_, _ = facades.Orm().Query().Table("backup_jobs").Where("id = ?", job.ID).Update(map[string]any{"status": BackupStatusFailed, "error_code": "BACKUP_ARCHIVE_INVALID", "error_message": "archive validation failed"})
		return models.BackupJob{}, ValidationPreview{}, err
	}
	manifestBody, _ := json.Marshal(preview)
	_, err = facades.Orm().Query().Table("backup_jobs").Where("id = ?", job.ID).Update(map[string]any{"storage_path": pathName, "size_bytes": info.Size(), "manifest_json": string(manifestBody)})
	if err != nil {
		return models.BackupJob{}, ValidationPreview{}, err
	}
	job.StoragePath, job.SizeBytes, job.ManifestJSON = pathName, info.Size(), string(manifestBody)
	_ = auditservices.NewAuditService().Record(operatorID, "backup.validate", map[string]any{"backup_id": job.ID})
	return job, preview, nil
}

func (s *BackupService) DownloadPath(job models.BackupJob) (string, error) {
	if job.Status != BackupStatusReady || job.StoragePath == "" {
		return "", errors.New("backup is not ready")
	}
	clean := filepath.Clean(job.StoragePath)
	root := backupRoot()
	if filepath.Dir(clean) != filepath.Clean(root) {
		return "", errors.New("backup path is invalid")
	}
	if _, err := os.Stat(clean); err != nil {
		return "", err
	}
	return clean, nil
}
