package services

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"goravel/app/facades"
	contentservices "goravel/app/modules/content/services"
	models "goravel/app/modules/footer_navigation/models"
)

var ErrNotFound = errors.New("footer navigation item not found")

type GroupInput struct {
	Title     string `json:"title"`
	Locale    string `json:"locale"`
	SortOrder int    `json:"sort_order"`
	Enabled   bool   `json:"is_enabled"`
}
type ItemInput struct {
	GroupID      uint   `json:"group_id"`
	ParentID     *uint  `json:"parent_id"`
	Label        string `json:"label"`
	TargetType   string `json:"target_type"`
	TargetValue  string `json:"target_value"`
	OpenInNewTab bool   `json:"open_in_new_tab"`
	SortOrder    int    `json:"sort_order"`
	Enabled      bool   `json:"is_enabled"`
}
type GroupResult struct {
	Group models.NavigationGroup  `json:"group"`
	Items []models.NavigationItem `json:"items"`
}

func validTargetType(value string) bool {
	return value == "page" || value == "route" || value == "external" || value == "friends"
}

var pageTargetPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func validateTarget(targetType, targetValue string) error {
	switch targetType {
	case "page":
		if !pageTargetPattern.MatchString(targetValue) {
			return errors.New("page target must be a slug")
		}
	case "route":
		if !strings.HasPrefix(targetValue, "/") || strings.HasPrefix(targetValue, "//") || strings.ContainsAny(targetValue, "\r\n") {
			return errors.New("route target must be a local path")
		}
	case "external":
		if err := contentservices.ValidateURL(targetValue, false); err != nil {
			return errors.New("external target URL is invalid")
		}
	case "friends":
		if targetValue != "friends" {
			return errors.New("friend-links target must be friends")
		}
	}
	return nil
}

type Service struct{}

func NewService() *Service { return &Service{} }

func (s *Service) List(locale string, publicOnly bool) ([]GroupResult, error) {
	groups := make([]models.NavigationGroup, 0)
	query := facades.Orm().Query().Table("footer_navigation_groups").OrderBy("sort_order").OrderBy("id")
	if publicOnly {
		query = query.Where("is_enabled = ?", true)
	}
	if locale = strings.TrimSpace(locale); locale != "" && locale != "all" {
		query = query.Where("locale IN (?, ?)", "all", locale)
	}
	if err := query.Get(&groups); err != nil {
		return nil, err
	}
	result := make([]GroupResult, 0, len(groups))
	for _, group := range groups {
		items := make([]models.NavigationItem, 0)
		itemQuery := facades.Orm().Query().Table("footer_navigation_items").Where("group_id = ?", group.ID).OrderBy("sort_order").OrderBy("id")
		if publicOnly {
			itemQuery = itemQuery.Where("is_enabled = ?", true)
		}
		if err := itemQuery.Get(&items); err != nil {
			return nil, err
		}
		result = append(result, GroupResult{Group: group, Items: items})
	}
	return result, nil
}

func (s *Service) SaveGroup(id uint, input GroupInput) (models.NavigationGroup, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Locale = strings.TrimSpace(input.Locale)
	if input.Title == "" {
		return models.NavigationGroup{}, errors.New("title is required")
	}
	if input.Locale == "" {
		input.Locale = "all"
	}
	values := map[string]any{"title": input.Title, "locale": input.Locale, "sort_order": input.SortOrder, "is_enabled": input.Enabled, "updated_at": time.Now().UTC()}
	if id == 0 {
		values["created_at"] = time.Now().UTC()
		if err := facades.Orm().Query().Table("footer_navigation_groups").Create(&values); err != nil {
			return models.NavigationGroup{}, err
		}
	} else if _, err := facades.Orm().Query().Table("footer_navigation_groups").Where("id = ?", id).Update(values); err != nil {
		return models.NavigationGroup{}, err
	}
	var row models.NavigationGroup
	if err := facades.Orm().Query().Table("footer_navigation_groups").Find(&row, id); err != nil && id > 0 {
		return row, ErrNotFound
	}
	if id == 0 {
		if err := facades.Orm().Query().Table("footer_navigation_groups").Where("title = ? AND locale = ?", input.Title, input.Locale).OrderByDesc("id").First(&row); err != nil {
			return row, err
		}
	}
	return row, nil
}

func (s *Service) SaveItem(id uint, input ItemInput) (models.NavigationItem, error) {
	input.Label = strings.TrimSpace(input.Label)
	input.TargetType = strings.TrimSpace(input.TargetType)
	input.TargetValue = strings.TrimSpace(input.TargetValue)
	if input.GroupID == 0 || input.Label == "" || input.TargetValue == "" || !validTargetType(input.TargetType) {
		return models.NavigationItem{}, errors.New("invalid navigation item")
	}
	if err := validateTarget(input.TargetType, input.TargetValue); err != nil {
		return models.NavigationItem{}, err
	}
	if input.ParentID != nil {
		if *input.ParentID == 0 || *input.ParentID == id {
			return models.NavigationItem{}, errors.New("navigation item cannot be its own parent")
		}
		var parent models.NavigationItem
		if err := facades.Orm().Query().Table("footer_navigation_items").Where("id = ? AND group_id = ?", *input.ParentID, input.GroupID).First(&parent); err != nil {
			return models.NavigationItem{}, errors.New("navigation parent is invalid")
		}
		if parent.ParentID != nil {
			return models.NavigationItem{}, errors.New("navigation nesting is limited to one level")
		}
	}
	values := map[string]any{"group_id": input.GroupID, "parent_id": input.ParentID, "label": input.Label, "target_type": input.TargetType, "target_value": input.TargetValue, "open_in_new_tab": input.OpenInNewTab, "sort_order": input.SortOrder, "is_enabled": input.Enabled, "updated_at": time.Now().UTC()}
	if id == 0 {
		values["created_at"] = time.Now().UTC()
		if err := facades.Orm().Query().Table("footer_navigation_items").Create(&values); err != nil {
			return models.NavigationItem{}, err
		}
	} else if _, err := facades.Orm().Query().Table("footer_navigation_items").Where("id = ?", id).Update(values); err != nil {
		return models.NavigationItem{}, err
	}
	var row models.NavigationItem
	if err := facades.Orm().Query().Table("footer_navigation_items").Find(&row, id); err != nil && id > 0 {
		return row, ErrNotFound
	}
	if id == 0 {
		var rows []models.NavigationItem
		if err := facades.Orm().Query().Table("footer_navigation_items").Where("group_id = ? AND label = ?", input.GroupID, input.Label).OrderByDesc("id").Get(&rows); err != nil || len(rows) == 0 {
			return row, err
		}
		row = rows[0]
	}
	return row, nil
}

func (s *Service) DeleteGroup(id uint) error {
	if id == 0 {
		return ErrNotFound
	}
	if _, err := facades.Orm().Query().Table("footer_navigation_items").Where("group_id = ?", id).Delete(); err != nil {
		return err
	}
	if _, err := facades.Orm().Query().Table("footer_navigation_groups").Where("id = ?", id).Delete(); err != nil {
		return err
	}
	return nil
}
func (s *Service) DeleteItem(id uint) error {
	if id == 0 {
		return ErrNotFound
	}
	if _, err := facades.Orm().Query().Table("footer_navigation_items").Where("id = ?", id).Delete(); err != nil {
		return err
	}
	return nil
}
