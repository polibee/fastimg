package feature

import (
	"errors"
	"fmt"
	"testing"

	"github.com/goravel/framework/contracts/database/orm"
	"goravel/app/facades"
	"goravel/app/models"
	mediaservices "goravel/app/services/media"
	"goravel/tests"
)

func TestPermanentMediaDeleteReleasesStorageOnlyAfterCleanupAndIsIdempotent(t *testing.T) {
	_ = tests.TestCase{}
	rollback := errors.New("rollback permanent media delete test")
	err := facades.Orm().Transaction(func(tx orm.Query) error {
		user := models.User{Name: "permanent-delete-test", Email: fmt.Sprintf("permanent-delete-%s@example.invalid", t.Name()), Status: "active"}
		if err := tx.Create(&user); err != nil {
			return err
		}
		asset := models.MediaAsset{UserID: user.ID, OriginalName: "sample.png", ContentType: "image/png", Format: "png", SizeBytes: 100, SHA256: "fixture", Status: "deleted"}
		if err := tx.Create(&asset); err != nil {
			return err
		}
		objects := []models.StorageObject{
			{Provider: "local", ObjectKey: fmt.Sprintf("fixture/%d/original.png", user.ID), ContentType: "image/png", SizeBytes: 100, SHA256: "original", Status: "ready"},
			{Provider: "local", ObjectKey: fmt.Sprintf("fixture/%d/thumbnail.png", user.ID), ContentType: "image/png", SizeBytes: 20, SHA256: "thumbnail", Status: "ready"},
		}
		for index := range objects {
			if err := tx.Create(&objects[index]); err != nil {
				return err
			}
			variant := models.MediaVariant{MediaAssetID: asset.ID, StorageObjectID: objects[index].ID, Name: fmt.Sprintf("variant-%d", index), Status: "ready"}
			if err := tx.Create(&variant); err != nil {
				return err
			}
		}
		if err := tx.Create(&models.UsageLedger{UserID: user.ID, ResourceType: "storage", Delta: 150, SourceType: "upload", SourceID: "fixture-upload", IdempotencyKey: fmt.Sprintf("fixture-upload-%d", user.ID), PeriodKey: "lifetime"}); err != nil {
			return err
		}

		prepared, cleanupObjects, err := mediaservices.PreparePermanentDeleteWithQuery(tx, user.ID, asset.ID)
		if err != nil {
			return err
		}
		if prepared.Status != "cleanup_pending" || len(cleanupObjects) != 2 {
			t.Fatalf("prepare returned status=%q and %d objects; want cleanup_pending and 2", prepared.Status, len(cleanupObjects))
		}
		if err := mediaservices.FinalizePermanentDeleteWithQuery(tx, user.ID, asset.ID); err != nil {
			return err
		}
		if err := mediaservices.FinalizePermanentDeleteWithQuery(tx, user.ID, asset.ID); err != nil {
			return err
		}
		var stored models.MediaAsset
		if err := tx.Find(&stored, asset.ID); err != nil {
			return err
		}
		if stored.Status != "physically_deleted" {
			t.Fatalf("media status = %q, want physically_deleted", stored.Status)
		}
		var balance int64
		if err := tx.Model(&models.UsageLedger{}).Where("user_id = ? AND resource_type = ?", user.ID, "storage").Sum("delta", &balance); err != nil {
			return err
		}
		if balance != 30 {
			t.Fatalf("storage balance = %d, want 30 after releasing 120 bytes", balance)
		}
		return rollback
	})
	if !errors.Is(err, rollback) {
		t.Fatalf("transaction error = %v, want intentional rollback", err)
	}
}
