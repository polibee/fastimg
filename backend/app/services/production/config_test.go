package production

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateProductionConfigRejectsUnsafeDefaults(t *testing.T) {
	env := map[string]string{
		"APP_ENV": "production", "APP_DEBUG": "true", "APP_URL": "http://127.0.0.1:53083",
		"APP_KEY": "short", "JWT_SECRET": "short", "DB_CONNECTION": "postgres", "DB_SSLMODE": "disable",
	}
	require.Error(t, Validate(env))
}

func TestValidateProductionConfigAcceptsReverseProxyDeployment(t *testing.T) {
	env := map[string]string{
		"APP_ENV": "production", "APP_DEBUG": "false", "APP_URL": "https://fastimg.example",
		"APP_KEY": "01234567890123456789012345678901", "JWT_SECRET": "01234567890123456789012345678901",
		"DB_CONNECTION": "postgres", "DB_SSLMODE": "verify-full", "CORS_ALLOWED_ORIGINS": "https://fastimg.example",
	}
	require.NoError(t, Validate(env))
}
