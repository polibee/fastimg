package controllers

import (
	"errors"
	"strconv"
	"strings"

	"github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
	advertisingservices "goravel/app/services/advertising"
)

type AdvertisingController struct {
	service *advertisingservices.Service
}

func NewAdvertisingController() *AdvertisingController {
	return &AdvertisingController{service: advertisingservices.NewService()}
}

func (c *AdvertisingController) Index(ctx http.Context) http.Response {
	placement := strings.TrimSpace(ctx.Request().Query("placement"))
	if placement != "" && !advertisingservices.ValidPlacement(placement) {
		return ctx.Response().Status(422).Json(http.Json{"code": "AD_PLACEMENT_INVALID"})
	}
	identity, err := facades.Auth(ctx).ID()
	if err != nil {
		return ctx.Response().Status(401).Json(http.Json{"code": "AUTH_UNAUTHORIZED"})
	}
	userID, err := strconv.ParseUint(identity, 10, 32)
	if err != nil || userID == 0 {
		return ctx.Response().Status(401).Json(http.Json{"code": "AUTH_UNAUTHORIZED"})
	}
	ads, err := c.service.ListPublishedForUser(uint(userID), placement)
	if err != nil {
		if errors.Is(err, advertisingservices.ErrInvalidPlacement) {
			return ctx.Response().Status(422).Json(http.Json{"code": "AD_PLACEMENT_INVALID"})
		}
		return ctx.Response().Status(500).Json(http.Json{"code": "ADS_UNAVAILABLE"})
	}
	return ctx.Response().Success().Json(http.Json{"data": ads})
}
