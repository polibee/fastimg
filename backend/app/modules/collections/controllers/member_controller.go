package controllers

import (
	"errors"
	"net/http"
	"strconv"

	httpcontract "github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
	collectionservices "goravel/app/services/collections"
)

type MemberCollectionController struct {
	service *collectionservices.Service
	kind    collectionservices.Kind
}

func NewMemberCollectionController(kind collectionservices.Kind) *MemberCollectionController {
	return &MemberCollectionController{service: collectionservices.NewService(collectionservices.NewDatabaseRepository()), kind: kind}
}

func (c *MemberCollectionController) Index(ctx httpcontract.Context) httpcontract.Response {
	userID, err := memberUserID(ctx)
	if err != nil {
		return ctx.Response().Status(http.StatusUnauthorized).Json(httpcontract.Json{"code": "AUTH_UNAUTHORIZED"})
	}
	items, err := c.service.List(ctx.Context(), c.kind, userID)
	if err != nil {
		return ctx.Response().Status(http.StatusInternalServerError).Json(httpcontract.Json{"code": "COLLECTIONS_UNAVAILABLE"})
	}
	return ctx.Response().Success().Json(httpcontract.Json{"data": items})
}

func (c *MemberCollectionController) Create(ctx httpcontract.Context) httpcontract.Response {
	userID, err := memberUserID(ctx)
	if err != nil {
		return ctx.Response().Status(http.StatusUnauthorized).Json(httpcontract.Json{"code": "AUTH_UNAUTHORIZED"})
	}
	var input collectionservices.Input
	if err := ctx.Request().Bind(&input); err != nil {
		return ctx.Response().Status(http.StatusUnprocessableEntity).Json(httpcontract.Json{"code": "COLLECTION_VALIDATION_FAILED"})
	}
	item, err := c.service.Create(ctx.Context(), c.kind, userID, input)
	if err != nil {
		return collectionError(ctx, err)
	}
	return ctx.Response().Status(http.StatusCreated).Json(httpcontract.Json{"data": item})
}

func (c *MemberCollectionController) Update(ctx httpcontract.Context) httpcontract.Response {
	userID, err := memberUserID(ctx)
	if err != nil {
		return ctx.Response().Status(http.StatusUnauthorized).Json(httpcontract.Json{"code": "AUTH_UNAUTHORIZED"})
	}
	id := ctx.Request().RouteInt64("id")
	if id <= 0 {
		return ctx.Response().Status(http.StatusNotFound).Json(httpcontract.Json{"code": "COLLECTION_NOT_FOUND"})
	}
	var input collectionservices.Input
	if err := ctx.Request().Bind(&input); err != nil {
		return ctx.Response().Status(http.StatusUnprocessableEntity).Json(httpcontract.Json{"code": "COLLECTION_VALIDATION_FAILED"})
	}
	item, err := c.service.Update(ctx.Context(), c.kind, userID, uint(id), input)
	if err != nil {
		return collectionError(ctx, err)
	}
	return ctx.Response().Success().Json(httpcontract.Json{"data": item})
}

func (c *MemberCollectionController) Delete(ctx httpcontract.Context) httpcontract.Response {
	userID, err := memberUserID(ctx)
	if err != nil {
		return ctx.Response().Status(http.StatusUnauthorized).Json(httpcontract.Json{"code": "AUTH_UNAUTHORIZED"})
	}
	id := ctx.Request().RouteInt64("id")
	if id <= 0 {
		return ctx.Response().Status(http.StatusNotFound).Json(httpcontract.Json{"code": "COLLECTION_NOT_FOUND"})
	}
	if err := c.service.Delete(ctx.Context(), c.kind, userID, uint(id)); err != nil {
		return collectionError(ctx, err)
	}
	return ctx.Response().NoContent(http.StatusNoContent)
}

func collectionError(ctx httpcontract.Context, err error) httpcontract.Response {
	switch {
	case errors.Is(err, collectionservices.ErrNotFound), errors.Is(err, collectionservices.ErrParentNotFound), errors.Is(err, collectionservices.ErrMediaNotFound):
		return ctx.Response().Status(http.StatusNotFound).Json(httpcontract.Json{"code": "COLLECTION_NOT_FOUND"})
	case errors.Is(err, collectionservices.ErrInvalidName), errors.Is(err, collectionservices.ErrInvalidKind), errors.Is(err, collectionservices.ErrInvalidVisibility):
		return ctx.Response().Status(http.StatusUnprocessableEntity).Json(httpcontract.Json{"code": "COLLECTION_VALIDATION_FAILED"})
	default:
		return ctx.Response().Status(http.StatusInternalServerError).Json(httpcontract.Json{"code": "COLLECTION_OPERATION_FAILED"})
	}
}

func memberUserID(ctx httpcontract.Context) (uint, error) {
	identity, err := facades.Auth(ctx).ID()
	if err != nil {
		return 0, err
	}
	parsed, err := strconv.ParseUint(identity, 10, 32)
	if err != nil || parsed == 0 {
		return 0, errors.New("invalid authenticated user")
	}
	return uint(parsed), nil
}
