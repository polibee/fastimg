package corspolicy

import (
	"net/url"
	"os"
	"strings"
)

var developmentOrigins = map[string]struct{}{
	"http://127.0.0.1:4173":  {},
	"http://127.0.0.1:4174":  {},
	"http://127.0.0.1:5173":  {},
	"http://127.0.0.1:5174":  {},
	"http://127.0.0.1:5175":  {},
	"http://127.0.0.1:5176":  {},
	"http://127.0.0.1:5180":  {},
	"http://127.0.0.1:5181":  {},
	"http://127.0.0.1:5182":  {},
	"http://127.0.0.1:53083": {},
	"http://127.0.0.1:53084": {},
}

func AllowedOrigin(origin string) bool {
	return AllowedOriginFor(origin, os.Getenv("APP_ENV"), os.Getenv("CORS_ALLOWED_ORIGINS"))
}

// AllowedOriginFor keeps CORS fail-closed in production. Origins are exact
// values, never URL prefixes or wildcard patterns, because authenticated
// browser requests must not be delegated to an arbitrary site.
func AllowedOriginFor(origin, environment, configured string) bool {
	origin = strings.TrimSpace(origin)
	if !validOrigin(origin) {
		return false
	}
	if strings.TrimSpace(configured) != "" {
		for _, candidate := range strings.FieldsFunc(configured, func(r rune) bool { return r == ',' || r == ';' || r == '\n' || r == '\r' }) {
			candidate = strings.TrimSpace(candidate)
			if candidate != "" && candidate != "*" && candidate == origin {
				return true
			}
		}
		return false
	}
	if strings.EqualFold(strings.TrimSpace(environment), "production") {
		return false
	}
	_, ok := developmentOrigins[origin]
	return ok
}

func validOrigin(origin string) bool {
	parsed, err := url.Parse(origin)
	return err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != "" && parsed.Path == "" && parsed.RawQuery == "" && parsed.Fragment == ""
}
