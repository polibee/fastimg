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

func TestMonthlyTransformQuotaCountsCompletedAndInFlightJobsByOwner(t *testing.T) {
	_ = tests.TestCase{}
	const userID uint = 4_000_000_014
	const otherUserID uint = 4_000_000_015
	rollback := errors.New("rollback monthly transform quota test")
	now := time.Now().UTC()
	periodKey := now.Format("2006-01")

	err := facades.Orm().Transaction(func(tx orm.Query) error {
		if err := tx.Create(&models.UsageLedger{
			UserID: userID, ResourceType: "transform", Delta: 1, SourceType: "upload",
			SourceID: "completed-transform", IdempotencyKey: "test-transform-completed", PeriodKey: periodKey,
		}); err != nil {
			return err
		}
		for index, session := range []models.UploadSession{
			{UserID: userID, IdempotencyKey: "transform-processing", Status: "processing"},
			{UserID: userID, IdempotencyKey: "transform-failed", Status: "failed"},
			{UserID: otherUserID, IdempotencyKey: "transform-other-user", Status: "processing"},
		} {
			session.SHA256 = fmt.Sprintf("transform-test-%d", index)
			if err := tx.Create(&session); err != nil {
				return err
			}
		}
		if err := quota.CheckMonthlyTransformsWithQuery(tx, userID, now, 2); !errors.Is(err, quota.ErrQuotaExceeded) {
			t.Fatalf("completed+in-flight transforms returned %v, want quota exceeded", err)
		}
		if err := quota.CheckMonthlyTransformsWithQuery(tx, userID, now, 3); err != nil {
			t.Fatalf("limit above completed+in-flight transforms returned %v", err)
		}
		if err := quota.CheckMonthlyTransformsWithQuery(tx, userID, now, 0); err != nil {
			t.Fatalf("zero limit should mean unlimited, got %v", err)
		}
		return rollback
	})
	if !errors.Is(err, rollback) {
		t.Fatalf("transaction error = %v, want intentional rollback", err)
	}
}
