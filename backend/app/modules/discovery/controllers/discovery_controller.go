package controllers

import (
	"errors"
	"net/http"
	"strconv"

	httpcontract "github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
	adminmiddleware "goravel/app/http/middleware"
	discoveryservices "goravel/app/services/discovery"
	planservices "goravel/app/services/plans"
	storageservices "goravel/app/services/storage"
)

type Controller struct{ service *discoveryservices.Service }

func NewController() *Controller {
	provider := storageservices.NewLocalProvider(facades.Storage().Disk("fastimg"))
	return &Controller{service: discoveryservices.NewDatabaseService(provider)}
}

func (c *Controller) Status(ctx httpcontract.Context) httpcontract.Response {
	return ctx.Response().Success().Json(httpcontract.Json{"data": c.service.Status(ctx.Context())})
}

func (c *Controller) Feed(ctx httpcontract.Context) httpcontract.Response {
	page, perPage := discoveryservices.NormalizePage(ctx.Request().QueryInt("page", 1), ctx.Request().QueryInt("per_page", 24))
	result, err := c.service.Feed(ctx.Context(), page, perPage)
	if err != nil {
		if errors.Is(err, discoveryservices.ErrDiscoveryDisabled) {
			return adminmiddleware.APIError(ctx, http.StatusNotFound, "DISCOVERY_DISABLED")
		}
		return adminmiddleware.APIError(ctx, http.StatusInternalServerError, "DISCOVERY_UNAVAILABLE")
	}
	return ctx.Response().Success().Json(httpcontract.Json{"data": result.Items, "meta": map[string]any{"page": result.Page, "per_page": result.PerPage, "total": result.Total}})
}

func (c *Controller) Content(ctx httpcontract.Context) httpcontract.Response {
	mediaID := ctx.Request().RouteInt64("id")
	if mediaID <= 0 {
		return adminmiddleware.APIError(ctx, http.StatusNotFound, "DISCOVERY_MEDIA_NOT_FOUND")
	}
	variant := ctx.Request().Query("variant", "thumbnail")
	contentType, content, err := c.service.Content(ctx.Context(), uint(mediaID), variant)
	if err != nil {
		if errors.Is(err, planservices.ErrBandwidthQuotaExceeded) {
			return adminmiddleware.APIError(ctx, http.StatusTooManyRequests, "BANDWIDTH_QUOTA_EXCEEDED")
		}
		if errors.Is(err, planservices.ErrBandwidthUnavailable) {
			return adminmiddleware.APIError(ctx, http.StatusServiceUnavailable, "BANDWIDTH_METERING_UNAVAILABLE")
		}
		if errors.Is(err, discoveryservices.ErrDiscoveryDisabled) || errors.Is(err, discoveryservices.ErrDiscoveryMediaNotFound) || errors.Is(err, discoveryservices.ErrDiscoveryVariantNotFound) {
			return adminmiddleware.APIError(ctx, http.StatusNotFound, "DISCOVERY_MEDIA_NOT_FOUND")
		}
		return adminmiddleware.APIError(ctx, http.StatusInternalServerError, "DISCOVERY_CONTENT_UNAVAILABLE")
	}
	return ctx.Response().Header("Cache-Control", "public, max-age=300").Header("X-Content-Type-Options", "nosniff").Data(http.StatusOK, contentType, content)
}

func (c *Controller) Submit(ctx httpcontract.Context) httpcontract.Response {
	identity, err := facades.Auth(ctx).ID()
	if err != nil {
		return adminmiddleware.APIError(ctx, http.StatusUnauthorized, "AUTH_UNAUTHORIZED")
	}
	userID, err := strconv.ParseUint(identity, 10, 32)
	if err != nil || userID == 0 {
		return adminmiddleware.APIError(ctx, http.StatusUnauthorized, "AUTH_UNAUTHORIZED")
	}
	mediaID := ctx.Request().RouteInt64("id")
	if mediaID <= 0 {
		return adminmiddleware.APIError(ctx, http.StatusNotFound, "DISCOVERY_MEDIA_NOT_FOUND")
	}
	if err := c.service.Submit(ctx.Context(), uint(userID), uint(mediaID)); err != nil {
		switch {
		case errors.Is(err, discoveryservices.ErrDiscoveryDisabled):
			return adminmiddleware.APIError(ctx, http.StatusNotFound, "DISCOVERY_DISABLED")
		case errors.Is(err, discoveryservices.ErrDiscoverySubmissionsOff):
			return adminmiddleware.APIError(ctx, http.StatusConflict, "DISCOVERY_SUBMISSIONS_DISABLED")
		case errors.Is(err, discoveryservices.ErrDiscoveryAlreadyPending):
			return adminmiddleware.APIError(ctx, http.StatusConflict, "DISCOVERY_ALREADY_SUBMITTED")
		case errors.Is(err, discoveryservices.ErrDiscoveryMediaNotFound):
			return adminmiddleware.APIError(ctx, http.StatusNotFound, "DISCOVERY_MEDIA_NOT_FOUND")
		default:
			return adminmiddleware.APIError(ctx, http.StatusInternalServerError, "DISCOVERY_SUBMIT_FAILED")
		}
	}
	return ctx.Response().Status(http.StatusAccepted).Json(httpcontract.Json{"data": map[string]any{"media_id": mediaID, "status": "pending"}})
}
