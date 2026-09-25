package controllers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSanitizeFallbackImageURLKeepsOnlyImageSafeHTTPOrSitePaths(t *testing.T) {
	require.Equal(t, "https://img.example.com/fallback.svg", sanitizeFallbackImageURL("https://img.example.com/fallback.svg"))
	require.Equal(t, "/public/fallback.svg", sanitizeFallbackImageURL("/public/fallback.svg"))
	require.Empty(t, sanitizeFallbackImageURL("javascript:alert(1)"))
	require.Empty(t, sanitizeFallbackImageURL("data:text/html,<svg></svg>"))
}
