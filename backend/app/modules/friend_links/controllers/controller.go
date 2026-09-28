package controllers

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	httpcontract "github.com/goravel/framework/contracts/http"
	"goravel/app/facades"
	adminmiddleware "goravel/app/http/middleware"
	models "goravel/app/modules/friend_links/models"
	services "goravel/app/modules/friend_links/services"
	auditservices "goravel/app/services/audit"
	settingsservices "goravel/app/services/settings"
)

type Controller struct{ service *services.Service }

func NewController() *Controller { return &Controller{service: services.NewService()} }
func publicPayload(row models.Submission) map[string]any {
	return map[string]any{"id": row.ID, "site_name": row.SiteName, "url": row.URL, "logo_url": row.LogoURL, "description": row.Description}
}

func (c *Controller) PublicList(ctx httpcontract.Context) httpcontract.Response {
	rows, err := c.service.List(models.StatusApproved)
	if err != nil {
		return adminmiddleware.APIError(ctx, 500, "FRIEND_LINKS_UNAVAILABLE")
	}
	data := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		data = append(data, publicPayload(row))
	}
	return ctx.Response().Header("Cache-Control", "public, max-age=300").Success().Json(httpcontract.Json{"data": data})
}

// Presentation returns the editable, public copy for the friend-links page.
// It deliberately exposes only presentation text; administrator settings and
// credentials never cross this public boundary.
func (c *Controller) Presentation(ctx httpcontract.Context) httpcontract.Response {
	locale := strings.TrimSpace(ctx.Request().Query("locale"))
	return ctx.Response().Header("Cache-Control", "public, max-age=0, must-revalidate").Success().Json(httpcontract.Json{"data": friendLinksPresentation(locale)})
}

func friendLinksPresentation(locale string) map[string]string {
	prefix := "friend_links.zh_cn."
	if strings.EqualFold(locale, "en-US") || strings.EqualFold(locale, "en-us") || strings.HasPrefix(strings.ToLower(locale), "en-") {
		prefix = "friend_links.en_us."
	}
	defaults := map[string]string{
		"eyebrow":            "社区连接",
		"title":              "友情链接",
		"description":        "展示经过管理员审核的站点。游客也可以提交申请，审核通过后才会公开。",
		"empty":              "暂时还没有已通过审核的友情链接。",
		"submit_title":       "申请交换友情链接",
		"submit_description": "请填写真实站点信息；申请不会立即公开。",
		"submitted":          "申请已提交，等待管理员审核。",
	}
	if prefix == "friend_links.en_us." {
		defaults = map[string]string{
			"eyebrow":            "Community",
			"title":              "Friend links",
			"description":        "Discover sites approved by an administrator. Guests can submit a request, but it is not public until approved.",
			"empty":              "There are no approved friend links yet.",
			"submit_title":       "Submit a friend-link request",
			"submit_description": "Provide accurate site information; requests are not public immediately.",
			"submitted":          "Your request was submitted for review.",
		}
	}
	settings := settingsservices.NewSettingService()
	result := make(map[string]string, len(defaults))
	for key, fallback := range defaults {
		result[key] = settings.Resolve(prefix+key, fallback)
	}
	return result
}
func (c *Controller) Submit(ctx httpcontract.Context) httpcontract.Response {
	var input services.SubmissionInput
	if err := ctx.Request().Bind(&input); err != nil {
		return adminmiddleware.APIError(ctx, 422, "FRIEND_LINK_INVALID")
	}
	if !allowSubmission(ctx.Request().Ip(), input.URL) {
		return adminmiddleware.APIError(ctx, http.StatusTooManyRequests, "FRIEND_LINK_RATE_LIMITED")
	}
	var userID *uint
	if identity, err := facades.Auth(ctx).ID(); err == nil {
		value, _ := strconv.ParseUint(identity, 10, 32)
		if value > 0 {
			id := uint(value)
			userID = &id
		}
	}
	row, err := c.service.Submit(input, userID)
	if err != nil {
		if errors.Is(err, services.ErrDuplicate) {
			return adminmiddleware.APIError(ctx, http.StatusConflict, "FRIEND_LINK_DUPLICATE")
		}
		return adminmiddleware.APIError(ctx, 422, "FRIEND_LINK_INVALID")
	}
	return ctx.Response().Status(http.StatusAccepted).Json(httpcontract.Json{"data": map[string]any{"id": row.ID, "status": row.Status}})
}

func allowSubmission(ip, rawURL string) bool {
	cache := facades.Cache()
	if cache == nil {
		return false
	}
	hash := func(value string) string {
		sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(value))))
		return hex.EncodeToString(sum[:])
	}
	ipKey := "friend-links:ip:" + hash(ip)
	urlKey := "friend-links:url:" + hash(rawURL)
	day := time.Now().UTC().Format("20060102")
	dailyKey := ipKey + ":daily:" + day
	if cache.Has(ipKey) || cache.Has(urlKey) {
		return false
	}
	if !cache.Add(ipKey, int64(1), time.Minute) || !cache.Add(urlKey, int64(1), time.Minute) {
		return false
	}
	cache.Add(dailyKey, int64(0), 24*time.Hour)
	count, err := cache.Increment(dailyKey, 1)
	if err != nil || count > 10 {
		return false
	}
	return true
}
func (c *Controller) AdminList(ctx httpcontract.Context) httpcontract.Response {
	page, _ := strconv.Atoi(ctx.Request().Query("page", "1"))
	perPage, _ := strconv.Atoi(ctx.Request().Query("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	rows, total, err := c.service.ListPaginated(ctx.Request().Query("status"), page, perPage)
	if err != nil {
		return adminmiddleware.APIError(ctx, 500, "FRIEND_LINKS_UNAVAILABLE")
	}
	lastPage := int64(1)
	if total > 0 {
		lastPage = (total + int64(perPage) - 1) / int64(perPage)
	}
	return ctx.Response().Success().Json(httpcontract.Json{"data": rows, "meta": httpcontract.Json{"page": page, "per_page": perPage, "total": total, "last_page": lastPage}})
}
func (c *Controller) Review(ctx httpcontract.Context) httpcontract.Response {
	var input services.ReviewInput
	if err := ctx.Request().Bind(&input); err != nil {
		return adminmiddleware.APIError(ctx, 422, "FRIEND_LINK_REVIEW_INVALID")
	}
	operatorID := authenticatedID(ctx)
	row, err := c.service.Review(uint(ctx.Request().RouteInt64("id")), input, operatorID)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			return adminmiddleware.APIError(ctx, 404, "FRIEND_LINK_NOT_FOUND")
		}
		return adminmiddleware.APIError(ctx, 422, "FRIEND_LINK_REVIEW_INVALID")
	}
	if operatorID > 0 {
		_ = auditservices.NewAuditService().Record(operatorID, "friend_links.review", map[string]any{"id": row.ID, "status": row.Status})
	}
	return ctx.Response().Success().Json(httpcontract.Json{"data": row})
}
func authenticatedID(ctx httpcontract.Context) uint {
	identity, err := facades.Auth(ctx).ID()
	if err != nil {
		return 0
	}
	var id uint
	_, _ = fmt.Sscan(identity, &id)
	return id
}
