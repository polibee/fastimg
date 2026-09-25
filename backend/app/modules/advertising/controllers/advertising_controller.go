package controllers

import (
	"strings"

	"github.com/goravel/framework/contracts/http"

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
	ads, err := c.service.ListPublished(placement)
	if err != nil {
		return ctx.Response().Status(500).Json(http.Json{"code": "ADS_UNAVAILABLE"})
	}
	return ctx.Response().Success().Json(http.Json{"data": ads})
}
