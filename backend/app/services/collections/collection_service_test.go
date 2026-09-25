package collections

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCollectionServiceNormalizesNamesAndPreservesOwnerScope(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)

	item, err := service.Create(context.Background(), KindFolder, 7, Input{Name: "  Assets  "})

	require.NoError(t, err)
	require.Equal(t, "Assets", item.Name)
	require.Equal(t, uint(7), repository.createdUserID)
	require.Equal(t, KindFolder, repository.createdKind)
}

func TestCollectionServiceRejectsInvalidAlbumVisibility(t *testing.T) {
	service := NewService(&fakeRepository{})

	_, err := service.Create(context.Background(), KindAlbum, 7, Input{Name: "Public", Visibility: "invalid"})

	require.ErrorIs(t, err, ErrInvalidVisibility)
}

func TestCollectionServiceRequiresOwnedParentFolder(t *testing.T) {
	repository := &fakeRepository{parentExists: false}
	service := NewService(repository)

	_, err := service.Create(context.Background(), KindFolder, 7, Input{Name: "Nested", ParentID: intPtr(99)})

	require.ErrorIs(t, err, ErrParentNotFound)
}

func TestCollectionServiceAddsOwnedMediaToAlbumWithDeduplicatedBatch(t *testing.T) {
	repository := &fakeAlbumMediaRepository{albumExists: true}
	service := NewService(repository)

	result, err := service.AddMediaToAlbum(context.Background(), 7, 11, []uint{5, 5, 8})

	require.NoError(t, err)
	require.Equal(t, []uint{5, 8}, repository.addedMediaIDs)
	require.Equal(t, []uint{5, 8}, result.Added)
	require.Empty(t, result.Skipped)
}

func TestCollectionServiceRejectsOversizedAlbumMediaBatch(t *testing.T) {
	service := NewService(&fakeAlbumMediaRepository{})
	mediaIDs := make([]uint, MaxAlbumMediaBatch+1)
	for index := range mediaIDs {
		mediaIDs[index] = uint(index + 1)
	}

	_, err := service.AddMediaToAlbum(context.Background(), 7, 11, mediaIDs)

	require.ErrorIs(t, err, ErrAlbumMediaBatchTooLarge)
}

func TestCollectionServiceRejectsAlbumMediaFromAnotherOwner(t *testing.T) {
	repository := &fakeAlbumMediaRepository{albumExists: true, validMediaIDs: map[uint]bool{5: true}}
	service := NewService(repository)

	_, err := service.AddMediaToAlbum(context.Background(), 7, 11, []uint{5, 8})

	require.ErrorIs(t, err, ErrMediaNotFound)
}

func TestCollectionServiceBatchRemovalSkipsMissingRelations(t *testing.T) {
	repository := &fakeAlbumMediaRepository{albumExists: true, removableMediaIDs: map[uint]bool{5: true}}
	service := NewService(repository)

	result, err := service.RemoveMediaFromAlbumBatch(context.Background(), 7, 11, []uint{5, 5, 8})

	require.NoError(t, err)
	payload, err := json.Marshal(result)
	require.NoError(t, err)
	require.JSONEq(t, `{"removed_ids":[5],"skipped_ids":[8]}`, string(payload))
}

func TestCollectionServiceReordersOwnedAlbumMedia(t *testing.T) {
	repository := &fakeAlbumMediaRepository{albumExists: true}
	service := NewService(repository)

	err := service.ReorderAlbumMedia(context.Background(), 7, 11, []uint{8, 5})

	require.NoError(t, err)
	require.Equal(t, []uint{8, 5}, repository.reorderedMediaIDs)
}

func TestCollectionServiceMovesOwnedMediaBetweenAlbums(t *testing.T) {
	repository := &fakeAlbumMediaRepository{albumExists: true}
	service := NewService(repository)

	result, err := service.MoveMediaBetweenAlbums(context.Background(), 7, 11, 12, []uint{5, 8})

	require.NoError(t, err)
	require.Equal(t, []uint{5, 8}, result.Moved)
	require.Empty(t, result.Skipped)
	require.Equal(t, uint(11), repository.moveSourceAlbumID)
	require.Equal(t, uint(12), repository.moveDestinationAlbumID)
	require.Equal(t, []uint{5, 8}, repository.movedMediaIDs)
}

func intPtr(value int) *int { return &value }

type fakeRepository struct {
	createdUserID uint
	createdKind   Kind
	parentExists  bool
}

func (r *fakeRepository) List(context.Context, Kind, uint) ([]Item, error) { return nil, nil }
func (r *fakeRepository) Find(context.Context, Kind, uint, uint) (Item, error) {
	return Item{}, ErrNotFound
}
func (r *fakeRepository) ParentExists(context.Context, uint, uint) (bool, error) {
	return r.parentExists, nil
}
func (r *fakeRepository) MediaExists(context.Context, uint, uint) (bool, error) { return true, nil }
func (r *fakeRepository) Create(_ context.Context, kind Kind, userID uint, input Input) (Item, error) {
	r.createdKind, r.createdUserID = kind, userID
	return Item{Name: input.Name, UserID: userID}, nil
}
func (r *fakeRepository) Update(context.Context, Kind, uint, uint, Input) (Item, error) {
	return Item{}, nil
}
func (r *fakeRepository) Delete(context.Context, Kind, uint, uint) error { return errors.New("unused") }

type fakeAlbumMediaRepository struct {
	fakeRepository
	albumExists            bool
	validMediaIDs          map[uint]bool
	removableMediaIDs      map[uint]bool
	addedMediaIDs          []uint
	reorderedMediaIDs      []uint
	moveSourceAlbumID      uint
	moveDestinationAlbumID uint
	movedMediaIDs          []uint
}

func (r *fakeAlbumMediaRepository) Find(_ context.Context, kind Kind, _ uint, _ uint) (Item, error) {
	if kind != KindAlbum || !r.albumExists {
		return Item{}, ErrNotFound
	}
	return Item{ID: 11, UserID: 7}, nil
}

func (r *fakeAlbumMediaRepository) AddMediaToAlbum(_ context.Context, _ uint, _ uint, mediaIDs []uint) (AlbumMediaMutation, error) {
	if len(r.validMediaIDs) > 0 {
		for _, mediaID := range mediaIDs {
			if !r.validMediaIDs[mediaID] {
				return AlbumMediaMutation{}, ErrMediaNotFound
			}
		}
	}
	r.addedMediaIDs = append([]uint(nil), mediaIDs...)
	return AlbumMediaMutation{Added: mediaIDs}, nil
}

func (r *fakeAlbumMediaRepository) RemoveMediaFromAlbum(_ context.Context, _ uint, _ uint, mediaID uint) error {
	if r.removableMediaIDs != nil {
		if !r.removableMediaIDs[mediaID] {
			return ErrMediaNotFound
		}
		delete(r.removableMediaIDs, mediaID)
	}
	return nil
}

func (r *fakeAlbumMediaRepository) ReorderAlbumMedia(_ context.Context, _ uint, _ uint, mediaIDs []uint) error {
	r.reorderedMediaIDs = append([]uint(nil), mediaIDs...)
	return nil
}

func (r *fakeAlbumMediaRepository) MoveMediaBetweenAlbums(_ context.Context, _ uint, sourceAlbumID, destinationAlbumID uint, mediaIDs []uint) (AlbumMediaMove, error) {
	r.moveSourceAlbumID = sourceAlbumID
	r.moveDestinationAlbumID = destinationAlbumID
	r.movedMediaIDs = append([]uint(nil), mediaIDs...)
	return AlbumMediaMove{Moved: mediaIDs}, nil
}

func (r *fakeAlbumMediaRepository) ListAlbumMediaIDs(context.Context, uint, uint) ([]uint, error) {
	return nil, nil
}
