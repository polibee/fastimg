package storage

import (
	"context"
	"errors"
	"mime"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var (
	ErrInvalidObjectKey      = errors.New("invalid storage object key")
	ErrInvalidObjectType     = errors.New("invalid storage content type")
	ErrObjectTypeMismatch    = errors.New("storage content type does not match object key")
	ErrObjectNotFound        = errors.New("storage object not found")
	ErrOperationNotSupported = errors.New("storage operation is not supported by this provider")
)

type ObjectMetadata struct {
	SizeBytes   int64
	ContentType string
}

type MultipartPart struct {
	Number int
	ETag   string
}

// Provider is the storage boundary shared by uploads, media retrieval, and
// later S3-compatible implementations.
type StorageProvider interface {
	Put(ctx context.Context, key, contentType string, content []byte) error
	CompleteMultipartUpload(ctx context.Context, key string, parts []MultipartPart) error
	Get(ctx context.Context, key string) ([]byte, error)
	GetMetadata(ctx context.Context, key string) (ObjectMetadata, error)
	// Delete must be idempotent: retries after partial cleanup must treat an
	// already-missing object as successfully deleted.
	Delete(ctx context.Context, key string) error
	CreateSignedURL(ctx context.Context, key string, expiresIn time.Duration) (string, error)
	Copy(ctx context.Context, from, to string) error
	Exists(ctx context.Context, key string) (bool, error)
}

type localDisk interface {
	Put(file, content string) error
	GetBytes(file string) ([]byte, error)
	Size(file string) (int64, error)
	MimeType(file string) (string, error)
	Delete(file ...string) error
	Copy(oldFile, newFile string) error
	Exists(file string) bool
}

type LocalProvider struct {
	disk localDisk
}

func NewLocalProvider(disk localDisk) *LocalProvider {
	return &LocalProvider{disk: disk}
}

func (p *LocalProvider) Put(_ context.Context, key, contentType string, content []byte) error {
	if err := validateObjectKey(key); err != nil {
		return err
	}
	if len(content) == 0 {
		return ErrObjectNotFound
	}
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil || mediaType == "" {
		return ErrInvalidObjectType
	}
	extensionType := mime.TypeByExtension(strings.ToLower(filepath.Ext(key)))
	if extensionType != "" {
		extensionMediaType, _, _ := mime.ParseMediaType(extensionType)
		if mediaType != extensionMediaType {
			return ErrObjectTypeMismatch
		}
	}
	return p.disk.Put(key, string(content))
}

func (p *LocalProvider) CompleteMultipartUpload(_ context.Context, key string, _ []MultipartPart) error {
	if err := validateObjectKey(key); err != nil {
		return err
	}
	return ErrOperationNotSupported
}

func (p *LocalProvider) Get(_ context.Context, key string) ([]byte, error) {
	if err := validateObjectKey(key); err != nil {
		return nil, err
	}
	if !p.disk.Exists(key) {
		return nil, ErrObjectNotFound
	}
	return p.disk.GetBytes(key)
}

func (p *LocalProvider) GetMetadata(_ context.Context, key string) (ObjectMetadata, error) {
	if err := validateObjectKey(key); err != nil {
		return ObjectMetadata{}, err
	}
	if !p.disk.Exists(key) {
		return ObjectMetadata{}, ErrObjectNotFound
	}
	size, err := p.disk.Size(key)
	if err != nil {
		return ObjectMetadata{}, err
	}
	contentType, err := p.disk.MimeType(key)
	if err != nil {
		return ObjectMetadata{}, err
	}
	return ObjectMetadata{SizeBytes: size, ContentType: contentType}, nil
}

func (p *LocalProvider) Delete(_ context.Context, key string) error {
	if err := validateObjectKey(key); err != nil {
		return err
	}
	if !p.disk.Exists(key) {
		return nil
	}
	return p.disk.Delete(key)
}

func (p *LocalProvider) CreateSignedURL(_ context.Context, key string, _ time.Duration) (string, error) {
	if err := validateObjectKey(key); err != nil {
		return "", err
	}
	return "", ErrOperationNotSupported
}

func (p *LocalProvider) Copy(_ context.Context, from, to string) error {
	if err := validateObjectKey(from); err != nil {
		return err
	}
	if err := validateObjectKey(to); err != nil {
		return err
	}
	return p.disk.Copy(from, to)
}

func (p *LocalProvider) Exists(_ context.Context, key string) (bool, error) {
	if err := validateObjectKey(key); err != nil {
		return false, err
	}
	return p.disk.Exists(key), nil
}

func validateObjectKey(key string) error {
	if key == "" || strings.Contains(key, "\\") || strings.HasPrefix(key, "/") || path.Clean(key) != key {
		return ErrInvalidObjectKey
	}
	for _, segment := range strings.Split(key, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return ErrInvalidObjectKey
		}
		for _, character := range segment {
			if (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') ||
				(character >= '0' && character <= '9') || character == '-' || character == '_' || character == '.' {
				continue
			}
			return ErrInvalidObjectKey
		}
	}
	return nil
}
