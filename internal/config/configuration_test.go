package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setRequiredEnv(t *testing.T) {
	t.Helper()
	t.Setenv("ONECHANNEL_SECRET", "oc-secret")
	t.Setenv("META_SECRET", "meta-secret")
	t.Setenv("POSTGRES_HOST", "localhost")
	t.Setenv("POSTGRES_DATABASE", "1channel_dev")
	t.Setenv("POSTGRES_USERNAME", "dev")
	t.Setenv("POSTGRES_PASSWORD", "dev")
	t.Setenv("NATS_HOST", "localhost")
	t.Setenv("NATS_USERNAME", "dev")
	t.Setenv("NATS_PASSWORD", "dev")
}

func TestNewConfiguration(t *testing.T) {
	tests := []struct {
		title         string
		setup         func(t *testing.T)
		expected      func(t *testing.T, cfg Configuration)
		expectedError string
	}{
		{
			title: "success - required vars with documented defaults",
			expected: func(t *testing.T, cfg Configuration) {
				t.Helper()
				assert.Equal(t, "1channel", cfg.AppName())
				assert.Equal(t, "8080", cfg.HTTPPort())
				assert.Equal(t, "info", cfg.LogLevel())
				assert.Equal(t, "oc-secret", cfg.OneChannelSecret())
				assert.Equal(t, "meta-secret", cfg.MetaSecret())
				assert.Equal(t, "localhost", cfg.PostgresHost())
				assert.Equal(t, uint16(5432), cfg.PostgresPort())
				assert.Equal(t, "1channel_dev", cfg.PostgresDatabase())
				assert.Equal(t, "dev", cfg.PostgresUsername())
				assert.Equal(t, "dev", cfg.PostgresPassword())
				assert.Equal(t, int32(4), cfg.PostgresMaxConns())
				assert.Equal(t, int32(0), cfg.PostgresMinConns())
				assert.Equal(t, time.Hour, cfg.PostgresMaxConnLifeTime())
				assert.Equal(t, 30*time.Minute, cfg.PostgresMaxConnIdleTime())
				assert.Equal(t, "localhost", cfg.NatsHost())
				assert.Equal(t, 4222, cfg.NatsPort())
				assert.Equal(t, "dev", cfg.NatsUsername())
				assert.Equal(t, "dev", cfg.NatsPassword())
				assert.Equal(t, 2*time.Second, cfg.NatsTimeout())
				assert.Equal(t, -1, cfg.NatsMaxReconnects())
				assert.Equal(t, 2*time.Second, cfg.NatsReconnectWait())
				assert.Equal(t, 2*time.Minute, cfg.NatsPingInterval())
				assert.Equal(t, 2, cfg.NatsMaxPingsOut())
			},
		},
		{
			title: "success - explicit vars override defaults",
			setup: func(t *testing.T) {
				t.Setenv("HTTP_PORT", "9090")
				t.Setenv("LOG_LEVEL", "debug")
				t.Setenv("POSTGRES_PORT", "6543")
				t.Setenv("POSTGRES_MAX_CONN_LIFETIME", "30m")
				t.Setenv("NATS_PORT", "14222")
			},
			expected: func(t *testing.T, cfg Configuration) {
				t.Helper()
				assert.Equal(t, "9090", cfg.HTTPPort())
				assert.Equal(t, "debug", cfg.LogLevel())
				assert.Equal(t, uint16(6543), cfg.PostgresPort())
				assert.Equal(t, 30*time.Minute, cfg.PostgresMaxConnLifeTime())
				assert.Equal(t, 14222, cfg.NatsPort())
			},
		},
		{
			title: "success - secrets are exposed verbatim",
			setup: func(t *testing.T) {
				t.Setenv("ONECHANNEL_SECRET", "super-secret-1")
				t.Setenv("META_SECRET", "super-secret-2")
			},
			expected: func(t *testing.T, cfg Configuration) {
				t.Helper()
				assert.Equal(t, "super-secret-1", cfg.OneChannelSecret())
				assert.Equal(t, "super-secret-2", cfg.MetaSecret())
			},
		},
		{
			title:         "failure - rejects garbage value in numeric postgres port",
			setup:         func(t *testing.T) { t.Setenv("POSTGRES_PORT", "not-a-number") },
			expectedError: `env: parse error on field "PostgresPortEnv" of type "uint16": strconv.ParseUint: parsing "not-a-number": invalid syntax`,
		},
		{
			title:         "failure - rejects empty onechannel secret",
			setup:         func(t *testing.T) { t.Setenv("ONECHANNEL_SECRET", "") },
			expectedError: `env: environment variable "ONECHANNEL_SECRET" should not be empty`,
		},
		{
			title:         "failure - rejects empty meta secret",
			setup:         func(t *testing.T) { t.Setenv("META_SECRET", "") },
			expectedError: `env: environment variable "META_SECRET" should not be empty`,
		},
		{
			title:         "failure - rejects empty postgres host",
			setup:         func(t *testing.T) { t.Setenv("POSTGRES_HOST", "") },
			expectedError: `env: environment variable "POSTGRES_HOST" should not be empty`,
		},
		{
			title:         "failure - rejects empty postgres database",
			setup:         func(t *testing.T) { t.Setenv("POSTGRES_DATABASE", "") },
			expectedError: `env: environment variable "POSTGRES_DATABASE" should not be empty`,
		},
		{
			title:         "failure - rejects empty postgres username",
			setup:         func(t *testing.T) { t.Setenv("POSTGRES_USERNAME", "") },
			expectedError: `env: environment variable "POSTGRES_USERNAME" should not be empty`,
		},
		{
			title:         "failure - rejects empty postgres password",
			setup:         func(t *testing.T) { t.Setenv("POSTGRES_PASSWORD", "") },
			expectedError: `env: environment variable "POSTGRES_PASSWORD" should not be empty`,
		},
		{
			title:         "failure - rejects empty nats host",
			setup:         func(t *testing.T) { t.Setenv("NATS_HOST", "") },
			expectedError: `env: environment variable "NATS_HOST" should not be empty`,
		},
		{
			title:         "failure - rejects empty nats username",
			setup:         func(t *testing.T) { t.Setenv("NATS_USERNAME", "") },
			expectedError: `env: environment variable "NATS_USERNAME" should not be empty`,
		},
		{
			title:         "failure - rejects empty nats password",
			setup:         func(t *testing.T) { t.Setenv("NATS_PASSWORD", "") },
			expectedError: `env: environment variable "NATS_PASSWORD" should not be empty`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			setRequiredEnv(t)
			if tt.setup != nil {
				tt.setup(t)
			}

			// Act
			cfg, err := NewConfiguration()

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
				return
			}
			require.NoError(t, err)
			require.NotNil(t, cfg)
			tt.expected(t, cfg)
		})
	}
}
