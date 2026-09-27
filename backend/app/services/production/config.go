package production

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
)

// Environment returns only configuration names needed by the production gate;
// it deliberately never logs or returns secret values to callers.
func Environment() map[string]string {
	keys := []string{"APP_ENV", "APP_DEBUG", "APP_URL", "APP_KEY", "JWT_SECRET", "DB_CONNECTION", "DB_SSLMODE", "CORS_ALLOWED_ORIGINS"}
	values := make(map[string]string, len(keys))
	for _, key := range keys {
		values[key] = getenv(key)
	}
	return values
}

// getenv is a variable so tests can supply a deterministic environment without
// mutating the process environment.
var getenv = func(key string) string { return lookupEnv(key) }

func lookupEnv(key string) string {
	// Kept in a tiny helper to make the validation contract easy to test.
	return os.Getenv(key)
}

func Validate(env map[string]string) error {
	if !strings.EqualFold(strings.TrimSpace(env["APP_ENV"]), "production") {
		return nil
	}
	var problems []error
	if !strings.EqualFold(strings.TrimSpace(env["APP_DEBUG"]), "false") {
		problems = append(problems, errors.New("APP_DEBUG must be false"))
	}
	if !validPublicURL(env["APP_URL"]) {
		problems = append(problems, errors.New("APP_URL must be an absolute non-loopback http(s) URL"))
	}
	if len(strings.TrimSpace(env["APP_KEY"])) < 32 {
		problems = append(problems, errors.New("APP_KEY must contain at least 32 characters"))
	}
	if len(strings.TrimSpace(env["JWT_SECRET"])) < 32 {
		problems = append(problems, errors.New("JWT_SECRET must contain at least 32 characters"))
	}
	if strings.EqualFold(strings.TrimSpace(env["DB_CONNECTION"]), "postgres") && (strings.TrimSpace(env["DB_SSLMODE"]) == "" || strings.EqualFold(strings.TrimSpace(env["DB_SSLMODE"]), "disable")) {
		problems = append(problems, errors.New("DB_SSLMODE must be an encrypted PostgreSQL mode in production"))
	}
	if strings.Contains(env["CORS_ALLOWED_ORIGINS"], "*") {
		problems = append(problems, errors.New("CORS_ALLOWED_ORIGINS must not contain wildcard origins"))
	}
	if len(problems) > 0 {
		return fmt.Errorf("%w: %v", errors.New("production configuration failed"), errors.Join(problems...))
	}
	return nil
}

func validPublicURL(raw string) bool {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	return host != "localhost" && host != "127.0.0.1" && host != "::1"
}
