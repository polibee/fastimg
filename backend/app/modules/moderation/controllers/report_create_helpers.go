package controllers

import (
	h "github.com/goravel/framework/contracts/http"
	"goravel/app/facades"
	"net/http"
	"strconv"
	"strings"
)

type rp struct {
	Reason      string `json:"reason"`
	Description string `json:"description"`
}
type rm struct {
	UserID uint `db:"user_id"`
}

func createReport(x h.Context) h.Response {
	v, e := facades.Auth(x).ID()
	if e != nil {
		return re(x, 401, "AUTH_UNAUTHORIZED")
	}
	uid, e := strconv.ParseUint(v, 10, 32)
	mid := x.Request().RouteInt64("id")
	if e != nil || uid == 0 {
		return re(x, 401, "AUTH_UNAUTHORIZED")
	}
	if mid < 1 {
		return re(x, 404, "MEDIA_NOT_FOUND")
	}
	var p rp
	if x.Request().Bind(&p) != nil {
		return re(x, 422, "REPORT_INVALID")
	}
	p.Reason, p.Description = strings.TrimSpace(p.Reason), strings.TrimSpace(p.Description)
	if len(p.Reason) < 2 || len(p.Reason) > 64 || len(p.Description) > 2000 {
		return re(x, 422, "REPORT_INVALID")
	}
	var m rm
	if e = facades.Orm().Query().Table("media_assets").Where("id = ? AND status = ? AND deleted_at IS NULL AND visibility = ? AND moderation_status = ?", mid, "ready", "public", "approved").First(&m); e != nil || m.UserID == 0 {
		return re(x, 404, "MEDIA_NOT_FOUND")
	}
	if m.UserID == uint(uid) {
		return re(x, 409, "REPORT_SELF_MEDIA")
	}
	if e = facades.Orm().Query().Table("media_reports").Create(&map[string]any{"media_asset_id": mid, "reporter_id": uid, "reason": p.Reason, "description": p.Description, "status": "pending"}); e != nil {
		return re(x, 500, "REPORT_UNAVAILABLE")
	}
	return x.Response().Status(http.StatusCreated).Json(h.Json{"data": map[string]any{"media_asset_id": mid, "status": "pending"}})
}
func re(x h.Context, s int, c string) h.Response {
	return x.Response().Status(s).Json(h.Json{"code": c})
}
