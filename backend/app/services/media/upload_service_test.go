package media

import (
	"context"
	"errors"
	"image"
	"image/png"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	storageservices "goravel/app/services/storage"
)

func TestUploadServiceStoresVariantsAndFinalizesOneMediaAsset(t *testing.T) {
	storage := newFakeStorageProvider()
	repository := &fakeUploadRepository{}
	service := NewUploadService(NewImageProcessor(ImageLimits{}), storage, repository)

	outcome, err := service.Upload(context.Background(), UploadInput{
		UserID: 7, OriginalName: "sample.png", DeclaredContentType: "image/png",
		IdempotencyKey: "request-123456", Content: testPNG(t, 64, 32),
	})

	require.NoError(t, err)
	require.Equal(t, "ready", outcome.Status)
	require.Equal(t, uint(22), outcome.MediaID)
	require.Equal(t, 3, len(storage.objects))
	require.Equal(t, uint(11), repository.completedSessionID)
	require.Empty(t, repository.failedSessionID)
}

func TestUploadServiceCleansObjectsAndMarksSessionFailedWhenStorageWriteFails(t *testing.T) {
	storage := newFakeStorageProvider()
	storage.failOn = "medium"
	repository := &fakeUploadRepository{}
	service := NewUploadService(NewImageProcessor(ImageLimits{}), storage, repository)

	_, err := service.Upload(context.Background(), UploadInput{
		UserID: 7, OriginalName: "sample.png", DeclaredContentType: "image/png",
		IdempotencyKey: "request-123456", Content: testPNG(t, 64, 32),
	})

	require.ErrorIs(t, err, ErrStorageWriteFailed)
	require.Equal(t, uint(11), repository.failedSessionID)
	require.Empty(t, storage.objects)
}

func TestUploadServiceRejectsStoredObjectsWhoseBytesDoNotMatchReservation(t *testing.T) {
	storage := newFakeStorageProvider()
	storage.tamperOn = "medium"
	repository := &fakeUploadRepository{}
	service := NewUploadService(NewImageProcessor(ImageLimits{}), storage, repository)

	_, err := service.Upload(context.Background(), UploadInput{
		UserID: 7, OriginalName: "sample.png", DeclaredContentType: "image/png",
		IdempotencyKey: "request-123456", Content: testPNG(t, 64, 32),
	})

	require.ErrorIs(t, err, ErrStorageWriteFailed)
	require.Equal(t, uint(11), repository.failedSessionID)
	require.Empty(t, storage.objects)
}

func TestUploadServiceKeepsObjectsForRecoveryWhenFinalizationFails(t *testing.T) {
	storage := newFakeStorageProvider()
	repository := &fakeUploadRepository{completeErr: errors.New("database temporarily unavailable")}
	service := NewUploadService(NewImageProcessor(ImageLimits{}), storage, repository)

	outcome, err := service.Upload(context.Background(), UploadInput{
		UserID: 7, OriginalName: "sample.png", DeclaredContentType: "image/png",
		IdempotencyKey: "request-123456", Content: testPNG(t, 64, 32),
	})

	require.NoError(t, err)
	require.Equal(t, "processing", outcome.Status)
	require.Len(t, storage.objects, 3)
	require.Empty(t, repository.failedSessionID)
}

func TestUploadServiceReturnsReadyIdempotentReplayWithoutWritingAgain(t *testing.T) {
	storage := newFakeStorageProvider()
	repository := &fakeUploadRepository{replayReady: true}
	service := NewUploadService(NewImageProcessor(ImageLimits{}), storage, repository)

	outcome, err := service.Upload(context.Background(), UploadInput{
		UserID: 7, OriginalName: "sample.png", DeclaredContentType: "image/png",
		IdempotencyKey: "request-123456", Content: testPNG(t, 64, 32),
	})

	require.NoError(t, err)
	require.Equal(t, "ready", outcome.Status)
	require.Equal(t, uint(22), outcome.MediaID)
	require.Empty(t, storage.objects)
	require.Empty(t, repository.completedSessionID)
}

func TestUploadServiceChecksQuotaBeforeProcessingImage(t *testing.T) {
	quotaErr := errors.New("monthly transform quota reached")
	repository := &fakeUploadRepository{allowanceErr: quotaErr}
	service := NewUploadService(NewImageProcessor(ImageLimits{}), newFakeStorageProvider(), repository)

	_, err := service.Upload(context.Background(), UploadInput{
		UserID: 7, OriginalName: "not-an-image.png", DeclaredContentType: "image/png",
		IdempotencyKey: "request-123456", Content: []byte("invalid image bytes"),
	})

	require.ErrorIs(t, err, quotaErr)
	require.Equal(t, uint(7), repository.allowanceUserID)
	require.Equal(t, int64(len("invalid image bytes")), repository.allowanceSize)
	require.Empty(t, repository.completedSessionID)
}

func TestUploadServiceDoesNotOverwriteObjectsForAnUploadAlreadyInProgress(t *testing.T) {
	storage := newFakeStorageProvider()
	repository := &fakeUploadRepository{inProgress: true}
	service := NewUploadService(NewImageProcessor(ImageLimits{}), storage, repository)

	outcome, err := service.Upload(context.Background(), UploadInput{
		UserID: 7, OriginalName: "sample.png", DeclaredContentType: "image/png",
		IdempotencyKey: "request-123456", Content: testPNG(t, 64, 32),
	})

	require.NoError(t, err)
	require.Equal(t, "processing", outcome.Status)
	require.Empty(t, storage.objects)
	require.Empty(t, repository.completedSessionID)
}

func TestUploadServiceReadsStatusWithinTheAuthenticatedUserScope(t *testing.T) {
	repository := &fakeUploadRepository{statusOutcome: UploadOutcome{SessionID: 11, MediaID: 22, Status: "processing"}}
	service := NewUploadService(NewImageProcessor(ImageLimits{}), newFakeStorageProvider(), repository)

	got, err := service.GetStatus(context.Background(), 7, 11)

	require.NoError(t, err)
	require.Equal(t, uint(7), repository.statusUserID)
	require.Equal(t, uint(11), repository.statusSessionID)
	require.Equal(t, "processing", got.Status)
}

func TestUploadServiceRejectsMissingIdentityBeforeRepositoryLookup(t *testing.T) {
	repository := &fakeUploadRepository{}
	service := NewUploadService(NewImageProcessor(ImageLimits{}), newFakeStorageProvider(), repository)

	_, err := service.GetStatus(context.Background(), 0, 11)

	require.ErrorIs(t, err, ErrInvalidUploadInput)
	require.Zero(t, repository.statusUserID)
}

func testPNG(t *testing.T, width, height int) []byte {
	t.Helper()
	var encoded strings.Builder
	require.NoError(t, png.Encode(&encoded, image.NewRGBA(image.Rect(0, 0, width, height))))
	return []byte(encoded.String())
}

type fakeUploadRepository struct {
	completedSessionID uint
	failedSessionID    uint
	allowanceErr       error
	allowanceUserID    uint
	allowanceSize      int64
	completeErr        error
	replayReady        bool
	inProgress         bool
	statusOutcome      UploadOutcome
	statusUserID       uint
	statusSessionID    uint
	recovery           UploadRecovery
	recoveryUserID     uint
	recoverySessionID  uint
}

func (r *fakeUploadRepository) CheckUploadAllowance(_ context.Context, input UploadMetadata) error {
	r.allowanceUserID = input.UserID
	r.allowanceSize = input.SizeBytes
	return r.allowanceErr
}

func (r *fakeUploadRepository) BeginUpload(_ context.Context, _ UploadMetadata, objects []PreparedObject) (UploadReservation, error) {
	keys := make(map[string]string, len(objects))
	for _, object := range objects {
		keys[object.Name] = object.Key
	}
	status := "processing"
	if r.replayReady {
		status = "ready"
	}
	return UploadReservation{SessionID: 11, MediaID: 22, Status: status, ObjectKeys: keys, InProgress: r.inProgress}, nil
}

func (r *fakeUploadRepository) CompleteUpload(_ context.Context, sessionID uint) error {
	r.completedSessionID = sessionID
	return r.completeErr
}

func (r *fakeUploadRepository) FailUpload(_ context.Context, sessionID uint, _ string) error {
	r.failedSessionID = sessionID
	return nil
}

func (r *fakeUploadRepository) GetUploadStatus(_ context.Context, userID, sessionID uint) (UploadOutcome, error) {
	r.statusUserID = userID
	r.statusSessionID = sessionID
	return r.statusOutcome, nil
}

func (r *fakeUploadRepository) ClaimStaleUpload(_ context.Context, userID, sessionID uint, _ time.Time) (UploadRecovery, error) {
	r.recoveryUserID = userID
	r.recoverySessionID = sessionID
	return r.recovery, nil
}

type fakeStorageProvider struct {
	objects  map[string][]byte
	types    map[string]string
	failOn   string
	tamperOn string
}

func newFakeStorageProvider() *fakeStorageProvider {
	return &fakeStorageProvider{objects: make(map[string][]byte), types: make(map[string]string)}
}

func (s *fakeStorageProvider) Put(_ context.Context, key, _ string, content []byte) error {
	if s.failOn != "" && strings.Contains(key, "/"+s.failOn+".") {
		return errors.New("disk full")
	}
	if s.tamperOn != "" && strings.Contains(key, "/"+s.tamperOn+".") {
		content = append(append([]byte(nil), content...), []byte("tampered")...)
	}
	s.objects[key] = append([]byte(nil), content...)
	s.types[key] = "image/png"
	return nil
}

func (s *fakeStorageProvider) CompleteMultipartUpload(context.Context, string, []storageservices.MultipartPart) error {
	return storageservices.ErrOperationNotSupported
}

func (s *fakeStorageProvider) Get(_ context.Context, key string) ([]byte, error) {
	body, ok := s.objects[key]
	if !ok {
		return nil, storageservices.ErrObjectNotFound
	}
	return append([]byte(nil), body...), nil
}

func (s *fakeStorageProvider) GetMetadata(_ context.Context, key string) (storageservices.ObjectMetadata, error) {
	body, ok := s.objects[key]
	if !ok {
		return storageservices.ObjectMetadata{}, storageservices.ErrObjectNotFound
	}
	return storageservices.ObjectMetadata{SizeBytes: int64(len(body)), ContentType: s.types[key]}, nil
}

func (s *fakeStorageProvider) Delete(_ context.Context, key string) error {
	delete(s.objects, key)
	delete(s.types, key)
	return nil
}

func (s *fakeStorageProvider) CreateSignedURL(context.Context, string, time.Duration) (string, error) {
	return "", storageservices.ErrOperationNotSupported
}

func (s *fakeStorageProvider) Copy(_ context.Context, from, to string) error {
	body, ok := s.objects[from]
	if !ok {
		return storageservices.ErrObjectNotFound
	}
	s.objects[to] = append([]byte(nil), body...)
	s.types[to] = s.types[from]
	return nil
}

func (s *fakeStorageProvider) Exists(_ context.Context, key string) (bool, error) {
	_, exists := s.objects[key]
	return exists, nil
}
