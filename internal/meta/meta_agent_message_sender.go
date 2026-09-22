package meta

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/jdgonzalez907/1channel/internal/config"
	"github.com/jdgonzalez907/1channel/internal/modules/conversations/domain"
)

const (
	whatsappAPIVersion = "v25.0"
	whatsappBaseURL    = "https://graph.facebook.com"
	messagingProduct   = "whatsapp"
	recipientType      = "individual"
	messageType        = "text"
	contentType        = "application/json"

	clientTimeout             = 30 * time.Second
	clientMaxIdleConns        = 10
	clientMaxIdleConnsPerHost = 5
	clientIdleConnTimeout     = 5 * time.Minute
	clientTLSHandshakeTimeout = 10 * time.Second
)

type metaAgentMessageSender struct {
	client        *http.Client
	baseURL       string
	phoneNumberID string
	accessToken   string
}

func NewMetaAgentMessageSender(cfg config.Configuration) domain.MessageSender {
	return &metaAgentMessageSender{
		client: &http.Client{
			Timeout: clientTimeout,
			Transport: &http.Transport{
				MaxIdleConns:        clientMaxIdleConns,
				MaxIdleConnsPerHost: clientMaxIdleConnsPerHost,
				IdleConnTimeout:     clientIdleConnTimeout,
				TLSHandshakeTimeout: clientTLSHandshakeTimeout,
			},
		},
		baseURL:       whatsappBaseURL,
		phoneNumberID: cfg.WhatsAppPhoneNumberID(),
		accessToken:   cfg.WhatsAppAccessToken(),
	}
}

type messageRequest struct {
	MessagingProduct string      `json:"messaging_product"`
	RecipientType    string      `json:"recipient_type"`
	To               string      `json:"to"`
	Type             string      `json:"type"`
	Text             messageBody `json:"text"`
	BizOpaqueData    string      `json:"biz_opaque_callback_data"`
}

type messageBody struct {
	Body string `json:"body"`
}

type messageResponse struct {
	Messages []struct {
		ID string `json:"id"`
	} `json:"messages"`
}

func (s *metaAgentMessageSender) Send(ctx context.Context, to string, message *domain.Message) (string, error) {
	req, err := s.buildRequest(ctx, to, message)
	if err != nil {
		return "", err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var response messageResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(response.Messages) == 0 {
		return "", fmt.Errorf("no messages in response")
	}

	return response.Messages[0].ID, nil
}

func (s *metaAgentMessageSender) buildRequest(ctx context.Context, to string, message *domain.Message) (*http.Request, error) {
	reqBody := messageRequest{
		MessagingProduct: messagingProduct,
		RecipientType:    recipientType,
		To:               to,
		Type:             messageType,
		Text:             messageBody{Body: message.Text()},
		BizOpaqueData:    message.ID().String(),
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	endpoint, err := url.JoinPath(s.baseURL, whatsappAPIVersion, s.phoneNumberID, "messages")
	if err != nil {
		return nil, fmt.Errorf("failed to build url: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer "+s.accessToken)

	return req, nil
}
