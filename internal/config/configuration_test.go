package config

import (
	"testing"
	"time"
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
		name    string
		setup   func(t *testing.T)
		wantErr bool
		verify  func(t *testing.T, cfg Configuration)
	}{
		{
			name:    "only required vars yields documented defaults",
			setup:   func(t *testing.T) {},
			wantErr: false,
			verify: func(t *testing.T, cfg Configuration) {
				t.Helper()
				if cfg.AppName() != "1channel" {
					t.Fatalf("expected AppName 1channel, got %q", cfg.AppName())
				}
				if cfg.HTTPPort() != "8080" {
					t.Fatalf("expected HTTPPort default 8080, got %q", cfg.HTTPPort())
				}
				if cfg.LogLevel() != "info" {
					t.Fatalf("expected LogLevel default info, got %q", cfg.LogLevel())
				}
				if cfg.OneChannelSecret() != "oc-secret" {
					t.Fatalf("expected OneChannelSecret oc-secret, got %q", cfg.OneChannelSecret())
				}
				if cfg.MetaSecret() != "meta-secret" {
					t.Fatalf("expected MetaSecret meta-secret, got %q", cfg.MetaSecret())
				}
				if cfg.PostgresHost() != "localhost" {
					t.Fatalf("expected PostgresHost localhost, got %q", cfg.PostgresHost())
				}
				if cfg.PostgresDatabase() != "1channel_dev" {
					t.Fatalf("expected PostgresDatabase 1channel_dev, got %q", cfg.PostgresDatabase())
				}
				if cfg.PostgresUsername() != "dev" {
					t.Fatalf("expected PostgresUsername dev, got %q", cfg.PostgresUsername())
				}
				if cfg.PostgresPassword() != "dev" {
					t.Fatalf("expected PostgresPassword dev, got %q", cfg.PostgresPassword())
				}
				if cfg.PostgresPort() != 5432 {
					t.Fatalf("expected PostgresPort default 5432, got %d", cfg.PostgresPort())
				}
				if cfg.PostgresMaxConns() != 4 {
					t.Fatalf("expected PostgresMaxConns default 4, got %d", cfg.PostgresMaxConns())
				}
				if cfg.PostgresMinConns() != 0 {
					t.Fatalf("expected PostgresMinConns default 0, got %d", cfg.PostgresMinConns())
				}
				if cfg.PostgresMaxConnLifeTime() != time.Hour {
					t.Fatalf("expected PostgresMaxConnLifeTime default 1h, got %v", cfg.PostgresMaxConnLifeTime())
				}
				if cfg.PostgresMaxConnIdleTime() != 30*time.Minute {
					t.Fatalf("expected PostgresMaxConnIdleTime default 30m, got %v", cfg.PostgresMaxConnIdleTime())
				}
				if cfg.NatsHost() != "localhost" {
					t.Fatalf("expected NatsHost localhost, got %q", cfg.NatsHost())
				}
				if cfg.NatsPort() != 4222 {
					t.Fatalf("expected NatsPort default 4222, got %d", cfg.NatsPort())
				}
				if cfg.NatsUsername() != "dev" {
					t.Fatalf("expected NatsUsername dev, got %q", cfg.NatsUsername())
				}
				if cfg.NatsPassword() != "dev" {
					t.Fatalf("expected NatsPassword dev, got %q", cfg.NatsPassword())
				}
				if cfg.NatsTimeout() != 2*time.Second {
					t.Fatalf("expected NatsTimeout default 2s, got %v", cfg.NatsTimeout())
				}
				if cfg.NatsMaxReconnects() != -1 {
					t.Fatalf("expected NatsMaxReconnects default -1, got %d", cfg.NatsMaxReconnects())
				}
				if cfg.NatsReconnectWait() != 2*time.Second {
					t.Fatalf("expected NatsReconnectWait default 2s, got %v", cfg.NatsReconnectWait())
				}
				if cfg.NatsPingInterval() != 2*time.Minute {
					t.Fatalf("expected NatsPingInterval default 2m, got %v", cfg.NatsPingInterval())
				}
				if cfg.NatsMaxPingsOut() != 2 {
					t.Fatalf("expected NatsMaxPingsOut default 2, got %d", cfg.NatsMaxPingsOut())
				}
			},
		},
		{
			name: "explicit vars override defaults",
			setup: func(t *testing.T) {
				t.Setenv("HTTP_PORT", "9090")
				t.Setenv("LOG_LEVEL", "debug")
				t.Setenv("POSTGRES_PORT", "6543")
				t.Setenv("POSTGRES_MAX_CONN_LIFETIME", "30m")
				t.Setenv("NATS_PORT", "14222")
			},
			wantErr: false,
			verify: func(t *testing.T, cfg Configuration) {
				t.Helper()
				if cfg.HTTPPort() != "9090" {
					t.Fatalf("expected HTTPPort 9090, got %q", cfg.HTTPPort())
				}
				if cfg.LogLevel() != "debug" {
					t.Fatalf("expected LogLevel debug, got %q", cfg.LogLevel())
				}
				if cfg.PostgresPort() != 6543 {
					t.Fatalf("expected PostgresPort 6543, got %d", cfg.PostgresPort())
				}
				if cfg.PostgresMaxConnLifeTime() != 30*time.Minute {
					t.Fatalf("expected PostgresMaxConnLifeTime 30m, got %v", cfg.PostgresMaxConnLifeTime())
				}
				if cfg.NatsPort() != 14222 {
					t.Fatalf("expected NatsPort 14222, got %d", cfg.NatsPort())
				}
			},
		},
		{
			name: "secrets are exposed verbatim",
			setup: func(t *testing.T) {
				t.Setenv("ONECHANNEL_SECRET", "super-secret-1")
				t.Setenv("META_SECRET", "super-secret-2")
			},
			wantErr: false,
			verify: func(t *testing.T, cfg Configuration) {
				t.Helper()
				if cfg.OneChannelSecret() != "super-secret-1" {
					t.Fatalf("expected OneChannelSecret override, got %q", cfg.OneChannelSecret())
				}
				if cfg.MetaSecret() != "super-secret-2" {
					t.Fatalf("expected MetaSecret override, got %q", cfg.MetaSecret())
				}
			},
		},
		{
			name: "numeric field with garbage value fails parse",
			setup: func(t *testing.T) {
				t.Setenv("POSTGRES_PORT", "not-a-number")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setRequiredEnv(t)
			tt.setup(t)

			cfg, err := NewConfiguration()

			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if tt.verify != nil {
				tt.verify(t, cfg)
			}
		})
	}
}

func TestNewConfigurationRequiredVars(t *testing.T) {
	requiredEnvVars := []string{
		"ONECHANNEL_SECRET",
		"META_SECRET",
		"POSTGRES_HOST",
		"POSTGRES_DATABASE",
		"POSTGRES_USERNAME",
		"POSTGRES_PASSWORD",
		"NATS_HOST",
		"NATS_USERNAME",
		"NATS_PASSWORD",
	}

	for _, envVar := range requiredEnvVars {
		t.Run(envVar+" missing rejects configuration", func(t *testing.T) {
			setRequiredEnv(t)
			t.Setenv(envVar, "")

			if _, err := NewConfiguration(); err == nil {
				t.Fatalf("expected error when %s is empty", envVar)
			}
		})
	}
}
