package media

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestUploadRecoveryLeaseOnlyExpiresStaleProcessingWork(t *testing.T) {
	now := time.Date(2026, time.September, 23, 12, 0, 0, 0, time.UTC)

	require.False(t, uploadRecoveryLeaseExpired("processing", now.Add(-9*time.Minute), now))
	require.False(t, uploadRecoveryLeaseExpired("processing", now.Add(-10*time.Minute), now))
	require.True(t, uploadRecoveryLeaseExpired("processing", now.Add(-11*time.Minute), now))
	require.False(t, uploadRecoveryLeaseExpired("ready", now.Add(-11*time.Minute), now))
	require.False(t, uploadRecoveryLeaseExpired("processing", time.Time{}, now))
}

func TestUploadRetryVerifiesReservedObjectsBeforeFinalizing(t *testing.T) {
	processor := NewImageProcessor(ImageLimits{})
	processed, err := processor.Process("sample.png", "image/png", testPNG(t, 64, 32))
	require.NoError(t, err)
	objects, err := preparedObjects(7, processed)
	require.NoError(t, err)

	storage := newFakeStorageProvider()
	keys := make(map[string]string, len(objects))
	for _, object := range objects {
		keys[object.Name] = object.Key
		require.NoError(t, storage.Put(context.Background(), object.Key, object.ContentType, object.Content))
	}
	repository := &fakeUploadRepository{recovery: UploadRecovery{
		Claimed:     true,
		Outcome:     UploadOutcome{SessionID: 11, MediaID: 22, Status: "processing", Name: "sample.png"},
		Reservation: UploadReservation{SessionID: 11, MediaID: 22, Status: "processing", ObjectKeys: keys},
		Objects:     objects,
	}}
	service := NewUploadService(processor, storage, repository)

	outcome, err := service.Retry(context.Background(), 7, 11)

	require.NoError(t, err)
	require.Equal(t, "ready", outcome.Status)
	require.Equal(t, uint(11), repository.completedSessionID)
	require.Empty(t, repository.failedSessionID)
}

func TestUploadRetryFailsAndCleansCorruptReservedObjects(t *testing.T) {
	processor := NewImageProcessor(ImageLimits{})
	processed, err := processor.Process("sample.png", "image/png", testPNG(t, 64, 32))
	require.NoError(t, err)
	objects, err := preparedObjects(7, processed)
	require.NoError(t, err)

	storage := newFakeStorageProvider()
	keys := make(map[string]string, len(objects))
	for index, object := range objects {
		keys[object.Name] = object.Key
		if index == 0 {
			storage.tamperOn = object.Name
		}
		require.NoError(t, storage.Put(context.Background(), object.Key, object.ContentType, object.Content))
	}
	repository := &fakeUploadRepository{recovery: UploadRecovery{
		Claimed:     true,
		Outcome:     UploadOutcome{SessionID: 11, MediaID: 22, Status: "processing", Name: "sample.png"},
		Reservation: UploadReservation{SessionID: 11, MediaID: 22, Status: "processing", ObjectKeys: keys},
		Objects:     objects,
	}}
	service := NewUploadService(processor, storage, repository)

	outcome, err := service.Retry(context.Background(), 7, 11)

	require.NoError(t, err)
	require.Equal(t, "failed", outcome.Status)
	require.Equal(t, "UPLOAD_RECOVERY_OBJECT_INVALID", outcome.ErrorCode)
	require.Equal(t, uint(11), repository.failedSessionID)
	require.Empty(t, storage.objects)
}

func TestUploadRetryFailsWhenOriginalObjectIsMissing(t *testing.T) {
	processor := NewImageProcessor(ImageLimits{})
	processed, err := processor.Process("sample.png", "image/png", testPNG(t, 64, 32))
	require.NoError(t, err)
	objects, err := preparedObjects(7, processed)
	require.NoError(t, err)

	storage := newFakeStorageProvider()
	keys := make(map[string]string, len(objects)-1)
	recoveryObjects := make([]PreparedObject, 0, len(objects)-1)
	for _, object := range objects {
		if object.Name == "original" {
			continue
		}
		keys[object.Name] = object.Key
		recoveryObjects = append(recoveryObjects, object)
		require.NoError(t, storage.Put(context.Background(), object.Key, object.ContentType, object.Content))
	}
	repository := &fakeUploadRepository{recovery: UploadRecovery{
		Claimed:     true,
		Outcome:     UploadOutcome{SessionID: 11, MediaID: 22, Status: "processing", Name: "sample.png"},
		Reservation: UploadReservation{SessionID: 11, MediaID: 22, Status: "processing", ObjectKeys: keys},
		Objects:     recoveryObjects,
	}}
	service := NewUploadService(processor, storage, repository)

	outcome, err := service.Retry(context.Background(), 7, 11)

	require.NoError(t, err)
	require.Equal(t, "failed", outcome.Status)
	require.Equal(t, "UPLOAD_RECOVERY_OBJECT_INVALID", outcome.ErrorCode)
	require.Equal(t, uint(11), repository.failedSessionID)
	require.Empty(t, storage.objects)
}
