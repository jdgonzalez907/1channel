package config

import (
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Configuration interface {
	AppName() string

	HTTPPort() string

	OneChannelSecret() string
	MetaSecret() string

	PostgresURL() string
	PostgresUsername() string
	PostgresPassword() string
	PostgresPort() uint16
	PostgresMaxConns() int32
	PostgresMinConns() int32
	PostgresMaxConnLifeTime() time.Duration
	PostgresMaxConnIdleTime() time.Duration

	NatsURL() string
	NatsPort() int
	NatsUsername() string
	NatsPassword() string
	NatsTimeout() time.Duration
	NatsMaxReconnects() int
	NatsReconnectWait() time.Duration
	NatsRetryOnFailedConnect() bool
	NatsPingInterval() time.Duration
	NatsMaxPingsOut() int
}

type configuration struct {
	HTTPPortEnv string `env:"HTTP_PORT" envDefault:"8080"`

	OneChannelSecretEnv string `env:"ONECHANNEL_SECRET,required,notEmpty"`
	MetaSecretEnv       string `env:"META_SECRET,required,notEmpty"`

	PostgresURLEnv             string        `env:"POSTGRES_URL,required,notEmpty"`
	PostgresPortEnv            uint16        `env:"POSTGRES_PORT" envDefault:"5432"`
	PostgresUsernameEnv        string        `env:"POSTGRES_USERNAME,required,notEmpty"`
	PostgresPasswordEnv        string        `env:"POSTGRES_PASSWORD,required,notEmpty"`
	PostgresMaxConnsEnv        int32         `env:"POSTGRES_MAX_CONNS" envDefault:"4"`
	PostgresMinConnsEnv        int32         `env:"POSTGRES_MIN_CONNS" envDefault:"0"`
	PostgresMaxConnLifeTimeEnv time.Duration `env:"POSTGRES_MAX_CONN_LIFETIME" envDefault:"1h"`
	PostgresMaxConnIdleTimeEnv time.Duration `env:"POSTGRES_MAX_CONN_IDLE_TIME" envDefault:"30m"`

	NatsURLEnv                  string        `env:"NATS_URL,required,notEmpty"`
	NatsPortEnv                 int           `env:"NATS_PORT" envDefault:"4222"`
	NatsUsernameEnv             string        `env:"NATS_USERNAME,required,notEmpty"`
	NatsPasswordEnv             string        `env:"NATS_PASSWORD,required,notEmpty"`
	NatsTimeoutEnv              time.Duration `env:"NATS_TIMEOUT" envDefault:"2s"`
	NatsMaxReconnectsEnv        int           `env:"NATS_MAX_RECONNECTS" envDefault:"-1"`
	NatsReconnectWaitEnv        time.Duration `env:"NATS_RECONNECT_WAIT" envDefault:"2s"`
	NatsRetryOnFailedConnectEnv bool          `env:"NATS_RETRY_ON_FAILED_CONNECT" envDefault:"true"`
	NatsPingIntervalEnv         time.Duration `env:"NATS_PING_INTERVAL" envDefault:"2m"`
	NatsMaxPingsOutEnv          int           `env:"NATS_MAX_PINGS_OUT" envDefault:"2"`
}

func NewConfiguration() (Configuration, error) {
	_ = godotenv.Load()

	var cfg configuration
	err := env.Parse(&cfg)
	if err != nil {
		return &cfg, err
	}

	return &cfg, nil
}

func (c *configuration) AppName() string                        { return "1channel" }
func (c *configuration) HTTPPort() string                       { return c.HTTPPortEnv }
func (c *configuration) OneChannelSecret() string               { return c.OneChannelSecretEnv }
func (c *configuration) MetaSecret() string                     { return c.MetaSecretEnv }
func (c *configuration) PostgresURL() string                    { return c.PostgresURLEnv }
func (c *configuration) PostgresPort() uint16                   { return c.PostgresPortEnv }
func (c *configuration) PostgresUsername() string               { return c.PostgresUsernameEnv }
func (c *configuration) PostgresPassword() string               { return c.PostgresPasswordEnv }
func (c *configuration) PostgresMaxConns() int32                { return c.PostgresMaxConnsEnv }
func (c *configuration) PostgresMinConns() int32                { return c.PostgresMinConnsEnv }
func (c *configuration) PostgresMaxConnLifeTime() time.Duration { return c.PostgresMaxConnLifeTimeEnv }
func (c *configuration) PostgresMaxConnIdleTime() time.Duration { return c.PostgresMaxConnIdleTimeEnv }
func (c *configuration) NatsURL() string                        { return c.NatsURLEnv }
func (c *configuration) NatsPort() int                          { return c.NatsPortEnv }
func (c *configuration) NatsUsername() string                   { return c.NatsUsernameEnv }
func (c *configuration) NatsPassword() string                   { return c.NatsPasswordEnv }
func (c *configuration) NatsTimeout() time.Duration             { return c.NatsTimeoutEnv }
func (c *configuration) NatsMaxReconnects() int                 { return c.NatsMaxReconnectsEnv }
func (c *configuration) NatsReconnectWait() time.Duration       { return c.NatsReconnectWaitEnv }
func (c *configuration) NatsRetryOnFailedConnect() bool         { return c.NatsRetryOnFailedConnectEnv }
func (c *configuration) NatsPingInterval() time.Duration        { return c.NatsPingIntervalEnv }
func (c *configuration) NatsMaxPingsOut() int                   { return c.NatsMaxPingsOutEnv }
