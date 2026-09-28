package services

import (
	"errors"
	"net"
	"net/url"
	"strings"
	"time"

	"goravel/app/facades"
	models "goravel/app/modules/friend_links/models"
)

var ErrNotFound = errors.New("friend link not found")
var ErrDuplicate = errors.New("friend link already submitted")

type SubmissionInput struct {
	SiteName     string `json:"site_name"`
	URL          string `json:"url"`
	LogoURL      string `json:"logo_url"`
	Description  string `json:"description"`
	ContactEmail string `json:"contact_email"`
}
type ReviewInput struct {
	Status     string `json:"status"`
	ReviewNote string `json:"review_note"`
}
type Service struct{}

func NewService() *Service { return &Service{} }

func validURL(raw string, optional bool) bool {
	raw = strings.TrimSpace(raw)
	if raw == "" && optional {
		return true
	}
	parsed, err := url.ParseRequestURI(raw)
	if err != nil || parsed == nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" || parsed.User != nil {
		return false
	}
	host := strings.ToLower(strings.TrimSuffix(parsed.Hostname(), "."))
	if host == "localhost" || strings.HasSuffix(host, ".local") || strings.HasSuffix(host, ".internal") || strings.HasSuffix(host, ".lan") {
		return false
	}
	if address := net.ParseIP(host); address != nil && (address.IsLoopback() || address.IsPrivate() || address.IsUnspecified() || address.IsLinkLocalUnicast()) {
		return false
	}
	return true
}

func (s *Service) List(status string) ([]models.Submission, error) {
	rows := make([]models.Submission, 0)
	query := facades.Orm().Query().OrderByDesc("created_at").OrderByDesc("id")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Get(&rows); err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *Service) Submit(input SubmissionInput, userID *uint) (models.Submission, error) {
	input.SiteName = strings.TrimSpace(input.SiteName)
	input.URL = strings.TrimSpace(input.URL)
	input.LogoURL = strings.TrimSpace(input.LogoURL)
	input.Description = strings.TrimSpace(input.Description)
	input.ContactEmail = strings.TrimSpace(input.ContactEmail)
	if input.SiteName == "" || !validURL(input.URL, false) || !validURL(input.LogoURL, true) {
		return models.Submission{}, errors.New("invalid friend link")
	}
	var existing []models.Submission
	if err := facades.Orm().Query().Where("url = ? AND status IN (?, ?)", input.URL, models.StatusPending, models.StatusApproved).Get(&existing); err != nil {
		return models.Submission{}, err
	}
	if len(existing) > 0 {
		return models.Submission{}, ErrDuplicate
	}
	now := time.Now().UTC()
	values := map[string]any{"site_name": input.SiteName, "url": input.URL, "logo_url": input.LogoURL, "description": input.Description, "contact_email": input.ContactEmail, "submitted_by": userID, "status": models.StatusPending, "created_at": now, "updated_at": now}
	if err := facades.Orm().Query().Table("friend_link_submissions").Create(&values); err != nil {
		return models.Submission{}, err
	}
	rows, err := s.List("")
	if err != nil || len(rows) == 0 {
		return models.Submission{}, err
	}
	return rows[0], nil
}

func (s *Service) Review(id uint, input ReviewInput, operatorID uint) (models.Submission, error) {
	if id == 0 || (input.Status != models.StatusApproved && input.Status != models.StatusRejected && input.Status != models.StatusHidden) {
		return models.Submission{}, errors.New("invalid review")
	}
	_, err := facades.Orm().Query().Table("friend_link_submissions").Where("id = ?", id).Update(map[string]any{"status": input.Status, "review_note": strings.TrimSpace(input.ReviewNote), "reviewed_by": operatorID, "reviewed_at": time.Now().UTC(), "updated_at": time.Now().UTC()})
	if err != nil {
		return models.Submission{}, ErrNotFound
	}
	rows, err := s.List("")
	if err != nil {
		return models.Submission{}, err
	}
	for _, row := range rows {
		if row.ID == id {
			return row, nil
		}
	}
	return models.Submission{}, ErrNotFound
}
