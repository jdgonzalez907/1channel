package nats

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go/jetstream"
)

const (
	messageEventReceivedStreamName = "META_EVENT_RECEIVED"

	MessageEventReceivedSubject = "meta.event_received"
)

type MessageEventReceivedStream struct {
	js jetstream.JetStream
}

func NewMessageMessageEventReceivedStream(ctx context.Context, js jetstream.JetStream) (*MessageEventReceivedStream, error) {
	cfg := jetstream.StreamConfig{
		Name:     messageEventReceivedStreamName,
		Subjects: []string{MessageEventReceivedSubject},
	}
	if _, err := js.CreateOrUpdateStream(ctx, cfg); err != nil {
		return nil, fmt.Errorf("ensuring stream %s: %w", cfg.Name, err)
	}
	return &MessageEventReceivedStream{js: js}, nil
}

func (s *MessageEventReceivedStream) Publish(ctx context.Context, data []byte) error {
	if _, err := s.js.Publish(ctx, MessageEventReceivedSubject, data); err != nil {
		return fmt.Errorf("publishing to stream %s: %w", messageEventReceivedStreamName, err)
	}
	return nil
}
