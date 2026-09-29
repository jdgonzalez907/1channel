package http

import "time"

type ContactResponse struct {
	ID         string `json:"id"`
	ExternalID string `json:"external_id"`
	Label      string `json:"label"`
}

type PersonalInformationResponse struct {
	ID                   string    `json:"id"`
	IdentificationNumber string    `json:"identification_number"`
	FirstName            *string   `json:"first_name"`
	LastName             *string   `json:"last_name"`
	PhoneNumber          *string   `json:"phone_number"`
	Email                *string   `json:"email"`
	Address              *string   `json:"address"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type LastMessageResponse struct {
	Text   *string   `json:"text"`
	SentAt time.Time `json:"sent_at"`
	Owner  string    `json:"owner"`
}

type ConversationListItemResponse struct {
	ID          string              `json:"id"`
	Status      string              `json:"status"`
	Contact     ContactResponse     `json:"contact"`
	LastMessage LastMessageResponse `json:"last_message"`
	UnreadCount int32               `json:"unread_count"`
}

type ConversationListResponse struct {
	Items            []ConversationListItemResponse `json:"items"`
	NextBeforeSentAt *string                        `json:"next_before_sent_at"`
	NextBeforeID     *string                        `json:"next_before_id"`
}

type ConversationMessageResponse struct {
	ID        string     `json:"id"`
	Status    string     `json:"status"`
	Type      string     `json:"type"`
	Text      *string    `json:"text"`
	Owner     string     `json:"owner"`
	SentAt    time.Time  `json:"sent_at"`
	ReadAt    *time.Time `json:"read_at"`
	EditedAt  *time.Time `json:"edited_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

type ConversationDetailResponse struct {
	ID                  string                        `json:"id"`
	Status              string                        `json:"status"`
	Contact             ContactResponse               `json:"contact"`
	PersonalInformation *PersonalInformationResponse  `json:"personal_information"`
	AgentID             *string                       `json:"agent_id"`
	UnreadCount         int32                         `json:"unread_count"`
	Messages            []ConversationMessageResponse `json:"messages"`
	NextBeforeSentAt    *string                       `json:"next_before_sent_at"`
	NextBeforeID        *string                       `json:"next_before_id"`
}
