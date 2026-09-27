package authservices

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/database/orm"

	"goravel/app/facades"
	"goravel/app/models"
	planservices "goravel/app/services/plans"
	settingsservices "goravel/app/services/settings"
)

var (
	ErrRegistrationInvalid      = errors.New("invalid registration")
	ErrRegistrationEmailTaken   = errors.New("registration email already exists")
	ErrRegistrationEmailBlocked = errors.New("registration email is not allowed")
	ErrVerificationInvalid      = errors.New("invalid or expired email verification token")
)

type RegistrationInput struct {
	Name     string
	Email    string
	Password string
}

type RegistrationResult struct {
	User                 *models.User
	VerificationToken    string
	VerificationRequired bool
}

type RegistrationService struct{}

func NewRegistrationService() *RegistrationService { return &RegistrationService{} }

func NormalizeRegistrationEmail(raw string) (string, error) {
	value := strings.ToLower(strings.TrimSpace(raw))
	parsed, err := mail.ParseAddress(value)
	if err != nil || parsed.Address != value || !strings.Contains(value, "@") {
		return "", ErrRegistrationInvalid
	}
	return value, nil
}

func (s *RegistrationService) Register(input RegistrationInput, policy RegistrationPolicy) (*RegistrationResult, error) {
	name := strings.TrimSpace(input.Name)
	if len([]rune(name)) < 2 || len([]rune(name)) > 80 || len(strings.TrimSpace(input.Password)) < 8 {
		return nil, ErrRegistrationInvalid
	}
	email, err := NormalizeRegistrationEmail(input.Email)
	if err != nil {
		return nil, err
	}
	if policy.EmailWhitelistEnabled {
		settings := settingsservices.NewSettingService()
		domains := ParseWhitelistDomains(settings.Resolve("auth.registration.email_whitelist_domains", ""))
		if len(domains) == 0 || !EmailAllowedByWhitelist(email, domains) {
			return nil, ErrRegistrationEmailBlocked
		}
	}
	exists, err := facades.Orm().Query().Where("email = ?", email).Exists()
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrRegistrationEmailTaken
	}
	hash, err := facades.Hash().Make(strings.TrimSpace(input.Password))
	if err != nil {
		return nil, err
	}
	user := &models.User{Name: name, Email: email, Password: hash, Locale: "zh-CN", Status: "active"}
	result := &RegistrationResult{User: user, VerificationRequired: policy.EmailVerificationRequired}
	if !policy.EmailVerificationRequired && facades.Schema().HasColumn("users", "email_verified_at") {
		now := time.Now().UTC()
		user.EmailVerifiedAt = &now
	}
	if err := facades.Orm().Transaction(func(tx orm.Query) error {
		if err := tx.Create(user); err != nil {
			return err
		}
		if user.EmailVerifiedAt != nil && facades.Schema().HasColumn("users", "email_verified_at") {
			if _, err := tx.Table("users").Where("id = ?", user.ID).Update("email_verified_at", user.EmailVerifiedAt); err != nil {
				return err
			}
		}
		if err := planservices.EnsureFreeSubscriptionWithQuery(tx, user.ID); err != nil {
			return err
		}
		if policy.EmailVerificationRequired {
			result.VerificationToken, err = s.createTokenWithQuery(tx, user.ID, policy.VerificationExpiryMinutes)
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *RegistrationService) createTokenWithQuery(query orm.Query, userID uint, expiryMinutes int) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	digest := sha256.Sum256([]byte(token))
	expiresAt := time.Now().UTC().Add(time.Duration(NormalizeVerificationExpiry(expiryMinutes)) * time.Minute)
	if err := query.Table("email_verification_tokens").Create(&map[string]any{
		"user_id": userID, "token_hash": hex.EncodeToString(digest[:]), "expires_at": expiresAt,
	}); err != nil {
		return "", err
	}
	return token, nil
}

func (s *RegistrationService) NewVerificationToken(userID uint, expiryMinutes int) (string, error) {
	return s.createTokenWithQuery(facades.Orm().Query(), userID, expiryMinutes)
}

func (s *RegistrationService) Verify(rawToken string) (*models.User, error) {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" || len(rawToken) > 256 {
		return nil, ErrVerificationInvalid
	}
	digest := sha256.Sum256([]byte(rawToken))
	hash := hex.EncodeToString(digest[:])
	var verification models.EmailVerificationToken
	if err := facades.Orm().Query().Where("token_hash = ? AND consumed_at IS NULL", hash).First(&verification); err != nil {
		return nil, ErrVerificationInvalid
	}
	now := time.Now().UTC()
	if verification.ExpiresAt.Before(now) {
		return nil, ErrVerificationInvalid
	}
	var user models.User
	if err := facades.Orm().Query().Where("id = ?", verification.UserID).First(&user); err != nil {
		return nil, ErrVerificationInvalid
	}
	if err := facades.Orm().Transaction(func(tx orm.Query) error {
		result, err := tx.Table("email_verification_tokens").Where("id = ? AND consumed_at IS NULL", verification.ID).Update("consumed_at", now)
		if err != nil {
			return err
		}
		if result.RowsAffected == 0 {
			return ErrVerificationInvalid
		}
		_, err = tx.Table("users").Where("id = ?", user.ID).Update("email_verified_at", now)
		return err
	}); err != nil {
		return nil, err
	}
	user.EmailVerifiedAt = &now
	return &user, nil
}
