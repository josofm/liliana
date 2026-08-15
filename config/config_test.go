package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig_EnvOverrides(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("HTTP_PORT", "9090")
	t.Setenv("LOG_LEVEL", "warn")
	t.Setenv("JWT_SECRET_KEY", "env-secret")
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("AUTO_MIGRATE", "true")
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:5173")

	cfg, err := NewConfig()
	require.NoError(t, err)

	assert.Equal(t, "test", cfg.App.Environment)
	assert.Equal(t, "9090", cfg.HTTP.Port)
	assert.Equal(t, "warn", cfg.Log.Level)
	assert.Equal(t, "env-secret", cfg.JWT.SecretKey)
	assert.Equal(t, "postgres://example", cfg.DB.URL)
	assert.True(t, cfg.DB.AutoMigrate)
	assert.Equal(t, "http://localhost:5173", cfg.HTTP.CORSAllowedOrigins)
}

func TestNewConfig_RenderPortFallback(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("HTTP_PORT", "")
	t.Setenv("PORT", "10000")
	t.Setenv("JWT_SECRET_KEY", "test-secret")

	cfg, err := NewConfig()
	require.NoError(t, err)

	assert.Equal(t, "10000", cfg.HTTP.Port)
}

func TestNewConfig_HTTPPortOverridesRenderPort(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("HTTP_PORT", "9090")
	t.Setenv("PORT", "10000")
	t.Setenv("JWT_SECRET_KEY", "test-secret")

	cfg, err := NewConfig()
	require.NoError(t, err)

	assert.Equal(t, "9090", cfg.HTTP.Port)
}

func TestNewConfig_ProductionRequiresSecrets(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("JWT_SECRET_KEY", "dev-secret-change-me")

	_, err := NewConfig()
	require.Error(t, err)
}

func TestNewConfig_ProductionFromEnvWithoutConfigFile(t *testing.T) {
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(originalDir))
	})

	t.Setenv("APP_ENV", "production")
	t.Setenv("HTTP_PORT", "")
	t.Setenv("PORT", "10000")
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("JWT_SECRET_KEY", "production-secret")

	cfg, err := NewConfig()
	require.NoError(t, err)

	assert.Equal(t, "production", cfg.App.Environment)
	assert.Equal(t, "10000", cfg.HTTP.Port)
	assert.Equal(t, "postgres://example", cfg.DB.URL)
	assert.Equal(t, "production-secret", cfg.JWT.SecretKey)
}
