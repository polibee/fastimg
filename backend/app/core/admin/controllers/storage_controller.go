package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	httpcontract "github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
	auditservices "goravel/app/services/audit"
	storageservices "goravel/app/services/storage"
)

type StorageController struct {
	connections *storageservices.ConnectionService
}

func NewStorageController() *StorageController {
	return &StorageController{connections: storageservices.NewConnectionService()}
}

func (c *StorageController) Index(ctx httpcontract.Context) httpcontract.Response {
	connections, err := c.connections.List()
	if err != nil {
		return ctx.Response().Status(http.StatusInternalServerError).Json(httpcontract.Json{"code": "STORAGE_CONNECTIONS_UNAVAILABLE"})
	}
	return ctx.Response().Success().Json(httpcontract.Json{"data": httpcontract.Json{
		"providers":   storageservices.ProviderDefinitions(),
		"connections": connections,
	}})
}

func (c *StorageController) Statistics(ctx httpcontract.Context) httpcontract.Response {
	statistics, err := c.connections.Statistics()
	if err != nil {
		return ctx.Response().Status(http.StatusInternalServerError).Json(httpcontract.Json{"code": "STORAGE_STATISTICS_UNAVAILABLE"})
	}
	return ctx.Response().Success().Json(httpcontract.Json{"data": statistics})
}

func (c *StorageController) Update(ctx httpcontract.Context) httpcontract.Response {
	var input storageservices.ConnectionInput
	if err := ctx.Request().Bind(&input); err != nil {
		return ctx.Response().Status(http.StatusUnprocessableEntity).Json(httpcontract.Json{"code": "STORAGE_CONFIGURATION_INVALID"})
	}
	providerCode := strings.TrimSpace(ctx.Request().Route("provider"))
	connection, err := c.connections.Save(providerCode, input)
	if err != nil {
		return storageError(ctx, err)
	}
	recordStorageAudit(ctx, "storage.connection.updated", providerCode, map[string]any{"enabled": connection.Enabled, "is_primary": connection.IsPrimary, "status": connection.Status})
	return ctx.Response().Success().Json(httpcontract.Json{"data": connection})
}

func (c *StorageController) Test(ctx httpcontract.Context) httpcontract.Response {
	providerCode := strings.TrimSpace(ctx.Request().Route("provider"))
	connection, err := c.connections.Test(providerCode)
	if err != nil {
		if errors.Is(err, storageservices.ErrProviderAdapterPending) {
			return ctx.Response().Status(http.StatusNotImplemented).Json(httpcontract.Json{"code": "STORAGE_PROVIDER_ADAPTER_PENDING"})
		}
		return storageError(ctx, err)
	}
	recordStorageAudit(ctx, "storage.connection.tested", providerCode, map[string]any{"status": connection.Status, "error_code": connection.LastErrorCode})
	return ctx.Response().Success().Json(httpcontract.Json{"data": connection})
}

func storageError(ctx httpcontract.Context, err error) httpcontract.Response {
	if errors.Is(err, storageservices.ErrStorageConnectionNotFound) {
		return ctx.Response().Status(http.StatusNotFound).Json(httpcontract.Json{"code": "STORAGE_CONNECTION_NOT_FOUND"})
	}
	if errors.Is(err, storageservices.ErrStorageConnectionInvalid) {
		return ctx.Response().Status(http.StatusUnprocessableEntity).Json(httpcontract.Json{"code": "STORAGE_CONFIGURATION_INVALID"})
	}
	if errors.Is(err, storageservices.ErrProviderAdapterPending) {
		return ctx.Response().Status(http.StatusConflict).Json(httpcontract.Json{"code": "STORAGE_PROVIDER_ADAPTER_PENDING"})
	}
	return ctx.Response().Status(http.StatusInternalServerError).Json(httpcontract.Json{"code": "STORAGE_CONNECTIONS_UNAVAILABLE"})
}

func recordStorageAudit(ctx httpcontract.Context, action, providerCode string, metadata map[string]any) {
	identity, err := facades.Auth(ctx).ID()
	if err != nil {
		return
	}
	userID, err := strconv.ParseUint(identity, 10, 32)
	if err != nil || userID == 0 {
		return
	}
	metadata["provider_code"] = providerCode
	_ = auditservices.NewAuditService().Record(uint(userID), action, metadata)
}
