package middleware

import (
	nethttp "net/http"
	"strconv"
	"strings"

	"goravel/app/facades"
	developerservices "goravel/app/services/developer"

	httpcontract "github.com/goravel/framework/contracts/http"
)

type memberTokenContext struct {
	TokenID uint
	UserID  uint
	Scopes  []string
}

type memberAuthenticationMiddleware struct{}

func RequireMemberAuthentication() httpcontract.Middleware { return memberAuthenticationMiddleware{} }

// RequireMemberSession keeps browser/member-session workflows separate from
// the deliberately small Personal API Token surface. Token callers can only
// reach routes that explicitly declare one of the API scopes.
func RequireMemberSession() httpcontract.Middleware { return memberSessionMiddleware{} }

func (m memberAuthenticationMiddleware) Signature() string { return "fastimg:member-authentication" }

func (m memberAuthenticationMiddleware) Handle(ctx httpcontract.Context) {
	header := strings.TrimSpace(ctx.Request().Header("Authorization"))
	if header == "" {
		header = strings.TrimSpace(ctx.Request().Header("X-API-Key"))
	}
	if header == "" {
		memberAuthFailure(ctx, "AUTH_UNAUTHORIZED")
		return
	}
	if _, err := facades.Auth(ctx).Parse(header); err == nil {
		ctx.Request().Next()
		return
	}

	raw := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
	if !strings.HasPrefix(raw, "fst_") {
		memberAuthFailure(ctx, "AUTH_UNAUTHORIZED")
		return
	}
	authenticated, err := developerservices.NewService(developerservices.NewDatabaseRepository()).Authenticate(ctx.Context(), raw, ctx.Request().Ip())
	if err != nil {
		if err == developerservices.ErrTokenExpired {
			memberAuthFailure(ctx, "TOKEN_EXPIRED")
		} else {
			memberAuthFailure(ctx, "TOKEN_INVALID")
		}
		return
	}
	rateLimit, err := developerservices.NewAPIRateLimiter().Allow(ctx.Context(), authenticated.TokenID, ctx.Request().Ip())
	if err != nil {
		AbortAPIError(ctx, nethttp.StatusServiceUnavailable, "TOKEN_RATE_LIMIT_STORE_UNAVAILABLE")
		return
	}
	response := ctx.Response().
		Header("X-RateLimit-Limit", strconv.Itoa(rateLimit.Limit)).
		Header("X-RateLimit-Remaining", strconv.Itoa(rateLimit.Remaining)).
		Header("X-RateLimit-Reset", strconv.FormatInt(rateLimit.ResetAt.Unix(), 10))
	if !rateLimit.Allowed {
		response.Header("Retry-After", strconv.Itoa(rateLimit.RetryAfter)).
			Status(nethttp.StatusTooManyRequests).
			Json(httpcontract.Json{"code": "TOKEN_RATE_LIMITED", "request_id": RequestID(ctx), "retryable": true, "retry_after_seconds": rateLimit.RetryAfter}).Abort()
		return
	}
	if _, err := facades.Auth(ctx).LoginUsingID(authenticated.UserID); err != nil {
		memberAuthFailure(ctx, "AUTH_UNAUTHORIZED")
		return
	}
	ctx.WithValue(memberTokenContextKey{}, memberTokenContext{TokenID: authenticated.TokenID, UserID: authenticated.UserID, Scopes: authenticated.Scopes})
	ctx.Request().Next()
}

type memberSessionMiddleware struct{}

func (m memberSessionMiddleware) Signature() string { return "fastimg:member-session" }

func (m memberSessionMiddleware) Handle(ctx httpcontract.Context) {
	if _, tokenRequest := ctx.Value(memberTokenContextKey{}).(memberTokenContext); tokenRequest {
		AbortAPIError(ctx, nethttp.StatusForbidden, "TOKEN_ENDPOINT_NOT_ALLOWED")
		return
	}
	if _, err := facades.Auth(ctx).ID(); err != nil {
		memberAuthFailure(ctx, "AUTH_UNAUTHORIZED")
		return
	}
	ctx.Request().Next()
}

type memberScopeMiddleware struct{ scope string }

func RequireMemberScope(scope string) httpcontract.Middleware {
	return memberScopeMiddleware{scope: scope}
}

func (m memberScopeMiddleware) Signature() string { return "fastimg:member-scope:" + m.scope }

func (m memberScopeMiddleware) Handle(ctx httpcontract.Context) {
	authenticated, ok := ctx.Value(memberTokenContextKey{}).(memberTokenContext)
	if !ok || authenticated.TokenID == 0 {
		ctx.Request().Next()
		return
	}
	if !developerservices.NewService(developerservices.NewDatabaseRepository()).HasScope(authenticated.Scopes, m.scope) {
		AbortAPIError(ctx, nethttp.StatusForbidden, "TOKEN_SCOPE_REQUIRED")
		return
	}
	ctx.Request().Next()
}

type memberTokenContextKey struct{}

func memberAuthFailure(ctx httpcontract.Context, code string) {
	AbortAPIError(ctx, nethttp.StatusUnauthorized, code)
}
