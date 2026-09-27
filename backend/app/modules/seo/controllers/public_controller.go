package controllers

import (
	"fmt"
	"html"
	"net/url"
	"strings"

	"github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
)

type PublicController struct{}

func NewPublicController() *PublicController { return &PublicController{} }

func (c *PublicController) Sitemap(ctx http.Context) http.Response {
	if !settingBool("sitemap.enabled", true) {
		return ctx.Response().Status(404).Json(http.Json{"code": "SITEMAP_DISABLED"})
	}
	base := publicBaseURL()
	paths := []string{"/", "/plans", "/discover"}
	paths = append(paths, settingPaths("sitemap.extra_paths")...)
	var body strings.Builder
	body.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">")
	for _, path := range paths {
		body.WriteString("<url><loc>")
		body.WriteString(html.EscapeString(base + path))
		body.WriteString("</loc></url>")
	}
	body.WriteString("</urlset>")
	return ctx.Response().Header("Content-Type", "application/xml; charset=utf-8").String(200, body.String())
}

func (c *PublicController) Robots(ctx http.Context) http.Response {
	body := "User-agent: *\nAllow: /\nDisallow: /admin/\nDisallow: /api/\n"
	if settingBool("sitemap.enabled", true) {
		body += fmt.Sprintf("Sitemap: %s/sitemap.xml\n", publicBaseURL())
	}
	return ctx.Response().Header("Content-Type", "text/plain; charset=utf-8").String(200, body)
}

// Presentation exposes only public, non-secret presentation settings needed
// by the member UI. Admin settings and credentials never cross this boundary.
func (c *PublicController) Presentation(ctx http.Context) http.Response {
	return ctx.Response().Json(200, http.Json{"data": http.Json{
		"watermark_fallback_image_url": sanitizeFallbackImageURL(settingValue("watermark.fallback_image_url")),
	}})
}

func sanitizeFallbackImageURL(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if strings.HasPrefix(value, "/") && !strings.HasPrefix(value, "//") {
		return value
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return ""
	}
	return parsed.String()
}

func settingBool(key string, fallback bool) bool {
	value := settingValue(key)
	if value == "" {
		return fallback
	}
	return strings.EqualFold(value, "true")
}

func settingPaths(key string) []string {
	value := settingValue(key)
	if value == "" {
		return nil
	}
	seen := map[string]struct{}{"/": {}, "/plans": {}, "/discover": {}}
	paths := make([]string, 0)
	for _, candidate := range strings.FieldsFunc(value, func(r rune) bool { return r == '\n' || r == '\r' || r == ',' || r == ';' }) {
		path := strings.TrimSpace(candidate)
		if path == "" || !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") || strings.Contains(path, "?") || strings.Contains(path, "#") || strings.HasPrefix(path, "/admin") || strings.HasPrefix(path, "/api") {
			continue
		}
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		paths = append(paths, path)
	}
	return paths
}

func settingValue(key string) string {
	if !facades.Schema().HasTable("system_settings") {
		return ""
	}
	var setting struct {
		Value string `orm:"value"`
	}
	if err := facades.Orm().Query().Table("system_settings").Where("key = ?", key).First(&setting); err != nil {
		return ""
	}
	return strings.TrimSpace(setting.Value)
}

func publicBaseURL() string {
	base := settingValue("site_url")
	if base == "" {
		base = strings.TrimSpace(facades.Config().GetString("app.url", "http://127.0.0.1:53085"))
	}
	return strings.TrimRight(base, "/")
}
