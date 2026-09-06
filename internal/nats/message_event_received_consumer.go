package nats

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go/jetstream"
)

const messageEventReceivedConsumerDurable = "1channel_message_event_received"

type MessageEventReceivedConsumer struct {
	cc     jetstream.ConsumeContext
	log    *slog.Logger
	handle func(context.Context, jetstream.Msg) error
}

func NewMessageMessageEventReceivedConsumer(ctx context.Context, js jetstream.JetStream, log *slog.Logger) (*MessageEventReceivedConsumer, error) {
	cfg := jetstream.ConsumerConfig{
		Durable:   messageEventReceivedConsumerDurable,
		AckPolicy: jetstream.AckExplicitPolicy,
	}
	consumer, err := js.CreateOrUpdateConsumer(ctx, messageEventReceivedStreamName, cfg)
	if err != nil {
		return nil, fmt.Errorf("ensuring consumer %s on stream %s: %w", cfg.Durable, messageEventReceivedStreamName, err)
	}

	c := &MessageEventReceivedConsumer{log: log}
	c.handle = tracePayload(log)

	cc, err := consumer.Consume(func(msg jetstream.Msg) { c.Consume(ctx, msg) })
	if err != nil {
		return nil, fmt.Errorf("starting pull loop for consumer %s: %w", cfg.Durable, err)
	}
	c.cc = cc

	log.Info("message event consumer running", "stream", messageEventReceivedStreamName, "durable", cfg.Durable)
	return c, nil
}

func (c *MessageEventReceivedConsumer) Consume(ctx context.Context, msg jetstream.Msg) {
	if err := c.handle(ctx, msg); err != nil {
		c.logError(err, msg)
		if err := msg.Nak(); err != nil {
			c.log.Warn("nacking message event", "subject", msg.Subject(), "error", err)
		}
		return
	}
	if err := msg.Ack(); err != nil {
		c.log.Warn("acking message event", "subject", msg.Subject(), "error", err)
	}
}

func (c *MessageEventReceivedConsumer) Stop() {
	c.cc.Stop()
}

func tracePayload(log *slog.Logger) func(context.Context, jetstream.Msg) error {
	return func(_ context.Context, msg jetstream.Msg) error {
		log.Debug("message event payload",
			"subject", msg.Subject(),
			"payload", string(msg.Data()),
		)
		return nil
	}
}

func (c *MessageEventReceivedConsumer) logError(handleErr error, msg jetstream.Msg) {
	attrs := []any{"subject", msg.Subject(), "error", handleErr}
	if md, err := msg.Metadata(); err == nil {
		attrs = append(attrs,
			"stream", md.Stream,
			"delivery_attempt", md.NumDelivered,
			"stream_seq", md.Sequence.Stream,
		)
	}
	c.log.Error("message event handling failed, nacking for redelivery", attrs...)
}
