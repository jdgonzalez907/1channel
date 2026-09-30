package meta

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/jdgonzalez907/1channel/internal/modules/conversations/domain"
)

const (
	graphBaseURL          = "https://graph.facebook.com"
	graphAPIVersion       = "v25.0"
	messagingTypeResponse = "RESPONSE"
	messagingPath         = "messages"
	accessTokenParam      = "access_token"
	contentTypeJSON       = "application/json"
	httpClientTimeout     = 4 * time.Second
)

var _ domain.AgentMessageSender = (*MessengerSender)(nil)

type MessengerSender struct {
	config  Config
	client  *http.Client
	baseURL string
}

func NewMessengerSender(config Config) *MessengerSender {
	return &MessengerSender{
		config:  config,
		client:  &http.Client{Timeout: httpClientTimeout},
		baseURL: graphBaseURL,
	}
}

func (s *MessengerSender) Send(ctx context.Context, to string, message *domain.Message) (string, error) {
	text := ""
	if message.Text() != nil {
		text = *message.Text()
	}

	payload, err := json.Marshal(sendMessageRequest{
		Recipient:     sendRecipient{ID: to},
		MessagingType: messagingTypeResponse,
		Message:       sendContent{Text: text},
	})
	if err != nil {
		return "", err
	}

	endpoint, err := url.Parse(fmt.Sprintf("%s/%s/%s/%s", s.baseURL, graphAPIVersion, s.config.PageID, messagingPath))
	if err != nil {
		return "", err
	}

	query := endpoint.Query()
	query.Set(accessTokenParam, s.config.PageAccessToken)
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", contentTypeJSON)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("meta send api returned status %d", resp.StatusCode)
	}

	var result sendMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.MessageID == "" {
		return "", errors.New("meta send api returned an empty message id")
	}

	return result.MessageID, nil
}
