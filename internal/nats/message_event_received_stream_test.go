package nats

import (
	"context"
	"errors"
	"testing"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/mock"
)

func TestNewMessageMessageEventReceivedStreamEnsuresConfiguredStream(t *testing.T) {
	js := &MockJetStream{}
	js.On("CreateOrUpdateStream", mock.Anything).
		Return(&MockStream{}, nil).
		Once()

	if _, err := NewMessageMessageEventReceivedStream(context.Background(), js); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	js.AssertCalled(t, "CreateOrUpdateStream", jetstream.StreamConfig{
		Name:     "META_EVENT_RECEIVED",
		Subjects: []string{"meta.event_received"},
	})
}

func TestNewMessageMessageEventReceivedStreamWrapsEnsureError(t *testing.T) {
	js := &MockJetStream{}
	boom := errors.New("boom")
	js.On("CreateOrUpdateStream", mock.Anything).
		Return(&MockStream{}, boom).
		Once()

	if _, err := NewMessageMessageEventReceivedStream(context.Background(), js); !errors.Is(err, boom) {
		t.Fatalf("expected wrapped boom, got %v", err)
	}
}

func TestMessageEventReceivedStreamPublish(t *testing.T) {
	tests := []struct {
		name    string
		ack     *jetstream.PubAck
		err     error
		wantErr bool
	}{
		{
			name:    "publishes to meta subject and succeeds",
			ack:     &jetstream.PubAck{Stream: "META_EVENT_RECEIVED"},
			wantErr: false,
		},
		{
			name:    "wraps publish failure",
			ack:     &jetstream.PubAck{},
			err:     errors.New("no responders"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			js := &MockJetStream{}
			js.On("Publish", "meta.event_received", []byte("payload")).
				Return(tt.ack, tt.err).
				Once()

			stream := &MessageEventReceivedStream{js: js}
			err := stream.Publish(context.Background(), []byte("payload"))

			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			js.AssertExpectations(t)
		})
	}
}
