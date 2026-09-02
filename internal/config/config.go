package config

import (
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	HTTPPort string `env:"HTTP_PORT,required,notEmpty"`

	OneChannelSecret string `env:"ONECHANNEL_SECRET,required,notEmpty"`
	MetaSecret       string `env:"META_SECRET,required,notEmpty"`

	DBURL             string        `env:"DB_URL,required,notEmpty"`
	DBMaxConns        int32         `env:"DB_MAX_CONNS,required,notEmpty"`
	DBMinConns        int32         `env:"DB_MIN_CONNS,required,notEmpty"`
	DBMaxConnIdleTime time.Duration `env:"DB_MAX_CONN_IDLE_TIME,required,notEmpty"`

	NatsURL           string        `env:"NATS_URL,required,notEmpty"`
	NatsClientName    string        `env:"NATS_CLIENT_NAME,required,notEmpty"`
	NatsTimeout       time.Duration `env:"NATS_TIMEOUT,required,notEmpty"`
	NatsMaxReconnects int           `env:"NATS_MAX_RECONNECTS,required,notEmpty"`
	NatsReconnectWait time.Duration `env:"NATS_RECONNECT_WAIT,required,notEmpty"`
}

func NewConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = env.Parse(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, err
}
