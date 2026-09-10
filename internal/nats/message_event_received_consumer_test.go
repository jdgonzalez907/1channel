package nats

import (
	"context"
	"errors"
	"testing"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewMessageMessageEventReceivedConsumer(t *testing.T) {
	tests := []struct {
		title         string
		setup         func(t *testing.T, m *MockJetStream)
		expectedError string
	}{
		{
			title: "success - ensures consumer and starts pull loop",
			setup: func(t *testing.T, m *MockJetStream) {
				t.Helper()
				cc := &MockConsumeContext{}
				consumer := &MockConsumer{}
				consumer.On("Consume", mock.Anything).Return(cc, nil).Once()
				m.On("CreateOrUpdateConsumer", messageEventReceivedStreamName, jetstream.ConsumerConfig{
					Durable:   "1channel_message_event_received",
					AckPolicy: jetstream.AckExplicitPolicy,
				}).Return(consumer, nil).Once()
			},
			expectedError: "",
		},
		{
			title: "failure - wraps consumer ensure error",
			setup: func(t *testing.T, m *MockJetStream) {
				t.Helper()
				m.On("CreateOrUpdateConsumer", messageEventReceivedStreamName, mock.Anything).
					Return((*MockConsumer)(nil), errors.New("boom")).Once()
			},
			expectedError: "ensuring consumer 1channel_message_event_received on stream META_EVENT_RECEIVED: boom",
		},
		{
			title: "failure - wraps consume loop start error",
			setup: func(t *testing.T, m *MockJetStream) {
				t.Helper()
				consumer := &MockConsumer{}
				consumer.On("Consume", mock.Anything).
					Return((*MockConsumeContext)(nil), errors.New("boom")).Once()
				m.On("CreateOrUpdateConsumer", messageEventReceivedStreamName, mock.Anything).
					Return(consumer, nil).Once()
			},
			expectedError: "starting pull loop for consumer 1channel_message_event_received: boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			js := &MockJetStream{}
			tt.setup(t, js)

			// Act
			consumer, err := NewMessageMessageEventReceivedConsumer(t.Context(), js, discardLogger())

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
				assert.Nil(t, consumer)
			} else {
				require.NoError(t, err)
				require.NotNil(t, consumer)
			}
			js.AssertExpectations(t)
		})
	}
}

func TestConsume(t *testing.T) {
	failingHandle := func(context.Context, jetstream.Msg) error { return errors.New("usecase failed") }

	tests := []struct {
		title     string
		handleErr bool
		setup     func(t *testing.T, m *MockMsg)
	}{
		{
			title: "success - traces payload and acks handled message",
			setup: func(t *testing.T, m *MockMsg) {
				t.Helper()
				m.On("Subject").Return(MessageEventReceivedSubject).Once()
				m.On("Data").Return([]byte(`{"object":"whatsapp_business_account"}`)).Once()
				m.On("Ack").Return(nil).Once()
			},
		},
		{
			title:     "failure - nacks message when handling fails",
			handleErr: true,
			setup: func(t *testing.T, m *MockMsg) {
				t.Helper()
				m.On("Subject").Return(MessageEventReceivedSubject).Once()
				m.On("Metadata").Return(&jetstream.MsgMetadata{
					Stream:       messageEventReceivedStreamName,
					NumDelivered: 2,
					Sequence:     jetstream.SequencePair{Stream: 42},
				}, nil).Once()
				m.On("Nak").Return(nil).Once()
			},
		},
		{
			title:     "failure - logs warning when nak itself fails",
			handleErr: true,
			setup: func(t *testing.T, m *MockMsg) {
				t.Helper()
				m.On("Subject").Return(MessageEventReceivedSubject).Twice()
				m.On("Metadata").Return((*jetstream.MsgMetadata)(nil), errors.New("no metadata")).Once()
				m.On("Nak").Return(errors.New("nak failed")).Once()
			},
		},
		{
			title: "failure - logs warning when ack itself fails",
			setup: func(t *testing.T, m *MockMsg) {
				t.Helper()
				m.On("Subject").Return(MessageEventReceivedSubject).Twice()
				m.On("Data").Return([]byte("payload")).Once()
				m.On("Ack").Return(errors.New("ack failed")).Once()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			msg := &MockMsg{}
			tt.setup(t, msg)
			consumer := &MessageEventReceivedConsumer{log: discardLogger()}
			if tt.handleErr {
				consumer.handle = failingHandle
			} else {
				consumer.handle = tracePayload(discardLogger())
			}

			// Act
			consumer.Consume(t.Context(), msg)

			// Assert
			msg.AssertExpectations(t)
			if tt.handleErr {
				msg.AssertNotCalled(t, "Ack")
			} else {
				msg.AssertNotCalled(t, "Nak")
			}
		})
	}
}

func TestStop(t *testing.T) {
	// Arrange
	cc := &MockConsumeContext{}
	cc.On("Stop").Return().Once()
	consumer := &MessageEventReceivedConsumer{cc: cc, log: discardLogger()}

	// Act
	consumer.Stop()

	// Assert
	cc.AssertExpectations(t)
}
