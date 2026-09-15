package meta

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jdgonzalez907/1channel/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestMessageEventHandlerHandle(t *testing.T) {
	const secret = "app-secret"

	body := []byte(`{"object":"whatsapp_business_account","entry":[]}`)
	bigBody := bytes.Repeat([]byte("a"), maxRequestBodyBytes+1)

	tests := []struct {
		title     string
		body      []byte
		signature string
		expStatus int
		expBody   string
		expLog    string
	}{
		{
			title:     "success - valid signature is accepted",
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
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			logs := captureSlogLogs(t)
			cfg := &config.MockConfiguration{}
			cfg.On("MetaSecret").Return(secret)
			handler := NewMessageEventHandler(cfg)

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
		})
	}
}
