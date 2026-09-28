package controllers

import (
	"errors"
	"testing"
)

func TestCanDeletePlanPriceRequiresArchivedAndUnreferenced(t *testing.T) {
	tests := []struct {
		name       string
		status     string
		referenced bool
		want       error
	}{
		{name: "active price must be archived first", status: "active", want: ErrPlanPriceDeleteRequiresArchive},
		{name: "archived historical price is protected", status: "archived", referenced: true, want: ErrPlanPriceDeleteProtected},
		{name: "archived unreferenced price can be deleted", status: "archived"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := canDeletePlanPrice(tt.status, tt.referenced); !errors.Is(err, tt.want) {
				t.Fatalf("canDeletePlanPrice() error = %v, want %v", err, tt.want)
			}
		})
	}
}
