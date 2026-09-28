package content

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"goravel/app/facades"
	contentmodels "goravel/app/modules/content/models"
)

var (
	ErrPageNotFound = errors.New("site page not found")
	ErrInvalidPage  = errors.New("site page is invalid")
)

var pageSlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type SaveDraftInput struct {
	ID             uint
	Slug           string
	Title          string
	Content        map[string]any
	Excerpt        string
	SEOTitle       string
	SEODescription string
	OperatorID     uint
}

type Service struct{}

func NewService() *Service { return &Service{} }

func (s *Service) GetPublished(_ context.Context, slug string) (contentmodels.SitePage, error) {
	var page contentmodels.SitePage
	slug = strings.ToLower(strings.TrimSpace(slug))
	if !pageSlugPattern.MatchString(slug) {
		return page, ErrPageNotFound
	}
	if err := facades.Orm().Query().Where("slug = ? AND status = ?", slug, contentmodels.StatusPublished).First(&page); err != nil {
		return page, ErrPageNotFound
	}
	return page, nil
}

func (s *Service) Get(_ context.Context, id uint) (contentmodels.SitePage, error) {
	var page contentmodels.SitePage
	if id == 0 || facades.Orm().Query().Where("id = ?", id).First(&page) != nil {
		return page, ErrPageNotFound
	}
	return page, nil
}

func (s *Service) SaveDraft(_ context.Context, input SaveDraftInput) (contentmodels.SitePage, error) {
	page := contentmodels.SitePage{}
	if input.ID > 0 {
		var err error
		page, err = s.Get(context.Background(), input.ID)
		if err != nil {
			return page, err
		}
	}
	clean, err := SanitizeDocument(input.Content)
	if err != nil {
		return page, fmt.Errorf("%w: %v", ErrInvalidPage, err)
	}
	if !pageSlugPattern.MatchString(strings.ToLower(strings.TrimSpace(input.Slug))) || strings.TrimSpace(input.Title) == "" {
		return page, ErrInvalidPage
	}
	contentJSON, err := json.Marshal(clean)
	if err != nil {
		return page, fmt.Errorf("%w: encode content", ErrInvalidPage)
	}
	now := time.Now().UTC()
	values := map[string]any{
		"slug": input.Slug, "title": input.Title, "content_json": string(contentJSON),
		"excerpt": input.Excerpt, "seo_title": input.SEOTitle, "seo_description": input.SEODescription,
		"status": contentmodels.StatusDraft, "updated_by": input.OperatorID, "updated_at": now,
	}
	if input.ID == 0 {
		values["created_by"] = input.OperatorID
		values["created_at"] = now
		if err := facades.Orm().Query().Table("site_pages").Create(&values); err != nil {
			return page, err
		}
		if err := facades.Orm().Query().Where("slug = ?", input.Slug).First(&page); err != nil {
			return page, err
		}
		return page, nil
	}
	if _, err := facades.Orm().Query().Table("site_pages").Where("id = ?", input.ID).Update(values); err != nil {
		return page, err
	}
	return s.Get(context.Background(), input.ID)
}

func (s *Service) Publish(_ context.Context, id uint, operatorID uint) (contentmodels.SitePage, error) {
	page, err := s.Get(context.Background(), id)
	if err != nil {
		return page, err
	}
	now := time.Now().UTC()
	if _, err := facades.Orm().Query().Table("site_pages").Where("id = ?", id).Update(map[string]any{"status": contentmodels.StatusPublished, "published_at": now, "updated_by": operatorID, "updated_at": now}); err != nil {
		return page, err
	}
	return s.Get(context.Background(), id)
}

func (s *Service) Archive(_ context.Context, id uint, operatorID uint) (contentmodels.SitePage, error) {
	page, err := s.Get(context.Background(), id)
	if err != nil {
		return page, err
	}
	now := time.Now().UTC()
	if _, err := facades.Orm().Query().Table("site_pages").Where("id = ?", id).Update(map[string]any{"status": contentmodels.StatusArchived, "updated_by": operatorID, "updated_at": now}); err != nil {
		return page, err
	}
	return s.Get(context.Background(), id)
}
