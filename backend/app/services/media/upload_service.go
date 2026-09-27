package media

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"strings"
	"time"
	"unicode"

	"goravel/app/facades"
	storageservices "goravel/app/services/storage"
)

var (
	ErrInvalidUploadInput = errors.New("invalid upload input")
	ErrStorageWriteFailed = errors.New("could not store uploaded image")
)

const uploadRecoveryGracePeriod = 10 * time.Minute

type UploadInput struct {
	UserID              uint
	Channel             UploadChannel
	OriginalName        string
	DeclaredContentType string
	IdempotencyKey      string
	Content             []byte
	WatermarkEnabled    bool
	WatermarkText       string
	WatermarkDomain     string
}

type UploadMetadata struct {
	UserID              uint
	StorageConnectionID uint
	Channel             UploadChannel
	OriginalName        string
	ContentType         string
	Format              string
	SHA256              string
	SizeBytes           int64
	Width               int64
	Height              int64
	IdempotencyKey      string
}

// UploadChannel identifies the entry point that consumed the upload.  The
// monthly API quota is intentionally scoped to Personal API Token uploads;
// browser/member uploads are governed by the daily upload and storage limits.
type UploadChannel string

const (
	UploadChannelMember UploadChannel = "member"
	UploadChannelAPI    UploadChannel = "api"
)

func CountsTowardMonthlyAPIUploads(channel UploadChannel) bool {
	return channel == UploadChannelAPI
}

type PreparedObject struct {
	Name        string
	Key         string
	ContentType string
	Content     []byte
	SHA256      string
	SizeBytes   int64
	Width       int64
	Height      int64
}

type UploadReservation struct {
	SessionID  uint
	MediaID    uint
	Status     string
	ObjectKeys map[string]string
	InProgress bool
}

type UploadRepository interface {
	CheckUploadAllowance(ctx context.Context, input UploadMetadata) error
	BeginUpload(ctx context.Context, input UploadMetadata, objects []PreparedObject) (UploadReservation, error)
	CompleteUpload(ctx context.Context, sessionID uint) error
	FailUpload(ctx context.Context, sessionID uint, errorCode string) error
	GetUploadStatus(ctx context.Context, userID, sessionID uint) (UploadOutcome, error)
	ClaimStaleUpload(ctx context.Context, userID, sessionID uint, now time.Time) (UploadRecovery, error)
}

type UploadRecovery struct {
	Outcome     UploadOutcome
	Reservation UploadReservation
	Objects     []PreparedObject
	Claimed     bool
}

type UploadOutcome struct {
	SessionID   uint
	MediaID     uint
	Status      string
	Name        string
	ContentType string
	SizeBytes   int64
	Width       int64
	Height      int64
	Replayed    bool
	ErrorCode   string
}

type UploadService struct {
	processor           *ImageProcessor
	storage             storageservices.StorageProvider
	repository          UploadRepository
	storageConnectionID uint
}

func NewUploadService(processor *ImageProcessor, storage storageservices.StorageProvider, repository UploadRepository) *UploadService {
	return NewUploadServiceWithConnection(processor, storage, repository, 0)
}

func NewUploadServiceWithConnection(processor *ImageProcessor, storage storageservices.StorageProvider, repository UploadRepository, storageConnectionID uint) *UploadService {
	return &UploadService{processor: processor, storage: storage, repository: repository, storageConnectionID: storageConnectionID}
}

func NewDatabaseUploadService() *UploadService {
	disk := facades.Storage().Disk("fastimg")
	var provider storageservices.StorageProvider = storageservices.NewLocalProvider(disk)
	storageConnectionID := uint(0)
	if configured, connection, err := storageservices.NewRuntimeRegistry(disk).Primary(); err == nil {
		provider = configured
		storageConnectionID = connection.ID
	}
	return NewUploadServiceWithConnection(NewImageProcessor(ImageLimits{}), provider, NewDatabaseRepository(), storageConnectionID)
}

func (s *UploadService) GetStatus(ctx context.Context, userID, sessionID uint) (UploadOutcome, error) {
	if userID == 0 || sessionID == 0 {
		return UploadOutcome{}, ErrInvalidUploadInput
	}
	return s.repository.GetUploadStatus(ctx, userID, sessionID)
}

func (s *UploadService) Retry(ctx context.Context, userID, sessionID uint) (UploadOutcome, error) {
	if userID == 0 || sessionID == 0 {
		return UploadOutcome{}, ErrInvalidUploadInput
	}
	now := time.Now().UTC()
	recovery, err := s.repository.ClaimStaleUpload(ctx, userID, sessionID, now)
	if err != nil {
		return UploadOutcome{}, err
	}
	outcome := recovery.Outcome
	if outcome.Status == "ready" || outcome.Status == "failed" {
		return outcome, nil
	}
	if outcome.Status != "processing" || !recovery.Claimed {
		return UploadOutcome{}, ErrUploadInProgress
	}
	if !recoveryHasRequiredVariants(recovery) {
		return s.failRecoveredUpload(ctx, recovery, outcome)
	}
	for _, object := range recovery.Objects {
		reservedKey, ok := recovery.Reservation.ObjectKeys[object.Name]
		if !ok || reservedKey == "" || reservedKey != object.Key {
			return s.failRecoveredUpload(ctx, recovery, outcome)
		}
		if err := verifyStoredObject(ctx, s.storage, object); err != nil {
			return s.failRecoveredUpload(ctx, recovery, outcome)
		}
	}
	if err := s.repository.CompleteUpload(ctx, sessionID); err != nil {
		outcome.Status = "processing"
		return outcome, nil
	}
	outcome.Status = "ready"
	return outcome, nil
}

func recoveryHasRequiredVariants(recovery UploadRecovery) bool {
	// New sessions reserve only the original. Accept the former three-variant
	// shape for an already persisted processing session so a deployment does not
	// strand an upload that started before this storage policy changed.
	if hasRequiredObjects(recovery, map[string]bool{"original": true}) {
		return true
	}
	return hasRequiredObjects(recovery, map[string]bool{"original": true, "thumbnail": true, "medium": true})
}

func hasRequiredObjects(recovery UploadRecovery, required map[string]bool) bool {
	if len(recovery.Objects) != len(required) || len(recovery.Reservation.ObjectKeys) != len(required) {
		return false
	}
	for _, object := range recovery.Objects {
		if !required[object.Name] || object.Key == "" || recovery.Reservation.ObjectKeys[object.Name] != object.Key {
			return false
		}
		delete(required, object.Name)
	}
	return len(required) == 0
}

func (s *UploadService) failRecoveredUpload(ctx context.Context, recovery UploadRecovery, outcome UploadOutcome) (UploadOutcome, error) {
	const errorCode = "UPLOAD_RECOVERY_OBJECT_INVALID"
	cleanupErr := s.deleteReservedObjects(ctx, recovery.Reservation, recovery.Objects)
	failErr := s.repository.FailUpload(ctx, recovery.Reservation.SessionID, errorCode)
	if err := errors.Join(cleanupErr, failErr); err != nil {
		return UploadOutcome{}, errors.Join(ErrStorageWriteFailed, err)
	}
	outcome.Status = "failed"
	outcome.ErrorCode = errorCode
	return outcome, nil
}

func uploadRecoveryLeaseExpired(status string, updatedAt, now time.Time) bool {
	return status == "processing" && !updatedAt.IsZero() && updatedAt.Before(now.Add(-uploadRecoveryGracePeriod))
}

func (s *UploadService) Upload(ctx context.Context, input UploadInput) (UploadOutcome, error) {
	if input.UserID == 0 || len(input.Content) == 0 || len(input.IdempotencyKey) < 8 || len(input.IdempotencyKey) > 160 {
		return UploadOutcome{}, ErrInvalidUploadInput
	}
	originalName := sanitizeOriginalName(input.OriginalName)
	if originalName == "" {
		return UploadOutcome{}, ErrInvalidUploadInput
	}
	sourceHash := sha256.Sum256(input.Content)
	metadata := UploadMetadata{
		UserID: input.UserID, StorageConnectionID: s.storageConnectionID, Channel: input.Channel, SizeBytes: int64(len(input.Content)),
		SHA256: hex.EncodeToString(sourceHash[:]), IdempotencyKey: input.IdempotencyKey,
	}
	if err := s.repository.CheckUploadAllowance(ctx, metadata); err != nil {
		return UploadOutcome{}, err
	}
	processed, err := s.processor.ProcessWithOptions(originalName, input.DeclaredContentType, input.Content, ProcessOptions{
		WatermarkEnabled: input.WatermarkEnabled,
		WatermarkText:    input.WatermarkText,
		WatermarkDomain:  input.WatermarkDomain,
	})
	if err != nil {
		return UploadOutcome{}, err
	}

	metadata.OriginalName, metadata.ContentType, metadata.Format = originalName, processed.ContentType, processed.Format
	metadata.Width, metadata.Height = processed.Width, processed.Height
	objects, err := preparedObjects(input.UserID, processed)
	if err != nil {
		return UploadOutcome{}, err
	}
	reservation, err := s.repository.BeginUpload(ctx, metadata, objects)
	if err != nil {
		return UploadOutcome{}, err
	}
	outcome := UploadOutcome{
		SessionID: reservation.SessionID, MediaID: reservation.MediaID, Status: reservation.Status,
		Name: originalName, ContentType: processed.ContentType, SizeBytes: int64(len(input.Content)),
		Width: processed.Width, Height: processed.Height, Replayed: reservation.Status == "ready",
	}
	if reservation.Status == "ready" {
		return outcome, nil
	}
	if reservation.InProgress {
		return outcome, nil
	}

	for index := range objects {
		key, ok := reservation.ObjectKeys[objects[index].Name]
		if !ok || key == "" {
			return UploadOutcome{}, ErrInvalidUploadInput
		}
		objects[index].Key = key
		err := s.storage.Put(ctx, key, objects[index].ContentType, objects[index].Content)
		if err == nil {
			err = verifyStoredObject(ctx, s.storage, objects[index])
		}
		if err != nil {
			cleanupErr := s.deleteReservedObjects(ctx, reservation, objects)
			failErr := s.repository.FailUpload(ctx, reservation.SessionID, "UPLOAD_STORAGE_WRITE_FAILED")
			return UploadOutcome{}, errors.Join(ErrStorageWriteFailed, err, cleanupErr, failErr)
		}
	}
	if err := s.repository.CompleteUpload(ctx, reservation.SessionID); err != nil {
		// Keep written objects: the persisted processing session is the recovery
		// record, and a retry may safely complete the same operation.
		outcome.Status = "processing"
		return outcome, nil
	}
	outcome.Status = "ready"
	return outcome, nil
}

func verifyStoredObject(ctx context.Context, provider storageservices.StorageProvider, expected PreparedObject) error {
	metadata, err := provider.GetMetadata(ctx, expected.Key)
	if err != nil {
		return err
	}
	if metadata.SizeBytes != expected.SizeBytes || metadata.ContentType != expected.ContentType {
		return errors.New("stored object metadata does not match upload reservation")
	}
	content, err := provider.Get(ctx, expected.Key)
	if err != nil {
		return err
	}
	hash := sha256.Sum256(content)
	if int64(len(content)) != expected.SizeBytes || hex.EncodeToString(hash[:]) != expected.SHA256 {
		return errors.New("stored object checksum does not match upload reservation")
	}
	return nil
}

func (s *UploadService) deleteReservedObjects(ctx context.Context, reservation UploadReservation, objects []PreparedObject) error {
	var cleanupErrors []error
	for _, object := range objects {
		key := reservation.ObjectKeys[object.Name]
		if key == "" {
			continue
		}
		if err := s.storage.Delete(ctx, key); err != nil {
			cleanupErrors = append(cleanupErrors, err)
		}
	}
	return errors.Join(cleanupErrors...)
}

func preparedObjects(userID uint, processed ProcessedImage) ([]PreparedObject, error) {
	assetID, err := randomObjectID()
	if err != nil {
		return nil, err
	}
	root := fmt.Sprintf("users/%d/%s", userID, assetID)
	inputs := []struct {
		name    string
		content []byte
	}{
		{name: "original", content: processed.Original},
	}
	objects := make([]PreparedObject, 0, len(inputs))
	for _, input := range inputs {
		config, _, err := image.DecodeConfig(bytes.NewReader(input.content))
		if err != nil {
			return nil, fmt.Errorf("%w: could not inspect generated variant", ErrImageProcessing)
		}
		hash := sha256.Sum256(input.content)
		objects = append(objects, PreparedObject{
			Name: input.name, Key: fmt.Sprintf("%s/%s.%s", root, input.name, extensionForFormat(processed.Format)),
			ContentType: processed.ContentType, Content: input.content,
			SHA256: hex.EncodeToString(hash[:]), SizeBytes: int64(len(input.content)),
			Width: int64(config.Width), Height: int64(config.Height),
		})
	}
	return objects, nil
}

func extensionForFormat(format string) string {
	if format == "jpeg" {
		return "jpg"
	}
	return format
}

func randomObjectID() (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw[:]), nil
}

func sanitizeOriginalName(filename string) string {
	filename = strings.ReplaceAll(filename, "\\", "/")
	if separator := strings.LastIndex(filename, "/"); separator >= 0 {
		filename = filename[separator+1:]
	}
	filename = strings.ToValidUTF8(filename, "")
	var result strings.Builder
	for _, character := range filename {
		if unicode.IsControl(character) {
			continue
		}
		result.WriteRune(character)
		if result.Len() >= 240 {
			break
		}
	}
	return strings.TrimSpace(result.String())
}
