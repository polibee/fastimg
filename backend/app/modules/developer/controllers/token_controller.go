package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	httpcontract "github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
	adminmiddleware "goravel/app/http/middleware"
	developerservices "goravel/app/services/developer"
)

type TokenController struct {
	service *developerservices.Service
}

type tokenRequest struct {
	Name      string     `json:"name"`
	Scopes    []string   `json:"scopes"`
	ExpiresAt *time.Time `json:"expires_at"`
}

func NewTokenController() *TokenController {
	return &TokenController{service: developerservices.NewService(developerservices.NewDatabaseRepository())}
}

func (c *TokenController) Index(ctx httpcontract.Context) httpcontract.Response {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return tokenFailure(ctx, http.StatusUnauthorized, "AUTH_UNAUTHORIZED")
	}
	items, err := c.service.List(ctx.Context(), userID)
	if err != nil {
		return tokenFailure(ctx, http.StatusInternalServerError, "TOKENS_UNAVAILABLE")
	}
	return ctx.Response().Success().Json(httpcontract.Json{"data": items})
}

func (c *TokenController) Create(ctx httpcontract.Context) httpcontract.Response {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return tokenFailure(ctx, http.StatusUnauthorized, "AUTH_UNAUTHORIZED")
	}
	var request tokenRequest
	if err := ctx.Request().Bind(&request); err != nil {
		return tokenFailure(ctx, http.StatusUnprocessableEntity, "TOKEN_VALIDATION_FAILED")
	}
	created, err := c.service.Create(ctx.Context(), developerservices.CreateInput{UserID: userID, Name: request.Name, Scopes: request.Scopes, ExpiresAt: request.ExpiresAt})
	if err != nil {
		return tokenServiceFailure(ctx, err)
	}
	return ctx.Response().Status(http.StatusCreated).Json(httpcontract.Json{"data": created})
}

func (c *TokenController) Revoke(ctx httpcontract.Context) httpcontract.Response {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return tokenFailure(ctx, http.StatusUnauthorized, "AUTH_UNAUTHORIZED")
	}
	id := ctx.Request().RouteInt64("id")
	if id <= 0 {
		return tokenFailure(ctx, http.StatusNotFound, "TOKEN_NOT_FOUND")
	}
	if err := c.service.Revoke(ctx.Context(), userID, uint(id)); err != nil {
		return tokenServiceFailure(ctx, err)
	}
	return ctx.Response().NoContent(http.StatusNoContent)
}

func (c *TokenController) Rotate(ctx httpcontract.Context) httpcontract.Response {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return tokenFailure(ctx, http.StatusUnauthorized, "AUTH_UNAUTHORIZED")
	}
	id := ctx.Request().RouteInt64("id")
	if id <= 0 {
		return tokenFailure(ctx, http.StatusNotFound, "TOKEN_NOT_FOUND")
	}
	created, err := c.service.Rotate(ctx.Context(), userID, uint(id))
	if err != nil {
		return tokenServiceFailure(ctx, err)
	}
	return ctx.Response().Status(http.StatusCreated).Json(httpcontract.Json{"data": created})
}

func authenticatedUserID(ctx httpcontract.Context) (uint, error) {
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

func tokenServiceFailure(ctx httpcontract.Context, err error) httpcontract.Response {
	switch {
	case errors.Is(err, developerservices.ErrTokenNotFound):
		return tokenFailure(ctx, http.StatusNotFound, "TOKEN_NOT_FOUND")
	case errors.Is(err, developerservices.ErrInvalidTokenName), errors.Is(err, developerservices.ErrInvalidTokenExpiry), errors.Is(err, developerservices.ErrInvalidTokenScope):
		return tokenFailure(ctx, http.StatusUnprocessableEntity, "TOKEN_VALIDATION_FAILED")
	default:
		return tokenFailure(ctx, http.StatusInternalServerError, "TOKEN_OPERATION_FAILED")
	}
}

func tokenFailure(ctx httpcontract.Context, status int, code string) httpcontract.Response {
	return adminmiddleware.APIError(ctx, status, code)
}
