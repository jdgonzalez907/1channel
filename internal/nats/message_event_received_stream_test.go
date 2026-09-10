package nats

import (
	"errors"
	"testing"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewMessageMessageEventReceivedStream(t *testing.T) {
	tests := []struct {
		title         string
		setup         func(t *testing.T, m *MockJetStream)
		expectedError string
	}{
		{
			title: "success - ensures stream with configured name and subject",
			setup: func(t *testing.T, m *MockJetStream) {
				t.Helper()
				m.On("CreateOrUpdateStream", jetstream.StreamConfig{
					Name:     messageEventReceivedStreamName,
					Subjects: []string{MessageEventReceivedSubject},
				}).Return(&MockStream{}, nil).Once()
			},
			expectedError: "",
		},
		{
			title: "failure - wraps ensure stream error",
			setup: func(t *testing.T, m *MockJetStream) {
				t.Helper()
				m.On("CreateOrUpdateStream", mock.Anything).
					Return(&MockStream{}, errors.New("boom")).Once()
			},
			expectedError: "ensuring stream META_EVENT_RECEIVED: boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			js := &MockJetStream{}
			tt.setup(t, js)

			// Act
			stream, err := NewMessageMessageEventReceivedStream(t.Context(), js)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
				assert.Nil(t, stream)
			} else {
				require.NoError(t, err)
				require.NotNil(t, stream)
			}
			js.AssertExpectations(t)
		})
	}
}

func TestPublish(t *testing.T) {
	tests := []struct {
		title         string
		setup         func(t *testing.T, m *MockJetStream)
		data          []byte
		expectedError string
	}{
		{
			title: "success - publishes payload to event subject",
			setup: func(t *testing.T, m *MockJetStream) {
				t.Helper()
				m.On("Publish", MessageEventReceivedSubject, []byte("payload")).
					Return(&jetstream.PubAck{Stream: messageEventReceivedStreamName}, nil).Once()
			},
			data:          []byte("payload"),
			expectedError: "",
		},
		{
			title: "failure - wraps publish error",
			setup: func(t *testing.T, m *MockJetStream) {
				t.Helper()
				m.On("Publish", MessageEventReceivedSubject, []byte("payload")).
					Return((*jetstream.PubAck)(nil), errors.New("no responders")).Once()
			},
			data:          []byte("payload"),
			expectedError: "publishing to stream META_EVENT_RECEIVED: no responders",
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			js := &MockJetStream{}
			tt.setup(t, js)
			stream := &MessageEventReceivedStream{js: js}

			// Act
			err := stream.Publish(t.Context(), tt.data)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
			} else {
				require.NoError(t, err)
			}
			js.AssertExpectations(t)
		})
	}
}
