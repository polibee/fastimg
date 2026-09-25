package middleware

import (
	"io"
	"net/http"

	httpcontract "github.com/goravel/framework/contracts/http"
)

const maxUploadRequestBytes int64 = 10_250_000 // 10 MB file plus bounded multipart fields and headers.
const maxBatchUploadRequestBytes int64 = 50_250_000

type uploadBodyLimitMiddleware struct {
	maximum int64
}

func (uploadBodyLimitMiddleware) Signature() string {
	return "fastimg:upload-body-limit"
}

func (middleware uploadBodyLimitMiddleware) Handle(ctx httpcontract.Context) {
	origin := ctx.Request().Origin()
	if origin == nil || origin.Body == nil {
		ctx.Request().Next()
		return
	}
	maximum := middleware.maximum
	if maximum <= 0 {
		maximum = maxUploadRequestBytes
	}
	if origin.ContentLength > maximum {
		_ = ctx.Response().Status(http.StatusRequestEntityTooLarge).Json(httpcontract.Json{
			"code": "UPLOAD_REQUEST_TOO_LARGE",
		}).Abort()
		return
	}
	origin.Body = limitUploadBody(origin.Body, maximum)
	ctx.Request().Next()
}

func UploadBodyLimit() httpcontract.Middleware {
	return uploadBodyLimitMiddleware{maximum: maxUploadRequestBytes}
}

func UploadBatchBodyLimit() httpcontract.Middleware {
	return uploadBodyLimitMiddleware{maximum: maxBatchUploadRequestBytes}
}

func limitUploadBody(body io.ReadCloser, maximum int64) io.ReadCloser {
	return http.MaxBytesReader(nil, body, maximum)
}
