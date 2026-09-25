package actions

import (
	"errors"
	"strings"
)

var ErrInvalidResolvePayload = errors.New("invalid report resolve payload")
var ErrReportNotFound = errors.New("report not found")
var ErrReportAlreadyResolved = errors.New("report already resolved")

func ParseResolvePayload(p map[string]any) (string, string, error) {
	if len(p) != 2 {
		return "", "", ErrInvalidResolvePayload
	}
	a, ok := p["action"].(string)
	if !ok {
		return "", "", ErrInvalidResolvePayload
	}
	n, ok := p["resolution"].(string)
	if !ok {
		return "", "", ErrInvalidResolvePayload
	}
	a, n = strings.TrimSpace(a), strings.TrimSpace(n)
	if n == "" || len(n) > 2000 || (a != "dismiss" && a != "hide_media" && a != "restore_media") {
		return "", "", ErrInvalidResolvePayload
	}
	return a, n, nil
}
