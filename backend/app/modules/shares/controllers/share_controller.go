package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	httpcontract "github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
	linkservices "goravel/app/services/links"
	planservices "goravel/app/services/plans"
	shareservices "goravel/app/services/shares"
	storageservices "goravel/app/services/storage"
)

type ShareController struct {
	service *shareservices.Service
	links   *linkservices.Service
}

type createShareRequest struct {
	ExpiresAt *time.Time `json:"expires_at"`
	Password  string     `json:"password"`
}

type signedURLRequest struct {
	Variant   string `json:"variant"`
	ExpiresIn int    `json:"expires_in"`
}

type hotlinkPolicyRequest struct {
	Mode           string `json:"mode"`
	AllowNoReferer bool   `json:"allow_no_referer"`
}

type hotlinkDomainRequest struct {
	Domain string `json:"domain"`
}

func NewShareController() *ShareController {
	provider := storageservices.NewLocalProvider(facades.Storage().Disk("fastimg"))
	links := linkservices.NewService(provider, facades.Config().GetString("app.key", ""))
	return &ShareController{service: shareservices.NewServiceWithPolicy(shareservices.NewDatabaseRepository(), provider, links), links: links}
}

func (c *ShareController) Create(ctx httpcontract.Context) httpcontract.Response {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return ctx.Response().Status(http.StatusUnauthorized).Json(httpcontract.Json{"code": "AUTH_UNAUTHORIZED"})
	}
	mediaID := ctx.Request().RouteInt64("id")
	if mediaID <= 0 {
		return ctx.Response().Status(http.StatusNotFound).Json(httpcontract.Json{"code": "MEDIA_NOT_FOUND"})
	}
	var input createShareRequest
	if err := ctx.Request().Bind(&input); err != nil {
		return ctx.Response().Status(http.StatusUnprocessableEntity).Json(httpcontract.Json{"code": "SHARE_VALIDATION_FAILED"})
	}
	view, err := c.service.Create(ctx.Context(), shareservices.CreateInput{UserID: userID, MediaID: uint(mediaID), ExpiresAt: input.ExpiresAt, Password: input.Password})
	if err != nil {
		return shareFailure(ctx, err)
	}
	return ctx.Response().Status(http.StatusCreated).Json(httpcontract.Json{"data": view})
}

func (c *ShareController) Index(ctx httpcontract.Context) httpcontract.Response {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return ctx.Response().Status(http.StatusUnauthorized).Json(httpcontract.Json{"code": "AUTH_UNAUTHORIZED"})
	}
	page, _ := strconv.Atoi(ctx.Request().Query("page", "1"))
	perPage, _ := strconv.Atoi(ctx.Request().Query("per_page", "20"))
	items, total, err := c.service.ListPage(ctx.Context(), userID, page, perPage)
	if err != nil {
		return ctx.Response().Status(http.StatusInternalServerError).Json(httpcontract.Json{"code": "SHARES_UNAVAILABLE"})
	}
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	lastPage := int64(1)
	if total > 0 {
		lastPage = (total + int64(perPage) - 1) / int64(perPage)
	}
	return ctx.Response().Success().Json(httpcontract.Json{"data": items, "meta": httpcontract.Json{"page": page, "per_page": perPage, "total": total, "last_page": lastPage}})
}

func (c *ShareController) Revoke(ctx httpcontract.Context) httpcontract.Response {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return ctx.Response().Status(http.StatusUnauthorized).Json(httpcontract.Json{"code": "AUTH_UNAUTHORIZED"})
	}
	shareID := ctx.Request().RouteInt64("id")
	if shareID <= 0 {
		return ctx.Response().Status(http.StatusNotFound).Json(httpcontract.Json{"code": "SHARE_NOT_FOUND"})
	}
	if err := c.service.Revoke(ctx.Context(), userID, uint(shareID)); err != nil {
		return shareFailure(ctx, err)
	}
	return ctx.Response().NoContent(http.StatusNoContent)
}

func (c *ShareController) Public(ctx httpcontract.Context) httpcontract.Response {
	token := ctx.Request().Route("token")
	if token == "" {
		return ctx.Response().Status(http.StatusNotFound).Json(httpcontract.Json{"code": "SHARE_NOT_FOUND"})
	}
	variant := ctx.Request().Query("variant", "original")
	password := ctx.Request().Query("password", "")
	content, err := c.service.PublicContent(ctx.Context(), token, variant, password, ctx.Request().Header("Referer"))
	if err != nil {
		if errors.Is(err, planservices.ErrBandwidthQuotaExceeded) {
			return ctx.Response().Status(http.StatusTooManyRequests).Json(httpcontract.Json{"code": "BANDWIDTH_QUOTA_EXCEEDED"})
		}
		if errors.Is(err, planservices.ErrBandwidthUnavailable) {
			return ctx.Response().Status(http.StatusServiceUnavailable).Json(httpcontract.Json{"code": "BANDWIDTH_METERING_UNAVAILABLE"})
		}
		if errors.Is(err, shareservices.ErrSharePasswordRequired) {
			return ctx.Response().Status(http.StatusUnauthorized).Json(httpcontract.Json{"code": "SHARE_PASSWORD_REQUIRED"})
		}
		return ctx.Response().Status(http.StatusNotFound).Json(httpcontract.Json{"code": "SHARE_NOT_FOUND"})
	}
	return ctx.Response().Header("Cache-Control", "no-store").Header("X-Content-Type-Options", "nosniff").Data(http.StatusOK, content.ContentType, content.Content)
}

func (c *ShareController) SignedURL(ctx httpcontract.Context) httpcontract.Response {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return ctx.Response().Status(http.StatusUnauthorized).Json(httpcontract.Json{"code": "AUTH_UNAUTHORIZED"})
	}
	mediaID := ctx.Request().RouteInt64("id")
	if mediaID <= 0 {
		return ctx.Response().Status(http.StatusNotFound).Json(httpcontract.Json{"code": "MEDIA_NOT_FOUND"})
	}
	input := signedURLRequest{Variant: ctx.Request().Query("variant", "original"), ExpiresIn: ctx.Request().QueryInt("expires_in", int(linkservices.DefaultSignedURLLifetime/time.Second))}
	if ctx.Request().Query("variant", "") == "" {
		var bound signedURLRequest
		if err := ctx.Request().Bind(&bound); err == nil {
			if bound.Variant != "" {
				input.Variant = bound.Variant
			}
			if bound.ExpiresIn != 0 {
				input.ExpiresIn = bound.ExpiresIn
			}
		}
	}
	view, err := c.links.CreateSignedURL(ctx.Context(), linkservices.SignedURLInput{UserID: userID, MediaID: uint(mediaID), Variant: input.Variant, ExpiresIn: time.Duration(input.ExpiresIn) * time.Second})
	if err != nil {
		return linkFailure(ctx, err)
	}
	return ctx.Response().Status(http.StatusCreated).Json(httpcontract.Json{"data": view})
}

func (c *ShareController) HotlinkPolicy(ctx httpcontract.Context) httpcontract.Response {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return ctx.Response().Status(http.StatusUnauthorized).Json(httpcontract.Json{"code": "AUTH_UNAUTHORIZED"})
	}
	mediaID := ctx.Request().RouteInt64("id")
	if mediaID <= 0 {
		return ctx.Response().Status(http.StatusNotFound).Json(httpcontract.Json{"code": "MEDIA_NOT_FOUND"})
	}
	policy, err := c.links.GetPolicy(ctx.Context(), userID, uint(mediaID))
	if err != nil {
		return linkFailure(ctx, err)
	}
	return ctx.Response().Success().Json(httpcontract.Json{"data": policy})
}

func (c *ShareController) UpdateHotlinkPolicy(ctx httpcontract.Context) httpcontract.Response {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return ctx.Response().Status(http.StatusUnauthorized).Json(httpcontract.Json{"code": "AUTH_UNAUTHORIZED"})
	}
	mediaID := ctx.Request().RouteInt64("id")
	var input hotlinkPolicyRequest
	if mediaID <= 0 || ctx.Request().Bind(&input) != nil {
		return ctx.Response().Status(http.StatusUnprocessableEntity).Json(httpcontract.Json{"code": "HOTLINK_POLICY_INVALID"})
	}
	policy, err := c.links.UpdatePolicy(ctx.Context(), userID, uint(mediaID), input.Mode, input.AllowNoReferer)
	if err != nil {
		return linkFailure(ctx, err)
	}
	return ctx.Response().Success().Json(httpcontract.Json{"data": policy})
}

func (c *ShareController) HotlinkDomains(ctx httpcontract.Context) httpcontract.Response {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return ctx.Response().Status(http.StatusUnauthorized).Json(httpcontract.Json{"code": "AUTH_UNAUTHORIZED"})
	}
	items, err := c.links.ListDomains(ctx.Context(), userID)
	if err != nil {
		return ctx.Response().Status(http.StatusInternalServerError).Json(httpcontract.Json{"code": "HOTLINK_DOMAINS_UNAVAILABLE"})
	}
	return ctx.Response().Success().Json(httpcontract.Json{"data": items})
}

func (c *ShareController) CreateHotlinkDomain(ctx httpcontract.Context) httpcontract.Response {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return ctx.Response().Status(http.StatusUnauthorized).Json(httpcontract.Json{"code": "AUTH_UNAUTHORIZED"})
	}
	var input hotlinkDomainRequest
	if err := ctx.Request().Bind(&input); err != nil {
		return ctx.Response().Status(http.StatusUnprocessableEntity).Json(httpcontract.Json{"code": "HOTLINK_DOMAIN_INVALID"})
	}
	item, err := c.links.AddDomain(ctx.Context(), userID, input.Domain)
	if err != nil {
		return linkFailure(ctx, err)
	}
	return ctx.Response().Status(http.StatusCreated).Json(httpcontract.Json{"data": item})
}

func (c *ShareController) DeleteHotlinkDomain(ctx httpcontract.Context) httpcontract.Response {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return ctx.Response().Status(http.StatusUnauthorized).Json(httpcontract.Json{"code": "AUTH_UNAUTHORIZED"})
	}
	if err := c.links.DeleteDomain(ctx.Context(), userID, uint(ctx.Request().RouteInt64("id"))); err != nil {
		return linkFailure(ctx, err)
	}
	return ctx.Response().NoContent(http.StatusNoContent)
}

func (c *ShareController) PublicSigned(ctx httpcontract.Context) httpcontract.Response {
	mediaID := ctx.Request().RouteInt64("id")
	content, err := c.links.PublicSignedContent(ctx.Context(), uint(mediaID), ctx.Request().Query("variant", "original"), ctx.Request().Query("expires", ""), ctx.Request().Query("signature", ""), ctx.Request().Header("Referer"), ctx.Request().Header("User-Agent"))
	if err != nil {
		if errors.Is(err, planservices.ErrBandwidthQuotaExceeded) {
			return ctx.Response().Status(http.StatusTooManyRequests).Json(httpcontract.Json{"code": "BANDWIDTH_QUOTA_EXCEEDED"})
		}
		if errors.Is(err, planservices.ErrBandwidthUnavailable) {
			return ctx.Response().Status(http.StatusServiceUnavailable).Json(httpcontract.Json{"code": "BANDWIDTH_METERING_UNAVAILABLE"})
		}
		return ctx.Response().Status(http.StatusNotFound).Json(httpcontract.Json{"code": "LINK_NOT_FOUND"})
	}
	return ctx.Response().Header("Cache-Control", "private, max-age=60").Header("X-Content-Type-Options", "nosniff").Data(http.StatusOK, content.ContentType, content.Content)
}

func shareFailure(ctx httpcontract.Context, err error) httpcontract.Response {
	switch {
	case errors.Is(err, shareservices.ErrShareNotFound), errors.Is(err, shareservices.ErrShareUnavailable):
		return ctx.Response().Status(http.StatusNotFound).Json(httpcontract.Json{"code": "SHARE_NOT_FOUND"})
	case errors.Is(err, shareservices.ErrInvalidShareExpiry):
		return ctx.Response().Status(http.StatusUnprocessableEntity).Json(httpcontract.Json{"code": "SHARE_EXPIRY_INVALID"})
	case errors.Is(err, shareservices.ErrInvalidSharePassword):
		return ctx.Response().Status(http.StatusUnprocessableEntity).Json(httpcontract.Json{"code": "SHARE_PASSWORD_INVALID"})
	default:
		return ctx.Response().Status(http.StatusInternalServerError).Json(httpcontract.Json{"code": "SHARE_OPERATION_FAILED"})
	}
}

func linkFailure(ctx httpcontract.Context, err error) httpcontract.Response {
	switch {
	case errors.Is(err, linkservices.ErrMediaNotFound), errors.Is(err, linkservices.ErrVariantNotFound):
		return ctx.Response().Status(http.StatusNotFound).Json(httpcontract.Json{"code": "MEDIA_NOT_FOUND"})
	case errors.Is(err, linkservices.ErrInvalidVariant):
		return ctx.Response().Status(http.StatusUnprocessableEntity).Json(httpcontract.Json{"code": "VARIANT_INVALID"})
	case errors.Is(err, linkservices.ErrInvalidSignedURLExpiry):
		return ctx.Response().Status(http.StatusUnprocessableEntity).Json(httpcontract.Json{"code": "SIGNED_URL_EXPIRY_INVALID"})
	case errors.Is(err, linkservices.ErrLinkSigningUnavailable):
		return ctx.Response().Status(http.StatusServiceUnavailable).Json(httpcontract.Json{"code": "LINK_SIGNING_UNAVAILABLE"})
	case errors.Is(err, linkservices.ErrHotlinkPolicyInvalid):
		return ctx.Response().Status(http.StatusUnprocessableEntity).Json(httpcontract.Json{"code": "HOTLINK_POLICY_INVALID"})
	case errors.Is(err, linkservices.ErrInvalidHotlinkDomain):
		return ctx.Response().Status(http.StatusUnprocessableEntity).Json(httpcontract.Json{"code": "HOTLINK_DOMAIN_INVALID"})
	case errors.Is(err, linkservices.ErrDomainNotFound):
		return ctx.Response().Status(http.StatusNotFound).Json(httpcontract.Json{"code": "HOTLINK_DOMAIN_NOT_FOUND"})
	case errors.Is(err, linkservices.ErrDomainConflict):
		return ctx.Response().Status(http.StatusConflict).Json(httpcontract.Json{"code": "HOTLINK_DOMAIN_EXISTS"})
	default:
		return ctx.Response().Status(http.StatusInternalServerError).Json(httpcontract.Json{"code": "LINK_OPERATION_FAILED"})
	}
}

func authenticatedUserID(ctx httpcontract.Context) (uint, error) {
	identity, err := facades.Auth(ctx).ID()
	if err != nil {
		return 0, err
	}
	parsed, err := strconv.ParseUint(identity, 10, 32)
	if err != nil || parsed == 0 {
		return 0, errors.New("invalid authenticated user")
	}
	return uint(parsed), nil
}
