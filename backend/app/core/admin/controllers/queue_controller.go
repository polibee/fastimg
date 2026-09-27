package controllers

import (
	"net/http"
	"strconv"
	"strings"

	httpcontract "github.com/goravel/framework/contracts/http"
	queuecontract "github.com/goravel/framework/contracts/queue"

	"goravel/app/facades"
	auditservices "goravel/app/services/audit"
)

const (
	QueueViewPermission  = "admin.tasks.view"
	QueueRetryPermission = "admin.tasks.retry"
)

type QueueController struct{}

func NewQueueController() *QueueController { return &QueueController{} }

type queueTaskItem struct {
	UUID       string `json:"uuid"`
	Connection string `json:"connection"`
	Queue      string `json:"queue"`
	Signature  string `json:"signature"`
	FailedAt   string `json:"failed_at"`
}

type queueTaskMeta struct {
	Page     int   `json:"page"`
	PerPage  int   `json:"per_page"`
	Total    int64 `json:"total"`
	LastPage int   `json:"last_page"`
}

type queueTaskPage struct {
	Data []queueTaskItem `json:"data"`
	Meta queueTaskMeta   `json:"meta"`
}

func (c *QueueController) Index(ctx httpcontract.Context) httpcontract.Response {
	query := resourceListQuery{
		Page:    positiveInt(ctx.Request().Query("page", "1"), 1),
		PerPage: positiveInt(ctx.Request().Query("per_page", "20"), 20),
	}
	if query.PerPage > 100 {
		query.PerPage = 100
	}
	failedJobs, err := facades.Queue().Failer().All()
	if err != nil {
		return ctx.Response().Status(http.StatusServiceUnavailable).Json(httpcontract.Json{"code": "QUEUE_TASKS_UNAVAILABLE"})
	}
	items := make([]queueTaskItem, 0, len(failedJobs))
	for _, failedJob := range failedJobs {
		items = append(items, queueTaskItemFromFailedJob(failedJob))
	}
	page := paginateQueueTasks(items, query.Page, query.PerPage)
	return ctx.Response().Success().Json(httpcontract.Json{"data": page.Data, "meta": page.Meta})
}

func (c *QueueController) Retry(ctx httpcontract.Context) httpcontract.Response {
	uuid := strings.TrimSpace(ctx.Request().Route("uuid"))
	if uuid == "" {
		return ctx.Response().Status(http.StatusUnprocessableEntity).Json(httpcontract.Json{"code": "QUEUE_TASK_INVALID"})
	}
	failedJobs, err := facades.Queue().Failer().Get("", "", []string{uuid})
	if err != nil {
		return ctx.Response().Status(http.StatusServiceUnavailable).Json(httpcontract.Json{"code": "QUEUE_TASKS_UNAVAILABLE"})
	}
	if len(failedJobs) == 0 {
		return ctx.Response().Status(http.StatusNotFound).Json(httpcontract.Json{"code": "QUEUE_TASK_NOT_FOUND"})
	}
	if err := failedJobs[0].Retry(); err != nil {
		return ctx.Response().Status(http.StatusInternalServerError).Json(httpcontract.Json{"code": "QUEUE_TASK_RETRY_FAILED"})
	}
	if userID, err := strconv.ParseUint(strings.TrimSpace(mustAuthID(ctx)), 10, 32); err == nil && userID > 0 {
		_ = auditservices.NewAuditService().Record(uint(userID), "queue.task.retry", map[string]any{
			"uuid":      uuid,
			"signature": failedJobs[0].Signature(),
			"queue":     failedJobs[0].Queue(),
		})
	}
	return ctx.Response().Success().Json(httpcontract.Json{"data": map[string]any{"uuid": uuid, "status": "queued"}})
}

func queueTaskItemFromFailedJob(failedJob queuecontract.FailedJob) queueTaskItem {
	item := queueTaskItem{
		UUID:       failedJob.UUID(),
		Connection: failedJob.Connection(),
		Queue:      failedJob.Queue(),
		Signature:  failedJob.Signature(),
	}
	if item.Signature == "" {
		item.Signature = "unknown"
	}
	if failedAt := failedJob.FailedAt(); failedAt != nil {
		item.FailedAt = failedAt.ToDateTimeString()
	}
	return item
}

func paginateQueueTasks(items []queueTaskItem, page, perPage int) queueTaskPage {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	total := int64(len(items))
	lastPage := int((total + int64(perPage) - 1) / int64(perPage))
	if lastPage == 0 {
		lastPage = 1
	}
	if page > lastPage {
		page = lastPage
	}
	start := (page - 1) * perPage
	if start > len(items) {
		start = len(items)
	}
	end := start + perPage
	if end > len(items) {
		end = len(items)
	}
	return queueTaskPage{
		Data: items[start:end],
		Meta: queueTaskMeta{Page: page, PerPage: perPage, Total: total, LastPage: lastPage},
	}
}

func mustAuthID(ctx httpcontract.Context) string {
	identity, err := facades.Auth(ctx).ID()
	if err != nil {
		return ""
	}
	return identity
}
