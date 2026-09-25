package collections

import (
	"context"
	"errors"
	"strings"
)

type Kind string

const (
	KindFolder Kind = "folders"
	KindAlbum  Kind = "albums"
)

var (
	ErrInvalidKind             = errors.New("invalid collection kind")
	ErrInvalidName             = errors.New("collection name is required")
	ErrInvalidVisibility       = errors.New("invalid album visibility")
	ErrParentNotFound          = errors.New("parent folder not found")
	ErrMediaNotFound           = errors.New("cover media not found")
	ErrNotFound                = errors.New("collection not found")
	ErrInvalidMediaIDs         = errors.New("album media ids are required")
	ErrAlbumMediaBatchTooLarge = errors.New("album media batch is too large")
	ErrAlbumMediaOrderInvalid  = errors.New("album media order must include every existing relation exactly once")
	ErrAlbumMoveInvalid        = errors.New("album move requires distinct owned albums")
)

const MaxAlbumMediaBatch = 100

type Item struct {
	ID           uint   `json:"id"`
	UserID       uint   `json:"user_id"`
	Name         string `json:"name"`
	ParentID     *int   `json:"parent_id,omitempty"`
	CoverMediaID *int   `json:"cover_media_id,omitempty"`
	Visibility   string `json:"visibility,omitempty"`
	MediaCount   int64  `json:"media_count"`
	CreatedAt    string `json:"created_at,omitempty"`
	UpdatedAt    string `json:"updated_at,omitempty"`
}

type Input struct {
	Name         string
	ParentID     *int   `json:"parent_id"`
	CoverMediaID *int   `json:"cover_media_id"`
	Visibility   string `json:"visibility"`
}

type AlbumMediaMutation struct {
	Added   []uint `json:"added_ids"`
	Skipped []uint `json:"skipped_ids"`
}

type AlbumMediaRemoval struct {
	Removed []uint `json:"removed_ids"`
	Skipped []uint `json:"skipped_ids"`
}

type AlbumMediaMove struct {
	Moved   []uint `json:"moved_ids"`
	Skipped []uint `json:"skipped_ids"`
}

type Repository interface {
	List(ctx context.Context, kind Kind, userID uint) ([]Item, error)
	Find(ctx context.Context, kind Kind, userID, id uint) (Item, error)
	ParentExists(ctx context.Context, userID, id uint) (bool, error)
	MediaExists(ctx context.Context, userID, id uint) (bool, error)
	Create(ctx context.Context, kind Kind, userID uint, input Input) (Item, error)
	Update(ctx context.Context, kind Kind, userID, id uint, input Input) (Item, error)
	Delete(ctx context.Context, kind Kind, userID, id uint) error
}

type AlbumMediaRepository interface {
	AddMediaToAlbum(ctx context.Context, userID, albumID uint, mediaIDs []uint) (AlbumMediaMutation, error)
	RemoveMediaFromAlbum(ctx context.Context, userID, albumID, mediaID uint) error
	ReorderAlbumMedia(ctx context.Context, userID, albumID uint, mediaIDs []uint) error
	MoveMediaBetweenAlbums(ctx context.Context, userID, sourceAlbumID, destinationAlbumID uint, mediaIDs []uint) (AlbumMediaMove, error)
	ListAlbumMediaIDs(ctx context.Context, userID, albumID uint) ([]uint, error)
}

type Service struct{ repository Repository }

func NewService(repository Repository) *Service { return &Service{repository: repository} }

func (s *Service) List(ctx context.Context, kind Kind, userID uint) ([]Item, error) {
	if err := validateKind(kind); err != nil {
		return nil, err
	}
	return s.repository.List(ctx, kind, userID)
}

func (s *Service) Create(ctx context.Context, kind Kind, userID uint, input Input) (Item, error) {
	input, err := normalizeInput(kind, input)
	if err != nil {
		return Item{}, err
	}
	if input.ParentID != nil && *input.ParentID > 0 {
		exists, err := s.repository.ParentExists(ctx, userID, uint(*input.ParentID))
		if err != nil {
			return Item{}, err
		}
		if !exists {
			return Item{}, ErrParentNotFound
		}
	}
	if input.CoverMediaID != nil && *input.CoverMediaID > 0 {
		exists, err := s.repository.MediaExists(ctx, userID, uint(*input.CoverMediaID))
		if err != nil {
			return Item{}, err
		}
		if !exists {
			return Item{}, ErrMediaNotFound
		}
	}
	return s.repository.Create(ctx, kind, userID, input)
}

func (s *Service) Update(ctx context.Context, kind Kind, userID, id uint, input Input) (Item, error) {
	if _, err := s.repository.Find(ctx, kind, userID, id); err != nil {
		return Item{}, err
	}
	return s.CreateOrUpdate(ctx, kind, userID, id, input)
}

func (s *Service) CreateOrUpdate(ctx context.Context, kind Kind, userID, id uint, input Input) (Item, error) {
	input, err := normalizeInput(kind, input)
	if err != nil {
		return Item{}, err
	}
	if input.ParentID != nil && *input.ParentID > 0 {
		exists, err := s.repository.ParentExists(ctx, userID, uint(*input.ParentID))
		if err != nil {
			return Item{}, err
		}
		if !exists {
			return Item{}, ErrParentNotFound
		}
	}
	if input.CoverMediaID != nil && *input.CoverMediaID > 0 {
		exists, err := s.repository.MediaExists(ctx, userID, uint(*input.CoverMediaID))
		if err != nil {
			return Item{}, err
		}
		if !exists {
			return Item{}, ErrMediaNotFound
		}
	}
	return s.repository.Update(ctx, kind, userID, id, input)
}

func (s *Service) Delete(ctx context.Context, kind Kind, userID, id uint) error {
	if err := validateKind(kind); err != nil {
		return err
	}
	return s.repository.Delete(ctx, kind, userID, id)
}

func (s *Service) AddMediaToAlbum(ctx context.Context, userID, albumID uint, mediaIDs []uint) (AlbumMediaMutation, error) {
	normalized, err := normalizeMediaIDs(mediaIDs)
	if err != nil {
		return AlbumMediaMutation{}, err
	}
	repository, ok := s.repository.(AlbumMediaRepository)
	if !ok {
		return AlbumMediaMutation{}, errors.New("album media association is not supported")
	}
	if _, err := s.repository.Find(ctx, KindAlbum, userID, albumID); err != nil {
		return AlbumMediaMutation{}, err
	}
	return repository.AddMediaToAlbum(ctx, userID, albumID, normalized)
}

func (s *Service) RemoveMediaFromAlbum(ctx context.Context, userID, albumID, mediaID uint) error {
	if mediaID == 0 {
		return ErrMediaNotFound
	}
	repository, ok := s.repository.(AlbumMediaRepository)
	if !ok {
		return errors.New("album media association is not supported")
	}
	if _, err := s.repository.Find(ctx, KindAlbum, userID, albumID); err != nil {
		return err
	}
	return repository.RemoveMediaFromAlbum(ctx, userID, albumID, mediaID)
}

func (s *Service) RemoveMediaFromAlbumBatch(ctx context.Context, userID, albumID uint, mediaIDs []uint) (AlbumMediaRemoval, error) {
	normalized, err := normalizeMediaIDs(mediaIDs)
	if err != nil {
		return AlbumMediaRemoval{}, err
	}
	repository, ok := s.repository.(AlbumMediaRepository)
	if !ok {
		return AlbumMediaRemoval{}, errors.New("album media association is not supported")
	}
	if _, err := s.repository.Find(ctx, KindAlbum, userID, albumID); err != nil {
		return AlbumMediaRemoval{}, err
	}
	mutation := AlbumMediaRemoval{}
	for _, mediaID := range normalized {
		if err := repository.RemoveMediaFromAlbum(ctx, userID, albumID, mediaID); err != nil {
			if errors.Is(err, ErrMediaNotFound) {
				mutation.Skipped = append(mutation.Skipped, mediaID)
				continue
			}
			return AlbumMediaRemoval{}, err
		}
		mutation.Removed = append(mutation.Removed, mediaID)
	}
	return mutation, nil
}

func (s *Service) ReorderAlbumMedia(ctx context.Context, userID, albumID uint, mediaIDs []uint) error {
	normalized, err := normalizeMediaIDs(mediaIDs)
	if err != nil {
		return err
	}
	repository, ok := s.repository.(AlbumMediaRepository)
	if !ok {
		return errors.New("album media association is not supported")
	}
	if _, err := s.repository.Find(ctx, KindAlbum, userID, albumID); err != nil {
		return err
	}
	return repository.ReorderAlbumMedia(ctx, userID, albumID, normalized)
}

func (s *Service) MoveMediaBetweenAlbums(ctx context.Context, userID, sourceAlbumID, destinationAlbumID uint, mediaIDs []uint) (AlbumMediaMove, error) {
	if sourceAlbumID == 0 || destinationAlbumID == 0 || sourceAlbumID == destinationAlbumID {
		return AlbumMediaMove{}, ErrAlbumMoveInvalid
	}
	normalized, err := normalizeMediaIDs(mediaIDs)
	if err != nil {
		return AlbumMediaMove{}, err
	}
	repository, ok := s.repository.(AlbumMediaRepository)
	if !ok {
		return AlbumMediaMove{}, errors.New("album media association is not supported")
	}
	if _, err := s.repository.Find(ctx, KindAlbum, userID, sourceAlbumID); err != nil {
		return AlbumMediaMove{}, err
	}
	if _, err := s.repository.Find(ctx, KindAlbum, userID, destinationAlbumID); err != nil {
		return AlbumMediaMove{}, err
	}
	return repository.MoveMediaBetweenAlbums(ctx, userID, sourceAlbumID, destinationAlbumID, normalized)
}

func (s *Service) ListAlbumMediaIDs(ctx context.Context, userID, albumID uint) ([]uint, error) {
	repository, ok := s.repository.(AlbumMediaRepository)
	if !ok {
		return nil, errors.New("album media association is not supported")
	}
	if _, err := s.repository.Find(ctx, KindAlbum, userID, albumID); err != nil {
		return nil, err
	}
	return repository.ListAlbumMediaIDs(ctx, userID, albumID)
}

func normalizeMediaIDs(mediaIDs []uint) ([]uint, error) {
	if len(mediaIDs) == 0 {
		return nil, ErrInvalidMediaIDs
	}
	seen := make(map[uint]struct{}, len(mediaIDs))
	result := make([]uint, 0, len(mediaIDs))
	for _, mediaID := range mediaIDs {
		if mediaID == 0 {
			return nil, ErrInvalidMediaIDs
		}
		if _, exists := seen[mediaID]; exists {
			continue
		}
		seen[mediaID] = struct{}{}
		result = append(result, mediaID)
	}
	if len(result) > MaxAlbumMediaBatch {
		return nil, ErrAlbumMediaBatchTooLarge
	}
	return result, nil
}

func normalizeInput(kind Kind, input Input) (Input, error) {
	if err := validateKind(kind); err != nil {
		return Input{}, err
	}
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" || len([]rune(input.Name)) > 120 {
		return Input{}, ErrInvalidName
	}
	if kind == KindAlbum {
		if input.Visibility == "" {
			input.Visibility = "private"
		}
		if input.Visibility != "private" && input.Visibility != "unlisted" && input.Visibility != "public" {
			return Input{}, ErrInvalidVisibility
		}
	} else {
		input.CoverMediaID = nil
		input.Visibility = ""
	}
	return input, nil
}

func validateKind(kind Kind) error {
	if kind != KindFolder && kind != KindAlbum {
		return ErrInvalidKind
	}
	return nil
}
