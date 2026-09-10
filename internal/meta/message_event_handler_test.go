package meta

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jdgonzalez907/1channel/internal/config"
	"github.com/jdgonzalez907/1channel/internal/nats"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newRegisteredStream(t *testing.T) (*nats.MessageEventReceivedStream, *nats.MockJetStream) {
	t.Helper()
	js := &nats.MockJetStream{}
	js.On("CreateOrUpdateStream", mock.Anything).Return(&nats.MockStream{}, nil).Once()
	stream, err := nats.NewMessageMessageEventReceivedStream(t.Context(), js)
	require.NoError(t, err)
	return stream, js
}

func TestMessageEventHandlerHandle(t *testing.T) {
	const secret = "app-secret"

	body := []byte(`{"object":"whatsapp_business_account","entry":[]}`)
	bigBody := bytes.Repeat([]byte("a"), maxRequestBodyBytes+1)

	tests := []struct {
		title     string
		setup     func(t *testing.T, js *nats.MockJetStream)
		body      []byte
		signature string
		expStatus int
		expBody   string
		expLog    string
	}{
		{
			title: "success - valid signature is published to nats and accepted",
			setup: func(t *testing.T, js *nats.MockJetStream) {
				t.Helper()
				js.On("Publish", nats.MessageEventReceivedSubject, body).
					Return(&jetstream.PubAck{}, nil).Once()
			},
			body:      body,
			signature: sign(secret, body),
			expStatus: http.StatusOK,
			expBody:   "",
			expLog:    "message event received",
		},
		{
			title:     "failure - missing signature header is bad request",
			body:      body,
			signature: "",
			expStatus: http.StatusBadRequest,
			expBody:   "Bad request\n",
			expLog:    "",
		},
		{
			title:     "failure - empty body is bad request",
			body:      []byte{},
			signature: sign(secret, []byte{}),
			expStatus: http.StatusBadRequest,
			expBody:   "Bad request\n",
			expLog:    "",
		},
		{
			title:     "failure - body over size limit is bad request",
			body:      bigBody,
			signature: sign(secret, bigBody),
			expStatus: http.StatusBadRequest,
			expBody:   "Bad request\n",
			expLog:    "",
		},
		{
			title:     "failure - tampered signature is unauthorized",
			body:      body,
			signature: sign("wrong-secret", body),
			expStatus: http.StatusUnauthorized,
			expBody:   "Unauthorized\n",
			expLog:    "",
		},
		{
			title: "failure - publish error logs and returns internal server error",
			setup: func(t *testing.T, js *nats.MockJetStream) {
				t.Helper()
				js.On("Publish", nats.MessageEventReceivedSubject, body).
					Return(&jetstream.PubAck{}, errors.New("nats down")).Once()
			},
			body:      body,
			signature: sign(secret, body),
			expStatus: http.StatusInternalServerError,
			expBody:   "Internal server error\n",
			expLog:    "publishing message event to nats",
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			logs := captureSlogLogs(t)
			stream, js := newRegisteredStream(t)
			if tt.setup != nil {
				tt.setup(t, js)
			}
			cfg := &config.MockConfiguration{}
			cfg.On("MetaSecret").Return(secret)
			handler := NewMessageEventHandler(cfg, stream)

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/meta/webhook", bytes.NewReader(tt.body))
			if tt.signature != "" {
				request.Header.Set(signatureHeader, tt.signature)
			}

			// Act
			handler.Handle(recorder, request)

			// Assert
			assert.Equal(t, tt.expStatus, recorder.Code)
			assert.Equal(t, tt.expBody, recorder.Body.String())
			if tt.expLog != "" {
				assert.Contains(t, logs.String(), tt.expLog)
			} else {
				assert.NotContains(t, logs.String(), "message event received")
			}
			if tt.setup == nil {
				js.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything)
			}
			js.AssertExpectations(t)
		})
	}
}
