package controllers

import (
	"io"
	"net/http"

	httpcontract "github.com/goravel/framework/contracts/http"

	billing "goravel/app/services/billing"
	"goravel/app/services/billing/providers"
)

type WebhookController struct{ service *billing.WebhookService }

func NewWebhookController() *WebhookController {
	return &WebhookController{service: billing.NewWebhookService(billing.DefaultGatewayRegistry())}
}

func (c *WebhookController) Receive(ctx httpcontract.Context) httpcontract.Response {
	billing.RefreshGatewayRegistry()
	body, err := io.ReadAll(io.LimitReader(ctx.Request().Origin().Body, 1<<20+1))
	if err != nil || len(body) > 1<<20 {
		return ctx.Response().Status(http.StatusRequestEntityTooLarge).Json(httpcontract.Json{"code": "WEBHOOK_BODY_TOO_LARGE"})
	}
	headers := make(map[string]string, len(ctx.Request().Headers()))
	for name, values := range ctx.Request().Headers() {
		if len(values) > 0 {
			headers[name] = values[0]
		}
	}
	event, duplicate, err := c.service.Ingest(ctx.Context(), ctx.Request().Route("gateway"), providers.WebhookRequest{Headers: headers, Body: body})
	if err != nil {
		return ctx.Response().Status(http.StatusUnauthorized).Json(httpcontract.Json{"code": "WEBHOOK_REJECTED"})
	}
	return ctx.Response().Success().Json(httpcontract.Json{"data": httpcontract.Json{"event_id": event.EventID, "accepted": true, "duplicate": duplicate}})
}
