package config

import (
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Configuration interface {
	AppName() string

	HTTPPort() string

	LogLevel() string

	OneChannelSecret() string
	MetaSecret() string

	PostgresHost() string
	PostgresPort() uint16
	PostgresDatabase() string
	PostgresUsername() string
	PostgresPassword() string
	PostgresMaxConns() int32
	PostgresMinConns() int32
	PostgresMaxConnLifeTime() time.Duration
	PostgresMaxConnIdleTime() time.Duration
}

type configuration struct {
	HTTPPortEnv string `env:"HTTP_PORT" envDefault:"8080"`

	LogLevelEnv string `env:"LOG_LEVEL" envDefault:"info"`

	OneChannelSecretEnv string `env:"ONECHANNEL_SECRET,required,notEmpty"`
	MetaSecretEnv       string `env:"META_SECRET,required,notEmpty"`

	PostgresHostEnv            string        `env:"POSTGRES_HOST,required,notEmpty"`
	PostgresPortEnv            uint16        `env:"POSTGRES_PORT" envDefault:"5432"`
	PostgresDatabaseEnv        string        `env:"POSTGRES_DATABASE,required,notEmpty"`
	PostgresUsernameEnv        string        `env:"POSTGRES_USERNAME,required,notEmpty"`
	PostgresPasswordEnv        string        `env:"POSTGRES_PASSWORD,required,notEmpty"`
	PostgresMaxConnsEnv        int32         `env:"POSTGRES_MAX_CONNS" envDefault:"4"`
	PostgresMinConnsEnv        int32         `env:"POSTGRES_MIN_CONNS" envDefault:"0"`
	PostgresMaxConnLifeTimeEnv time.Duration `env:"POSTGRES_MAX_CONN_LIFETIME" envDefault:"1h"`
	PostgresMaxConnIdleTimeEnv time.Duration `env:"POSTGRES_MAX_CONN_IDLE_TIME" envDefault:"30m"`
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
func (c *configuration) LogLevel() string                       { return c.LogLevelEnv }
func (c *configuration) OneChannelSecret() string               { return c.OneChannelSecretEnv }
func (c *configuration) MetaSecret() string                     { return c.MetaSecretEnv }
func (c *configuration) PostgresHost() string                   { return c.PostgresHostEnv }
func (c *configuration) PostgresPort() uint16                   { return c.PostgresPortEnv }
func (c *configuration) PostgresDatabase() string               { return c.PostgresDatabaseEnv }
func (c *configuration) PostgresUsername() string               { return c.PostgresUsernameEnv }
func (c *configuration) PostgresPassword() string               { return c.PostgresPasswordEnv }
func (c *configuration) PostgresMaxConns() int32                { return c.PostgresMaxConnsEnv }
func (c *configuration) PostgresMinConns() int32                { return c.PostgresMinConnsEnv }
func (c *configuration) PostgresMaxConnLifeTime() time.Duration { return c.PostgresMaxConnLifeTimeEnv }
func (c *configuration) PostgresMaxConnIdleTime() time.Duration { return c.PostgresMaxConnIdleTimeEnv }
