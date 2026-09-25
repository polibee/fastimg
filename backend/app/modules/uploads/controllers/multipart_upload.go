package controllers

import (
	"io"
	"net/http"
	"strings"
)

var (
	errUploadFileRequired  = uploadInputError{"UPLOAD_FILE_REQUIRED"}
	errMultipleUploadFiles = uploadInputError{"UPLOAD_SINGLE_FILE_ONLY"}
	errUploadFileTooLarge  = uploadInputError{"UPLOAD_FILE_TOO_LARGE"}
	errUploadFileEmpty     = uploadInputError{"UPLOAD_FILE_EMPTY"}
	errBatchFilesTooMany   = uploadInputError{"UPLOAD_BATCH_TOO_MANY_FILES"}
	errBatchTotalTooLarge  = uploadInputError{"UPLOAD_BATCH_TOO_LARGE"}
)

type uploadInputError struct {
	code string
}

func (e uploadInputError) Error() string { return e.code }

type multipartUpload struct {
	OriginalName string
	ContentType  string
	Content      []byte
}

func readMultipartUpload(request *http.Request, maxFileBytes int64) (multipartUpload, error) {
	if maxFileBytes <= 0 {
		return multipartUpload{}, errUploadFileTooLarge
	}
	if request.MultipartForm == nil {
		parseErr := request.ParseMultipartForm(4 << 20)
		if request.MultipartForm != nil {
			defer request.MultipartForm.RemoveAll()
		}
		if parseErr != nil {
			return multipartUpload{}, parseErr
		}
	} else {
		defer request.MultipartForm.RemoveAll()
	}
	if request.MultipartForm == nil {
		return multipartUpload{}, errUploadFileRequired
	}
	files := request.MultipartForm.File["file"]
	if len(files) == 0 {
		return multipartUpload{}, errUploadFileRequired
	}
	if len(files) != 1 {
		return multipartUpload{}, errMultipleUploadFiles
	}
	header := files[0]
	if header.Size > maxFileBytes {
		return multipartUpload{}, errUploadFileTooLarge
	}
	file, err := header.Open()
	if err != nil {
		return multipartUpload{}, err
	}
	defer file.Close()
	content, err := io.ReadAll(io.LimitReader(file, maxFileBytes+1))
	if err != nil {
		return multipartUpload{}, err
	}
	if int64(len(content)) > maxFileBytes {
		return multipartUpload{}, errUploadFileTooLarge
	}
	if len(content) == 0 {
		return multipartUpload{}, errUploadFileEmpty
	}
	return multipartUpload{
		OriginalName: header.Filename,
		ContentType:  strings.TrimSpace(header.Header.Get("Content-Type")),
		Content:      content,
	}, nil
}

func readMultipartUploads(request *http.Request, maxFileBytes int64, maxFiles int, maxTotalBytes int64) ([]multipartUpload, error) {
	if maxFileBytes <= 0 || maxFiles <= 0 || maxTotalBytes <= 0 {
		return nil, errBatchTotalTooLarge
	}
	if request.MultipartForm == nil {
		parseErr := request.ParseMultipartForm(4 << 20)
		if request.MultipartForm != nil {
			defer request.MultipartForm.RemoveAll()
		}
		if parseErr != nil {
			return nil, parseErr
		}
	} else {
		defer request.MultipartForm.RemoveAll()
	}
	if request.MultipartForm == nil {
		return nil, errUploadFileRequired
	}
	files := request.MultipartForm.File["files[]"]
	if len(files) == 0 {
		files = request.MultipartForm.File["files"]
	}
	if len(files) == 0 {
		return nil, errUploadFileRequired
	}
	if len(files) > maxFiles {
		return nil, errBatchFilesTooMany
	}
	uploads := make([]multipartUpload, 0, len(files))
	var totalBytes int64
	for _, header := range files {
		if header.Size > maxFileBytes {
			return nil, errUploadFileTooLarge
		}
		file, err := header.Open()
		if err != nil {
			return nil, err
		}
		content, readErr := io.ReadAll(io.LimitReader(file, maxFileBytes+1))
		closeErr := file.Close()
		if readErr != nil {
			return nil, readErr
		}
		if closeErr != nil {
			return nil, closeErr
		}
		if int64(len(content)) > maxFileBytes {
			return nil, errUploadFileTooLarge
		}
		if len(content) == 0 {
			return nil, errUploadFileEmpty
		}
		if totalBytes > maxTotalBytes-int64(len(content)) {
			return nil, errBatchTotalTooLarge
		}
		totalBytes += int64(len(content))
		uploads = append(uploads, multipartUpload{
			OriginalName: header.Filename,
			ContentType:  strings.TrimSpace(header.Header.Get("Content-Type")),
			Content:      content,
		})
	}
	return uploads, nil
}
