package developer

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"goravel/app/models"
)

type fakeTokenRepository struct {
	created models.ApiToken
	raw     string
	items   []models.ApiToken
}

func (f *fakeTokenRepository) Create(_ context.Context, input CreateInput) (models.ApiToken, string, error) {
	name, scopes, err := normalizeCreate(input)
	if err != nil {
		return models.ApiToken{}, "", err
	}
	encoded, _ := encodeScopes(scopes)
	f.created = models.ApiToken{UserID: input.UserID, Name: name, TokenPrefix: "fst_abcd", ScopesJSON: encoded, Status: "active", ExpiresAt: input.ExpiresAt}
	f.created.Model.ID = 9
	f.raw = "fst_raw-token"
	return f.created, f.raw, nil
}

func (f *fakeTokenRepository) ListOwned(context.Context, uint) ([]models.ApiToken, error) {
	return f.items, nil
}

func (f *fakeTokenRepository) Revoke(context.Context, uint, uint) error { return nil }

func (f *fakeTokenRepository) Rotate(context.Context, uint, uint) (models.ApiToken, string, error) {
	return f.created, f.raw, nil
}

func (f *fakeTokenRepository) Authenticate(context.Context, string, string) (Authenticated, error) {
	return Authenticated{}, nil
}

func encodeScopes(scopes []string) (string, error) {
	encoded, err := json.Marshal(scopes)
	return string(encoded), err
}

func TestCreateAlwaysIncludesBaseScopesAndReturnsRawTokenOnce(t *testing.T) {
	repository := &fakeTokenRepository{}
	created, err := NewService(repository).Create(context.Background(), CreateInput{UserID: 7, Name: "blog", Scopes: []string{ScopeMediaRead}})
	if err != nil {
		t.Fatal(err)
	}
	if created.Token == "" || created.Token != repository.raw {
		t.Fatalf("expected one-time raw token, got %q", created.Token)
	}
	if got := created.Scopes; len(got) != 3 || got[0] != ScopeUploadWrite || got[1] != ScopeMediaRead || got[2] != ScopeMediaDelete {
		t.Fatalf("base scopes missing: %#v", got)
	}
}

func TestCreateRejectsUnknownScopeAndPastExpiry(t *testing.T) {
	repository := &fakeTokenRepository{}
	service := NewService(repository)
	if _, err := service.Create(context.Background(), CreateInput{UserID: 7, Name: "blog", Scopes: []string{"admin:all"}}); err != ErrInvalidTokenScope {
		t.Fatalf("unknown scope error = %v", err)
	}
	past := time.Now().UTC().Add(-time.Minute)
	if _, err := service.Create(context.Background(), CreateInput{UserID: 7, Name: "blog", ExpiresAt: &past}); err != ErrInvalidTokenExpiry {
		t.Fatalf("past expiry error = %v", err)
	}
}

func TestCreateAcceptsPermanentExpiry(t *testing.T) {
	repository := &fakeTokenRepository{}
	created, err := NewService(repository).Create(context.Background(), CreateInput{UserID: 7, Name: "permanent", ExpiresAt: nil})
	if err != nil {
		t.Fatal(err)
	}
	if created.ExpiresAt != nil || repository.created.ExpiresAt != nil {
		t.Fatalf("permanent token must keep a null expiry: created=%v stored=%v", created.ExpiresAt, repository.created.ExpiresAt)
	}
}

func TestHasScopeOnlyMatchesExactScope(t *testing.T) {
	service := NewService(&fakeTokenRepository{})
	if !service.HasScope([]string{ScopeUploadWrite}, ScopeUploadWrite) || service.HasScope([]string{ScopeUploadWrite}, "upload") {
		t.Fatal("scope checks must be exact")
	}
}
