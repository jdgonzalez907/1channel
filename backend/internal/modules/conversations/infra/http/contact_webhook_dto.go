package http

type ContactWebhookRequest struct {
	Event             string  `json:"event"`
	ExternalContactID string  `json:"external_contact_id"`
	ExternalMessageID string  `json:"external_message_id"`
	Text              *string `json:"text"`
}
