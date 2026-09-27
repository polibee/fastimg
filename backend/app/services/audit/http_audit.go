package auditservices

import (
	"encoding/json"
	"mime"
	"strings"
)

type HTTPAuditInput struct {
	Method            string
	Path              string
	Query             map[string]string
	Route             map[string]string
	RequestBody       any
	Status            int
	ContentType       string
	ResponseBody      []byte
	ResponseBodyBytes int
	ResponseTruncated bool
}

func ShouldAuditHTTP(method, path string) bool {
	if strings.EqualFold(method, "OPTIONS") || !strings.HasPrefix(path, "/api/v1/") {
		return false
	}
	return path != "/api/v1/admin/audit-logs"
}

// ShouldRecordHTTP keeps the audit trail focused on security-sensitive
// actions, mutations, and failures. Successful media delivery and read-only
// page data already have domain-specific access/statistics records and should
// not flood the administrator audit view.
func ShouldRecordHTTP(method, path string, status int) bool {
	if status >= 400 {
		return true
	}
	if ClassifyHTTPAction(method, path) == "auth.refresh" {
		return false
	}
	return !strings.EqualFold(method, "GET") && !strings.EqualFold(method, "HEAD")
}

func ClassifyHTTPAction(method, path string) string {
	path = strings.SplitN(path, "?", 2)[0]
	path = strings.TrimSuffix(path, "/")
	if strings.HasPrefix(path, "/api/v1/auth/") {
		return "auth." + strings.TrimPrefix(path, "/api/v1/auth/")
	}
	switch {
	case strings.HasPrefix(path, "/api/v1/payment-gateways/") && strings.HasSuffix(path, "/webhook"):
		return "billing.webhook.receive"
	case strings.HasSuffix(path, "/payments") && strings.HasPrefix(path, "/api/v1/orders/"):
		if strings.EqualFold(method, "POST") {
			return "billing.payment.create"
		}
		return "billing.payment.view"
	case strings.HasPrefix(path, "/api/v1/orders"):
		return "billing.order." + httpActionVerb(method)
	case strings.HasPrefix(path, "/api/v1/media/") && strings.HasSuffix(path, "/reports"):
		return "moderation.report." + httpActionVerb(method)
	case path == "/api/v1/uploads" || strings.HasPrefix(path, "/api/v1/uploads/"):
		return "media.upload"
	case strings.HasPrefix(path, "/api/v1/media"):
		return "media." + httpActionVerb(method)
	case strings.HasPrefix(path, "/api/v1/admin/settings"):
		return "settings." + httpActionVerb(method)
	case strings.HasPrefix(path, "/api/v1/admin/reports"):
		return "moderation.report." + httpActionVerb(method)
	case strings.HasPrefix(path, "/api/v1/admin/"):
		return "admin." + httpActionVerb(method)
	case strings.HasPrefix(path, "/api/v1/tokens"):
		return "developer.token." + httpActionVerb(method)
	default:
		return "http." + strings.ToLower(httpActionVerb(method))
	}
}

func httpActionVerb(method string) string {
	switch strings.ToUpper(strings.TrimSpace(method)) {
	case "POST":
		return "create"
	case "PUT", "PATCH":
		return "update"
	case "DELETE":
		return "delete"
	default:
		return "read"
	}
}

// RequestBodyForAudit avoids parsing file bodies into audit events and leaves
// multipart parsing to the route after its request-size limit is installed.
func RequestBodyForAudit(contentType string, readAll func() map[string]any) map[string]any {
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(contentType)), "multipart/") {
		return map[string]any{"omitted": "multipart"}
	}
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err == nil && strings.HasPrefix(strings.ToLower(mediaType), "multipart/") {
		return map[string]any{"omitted": "multipart"}
	}
	return readAll()
}

func BuildHTTPAuditMetadata(input HTTPAuditInput) map[string]any {
	bodyBytes := input.ResponseBodyBytes
	if bodyBytes == 0 {
		bodyBytes = len(input.ResponseBody)
	}
	responseBody := any(nil)
	if input.ResponseTruncated {
		responseBody = map[string]any{"truncated": true, "original_bytes": bodyBytes}
	} else if len(input.ResponseBody) > 0 {
		if strings.Contains(strings.ToLower(input.ContentType), "json") {
			if err := json.Unmarshal(input.ResponseBody, &responseBody); err != nil {
				responseBody = map[string]any{"truncated": true, "invalid_json": true, "bytes": len(input.ResponseBody)}
			}
		}
	}
	metadata := map[string]any{
		"category":  auditCategory(ClassifyHTTPAction(input.Method, input.Path)),
		"operation": ClassifyHTTPAction(input.Method, input.Path),
		"outcome":   auditOutcome(input.Status),
		"request": map[string]any{
			"method": input.Method,
			"path":   input.Path,
			"query":  BoundedValue(input.Query),
			"route":  BoundedValue(input.Route),
			"body":   BoundedValue(input.RequestBody),
		},
		"response": map[string]any{
			"status":       input.Status,
			"content_type": input.ContentType,
			"body":         BoundedValue(responseBody),
		},
	}
	if len(input.ResponseBody) > 0 && !strings.Contains(strings.ToLower(input.ContentType), "json") {
		response := metadata["response"].(map[string]any)
		delete(response, "body")
		response["body_bytes"] = bodyBytes
	}
	return metadata
}

func auditCategory(action string) string {
	if index := strings.IndexByte(action, '.'); index > 0 {
		return action[:index]
	}
	return "other"
}

func auditOutcome(status int) string {
	if status >= 400 {
		return "error"
	}
	return "success"
}
