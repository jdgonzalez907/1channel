package meta

type webhookPayload struct {
	Object string         `json:"object"`
	Entry  []webhookEntry `json:"entry"`
}

type webhookEntry struct {
	ID        string           `json:"id"`
	Time      int64            `json:"time"`
	Messaging []messagingEvent `json:"messaging"`
}

type messagingEvent struct {
	Sender      webhookParticipant  `json:"sender"`
	Recipient   webhookParticipant  `json:"recipient"`
	Timestamp   int64               `json:"timestamp"`
	Message     *messagePayload     `json:"message"`
	MessageEdit *messageEditPayload `json:"message_edit"`
}

type webhookParticipant struct {
	ID string `json:"id"`
}

type messagePayload struct {
	Mid    string  `json:"mid"`
	Text   *string `json:"text"`
	IsEcho bool    `json:"is_echo"`
}

type messageEditPayload struct {
	Mid  string `json:"mid"`
	Text string `json:"text"`
}

type sendMessageRequest struct {
	Recipient     sendRecipient `json:"recipient"`
	MessagingType string        `json:"messaging_type"`
	Message       sendContent   `json:"message"`
}

type sendRecipient struct {
	ID string `json:"id"`
}

type sendContent struct {
	Text string `json:"text"`
}
