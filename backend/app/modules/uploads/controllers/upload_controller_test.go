package controllers

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	"github.com/stretchr/testify/require"

	"goravel/app/services/media"
	"goravel/app/services/quota"
)

func TestUploadErrorResponseUsesStableClientCodes(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{name: "bad image", err: media.ErrImageFormatMismatch, wantStatus: http.StatusUnprocessableEntity, wantCode: "IMAGE_FORMAT_INVALID"},
		{name: "storage quota", err: quota.ErrStorageQuotaExceeded, wantStatus: http.StatusInsufficientStorage, wantCode: "STORAGE_QUOTA_EXCEEDED"},
		{name: "daily upload cap", err: media.ErrDailyUploadLimit, wantStatus: http.StatusTooManyRequests, wantCode: "DAILY_UPLOAD_LIMIT_REACHED"},
		{name: "monthly API upload cap", err: media.ErrMonthlyAPIUploadLimit, wantStatus: http.StatusTooManyRequests, wantCode: "MONTHLY_API_UPLOAD_LIMIT_REACHED"},
		{name: "monthly transform cap", err: media.ErrMonthlyTransformLimit, wantStatus: http.StatusTooManyRequests, wantCode: "MONTHLY_TRANSFORM_LIMIT_REACHED"},
		{name: "duplicate key", err: media.ErrIdempotencyConflict, wantStatus: http.StatusConflict, wantCode: "IDEMPOTENCY_KEY_REUSED"},
		{name: "disabled account", err: media.ErrAccountUnavailable, wantStatus: http.StatusForbidden, wantCode: "ACCOUNT_UPLOAD_DISABLED"},
		{name: "request body cap", err: &http.MaxBytesError{Limit: 100}, wantStatus: http.StatusRequestEntityTooLarge, wantCode: "UPLOAD_REQUEST_TOO_LARGE"},
		{name: "unknown", err: errors.New("internal"), wantStatus: http.StatusInternalServerError, wantCode: "UPLOAD_FAILED"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status, code := uploadErrorResponse(test.err)
			require.Equal(t, test.wantStatus, status)
			require.Equal(t, test.wantCode, code)
		})
	}
}

func TestUploadStatusURLUsesSessionID(t *testing.T) {
	require.Equal(t, "/api/v1/uploads/42", uploadStatusURL(42))
}

func TestReadMultipartUploadReturnsBoundedImageFile(t *testing.T) {
	request := multipartUploadRequest(t, []struct {
		name string
		data []byte
	}{{name: "photo.png", data: []byte("png-data")}})

	got, err := readMultipartUpload(request, 16)
	require.NoError(t, err)
	require.Equal(t, "photo.png", got.OriginalName)
	require.Equal(t, "image/png", got.ContentType)
	require.Equal(t, []byte("png-data"), got.Content)
}

func TestReadMultipartUploadRejectsMultipleFilesAndOversizedFile(t *testing.T) {
	request := multipartUploadRequest(t, []struct {
		name string
		data []byte
	}{{name: "one.png", data: []byte("1")}, {name: "two.png", data: []byte("2")}})
	_, err := readMultipartUpload(request, 16)
	require.ErrorIs(t, err, errMultipleUploadFiles)

	request = multipartUploadRequest(t, []struct {
		name string
		data []byte
	}{{name: "large.png", data: bytes.Repeat([]byte("x"), 17)}})
	_, err = readMultipartUpload(request, 16)
	require.ErrorIs(t, err, errUploadFileTooLarge)
}

func TestReadMultipartUploadsPreservesOrderAndRejectsBatchLimits(t *testing.T) {
	request := multipartUploadsRequest(t, []struct {
		name string
		data []byte
	}{{name: "first.png", data: []byte("first")}, {name: "second.jpg", data: []byte("second")}})

	uploads, err := readMultipartUploads(request, 8, 3, 12)
	require.NoError(t, err)
	require.Len(t, uploads, 2)
	require.Equal(t, "first.png", uploads[0].OriginalName)
	require.Equal(t, "second.jpg", uploads[1].OriginalName)

	request = multipartUploadsRequest(t, []struct {
		name string
		data []byte
	}{{name: "one.png", data: []byte("1")}, {name: "two.png", data: []byte("2")}, {name: "three.png", data: []byte("3")}})
	_, err = readMultipartUploads(request, 8, 2, 12)
	require.EqualError(t, err, "UPLOAD_BATCH_TOO_MANY_FILES")

	request = multipartUploadsRequest(t, []struct {
		name string
		data []byte
	}{{name: "large-a.png", data: []byte("123456")}, {name: "large-b.png", data: []byte("123456")}})
	_, err = readMultipartUploads(request, 8, 3, 10)
	require.EqualError(t, err, "UPLOAD_BATCH_TOO_LARGE")
}

func TestReadMultipartUploadsRejectsEmptyBatchAndEmptyFile(t *testing.T) {
	request := multipartUploadsRequest(t, nil)
	_, err := readMultipartUploads(request, 8, 3, 12)
	require.EqualError(t, err, "UPLOAD_FILE_REQUIRED")

	request = multipartUploadsRequest(t, []struct {
		name string
		data []byte
	}{{name: "empty.png", data: nil}})
	_, err = readMultipartUploads(request, 8, 3, 12)
	require.EqualError(t, err, "UPLOAD_FILE_EMPTY")
}

func TestBatchIdempotencyKeyIsStableAndIndependentPerFile(t *testing.T) {
	first := batchIdempotencyKey("batch-request-123", 0)
	second := batchIdempotencyKey("batch-request-123", 1)
	require.NotEqual(t, first, second)
	require.Equal(t, first, batchIdempotencyKey("batch-request-123", 0))
	require.LessOrEqual(t, len(first), 160)
}

func multipartUploadRequest(t *testing.T, files []struct {
	name string
	data []byte
}) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for _, file := range files {
		header := make(textproto.MIMEHeader)
		header.Set("Content-Disposition", `form-data; name="file"; filename="`+file.name+`"`)
		header.Set("Content-Type", "image/png")
		part, err := writer.CreatePart(header)
		require.NoError(t, err)
		_, err = part.Write(file.data)
		require.NoError(t, err)
	}
	require.NoError(t, writer.Close())
	request := httptest.NewRequest(http.MethodPost, "/api/v1/uploads", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return request
}

func multipartUploadsRequest(t *testing.T, files []struct {
	name string
	data []byte
}) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for _, file := range files {
		header := make(textproto.MIMEHeader)
		header.Set("Content-Disposition", `form-data; name="files[]"; filename="`+file.name+`"`)
		header.Set("Content-Type", "image/png")
		part, err := writer.CreatePart(header)
		require.NoError(t, err)
		_, err = part.Write(file.data)
		require.NoError(t, err)
	}
	require.NoError(t, writer.Close())
	request := httptest.NewRequest(http.MethodPost, "/api/v1/uploads/batch", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return request
}
