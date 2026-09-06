package nats

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jdgonzalez907/1channel/internal/config"
	natsgo "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type Client struct {
	conn *natsgo.Conn
	js   jetstream.JetStream
	log  *slog.Logger

	Streams struct {
		MessageEventReceived *MessageEventReceivedStream
	}
}

func New(cfg config.Configuration, log *slog.Logger) (*Client, error) {
	url := fmt.Sprintf("nats://%s:%d", cfg.NatsHost(), cfg.NatsPort())

	opts := []natsgo.Option{
		natsgo.Name(cfg.AppName()),
		natsgo.UserInfo(cfg.NatsUsername(), cfg.NatsPassword()),
		natsgo.Timeout(cfg.NatsTimeout()),
		natsgo.ReconnectWait(cfg.NatsReconnectWait()),
		natsgo.MaxReconnects(cfg.NatsMaxReconnects()),
		natsgo.PingInterval(cfg.NatsPingInterval()),
		natsgo.MaxPingsOutstanding(cfg.NatsMaxPingsOut()),
		natsgo.ReconnectHandler(func(conn *natsgo.Conn) {
			log.Info("nats reconnected", "url", conn.ConnectedUrl())
		}),
		natsgo.DisconnectErrHandler(func(conn *natsgo.Conn, err error) {
			log.Warn("nats disconnected", "url", url, "error", err)
		}),
		natsgo.ClosedHandler(func(conn *natsgo.Conn) {
			log.Info("nats connection closed")
		}),
	}

	conn, err := natsgo.Connect(url, opts...)
	if err != nil {
		return nil, fmt.Errorf("connecting to nats at %s: %w", url, err)
	}

	js, err := jetstream.New(conn)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("creating jetstream context: %w", err)
	}

	log.Info("nats connected", "url", conn.ConnectedUrl())
	return &Client{conn: conn, js: js, log: log}, nil
}

func (c *Client) Drain() error {
	return c.conn.Drain()
}

func (c *Client) Register(ctx context.Context) error {
	messageEventReceivedStream, err := NewMessageMessageEventReceivedStream(ctx, c.js)
	if err != nil {
		return err
	}
	c.Streams.MessageEventReceived = messageEventReceivedStream

	if c.Streams.MessageEventReceived == nil {
		return errors.New("registering nats streams: EventReceived is nil")
	}

	consumer, err := NewMessageMessageEventReceivedConsumer(ctx, c.js, c.log)
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		consumer.Stop()
	}()

	return nil
}
