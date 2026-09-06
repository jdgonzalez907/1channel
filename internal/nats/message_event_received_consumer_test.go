package nats

import (
	"context"
	"errors"
	"testing"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/mock"
)

func TestNewMessageMessageEventReceivedConsumerEnsuresAndSubscribes(t *testing.T) {
	cc := &MockConsumeContext{}
	cc.On("Stop").Return()

	mockConsumer := &MockConsumer{}
	mockConsumer.On("Consume", mock.Anything).Return(cc, nil).Once()

	js := &MockJetStream{}
	js.On("CreateOrUpdateConsumer", messageEventReceivedStreamName, mock.Anything).
		Return(mockConsumer, nil).
		Once()

	consumer, err := NewMessageMessageEventReceivedConsumer(context.Background(), js, discardLogger())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	mockConsumer.AssertExpectations(t)
	consumer.Stop()
	cc.AssertCalled(t, "Stop")
}

func TestNewMessageMessageEventReceivedConsumerFailsWhenEnsureFails(t *testing.T) {
	js := &MockJetStream{}
	js.On("CreateOrUpdateConsumer", messageEventReceivedStreamName, mock.Anything).
		Return((*MockConsumer)(nil), errors.New("boom")).
		Once()

	if _, err := NewMessageMessageEventReceivedConsumer(context.Background(), js, discardLogger()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestConsumeAcksHandledMessage(t *testing.T) {
	consumer := &MessageEventReceivedConsumer{log: discardLogger(), handle: tracePayload(discardLogger())}

	msg := &MockMsg{}
	msg.On("Subject").Return("meta.event_received")
	msg.On("Data").Return([]byte(`{"object":"whatsapp_business_account"}`))
	msg.On("Ack").Return(nil)

	consumer.Consume(context.Background(), msg)

	msg.AssertExpectations(t)
	msg.AssertNotCalled(t, "Nak")
}

func TestConsumeNacksFailedMessage(t *testing.T) {
	consumer := &MessageEventReceivedConsumer{
		log:    discardLogger(),
		handle: func(context.Context, jetstream.Msg) error { return errors.New("usecase failed") },
	}

	msg := &MockMsg{}
	msg.On("Subject").Return("meta.event_received")
	msg.On("Metadata").Return(&jetstream.MsgMetadata{
		Stream:       "META_EVENT_RECEIVED",
		NumDelivered: 2,
		Sequence:     jetstream.SequencePair{Stream: 42},
	}, nil)
	msg.On("Nak").Return(nil)

	consumer.Consume(context.Background(), msg)

	msg.AssertCalled(t, "Nak")
	msg.AssertNotCalled(t, "Ack")
}
