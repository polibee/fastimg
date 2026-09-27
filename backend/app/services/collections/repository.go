package collections

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"goravel/app/facades"

	"github.com/goravel/framework/contracts/database/orm"
)

type DatabaseRepository struct{}

func NewDatabaseRepository() *DatabaseRepository { return &DatabaseRepository{} }

func collectionTable(kind Kind) (string, error) {
	switch kind {
	case KindFolder:
		return "folders", nil
	case KindAlbum:
		return "albums", nil
	default:
		return "", ErrInvalidKind
	}
}

func (r *DatabaseRepository) List(_ context.Context, kind Kind, userID uint) ([]Item, error) {
	table, err := collectionTable(kind)
	if err != nil {
		return nil, err
	}
	var rows []map[string]any
	if err := facades.Orm().Query().Table(table).Where("user_id = ?", userID).OrderBy("id", "desc").Get(&rows); err != nil {
		return nil, err
	}
	items := make([]Item, 0, len(rows))
	for _, row := range rows {
		item := itemFromRow(kind, row)
		if kind == KindAlbum {
			item.MediaCount, err = facades.Orm().Query().Table("album_media").Where("album_id = ? AND user_id = ?", item.ID, userID).Count()
			if err != nil {
				return nil, err
			}
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *DatabaseRepository) FindPublicAlbum(_ context.Context, albumID uint) (PublicAlbum, error) {
	var albumRows []map[string]any
	if err := facades.Orm().Query().Table("albums").Where("id = ? AND visibility = ?", albumID, "public").Get(&albumRows); err != nil {
		return PublicAlbum{}, err
	}
	if len(albumRows) == 0 {
		return PublicAlbum{}, ErrNotFound
	}
	album := itemFromRow(KindAlbum, albumRows[0])

	var relations []map[string]any
	if err := facades.Orm().Query().Table("album_media").Where("album_id = ? AND user_id = ?", album.ID, album.UserID).OrderBy("sort_order", "asc").OrderBy("id", "asc").Get(&relations); err != nil {
		return PublicAlbum{}, err
	}
	mediaIDs := make([]any, 0, len(relations))
	for _, relation := range relations {
		mediaIDs = append(mediaIDs, rowUint(relation["media_asset_id"]))
	}
	mediaByID := make(map[uint]map[string]any, len(mediaIDs))
	if len(mediaIDs) > 0 {
		var mediaRows []map[string]any
		if err := facades.Orm().Query().Table("media_assets").Where("user_id = ? AND status = ? AND deleted_at IS NULL AND visibility = ? AND moderation_status <> ?", album.UserID, "ready", "public", "rejected").WhereIn("id", mediaIDs).Get(&mediaRows); err != nil {
			return PublicAlbum{}, err
		}
		for _, media := range mediaRows {
			mediaByID[rowUint(media["id"])] = media
		}
	}
	media := make([]PublicAlbumMedia, 0, len(mediaByID))
	for _, relation := range relations {
		mediaRow, ok := mediaByID[rowUint(relation["media_asset_id"])]
		if !ok {
			continue
		}
		media = append(media, PublicAlbumMedia{
			ID:           rowUint(mediaRow["id"]),
			OriginalName: rowString(mediaRow["original_name"]),
			ContentType:  rowString(mediaRow["content_type"]),
			Width:        rowInt64(mediaRow["width"]),
			Height:       rowInt64(mediaRow["height"]),
			SizeBytes:    rowInt64(mediaRow["size_bytes"]),
			CreatedAt:    rowString(mediaRow["created_at"]),
		})
	}
	return PublicAlbum{ID: album.ID, Name: album.Name, Visibility: album.Visibility, Media: media, CreatedAt: album.CreatedAt, UpdatedAt: album.UpdatedAt}, nil
}

func (r *DatabaseRepository) Find(_ context.Context, kind Kind, userID, id uint) (Item, error) {
	table, err := collectionTable(kind)
	if err != nil {
		return Item{}, err
	}
	var rows []map[string]any
	if err := facades.Orm().Query().Table(table).Where("id = ? AND user_id = ?", id, userID).Get(&rows); err != nil {
		return Item{}, ErrNotFound
	}
	if len(rows) == 0 {
		return Item{}, ErrNotFound
	}
	return itemFromRow(kind, rows[0]), nil
}

func (r *DatabaseRepository) ParentExists(_ context.Context, userID, id uint) (bool, error) {
	return facades.Orm().Query().Table("folders").Where("id = ? AND user_id = ?", id, userID).Exists()
}

func (r *DatabaseRepository) MediaExists(_ context.Context, userID, id uint) (bool, error) {
	return facades.Orm().Query().Table("media_assets").Where("id = ? AND user_id = ? AND status = ?", id, userID, "ready").Exists()
}

func (r *DatabaseRepository) Create(_ context.Context, kind Kind, userID uint, input Input) (Item, error) {
	table, err := collectionTable(kind)
	if err != nil {
		return Item{}, err
	}
	columns := []string{"name", "user_id"}
	args := []any{input.Name, userID}
	if kind == KindFolder {
		columns = append(columns, "parent_id")
		args = append(args, input.ParentID)
	} else {
		columns = append(columns, "cover_media_id", "visibility")
		args = append(args, input.CoverMediaID, input.Visibility)
	}
	return r.insert(table, columns, args, kind, userID)
}

func (r *DatabaseRepository) Update(_ context.Context, kind Kind, userID, id uint, input Input) (Item, error) {
	table, err := collectionTable(kind)
	if err != nil {
		return Item{}, err
	}
	values := map[string]any{"name": input.Name, "updated_at": time.Now().UTC()}
	if kind == KindFolder {
		values["parent_id"] = input.ParentID
	} else {
		values["cover_media_id"] = input.CoverMediaID
		values["visibility"] = input.Visibility
	}
	result, err := facades.Orm().Query().Table(table).Where("id = ? AND user_id = ?", id, userID).Update(values)
	if err != nil {
		return Item{}, err
	}
	if result.RowsAffected == 0 {
		return Item{}, ErrNotFound
	}
	return r.Find(context.Background(), kind, userID, id)
}

func (r *DatabaseRepository) Delete(_ context.Context, kind Kind, userID, id uint) error {
	table, err := collectionTable(kind)
	if err != nil {
		return err
	}
	result, err := facades.Orm().Query().Table(table).Where("id = ? AND user_id = ?", id, userID).Delete()
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *DatabaseRepository) AddMediaToAlbum(_ context.Context, userID, albumID uint, mediaIDs []uint) (mutation AlbumMediaMutation, err error) {
	err = facades.Orm().Transaction(func(tx orm.Query) error {
		if exists, checkErr := tx.Table("albums").Where("id = ? AND user_id = ?", albumID, userID).Exists(); checkErr != nil {
			return checkErr
		} else if !exists {
			return ErrNotFound
		}
		var albumRows []map[string]any
		if err := tx.Table("albums").Where("id = ? AND user_id = ?", albumID, userID).LockForUpdate().Get(&albumRows); err != nil {
			return ErrNotFound
		}
		if len(albumRows) == 0 {
			return ErrNotFound
		}

		mediaArgs := uintArgs(mediaIDs)
		var mediaRows []map[string]any
		if err := tx.Table("media_assets").Where("user_id = ? AND status = ?", userID, "ready").WhereIn("id", mediaArgs).Get(&mediaRows); err != nil {
			return err
		}
		if len(mediaRows) != len(mediaIDs) {
			return ErrMediaNotFound
		}

		var existingRows []map[string]any
		if err := tx.Table("album_media").Where("album_id = ? AND user_id = ?", albumID, userID).WhereIn("media_asset_id", mediaArgs).Get(&existingRows); err != nil {
			return err
		}
		existing := make(map[uint]struct{}, len(existingRows))
		for _, row := range existingRows {
			existing[rowUint(row["media_asset_id"])] = struct{}{}
		}
		for _, mediaID := range mediaIDs {
			if _, alreadyAdded := existing[mediaID]; alreadyAdded {
				mutation.Skipped = append(mutation.Skipped, mediaID)
				continue
			}
			if err := tx.Table("album_media").Create(&map[string]any{
				"album_id":       albumID,
				"media_asset_id": mediaID,
				"user_id":        userID,
				"sort_order":     0,
			}); err != nil {
				return err
			}
			mutation.Added = append(mutation.Added, mediaID)
		}
		return nil
	})
	return mutation, err
}

func (r *DatabaseRepository) RemoveMediaFromAlbum(_ context.Context, userID, albumID, mediaID uint) error {
	if exists, err := facades.Orm().Query().Table("albums").Where("id = ? AND user_id = ?", albumID, userID).Exists(); err != nil {
		return err
	} else if !exists {
		return ErrNotFound
	}
	result, err := facades.Orm().Query().Table("album_media").Where("album_id = ? AND user_id = ? AND media_asset_id = ?", albumID, userID, mediaID).Delete()
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return ErrMediaNotFound
	}
	return nil
}

func (r *DatabaseRepository) ReorderAlbumMedia(_ context.Context, userID, albumID uint, mediaIDs []uint) error {
	return facades.Orm().Transaction(func(tx orm.Query) error {
		if exists, err := tx.Table("albums").Where("id = ? AND user_id = ?", albumID, userID).Exists(); err != nil {
			return err
		} else if !exists {
			return ErrNotFound
		}
		var rows []map[string]any
		if err := tx.Table("album_media").Where("album_id = ? AND user_id = ?", albumID, userID).LockForUpdate().Get(&rows); err != nil {
			return err
		}
		if len(rows) != len(mediaIDs) {
			return ErrAlbumMediaOrderInvalid
		}
		existing := make(map[uint]struct{}, len(rows))
		for _, row := range rows {
			existing[rowUint(row["media_asset_id"])] = struct{}{}
		}
		for _, mediaID := range mediaIDs {
			if _, ok := existing[mediaID]; !ok {
				return ErrAlbumMediaOrderInvalid
			}
			delete(existing, mediaID)
		}
		if len(existing) != 0 {
			return ErrAlbumMediaOrderInvalid
		}
		now := time.Now().UTC()
		for order, mediaID := range mediaIDs {
			result, err := tx.Table("album_media").Where("album_id = ? AND user_id = ? AND media_asset_id = ?", albumID, userID, mediaID).Update(map[string]any{
				"sort_order": order,
				"updated_at": now,
			})
			if err != nil {
				return err
			}
			if result.RowsAffected == 0 {
				return ErrAlbumMediaOrderInvalid
			}
		}
		return nil
	})
}

func (r *DatabaseRepository) MoveMediaBetweenAlbums(_ context.Context, userID, sourceAlbumID, destinationAlbumID uint, mediaIDs []uint) (mutation AlbumMediaMove, err error) {
	err = facades.Orm().Transaction(func(tx orm.Query) error {
		albumIDs := []uint{sourceAlbumID, destinationAlbumID}
		sort.Slice(albumIDs, func(left, right int) bool { return albumIDs[left] < albumIDs[right] })
		for _, albumID := range albumIDs {
			if exists, checkErr := tx.Table("albums").Where("id = ? AND user_id = ?", albumID, userID).Exists(); checkErr != nil {
				return checkErr
			} else if !exists {
				return ErrNotFound
			}
			var locked []map[string]any
			if err := tx.Table("albums").Where("id = ? AND user_id = ?", albumID, userID).LockForUpdate().Get(&locked); err != nil || len(locked) == 0 {
				if err != nil {
					return err
				}
				return ErrNotFound
			}
		}

		mediaArgs := uintArgs(mediaIDs)
		var sourceRows []map[string]any
		if err := tx.Table("album_media").Where("album_id = ? AND user_id = ?", sourceAlbumID, userID).WhereIn("media_asset_id", mediaArgs).LockForUpdate().Get(&sourceRows); err != nil {
			return err
		}
		var destinationRows []map[string]any
		if err := tx.Table("album_media").Where("album_id = ? AND user_id = ?", destinationAlbumID, userID).WhereIn("media_asset_id", mediaArgs).LockForUpdate().Get(&destinationRows); err != nil {
			return err
		}
		source := make(map[uint]struct{}, len(sourceRows))
		for _, row := range sourceRows {
			source[rowUint(row["media_asset_id"])] = struct{}{}
		}
		destination := make(map[uint]struct{}, len(destinationRows))
		for _, row := range destinationRows {
			destination[rowUint(row["media_asset_id"])] = struct{}{}
		}
		destinationCount, err := tx.Table("album_media").Where("album_id = ? AND user_id = ?", destinationAlbumID, userID).Count()
		if err != nil {
			return err
		}
		now := time.Now().UTC()
		for _, mediaID := range mediaIDs {
			if _, exists := source[mediaID]; !exists {
				mutation.Skipped = append(mutation.Skipped, mediaID)
				continue
			}
			if _, exists := destination[mediaID]; exists {
				mutation.Skipped = append(mutation.Skipped, mediaID)
				continue
			}
			if err := tx.Table("album_media").Create(&map[string]any{
				"album_id":       destinationAlbumID,
				"media_asset_id": mediaID,
				"user_id":        userID,
				"sort_order":     int(destinationCount),
				"created_at":     now,
				"updated_at":     now,
			}); err != nil {
				return err
			}
			if _, err := tx.Table("album_media").Where("album_id = ? AND user_id = ? AND media_asset_id = ?", sourceAlbumID, userID, mediaID).Delete(); err != nil {
				return err
			}
			destinationCount++
			mutation.Moved = append(mutation.Moved, mediaID)
		}
		return nil
	})
	return mutation, err
}

func (r *DatabaseRepository) ListAlbumMediaIDs(_ context.Context, userID, albumID uint) ([]uint, error) {
	if exists, err := facades.Orm().Query().Table("albums").Where("id = ? AND user_id = ?", albumID, userID).Exists(); err != nil {
		return nil, err
	} else if !exists {
		return nil, ErrNotFound
	}
	var rows []map[string]any
	if err := facades.Orm().Query().Table("album_media").Where("album_id = ? AND user_id = ?", albumID, userID).OrderBy("sort_order", "asc").OrderBy("id", "desc").Get(&rows); err != nil {
		return nil, err
	}
	mediaIDs := make([]uint, 0, len(rows))
	for _, row := range rows {
		mediaIDs = append(mediaIDs, rowUint(row["media_asset_id"]))
	}
	return mediaIDs, nil
}

func (r *DatabaseRepository) ListAdminAlbumMedia(_ context.Context, albumID uint) ([]AdminAlbumMediaItem, error) {
	var albumRows []map[string]any
	if err := facades.Orm().Query().Table("albums").Where("id = ?", albumID).Get(&albumRows); err != nil || len(albumRows) == 0 {
		return nil, ErrNotFound
	}
	var relations []map[string]any
	if err := facades.Orm().Query().Table("album_media").Where("album_id = ?", albumID).OrderBy("sort_order", "asc").OrderBy("id", "asc").Get(&relations); err != nil {
		return nil, err
	}
	if len(relations) == 0 {
		return []AdminAlbumMediaItem{}, nil
	}
	mediaIDs := make([]any, 0, len(relations))
	for _, row := range relations {
		mediaIDs = append(mediaIDs, rowUint(row["media_asset_id"]))
	}
	var mediaRows []map[string]any
	if err := facades.Orm().Query().Table("media_assets").Where("status = ?", "ready").WhereIn("id", mediaIDs).Get(&mediaRows); err != nil {
		return nil, err
	}
	byID := make(map[uint]map[string]any, len(mediaRows))
	for _, row := range mediaRows {
		byID[rowUint(row["id"])] = row
	}
	items := make([]AdminAlbumMediaItem, 0, len(relations))
	for _, relation := range relations {
		mediaID := rowUint(relation["media_asset_id"])
		media, ok := byID[mediaID]
		if !ok {
			continue
		}
		items = append(items, AdminAlbumMediaItem{ID: mediaID, OriginalName: fmt.Sprint(media["original_name"]), UserID: rowUint(media["user_id"]), Status: fmt.Sprint(media["status"]), Visibility: fmt.Sprint(media["visibility"]), SortOrder: int(rowUint(relation["sort_order"]))})
	}
	return items, nil
}

func (r *DatabaseRepository) AddAdminMediaToAlbum(_ context.Context, albumID uint, mediaIDs []uint) (mutation AdminAlbumMediaMutation, err error) {
	err = facades.Orm().Transaction(func(tx orm.Query) error {
		var albums []map[string]any
		if err := tx.Table("albums").Where("id = ?", albumID).LockForUpdate().Get(&albums); err != nil || len(albums) == 0 {
			return ErrNotFound
		}
		ownerID := rowUint(albums[0]["user_id"])
		mediaArgs := uintArgs(mediaIDs)
		var mediaRows []map[string]any
		if err := tx.Table("media_assets").Where("user_id = ? AND status = ?", ownerID, "ready").WhereIn("id", mediaArgs).Get(&mediaRows); err != nil {
			return err
		}
		if len(mediaRows) != len(mediaIDs) {
			return ErrMediaNotFound
		}
		var existing []map[string]any
		if err := tx.Table("album_media").Where("album_id = ?", albumID).WhereIn("media_asset_id", mediaArgs).Get(&existing); err != nil {
			return err
		}
		seen := make(map[uint]struct{}, len(existing))
		for _, row := range existing {
			seen[rowUint(row["media_asset_id"])] = struct{}{}
		}
		for _, mediaID := range mediaIDs {
			if _, ok := seen[mediaID]; ok {
				continue
			}
			if err := tx.Table("album_media").Create(&map[string]any{"album_id": albumID, "media_asset_id": mediaID, "user_id": ownerID, "sort_order": 0}); err != nil {
				return err
			}
			mutation.Changed = append(mutation.Changed, mediaID)
		}
		return nil
	})
	return mutation, err
}

func (r *DatabaseRepository) RemoveAdminMediaFromAlbum(_ context.Context, albumID uint, mediaIDs []uint) (mutation AdminAlbumMediaMutation, err error) {
	if exists, checkErr := facades.Orm().Query().Table("albums").Where("id = ?", albumID).Exists(); checkErr != nil {
		return mutation, checkErr
	} else if !exists {
		return mutation, ErrNotFound
	}
	for _, mediaID := range mediaIDs {
		result, deleteErr := facades.Orm().Query().Table("album_media").Where("album_id = ? AND media_asset_id = ?", albumID, mediaID).Delete()
		if deleteErr != nil {
			return mutation, deleteErr
		}
		if result.RowsAffected > 0 {
			mutation.Changed = append(mutation.Changed, mediaID)
		}
	}
	return mutation, nil
}

func (r *DatabaseRepository) insert(table string, columns []string, args []any, kind Kind, userID uint) (Item, error) {
	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	statement := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING id", table, joinColumns(columns), joinColumns(placeholders))
	var row struct {
		ID uint `gorm:"column:id"`
	}
	if err := facades.Orm().Query().Raw(statement, args...).Scan(&row); err != nil {
		return Item{}, err
	}
	if row.ID == 0 {
		return Item{}, ErrNotFound
	}
	return r.Find(context.Background(), kind, userID, row.ID)
}

func joinColumns(values []string) string {
	result := ""
	for i, value := range values {
		if i > 0 {
			result += ", "
		}
		result += value
	}
	return result
}

func itemFromRow(kind Kind, row map[string]any) Item {
	item := Item{ID: rowUint(row["id"]), UserID: rowUint(row["user_id"]), Name: fmt.Sprint(row["name"]), CreatedAt: rowString(row["created_at"]), UpdatedAt: rowString(row["updated_at"])}
	if kind == KindFolder {
		item.ParentID = rowIntPtr(row["parent_id"])
	} else {
		item.CoverMediaID = rowIntPtr(row["cover_media_id"])
		item.Visibility = fmt.Sprint(row["visibility"])
	}
	return item
}

func rowUint(value any) uint {
	parsed, _ := strconv.ParseUint(fmt.Sprint(value), 10, 64)
	return uint(parsed)
}

func rowInt64(value any) int64 {
	parsed, _ := strconv.ParseInt(fmt.Sprint(value), 10, 64)
	return parsed
}

func rowIntPtr(value any) *int {
	if value == nil || fmt.Sprint(value) == "<nil>" || fmt.Sprint(value) == "" {
		return nil
	}
	parsed, err := strconv.Atoi(fmt.Sprint(value))
	if err != nil {
		return nil
	}
	return &parsed
}

func rowString(value any) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}

func uintArgs(values []uint) []any {
	args := make([]any, len(values))
	for index, value := range values {
		args[index] = value
	}
	return args
}
