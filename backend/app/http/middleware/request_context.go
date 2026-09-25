package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	httpcontract "github.com/goravel/framework/contracts/http"
)

type requestIDContextKey struct{}

type requestContextMiddleware struct{}

func (requestContextMiddleware) Signature() string { return "fastimg:request-context" }

func (requestContextMiddleware) Handle(ctx httpcontract.Context) {
	requestID := normalizeRequestID(ctx.Request().Header("X-Request-ID"))
	ctx.WithValue(requestIDContextKey{}, requestID)
	ctx.Response().Header("X-Request-ID", requestID)
	ctx.Request().Next()
}

func RequestContext() httpcontract.Middleware { return requestContextMiddleware{} }

func RequestID(ctx httpcontract.Context) string {
	if value, ok := ctx.Value(requestIDContextKey{}).(string); ok && value != "" {
		return value
	}
	return normalizeRequestID(ctx.Request().Header("X-Request-ID"))
}

func APIError(ctx httpcontract.Context, status int, code string) httpcontract.Response {
	return ctx.Response().Status(status).Json(apiErrorPayload(ctx, status, code))
}

func AbortAPIError(ctx httpcontract.Context, status int, code string) {
	ctx.Response().Status(status).Json(apiErrorPayload(ctx, status, code)).Abort()
}

func apiErrorPayload(ctx httpcontract.Context, status int, code string) httpcontract.Json {
	return httpcontract.Json{
		"code":       code,
		"request_id": RequestID(ctx),
		"retryable":  retryableStatus(status),
	}
}

func normalizeRequestID(input string) string {
	input = strings.TrimSpace(input)
	if input != "" && len(input) <= 64 {
		for _, char := range input {
			if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
				(char >= '0' && char <= '9') || char == '.' || char == '_' || char == ':' || char == '-' {
				continue
			}
			return newRequestID()
		}
		return input
	}
	return newRequestID()
}

func newRequestID() string {
	buffer := make([]byte, 12)
	if _, err := rand.Read(buffer); err == nil {
		return "req_" + hex.EncodeToString(buffer)
	}
	return fmt.Sprintf("req_%x", time.Now().UTC().UnixNano())
}

func retryableStatus(status int) bool {
	return status == 429 || status == 500 || status == 502 || status == 503 || status == 504
}
