package actions

import (
	"errors"
	"strconv"

	h "github.com/goravel/framework/contracts/http"

	a "goravel/app/core/admin/actions"
	"goravel/app/facades"
	mediaservices "goravel/app/services/media"
)

type MediaModerationHandler struct{}

func NewMediaModerationHandler() *MediaModerationHandler { return &MediaModerationHandler{} }
func (*MediaModerationHandler) Kind() string             { return "media-moderation" }
func (*MediaModerationHandler) Payload() string          { return "media-moderation" }

func (handler *MediaModerationHandler) Execute(ctx h.Context, request a.Request) (a.Result, error) {
	operation, ok := request.Payload["operation"].(string)
	if !ok || operation == "permanent_delete" {
		return a.Result{}, a.ErrPayloadContract
	}
	return executeMediaOperation(ctx, request, operation)
}

type MediaPermanentDeleteHandler struct{}

func NewMediaPermanentDeleteHandler() *MediaPermanentDeleteHandler {
	return &MediaPermanentDeleteHandler{}
}
func (*MediaPermanentDeleteHandler) Kind() string    { return "media-permanent-delete" }
func (*MediaPermanentDeleteHandler) Payload() string { return "media-permanent-delete" }

func (handler *MediaPermanentDeleteHandler) Execute(ctx h.Context, request a.Request) (a.Result, error) {
	if request.Payload["confirmation"] != "permanently-delete" {
		return a.Result{}, errors.New("permanent media deletion confirmation is required")
	}
	return executeMediaOperation(ctx, request, "permanent_delete")
}

func executeMediaOperation(ctx h.Context, request a.Request, operation string) (a.Result, error) {
	identity, err := facades.Auth(ctx).ID()
	if err != nil {
		return a.Result{}, err
	}
	operatorID, err := strconv.ParseUint(identity, 10, 32)
	if err != nil || operatorID == 0 {
		return a.Result{}, errors.New("invalid administrator identity")
	}
	result := a.Result{Action: request.Action, Requested: len(request.IDs)}
	service := mediaservices.NewAdminService()
	for _, id := range request.IDs {
		if err := service.Apply(ctx.Context(), uint(operatorID), uint(id), operation); err != nil {
			result.Failed++
			result.Failures = append(result.Failures, a.Failure{ID: id, Code: "MEDIA_ADMIN_OPERATION_FAILED"})
			continue
		}
		result.Succeeded++
	}
	return result, nil
}
