package http

import "time"

type ContactResponse struct {
	ID         string    `json:"id"`
	ExternalID string    `json:"external_id"`
	CreatedAt  time.Time `json:"created_at"`
}
