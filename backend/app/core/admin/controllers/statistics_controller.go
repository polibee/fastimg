package controllers

import (
	"net/http"
	"time"

	httpcontract "github.com/goravel/framework/contracts/http"

	statisticsservices "goravel/app/services/statistics"
)

type StatisticsController struct{ service *statisticsservices.Service }

func NewStatisticsController() *StatisticsController {
	return &StatisticsController{service: statisticsservices.NewService()}
}

func (c *StatisticsController) Trends(ctx httpcontract.Context) httpcontract.Response {
	from, to, err := parseTrendRange(ctx)
	if err != nil {
		return ctx.Response().Status(http.StatusUnprocessableEntity).Json(httpcontract.Json{"code": "STATISTICS_RANGE_INVALID"})
	}
	report, err := c.service.Trends(ctx.Context(), from, to)
	if err != nil {
		return ctx.Response().Status(http.StatusInternalServerError).Json(httpcontract.Json{"code": "STATISTICS_UNAVAILABLE"})
	}
	return ctx.Response().Success().Json(httpcontract.Json{"data": report})
}

func parseTrendRange(ctx httpcontract.Context) (time.Time, time.Time, error) {
	to := time.Now().UTC()
	from := to.AddDate(0, 0, -29)
	if value := ctx.Request().Query("from"); value != "" {
		parsed, err := time.Parse("2006-01-02", value)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		from = parsed.UTC()
	}
	if value := ctx.Request().Query("to"); value != "" {
		parsed, err := time.Parse("2006-01-02", value)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		to = parsed.UTC()
	}
	return from, to, nil
}
