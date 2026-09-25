package feature

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/goravel/framework/contracts/database/orm"

	"goravel/app/facades"
	"goravel/app/models"
	"goravel/app/services/quota"
	"goravel/tests"
)

func TestMonthlyAPIUploadQuotaCountsReadyAndProcessingSessions(t *testing.T) {
	_ = tests.TestCase{}
	const userID uint = 4_000_000_012
	const otherUserID uint = 4_000_000_013
	rollback := errors.New("rollback monthly upload quota test")
	monthStart := time.Now().UTC()

	err := facades.Orm().Transaction(func(tx orm.Query) error {
		for index, session := range []models.UploadSession{
			{UserID: userID, IdempotencyKey: "monthly-ready", Status: "ready"},
			{UserID: userID, IdempotencyKey: "monthly-processing", Status: "processing"},
			{UserID: userID, IdempotencyKey: "monthly-failed", Status: "failed"},
			{UserID: otherUserID, IdempotencyKey: "monthly-other-user", Status: "ready"},
		} {
			session.SHA256 = fmt.Sprintf("test-%d", index)
			if err := tx.Create(&session); err != nil {
				return err
			}
		}

		if err := quota.CheckMonthlyAPIUploadsWithQuery(tx, userID, monthStart, 2); !errors.Is(err, quota.ErrQuotaExceeded) {
			t.Fatalf("limit at completed+reserved count returned %v, want quota exceeded", err)
		}
		if err := quota.CheckMonthlyAPIUploadsWithQuery(tx, userID, monthStart, 3); err != nil {
			t.Fatalf("limit above completed+reserved count returned %v", err)
		}
		if err := quota.CheckMonthlyAPIUploadsWithQuery(tx, userID, monthStart, 0); err != nil {
			t.Fatalf("zero limit should mean unlimited, got %v", err)
		}
		return rollback
	})
	if !errors.Is(err, rollback) {
		t.Fatalf("transaction error = %v, want intentional rollback", err)
	}
}
