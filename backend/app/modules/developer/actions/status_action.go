package actions

import (
	"errors"
	"strings"

	adminactions "goravel/app/core/admin/actions"
	developerservices "goravel/app/services/developer"

	"github.com/goravel/framework/contracts/http"
)

var (
	ErrUnknownParameter = errors.New("unknown api token status action parameter")
	ErrInvalidStatus    = errors.New("invalid api token status")
)

type SetStatusHandler struct {
	repository *developerservices.DatabaseRepository
}

func NewSetStatusHandler() *SetStatusHandler {
	return &SetStatusHandler{repository: developerservices.NewDatabaseRepository()}
}

func (h *SetStatusHandler) Kind() string { return "api-token-status" }

func (h *SetStatusHandler) Payload() string { return "api-token-status" }

func ParseStatusParams(payload map[string]any) (string, error) {
	if len(payload) != 1 {
		for name := range payload {
			if name != "status" {
				return "", ErrUnknownParameter
			}
		}
	}
	value, ok := payload["status"].(string)
	if !ok {
		return "", ErrInvalidStatus
	}
	status := strings.TrimSpace(value)
	switch status {
	case "disabled", "revoked":
		return status, nil
	default:
		return "", ErrInvalidStatus
	}
}

func (h *SetStatusHandler) Execute(ctx http.Context, request adminactions.Request) (adminactions.Result, error) {
	status, err := ParseStatusParams(request.Payload)
	if err != nil {
		return adminactions.Result{}, err
	}
	result := adminactions.Result{Action: request.Action, Requested: len(request.IDs)}
	for _, id := range request.IDs {
		if err := h.repository.SetStatus(ctx.Context(), uint(id), status); err != nil {
			result.Failed++
			result.Failures = append(result.Failures, adminactions.Failure{ID: id, Code: apiTokenStatusErrorCode(err)})
			continue
		}
		result.Succeeded++
	}
	return result, nil
}

func apiTokenStatusErrorCode(err error) string {
	if errors.Is(err, developerservices.ErrInvalidTokenStatus) {
		return "VALIDATION_ERROR"
	}
	if errors.Is(err, developerservices.ErrTokenNotFound) {
		return "RESOURCE_NOT_FOUND"
	}
	return "INTERNAL_ERROR"
}
