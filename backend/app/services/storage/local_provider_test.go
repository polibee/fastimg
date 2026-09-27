package storage

import (
	"context"
	"errors"
	"mime"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLocalProviderStoresReadsCopiesAndDeletesObjects(t *testing.T) {
	disk := newMemoryDisk()
	provider := NewLocalProvider(disk)
	ctx := context.Background()
	body := []byte("image bytes")

	require.NoError(t, provider.Put(ctx, "users/7/asset/original.png", "image/png", body))
	exists, err := provider.Exists(ctx, "users/7/asset/original.png")
	require.NoError(t, err)
	require.True(t, exists)
	got, err := provider.Get(ctx, "users/7/asset/original.png")
	require.NoError(t, err)
	require.Equal(t, body, got)
	metadata, err := provider.GetMetadata(ctx, "users/7/asset/original.png")
	require.NoError(t, err)
	require.Equal(t, int64(len(body)), metadata.SizeBytes)
	require.Equal(t, "image/png", metadata.ContentType)

	require.NoError(t, provider.Copy(ctx, "users/7/asset/original.png", "users/7/asset/copy.png"))
	exists, err = provider.Exists(ctx, "users/7/asset/copy.png")
	require.NoError(t, err)
	require.True(t, exists)
	require.NoError(t, provider.Delete(ctx, "users/7/asset/copy.png"))
	exists, err = provider.Exists(ctx, "users/7/asset/copy.png")
	require.NoError(t, err)
	require.False(t, exists)
}

func TestLocalProviderRejectsUnsafeObjectKeys(t *testing.T) {
	provider := NewLocalProvider(newMemoryDisk())
	for _, key := range []string{"../outside.png", "users/../outside.png", "/absolute.png", `users\7\photo.png`} {
		t.Run(key, func(t *testing.T) {
			err := provider.Put(context.Background(), key, "image/png", []byte("x"))
			require.ErrorIs(t, err, ErrInvalidObjectKey)
		})
	}
}

func TestLocalProviderDoesNotPretendToSupportSignedURLsOrMultipart(t *testing.T) {
	provider := NewLocalProvider(newMemoryDisk())
	_, err := provider.CreateSignedURL(context.Background(), "users/7/photo.png", 0)
	require.ErrorIs(t, err, ErrOperationNotSupported)
	err = provider.CompleteMultipartUpload(context.Background(), "users/7/photo.png", nil)
	require.ErrorIs(t, err, ErrOperationNotSupported)
}

type memoryDisk struct {
	objects map[string][]byte
	types   map[string]string
}

func newMemoryDisk() *memoryDisk {
	return &memoryDisk{objects: make(map[string][]byte), types: make(map[string]string)}
}

func (d *memoryDisk) Put(key, content string) error {
	d.objects[key] = []byte(content)
	d.types[key] = mime.TypeByExtension(filepath.Ext(key))
	return nil
}

func (d *memoryDisk) GetBytes(key string) ([]byte, error) {
	body, ok := d.objects[key]
	if !ok {
		return nil, errors.New("missing")
	}
	return append([]byte(nil), body...), nil
}

func (d *memoryDisk) Size(key string) (int64, error) {
	body, ok := d.objects[key]
	if !ok {
		return 0, errors.New("missing")
	}
	return int64(len(body)), nil
}

func (d *memoryDisk) MimeType(key string) (string, error) {
	contentType, ok := d.types[key]
	if !ok {
		return "", errors.New("missing")
	}
	return contentType, nil
}

func (d *memoryDisk) Delete(keys ...string) error {
	for _, key := range keys {
		delete(d.objects, key)
		delete(d.types, key)
	}
	return nil
}

func (d *memoryDisk) Copy(from, to string) error {
	body, ok := d.objects[from]
	if !ok {
		return errors.New("missing")
	}
	d.objects[to] = append([]byte(nil), body...)
	d.types[to] = d.types[from]
	return nil
}

func (d *memoryDisk) Exists(key string) bool {
	_, ok := d.objects[key]
	return ok
}
