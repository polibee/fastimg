package media

import (
	"context"
	"errors"
	"testing"

	"github.com/goravel/framework/database/orm"
	"github.com/stretchr/testify/require"
	"goravel/app/models"
)

func TestPermanentDeleteRemovesUniqueObjectsBeforeFinalizingStorageRelease(t *testing.T) {
	repository := &fakePermanentDeleteRepository{
		fakeMediaLibraryRepository: &fakeMediaLibraryRepository{},
		asset:                      models.MediaAsset{Model: orm.Model{ID: 42}, UserID: 7, Status: "deleted"},
		objects: []models.StorageObject{
			{Model: orm.Model{ID: 1}, ObjectKey: "media/42/original.png", Status: "ready"},
			{Model: orm.Model{ID: 2}, ObjectKey: "media/42/thumbnail.png", Status: "ready"},
			{Model: orm.Model{ID: 3}, ObjectKey: "media/42/original.png", Status: "ready"},
		},
	}
	storage := &countingDeleteStorage{fakeStorageProvider: newFakeStorageProvider(), calls: make(map[string]int)}
	storage.objects["media/42/original.png"] = []byte("original")
	storage.objects["media/42/thumbnail.png"] = []byte("thumbnail")
	service := NewMediaLibraryService(repository, storage)

	asset, err := service.PermanentDelete(context.Background(), 7, 42)

	require.NoError(t, err)
	require.Equal(t, "physically_deleted", asset.Status)
	require.Empty(t, storage.objects)
	require.Equal(t, map[string]int{"media/42/original.png": 1, "media/42/thumbnail.png": 1}, storage.calls)
	require.Equal(t, 1, repository.finalizeCalls)
	require.Equal(t, uint(7), repository.finalizedUserID)
	require.Equal(t, uint(42), repository.finalizedMediaID)
}

func TestPermanentDeleteDoesNotFinalizeWhenStorageRemovalFails(t *testing.T) {
	repository := &fakePermanentDeleteRepository{
		fakeMediaLibraryRepository: &fakeMediaLibraryRepository{},
		asset:                      models.MediaAsset{Model: orm.Model{ID: 42}, UserID: 7, Status: "deleted"},
		objects:                    []models.StorageObject{{Model: orm.Model{ID: 1}, ObjectKey: "media/42/original.png", Status: "ready"}},
	}
	storage := &failingDeleteStorage{fakeStorageProvider: newFakeStorageProvider(), err: errors.New("disk unavailable")}
	service := NewMediaLibraryService(repository, storage)

	_, err := service.PermanentDelete(context.Background(), 7, 42)

	require.ErrorIs(t, err, storage.err)
	require.Zero(t, repository.finalizeCalls)
}

func TestPermanentDeleteRequiresExplicitConfirmationPhrase(t *testing.T) {
	for _, input := range []string{"", "true", " delete", "permanently-delete "} {
		if PermanentDeleteConfirmed(input) {
			t.Errorf("permanentDeleteConfirmed(%q) = true, want false", input)
		}
	}
	if !PermanentDeleteConfirmed("permanently-delete") {
		t.Fatal("permanentDeleteConfirmed(permanently-delete) = false, want true")
	}
}

func TestEmptyTrashRequiresExplicitConfirmationPhrase(t *testing.T) {
	for _, input := range []string{"", "true", "empty-trash ", " permanently-delete"} {
		if EmptyTrashConfirmed(input) {
			t.Errorf("emptyTrashConfirmed(%q) = true, want false", input)
		}
	}
	if !EmptyTrashConfirmed("empty-trash") {
		t.Fatal("emptyTrashConfirmed(empty-trash) = false, want true")
	}
}

type fakePermanentDeleteRepository struct {
	*fakeMediaLibraryRepository
	asset            models.MediaAsset
	objects          []models.StorageObject
	finalizeCalls    int
	finalizedUserID  uint
	finalizedMediaID uint
}

func (r *fakePermanentDeleteRepository) PreparePermanentDelete(_ context.Context, userID, mediaID uint) (models.MediaAsset, []models.StorageObject, error) {
	if r.asset.UserID != userID || r.asset.ID != mediaID || r.asset.Status != "deleted" {
		return models.MediaAsset{}, nil, ErrMediaNotFound
	}
	return r.asset, append([]models.StorageObject(nil), r.objects...), nil
}

func (r *fakePermanentDeleteRepository) FinalizePermanentDelete(_ context.Context, userID, mediaID uint) error {
	r.finalizeCalls++
	r.finalizedUserID, r.finalizedMediaID = userID, mediaID
	r.asset.Status = "physically_deleted"
	return nil
}

type failingDeleteStorage struct {
	*fakeStorageProvider
	err error
}

func (s *failingDeleteStorage) Delete(context.Context, string) error { return s.err }

type countingDeleteStorage struct {
	*fakeStorageProvider
	calls map[string]int
}

func (s *countingDeleteStorage) Delete(ctx context.Context, key string) error {
	s.calls[key]++
	return s.fakeStorageProvider.Delete(ctx, key)
}
