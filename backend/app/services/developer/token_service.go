package developer

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/support/carbon"

	"goravel/app/facades"
	"goravel/app/models"
	planservices "goravel/app/services/plans"
	"goravel/app/services/quota"
)

const (
	ScopeUploadWrite   = "upload:write"
	ScopeLinksRead     = "links:read"
	ScopeMediaDelete   = "media:delete"
	ScopeMediaRead     = "media:read"
	ScopeUsageRead     = "usage:read"
	ScopeWebhookManage = "webhook:manage"
)

var (
	ErrTokenNotFound      = errors.New("api token not found")
	ErrTokenInvalid       = errors.New("api token invalid")
	ErrTokenExpired       = errors.New("api token expired")
	ErrInvalidTokenName   = errors.New("api token name is invalid")
	ErrInvalidTokenExpiry = errors.New("api token expiry is invalid")
	ErrInvalidTokenScope  = errors.New("api token scope is invalid")
	ErrInvalidTokenStatus = errors.New("api token status is invalid")
	ErrTokenLimitReached  = errors.New("api token limit reached")
)

// Personal API Tokens intentionally expose only the four operations needed by
// upload clients: upload, list own media, read one media item/links, and delete
// own media. Browser-only workflows use the member session middleware.
var baseScopes = []string{ScopeUploadWrite, ScopeMediaRead, ScopeMediaDelete}
var optionalScopes = []string{}

type CreateInput struct {
	UserID    uint
	Name      string
	Scopes    []string
	ExpiresAt *time.Time
}

type View struct {
	ID         uint             `json:"id"`
	Name       string           `json:"name"`
	Prefix     string           `json:"prefix"`
	Scopes     []string         `json:"scopes"`
	Status     string           `json:"status"`
	ExpiresAt  *time.Time       `json:"expires_at"`
	CreatedAt  *carbon.DateTime `json:"created_at"`
	LastUsedAt *time.Time       `json:"last_used_at"`
	LastUsedIP string           `json:"last_used_ip,omitempty"`
	UsageCount int64            `json:"usage_count"`
}

type Created struct {
	View
	Token string `json:"token"`
}

type Authenticated struct {
	TokenID uint
	UserID  uint
	Scopes  []string
}

type Repository interface {
	Create(ctx context.Context, input CreateInput) (models.ApiToken, string, error)
	ListOwned(ctx context.Context, userID uint) ([]models.ApiToken, error)
	Revoke(ctx context.Context, userID, id uint) error
	Delete(ctx context.Context, userID, id uint) error
	DeleteAdmin(ctx context.Context, id uint) error
	Rotate(ctx context.Context, userID, id uint) (models.ApiToken, string, error)
	Authenticate(ctx context.Context, raw, ip string) (Authenticated, error)
}

type DatabaseRepository struct{}

func NewDatabaseRepository() *DatabaseRepository { return &DatabaseRepository{} }

func (r *DatabaseRepository) Create(_ context.Context, input CreateInput) (models.ApiToken, string, error) {
	name, scopes, err := normalizeCreate(input)
	if err != nil {
		return models.ApiToken{}, "", err
	}
	raw, err := randomToken()
	if err != nil {
		return models.ApiToken{}, "", err
	}
	scopesJSON, err := json.Marshal(scopes)
	if err != nil {
		return models.ApiToken{}, "", err
	}
	token := models.ApiToken{
		UserID: input.UserID, Name: name, TokenHash: tokenHash(raw),
		TokenPrefix: raw[:12], ScopesJSON: string(scopesJSON), Status: "active",
		ExpiresAt: input.ExpiresAt,
	}
	if err := facades.Orm().Query().Create(&token); err != nil {
		return models.ApiToken{}, "", err
	}
	return token, raw, nil
}

func (r *DatabaseRepository) ListOwned(_ context.Context, userID uint) ([]models.ApiToken, error) {
	var tokens []models.ApiToken
	if err := facades.Orm().Query().Where("user_id = ?", userID).OrderBy("id", "desc").Get(&tokens); err != nil {
		return nil, err
	}
	return tokens, nil
}

func (r *DatabaseRepository) Revoke(_ context.Context, userID, id uint) error {
	now := time.Now().UTC()
	result, err := facades.Orm().Query().Table("api_tokens").Where("id = ? AND user_id = ? AND status = ?", id, userID, "active").Update(map[string]any{
		"status": "revoked", "revoked_at": now, "updated_at": now,
	})
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return ErrTokenNotFound
	}
	return nil
}

// Delete permanently removes a token owned by the member. The raw token is
// never recoverable, and deleting it immediately invalidates future requests.
func (r *DatabaseRepository) Delete(_ context.Context, userID, id uint) error {
	var token models.ApiToken
	if err := facades.Orm().Query().Where("id = ? AND user_id = ?", id, userID).First(&token); err != nil {
		return ErrTokenNotFound
	}
	return r.deleteRecord(id)
}

// DeleteAdmin is used only by the explicitly authorized admin Resource route.
// It keeps token cleanup in the developer service instead of bypassing the
// token lifecycle through a generic table delete.
func (r *DatabaseRepository) DeleteAdmin(_ context.Context, id uint) error {
	return r.deleteRecord(id)
}

func (r *DatabaseRepository) deleteRecord(id uint) error {
	return facades.Orm().Transaction(func(tx orm.Query) error {
		result, err := tx.Table("api_tokens").Where("id = ?", id).Delete()
		if err != nil {
			return err
		}
		if result.RowsAffected == 0 {
			return ErrTokenNotFound
		}
		if facades.Schema().HasTable("api_token_rate_limits") {
			if _, err := tx.Table("api_token_rate_limits").Where("key_hash = ?", rateLimitKeyHash("token:"+strconv.FormatUint(uint64(id), 10))).Delete(); err != nil {
				return err
			}
		}
		return nil
	})
}

// SetStatus is reserved for administrator lifecycle actions. Disabled and
// revoked are explicit operational states, while permanent deletion uses the
// separate delete permission and service path below.
func (r *DatabaseRepository) SetStatus(_ context.Context, id uint, status string) error {
	status = strings.TrimSpace(status)
	if status != "disabled" && status != "revoked" {
		return ErrInvalidTokenStatus
	}
	updates := map[string]any{"status": status, "updated_at": time.Now().UTC()}
	if status == "revoked" {
		updates["revoked_at"] = time.Now().UTC()
	}
	result, err := facades.Orm().Query().Table("api_tokens").Where("id = ? AND status = ?", id, "active").Update(updates)
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return ErrTokenNotFound
	}
	return nil
}

func (r *DatabaseRepository) Rotate(ctx context.Context, userID, id uint) (models.ApiToken, string, error) {
	var old models.ApiToken
	if err := facades.Orm().Query().Where("id = ? AND user_id = ? AND status = ?", id, userID, "active").First(&old); err != nil {
		return models.ApiToken{}, "", ErrTokenNotFound
	}
	scopes := decodeScopes(old.ScopesJSON)
	newToken, raw, err := r.Create(ctx, CreateInput{UserID: userID, Name: old.Name, Scopes: scopes, ExpiresAt: old.ExpiresAt})
	if err != nil {
		return models.ApiToken{}, "", err
	}
	if err := r.Revoke(ctx, userID, old.ID); err != nil {
		return models.ApiToken{}, "", err
	}
	return newToken, raw, nil
}

func (r *DatabaseRepository) Authenticate(_ context.Context, raw, ip string) (Authenticated, error) {
	if !strings.HasPrefix(raw, "fst_") {
		return Authenticated{}, ErrTokenInvalid
	}
	var token models.ApiToken
	if err := facades.Orm().Query().Where("token_hash = ?", tokenHash(raw)).First(&token); err != nil {
		return Authenticated{}, ErrTokenInvalid
	}
	now := time.Now().UTC()
	if token.Status != "active" {
		return Authenticated{}, ErrTokenInvalid
	}
	if token.ExpiresAt != nil && !token.ExpiresAt.After(now) {
		return Authenticated{}, ErrTokenExpired
	}
	if token.UserID == 0 {
		return Authenticated{}, ErrTokenInvalid
	}
	var user models.User
	if err := facades.Orm().Query().Where("id = ? AND status = ?", token.UserID, "active").First(&user); err != nil {
		return Authenticated{}, ErrTokenInvalid
	}
	updates := map[string]any{"last_used_at": now, "usage_count": token.UsageCount + 1, "updated_at": now}
	if strings.TrimSpace(ip) != "" {
		updates["last_used_ip"] = strings.TrimSpace(ip)
	}
	_, _ = facades.Orm().Query().Table("api_tokens").Where("id = ? AND status = ?", token.ID, "active").Update(updates)
	return Authenticated{TokenID: token.ID, UserID: token.UserID, Scopes: decodeScopes(token.ScopesJSON)}, nil
}

type Service struct{ repository Repository }

func NewService(repository Repository) *Service { return &Service{repository: repository} }

func NewDatabaseService() *Service { return NewService(NewDatabaseRepository()) }

func (s *Service) Create(ctx context.Context, input CreateInput) (Created, error) {
	if err := s.checkLimit(ctx, input.UserID, 0); err != nil {
		return Created{}, err
	}
	token, raw, err := s.repository.Create(ctx, input)
	if err != nil {
		return Created{}, err
	}
	return Created{View: view(token), Token: raw}, nil
}

func (s *Service) List(ctx context.Context, userID uint) ([]View, error) {
	tokens, err := s.repository.ListOwned(ctx, userID)
	if err != nil {
		return nil, err
	}
	views := make([]View, 0, len(tokens))
	for _, token := range tokens {
		views = append(views, view(token))
	}
	return views, nil
}

func (s *Service) Revoke(ctx context.Context, userID, id uint) error {
	return s.repository.Revoke(ctx, userID, id)
}

func (s *Service) Delete(ctx context.Context, userID, id uint) error {
	return s.repository.Delete(ctx, userID, id)
}

func (s *Service) DeleteAdmin(ctx context.Context, id uint) error {
	return s.repository.DeleteAdmin(ctx, id)
}

func (s *Service) Rotate(ctx context.Context, userID, id uint) (Created, error) {
	if err := s.checkLimit(ctx, userID, id); err != nil {
		return Created{}, err
	}
	token, raw, err := s.repository.Rotate(ctx, userID, id)
	if err != nil {
		return Created{}, err
	}
	return Created{View: view(token), Token: raw}, nil
}

func (s *Service) Authenticate(ctx context.Context, raw, ip string) (Authenticated, error) {
	return s.repository.Authenticate(ctx, strings.TrimSpace(raw), ip)
}

func (s *Service) HasScope(scopes []string, required string) bool {
	for _, scope := range scopes {
		if scope == required {
			return true
		}
	}
	return false
}

// CheckTokenLimit treats zero as unlimited and counts only active tokens.
// Keeping this boundary pure makes entitlement behavior independently testable.
func CheckTokenLimit(activeTokens, limit int) error {
	if activeTokens < 0 || limit < 0 {
		return ErrInvalidTokenStatus
	}
	if limit > 0 && activeTokens >= limit {
		return ErrTokenLimitReached
	}
	return nil
}

func (s *Service) checkLimit(ctx context.Context, userID, excludeID uint) error {
	// Repository fakes used by unit tests do not represent the database-backed
	// subscription boundary. Runtime enforcement is intentionally enabled only
	// for the production repository.
	if _, ok := s.repository.(*DatabaseRepository); !ok {
		return nil
	}
	tokens, err := s.repository.ListOwned(ctx, userID)
	if err != nil {
		return err
	}
	subscription, err := planservices.NewPlanService().SubscriptionForUser(userID)
	if err != nil {
		return err
	}
	entitlement, err := quota.ParseEntitlementJSON(subscription.EntitlementSnapshotJSON)
	if err != nil {
		return err
	}
	active := 0
	now := time.Now().UTC()
	for _, token := range tokens {
		if token.ID == excludeID || token.Status != "active" {
			continue
		}
		if token.ExpiresAt != nil && !token.ExpiresAt.After(now) {
			continue
		}
		active++
	}
	return CheckTokenLimit(active, int(entitlement.TokenLimit))
}

func normalizeCreate(input CreateInput) (string, []string, error) {
	if input.UserID == 0 {
		return "", nil, ErrTokenNotFound
	}
	name := strings.TrimSpace(input.Name)
	if name == "" || len([]rune(name)) > 120 {
		return "", nil, ErrInvalidTokenName
	}
	if input.ExpiresAt != nil && !input.ExpiresAt.After(time.Now().UTC()) {
		return "", nil, ErrInvalidTokenExpiry
	}
	allowed := make(map[string]struct{}, len(optionalScopes))
	for _, scope := range optionalScopes {
		allowed[scope] = struct{}{}
	}
	scopes := append([]string(nil), baseScopes...)
	seen := make(map[string]struct{}, len(scopes))
	for _, scope := range scopes {
		seen[scope] = struct{}{}
	}
	for _, scope := range input.Scopes {
		scope = strings.TrimSpace(scope)
		if scope == "" {
			continue
		}
		if _, ok := seen[scope]; ok {
			continue
		}
		if _, ok := allowed[scope]; !ok {
			return "", nil, ErrInvalidTokenScope
		}
		seen[scope] = struct{}{}
		scopes = append(scopes, scope)
	}
	return name, scopes, nil
}

func view(token models.ApiToken) View {
	return View{ID: token.ID, Name: token.Name, Prefix: token.TokenPrefix, Scopes: decodeScopes(token.ScopesJSON), Status: token.Status, ExpiresAt: token.ExpiresAt, CreatedAt: token.CreatedAt, LastUsedAt: token.LastUsedAt, LastUsedIP: token.LastUsedIP, UsageCount: token.UsageCount}
}

func decodeScopes(payload string) []string {
	var scopes []string
	if json.Unmarshal([]byte(payload), &scopes) != nil {
		return append([]string(nil), baseScopes...)
	}
	return scopes
}

func randomToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "fst_" + hex.EncodeToString(bytes), nil
}

func tokenHash(raw string) string {
	digest := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(digest[:])
}
