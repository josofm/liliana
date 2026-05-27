package config

import (
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

	cfg, err := NewConfig()
	require.NoError(t, err)

	assert.Equal(t, "test", cfg.App.Environment)
	assert.Equal(t, "9090", cfg.HTTP.Port)
	assert.Equal(t, "warn", cfg.Log.Level)
	assert.Equal(t, "env-secret", cfg.JWT.SecretKey)
	assert.Equal(t, "postgres://example", cfg.DB.URL)
	assert.True(t, cfg.DB.AutoMigrate)
}

func TestNewConfig_RenderPortFallback(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("HTTP_PORT", "")
	t.Setenv("PORT", "10000")

	cfg, err := NewConfig()
	require.NoError(t, err)

	assert.Equal(t, "10000", cfg.HTTP.Port)
}

func TestNewConfig_HTTPPortOverridesRenderPort(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("HTTP_PORT", "9090")
	t.Setenv("PORT", "10000")

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
