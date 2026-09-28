package http

type SendAgentMessageRequest struct {
	Text string `json:"text"`
}

type MarkMessagesReadRequest struct {
	Status string `json:"status"`
}

type ResolveConversationRequest struct {
	Status string `json:"status"`
}

type MessageResponse struct {
	ID string `json:"id"`
}
