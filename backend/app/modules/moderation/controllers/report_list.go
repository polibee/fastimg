package controllers

import (
	h "github.com/goravel/framework/contracts/http"
	"goravel/app/facades"
	"net/http"
	"strconv"
)

func listReports(x h.Context) h.Response {
	v, e := facades.Auth(x).ID()
	if e != nil {
		return re(x, 401, "AUTH_UNAUTHORIZED")
	}
	uid, e := strconv.ParseUint(v, 10, 32)
	if e != nil || uid == 0 {
		return re(x, 401, "AUTH_UNAUTHORIZED")
	}
	p, _ := strconv.Atoi(x.Request().Query("page", "1"))
	if p < 1 {
		p = 1
	}
	pp, _ := strconv.Atoi(x.Request().Query("per_page", "20"))
	if pp < 1 {
		pp = 20
	}
	if pp > 100 {
		pp = 100
	}
	var rows []map[string]any
	var total int64
	e = facades.Orm().Query().Table("media_reports").Where("reporter_id = ?", uid).OrderByDesc("id").Paginate(p, pp, &rows, &total)
	if e != nil {
		return re(x, 500, "REPORTS_UNAVAILABLE")
	}
	for _, r := range rows {
		delete(r, "reporter_id")
	}
	return x.Response().Json(http.StatusOK, h.Json{"data": rows, "meta": map[string]any{"page": p, "per_page": pp, "total": total}})
}
