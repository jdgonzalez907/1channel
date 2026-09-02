package config

import (
	"context"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func NewNatsConnection(ctx context.Context, cfg *Config) (*nats.Conn, jetstream.JetStream, error) {
	nc, err := nats.Connect(cfg.NatsURL,
		nats.Name(cfg.NatsClientName),
		nats.Timeout(cfg.NatsTimeout),
		nats.MaxReconnects(cfg.NatsMaxReconnects),
		nats.ReconnectWait(cfg.NatsReconnectWait),
	)
	if err != nil {
		return nil, nil, err
	}

	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()
		return nil, nil, err
	}

	ctxPing, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if _, err := js.AccountInfo(ctxPing); err != nil {
		nc.Close()
		return nil, nil, err
	}

	return nc, js, nil
}
