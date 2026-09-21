package meta

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jdgonzalez907/1channel/internal/config"
	"github.com/jdgonzalez907/1channel/internal/modules/conversations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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
	}{
		{
			title:     "success - valid signature is accepted",
			body:      body,
			signature: sign(secret, body),
			expStatus: http.StatusOK,
			expBody:   "",
		},
		{
			title:     "failure - missing signature header is bad request",
			body:      body,
			signature: "",
			expStatus: http.StatusBadRequest,
			expBody:   "Bad request\n",
		},
		{
			title:     "failure - empty body is bad request",
			body:      []byte{},
			signature: sign(secret, []byte{}),
			expStatus: http.StatusBadRequest,
			expBody:   "Bad request\n",
		},
		{
			title:     "failure - body over size limit is bad request",
			body:      bigBody,
			signature: sign(secret, bigBody),
			expStatus: http.StatusBadRequest,
			expBody:   "Bad request\n",
		},
		{
			title:     "failure - tampered signature is unauthorized",
			body:      body,
			signature: sign("wrong-secret", body),
			expStatus: http.StatusUnauthorized,
			expBody:   "Unauthorized\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			cfg := &config.MockConfiguration{}
			cfg.On("MetaSecret").Return(secret)
			mockAPI := conversations.NewMockConversationsAPI()
			handler := NewMessageEventHandler(cfg, mockAPI)

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/meta/webhook", bytes.NewReader(tt.body))
			if tt.signature != "" {
				request.Header.Set(signatureHeader, tt.signature)
			}

			handler.Handle(recorder, request)

			assert.Equal(t, tt.expStatus, recorder.Code)
			assert.Equal(t, tt.expBody, recorder.Body.String())
		})
	}
}

func TestMessageEventHandlerProcessTextMessage(t *testing.T) {
	const secret = "app-secret"

	textBody := buildWebhookBody(t, webhookPayload{
		Entry: []entry{{Changes: []change{{Value: value{
			Messages: []message{{
				ID:        "wamid.abc123",
				From:      "5491112345678",
				Timestamp: "1694000000",
				Type:      "text",
				Text:      &textBody{Body: "hola"},
			}},
		}}}}},
	})

	t.Run("success - text message calls ReceiveContactMessage", func(t *testing.T) {
		cfg := &config.MockConfiguration{}
		cfg.On("MetaSecret").Return(secret)
		mockAPI := conversations.NewMockConversationsAPI()
		mockAPI.On("ReceiveContactMessage", mock.Anything, mock.MatchedBy(func(input conversations.ReceiveContactMessageInput) bool {
			return input.ExternalMessageID == "wamid.abc123" &&
				input.ExternalContactID == "5491112345678" &&
				input.Text == "hola" &&
				input.ReceivedAt.Unix() == 1694000000
		})).Return(nil).Once()

		handler := NewMessageEventHandler(cfg, mockAPI)
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/meta/webhook", bytes.NewReader(textBody))
		request.Header.Set(signatureHeader, sign(secret, textBody))

		handler.Handle(recorder, request)

		assert.Equal(t, http.StatusOK, recorder.Code)
		mockAPI.AssertExpectations(t)
	})

	t.Run("failure - ReceiveContactMessage error returns 500", func(t *testing.T) {
		cfg := &config.MockConfiguration{}
		cfg.On("MetaSecret").Return(secret)
		mockAPI := conversations.NewMockConversationsAPI()
		mockAPI.On("ReceiveContactMessage", mock.Anything, mock.Anything).Return(assert.AnError).Once()

		handler := NewMessageEventHandler(cfg, mockAPI)
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/meta/webhook", bytes.NewReader(textBody))
		request.Header.Set(signatureHeader, sign(secret, textBody))

		handler.Handle(recorder, request)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		mockAPI.AssertExpectations(t)
	})
}

func TestMessageEventHandlerSkipsNonTextMessages(t *testing.T) {
	const secret = "app-secret"

	imageBody := buildWebhookBody(t, webhookPayload{
		Entry: []entry{{Changes: []change{{Value: value{
			Messages: []message{{
				ID:        "wamid.img456",
				From:      "5491112345678",
				Timestamp: "1694000000",
				Type:      "image",
			}},
		}}}}},
	})

	cfg := &config.MockConfiguration{}
	cfg.On("MetaSecret").Return(secret)
	mockAPI := conversations.NewMockConversationsAPI()

	handler := NewMessageEventHandler(cfg, mockAPI)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/meta/webhook", bytes.NewReader(imageBody))
	request.Header.Set(signatureHeader, sign(secret, imageBody))

	handler.Handle(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockAPI.AssertNotCalled(t, "ReceiveContactMessage")
}

func TestMessageEventHandlerProcessesMultipleMessages(t *testing.T) {
	const secret = "app-secret"

	multiBody := buildWebhookBody(t, webhookPayload{
		Entry: []entry{{Changes: []change{{Value: value{
			Messages: []message{
				{ID: "wamid.1", From: "5491111111111", Timestamp: "1694000000", Type: "text", Text: &textBody{Body: "msg1"}},
				{ID: "wamid.2", From: "5492222222222", Timestamp: "1694000001", Type: "text", Text: &textBody{Body: "msg2"}},
			},
		}}}}},
	})

	cfg := &config.MockConfiguration{}
	cfg.On("MetaSecret").Return(secret)
	mockAPI := conversations.NewMockConversationsAPI()
	mockAPI.On("ReceiveContactMessage", mock.Anything, mock.MatchedBy(func(input conversations.ReceiveContactMessageInput) bool {
		return input.ExternalMessageID == "wamid.1"
	})).Return(nil).Once()
	mockAPI.On("ReceiveContactMessage", mock.Anything, mock.MatchedBy(func(input conversations.ReceiveContactMessageInput) bool {
		return input.ExternalMessageID == "wamid.2"
	})).Return(nil).Once()

	handler := NewMessageEventHandler(cfg, mockAPI)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/meta/webhook", bytes.NewReader(multiBody))
	request.Header.Set(signatureHeader, sign(secret, multiBody))

	handler.Handle(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockAPI.AssertNumberOfCalls(t, "ReceiveContactMessage", 2)
}

func TestMessageEventHandlerStopsOnFirstError(t *testing.T) {
	const secret = "app-secret"

	multiBody := buildWebhookBody(t, webhookPayload{
		Entry: []entry{{Changes: []change{{Value: value{
			Messages: []message{
				{ID: "wamid.1", From: "5491111111111", Timestamp: "1694000000", Type: "text", Text: &textBody{Body: "msg1"}},
				{ID: "wamid.2", From: "5492222222222", Timestamp: "1694000001", Type: "text", Text: &textBody{Body: "msg2"}},
			},
		}}}}},
	})

	cfg := &config.MockConfiguration{}
	cfg.On("MetaSecret").Return(secret)
	mockAPI := conversations.NewMockConversationsAPI()
	mockAPI.On("ReceiveContactMessage", mock.Anything, mock.MatchedBy(func(input conversations.ReceiveContactMessageInput) bool {
		return input.ExternalMessageID == "wamid.1"
	})).Return(assert.AnError).Once()

	handler := NewMessageEventHandler(cfg, mockAPI)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/meta/webhook", bytes.NewReader(multiBody))
	request.Header.Set(signatureHeader, sign(secret, multiBody))

	handler.Handle(recorder, request)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockAPI.AssertNotCalled(t, "ReceiveContactMessage", mock.Anything, mock.MatchedBy(func(input conversations.ReceiveContactMessageInput) bool {
		return input.ExternalMessageID == "wamid.2"
	}))
}

func buildWebhookBody(t *testing.T, payload webhookPayload) []byte {
	t.Helper()
	body, err := json.Marshal(payload)
	require.NoError(t, err)
	return body
}
