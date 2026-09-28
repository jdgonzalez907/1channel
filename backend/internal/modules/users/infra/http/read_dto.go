package http

import "time"

type UserDetailResponse struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}
