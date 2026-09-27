package media

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"goravel/app/models"
)

func TestMediaLibraryDoesNotReadStorageBeforeOwnershipCheck(t *testing.T) {
	repository := &fakeMediaLibraryRepository{variantErr: ErrMediaNotFound}
	storage := newFakeStorageProvider()
	service := NewMediaLibraryService(repository, storage)

	_, err := service.GetContent(context.Background(), 9, 22, "thumbnail")

	require.ErrorIs(t, err, ErrMediaNotFound)
	require.Empty(t, storage.objects)
}

func TestMediaLibraryDetailsAreOwnerScopedAndIncludeReadyVariants(t *testing.T) {
	asset := models.MediaAsset{UserID: 9, OriginalName: "photo.png", Status: "ready", Width: 800, Height: 600}
	repository := &fakeMediaLibraryRepository{asset: asset, detailErr: map[string]error{"medium": ErrVariantNotFound}}
	service := NewMediaLibraryService(repository, newFakeStorageProvider())

	item, err := service.GetDetails(context.Background(), 9, 22)

	require.NoError(t, err)
	require.Equal(t, asset, item.Asset)
	require.Len(t, item.Variants, 2)
	require.Equal(t, []string{"original", "thumbnail"}, []string{item.Variants[0].Variant.Name, item.Variants[1].Variant.Name})
}

func TestMediaLibraryDetailsDoNotReturnAnotherUsersMedia(t *testing.T) {
	repository := &fakeMediaLibraryRepository{detailErr: map[string]error{}}
	service := NewMediaLibraryService(repository, newFakeStorageProvider())

	_, err := service.GetDetails(context.Background(), 8, 22)

	require.ErrorIs(t, err, ErrMediaNotFound)
}

type fakeMediaLibraryRepository struct {
	variantErr error
	asset      models.MediaAsset
	detailErr  map[string]error
}

func (r *fakeMediaLibraryRepository) ListOwned(context.Context, uint, bool, string, int, int) ([]MediaListItem, int64, error) {
	return nil, 0, nil
}

func (r *fakeMediaLibraryRepository) FindOwned(_ context.Context, userID, _ uint, _ bool) (models.MediaAsset, error) {
	if r.asset.UserID != userID {
		return models.MediaAsset{}, ErrMediaNotFound
	}
	return r.asset, nil
}

func (r *fakeMediaLibraryRepository) FindOwnedVariant(_ context.Context, userID, _ uint, name string) (MediaVariantObject, error) {
	if r.asset.UserID != userID {
		return MediaVariantObject{}, ErrMediaNotFound
	}
	if err := r.detailErr[name]; err != nil {
		return MediaVariantObject{}, err
	}
	if r.variantErr != nil {
		return MediaVariantObject{}, r.variantErr
	}
	return MediaVariantObject{Variant: models.MediaVariant{Name: name, Status: "ready"}, Object: models.StorageObject{Status: "ready", ContentType: "image/png"}}, nil
}

func (r *fakeMediaLibraryRepository) SoftDelete(context.Context, uint, uint) (models.MediaAsset, error) {
	return models.MediaAsset{}, errors.New("unexpected call")
}

func (r *fakeMediaLibraryRepository) UpdateVisibility(context.Context, uint, uint, string) (models.MediaAsset, error) {
	return models.MediaAsset{}, errors.New("unexpected call")
}

func (r *fakeMediaLibraryRepository) Restore(context.Context, uint, uint) (models.MediaAsset, error) {
	return models.MediaAsset{}, errors.New("unexpected call")
}
