package media

import (
	"errors"
	"strings"
)

var ErrVariantFormatUnavailable = errors.New("requested image variant format is unavailable")

type VariantRequest struct {
	Format  string
	Width   int
	Quality int
}

func ValidateVariantRequest(format string, width, quality int) (VariantRequest, error) {
	format = strings.ToLower(strings.TrimSpace(format))
	if format == "" {
		format = "jpeg"
	}
	if format == "webp" || format == "avif" {
		return VariantRequest{}, ErrVariantFormatUnavailable
	}
	if format != "jpeg" && format != "png" {
		return VariantRequest{}, ErrInvalidVariantName
	}
	if width < 1 || width > 4096 {
		return VariantRequest{}, errors.New("variant width is outside the allowed range")
	}
	if quality == 0 {
		quality = 80
	}
	if quality < 1 || quality > 100 {
		return VariantRequest{}, errors.New("variant quality is outside the allowed range")
	}
	return VariantRequest{Format: format, Width: width, Quality: quality}, nil
}
