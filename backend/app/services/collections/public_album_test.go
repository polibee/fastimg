package collections

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPublicAlbumUsesPublicOnlyRepositoryBoundary(t *testing.T) {
	repository := &publicAlbumRepositoryFake{fakeAlbumMediaRepository: &fakeAlbumMediaRepository{}}
	service := NewService(repository)

	album, err := service.PublicAlbum(context.Background(), 11)

	require.NoError(t, err)
	require.Equal(t, uint(11), album.ID)
	require.Equal(t, "public", album.Visibility)
	require.Len(t, album.Media, 1)
	require.Equal(t, uint(51), album.Media[0].ID)
}

func TestPublicAlbumDoesNotInventAVisibilityBypass(t *testing.T) {
	repository := &publicAlbumRepositoryFake{
		fakeAlbumMediaRepository: &fakeAlbumMediaRepository{},
		err:                      ErrNotFound,
	}
	service := NewService(repository)

	_, err := service.PublicAlbum(context.Background(), 12)

	require.ErrorIs(t, err, ErrNotFound)
}

type publicAlbumRepositoryFake struct {
	*fakeAlbumMediaRepository
	err error
}

func (r *publicAlbumRepositoryFake) FindPublicAlbum(_ context.Context, albumID uint) (PublicAlbum, error) {
	if r.err != nil {
		return PublicAlbum{}, r.err
	}
	return PublicAlbum{
		ID:         albumID,
		Name:       "Public album",
		Visibility: "public",
		Media:      []PublicAlbumMedia{{ID: 51, OriginalName: "public.jpg", ContentType: "image/jpeg"}},
	}, nil
}
