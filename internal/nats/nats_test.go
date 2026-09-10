package nats

import (
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/jdgonzalez907/1channel/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestNew(t *testing.T) {
	t.Run("failure - returns wrapped error when server is unreachable", func(t *testing.T) {
		// Arrange
		cfg := config.NewMockNatsConfiguration("127.0.0.1", 14222, 200*time.Millisecond)

		// Act
		client, err := New(cfg, discardLogger())

		// Assert
		require.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "connecting to nats at nats://127.0.0.1:14222")
	})
}

func TestRegister(t *testing.T) {
	tests := []struct {
		title         string
		setup         func(t *testing.T, m *MockJetStream)
		expStreamSet  bool
		expectedError string
	}{
		{
			title: "success - registers stream and starts consumer",
			setup: func(t *testing.T, m *MockJetStream) {
				t.Helper()
				cc := &MockConsumeContext{}
				cc.On("Stop").Return()
				consumer := &MockConsumer{}
				consumer.On("Consume", mock.Anything).Return(cc, nil).Once()
				m.On("CreateOrUpdateStream", mock.Anything).Return(&MockStream{}, nil).Once()
				m.On("CreateOrUpdateConsumer", messageEventReceivedStreamName, mock.Anything).
					Return(consumer, nil).Once()
			},
			expectedError: "",
			expStreamSet:  true,
		},
		{
			title: "failure - wraps stream registration error",
			setup: func(t *testing.T, m *MockJetStream) {
				t.Helper()
				m.On("CreateOrUpdateStream", mock.Anything).
					Return(&MockStream{}, errors.New("boom")).Once()
			},
			expectedError: "ensuring stream META_EVENT_RECEIVED: boom",
			expStreamSet:  false,
		},
		{
			title: "failure - wraps consumer registration error",
			setup: func(t *testing.T, m *MockJetStream) {
				t.Helper()
				m.On("CreateOrUpdateStream", mock.Anything).Return(&MockStream{}, nil).Once()
				m.On("CreateOrUpdateConsumer", messageEventReceivedStreamName, mock.Anything).
					Return((*MockConsumer)(nil), errors.New("boom")).Once()
			},
			expectedError: "ensuring consumer 1channel_message_event_received on stream META_EVENT_RECEIVED: boom",
			expStreamSet:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			js := &MockJetStream{}
			tt.setup(t, js)
			client := &Client{js: js, log: discardLogger()}

			// Act
			err := client.Register(t.Context())

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.expStreamSet, client.Streams.MessageEventReceived != nil)
			js.AssertExpectations(t)
		})
	}
}
