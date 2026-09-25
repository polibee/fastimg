package media

import (
	"fmt"
	"html"
	"net/url"
	"strings"

	"goravel/app/facades"
)

// Links returns member-visible authenticated link formats. Upload responses
// should use LinkFormatsFromVariants with signed public URLs instead.
func Links(mediaID uint, originalName string) map[string]string {
	base := strings.TrimRight(facades.Config().GetString("app.url", "http://127.0.0.1:53083"), "/")
	original := absoluteURL(base, contentURL(mediaID, "original"))
	thumbnail := absoluteURL(base, contentURL(mediaID, "thumbnail"))
	medium := absoluteURL(base, contentURL(mediaID, "medium"))
	return LinkFormatsFromVariants(originalName, map[string]string{"original": original, "thumbnail": thumbnail, "medium": medium})
}

// LinkFormatsFromVariants creates the formats users copy into forums, docs and
// HTML. The supplied URLs must already be public absolute URLs.
func LinkFormatsFromVariants(originalName string, variants map[string]string) map[string]string {
	original := variants["original"]
	thumbnail := variants["thumbnail"]
	medium := variants["medium"]
	return map[string]string{
		"original":  original,
		"thumbnail": thumbnail,
		"medium":    medium,
		"url":       original,
		"markdown":  fmt.Sprintf("![%s](%s)", originalName, original),
		"html":      fmt.Sprintf(`<img src="%s" alt="%s">`, original, html.EscapeString(originalName)),
		"bbcode":    fmt.Sprintf("[img]%s[/img]", original),
	}
}

func absoluteURL(base, path string) string {
	if parsed, err := url.Parse(path); err == nil && parsed.IsAbs() {
		return path
	}
	return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(path, "/")
}

func contentURL(mediaID uint, variant string) string {
	return fmt.Sprintf("/api/v1/media/%d/content?variant=%s", mediaID, variant)
}
