package feature

import (
	"context"
	"testing"

	mediaservices "goravel/app/services/media"
	"goravel/tests"
)

func TestMediaRepositoryListOwnedSupportsDescendingSort(t *testing.T) {
	_ = tests.TestCase{}

	_, _, err := mediaservices.NewDatabaseRepository().ListOwned(context.Background(), 4_000_000_000, false, "", 1, 48)
	if err != nil {
		t.Fatalf("ListOwned returned an error for an empty owner result: %v", err)
	}
}
